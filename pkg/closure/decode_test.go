package closure

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const orderedRequest = `{"schema_version":1,"document":{"text":"` + "`expand`" + `","language":"spl"},"bundle":{"schema_version":1,"scope_id":"scope-A","collections":[{"kind":"lookup","coverage":"partial"},{"kind":"macro","coverage":"complete"}],"objects":[{"id":"lookup-1","kind":"lookup","name":"users","source_id":"users.csv"},{"id":"macro-1","kind":"macro","name":"expand","source_id":"macros.conf","document":{"text":"lookup users | lookup missing"},"arity":0,"arguments":[],"eval_based":false,"validation":"opaque","relations":[{"kind":"lookup","name":"missing","start":15,"end":29},{"kind":"lookup","name":"users","start":0,"end":12}]}]},"bindings":[]}`

const reorderedRequest = `{"bindings":[],"bundle":{"objects":[{"relations":[{"end":12,"start":0,"name":"users","kind":"lookup"},{"end":29,"start":15,"name":"missing","kind":"lookup"}],"validation":"opaque","eval_based":false,"arguments":[],"arity":0,"document":{"language":"spl","profile":"splunkd","version":"current","text":"lookup users | lookup missing"},"source_id":"macros.conf","name":"expand","kind":"macro","id":"macro-1"},{"source_id":"users.csv","name":"users","kind":"lookup","id":"lookup-1"}],"collections":[{"coverage":"complete","kind":"macro"},{"coverage":"partial","kind":"lookup"}],"scope_id":"scope-A","schema_version":1},"document":{"version":"current","profile":"splunkd","language":"spl","text":"` + "`expand`" + `"},"schema_version":1}`

