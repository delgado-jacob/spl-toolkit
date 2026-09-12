package corpus

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCorpusRequestFixture(t *testing.T) {
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
		if tc.Kind != "corpus" {
			continue
		}
		t.Run(tc.Name, func(t *testing.T) {
			_, err := DecodeRequest(tc.Request)
			if (err == nil) != tc.Valid {
				t.Fatalf("valid=%v, err=%v", tc.Valid, err)
			}
		})
	}
}

func TestCorpusRequestStrict(t *testing.T) {
	valid := `{"schema_version":1,"documents":[{"id":"one","document":{"text":""}}]}`
	r, err := DecodeRequest([]byte(valid))
	if err != nil || len(r.Documents) != 1 || r.Documents[0].ID != "one" || r.Documents[0].Document.Text != "" || r.Documents[0].Document.Language != "spl" {
		t.Fatalf("empty source should be a normalized document: %+v, %v", r, err)
	}
	for _, tc := range []struct{ name, raw string }{
		{"duplicate ids", `{"schema_version":1,"documents":[{"id":"same","document":{"text":"a"}},{"id":"same","document":{"text":"b"}}]}`},
		{"duplicate wrapper key", `{"schema_version":1,"schema_version":1,"documents":[{"id":"x","document":{"text":""}}]}`},
		{"duplicate document key", `{"schema_version":1,"documents":[{"id":"x","document":{"text":"a","text":"b"}}]}`},
		{"unknown key", `{"schema_version":1,"documents":[{"id":"x","document":{"text":""},"path":"a.spl"}]}`},
		{"null text", `{"schema_version":1,"documents":[{"id":"x","document":{"text":null}}]}`},
		{"unpaired surrogate", `{"schema_version":1,"documents":[{"id":"x","document":{"text":"\ud800"}}]}`},
		{"unsupported language", `{"schema_version":1,"documents":[{"id":"x","document":{"text":"","language":"sql"}}]}`},
		{"empty documents", `{"schema_version":1,"documents":[]}`},
		{"bad version", `{"schema_version":1.0,"documents":[{"id":"x","document":{"text":""}}]}`},
		{"null target", `{"schema_version":1,"documents":[{"id":"x","document":{"text":""}}],"validation_target":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeRequest([]byte(tc.raw)); !IsInputError(err) {
				t.Fatalf("expected input error, got %v", err)
			}
		})
	}
}
