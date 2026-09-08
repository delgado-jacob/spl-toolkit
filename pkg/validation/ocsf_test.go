package validation

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"
)

func readOCSFFixture(t *testing.T, variant string) json.RawMessage {
	t.Helper()
	hashes := map[string]string{"base": "9b609f8fb670772f04191c1c276b46d34d6e9110d2417c71fa89c4f54c585137", "windows": "19af77ce259f3ff57debc33e52da51b8220a1af59399d06553f29629678595e9"}
	f, err := os.Open("../../testdata/schemas/ocsf/1.6.0/" + variant + ".json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != hashes[variant] {
		t.Fatalf("%s hash %s", variant, got)
	}
	return raw
}

func TestOCSFRealAuthentication(t *testing.T) {
	raw := readOCSFFixture(t, "base")
	prepared, err := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw,
		Selection: &OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{}, Extensions: []string{}}})
	if err != nil {
		t.Fatal(err)
	}
	for name, outcome := range map[string]string{"time": "required", "unmapped.vendor_field": "permitted_unspecified"} {
		if got := prepared.project(name); got.Outcome != outcome {
			t.Fatalf("%s=%+v", name, got)
		}
	}
}

func TestOCSFRejectMalformedCatalogs(t *testing.T) {
	raw := edgeOCSF(t)
	cases := map[string]func(map[string]any){
		"compile version": func(m map[string]any) { m["compile_version"] = 2 },
		"null version":    func(m map[string]any) { m["version"] = nil },
		"missing objects": func(m map[string]any) { delete(m, "objects") },
		"null attributes": func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["attributes"] = nil },
		"bad requirement": func(m map[string]any) {
			m["classes"].(map[string]any)["a"].(map[string]any)["attributes"].(map[string]any)["time"].(map[string]any)["requirement"] = "sometimes"
		},
		"dangling object":       func(m map[string]any) { delete(m["objects"].(map[string]any), "empty") },
		"dangling category":     func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["category"] = "lost" },
		"category uid mismatch": func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["category_uid"] = 2 },
		"colliding class uid":   func(m map[string]any) { m["classes"].(map[string]any)["b"].(map[string]any)["uid"] = 1001 },
		"colliding category uid": func(m map[string]any) {
			m["categories"].(map[string]any)["attributes"].(map[string]any)["empty"].(map[string]any)["uid"] = 1
		},
		"sentinel lookalike": func(m map[string]any) { m["classes"].(map[string]any)["base_event"].(map[string]any)["uid"] = 9 },
		"sentinel wrong category": func(m map[string]any) {
			m["classes"].(map[string]any)["base_event"].(map[string]any)["category"] = "test"
		},
		"zero uid lookalike": func(m map[string]any) { m["classes"].(map[string]any)["b"].(map[string]any)["uid"] = 0 },
		"wrong name":         func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["name"] = "wrong" },
		"unknown profile dependency": func(m map[string]any) {
			m["classes"].(map[string]any)["a"].(map[string]any)["attributes"].(map[string]any)["time"].(map[string]any)["profiles"] = []any{"lost"}
		},
		"hidden selected":   func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["hidden"] = true },
		"abstract selected": func(m map[string]any) { m["classes"].(map[string]any)["a"].(map[string]any)["abstract"] = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			var m map[string]any
			if e := json.Unmarshal(raw, &m); e != nil {
				t.Fatal(e)
			}
			mutate(m)
			b, _ := json.Marshal(m)
			if _, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: b, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}}); !IsInputError(e) {
				t.Fatalf("want InputError, got %v", e)
			}
		})
	}
	for name, s := range map[string]OCSFSelection{"base": {Class: "base_event"}, "empty category": {Category: "empty"}, "unknown class": {Class: "lost"}, "unknown profile": {Class: "a", Profiles: []string{"lost"}}, "inapplicable profile": {Class: "b", Profiles: []string{"p"}}, "wrong version": {Version: "1.5.0", Class: "a"}, "extensions": {Class: "a", Extensions: []string{"win"}}} {
		t.Run(name, func(t *testing.T) {
			if s.Version == "" {
				s.Version = "1.6.0"
			}
			if _, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &s}); !IsInputError(e) {
				t.Fatalf("want InputError, got %v", e)
			}
		})
	}
}

func TestOCSFDispatcherAndTypedGate(t *testing.T) {
	p, e := prepareSchemaTarget(SchemaTarget{Kind: "ocsf", Catalog: edgeOCSF(t), Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}})
	if e != nil {
		t.Fatal(e)
	}
	assertOCSF(t, p, "time", "required")
	q, e := prepareSchemaTarget(SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}},"additionalProperties":false}`)})
	if e != nil {
		t.Fatal(e)
	}
	if q.project("x").Outcome != "optional" {
		t.Fatal("JSON dispatcher")
	}
	for _, target := range []SchemaTarget{{Kind: "future"}, {Kind: "ocsf", Catalog: edgeOCSF(t)}, {Kind: "ocsf", Catalog: edgeOCSF(t), Schema: json.RawMessage(`{}`), Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}}} {
		if _, e := prepareSchemaTarget(target); !IsInputError(e) {
			t.Fatalf("invalid target accepted %v", e)
		}
	}
}

func TestOCSFRejectUndeclaredNamespace(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	classes := m["classes"].(map[string]any)
	classes["lost/a"] = classes["a"]
	delete(classes, "a")
	raw, _ := json.Marshal(m)
	if _, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "lost/a"}}); !IsInputError(e) {
		t.Fatalf("uncompiled namespace accepted: %v", e)
	}
}

func TestOCSFRejectNullArrayMetadata(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	m["classes"].(map[string]any)["a"].(map[string]any)["attributes"].(map[string]any)["array"].(map[string]any)["is_array"] = nil
	raw, _ := json.Marshal(m)
	if _, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "a"}}); !IsInputError(e) {
		t.Fatalf("null is_array silently closed descendants: %v", e)
	}
}
