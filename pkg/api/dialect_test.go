package api

import (
	"bytes"
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
	"net/http/httptest"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestDialectCapabilitiesRESTStrictSelectors(t *testing.T) {
	for _, query := range []string{"language=spl2&language=spl2", "language=spl&language=spl2", "profile=splunkd&profile=splunkd", "version=current&version=current", "extra=x", "compatibility-version=current", "language=unknown", "profile=cloud", "version=next", "language=%FF", "profile=%FF", "version=%FF", "language=%ZZ", "language=spl2;profile=splunkd"} {
		w := serveAnalysisRequest(t, "GET", "/api/v1/capabilities?"+query, nil, "")
		if w.Code != 400 {
			t.Fatalf("%q: %d %s", query, w.Code, w.Body)
		}
	}
}

func TestDialectAnalysisRESTStrictDocuments(t *testing.T) {
	for _, body := range []string{`{}`, `{"text":null}`, `{"text":"table host","language":null}`, `{"text":"table host","text":"table other"}`, `{"text":"table host","language":"spl","language":"spl2"}`, `{"Text":"table host"}`, `{"text":"table host","profile":null}`, `{"text":"table host","version":null}`, `{"text":"table host","source_id":null}`} {
		w := serveAnalysisRequest(t, "POST", "/api/v1/query/analyze", []byte(body), "application/json")
		if w.Code != 400 {
			t.Fatalf("%s: %d %s", body, w.Code, w.Body)
		}
	}
}

func TestDialectLegacyRESTSelectors(t *testing.T) {
	for _, operation := range []string{"map", "discover", "validate"} {
		base := map[string]any{"query": "search host=x"}
		if operation == "map" {
			base["mappings"] = []map[string]string{{"source": "host", "target": "server"}}
		}
		payload, _ := json.Marshal(base)
		want := serveAnalysisRequest(t, "POST", "/api/v1/query/"+operation, payload, "application/json")
		if want.Code != 200 {
			t.Fatal(want.Code, want.Body)
		}
		for _, selectors := range []map[string]string{{"language": "spl", "profile": "splunkd", "version": "current"}, {"language": "", "profile": "", "version": ""}, {"language": "spl2"}, {"language": "unknown"}, {"profile": "cloud"}, {"version": "next"}} {
			request := map[string]any{}
			for key, value := range base {
				request[key] = value
			}
			for key, value := range selectors {
				request[key] = value
			}
			data, _ := json.Marshal(request)
			w := serveAnalysisRequest(t, "POST", "/api/v1/query/"+operation, data, "application/json")
			if selectors["language"] == "spl" || len(selectors) == 3 {
				if w.Code != 200 || w.Body.String() != want.Body.String() {
					t.Fatalf("legacy response changed: %d %s", w.Code, w.Body)
				}
			} else {
				if w.Code != 400 {
					t.Fatalf("selectors %v: %d %s", selectors, w.Code, w.Body)
				}
				if selectors["language"] == "spl2" && (!strings.Contains(w.Body.String(), "unsupported_dialect_for_operation") || !strings.Contains(w.Body.String(), "analyze")) {
					t.Fatalf("missing rejection guidance: %s", w.Body)
				}
			}
		}
		for _, selector := range []string{`"language":"\ud800"`, `"profile":"\ud800"`, `"version":"\ud800"`} {
			body := []byte(`{"query":"search host=x",` + selector + `}`)
			w := serveAnalysisRequest(t, "POST", "/api/v1/query/"+operation, body, "application/json")
			if w.Code != 400 || strings.Contains(w.Body.String(), "�") {
				t.Fatalf("selector Unicode replaced: %d %s", w.Code, w.Body)
			}
		}
	}
}

func TestDialectRESTExactBodyLimits(t *testing.T) {
	for _, tc := range []struct {
		route, body string
		limit       int
	}{
		{"analyze", `{"text":"FROM main | table host","language":"spl2"}`, 1 << 20},
		{"validate-fields", `{"document":{"text":"FROM main | table host","language":"spl2"},"catalog":["host"]}`, 1 << 20},
		{"validate-fields/batch", `{"documents":[{"text":"table host"},{"text":"FROM main | table host","language":"spl2"}],"catalog":["host"]}`, 1 << 20},
		{"validate-schema", `{"document":{"text":"FROM main | table host","language":"spl2"},"target":{"kind":"json_schema","schema":{"properties":{"host":{}}}}}`, 8 << 20},
		{"validate-schema/batch", `{"documents":[{"text":"table host"},{"text":"FROM main | table host","language":"spl2"}],"target":{"kind":"json_schema","schema":{"properties":{"host":{}}}}}`, 8 << 20},
	} {
		for _, extra := range []int{0, 1} {
			body := append([]byte(tc.body), bytes.Repeat([]byte(" "), tc.limit-len(tc.body)+extra)...)
			r := httptest.NewRequest("POST", "/api/v1/query/"+tc.route, bytes.NewReader(body))
			r.ContentLength = -1
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			NewServer().Handler().ServeHTTP(w, r)
			expected := 200
			if extra == 1 {
				expected = 400
			}
			if w.Code != expected {
				t.Fatalf("%s bound+%d: %d %s", tc.route, extra, w.Code, w.Body)
			}
		}
	}
}

func TestDialectMaintainedAPIExamples(t *testing.T) {
	data, err := os.ReadFile("../../docs/api-server.md")
	if err != nil {
		t.Fatal(err)
	}
	pattern := regexp.MustCompile("(?s)<!-- api-example: ([a-z0-9-]+) (/query/[a-z/-]+) (valid|invalid|incomplete|legacy) -->\\n```json\\n(.*?)\\n```")
	examples := pattern.FindAllSubmatch(data, -1)
	if len(examples) != 14 || bytes.Count(data, []byte("```json\n")) != len(examples) {
		t.Fatalf("request marker coverage: %d", len(examples))
	}
	seen := map[string]bool{}
	for _, example := range examples {
		id, route, status, body := string(example[1]), string(example[2]), string(example[3]), example[4]
		t.Run(id, func(t *testing.T) {
			if seen[id] {
				t.Fatal("duplicate example ID")
			}
			seen[id] = true
			w := serveAnalysisRequest(t, "POST", "/api/v1"+route, body, "application/json")
			if w.Code != 200 {
				t.Fatalf("HTTP %d: %s", w.Code, w.Body)
			}
			var want any
			switch route {
			case "/query/map":
				want = MapQueryResponse{OriginalQuery: "search src_ip=1", MappedQuery: "search source_ip=1", Success: true}
			case "/query/analyze":
				var doc analysis.QueryDocument
				if err := json.Unmarshal(body, &doc); err != nil {
					t.Fatal(err)
				}
				report, e := analysis.Analyze(doc)
				if e != nil {
					t.Fatal(e)
				}
				want = report
			case "/query/validate-fields":
				request, e := validation.DecodeRequest(body)
				if e != nil {
					t.Fatal(e)
				}
				report, e := validation.Validate(request.Document, request.Catalog)
				if e != nil {
					t.Fatal(e)
				}
				want = report
			case "/query/validate-fields/batch":
				request, e := validation.DecodeBatchRequest(body)
				if e != nil {
					t.Fatal(e)
				}
				report, e := validation.ValidateBatch(request.Documents, request.Catalog)
				if e != nil {
					t.Fatal(e)
				}
				want = report
			case "/query/validate-schema":
				request, e := validation.DecodeSchemaRequest(body)
				if e != nil {
					t.Fatal(e)
				}
				report, e := validation.ValidateSchema(request.Document, request.Target)
				if e != nil {
					t.Fatal(e)
				}
				want = report
			case "/query/validate-schema/batch":
				request, e := validation.DecodeSchemaBatchRequest(body)
				if e != nil {
					t.Fatal(e)
				}
				report, e := validation.ValidateSchemaBatch(request.Documents, request.Target)
				if e != nil {
					t.Fatal(e)
				}
				want = report
			default:
				t.Fatalf("unhandled documented route %s", route)
			}
			expected, e := json.Marshal(want)
			if e != nil {
				t.Fatal(e)
			}
			var gotJSON, wantJSON any
			if e = json.Unmarshal(w.Body.Bytes(), &gotJSON); e != nil {
				t.Fatal(e)
			}
			if e = json.Unmarshal(expected, &wantJSON); e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(gotJSON, wantJSON) {
				t.Fatalf("canonical JSON mismatch: %s", w.Body)
			}
			if status != "legacy" && gotJSON.(map[string]any)["status"] != status {
				t.Fatalf("documented content status %s: %s", status, w.Body)
			}
		})
	}
	query := "/api/v1/capabilities?language=spl2&profile=splunkd&version=current"
	if !bytes.Contains(data, []byte(query)) {
		t.Fatal("capabilities request missing")
	}
	w := serveAnalysisRequest(t, "GET", query, nil, "")
	want, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: "spl2", Profile: "splunkd", Version: "current"})
	if err != nil {
		t.Fatal(err)
	}
	var manifest analysis.CapabilityManifest
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body)
	}
	if err = json.Unmarshal(w.Body.Bytes(), &manifest); err != nil || !reflect.DeepEqual(manifest, want) {
		t.Fatalf("capabilities JSON mismatch: %v %s", err, w.Body)
	}
}

