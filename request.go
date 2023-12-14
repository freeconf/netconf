package netconf

import (
	"encoding/xml"
	"io"
)

type Request struct {
	Hello *HelloMsg
	Rpc   *RpcMsg
	Other *Msg
}

func (r Request) String() string {
	if r.Hello != nil {
		return "hello"
	}
	if r.Rpc != nil {
		return "rpc"
	}
	if r.Other != nil {
		return r.Other.XMLName.Local
	}
	return "unknown"
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
