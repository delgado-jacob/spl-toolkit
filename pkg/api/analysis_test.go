package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type analysisAPICorpusCase struct {
	ID       string                 `json:"id"`
	Document analysis.QueryDocument `json:"document"`
	Expected *analysis.Result       `json:"expected"`
}

func loadAnalysisAPICorpus(t *testing.T) []analysisAPICorpusCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/analysis/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Version string                  `json:"version"`
		Cases   []analysisAPICorpusCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	if corpus.Version != "1" || len(corpus.Cases) == 0 {
		t.Fatalf("missing reviewed corpus: version=%q cases=%d", corpus.Version, len(corpus.Cases))
	}
	return corpus.Cases
}

func serveAnalysisRequest(t *testing.T, method, path string, body []byte, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewReader(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response := httptest.NewRecorder()
	NewServer().Handler().ServeHTTP(response, request)
	return response
}

func TestAnalysisRESTReportsMatchCorpus(t *testing.T) {
	server := NewServer()
	for _, corpusCase := range loadAnalysisAPICorpus(t) {
		t.Run(corpusCase.ID, func(t *testing.T) {
			body, err := json.Marshal(corpusCase.Document)
			if err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPost, "/api/v1/query/analyze", bytes.NewReader(body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
			}
			var got analysis.Result
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, corpusCase.Expected) {
				t.Fatalf("report mismatch\ngot:  %#v\nwant: %#v", &got, corpusCase.Expected)
			}
		})
	}
}

func TestAnalysisRESTEmptyQueriesAreReports(t *testing.T) {
	for _, text := range []string{"", " \t\r\n"} {
		body, err := json.Marshal(analysis.QueryDocument{Text: text})
		if err != nil {
			t.Fatal(err)
		}
		response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/analyze", body, "application/json")
		if response.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
		}
		var report analysis.Result
		if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
			t.Fatal(err)
		}
		if report.Status != analysis.Invalid || report.Document.Text != text {
			t.Fatalf("unexpected report: %#v", report)
		}
	}
}

func TestAnalysisRESTRejectsDocumentAndJSONErrors(t *testing.T) {
	invalidText := append([]byte(`{"text":"`), 0xff)
	invalidText = append(invalidText, []byte(`"}`)...)
	invalidSource := append([]byte(`{"text":"search a=1","source_id":"`), 0xff)
	invalidSource = append(invalidSource, []byte(`"}`)...)
	overLimit := append([]byte(`{"text":"`), bytes.Repeat([]byte("a"), (1<<20)+1)...)
	overLimit = append(overLimit, []byte(`"}`)...)

	for _, test := range []struct {
		name        string
		body        []byte
		contentType string
	}{
		{name: "malformed", body: []byte(`{"text":`), contentType: "application/json"},
		{name: "null document", body: []byte(`null`), contentType: "application/json"},
		{name: "trailing document", body: []byte(`{"text":"search a=1"} {}`), contentType: "application/json"},
		{name: "unknown field", body: []byte(`{"text":"search a=1","extra":true}`), contentType: "application/json"},
		{name: "unsupported language", body: []byte(`{"text":"search a=1","language":"unknown"}`), contentType: "application/json"},
		{name: "unsupported profile", body: []byte(`{"text":"search a=1","profile":"cloud"}`), contentType: "application/json"},
		{name: "unsupported version", body: []byte(`{"text":"search a=1","version":"9.4"}`), contentType: "application/json"},
		{name: "invalid text UTF-8", body: invalidText, contentType: "application/json"},
		{name: "invalid source UTF-8", body: invalidSource, contentType: "application/json"},
		{name: "wrong content type", body: []byte(`{"text":"search a=1"}`), contentType: "text/plain"},
		{name: "missing content type", body: []byte(`{"text":"search a=1"}`)},
		{name: "body too large", body: overLimit, contentType: "application/json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/analyze", test.body, test.contentType)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
			}
			var apiError ErrorResponse
			if err := json.Unmarshal(response.Body.Bytes(), &apiError); err != nil || !apiError.Error {
				t.Fatalf("error response=%#v decode=%v body=%s", apiError, err, response.Body.Bytes())
			}
		})
	}
}

func TestAnalysisRESTRejectsSurrogateEscapesBeforeDecoderReplacement(t *testing.T) {
	for _, test := range []struct {
		name string
		body []byte
	}{
		{name: "lone high surrogate", body: []byte(`{"text":"\ud800"}`)},
		{name: "high surrogate without low", body: []byte(`{"text":"\ud800x"}`)},
		{name: "high followed by ordinary escape", body: []byte(`{"text":"\ud800\u0061"}`)},
		{name: "lone low surrogate", body: []byte(`{"text":"search a=1","source_id":"\udc00"}`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/analyze", test.body, "application/json")
			if response.Code != http.StatusBadRequest {
				t.Fatalf("surrogate escape reached decoding: status=%d body=%s", response.Code, response.Body.Bytes())
			}
			if bytes.Contains(response.Body.Bytes(), []byte("�")) {
				t.Fatalf("response contains a decoder replacement: %s", response.Body.Bytes())
			}
		})
	}
}

