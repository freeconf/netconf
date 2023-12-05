package netconf

import (
	"encoding/xml"
	"io"

	"github.com/freeconf/yang/nodeutil"
)

type Request struct {
	Hello *HelloMsg
	Rpc   *RpcMsg
	Other *Msg
}

func DecodeRequest(in io.Reader) (*Request, error) {
	dec := xml.NewDecoder(in)
	var req Request
	err := dec.Decode(&req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

type Msg struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content string     `xml:",innerxml"`
	Elems   []*Msg     `xml:",any"`
}

func (m *Request) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	switch start.Name.Local {
	case "rpc":
		m.Rpc = &RpcMsg{}
		return d.DecodeElement(m.Rpc, &start)
	case "hello":
		m.Hello = &HelloMsg{}
		return d.DecodeElement(m.Hello, &start)
	}
	m.Other = &Msg{}
	return d.DecodeElement(m.Other, &start)
}

type RpcReply struct {
	XMLName   xml.Name            `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageId string              `xml:"message-id,attr"`
	OK        *Msg                `xml:"ok,omitempty"`
	Config    []*nodeutil.XMLWtr2 `xml:"data,omitempty"`
}

type HelloMsg struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	Capabilities []*Msg   `xml:"capabilities>capability"`
	SessionId    string   `xml:"session-id,omitempty"`
}

type RpcMsg struct {
	XMLName   xml.Name   `xml:"rpc"`
	MessageId string     `xml:"message-id,attr"`
	Attrs     []xml.Attr `xml:"-"`
	GetConfig *RpcGet    `xml:"get-config,omitempty"`
	Get       *RpcGet    `xml:"get,omitempty"`
	Close     *Msg       `xml:"close-session,omitempty"`
}

type RpcSource struct {
	XMLName xml.Name `xml:"source"`
	Elem    *Msg     `xml:",any"`
}

type RpcGet struct {
	Filter *RpcFilter `xml:"filter,omitempty"`
}

type RpcFilter struct {
	XMLName xml.Name `xml:"filter"`
	Type    string   `xml:"type,attr"`
	Elems   []*Msg   `xml:",any"`
}
