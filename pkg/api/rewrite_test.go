package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

const apiRewriteRule = `{"id":"rename","kind":"field","source":{"name":"src"},"target":{"name":"user"}}`

func TestRewriteRESTCanonicalSingleAndBatchReports(t *testing.T) {
	handler := NewServer().Handler()
	for _, batch := range []bool{false, true} {
		route, key, value := "rewrite", "document", `{"text":"search src=x","source_id":"one"}`
		if batch {
			route, key, value = "rewrite/batch", "documents", `[{"text":"search src=x","source_id":"one"},{"text":"| mystery src","source_id":"two"}]`
		}
		payload := []byte(`{"schema_version":1,"mode":"apply","` + key + `":` + value + `,"rules":[` + apiRewriteRule + `]}`)
		decodedSingle, decodedBatch, err := rewrite.Request{}, rewrite.BatchRequest{}, error(nil)
		var want any
		if batch {
			decodedBatch, err = rewrite.DecodeBatchRequest(payload)
			if err == nil {
				want, err = rewrite.RewriteBatch(decodedBatch)
			}
		} else {
			decodedSingle, err = rewrite.DecodeRequest(payload)
			if err == nil {
				want, err = rewrite.Rewrite(decodedSingle)
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		response := schemaREST(t, handler, route, payload, "application/json; charset=utf-8", false)
		if response.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", route, response.Code, response.Body)
		}
		expected, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(bytes.TrimSpace(response.Body.Bytes()), expected) {
			t.Fatalf("%s parity mismatch", route)
		}
	}
}

func TestRewriteRESTStrictRequestsAndBodyLimit(t *testing.T) {
	handler := NewServer().Handler()
	valid := `{"schema_version":1,"document":{"text":"search src=x"},"rules":[` + apiRewriteRule + `]}`
	for _, bad := range []string{`null`, `[]`, `{}`, valid + ` {}`, strings.Replace(valid, `"rules":`, `"extra":1,"rules":`, 1), strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1), strings.Replace(valid, `"document":{`, `"document":null,"unused":{`, 1), strings.Replace(valid, `"rules":[`, `"rules":null,"unused":[`, 1)} {
		response := schemaREST(t, handler, "rewrite", []byte(bad), "application/json", false)
		if response.Code != http.StatusBadRequest {
			t.Errorf("bad=%s status=%d body=%s", bad, response.Code, response.Body)
		}
	}
	for _, contentType := range []string{"", "text/plain", "application/json; malformed"} {
		response := schemaREST(t, handler, "rewrite", []byte(valid), contentType, false)
		if response.Code != http.StatusBadRequest {
			t.Errorf("content-type=%q status=%d", contentType, response.Code)
		}
	}
	base := []byte(valid)
	exact := append(base, bytes.Repeat([]byte(" "), (8<<20)-len(base))...)
	if response := schemaREST(t, handler, "rewrite", exact, "application/json", false); response.Code != http.StatusOK {
		t.Fatalf("exact limit: %d %s", response.Code, response.Body)
	}
	for _, unknown := range []bool{false, true} {
		response := schemaREST(t, handler, "rewrite", append(exact, ' '), "application/json", unknown)
		if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "request body exceeds 8 MiB limit") {
			t.Fatalf("overflow unknown=%v: %d %s", unknown, response.Code, response.Body)
		}
	}
}

func TestRewriteRESTBatchKnownLengthBodyLimit(t *testing.T) {
	handler := NewServer().Handler()
	base := []byte(`{"schema_version":1,"documents":[{"text":"search src=x"}],"rules":[` + apiRewriteRule + `]}`)
	exact := append(base, bytes.Repeat([]byte(" "), (8<<20)-len(base))...)
	if response := schemaREST(t, handler, "rewrite/batch", exact, "application/json", false); response.Code != http.StatusOK {
		t.Fatalf("exact known-length batch limit: %d %s", response.Code, response.Body)
	}
	response := schemaREST(t, handler, "rewrite/batch", append(exact, ' '), "application/json", false)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "request body exceeds 8 MiB limit") {
		t.Fatalf("overflow known-length batch limit: %d %s", response.Code, response.Body)
	}
}

