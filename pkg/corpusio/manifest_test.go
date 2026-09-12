package corpusio

import (
	"encoding/json"
	"os"
	"testing"
)

func TestManifestFixture(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/tooling/requests.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name    string          `json:"name"`
		Kind    string          `json:"kind"`
		Valid   bool            `json:"valid"`
		Request json.RawMessage `json:"request"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		if tc.Kind != "manifest" {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			_, err := DecodeManifest(tc.Request)
			if (err == nil) != tc.Valid {
				t.Fatalf("valid=%v, err=%v", tc.Valid, err)
			}
		})
	}
}

func TestManifestExactlyOneSource(t *testing.T) {
	valid := `{"schema_version":1,"documents":[{"id":"empty","text":""},{"id":"path","path":"sub/query.spl2"}]}`
	m, err := DecodeManifest([]byte(valid))
	if err != nil || len(m.Documents) != 2 || m.Documents[0].Text == nil || *m.Documents[0].Text != "" || m.Documents[1].Path == nil {
		t.Fatalf("presence contract: %+v, %v", m, err)
	}
	if m.Base != "." || m.Documents[0].SourceID != "empty" || m.Documents[1].SourceID != "sub/query.spl2" || m.Documents[1].Language != "spl2" {
		t.Fatalf("manifest defaults: %+v", m)
	}
	explicit, err := DecodeManifest([]byte(`{"schema_version":1,"documents":[{"id":"q","path":"query.txt","language":"spl","source_id":""}]}`))
	if err != nil || explicit.Documents[0].SourceID != "" {
		t.Fatalf("explicit opaque source id: %+v, %v", explicit, err)
	}
	for _, tc := range []struct{ name, raw string }{
		{"both", `{"schema_version":1,"documents":[{"id":"a","text":"","path":"a.spl"}]}`},
		{"neither", `{"schema_version":1,"documents":[{"id":"a"}]}`},
		{"null", `{"schema_version":1,"documents":[{"id":"a","text":null}]}`},
		{"duplicate id", `{"schema_version":1,"documents":[{"id":"a","text":""},{"id":"a","text":""}]}`},
		{"duplicate key", `{"schema_version":1,"documents":[{"id":"a","text":"","text":"b"}]}`},
		{"surrogate", `{"schema_version":1,"documents":[{"id":"a","text":"\ud800"}]}`},
		{"unsupported", `{"schema_version":1,"documents":[{"id":"a","text":"","language":"sql"}]}`},
		{"absolute", `{"schema_version":1,"documents":[{"id":"a","path":"/tmp/a.spl"}]}`},
		{"drive", `{"schema_version":1,"documents":[{"id":"a","path":"C:/a.spl"}]}`},
		{"UNC", `{"schema_version":1,"documents":[{"id":"a","path":"//server/a.spl"}]}`},
		{"parent", `{"schema_version":1,"documents":[{"id":"a","path":"sub/../a.spl"}]}`},
		{"backslash", `{"schema_version":1,"documents":[{"id":"a","path":"sub\\a.spl"}]}`},
		{"nul path", `{"schema_version":1,"documents":[{"id":"a","path":"a\u0000.spl"}]}`},
		{"missing dialect", `{"schema_version":1,"documents":[{"id":"a","path":"a.txt"}]}`},
		{"empty selection", `{"schema_version":1,"documents":[]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeManifest([]byte(tc.raw)); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
}
