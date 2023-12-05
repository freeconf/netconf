package netconf

import (
	"encoding/xml"
	"testing"

	"github.com/freeconf/yang/fc"
)

// func TestRpcResponse(t *testing.T) {
// 	var buf bytes.Buffer
// 	r := ResponseWriter{Out: &buf}
// 	rpc := &RpcMsg{}
// 	r.openRpcReplyContent(rpc, "x")
// 	r.closeRpcReplyContent(rpc, "x")
// 	fc.AssertEqual(t, `<rpc-reply><x></x></rpc-reply>`, buf.String())
// }

func TestXmlWriter(t *testing.T) {
	tests := []struct {
		msg      any
		expected string
	}{
		{
			msg: &HelloMsg{
				Capabilities: []*Msg{
					{Content: "xyz"},
				},
				SessionId: "99",
			},
			expected: `<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0"><capabilities><capability>xyz</capability></capabilities><session-id>99</session-id></hello>`,
		},
		{
			msg:      &RpcReply{OK: &Msg{}, MessageId: "abc"},
			expected: `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="abc"><ok></ok></rpc-reply>`,
		},
	}
	for _, test := range tests {
		actual, err := xml.Marshal(test.msg)
		fc.AssertEqual(t, nil, err)
		fc.AssertEqual(t, test.expected, string(actual))
	}
}
