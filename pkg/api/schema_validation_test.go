package api

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func schemaREST(t *testing.T, handler http.Handler, route string, payload []byte, contentType string, unknownLength bool) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, "/api/v1/query/"+route, bytes.NewReader(payload))
	r.Header.Set("Content-Type", contentType)
	if unknownLength {
		r.ContentLength = -1
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestSchemaRESTCanonicalReports(t *testing.T) {
	handler := NewServer().Handler()
	for _, schema := range []string{`{"properties":{"host":{}},"additionalProperties":false}`, `{}`, `{"$ref":"https://offline.invalid/missing"}`} {
		target := validation.SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(schema)}
		documents := []analysis.QueryDocument{{Text: "table absent", SourceID: "first"}, {Text: "search host=\"café\"\r\n| table h*\r\n", SourceID: "second"}}
		for _, batch := range []bool{false, true} {
			var request, want any
			route := "validate-schema"
			var err error
			if batch {
				route += "/batch"
				request = validation.SchemaBatchRequest{Documents: documents, Target: target}
				want, err = validation.ValidateSchemaBatch(documents, target)
			} else {
				request = validation.SchemaRequest{Document: documents[0], Target: target}
				want, err = validation.ValidateSchema(documents[0], target)
			}
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			w := schemaREST(t, handler, route, payload, "application/json; charset=utf-8", false)
			if w.Code != 200 {
				t.Fatalf("%s: %d %s", route, w.Code, w.Body)
			}
			expected, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), expected) {
				t.Fatalf("parity mismatch: %s", w.Body)
			}
		}
	}
}

func TestSchemaRESTStrictWrappers(t *testing.T) {
	handler := NewServer().Handler()
	target := `{"kind":"json_schema","schema":{}}`
	badTargets := []string{`null`, `[]`, `{}`, `{"kind":"wat","schema":{}}`, `{"kind":"json_schema","schema":{},"catalog":null}`, `{"kind":"json_schema","schema":null}`, `{"kind":"json_schema","schema":{},"extra":1}`, `{"kind":"json_schema","kind":"json_schema","schema":{}}`, `{"kind":"json_schema","schema":{"title":"a","title":"b"}}`, `{"kind":"json_schema","schema":{},"resources":null}`, `{"kind":"json_schema","schema":{},"base_uri":null}`, `{"kind":"json_schema","schema":{"$schema":"http://json-schema.org/draft-04/schema#"}}`,
		`{"kind":"ocsf","catalog":{},"selection":null}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"x","profiles":null}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"x","extensions":null}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"x","class_uid":1}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class_uid":1.1}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class_uid":9223372036854775808}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"x","profiles":["p","p"]}}`, `{"kind":"ocsf","catalog":{},"selection":{"version":"1.6.0","class":"x","unknown":1}}`,
	}
	for _, batch := range []bool{false, true} {
		route := "validate-schema"
		key := "document"
		doc := `{"text":"table host"}`
		if batch {
			route += "/batch"
			key = "documents"
			doc = "[" + doc + "]"
		}
		wrap := func(d, t string) string { return fmt.Sprintf(`{"%s":%s,"target":%s}`, key, d, t) }
		bad := []string{`null`, `[]`, `{}`, wrap("null", target), wrap(doc, "null"), wrap(doc, target) + `{}`, strings.TrimSuffix(wrap(doc, target), "}") + `,"extra":1}`, strings.TrimSuffix(wrap(doc, target), "}") + `,"target":` + target + `}`}
		for _, tg := range badTargets {
			bad = append(bad, wrap(doc, tg))
		}
		for _, d := range []string{`{}`, `{"text":null}`, `{"text":"x","unknown":1}`, `{"text":"x","text":"y"}`, `{"text":"\ud800"}`, "{\"text\":\"\xff\"}", `{"text":"x","profile":null}`} {
			if batch {
				d = "[" + d + "]"
			}
			bad = append(bad, wrap(d, target))
		}
		if batch {
			bad = append(bad, wrap("[]", target), wrap("{}", target))
		}
		for _, payload := range bad {
			w := schemaREST(t, handler, route, []byte(payload), "application/json", false)
			if w.Code != 400 || !strings.Contains(w.Body.String(), `"error":true`) {
				t.Errorf("%s %s: %d %s", route, payload, w.Code, w.Body)
			}
		}
		for _, ct := range []string{"", "text/plain", "application/json; malformed"} {
			w := schemaREST(t, handler, route, []byte(wrap(doc, target)), ct, false)
			if w.Code != 400 {
				t.Errorf("content type %q: %d", ct, w.Code)
			}
		}
	}
}

