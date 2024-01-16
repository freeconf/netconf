package netconf

import (
	"fmt"
	"time"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/patch/xml"
	"github.com/freeconf/yang/xpath"

	"github.com/freeconf/yang/nodeutil"
)

// data structures used in requests and responses

type Msg struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
	Elems   []*Msg     `xml:",any"`
}

type MsgLeaf struct {
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",chardata"`
}

type RpcReply struct {
	XMLName   xml.Name            `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc-reply"`
	MessageId string              `xml:"message-id,attr"`
	OK        *Msg                `xml:"ok,omitempty"`
	Data      *RpcData            `xml:"data,omitempty"`
	Out       []*nodeutil.XMLWtr2 `xml:",any"`
}

type RpcData struct {
	Nodes []*nodeutil.XMLWtr2
}

type HelloMsg struct {
	XMLName      xml.Name `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 hello"`
	Capabilities []*Msg   `xml:"capabilities>capability"`
	SessionId    string   `xml:"session-id,omitempty"`
}

type RpcMsg struct {
	XMLName            xml.Name            `xml:"urn:ietf:params:xml:ns:netconf:base:1.0 rpc"`
	MessageId          string              `xml:"message-id,attr"`
	Attrs              []xml.Attr          `xml:",any,attr"`
	GetConfig          *RpcGet             `xml:"get-config,omitempty"`
	Get                *RpcGet             `xml:"get,omitempty"`
	EditConfig         *RpcEdit            `xml:"edit-config,omitempty"`
	Copy               *RpcCopy            `xml:"copy-config,omitempty"`
	Delete             *RpcEdit            `xml:"delete-config,omitempty"`
	Close              *Msg                `xml:"close-session,omitempty"`
	Kill               *Msg                `xml:"kill-session,omitempty"`
	CreateSubscription *CreateSubscription `xml:"create-subscription,omitempty"`
	Action             *nodeutil.XmlNode   `xml:",any"`
}

type CreateSubscription struct {
	StartTime time.Time  `xml:"startTime,omitempty"`
	StopTime  time.Time  `xml:"stopTime,omitempty"`
	Filter    *RpcFilter `xml:"filter,omitempty"`
	Stream    string     `xml:"stream,omitempty"`
}

type RpcCopy struct {
	Source *RpcSource `xml:"source,omitempty"`
	Target *Msg       `xml:"target,omitempty"`
}

type RpcEdit struct {
	Target *Msg `xml:"target,omitempty"`

	// allowed: merge(default), replace, create, delete, remove
	DefaultOperation string `xml:"default-operation,omitempty"`

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
	Source *Msg       `xml:"source,omitempty"`
	Filter *RpcFilter `xml:"filter,omitempty"`
}

type RpcFilter struct {
	Type string `xml:"type,attr"`

	// when Type is "xpath"
	Select     string `xml:"select,omitempty"`
	shortcodes map[string]string

	Elems []*Msg `xml:",any"`
}

type Notification struct {
	EventTime time.Time        `xml:"eventTime"`
	Event     nodeutil.XmlNode `xml:"event"`
}

func addNamespaces(ns map[string]*meta.Module, m *meta.Module) {
	ns[m.Namespace()] = m
	for _, child := range m.Imports() {
		addNamespaces(ns, child.Module())
	}
}

func deviceNamespaces(d device.Device, shortcodes map[string]string) xpath.ShortcodeToModule {
	namespaces := make(map[string]*meta.Module)
	for _, m := range d.Modules() {
		addNamespaces(namespaces, m)
	}
	return func(shortcode string) (*meta.Module, error) {
		ns, found := shortcodes[shortcode]
		if !found {
			return nil, fmt.Errorf("namespace for shortcode '%s' not found", shortcode)
		}
		mod, found := namespaces[ns]
		if !found {
			return nil, fmt.Errorf("module for namespace '%s' not found", ns)
		}
		return mod, nil
	}
}

func (f *RpcFilter) CompileXPath(d device.Device) (*node.Selection, error) {
	if f.Type == "xpath" && f.Select != "" {
		lookup := deviceNamespaces(d, f.shortcodes)
		top, err := xpath.Parse2(lookup, f.Select)
		if err != nil {
			return nil, err
		}
		b, err := d.Browser(top.Ident)
		if err != nil {
			return nil, err
		}
		root := b.Root()
		if top.Next == nil {
			return root, nil
		}
		return root.XFind(top.Next)
	}
	return nil, fmt.Errorf("not a valid xpath filter")
}

func (rf *RpcFilter) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	copy := struct {
		Type   string `xml:"type,attr"`
		Select string `xml:"select,omitempty"`
		Elems  []*Msg `xml:",any"`
	}{}
	if err := d.DecodeElement(&copy, &start); err != nil {
		return err
	}
	rf.Elems = copy.Elems
	rf.Type = copy.Type
	rf.Select = copy.Select
	return nil
}
