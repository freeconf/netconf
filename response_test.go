package netconf

import (
	"testing"

	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/patch/xml"
)

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
			msg:      &RpcReply{MessageId: "abc", OK: &Msg{}},
			expected: `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="abc"><ok></ok></rpc-reply>`,
		},
		{
			msg:      &RpcReply{MessageId: "abc", Out: []*nodeutil.XMLWtr2{{XMLName: xml.Name{Local: "foo", Space: "bar"}}}},
			expected: `<rpc-reply xmlns="urn:ietf:params:xml:ns:netconf:base:1.0" message-id="abc"><foo xmlns="bar"></foo></rpc-reply>`,
		},
	}
	for _, test := range tests {
		actual, err := xml.Marshal(test.msg)
		fc.AssertEqual(t, nil, err)
		fc.AssertEqual(t, test.expected, string(actual))
	}
}
