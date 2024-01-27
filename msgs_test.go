package netconf

import (
	"bytes"
	"flag"
	"strings"
	"testing"
	"time"

	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/patch/xml"
)

var updateFlag = flag.Bool("update", false, "update golden files instead of verifying against them")

func TestMsg(t *testing.T) {
	msg := &Msg{
		XMLName: xml.Name{Local: "x", Space: "Y"},
		Elems: []*Msg{
			{
				XMLName: xml.Name{Local: "y"},
			},
		},
	}
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg, &buf, false, false))
	fc.AssertEqual(t, `<x xmlns="Y"><y></y></x>`, buf.String())
}

func TestGetConfig(t *testing.T) {
	payload := `
	<rpc message-id="101" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<get-config>
			<source><running/></source>
			<filter type="subtree">
				<top xmlns="http://example.com/schema/1.2/config">
					<users />
				</top>
			</filter>
		</get-config>
	</rpc>`

	// read/decode
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Rpc != nil)
	fc.RequireEqual(t, true, msg.Rpc.GetConfig != nil)
	fc.AssertEqual(t, "101", msg.Rpc.MessageId)
	fc.AssertEqual(t, true, msg.Rpc.GetConfig.Filter != nil)
	fc.AssertEqual(t, "subtree", msg.Rpc.GetConfig.Filter.Type)
	fc.AssertEqual(t, 1, len(msg.Rpc.GetConfig.Filter.Elems))

	// only nec. when trying to write back what we read. RFC explicitly states
	// the whitespace might matter.
	CleanWhitespace(msg.Rpc.GetConfig.Source.Elems...)
	CleanWhitespace(msg.Rpc.GetConfig.Filter.Elems...)

	// write/encode
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg.Rpc, &buf, false, true))
	fc.Gold(t, *updateFlag, buf.Bytes(), "testdata/gold/get-config.xml")
}

func CleanWhitespace(elems ...*Msg) {
	for _, x := range elems {
		x.Content = strings.TrimSpace(x.Content)
		x.XMLName.Space = ""
		CleanWhitespace(x.Elems...)
	}
}

func TestAction(t *testing.T) {
	payload := `
	<rpc message-id="101" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
	  <rock-the-house xmlns="urn:example:rock">
	     <zip-code>27606-0100</zip-code>
      </rock-the-house>	
	</rpc>`

	// read/decode
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Rpc != nil)
	fc.RequireEqual(t, true, msg.Rpc.Action != nil)
	fc.AssertEqual(t, "rock-the-house", msg.Rpc.Action.XMLName.Local)
}

func TestHello(t *testing.T) {
	payload := `
	<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<capabilities>
			<capability>urn:ietf:params:netconf:base:1.1</capability>
			<capability>urn:ietf:params:netconf:capability:startup:1.0</capability>
			<capability>http://example.net/router/2.3/myfeature</capability>
		</capabilities>
		<session-id>4</session-id>
	</hello>`

	// read/decode
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Hello != nil)
	fc.RequireEqual(t, "4", msg.Hello.SessionId)
	fc.RequireEqual(t, 3, len(msg.Hello.Capabilities))
	fc.RequireEqual(t, "http://example.net/router/2.3/myfeature", strings.TrimSpace(string(msg.Hello.Capabilities[2].Content)))

	// write/encode
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg.Hello, &buf, false, true))
	fc.Gold(t, *updateFlag, buf.Bytes(), "testdata/gold/hello.xml")
}

func TestCreateSupscription(t *testing.T) {
	payload := `
	<rpc message-id="101" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<create-subscription xmlns="urn:ietf:params:xml:ns:netconf:notification:1.0">
			<stream>car:update</stream>
		</create-subscription>
	</rpc>`

	// read/decode
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Rpc != nil)
	fc.RequireEqual(t, true, msg.Rpc.CreateSubscription != nil)
	fc.AssertEqual(t, "car:update", msg.Rpc.CreateSubscription.Stream)

	// write/encode
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg.Rpc, &buf, false, true))
	fc.Gold(t, *updateFlag, buf.Bytes(), "testdata/gold/create-subscription.xml")
}

func TestNotification(t *testing.T) {
	// any fixed date will do
	t0, _ := time.Parse("2006-01-02T15:04:05Z", "2014-01-01T12:00:00Z")
	resp := Notification{
		EventTime: t0,
		Elems: []*nodeutil.XMLWtr2{
			{
				XMLName: xml.Name{Space: "x", Local: "y"},
				Content: "hello",
			},
		},
	}
	var actual bytes.Buffer
	err := WriteResponse(resp, &actual)
	fc.AssertEqual(t, nil, err)
	fc.Gold(t, *updateFlag, actual.Bytes(), "testdata/gold/notification.xml")
}
