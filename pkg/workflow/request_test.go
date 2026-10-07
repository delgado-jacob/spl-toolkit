package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"testing"
)

func TestDecodeWorkflowRejectsDuplicateRootKey(t *testing.T) {
	raw, err := json.Marshal(seedRequest(t, "from main"))
	if err != nil {
		t.Fatal(err)
	}
	raw = append([]byte(`{"schema_version":1,`), raw[1:]...)
	if _, err := DecodeRequest(raw); err == nil {
		t.Fatal("accepted duplicate root key")
	}
}

func TestDecodeWorkflowRejectsInvalidAdmission(t *testing.T) {
	valid, err := json.Marshal(seedRequest(t, "from main"))
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		raw  []byte
		path string
	}{
		"trailing": {append(append([]byte{}, valid...), []byte(` {}`)...), ""},
		"unicode":  {[]byte(`{"schema_version":1,"documents":[{"id":"\ud800"}]}`), ""},
		"utf8":     {append([]byte(`{"format":"`), 0xff), ""},
		"unknown":  {append([]byte(`{"extra":1,`), valid[1:]...), "/extra"},
		"fraction": {bytes.Replace(valid, []byte(`"schema_version":1`), []byte(`"schema_version":1.0`), 1), "/schema_version"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeRequest(c.raw)
			if err == nil {
				t.Fatal("accepted invalid input")
			}
			d, ok := RequestErrorDetails(err)
			if !ok || d.Path != c.path || !validation.IsInputError(err) {
				t.Fatalf("missing structured input detail: %#v %v", d, err)
			}
		})
	}
	mutations := map[string]func(*Request){
		"null documents":      func(r *Request) { r.Documents = nil },
		"null entries":        func(r *Request) { r.Settings.Entries = nil },
		"null bindings":       func(r *Request) { r.Settings.Entries[0].Compatibility.InputBindings = nil },
		"invalid selector":    func(r *Request) { r.Settings.Entries[0].Compatibility.QueryScope.App = environment.Selector{} },
		"duplicate documents": func(r *Request) { r.Documents = append(r.Documents, r.Documents[0]) },
		"blank document":      func(r *Request) { r.Documents[0].ID = " " },
		"duplicate settings":  func(r *Request) { r.Settings.Entries = append(r.Settings.Entries, r.Settings.Entries[0]) },
		"blank settings":      func(r *Request) { r.Settings.Entries[0].ID = " " },
		"missing settings":    func(r *Request) { r.Settings.Entries = []EntrySettings{} },
		"unused settings":     func(r *Request) { r.Settings.Entries[0].ID = "other" },
		"neither mode":        func(r *Request) { r.Settings.Entries[0].Compatibility = nil },
		"both modes": func(r *Request) {
			r.Settings.Entries[0].Resolution = &ResolveSettings{Resolutions: []resolution.Resolution{}, Compatibility: compatibility.ResolutionAssessment{QueryScope: r.Settings.Entries[0].Compatibility.QueryScope, InputBindings: []compatibility.ResolutionBinding{}}}
		},
		"bad capability": func(r *Request) { r.Documents[0].Document.Language = "bogus" },
		"bad format":     func(r *Request) { r.Format = "yaml" },
		"schema version": func(r *Request) { r.Settings.SchemaVersion = 2 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			r := seedRequest(t, "from main")
			mutate(&r)
			raw, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeRequest(raw); err == nil {
				t.Fatal("accepted invalid input")
			}
			if _, err := normalizeRequest(r); err == nil {
				t.Fatal("accepted invalid typed input")
			}
		})
	}
}
func TestDecodeWorkflowNestedKeysAndNullOptionals(t *testing.T) {
	raw, _ := json.Marshal(seedRequest(t, "from main"))
	for name, bad := range map[string][]byte{
		"nested duplicate":    bytes.Replace(raw, []byte(`"id":"d1"`), []byte(`"id":"d1","id":"d1"`), 1),
		"nested unknown":      bytes.Replace(raw, []byte(`"text":`), []byte(`"extra":true,"text":`), 1),
		"optional null":       append([]byte(`{"format":null,`), raw[1:]...),
		"selector both empty": bytes.Replace(raw, []byte(`"all":true`), []byte(`"all":true,"values":[]`), 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeRequest(bad); err == nil {
				t.Fatal("accepted invalid nested shape")
			}
		})
	}
}
func TestDecodeWorkflowSettingsIndependentAndDetached(t *testing.T) {
	r := seedRequest(t, "from main")
	r.Settings.Entries[0].ID = "acquired-later"
	raw, _ := json.Marshal(r.Settings)
	s, err := DecodeSettings(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Entries) != 1 {
		t.Fatal("lost settings")
	}
	normalized, err := normalizeSettings(s)
	if err != nil {
		t.Fatal(err)
	}
	normalized.Entries[0].ID = "changed"
	if s.Entries[0].ID != "acquired-later" {
		t.Fatal("normalization aliases input")
	}
}
func TestDecodeWorkflowTypedInvalidUTF8(t *testing.T) {
	r := seedRequest(t, string([]byte{0xff}))
	if _, err := normalizeRequest(r); err == nil {
		t.Fatal("accepted invalid typed UTF-8")
	}
	r = seedRequest(t, "from main")
	r.Settings.Snapshot.ScopeID = string([]byte{0xff})
	if _, err := normalizeSettings(r.Settings); err == nil {
		t.Fatal("accepted invalid settings UTF-8")
	}
}
func TestDecodeWorkflowValidFormatsAndErrorDetailDetachment(t *testing.T) {
	for _, format := range []string{"", "text", "json", "sarif", "graph", "bom"} {
		r := seedRequest(t, "from main")
		r.Format = format
		raw, _ := json.Marshal(r)
		got, err := DecodeRequest(raw)
		if err != nil {
			t.Fatal(err)
		}
		if got.Settings.Entries[0].Compatibility.InputBindings == nil {
			t.Fatal("empty required array became nil")
		}
	}
	err := requestErrorOffset("request_invalid", "/schema_version", "invalid", 4)
	first, ok := RequestErrorDetails(fmt.Errorf("wrapped: %w", err))
	if !ok {
		t.Fatal("missing wrapped detail")
	}
	*first.ByteOffset = 9
	second, _ := RequestErrorDetails(err)
	if *second.ByteOffset != 4 {
		t.Fatal("details alias internal offset")
	}
}