func TestAnalysisRESTPreservesValidUnicodeEscapes(t *testing.T) {
	body := []byte(`{"text":"search a=\"\ud83d\ude00\"","source_id":"source:\\ud800"}`)
	response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/analyze", body, "application/json")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
	}
	var report analysis.Result
	if err := json.Unmarshal(response.Body.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Document.Text != "search a=\"😀\"" || report.Document.SourceID != `source:\ud800` {
		t.Fatalf("decoded source was not preserved: %#v", report.Document)
	}
	if strings.Contains(report.Document.Text, "�") || strings.Contains(report.Document.SourceID, "�") {
		t.Fatalf("decoder replacement escaped validation: %#v", report.Document)
	}
}

func TestCapabilitiesRESTMatchesKernel(t *testing.T) {
	response := serveAnalysisRequest(t, http.MethodGet, "/api/v1/capabilities", nil, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
	}
	var got analysis.CapabilityManifest
	if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if want := analysis.Capabilities(); !reflect.DeepEqual(got, want) {
		t.Fatalf("manifest mismatch\ngot:  %#v\nwant: %#v", got, want)
	}
	if got.ToolkitVersion == "" || len(got.Records) == 0 || len(got.Evidence) == 0 {
		t.Fatalf("incomplete manifest: toolkit=%q records=%d evidence=%d", got.ToolkitVersion, len(got.Records), len(got.Evidence))
	}
	got.Records[0].Dimensions.Syntax.EvidenceIDs = append(got.Records[0].Dimensions.Syntax.EvidenceIDs, "mutated")
	got.Evidence[0].ID = "mutated"
	response = serveAnalysisRequest(t, http.MethodGet, "/api/v1/capabilities", nil, "")
	var fresh analysis.CapabilityManifest
	if response.Code != http.StatusOK {
		t.Fatalf("fresh status=%d body=%s", response.Code, response.Body.Bytes())
	}
	if err := json.Unmarshal(response.Body.Bytes(), &fresh); err != nil {
		t.Fatal(err)
	}
	if want := analysis.Capabilities(); !reflect.DeepEqual(fresh, want) {
		t.Fatalf("manifest mutation escaped into a later response\ngot:  %#v\nwant: %#v", fresh, want)
	}
}

