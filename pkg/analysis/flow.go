package analysis

import "sort"

type fieldValueState uint8

const (
	fieldValueUnknown fieldValueState = iota
	fieldValueElementSelected
)

type trackedField struct {
	FieldBinding
	identity   fieldIdentity
	source     bool
	valueState fieldValueState
}
type environment struct {
	rewrite         *rewriteFlow
	fields          map[fieldIdentityKey]trackedField
	fieldOrder      []fieldIdentityKey
	identities      map[fieldIdentityKey]fieldIdentity
	removed         map[fieldIdentityKey]bool
	removedOrder    []fieldIdentityKey
	ambiguous       map[string]bool
	open, uncertain bool
	requirements    requirementEnvironment
}

func newEnvironment() *environment {
	return newEnvironmentWithRequirementTrace(nil)
}
func newEnvironmentWithRequirementTrace(trace *requirementTrace) *environment {
	return &environment{fields: map[fieldIdentityKey]trackedField{}, identities: map[fieldIdentityKey]fieldIdentity{}, removed: map[fieldIdentityKey]bool{}, ambiguous: map[string]bool{}, open: true, requirements: newRequirementEnvironment(trace)}
}
func copyIDs(ids []string) []string { return append([]string{}, ids...) }
func (e *environment) snapshot() FieldState {
	s := FieldState{Fields: []FieldBinding{}, Removed: []string{}, Open: e.open, Uncertain: e.uncertain}
	projected := map[string]int{}
	for _, key := range e.orderedFieldKeys() {
		f, exists := e.fields[key]
		if !exists {
			continue
		}
		if index, found := projected[f.Name]; found {
			binding := &s.Fields[index]
			binding.OriginReferenceIDs = uniqueIDs(binding.OriginReferenceIDs, f.OriginReferenceIDs)
			binding.Conditional = binding.Conditional || f.Conditional
			s.Uncertain = true
			continue
		}
		f.OriginReferenceIDs = copyIDs(f.OriginReferenceIDs)
		projected[f.Name] = len(s.Fields)
		s.Fields = append(s.Fields, f.FieldBinding)
	}
	removed := map[string]bool{}
	for _, key := range e.orderedRemovedKeys() {
		if !e.removed[key] {
			continue
		}
		identity := e.identities[key]
		name := identity.PublicName
		if name == "" {
			name = string(key)
		}
		if _, live := projected[name]; live {
			continue
		}
		if !removed[name] {
			s.Removed = append(s.Removed, name)
			removed[name] = true
		}
	}
	sort.Slice(s.Fields, func(i, j int) bool { return s.Fields[i].Name < s.Fields[j].Name })
	sort.Strings(s.Removed)
	return s
}
func (e *environment) clone() *environment {
	n := newEnvironmentWithRequirementTrace(e.requirements.trace)
	n.rewrite = e.rewrite.clone()
	n.open = e.open
	n.uncertain = e.uncertain
	n.requirements = e.requirements.clone()
	for k, v := range e.fields {
		v.OriginReferenceIDs = copyIDs(v.OriginReferenceIDs)
		v.identity = v.identity.clone()
		n.fields[k] = v
	}
	n.fieldOrder = append([]fieldIdentityKey{}, e.fieldOrder...)
	for key, identity := range e.identities {
		n.identities[key] = identity.clone()
	}
	for k, v := range e.removed {
		n.removed[k] = v
	}
	n.removedOrder = append([]fieldIdentityKey{}, e.removedOrder...)
	for name, ambiguous := range e.ambiguous {
		n.ambiguous[name] = ambiguous
	}
	return n
}
func (e *environment) remove(name string) {
	e.removeIdentity(atomicFieldIdentity(name))
}
func (e *environment) install(name string, ids []string, conditional bool) {
	e.installIdentity(atomicFieldIdentity(name), ids, conditional, false)
}