func TestBundleDigestCanonicalOrderAndNoMutation(t *testing.T) {
	firstRaw := []byte(orderedRequest)
	originalRaw := append([]byte(nil), firstRaw...)
	first, err := DecodeRequest(firstRaw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstRaw, originalRaw) {
		t.Fatal("decoder changed caller bytes")
	}
	second, err := DecodeRequest([]byte(reorderedRequest))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := json.Marshal(first.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	a, err := BundleDigest(first.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	b, err := BundleDigest(second.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	if a != b || !strings.HasPrefix(a, "sha256:") {
		t.Fatalf("digests differ: %q %q", a, b)
	}
	after, err := json.Marshal(first.Bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(snapshot, after) {
		t.Fatal("digest changed decoded bundle")
	}
	changed := first.Bundle
	changed.Objects = append([]Definition(nil), first.Bundle.Objects...)
	changedDocument := *changed.Objects[1].Document
	changedDocument.Text += " "
	changed.Objects[1].Document = &changedDocument
	changedDigest, err := BundleDigest(changed)
	if err != nil {
		t.Fatal(err)
	}
	if changedDigest == a {
		t.Fatal("exact source text did not affect digest")
	}
	if first.Document.Language != "spl" || first.Document.Profile != "splunkd" || first.Document.Version != "current" {
		t.Fatalf("document selectors not normalized: %+v", first.Document)
	}
	if first.Bundle.Collections == nil || first.Bundle.Objects == nil || first.Bindings == nil || first.Bundle.Objects[0].Relations == nil || first.Bundle.Objects[1].Arguments == nil {
		t.Fatal("canonical arrays must not be nil")
	}
	if first.Bundle.Objects[1].Document == nil || first.Bundle.Objects[1].Document.Text != "lookup users | lookup missing" {
		t.Fatal("definition text changed")
	}
}

func TestDecodeAbsentRelationTargetIsContent(t *testing.T) {
	request, err := DecodeRequest([]byte(orderedRequest))
	if err != nil {
		t.Fatal(err)
	}
	got := request.Bundle.Objects[1].Relations[0]
	if got.Kind != "lookup" || got.Name != "missing" {
		t.Fatalf("absent relation target lost: %+v", got)
	}
}

func TestDecodeRejectsMalformedInput(t *testing.T) {
	base := orderedRequest
	cases := []struct{ name, raw string }{
		{"duplicate root member", strings.Replace(base, `"schema_version":1,`, `"schema_version":1,"schema_version":1,`, 1)},
		{"unknown root member", strings.Replace(base, `"bindings":[]`, `"path":"x","bindings":[]`, 1)},
		{"null document", strings.Replace(base, `"document":{"text":"`+"`expand`"+`","language":"spl"}`, `"document":null`, 1)},
		{"null collections", strings.Replace(base, `"collections":[{"kind":"lookup","coverage":"partial"},{"kind":"macro","coverage":"complete"}]`, `"collections":null`, 1)},
		{"null objects", `{"schema_version":1,"document":{"text":""},"bundle":{"schema_version":1,"scope_id":"s","objects":null},"bindings":[]}`},
		{"null bindings", strings.Replace(base, `"bindings":[]`, `"bindings":null`, 1)},
		{"trailing value", base + ` {}`},
		{"unsupported selector", strings.Replace(base, `"language":"spl"`, `"language":"sql"`, 1)},
		{"duplicate object id", strings.Replace(base, `"objects":[`, `"objects":[{"id":"lookup-1","kind":"lookup","name":"other","source_id":"other"},`, 1)},
		{"inconsistent macro arity", strings.Replace(base, `"arity":0`, `"arity":1`, 1)},
		{"arguments on non-macro", strings.Replace(base, `"name":"users","source_id":"users.csv"`, `"name":"users","source_id":"users.csv","arguments":[]`, 1)},
		{"repeated argument names", strings.Replace(base, `"arity":0,"arguments":[]`, `"arity":2,"arguments":["x","x"]`, 1)},
		{"contradictory collection", strings.Replace(base, `"collections":[`, `"collections":[{"kind":"macro","coverage":"partial"},`, 1)},
		{"invalid coverage", strings.Replace(base, `"coverage":"partial"`, `"coverage":"unknown"`, 1)},
		{"invalid object kind", strings.Replace(base, `"kind":"lookup","name":"users"`, `"kind":"index","name":"users"`, 1)},
		{"relation with two evidence forms", strings.Replace(base, `"end":29}`, `"end":29,"property":"/x"}`, 1)},
		{"relation with no evidence", strings.Replace(base, `,"start":15,"end":29`, ``, 1)},
		{"relation range out of bounds", strings.Replace(base, `"end":29`, `"end":999`, 1)},
		{"relation invalid pointer", strings.Replace(base, `"start":15,"end":29`, `"property":"not-a-pointer"`, 1)},
		{"binding digest malformed", strings.Replace(base, `"bindings":[]`, `"bindings":[{"document_digest":"sha256:bad","kind":"macro","start":0,"end":8,"object_id":"macro-1"}]`, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := DecodeRequest([]byte(tc.raw))
			if !IsInputError(err) {
				t.Fatalf("expected input error, got %+v, %v", got, err)
			}
			if !reflect.DeepEqual(got, Request{}) {
				t.Fatalf("invalid input returned request: %+v", got)
			}
		})
	}
	badUTF8 := append([]byte(orderedRequest), 0xff)
	if got, err := DecodeRequest(badUTF8); !IsInputError(err) || !reflect.DeepEqual(got, Request{}) {
		t.Fatalf("invalid UTF-8: %+v, %v", got, err)
	}
}

func TestDecodeBindingStructureAndKindVocabulary(t *testing.T) {
	sum := sha256.Sum256([]byte("`expand`"))
	digest := "sha256:" + hex.EncodeToString(sum[:])
	binding := fmt.Sprintf(`"bindings":[{"document_digest":%q,"kind":"macro","start":0,"end":8,"object_id":"macro-1"}]`, digest)
	raw := strings.Replace(orderedRequest, `"bindings":[]`, binding, 1)
	got, err := DecodeRequest([]byte(raw))
	if err != nil || len(got.Bindings) != 1 {
		t.Fatalf("well-formed binding: %+v, %v", got.Bindings, err)
	}
	for _, tc := range []struct{ name, raw string }{
		{"wrong object", strings.Replace(raw, `"object_id":"macro-1"`, `"object_id":"absent"`, 1)},
		{"wrong kind", strings.Replace(raw, `"kind":"macro","start":0`, `"kind":"lookup","start":0`, 1)},
		{"range outside document", strings.Replace(raw, `"end":8,"object_id"`, `"end":9,"object_id"`, 1)},
		{"stale digest", strings.Replace(raw, digest, `sha256:`+strings.Repeat("0", 64), 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if r, e := DecodeRequest([]byte(tc.raw)); !IsInputError(e) || !reflect.DeepEqual(r, Request{}) {
				t.Fatalf("got %+v, %v", r, e)
			}
		})
	}
	for _, kind := range []string{"macro", "lookup", "data_model", "dataset", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module", "function", "external_command"} {
		t.Run(kind, func(t *testing.T) {
			collection := fmt.Sprintf(`{"kind":%q,"coverage":"unavailable"}`, kind)
			input := fmt.Sprintf(`{"schema_version":1,"document":{"text":""},"bundle":{"schema_version":1,"scope_id":"s","collections":[%s],"objects":[]},"bindings":[]}`, collection)
			if _, err := DecodeRequest([]byte(input)); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestDecodeObjectsDoNotRequireCollectionCoverage(t *testing.T) {
	for _, tc := range []struct{ name, raw string }{
		{"omitted", strings.Replace(orderedRequest, `{"kind":"lookup","coverage":"partial"},`, ``, 1)},
		{"unavailable", strings.Replace(orderedRequest, `"kind":"lookup","coverage":"partial"`, `"kind":"lookup","coverage":"unavailable"`, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := DecodeRequest([]byte(tc.raw))
			if err != nil || len(r.Bundle.Objects) != 2 {
				t.Fatalf("supplied object rejected: %+v, %v", r.Bundle.Objects, err)
			}
		})
	}
}
