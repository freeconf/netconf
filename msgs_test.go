package netconf

import (
	"bytes"
	"encoding/xml"
	"flag"
	"strings"
	"testing"

	"github.com/freeconf/yang/fc"
)

var updateFlag = flag.Bool("update", false, "update golden files instead of verifying against them")

func TestMsg(t *testing.T) {
	msg := &Msg{
		XMLName: xml.Name{Local: "x", Space: "Y"},
		Elems: []*Msg{
			{
				XMLName: xml.Name{Local: "y", Space: "Y"},
			},
		},
	}
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg, &buf, false, true))
	fc.AssertEqual(t, `<x></x>`, buf.String())

}

func TestGetConfig(t *testing.T) {
	payload := `
	<rpc message-id="101" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<get-config>
			<source>
				<running/>
			</source>
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
	CleanWhitespace(msg.Rpc.GetConfig.Filter.Elems...)

	// write/encode
	var buf bytes.Buffer
	fc.AssertEqual(t, nil, WriteResponseWithOptions(msg.Rpc, &buf, false, true))
	fc.Gold(t, *updateFlag, buf.Bytes(), "testdata/gold/get-config.xml")
}

func CleanWhitespace(elems ...*Msg) {
	for _, x := range elems {
		x.Content = strings.TrimSpace(x.Content)
		CleanWhitespace(x.Elems...)
	}
}

func TestHello(t *testing.T) {
	payload := `
	<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<capabilities>
			<capability>
				urn:ietf:params:netconf:base:1.1
			</capability>
			<capability>
				urn:ietf:params:netconf:capability:startup:1.0
			</capability>
			<capability>
				http://example.net/router/2.3/myfeature
			</capability>
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
