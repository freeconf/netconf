package netconf

import (
	"strings"
	"testing"

	"github.com/freeconf/yang/fc"
)

func TestDecodeGetConfig(t *testing.T) {
	payload := `
	<rpc message-id="101" xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
		<get-config>
			<source>
				<running/>
			</source>
			<filter type="subtree">
				<top xmlns="http://example.com/schema/1.2/config">
					<users/>
				</top>
			</filter>
		</get-config>
	</rpc>`
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Rpc != nil)
	fc.RequireEqual(t, true, msg.Rpc.GetConfig != nil)
	fc.AssertEqual(t, "101", msg.Rpc.MessageId)
	fc.AssertEqual(t, true, msg.Rpc.GetConfig.Filter != nil)
	fc.AssertEqual(t, "subtree", msg.Rpc.GetConfig.Filter.Type)
	fc.AssertEqual(t, 1, len(msg.Rpc.GetConfig.Filter.Elems))
}

func TestDecodeHello(t *testing.T) {
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
	msg, err := DecodeRequest(strings.NewReader(payload))
	fc.RequireEqual(t, nil, err)
	fc.RequireEqual(t, true, msg.Hello != nil)
	fc.RequireEqual(t, "4", msg.Hello.SessionId)
	fc.RequireEqual(t, 3, len(msg.Hello.Capabilities))
	fc.RequireEqual(t, "http://example.net/router/2.3/myfeature", strings.TrimSpace(string(msg.Hello.Capabilities[2].Content)))
}
