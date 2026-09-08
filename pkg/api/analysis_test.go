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
