package netconf

import (
	"fmt"
	"io"
	"strconv"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

type Session struct {
	dev device.Device
	mgr SessionManager
	out io.Writer
	in  io.Reader
	Id  int64
}

func NewSession(mgr SessionManager, dev device.Device, in io.Reader, out io.Writer) *Session {
	return &Session{
		mgr: mgr,
		dev: dev,
		Id:  mgr.NextSessionId(),
		in:  in,
		out: out,
	}
}

func (ses *Session) readMessages() error {
	for {
		err := ses.readRequest()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("ending stream")
				return nil
			}
			fmt.Printf("err reading xml message. err=%s\n", err)
			return err
		}
	}
}

func (ses *Session) readRequest() error {
	req, err := DecodeRequest(ses.in)
	if err != nil {
		return err
	}
	if req.Rpc != nil {
		return ses.handleRpc(req.Rpc)
	}
	if req.Hello != nil {
		return ses.handleHello(req.Hello)
	}
	switch req.Other.XMLName.Local {
	case "close":
		// ignore write err, we were just being polite by returning ok
		WriteResponse(RpcReply{OK: &Msg{}}, ses.out, true)
		return io.EOF
	}
	return fmt.Errorf("unsupported message")
}

func (ses *Session) readFilter(f *RpcFilter, c node.ContentConstraint) ([]*node.Selection, error) {
	var sels = make([]*node.Selection, 0, len(ses.dev.Modules()))
	for name := range ses.dev.Modules() {
		b, err := ses.dev.Browser(name)
		if err != nil {
			return nil, err
		}
		sels = append(sels, b.Root())
	}
	return sels, nil
}

func (ses *Session) Hello() *HelloMsg {
	return &HelloMsg{
		SessionId: strconv.FormatInt(ses.Id, 10),
		Capabilities: []*Msg{
			{Content: []byte("urn:ietf:params:netconf:base:1.1")},
		},
	}
}

func (ses *Session) handleGetConfig(rpc *RpcMsg, get *RpcGet, c node.ContentConstraint) error {
	sels, err := ses.readFilter(get.Filter, c)
	if err != nil {
		return err
	}
	resp := &RpcReply{}
	for _, sel := range sels {
		cfg := &nodeutil.XMLWtr2{}
		if err := sel.UpsertInto(cfg); err != nil {
			return err
		}
		resp.Config = append(resp.Config, cfg)
	}
	return WriteResponse(resp, ses.out, true)
}

func (ses *Session) handleRpc(rpc *RpcMsg) error {
	if rpc.GetConfig != nil {
		return ses.handleGetConfig(rpc, rpc.GetConfig, node.ContentConfig)
	}
	if rpc.Get != nil {
		return ses.handleGetConfig(rpc, rpc.Get, node.ContentOperational)
	}
	return fmt.Errorf("unsupported rpc command")
}

func (ses *Session) handleHello(h *HelloMsg) error {
	return nil
}
