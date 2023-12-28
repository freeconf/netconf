package netconf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/parser"
	"github.com/freeconf/yang/patch/xml"
	"github.com/freeconf/yang/source"
)

func TestSubtreeCompile(t *testing.T) {
	tests := []struct {
		filter   string
		expected *subtreeFilter
	}{
		{
			filter: `<x><y /></x>`,
			expected: &subtreeFilter{
				containment: map[xml.Name]*subtreeFilter{
					{Local: "x"}: {
						selection: []xml.Name{
							{Local: "y"},
						},
					},
				},
			},
		},
		{
			filter: `<x><y>burrito</y></x>`,
			expected: &subtreeFilter{
				containment: map[xml.Name]*subtreeFilter{
					{Local: "x"}: {
						matching: []contentMatching{
							{
								field: xml.Name{Local: "y"},
								value: "burrito",
							},
						},
					},
				},
			},
		},
		{
			filter: `<x><z/><y>burrito</y></x>`,
			expected: &subtreeFilter{
				containment: map[xml.Name]*subtreeFilter{
					{Local: "x"}: {
						selection: []xml.Name{
							{Local: "z"},
						},
						matching: []contentMatching{
							{
								field: xml.Name{Local: "y"},
								value: "burrito",
							},
						},
					},
				},
			},
		},
		{
			filter: `<x><z/><y>burrito</y><q abc="123"/></x>`,
			expected: &subtreeFilter{
				containment: map[xml.Name]*subtreeFilter{
					{Local: "x"}: {
						selection: []xml.Name{
							{Local: "z"},
						},
						matching: []contentMatching{
							{
								field: xml.Name{Local: "y"},
								value: "burrito",
							},
						},
						containment: map[xml.Name]*subtreeFilter{
							{Local: "q"}: {
								matching: []contentMatching{
									{
										field: xml.Name{Local: "abc"},
										value: "123",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, test := range tests {
		var xf Msg
		d := xml.NewDecoder(strings.NewReader(test.filter))
		fc.RequireEqual(t, nil, d.Decode(&xf))
		var f subtreeFilter
		fc.AssertEqual(t, nil, compileSubtreeComponents(&xf, &f))
		var actual bytes.Buffer
		dumpTestFilter(&actual, &f, "")
		fc.AssertEqual(t, test.expected, &f, actual.String())
	}
}

func newTestData() *node.Browser {
	data := `{
		"users" : {
			"user" : [{
				"name": "root",
				"type": "superuser",
				"full-name": "Charlie Root",
				"company-info" : {
					"dept" : 1,
					"id" : 1
				}	
			},{
				"name": "fred",
				"type": "admin",
				"full-name": "Fred Flintstone",
				"company-info" : {
					"dept" : 2,
					"id" : 2
				}	
			},{
				"name": "barney",
				"type": "admin",
				"full-name": "Barney Rubble",
				"company-info" : {
					"dept" : 2,
					"id" : 3
				}	
			}]
		},
		"machines": [{
			"id" : "00001",
			"os" : "linux",
			"purpose" : "webserver"
		},{
			"id" : "00002",
			"os" : "linux",
			"purpose" : "dns"
		}]
	}`
	n := nodeutil.ReadJSON(data)
	m := parser.RequireModule(source.Dir("./testdata/yang"), "top")
	return node.NewBrowser(m, n)
}

func TestSubtree(t *testing.T) {
	b := newTestData()
	tests := []struct {
		desc     string
		filter   string
		expected string
	}{
		{
			desc:     "single containment",
			filter:   `<top><users/></top>`,
			expected: "testdata/filter/gold/users.xml",
		},
		{
			desc:     "single but different containment",
			filter:   `<top><machines/></top>`,
			expected: "testdata/filter/gold/machines.xml",
		},
		{
			desc: "nested containment, unrelated containment",
			filter: `<top>
				<bogus/>
				<users>
					<user/>
				</users>
			</top>`,
			expected: "testdata/filter/gold/users.xml",
		},
		{
			desc: "containment, then selection",
			filter: `<top>
			  <users>
			    <user>
				  <name/>
				</user>
			  </users>
			</top>`,
			expected: "testdata/filter/gold/user-names.xml",
		},
		{
			desc: "nested containment, with single matching",
			filter: `<top>
			  <users>
			    <user>
				  <name>fred</name>
				</user>
			  </users>
			</top>`,
			expected: "testdata/filter/gold/fred.xml",
		},
		{
			desc: "nested containment, with double matching",
			filter: `<top>
		      <users>
			    <user>
				  <name>fred</name>
				  <type>admin</type>
				</user>
			  </users>
			</top>`,
			expected: "testdata/filter/gold/fred-admin.xml",
		},
		{
			desc: "nested containment, selection and matching",
			filter: `<top>
			  <users>
			    <user>
				  <name>fred</name>
				  <type/>
				</user>
		       </users>
			</top>`,
			expected: "testdata/filter/gold/fred-admin.xml",
		},
	}
	for _, test := range tests {
		t.Log(test.desc)
		sel := b.Root()
		f := compileTestFilter(t, test.filter)
		sel.Constraints.AddConstraint("netconf", 0, 0, f)
		actual, err := nodeutil.WriteXMLDoc(sel, true)
		fc.AssertEqual(t, nil, err)
		fc.Gold(t, *updateFlag, []byte(actual), test.expected)
	}
}

func compileTestFilter(t *testing.T, fstr string) *subtreeFilter {
	d := xml.NewDecoder(strings.NewReader(fstr))
	var xf Msg
	fc.RequireEqual(t, nil, d.Decode(&xf))
	var f subtreeFilter
	fc.RequireEqual(t, nil, compileSubtree(&xf, &f))
	return &f
}

func dumpTestFilter(w io.Writer, f *subtreeFilter, indent string) {
	for _, m := range f.matching {
		fmt.Fprintf(w, "%s  matching:%s=%s\n", indent, m.field, m.value)
	}
	for _, s := range f.selection {
		fmt.Fprintf(w, "%s  selection:%s\n", indent, s)
	}
	for elem, child := range f.containment {
		fmt.Fprintf(w, "%scontainment:%s [\n", indent, elem)
		dumpTestFilter(w, child, indent+"  ")
		fmt.Fprintf(w, "%s]\n", indent)
	}
}