func TestAnalysisOpenAPIDocument(t *testing.T) {
	response := serveAnalysisRequest(t, http.MethodGet, "/api/v1/openapi.json", nil, "")
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
	}
	var spec struct {
		Paths      map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]struct {
				Properties map[string]struct {
					Type string `json:"type"`
				} `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/query/analyze", "/capabilities"} {
		if _, ok := spec.Paths[path]; !ok {
			t.Errorf("OpenAPI missing %s", path)
		}
	}
	result, ok := spec.Components.Schemas["analysis.Result"]
	if !ok {
		t.Fatalf("OpenAPI missing analysis.Result; schemas=%v", reflect.ValueOf(spec.Components.Schemas).MapKeys())
	}
	if got := result.Properties["schema_version"].Type; got != "integer" {
		t.Errorf("analysis.Result schema_version type=%q", got)
	}
	if raw := string(spec.Paths["/query/analyze"]); !strings.Contains(raw, "analysis.Result") {
		t.Errorf("analyze response does not reference canonical report: %s", raw)
	}
}

func TestRequirementsRESTContentStatuses(t *testing.T) {
	for _, test := range []struct {
		document analysis.QueryDocument
		status   analysis.Status
	}{
		{document: analysis.QueryDocument{Text: "search host=web", Language: "spl", SourceID: "valid-spl"}, status: analysis.Valid},
		{document: analysis.QueryDocument{Text: "search host=* | fields - host | table host", Language: "spl", SourceID: "invalid-spl"}, status: analysis.Invalid},
		{document: analysis.QueryDocument{Text: "search host=web | mystery", Language: "spl", SourceID: "incomplete-spl"}, status: analysis.Incomplete},
		{document: analysis.QueryDocument{Text: "FROM main | table host", Language: "spl2", SourceID: "valid-spl2"}, status: analysis.Valid},
		{document: analysis.QueryDocument{Text: "FROM [{a:1}] | table b", Language: "spl2", SourceID: "invalid-spl2"}, status: analysis.Invalid},
		{document: analysis.QueryDocument{Text: "FROM main | mystery", Language: "spl2", SourceID: "incomplete-spl2"}, status: analysis.Incomplete},
	} {
		test := test
		document := test.document
		t.Run(document.SourceID, func(t *testing.T) {
			body, err := json.Marshal(document)
			if err != nil {
				t.Fatal(err)
			}
			response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/requirements", body, "application/json; charset=utf-8")
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
			}
			var got analysis.RequirementSet
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			want, err := analysis.Requirements(document)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("requirement set mismatch\ngot:  %#v\nwant: %#v", &got, want)
			}
			if got.QueryStatus != test.status {
				t.Fatalf("query status=%q, want %q", got.QueryStatus, test.status)
			}
		})
	}
}

func TestRequirementsRESTRejectsRequestBoundaryErrors(t *testing.T) {
	invalidText := append([]byte(`{"text":"`), 0xff)
	invalidText = append(invalidText, []byte(`"}`)...)
	overLimit := append([]byte(`{"text":"`), bytes.Repeat([]byte("a"), (1<<20)+1)...)
	overLimit = append(overLimit, []byte(`"}`)...)

	for _, test := range []struct {
		name        string
		body        []byte
		contentType string
	}{
		{name: "duplicate member", body: []byte(`{"text":"search a=1","text":"search b=2"}`), contentType: "application/json"},
		{name: "unknown member", body: []byte(`{"text":"search a=1","extra":true}`), contentType: "application/json"},
		{name: "file member", body: []byte(`{"text":"search a=1","file":"query.spl"}`), contentType: "application/json"},
		{name: "network member", body: []byte(`{"text":"search a=1","url":"https://example.invalid/query.spl"}`), contentType: "application/json"},
		{name: "empty", body: nil, contentType: "application/json"},
		{name: "malformed", body: []byte(`{"text":`), contentType: "application/json"},
		{name: "null", body: []byte(`null`), contentType: "application/json"},
		{name: "invalid unicode", body: invalidText, contentType: "application/json"},
		{name: "wrong content type", body: []byte(`{"text":"search a=1"}`), contentType: "text/plain"},
		{name: "missing content type", body: []byte(`{"text":"search a=1"}`)},
		{name: "trailing JSON", body: []byte(`{"text":"search a=1"} {}`), contentType: "application/json"},
		{name: "unsupported language", body: []byte(`{"text":"search a=1","language":"unknown"}`), contentType: "application/json"},
		{name: "unsupported profile", body: []byte(`{"text":"search a=1","profile":"cloud"}`), contentType: "application/json"},
		{name: "unsupported version", body: []byte(`{"text":"search a=1","version":"9.4"}`), contentType: "application/json"},
		{name: "body too large", body: overLimit, contentType: "application/json"},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/requirements", test.body, test.contentType)
			if response.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.Bytes())
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode error response: %v\n%s", err, response.Body.Bytes())
			}
			if _, ok := got["error"]; !ok {
				t.Fatalf("missing error response: %s", response.Body.Bytes())
			}
			for _, key := range []string{"schema_version", "query", "query_status", "coverage", "items", "gaps", "diagnostics"} {
				if _, ok := got[key]; ok {
					t.Fatalf("partial requirement report contains %q: %s", key, response.Body.Bytes())
				}
			}
		})
	}
}

func TestRequirementsRESTLexerBudgetParity(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		language := language
		t.Run(language, func(t *testing.T) {
			exact := strings.Repeat("x ", 2048)
			for _, test := range []struct {
				name          string
				text          string
				resourceLimit bool
			}{
				{name: "exact", text: exact},
				{name: "plus one", text: exact + "x", resourceLimit: true},
				{name: "long sparse", text: `search note="` + strings.Repeat("a", 64*1024) + `"`},
			} {
				t.Run(test.name, func(t *testing.T) {
					document := analysis.QueryDocument{Text: test.text, Language: language, Profile: "splunkd", Version: "current", SourceID: "adapter-budget"}
					body, err := json.Marshal(document)
					if err != nil {
						t.Fatal(err)
					}

					for _, operation := range []string{"analyze", "requirements"} {
						response := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/"+operation, body, "application/json")
						if response.Code != http.StatusOK {
							t.Fatalf("%s status=%d body=%s", operation, response.Code, response.Body.Bytes())
						}
						if operation == "analyze" {
							var got analysis.Result
							if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
								t.Fatal(err)
							}
							want, err := analysis.Analyze(document)
							if err != nil || !reflect.DeepEqual(&got, want) {
								t.Fatalf("analyze parity: err=%v", err)
							}
							gotLimit := len(got.Diagnostics) == 1 && got.Diagnostics[0].Code == analysis.CodeAnalysisResourceLimit
							if gotLimit != test.resourceLimit {
								t.Fatalf("analyze resource limit=%t, want %t", gotLimit, test.resourceLimit)
							}
							continue
						}

						var got analysis.RequirementSet
						if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
							t.Fatal(err)
						}
						want, err := analysis.Requirements(document)
						if err != nil || !reflect.DeepEqual(&got, want) {
							t.Fatalf("requirements parity: err=%v", err)
						}
						gotLimit := len(got.Diagnostics) == 1 && got.Diagnostics[0].Code == analysis.CodeAnalysisResourceLimit
						if gotLimit != test.resourceLimit {
							t.Fatalf("requirements resource limit=%t, want %t", gotLimit, test.resourceLimit)
						}
					}
				})
			}
		})
	}
}
