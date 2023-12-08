package netconf

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

type Session struct {
	dev   device.Device
	mgr   SessionManager
	out   io.Writer
	inRaw io.Reader
	in    <-chan io.Reader
	Id    int64
}

func NewSession(mgr SessionManager, dev device.Device, in io.Reader, out io.Writer) *Session {
	return &Session{
		mgr:   mgr,
		dev:   dev,
		Id:    mgr.NextSessionId(),
		inRaw: in,
		in:    NewChunkedRdr(in),
		out:   out,
	}
}

func (ses *Session) readMessages() error {
	hello, err := DecodeRequest(ses.inRaw)
	if err != nil {
		return err
	}
	if hello.Hello == nil {
		return errors.New("expected initial hello message")
	}
	if err = ses.handleHello(hello.Hello); err != nil {
		return err
	}
	for {
		err := ses.readRequest()
		if err != nil {
			if err == io.EOF {
				fc.Debug.Printf("ending session %d", ses.Id)
				return nil
			}
			return err
		}
	}
}

func (ses *Session) readRequest() error {
	req, err := DecodeRequest(<-ses.in)
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

func (ses *Session) readFilter(f *RpcFilter, c node.ContentConstraint) ([]*node.Selection, error) {
	var sels = make([]*node.Selection, 0, len(ses.dev.Modules()))
	for name := range ses.dev.Modules() {
		b, err := ses.dev.Browser(name)
		if err != nil {
			return nil, err
		}
		sel := b.Root()
		sel.Constraints.AddConstraint("content", 0, 0, c)

		// TODO apply filter

		sels = append(sels, sel)
	}
	return sels, nil
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
		},
	}
}

func (ses *Session) handleGetConfig(req *RpcMsg, get *RpcGet, resp *RpcReply, c node.ContentConstraint) error {
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

func (ses *Session) handleRpc(rpc *RpcMsg) error {
	var err error
	resp := &RpcReply{MessageId: rpc.MessageId}
	if rpc.GetConfig != nil {
		err = ses.handleGetConfig(rpc, rpc.GetConfig, resp, node.ContentConfig)
	} else if rpc.Get != nil {
		err = ses.handleGetConfig(rpc, rpc.Get, resp, node.ContentOperational)
	} else if rpc.Close != nil {
		resp.OK = &Msg{}

		// TODO: properly close session
		// err = io.EOF
	} else {
		err = fmt.Errorf("unrecognized rpc command")
	}
	if err != nil {
		fmt.Printf("got err %s\n", err)
		return err
	}
	out := NewChunkedWtr(ses.out)
	defer out.Close()
	return WriteResponse(resp, out)
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
