package netconf

import (
	"context"
	"strings"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/patch/xml"
	"github.com/freeconf/yang/val"
)

type subtreeContextKeyType struct{}

var subtreeContextKey = subtreeContextKeyType{}

type subtreeFilter struct {
	containment map[xml.Name]*subtreeFilter
	selection   []xml.Name
	matching    []contentMatching
}

var subtreeSelectNone = &subtreeFilter{}

func (f *subtreeFilter) selectAll() bool {
	return len(f.containment) == 0 && len(f.selection) == 0 && len(f.matching) == 0
}

func (f *subtreeFilter) selected(ident string) (*subtreeFilter, bool) {
	if f == subtreeSelectNone {
		return f, false
	}
	if f.selectAll() {
		return f, true
	}
	if c, found := f.containment[xml.Name{Local: ident}]; found {
		return c, true
	}
	for _, s := range f.selection {
		if ident == s.Local {
			return &subtreeFilter{}, true
		}
	}
	for _, m := range f.matching {
		if ident == m.field.Local {
			return f, true
		}
	}
	return subtreeSelectNone, false
}

type contentMatching struct {
	field xml.Name
	value string
}

func compileSubtree(x *Msg, f *subtreeFilter) error {
	for _, e := range x.Elems {
		if err := compileSubtreeComponents(e, f); err != nil {
			return err
		}
	}
	return nil
}

func compileSubtreeComponents(x *Msg, f *subtreeFilter) error {
	if len(x.Elems) > 0 || len(x.Attrs) > 0 {
		child := &subtreeFilter{}
		if f.containment == nil {
			f.containment = make(map[xml.Name]*subtreeFilter)
		}
		f.containment[x.XMLName] = child
		if len(x.Attrs) > 0 {
			for _, a := range x.Attrs {
				child.matching = append(child.matching, contentMatching{
					field: a.Name,
					value: a.Value,
				})
			}
		} else {
			for _, e := range x.Elems {
				if err := compileSubtreeComponents(e, child); err != nil {
					return err
				}
			}
		}

		return nil
	}
	s := strings.TrimSpace(x.Content)
	if s != "" {
		f.matching = append(f.matching, contentMatching{
			field: x.XMLName,
			value: s,
		})
		return nil
	}
	f.selection = append(f.selection, x.XMLName)
	return nil
}

func (root *subtreeFilter) CheckContainerPreConstraints(r *node.ChildRequest) (bool, error) {
	f := root.currentFilter(r.Selection)
	_, selected := f.selected(r.Meta.Ident())
	return selected, nil
}

func (root *subtreeFilter) CheckFieldPreConstraints(r *node.FieldRequest, hnd *node.ValueHandle) (bool, error) {
	f := root.currentFilter(r.Selection)
	_, selected := f.selected(r.Meta.Ident())
	return selected, nil
}

func (root *subtreeFilter) CheckListPostConstraints(r node.ListRequest, child *node.Selection, key []val.Value) (bool, bool, error) {
	f := root.currentFilter(r.Selection)
	if f.selectAll() {
		return true, true, nil
	}
	for _, m := range f.matching {
		fs, err := child.Find(m.field.Local)
		if err != nil {
			return true, false, err
		}
		v, err := fs.Get()
		if err != nil {
			return true, false, err
		}
		if v.String() != m.value {
			return true, false, nil
		}
	}
	return true, true, nil
}

func (root *subtreeFilter) currentFilter(s *node.Selection) *subtreeFilter {
	if found := s.Context.Value(subtreeContextKey); found != nil {
		return found.(*subtreeFilter)
	}
	return root
}

func (root *subtreeFilter) ContextConstraint(s *node.Selection) context.Context {
	if s.InsideList {
		return s.Context
	}
	f := root.currentFilter(s)
	ctx := s.Context
	next, _ := f.selected(s.Meta().Ident())
	return context.WithValue(ctx, subtreeContextKey, next)
}