func TestDialectOpenAPISelectorsAndMetadata(t *testing.T) {
	w := serveAnalysisRequest(t, "GET", "/api/v1/openapi.json", nil, "")
	var spec struct {
		Paths map[string]map[string]struct {
			Parameters []struct {
				Name   string `json:"name"`
				Schema struct {
					Enum []string `json:"enum"`
				} `json:"schema"`
			} `json:"parameters"`
		} `json:"paths"`
		Components struct {
			Schemas map[string]struct {
				Required   []string `json:"required"`
				Properties map[string]struct {
					Type string   `json:"type"`
					Enum []string `json:"enum"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	params := spec.Paths["/capabilities"]["get"].Parameters
	if len(params) != 3 {
		t.Fatalf("capability query parameters: %+v", params)
	}
	for _, p := range params {
		want := map[string][]string{"language": {"", "spl", "spl2"}, "profile": {"", "splunkd"}, "version": {"", "current"}}[p.Name]
		if !reflect.DeepEqual(p.Schema.Enum, want) {
			t.Fatalf("selector schema %+v", p)
		}
	}
	schemas := spec.Components.Schemas
	for _, name := range []string{"api.AnalysisRequest", "validation.QueryDocument"} {
		if !reflect.DeepEqual(schemas[name].Properties["language"].Enum, []string{"", "spl", "spl2"}) {
			t.Fatalf("missing strict SPL2 request schema: %s", name)
		}
	}
	for _, tc := range []struct{ schema, key, kind string }{{"analysis.CapabilityManifest", "documentation_snapshot", "string"}, {"analysis.Lineage", "phase", "string"}, {"analysis.Lineage", "execution_order", "integer"}} {
		if schemas[tc.schema].Properties[tc.key].Type != tc.kind {
			t.Fatalf("missing metadata %s.%s", tc.schema, tc.key)
		}
		for _, required := range schemas[tc.schema].Required {
			if required == tc.key {
				t.Fatalf("additive metadata became mandatory: %+v", tc)
			}
		}
	}
}

func TestDialectValidationRESTSelectorErrorsAreAtomic(t *testing.T) {
	for _, kind := range []string{"fields", "schema"} {
		target := `"catalog":["host"]`
		if kind == "schema" {
			target = `"target":{"kind":"json_schema","schema":{"properties":{"host":{}}}}`
		}
		for _, selector := range []string{`"language":"unknown"`, `"profile":"cloud"`, `"version":"next"`, `"language":null`, `"profile":null`, `"version":null`, `"source_id":null`, `"language":"\ud800"`, `"profile":"\ud800"`, `"version":"\ud800"`, `"source_id":"\ud800"`, `"language":"spl2","language":"spl"`} {
			doc := `{"text":"FROM main | table host",` + selector + `}`
			for _, batch := range []bool{false, true} {
				body := `{"document":` + doc + `,` + target + `}`
				route := "/api/v1/query/validate-" + kind
				if batch {
					body = `{"documents":[{"text":"table host"},` + doc + `],` + target + `}`
					route += "/batch"
				}
				w := serveAnalysisRequest(t, "POST", route, []byte(body), "application/json")
				if w.Code != 400 || strings.Contains(w.Body.String(), `"reports"`) {
					t.Fatalf("selector %s: HTTP %d %s", selector, w.Code, w.Body)
				}
			}
		}
		for _, selector := range []string{`"language":"spl2"`, `"profile":""`, `"version":"current"`, `"source_id":""`} {
			body := `{"documents":[{"text":"table host"}],` + target + `,` + selector + `}`
			w := serveAnalysisRequest(t, "POST", "/api/v1/query/validate-"+kind+"/batch", []byte(body), "application/json")
			if w.Code != 400 || strings.Contains(w.Body.String(), `"reports"`) {
				t.Fatalf("global batch selector: HTTP %d %s", w.Code, w.Body)
			}
		}
	}
}

func TestDialectLegacyRESTRejectsMalformedSelectorMembers(t *testing.T) {
	for _, operation := range []string{"map", "discover", "validate"} {
		for _, selector := range []string{`"language":null`, `"profile":null`, `"version":null`, `"language":"spl2","language":"spl"`, `"profile":"cloud","profile":""`, `"version":"next","version":"current"`, `"Language":"spl2","language":"spl"`} {
			body := `{"query":"search host=x","mappings":[{"source":"host","target":"server"}],` + selector + `}`
			if operation != "map" {
				body = `{"query":"search host=x",` + selector + `}`
			}
			w := serveAnalysisRequest(t, "POST", "/api/v1/query/"+operation, []byte(body), "application/json")
			if w.Code != 400 {
				t.Fatalf("%s %s: HTTP %d %s", operation, selector, w.Code, w.Body)
			}
		}
	}
}

func TestDialectLegacyRESTUnicodeFoldedSelectors(t *testing.T) {
	for _, operation := range []string{"map", "discover", "validate"} {
		t.Run(operation, func(t *testing.T) {
			body := `{"query":"search host=x"`
			if operation == "map" {
				body += `,"mappings":[{"source":"host","target":"server"}]`
			}
			route := "/api/v1/query/" + operation
			baseline := serveAnalysisRequest(t, "POST", route, []byte(body+`}`), "application/json")
			if baseline.Code != 200 {
				t.Fatalf("valid request control: HTTP %d %s", baseline.Code, baseline.Body)
			}
			for _, tc := range []struct {
				name, members, errorFragment string
			}{
				{"long-s null", `"verſion":null`, "expected a string"},
				{"escaped long-s null", `"ver\u017fion":null`, "expected a string"},
				{"folded duplicates", `"verſion":"next","verſion":"current"`, "duplicate property"},
				{"folded then canonical", `"verſion":"next","version":"current"`, "duplicate property"},
				{"canonical then folded", `"version":"next","VERſION":"current"`, "duplicate property"},
				{"equal duplicate aliases", `"verſion":"current","VERSION":"current"`, "duplicate property"},
				{"folded null then default", `"verſion":null,"version":""`, "duplicate property"},
				{"folded unsupported", `"verſion":"next"`, "unsupported compatibility version"},
				{"language null control", `"LANGUAGE":null`, "expected a string"},
				{"profile null control", `"PROFILE":null`, "expected a string"},
				{"folded current", `"verſion":"current"`, ""},
				{"folded empty default", `"verſion":""`, ""},
				{"mixed-case selectors", `"Language":"spl","ProFiLe":"splunkd","VERſION":"current"`, ""},
				{"empty folded selectors", `"LANGUAGE":"","PROFILE":"","ver\u017fion":""`, ""},
			} {
				t.Run(tc.name, func(t *testing.T) {
					response := serveAnalysisRequest(t, "POST", route, []byte(body+`,`+tc.members+`}`), "application/json")
					if tc.errorFragment == "" {
						if response.Code != 200 || response.Body.String() != baseline.Body.String() {
							t.Fatalf("folded/default selector changed legacy response: HTTP %d %s", response.Code, response.Body)
						}
						return
					}
					var rejection ErrorResponse
					if err := json.Unmarshal(response.Body.Bytes(), &rejection); err != nil {
						t.Fatal(err)
					}
					if response.Code != 400 || !rejection.Error || !strings.Contains(rejection.Message, tc.errorFragment) {
						t.Fatalf("selector %s: HTTP %d %s", tc.members, response.Code, response.Body)
					}
				})
			}
		})
	}
}
