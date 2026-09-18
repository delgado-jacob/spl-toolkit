package analysis

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

// These tests catch document mutation, unsupported-option acceptance, and null report arrays.
func TestDocumentAndUnicodeSource(t *testing.T) {
	text := "  search host=\"é\"\n| eval 'café'=1  "
	r, err := Analyze(QueryDocument{Text: text, SourceID: "queries/é.spl"})
	if err != nil {
		t.Fatal(err)
	}
	if r.Document.Text != text || r.Document.SourceID != "queries/é.spl" {
		t.Fatal(r.Document)
	}
	if r.Document.Language != "spl" || r.Document.Profile != "splunkd" || r.Document.Version != "current" {
		t.Fatal(r.Document)
	}
	if len(r.Stages) != 2 || r.Stages[1].Command != "eval" {
		t.Fatal(r.Stages)
	}
	if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete {
		t.Fatal(r.Status, r.Coverage)
	}
	if r.SchemaVersion != 1 || Capabilities().SchemaVersion != 1 {
		t.Fatal("report schema")
	}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if err = json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	assertNoNull(t, decoded)
	b, err = json.Marshal(Capabilities())
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	assertNoNull(t, decoded)
}
func assertNoNull(t *testing.T, v any) {
	t.Helper()
	switch x := v.(type) {
	case nil:
		t.Fatal("null report value")
	case []any:
		for _, v := range x {
			assertNoNull(t, v)
		}
	case map[string]any:
		for _, v := range x {
			assertNoNull(t, v)
		}
	}
}

func TestAnalysisCollectionsAreInitialized(t *testing.T) {
	for _, options := range []CapabilityOptions{{}, {Language: "spl2"}} {
		manifest, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		assertNoNull(t, decoded)
	}
}

func TestDocumentUnsupportedOptions(t *testing.T) {
	for _, tc := range []struct {
		doc  QueryDocument
		want string
	}{
		{QueryDocument{Language: "SPL"}, `unsupported language "SPL"`},
		{QueryDocument{Profile: "cloud"}, `unsupported profile "cloud"`},
		{QueryDocument{Version: "9.0"}, `unsupported compatibility version "9.0"`},
	} {
		r, err := Analyze(tc.doc)
		if r != nil || err == nil || err.Error() != tc.want {
			t.Fatalf("%+v: %v %v", tc.doc, r, err)
		}
	}
}
func TestDocumentDeterministicConcurrent(t *testing.T) {
	doc := QueryDocument{Text: "search host=web | eval x=lower(host) | mystery x"}
	r, _ := Analyze(doc)
	want, _ := json.Marshal(r)
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := Analyze(doc)
			if e != nil {
				t.Error(e)
				return
			}
			got, _ := json.Marshal(r)
			if string(got) != string(want) {
				t.Error("nondeterministic result")
			}
		}()
	}
	wg.Wait()
	if strings.Contains(string(want), ":null") {
		t.Fatal(string(want))
	}
}