func TestSchemaRESTBodyLimitsAndLegacyRetention(t *testing.T) {
	handler := NewServer().Handler()
	for _, route := range []string{"validate-schema", "validate-schema/batch"} {
		doc := `"document":{"text":"table host"}`
		if strings.HasSuffix(route, "/batch") {
			doc = `"documents":[{"text":"table host"}]`
		}
		base := []byte(`{` + doc + `,"target":{"kind":"json_schema","schema":{}}}`)
		exact := append(base, bytes.Repeat([]byte(" "), (8<<20)-len(base))...)
		w := schemaREST(t, handler, route, exact, "application/json", false)
		if w.Code != 200 {
			t.Fatalf("exact limit %s: %d %s", route, w.Code, w.Body)
		}
		for _, unknown := range []bool{false, true} {
			w = schemaREST(t, handler, route, append(exact, ' '), "application/json", unknown)
			if w.Code != 400 || !strings.Contains(w.Body.String(), "request body exceeds 8 MiB limit") {
				t.Fatalf("oversize %s unknown=%v: %d %s", route, unknown, w.Code, w.Body)
			}
		}
	}
	for _, route := range []string{"validate-fields", "validate-fields/batch"} {
		doc := `"document":{"text":"table host"}`
		if strings.HasSuffix(route, "/batch") {
			doc = `"documents":[{"text":"table host"}]`
		}
		base := []byte(`{` + doc + `,"catalog":["host"]}`)
		payload := append(base, bytes.Repeat([]byte(" "), (1<<20)+1-len(base))...)
		w := schemaREST(t, handler, route, payload, "application/json", true)
		if w.Code != 400 {
			t.Fatalf("legacy %s: %d", route, w.Code)
		}
	}
}

func readSchemaFixture(t *testing.T, variant string) json.RawMessage {
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

func TestSchemaRESTAcceptsOfficialCatalogs(t *testing.T) {
	handler := NewServer().Handler()
	for _, variant := range []string{"base", "windows"} {
		raw := readSchemaFixture(t, variant)
		selection := &validation.OCSFSelection{Version: "1.6.0", Class: "authentication", Profiles: []string{}, Extensions: []string{}}
		if variant == "windows" {
			selection.Class = "win/registry_key_activity"
			selection.Extensions = []string{"win"}
		}
		target := validation.SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: selection}
		document := analysis.QueryDocument{Text: "table time"}
		for _, batch := range []bool{false, true} {
			route := "validate-schema"
			var request, want any
			var err error
			if batch {
				route += "/batch"
				request = validation.SchemaBatchRequest{Documents: []analysis.QueryDocument{document}, Target: target}
				want, err = validation.ValidateSchemaBatch([]analysis.QueryDocument{document}, target)
			} else {
				request = validation.SchemaRequest{Document: document, Target: target}
				want, err = validation.ValidateSchema(document, target)
			}
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			if len(payload) <= 1<<20 || len(payload) >= 8<<20 {
				t.Fatalf("wrong real compact payload size: %d", len(payload))
			}
			t.Logf("%s %s compact bytes=%d raw catalog bytes=%d", variant, route, len(payload), len(raw))
			w := schemaREST(t, handler, route, payload, "application/json", false)
			if w.Code != 200 {
				t.Fatalf("%s %s: %d %s", variant, route, w.Code, w.Body)
			}
			var got any
			if batch {
				got = &validation.SchemaBatchReport{}
			} else {
				got = &validation.SchemaReport{}
			}
			if err = json.Unmarshal(w.Body.Bytes(), got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("report mismatch: %s", w.Body)
			}
			// Preserve the genuine, hash-checked raw catalog formatting through the route too.
			selectionJSON, err := json.Marshal(selection)
			if err != nil {
				t.Fatal(err)
			}
			docJSON, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			documentMember := `"document":` + string(docJSON)
			if batch {
				documentMember = `"documents":[` + string(docJSON) + `]`
			}
			rawPayload := []byte(`{` + documentMember + `,"target":{"kind":"ocsf","catalog":` + string(raw) + `,"selection":` + string(selectionJSON) + `}}`)
			if len(rawPayload) <= 1<<20 || len(rawPayload) >= 8<<20 {
				t.Fatalf("wrong raw payload size %d", len(rawPayload))
			}

			t.Logf("%s %s raw request bytes=%d", variant, route, len(rawPayload))
			w = schemaREST(t, handler, route, rawPayload, "application/json", false)
			if w.Code != 200 {
				t.Fatalf("raw catalog: %d %s", w.Code, w.Body)
			}
			if err = json.Unmarshal(w.Body.Bytes(), got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("raw report mismatch: %s", w.Body)
			}
		}
	}
}

func TestSchemaRESTConcurrentReports(t *testing.T) {
	handler := NewServer().Handler()
	for _, batch := range []bool{false, true} {
		route := "validate-schema"
		document := `"document":{"text":"table host"}`
		if batch {
			route += "/batch"
			document = `"documents":[{"text":"table host"},{"text":"table absent"}]`
		}
		payload := []byte(`{` + document + `,"target":{"kind":"json_schema","schema":{"properties":{"host":{}}}}}`)
		expected := schemaREST(t, handler, route, payload, "application/json", false)
		if expected.Code != 200 {
			t.Fatalf("%d %s", expected.Code, expected.Body)
		}
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w := schemaREST(t, handler, route, payload, "application/json", false)
				if w.Code != 200 || w.Body.String() != expected.Body.String() {
					t.Errorf("concurrent mismatch %d %s", w.Code, w.Body)
				}
			}()
		}
		wg.Wait()
	}
}
