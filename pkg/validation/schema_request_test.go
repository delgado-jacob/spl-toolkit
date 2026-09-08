package validation

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

func TestDecodeSchemaWrappers(t *testing.T) {
	target := `"target":{"kind":"json_schema","schema":true}`
	r, e := DecodeSchemaRequest([]byte(`{"document":{"text":"table host"},` + target + `}`))
	if e != nil || r.Document.Language != "spl" {
		t.Fatalf("%+v %v", r, e)
	}
	b, e := DecodeSchemaBatchRequest([]byte(`{"documents":[{"text":"table host"}],` + target + `}`))
	if e != nil || len(b.Documents) != 1 {
		t.Fatalf("%+v %v", b, e)
	}
	for _, raw := range []string{`null`, `{}`, `{"document":null,` + target + `}`, `{"document":{"text":"x","text":"y"},` + target + `}`, `{"document":{"text":"\ud800"},` + target + `}`, `{"document":{"text":"x","language":"sql"},` + target + `}`, `{"document":{"text":"x"},` + target + `,"extra":0}`, `{"document":{"text":"x"},"target":null}`} {
		if _, e := DecodeSchemaRequest([]byte(raw)); !IsInputError(e) {
			t.Fatalf("accepted %s: %v", raw, e)
		}
	}
	for _, raw := range []string{`null`, `{"documents":[],` + target + `}`, `{"documents":null,` + target + `}`, `{"documents":[null],` + target + `}`} {
		if _, e := DecodeSchemaBatchRequest([]byte(raw)); !IsInputError(e) {
			t.Fatalf("accepted %s: %v", raw, e)
		}
	}
}

func TestSchemaMalformedRequestCorpus(t *testing.T) {
	raw, e := os.ReadFile("../../testdata/schemas/requests.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		SchemaVersion int `json:"schema_version"`
		Cases         []struct {
			ID    string `json:"id"`
			Batch bool   `json:"batch"`
			Input string `json:"input"`
			Hex   string `json:"hex,omitempty"`
		} `json:"cases"`
	}
	if e = json.Unmarshal(raw, &corpus); e != nil {
		t.Fatal(e)
	}
	if corpus.SchemaVersion != 1 || len(corpus.Cases) < 20 {
		t.Fatal("incomplete request corpus")
	}
	for _, tc := range corpus.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			data := []byte(tc.Input)
			if tc.Hex != "" {
				data, e = hex.DecodeString(tc.Hex)
				if e != nil {
					t.Fatal(e)
				}
			}
			if tc.Batch {
				r, e := DecodeSchemaBatchRequest(data)
				if e == nil {
					var b *SchemaBatchReport
					b, e = ValidateSchemaBatch(r.Documents, r.Target)
					if b != nil {
						t.Fatal("invalid input returned batch")
					}
				}
				if !IsInputError(e) {
					t.Fatalf("want InputError: %v", e)
				}
			} else {
				r, e := DecodeSchemaRequest(data)
				if e == nil {
					var report *SchemaReport
					report, e = ValidateSchema(r.Document, r.Target)
					if report != nil {
						t.Fatal("invalid input returned report")
					}
				}
				if !IsInputError(e) {
					t.Fatalf("want InputError: %v", e)
				}
			}
		})
	}
}