func TestDecodeWorkflowResolutionRequiredArrays(t *testing.T) {
	r := seedRequest(t, "from $source")
	scope := r.Settings.Entries[0].Compatibility.QueryScope
	r.Settings.Entries[0].Compatibility = nil
	r.Settings.Entries[0].Resolution = &ResolveSettings{Resolutions: []resolution.Resolution{{Placeholder: "$source", Kind: "dataset", Values: nil}}, Compatibility: compatibility.ResolutionAssessment{QueryScope: scope, InputBindings: []compatibility.ResolutionBinding{}}}
	raw, _ := json.Marshal(r)
	if _, err := DecodeRequest(raw); err == nil {
		t.Fatal("accepted null choice values")
	}
	if _, err := normalizeRequest(r); err == nil {
		t.Fatal("accepted nil typed choice values")
	}
	r.Settings.Entries[0].Resolution.Resolutions[0].Values = []string{}
	raw, _ = json.Marshal(r)
	if _, err := DecodeRequest(raw); err != nil {
		t.Fatal(err)
	}
	if _, err := normalizeRequest(r); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeWorkflowUnicodeOffsets(t *testing.T) {
	for _, raw := range [][]byte{append([]byte(`{"format":"`), 0xff), []byte(`{"format":"\ud800"}`), []byte(`{"format":"ok\udc00"}`), []byte(`{"format":"\ud800\udc00\ud800"}`)} {
		expected := bytes.IndexByte(raw, 0xff)
		if expected < 0 {
			expected = bytes.LastIndex(raw, []byte(`\u`))
		}
		_, err := DecodeRequest(raw)
		detail, ok := RequestErrorDetails(err)
		if !ok || detail.ByteOffset == nil || *detail.ByteOffset != expected {
			t.Fatalf("offset for %q: %#v", raw, detail)
		}
	}
}
func TestDecodeWorkflowDelegatesSelectorValueSemantics(t *testing.T) {
	for _, values := range [][]string{{" "}, {"app", "app"}} {
		r := seedRequest(t, "from main")
		r.Settings.Entries[0].Compatibility.QueryScope.App = environment.Selector{Values: values}
		raw, _ := json.Marshal(r)
		if _, err := DecodeRequest(raw); err != nil {
			t.Fatal(err)
		}
		if _, err := normalizeRequest(r); err != nil {
			t.Fatal(err)
		}
	}
}
