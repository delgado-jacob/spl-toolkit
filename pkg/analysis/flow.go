package analysis

import "sort"

type trackedField struct {
	FieldBinding
	source bool
}
type environment struct {
	rewrite         *rewriteFlow
	fields          map[string]trackedField
	removed         map[string]bool
	open, uncertain bool
}

func newEnvironment() *environment {
	return &environment{fields: map[string]trackedField{}, removed: map[string]bool{}, open: true}
}
func copyIDs(ids []string) []string { return append([]string{}, ids...) }
func (e *environment) snapshot() FieldState {
	s := FieldState{Fields: []FieldBinding{}, Removed: []string{}, Open: e.open, Uncertain: e.uncertain}
	for _, f := range e.fields {
		f.OriginReferenceIDs = copyIDs(f.OriginReferenceIDs)
		s.Fields = append(s.Fields, f.FieldBinding)
	}
	for n := range e.removed {
		s.Removed = append(s.Removed, n)
	}
	sort.Slice(s.Fields, func(i, j int) bool { return s.Fields[i].Name < s.Fields[j].Name })
	sort.Strings(s.Removed)
	return s
}
func (e *environment) clone() *environment {
	n := newEnvironment()
	n.rewrite = e.rewrite.clone()
	n.open = e.open
	n.uncertain = e.uncertain
	for k, v := range e.fields {
		v.OriginReferenceIDs = copyIDs(v.OriginReferenceIDs)
		n.fields[k] = v
	}
	for k, v := range e.removed {
		n.removed[k] = v
	}
	return n
}
func (e *environment) remove(name string) {
	e.rewriteInvalidate(name)
	delete(e.fields, name)
	e.removed[name] = true
}
func (e *environment) install(name string, ids []string, conditional bool) {
	e.rewriteInvalidate(name)
	e.fields[name] = trackedField{FieldBinding: FieldBinding{Name: name, OriginReferenceIDs: copyIDs(ids), Conditional: conditional}}
	delete(e.removed, name)
}
func uniqueIDs(groups ...[]string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, g := range groups {
		for _, id := range g {
			if !seen[id] {
				out = append(out, id)
				seen[id] = true
			}
		}
	}
	return out
}
