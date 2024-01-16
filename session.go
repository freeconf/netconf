package netconf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/estream"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/patch/xml"
)

type Session struct {
	dev   device.Device
	mgr   SessionManager
	out   io.Writer
	inRaw io.Reader
	in    <-chan io.Reader
	Id    int64
	user  string
}

// ErrEOS signals the session should be closed gracefully.
var ErrEOS = errors.New("end of session") // not really an error but linter wants "Err" prefix

func NewSession(mgr SessionManager, user string, dev device.Device, in io.Reader, out io.Writer) *Session {
	return &Session{
		mgr:   mgr,
		dev:   dev,
		Id:    mgr.NextSessionId(),
		inRaw: in,
		in:    NewChunkedRdr(in),
		out:   out,
		user:  user,
	}
}

func (ses *Session) User() string {
	return ses.user
}

func (ses *Session) readMessages(ctx context.Context) error {
	hello, err := DecodeRequest(ses.inRaw)
	if err != nil {
		return err
	}
	if hello.Hello == nil {
		return errors.New("expected initial hello message")
	}
	fc.Debug.Printf("got hello request ses=%d", ses.Id)
	if err = ses.handleHello(hello.Hello); err != nil {
		return err
	}
	for {
		err := ses.readRequest(ctx)
		if err != nil {
			return err
		}
	}
}

func (ses *Session) readRequest(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ErrEOS
	case in := <-ses.in:
		req, err := DecodeRequest(in)
		if err != nil || req == nil {
			return err
		}
		if req.Rpc != nil {
			return ses.handleRpc(req.Rpc)
		}
		if req.Hello != nil {
			return ses.handleHello(req.Hello)
		}
		return fmt.Errorf("unsupported message %s", req.Other.XMLName.Local)
	}
}

func (ses *Session) readFilter(f *RpcFilter, c node.ContentConstraint) ([]*node.Selection, error) {
	if f == nil {
		// Sec 6.4.1 - no filter returns all data
		f = &RpcFilter{}
		for name := range ses.dev.Modules() {
			f.Elems = append(f.Elems, &Msg{XMLName: xml.Name{Local: name}})
		}
	} else if len(f.Elems) == 0 {
		// Sec 6.4.1 - empty filter returns nothing
		return nil, nil
	}

	var sels = make([]*node.Selection, 0)
	for _, e := range f.Elems {
		b, err := ses.dev.Browser(e.XMLName.Local)
		if err != nil {
			return nil, err
		}
		if e.XMLName.Space != "" {
			if b.Meta.Namespace() != e.XMLName.Space {
				continue
			}
		}
		sel := b.Root()
		sel.Constraints.AddConstraint("content", 0, 0, c)
		if f.Type == "xpath" {
			sel, err := f.CompileXPath(ses.dev)
			if err != nil {
				return nil, err
			}
			sels = append(sels, sel)
		} else if f.Type == "subtree" || f.Type == "" {
			var f subtreeFilter
			if err := compileSubtree(e, &f); err != nil {
				return nil, err
			}
			sel.Constraints.AddConstraint("filter", 10, 0, &f)

			sels = append(sels, sel)
		}
	}
	return sels, nil
}

func (ses *Session) findBrowserByNs(ns string) (*node.Browser, error) {
	for _, mod := range ses.dev.Modules() {
		if mod.Namespace() == ns {
			return ses.dev.Browser(mod.Ident())
		}
	}
	return nil, fmt.Errorf("browser for namespace '%s' not found", ns)
}

const (
	Base_1_1 = "urn:ietf:params:netconf:base:1.1"
)

func (ses *Session) Hello() *HelloMsg {
	// For recognized capabilties, see
	//   https://datatracker.ietf.org/doc/html/rfc6241#section-10.4
	return &HelloMsg{
		SessionId: strconv.FormatInt(ses.Id, 10),
		Capabilities: []*Msg{
			{Content: Base_1_1},
			{Content: "urn:ietf:params:netconf:capability:notification:1.0"},
		},
	}
}

func (ses *Session) handleGet(get *RpcGet, resp *RpcReply, c node.ContentConstraint) error {
	sels, err := ses.readFilter(get.Filter, c)
	if err != nil {
		return err
	}
	resp.Data = &RpcData{}
	for _, sel := range sels {
		mod := meta.OriginalModule(sel.Meta())
		cfg := &nodeutil.XMLWtr2{
			XMLName: xml.Name{
				Local: sel.Meta().Ident(),
				Space: mod.Namespace(),
			},
		}
		if err := sel.UpsertInto(cfg); err != nil {
			return err
		}
		resp.Data.Nodes = append(resp.Data.Nodes, cfg)
	}
	return nil
}

