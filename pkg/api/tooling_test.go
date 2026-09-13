package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/document"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)

func TestToolingHTTPBounds(t *testing.T) {
	h := NewServer().Handler()
	baseRequest := `{"schema_version":1,"documents":[{"id":"one","document":{"text":"search host=x"}}]}`
	target := `{"kind":"field_list","catalog":{"fields":["host"]}}`
	requests := map[string]string{"corpus/scan": baseRequest, "corpus/graph": baseRequest, "corpus/sarif": baseRequest,
		"query/document":        `{"text":"search host=x"}`,
		"corpus/impact-schema":  strings.TrimSuffix(baseRequest, "}") + `,"before_target":` + target + `,"after_target":` + target + `}`,
		"corpus/impact-mapping": strings.TrimSuffix(baseRequest, "}") + `,"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]}}`}
	for route, raw := range requests {
		base := []byte(raw)
		for _, unknown := range []bool{false, true} {
			for _, extra := range []int{0, 1} {
				body := append(append([]byte{}, base...), bytes.Repeat([]byte(" "), (8<<20)-len(base)+extra)...)
				r := httptest.NewRequest(http.MethodPost, "/api/v1/"+route, bytes.NewReader(body))
				r.Header.Set("Content-Type", "application/json")
				if unknown {
					r.ContentLength = -1
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				want := 200
				if extra > 0 {
					want = 400
				}
				if w.Code != want {
					t.Fatalf("%s unknown %t extra %d: %d %s", route, unknown, extra, w.Code, w.Body)
				}
			}
		}
	}
}

func toolingHTTP(t *testing.T, route string, body []byte) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "/api/v1/"+route, bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewServer().Handler().ServeHTTP(w, r)
	return w
}

func TestToolingHTTPCanonicalParity(t *testing.T) {
	docs := `[{"id":"a","document":{"text":"search host=\"😀\"\r\n| table host","source_id":"a"}},{"id":"b","document":{"text":"| mystery | table host"}}]`
	base := `{"schema_version":1,"documents":` + docs + `}`
	target := `{"kind":"field_list","catalog":{"fields":["host"]}}`
	requests := map[string]string{"corpus/scan": base, "corpus/graph": base, "corpus/sarif": base, "corpus/impact-schema": `{"schema_version":1,"documents":` + docs + `,"before_target":` + target + `,"after_target":{"kind":"field_list","catalog":{"fields":[]}}}`, "corpus/impact-mapping": `{"schema_version":1,"documents":` + docs + `,"before_rules":{"schema_version":1,"rules":[]},"after_rules":{"schema_version":1,"rules":[]}}`, "query/document": `{"text":"search host=\"😀\"\r\n| table host","source_id":"a"}`}
	for route, raw := range requests {
		var want any
		var err error
		switch route {
		case "corpus/scan", "corpus/graph", "corpus/sarif":
			r, e := corpus.DecodeRequest([]byte(raw))
			if e != nil {
				t.Fatal(e)
			}
			report, e := corpus.Scan(r)
			err = e
			want = report
			if route == "corpus/graph" {
				want, err = graph.Export(report)
			} else if route == "corpus/sarif" {
				want, err = sarif.Export(report)
			}
		case "corpus/impact-schema":
			r, e := impact.DecodeSchemaRequest([]byte(raw))
			if e != nil {
				t.Fatal(e)
			}
			want, err = impact.CompareSchemas(r)
		case "corpus/impact-mapping":
			r, e := impact.DecodeMappingRequest([]byte(raw))
			if e != nil {
				t.Fatal(e)
			}
			want, err = impact.CompareMappings(r)
		case "query/document":
			var d analysis.QueryDocument
			if err = json.Unmarshal([]byte(raw), &d); err != nil {
				t.Fatal(err)
			}
			r, e := analysis.Analyze(d)
			if e != nil {
				t.Fatal(e)
			}
			want, err = document.New(r, document.RevisionContext{ToolVersion: buildinfo.Version, ContractVersion: "1"})
		}
		if err != nil {
			t.Fatal(err)
		}
		expected, _ := json.Marshal(want)
		w := toolingHTTP(t, route, []byte(raw))
		if w.Code != 200 || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), expected) {
			t.Fatalf("%s parity %d %s", route, w.Code, w.Body)
		}
		for _, bad := range []string{"null", raw + " {}", strings.Replace(raw, "{", `{"path":"/etc/passwd",`, 1), strings.Replace(raw, "{", `{"directory":"/",`, 1)} {
			w = toolingHTTP(t, route, []byte(bad))
			if w.Code != 400 {
				t.Fatalf("%s bad request: %d", route, w.Code)
			}
		}
	}
}

func TestToolingHTTPTwoGenuineCatalogs(t *testing.T) {
	before := map[string]any{"kind": "ocsf", "catalog": readSchemaFixture(t, "base"), "selection": map[string]any{"version": "1.6.0", "class": "authentication"}}
	after := map[string]any{"kind": "ocsf", "catalog": readSchemaFixture(t, "windows"), "selection": map[string]any{"version": "1.6.0", "class": "win/registry_key_activity", "extensions": []string{"win"}}}
	raw, err := json.Marshal(map[string]any{"schema_version": 1, "documents": []any{map[string]any{"id": "one", "document": map[string]any{"text": "table actor.user.name"}}}, "before_target": before, "after_target": after})
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) > 8<<20 {
		t.Fatalf("two genuine catalogs exceed bound: %d", len(raw))
	}
	request, err := impact.DecodeSchemaRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	want, err := impact.CompareSchemas(request)
	if err != nil {
		t.Fatal(err)
	}
	expected, _ := json.Marshal(want)
	w := toolingHTTP(t, "corpus/impact-schema", raw)
	if w.Code != 200 || !bytes.Equal(bytes.TrimSpace(w.Body.Bytes()), expected) {
		t.Fatalf("catalog parity: %d %s", w.Code, w.Body)
	}
}
