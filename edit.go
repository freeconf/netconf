package netconf

import (
	"fmt"
	"strings"

	"github.com/freeconf/yang/meta"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/patch/xml"
	"github.com/freeconf/yang/val"
)

type edit struct {
	path string
	op   string
	n    *nodeutil.XmlNode
}

type editBuilder struct {
	edits []edit
}

func (b *editBuilder) enter(base *nodeutil.XmlNode) node.Node {
	return &nodeutil.Extend{
		Base: base,
		OnChild: func(parent node.Node, r node.ChildRequest) (node.Node, error) {
			n, err := parent.Child(r)
			if n == nil || err != nil {
				return n, err
			}
			from := n.(*nodeutil.XmlNode)
			if op := getOp(from.Attr); op != "" {
				b.edits = append(b.edits, edit{
					op:   op,
					path: r.Selection.Path.StringNoModule(),
					n:    from,
				})
			}
			return b.enter(from), nil
		},
		OnNext: func(parent node.Node, r node.ListRequest) (node.Node, []val.Value, error) {
			n, k, err := parent.Next(r)
			if n == nil || err != nil {
				return n, k, err
			}
			from := n.(*nodeutil.XmlNode)
			if op := getOp(from.Attr); op != "" {
				p := r.Selection.Path
				if k != nil {
					// this distinguished operations on the list item and not the
					// list itself
					p = &node.Path{
						Parent: p.Parent,
						Meta:   p.Meta,
						Key:    k,
					}
				}
				b.edits = append(b.edits, edit{
					op:   op,
					path: p.StringNoModule(),
					n:    from,
				})
			}
			return b.enter(from), k, nil
		},
		OnField: func(parent node.Node, r node.FieldRequest, hnd *node.ValueHandle) error {
			err := parent.Field(r, hnd)
			if hnd.Val == nil || err != nil {
				return err
			}
			from := parent.(*nodeutil.XmlNode)
			ndx := from.Find(0, r.Meta)

			// TODO: This won't work with leaf-lists if first leaf list
			// isn't the element w/the operaton on it
			fieldElem := from.Nodes[ndx]

			if op := getOp(fieldElem.Attr); op != "" {
				containerPath := r.Selection.Path.StringNoModule()
				if containerPath != "" {
					containerPath = containerPath + "/"
				}
				b.edits = append(b.edits, edit{
					op:   op,
					n:    from,
					path: containerPath + r.Meta.Ident(),
				})
			}
			return nil
		},
	}
}

func buildEdits(defaultOp string, data *nodeutil.XmlNode, m *meta.Module) ([]edit, error) {
	var bldr editBuilder

	b := node.NewBrowser(m, bldr.enter(data))
	if err := b.Root().InsertInto(nodeutil.Null()); err != nil {
		return nil, err
	}
	edits := bldr.edits
	// check for edits that conflict w/eachother.  If there's an edit under another edit
	// then the edit before it would path that starts w/conflicting edit path
	for i := 1; i < len(edits); i++ {
		if strings.HasPrefix(edits[i].path, edits[i-1].path) {
			return nil, fmt.Errorf("edit %s:%s conflicts with edit %s:%s",
				edits[i].op, edits[i].path,
				edits[i-1].op, edits[i-1].path)
		}
	}

	if len(edits) == 0 {
		edits = []edit{
			{
				op:   defaultOp,
				path: "",
				n:    data,
			},
		}
	}
	return edits, nil
}

func getOp(attrs []xml.Attr) string {
	for _, a := range attrs {
		if a.Name.Local == "operation" && a.Name.Space == Base_1_1 {
			return a.Value
		}
	}
	return ""
}