func (ses *Session) handleEdit(edit *RpcEdit, resp *RpcReply) error {
	defaultOp := edit.DefaultOperation
	if defaultOp == "" {
		defaultOp = "merge"
	}
	if edit.Config == nil {
		return fmt.Errorf("edit config with no config specified")
	}
	for _, n := range edit.Config.Nodes {
		b, err := ses.dev.Browser(n.XMLName.Local)
		if err != nil {
			return err
		}
		edits, err := buildEdits(defaultOp, n, b.Meta)
		if err != nil {
			return err
		}
		root := b.Root()
		for _, e := range edits {
			sel, err := root.Find(e.path)
			if err != nil {
				return err
			}
			switch e.op {
			case "merge":
				err = sel.UpsertFrom(e.n)
			case "replace":
				err = sel.ReplaceFrom(e.n)
			case "create":
				err = sel.InsertFrom(e.n)
			case "remove":
				if sel != nil {
					err = sel.Delete()
				}
			case "delete":
				if sel == nil {
					err = fmt.Errorf("node with path '%s' does not exist.  try remove operation to ignore this error", e.path)
				} else {
					err = sel.Delete()
				}
			default:
				return fmt.Errorf("edit config operation '%s' not implemented or recognized", e.op)
			}
			if err == nil {
				return err
			}
		}
	}
	return nil
}

func (ses *Session) handleAction(rpc *nodeutil.XmlNode) (*nodeutil.XMLWtr2, error) {
	b, err := ses.findBrowserByNs(rpc.XMLName.Space)
	if err != nil {
		return nil, err
	}
	sel, err := b.Root().Find(rpc.XMLName.Local)
	if err != nil {
		return nil, err
	}
	out, err := sel.Action(rpc)
	if err != nil {
		return nil, err
	}
	if out != nil {
		var resp nodeutil.XMLWtr2
		if err := out.UpsertInto(&resp); err != nil {
			return nil, err
		}
		return &resp, nil
	}
	return nil, nil
}

func (ses *Session) handleRpc(rpc *RpcMsg) error {
	close := false
	var err error
	resp := &RpcReply{MessageId: rpc.MessageId}
	if rpc.GetConfig != nil {
		fc.Debug.Printf("get config message ses=%d", ses.Id)
		err = ses.handleGet(rpc.GetConfig, resp, node.ContentConfig)
	} else if rpc.Get != nil {
		fc.Debug.Printf("get metrics message ses=%d", ses.Id)
		err = ses.handleGet(rpc.Get, resp, node.ContentOperational)
	} else if rpc.EditConfig != nil {
		fc.Debug.Printf("edit message ses=%d", ses.Id)
		err = ses.handleEdit(rpc.EditConfig, resp)
	} else if rpc.Kill != nil {
		fc.Debug.Printf("kill message ses=%d", ses.Id)
		// abort, unlock, rollback
		resp.OK = &Msg{}
		close = true
	} else if rpc.Close != nil {
		fc.Debug.Printf("close message ses=%d", ses.Id)
		resp.OK = &Msg{}
		close = true
	} else if rpc.CreateSubscription != nil {
		fc.Debug.Printf("create subscription message ses=%d", ses.Id)
		err = ses.handleCreateSubscription(rpc.CreateSubscription)
	} else if rpc.Action != nil {
		fc.Debug.Printf("rpc/action ses=%d", ses.Id)
		var out *nodeutil.XMLWtr2
		out, err = ses.handleAction(rpc.Action)
		if err == nil {
			if out == nil {
				resp.OK = &Msg{}
			} else {
				resp.Out = out.Elem
			}
		}
	} else {
		err = fmt.Errorf("unrecognized rpc command")
	}
	if err != nil {
		return err
	}
	out := NewChunkedWtr(ses.out)
	defer out.Close()
	if err = WriteResponse(resp, out); err != nil {
		return err
	}
	if close {
		return ErrEOS
	}
	return nil
}

func (ses *Session) handleCreateSubscription(create *CreateSubscription) error {
	req := estream.EstablishRequest{
		Stream: create.Stream,
		// pass along possible filter
	}
	sub, err := ses.mgr.StreamService().EstablishSubscription(req)
	if err != nil {
		return err
	}
	name := fmt.Sprintf("sub-%s", sub.Id)
	err = sub.AddReceiver(name, func(e estream.ReceiverEvent) (estream.RecvState, error) {
		resp := Notification{
			EventTime: e.EventTime,
		}
		if err := e.Event.UpsertInto(&resp.Event); err != nil {
			return estream.RecvStateSuspended, fmt.Errorf("error encoding event %w", err)
		}
		out := NewChunkedWtr(ses.out)
		defer out.Close()
		if err = WriteResponse(resp, out); err != nil {
			return estream.RecvStateSuspended, fmt.Errorf("error encoding notification %w", err)
		}
		return estream.RecvStateActive, nil
	})
	return err
}

func (ses *Session) handleHello(h *HelloMsg) error {
	if h.SessionId != "" {
		// RFC6241 Section 8.1
		return fmt.Errorf("session id not allowed from client")
	}
	versionOk := false
	for _, c := range h.Capabilities {
		if c.Content == Base_1_1 {
			versionOk = true

		}
	}
	if !versionOk {
		return fmt.Errorf("only compatible base version '%s'", Base_1_1)
	}
	return nil
}