func (e *environment) installIdentity(identity fieldIdentity, ids []string, conditional, source bool) (bool, []string) {
	key, exact := identity.privateKey()
	if !exact {
		return false, nil
	}
	if !source {
		e.rewriteInvalidate(identity.PublicName)
	}
	collision, origins := e.registerIdentityField(trackedField{FieldBinding: FieldBinding{Name: identity.PublicName, OriginReferenceIDs: copyIDs(ids), Conditional: conditional}, identity: identity.clone(), source: source})
	delete(e.removed, key)
	return collision, origins
}

func (e *environment) registerIdentityField(field trackedField) (bool, []string) {
	key, exact := field.identity.privateKey()
	if !exact {
		field.identity = atomicFieldIdentity(field.Name)
		key, _ = field.identity.privateKey()
	}
	if _, exists := e.fields[key]; !exists {
		e.fieldOrder = append(e.fieldOrder, key)
	}
	field.OriginReferenceIDs = copyIDs(field.OriginReferenceIDs)
	field.identity = field.identity.clone()
	e.fields[key] = field
	e.identities[key] = field.identity.clone()
	return e.detectCollision(field.Name)
}

func (e *environment) removeIdentity(identity fieldIdentity) {
	key, exact := identity.privateKey()
	if !exact {
		return
	}
	e.rewriteInvalidate(identity.PublicName)
	e.identities[key] = identity.clone()
	delete(e.fields, key)
	if !e.removed[key] {
		e.removedOrder = append(e.removedOrder, key)
	}
	e.removed[key] = true
}

func (e *environment) field(identity fieldIdentity) (trackedField, bool) {
	key, exact := identity.privateKey()
	if !exact {
		return trackedField{}, false
	}
	field, known := e.fields[key]
	return field, known
}

// selectElement records a private value-shape fact without changing public
// field presence, provenance, identity, or requirement evidence.
func (e *environment) selectElement(identity fieldIdentity) bool {
	key, exact := identity.privateKey()
	if !exact {
		return false
	}
	field, known := e.fields[key]
	if !known {
		return false
	}
	field.valueState = fieldValueElementSelected
	e.fields[key] = field
	return true
}

func (e *environment) hasOtherIdentity(identity fieldIdentity) bool {
	key, exact := identity.privateKey()
	if !exact {
		return false
	}
	for candidateKey, field := range e.fields {
		if candidateKey != key && field.Name == identity.PublicName {
			return true
		}
	}
	return false
}

func (e *environment) detectCollision(publicName string) (bool, []string) {
	identities := map[fieldIdentityKey]bool{}
	origins := []string{}
	for _, key := range e.orderedFieldKeys() {
		field, exists := e.fields[key]
		if !exists || field.Name != publicName {
			continue
		}
		identityKey, exact := field.identity.privateKey()
		if !exact {
			identityKey, _ = atomicFieldIdentity(field.Name).privateKey()
		}
		identities[identityKey] = true
		origins = uniqueIDs(origins, field.OriginReferenceIDs)
	}
	if len(identities) < 2 {
		return false, origins
	}
	first := !e.ambiguous[publicName]
	e.ambiguous[publicName] = true
	e.uncertain = true
	return first, origins
}

func (e *environment) orderedFieldKeys() []fieldIdentityKey {
	return orderedIdentityKeys(e.fieldOrder, e.fields)
}

func (e *environment) orderedRemovedKeys() []fieldIdentityKey {
	out := append([]fieldIdentityKey{}, e.removedOrder...)
	seen := map[fieldIdentityKey]bool{}
	for _, key := range out {
		seen[key] = true
	}
	missing := []string{}
	for key := range e.removed {
		if !seen[key] {
			missing = append(missing, string(key))
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		out = append(out, fieldIdentityKey(key))
	}
	return out
}

func orderedIdentityKeys[T any](order []fieldIdentityKey, values map[fieldIdentityKey]T) []fieldIdentityKey {
	out := append([]fieldIdentityKey{}, order...)
	seen := map[fieldIdentityKey]bool{}
	for _, key := range out {
		seen[key] = true
	}
	missing := []string{}
	for key := range values {
		if !seen[key] {
			missing = append(missing, string(key))
		}
	}
	sort.Strings(missing)
	for _, key := range missing {
		out = append(out, fieldIdentityKey(key))
	}
	return out
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
