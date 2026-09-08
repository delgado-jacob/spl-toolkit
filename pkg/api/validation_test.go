package api

import (
	"bytes"
	"encoding/json"
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

func TestValidateFieldsRESTCanonicalReports(t *testing.T) {
	server := NewServer()
	for _, query := range []string{"search missing=x", "search host=x\n", "| mystery", "", "search host=\"é😀\"\r\n"} {
		for _, catalog := range []string{`["host"]`, `{"fields":["host"],"optional_fields":["user"],"identity":"local","version":"v1"}`} {
			doc := analysis.QueryDocument{Text: query, SourceID: "original"}
			raw, err := json.Marshal(doc)
			if err != nil {
				t.Fatal(err)
			}
			body := `{"document":` + string(raw) + `,"catalog":` + catalog + `}`
			request := httptest.NewRequest(http.MethodPost, "/api/v1/query/validate-fields", strings.NewReader(body))
			request.Header.Set("Content-Type", "application/json; charset=utf-8")
			response := httptest.NewRecorder()
			server.Handler().ServeHTTP(response, request)
			if response.Code != 200 {
				t.Fatalf("status=%d body=%s", response.Code, response.Body)
			}
			var got validation.Report
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			cat, err := validation.DecodeFieldCatalog([]byte(catalog))
			if err != nil {
				t.Fatal(err)
			}
			want, err := validation.Validate(doc, cat)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(&got, want) {
				t.Fatalf("canonical report mismatch: %s", response.Body)
			}
		}
	}
}

func TestValidateFieldsRESTBatchAndIsolation(t *testing.T) {
	server := NewServer()
	var wg sync.WaitGroup
	for _, catalog := range []string{`["host"]`, `[]`} {
		wg.Add(1)
		go func(catalog string) {
			defer wg.Done()
			body := `{"documents":[{"text":"search host=x\n","source_id":"first"},{"text":"| mystery","source_id":"second"}],"catalog":` + catalog + `}`
			for i := 0; i < 3; i++ {
				request := httptest.NewRequest(http.MethodPost, "/api/v1/query/validate-fields/batch", strings.NewReader(body))
				request.Header.Set("Content-Type", "application/json")
				response := httptest.NewRecorder()
				server.Handler().ServeHTTP(response, request)
				if response.Code != 200 {
					t.Errorf("status=%d body=%s", response.Code, response.Body)
					return
				}
				var got validation.BatchReport
				if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
					t.Error(err)
					return
				}
				req, err := validation.DecodeBatchRequest([]byte(body))
				if err != nil {
					t.Error(err)
					return
				}
				want, err := validation.ValidateBatch(req.Documents, req.Catalog)
				if err != nil {
					t.Error(err)
					return
				}
				if !reflect.DeepEqual(&got, want) {
					t.Errorf("batch mismatch: %s", response.Body)
				}
			}
		}(catalog)
	}
	wg.Wait()
}

