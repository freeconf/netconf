package netconf

import (
	"strings"
	"testing"

	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/parser"
	"github.com/freeconf/yang/source"
)

func TestBuildEdits(t *testing.T) {
	tests := []struct {
		editStr  string
		expected []edit
	}{
		{
			editStr: `
				<car xmlns:nc="urn:ietf:params:netconf:base:1.1">
					<speed nc:operation="merge">10</speed>
				</car>
			`,
			expected: []edit{
				{
					path: "speed",
					op:   "merge",
				},
			},
		},
		{
			editStr: `
				<car xmlns:nc="urn:ietf:params:netconf:base:1.1">
				    <tire>
					   <pos>0</pos>
					   <size nc:operation="delete">16</size>
					</tire>
					<speed nc:operation="merge">10</speed>
				</car>
			`,
			expected: []edit{
				{
					path: "tire=0/size",
					op:   "delete",
				},
				{
					path: "speed",
					op:   "merge",
				},
			},
		},
		{
			editStr: `
				<car xmlns:nc="urn:ietf:params:netconf:base:1.1">
				    <tire nc:operation="delete">
					   <pos>0</pos>
					</tire>
					<speed nc:operation="merge">10</speed>
				</car>
			`,
			expected: []edit{
				{
					path: "tire=0",
					op:   "delete",
				},
				{
					path: "speed",
					op:   "merge",
				},
			},
		},
	}
	for _, test := range tests {
		m := parser.RequireModule(source.Dir("./testdata/yang"), "car")
		n, err := nodeutil.ReadXMLDoc(strings.NewReader(test.editStr))
		fc.RequireEqual(t, nil, err)
		edits, err := buildEdits("merge", n, m)
		fc.AssertEqual(t, nil, err)
		for i := range edits {
			fc.AssertEqual(t, true, edits[i].n != nil)
			edits[i].n = nil
		}
		fc.AssertEqual(t, test.expected, edits)
	}
}
