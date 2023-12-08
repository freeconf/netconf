package netconf

import (
	"encoding/xml"

	"github.com/freeconf/yang/nodeutil"
)

// data structures used in requests and responses

type Msg struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:"-"`
	Content string     `xml:",chardata"`
	Elems   []*Msg     `xml:",any"`
}

type MsgLeaf struct {
	Attrs   []xml.Attr `xml:"-"`
	Content string     `xml:",innerxml"`
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
	XMLName   xml.Name   `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc"`
	MessageId string     `xml:"message-id,attr"`
	Attrs     []xml.Attr `xml:"-"`
	GetConfig *RpcGet    `xml:"get-config,omitempty"`
	Get       *RpcGet    `xml:"get,omitempty"`
	Edit      *RpcEdit   `xml:"edit-config,omitempty"`
	Copy      *RpcCopy   `xml:"copy-config,omitempty"`
	Delete    *RpcEdit   `xml:"delete-config,omitempty"`
	Close     *Msg       `xml:"close-session,omitempty"`
	Kill      *Msg       `xml:"kill-session,omitempty"`
}

type RpcCopy struct {
	Source *RpcSource `xml:"source,omitempty"`
	Target *Msg       `xml:"target,omitempty"`
}

type RpcEdit struct {
	// allowed: merge(default), replace, create, delete, remove
	Operation string `xml:"operation,attr"`

	Target *Msg `xml:"target,omitempty"`

	// allowed: merge(default), replace, none
	DefaultOperation *Msg `xml:"default-operation,omitempty"`

	// allowed: test-then-set(default), set, test-only
	TestOperation *Msg `xml:"test-operation,omitempty"`

	// allowed: stop-on-error(default), continue-on-error, rollback-on-error
	ErrorOption *Msg `xml:"error-option,omitempty"`

	Config *nodeutil.XmlNode `xml:"config"`
	Elem   *Msg              `xml:",any"`
}

type RpcSource struct {
	Url     *MsgLeaf `xml:"url,omitempty"`
	Content string   `xml:",innerxml"`
}

type RpcGet struct {
	Source *MsgLeaf   `xml:"source,omitempty"`
	Filter *RpcFilter `xml:"filter,omitempty"`
}

type RpcFilter struct {
	Type  string `xml:"type,attr"`
	Elems []*Msg `xml:",any"`
}