func TestRewriteRESTRejectsMalformedTargetsAndBatchWrappers(t *testing.T) {
	for _, batch := range []bool{false, true} {
		route, key, document := "rewrite", "document", `{"text":"search src=x"}`
		if batch {
			route, key, document = "rewrite/batch", "documents", `[{"text":"search src=x"}]`
		}
		for _, target := range []string{`null`, `{}`, `{"kind":"field_list"}`, `{"kind":"field_list","catalog":null}`, `{"kind":"field_list","catalog":["x"],"path":"x"}`, `{"kind":"json_schema","schema":null}`, `{"kind":"ocsf","catalog":{},"selection":null}`} {
			payload := []byte(`{"schema_version":1,"` + key + `":` + document + `,"rules":[],"validation_target":` + target + `}`)
			response := schemaREST(t, NewServer().Handler(), route, payload, "application/json", false)
			if response.Code != http.StatusBadRequest || strings.Contains(response.Body.String(), `"reports"`) {
				t.Errorf("%s target=%s: %d %s", route, target, response.Code, response.Body)
			}
		}
		payload := []byte(`{"schema_version":1,"` + key + `":` + document + `,"rules":[],"rules":[]}`)
		if response := schemaREST(t, NewServer().Handler(), route, payload, "application/json", false); response.Code != http.StatusBadRequest {
			t.Errorf("%s duplicate: %d", route, response.Code)
		}
	}
}

func TestWriteRewriteErrorClassifiesInputAndInternalFailures(t *testing.T) {
	s := NewServer()
	for _, tc := range []struct {
		err  error
		want int
	}{{errors.New("boom"), 500}, {&rewrite.InputError{Err: errors.New("bad")}, 400}} {
		w := httptest.NewRecorder()
		s.writeRewriteError(w, tc.err)
		if w.Code != tc.want {
			t.Fatalf("err=%v status=%d body=%s", tc.err, w.Code, w.Body)
		}
	}
}

func TestRewriteRESTPreservesQueryStatusesAtHTTP200(t *testing.T) {
	for _, document := range []analysis.QueryDocument{{Text: "|"}, {Text: "| mystery src"}, {Text: "search src=x"}} {
		documentJSON, err := json.Marshal(document)
		if err != nil {
			t.Fatal(err)
		}
		payload := []byte(`{"schema_version":1,"document":` + string(documentJSON) + `,"rules":[]}`)
		response := schemaREST(t, NewServer().Handler(), "rewrite", payload, "application/json", false)
		if response.Code != http.StatusOK {
			t.Fatalf("document=%q status=%d body=%s", document.Text, response.Code, response.Body)
		}
		var got rewrite.Result
		if err = json.Unmarshal(response.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		request, err := rewrite.DecodeRequest(payload)
		if err != nil {
			t.Fatal(err)
		}
		want, err := rewrite.Rewrite(request)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(&got, want) {
			t.Fatalf("document=%q parity mismatch", document.Text)
		}
	}
}

func TestRewriteRESTAcceptsOfficialOCSFCatalogs(t *testing.T) {
	for _, variant := range []string{"base", "windows"} {
		catalog := readSchemaFixture(t, variant)
		class := "authentication"
		extensions := `[]`
		if variant == "windows" {
			class, extensions = "win/registry_key_activity", `["win"]`
		}
		target := `{"kind":"ocsf","catalog":` + string(catalog) + `,"selection":{"version":"1.6.0","class":"` + class + `","profiles":[],"extensions":` + extensions + `}}`
		for _, batch := range []bool{false, true} {
			route, key, value := "rewrite", "document", `{"text":"table time"}`
			if batch {
				route, key, value = "rewrite/batch", "documents", `[{"text":"table time"}]`
			}
			payload := []byte(`{"schema_version":1,"` + key + `":` + value + `,"rules":[],"validation_target":` + target + `}`)
			if len(payload) <= 1<<20 || len(payload) >= 8<<20 {
				t.Fatalf("%s payload size=%d", variant, len(payload))
			}
			response := schemaREST(t, NewServer().Handler(), route, payload, "application/json", false)
			if response.Code != http.StatusOK {
				t.Fatalf("%s %s: %d %s", variant, route, response.Code, response.Body)
			}
		}
	}
}
