package netconf

import (
	"context"
	"encoding/xml"
	"strings"

	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/val"
)

type filterContextKeyType struct{}

var filterContextKey = filterContextKeyType{}

type filter struct {
	containment map[xml.Name]*filter
	selection   []xml.Name
	matching    []contentMatching
}

var filterSelectNone = &filter{}

func (f *filter) selectAll() bool {
	return len(f.containment) == 0 && len(f.selection) == 0 && len(f.matching) == 0
}

func (f *filter) selected(ident string) (*filter, bool) {
	if f == filterSelectNone {
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
			return &filter{}, true
		}
	}
	for _, m := range f.matching {
		if ident == m.field.Local {
			return f, true
		}
	}
	return filterSelectNone, false
}

type contentMatching struct {
	field xml.Name
	value string
}

func compileFilter(x *Msg, f *filter) error {
	for _, e := range x.Elems {
		if err := compileFilterComponents(e, f); err != nil {
			return err
		}
	}
	return nil
}

func compileFilterComponents(x *Msg, f *filter) error {
	if len(x.Elems) > 0 || len(x.Attrs) > 0 {
		child := &filter{}
		if f.containment == nil {
			f.containment = make(map[xml.Name]*filter)
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
				if err := compileFilterComponents(e, child); err != nil {
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

func (root *filter) CheckContainerPreConstraints(r *node.ChildRequest) (bool, error) {
	f := root.currentFilter(r.Selection)
	_, selected := f.selected(r.Meta.Ident())
	return selected, nil
}

func (root *filter) CheckFieldPreConstraints(r *node.FieldRequest, hnd *node.ValueHandle) (bool, error) {
	f := root.currentFilter(r.Selection)
	_, selected := f.selected(r.Meta.Ident())
	return selected, nil
}

func (root *filter) CheckListPostConstraints(r node.ListRequest, child *node.Selection, key []val.Value) (bool, bool, error) {
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

func (root *filter) currentFilter(s *node.Selection) *filter {
	if found := s.Context.Value(filterContextKey); found != nil {
		return found.(*filter)
	}
	return root
}

func (root *filter) ContextConstraint(s *node.Selection) context.Context {
	if s.InsideList {
		return s.Context
	}
	f := root.currentFilter(s)
	ctx := s.Context
	next, _ := f.selected(s.Meta().Ident())
	return context.WithValue(ctx, filterContextKey, next)
}