func TestValidateFieldsRESTInputErrors(t *testing.T) {
	for _, route := range []string{"/api/v1/query/validate-fields", "/api/v1/query/validate-fields/batch"} {
		valid := `{"document":{"text":"search host=x"},"catalog":["host"]}`
		if strings.HasSuffix(route, "/batch") {
			valid = `{"documents":[{"text":"search host=x"}],"catalog":["host"]}`
		}
		for _, body := range []string{`null`, `{}`, `{"catalog":[]}`, valid + ` {}`, strings.Replace(valid, `"catalog":["host"]`, `"catalog":null`, 1), strings.Replace(valid, `"catalog":["host"]`, `"catalog":[],"catalog":[]`, 1), strings.Replace(valid, `"catalog":["host"]`, `"catalog":{"fields":[],"url":"https://example.com"}`, 1), strings.Replace(valid, `"text":"search host=x"`, `"text":null`, 1), strings.Replace(valid, `"text":"search host=x"`, `"text":"x","text":"y"`, 1), strings.Replace(valid, `"text":"search host=x"`, `"text":"x","unknown":true`, 1), strings.Replace(valid, `"text":"search host=x"`, `"text":"x","language":"unknown"`, 1), strings.Replace(valid, `"text":"search host=x"`, `"text":"\ud800"`, 1), strings.Replace(valid, `search host=x`, string([]byte{255}), 1), strings.Replace(valid, `["host"]`, `["host","host"]`, 1), strings.Replace(valid, `"catalog":`, `"unknown":`, 1)} {
			response := serveAnalysisRequest(t, http.MethodPost, route, []byte(body), "application/json")
			if response.Code != 400 {
				t.Errorf("route=%s input=%s status=%d body=%s", route, body, response.Code, response.Body)
			}
		}
		for _, ct := range []string{"", "text/plain"} {
			if r := serveAnalysisRequest(t, http.MethodPost, route, []byte(valid), ct); r.Code != 400 {
				t.Errorf("content type %q status=%d", ct, r.Code)
			}
		}
		over := bytes.Repeat([]byte(" "), 1<<20)
		over = append(over, []byte(valid)...)
		if r := serveAnalysisRequest(t, http.MethodPost, route, over, "application/json"); r.Code != 400 {
			t.Errorf("body limit status=%d", r.Code)
		}
		if r := serveAnalysisRequest(t, http.MethodGet, route, nil, ""); r.Code != 405 {
			t.Errorf("method status=%d", r.Code)
		}
	}
	for _, body := range []string{`{"documents":[],"catalog":[]}`, `{"documents":null,"catalog":[]}`} {
		if r := serveAnalysisRequest(t, http.MethodPost, "/api/v1/query/validate-fields/batch", []byte(body), "application/json"); r.Code != 400 {
			t.Errorf("empty batch status=%d", r.Code)
		}
	}
}

func TestValidateFieldsOpenAPICatalogShapes(t *testing.T) {
	response := serveAnalysisRequest(t, http.MethodGet, "/api/v1/openapi.json", nil, "")
	var spec struct {
		Paths      map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &spec); err != nil {
		t.Fatal(err)
	}
	static, err := os.ReadFile("../../docs/swagger.json")
	if err != nil {
		t.Fatal(err)
	}
	var servedSpec, staticSpec map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &servedSpec); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(static, &staticSpec); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(servedSpec["components"], staticSpec["components"]) {
		t.Fatal("served and generated schemas differ")
	}
	for _, name := range []string{"validation.Request", "validation.BatchRequest", "validation.QueryDocument", "validation.FieldCatalog"} {
		var schema map[string]any
		if err := json.Unmarshal(spec.Components.Schemas[name], &schema); err != nil {
			t.Fatal(err)
		}
		if schema["additionalProperties"] != false {
			t.Errorf("input schema %s is not strict", name)
		}
	}
	for _, route := range []string{"/query/validate-fields", "/query/validate-fields/batch"} {
		if _, ok := spec.Paths[route]; !ok {
			t.Errorf("missing route %s", route)
		}
	}
	for _, name := range []string{"validation.Request", "validation.BatchRequest"} {
		var schema struct {
			Required   []string `json:"required"`
			Properties map[string]struct {
				OneOf []json.RawMessage `json:"oneOf"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(spec.Components.Schemas[name], &schema); err != nil {
			t.Errorf("missing schema %s: %v", name, err)
			continue
		}
		if len(schema.Required) != 2 || len(schema.Properties["catalog"].OneOf) != 2 {
			t.Errorf("schema %s does not require documents/catalog or permit both catalog shapes: %s", name, spec.Components.Schemas[name])
		}
	}
	for _, name := range []string{"validation.Report", "validation.BatchReport"} {
		var schema struct {
			Properties map[string]struct {
				Type string `json:"type"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(spec.Components.Schemas[name], &schema); err != nil {
			t.Errorf("missing report %s: %v", name, err)
			continue
		}
		if schema.Properties["schema_version"].Type != "integer" {
			t.Errorf("noninteger schema version %s", name)
		}
	}
}
