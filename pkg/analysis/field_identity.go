package analysis

import (
	"strconv"
	"strings"
)

const CodeAmbiguousField = "SPL_AMBIGUOUS_FIELD"

type fieldIdentityKind uint8

const (
	fieldIdentityAtomic fieldIdentityKind = iota
	fieldIdentityPath
)

type fieldIdentity struct {
	Kind       fieldIdentityKind
	Qualifier  string
	Segments   []string
	PublicName string
	exact      bool
}

type fieldIdentityKey = string

const fieldIdentityKeyPrefix = "\x00spl2-field:"

func atomicFieldIdentity(name string) fieldIdentity {
	return fieldIdentity{Kind: fieldIdentityAtomic, Segments: []string{name}, PublicName: name, exact: name != ""}
}

func pathFieldIdentity(qualifier string, segments []string) fieldIdentity {
	segments = append([]string{}, segments...)
	return fieldIdentity{Kind: fieldIdentityPath, Qualifier: qualifier, Segments: segments, PublicName: strings.Join(segments, "."), exact: len(segments) > 0}
}

func dynamicFieldIdentity(publicName string) fieldIdentity {
	return fieldIdentity{Kind: fieldIdentityPath, PublicName: publicName}
}

func (i fieldIdentity) privateKey() (fieldIdentityKey, bool) {
	if !i.exact || i.PublicName == "" {
		return "", false
	}
	var key strings.Builder
	key.WriteString(fieldIdentityKeyPrefix)
	if i.Kind == fieldIdentityAtomic {
		key.WriteByte('a')
		writeIdentityPart(&key, i.PublicName)
		return fieldIdentityKey(key.String()), true
	}
	key.WriteByte('p')
	writeIdentityPart(&key, i.Qualifier)
	writeIdentityPart(&key, strconv.Itoa(len(i.Segments)))
	for _, segment := range i.Segments {
		writeIdentityPart(&key, segment)
	}
	return fieldIdentityKey(key.String()), true
}

func writeIdentityPart(key *strings.Builder, value string) {
	key.WriteString(strconv.Itoa(len(value)))
	key.WriteByte(':')
	key.WriteString(value)
}

func (i fieldIdentity) clone() fieldIdentity {
	i.Segments = append([]string{}, i.Segments...)
	return i
}
