package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
)

func workflowHTTPSeed(t *testing.T, texts ...string) workflow.Request {
	t.Helper()
	raw, err := os.ReadFile("../../examples/resolution/request.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := resolution.DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := workflow.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{}, Settings: workflow.Settings{SchemaVersion: 1, Snapshot: seed.Compatibility.Snapshot, SchemaBundle: seed.Compatibility.SchemaBundle, Entries: []workflow.EntrySettings{}}}
	for i, text := range texts {
		id := string(rune('a' + i))
		q.Documents = append(q.Documents, corpus.RequestDocument{ID: id, Document: analysis.QueryDocument{Text: text, Language: "spl2", SourceID: id}})
		q.Settings.Entries = append(q.Settings.Entries, workflow.EntrySettings{ID: id, Compatibility: &workflow.CheckSettings{QueryScope: seed.Compatibility.QueryScope, InputBindings: []compatibility.InputBinding{}}})
	}
	return q
}
func workflowHTTPJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func workflowHTTPPost(t *testing.T, server *httptest.Server, route string, raw []byte, mime string, chunked bool) (int, string, []byte) {
	t.Helper()
	var reader io.Reader = bytes.NewReader(raw)
	if chunked {
		reader = struct{ io.Reader }{reader}
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/workflow/"+route, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", mime)
	response, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	return response.StatusCode, response.Header.Get("Content-Type"), body
}
func TestWorkflowHTTPAllOperationsCanonicalParity(t *testing.T) {
	server := httptest.NewServer(NewServer().mux)
	defer server.Close()
	for _, texts := range [][]string{{"from [{id:1}]"}, {"from [{id:1}]", "from [{id:1}] | eval broken =", "from [{id:1}] | unknowable_command"}} {
		request := workflowHTTPSeed(t, texts...)
		report, err := workflow.Assess(request)
		if err != nil {
			t.Fatal(err)
		}
		comparisonRequest := workflow.CompareRequest{SchemaVersion: 1, Before: *report, After: *report}
		comparison, err := workflow.Compare(comparisonRequest)
		if err != nil {
			t.Fatal(err)
		}
		evidenceRequest := workflow.EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}}
		evidence, err := workflow.Evidence(evidenceRequest)
		if err != nil {
			t.Fatal(err)
		}
		proposal := request.Documents[0].Document
		proposal.Text = "from [{id:1}] | fields id"
		recheckRequest := workflow.RecheckRequest{SchemaVersion: 1, Context: workflow.RecheckContext{Original: request.Documents[0], Snapshot: request.Settings.Snapshot, SchemaBundle: request.Settings.SchemaBundle}, Proposal: workflow.Proposal{Document: &proposal, Settings: request.Settings.Entries[0]}}
		recheck, err := workflow.Recheck(recheckRequest)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			route         string
			request, want any
		}{{"assess", request, report}, {"compare", comparisonRequest, comparison}, {"evidence", evidenceRequest, evidence}, {"recheck", recheckRequest, recheck}} {
			code, mime, body := workflowHTTPPost(t, server, tc.route, workflowHTTPJSON(t, tc.request), "application/json", false)
			if code != 200 {
				t.Fatalf("%s %d %s", tc.route, code, body)
			}
			if mime != "application/json" && mime != "text/plain; charset=utf-8" {
				t.Fatalf("MIME %s", mime)
			}
			var got, want any
			if err = json.Unmarshal(body, &got); err != nil {
				t.Fatal(err)
			}
			if err = json.Unmarshal(workflowHTTPJSON(t, tc.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("%s differs from canonical result", tc.route)
			}
		}
	}
}
func TestWorkflowHTTPAssessmentFormats(t *testing.T) {
	server := httptest.NewServer(NewServer().mux)
	defer server.Close()
	for _, format := range []string{"", "json", "text", "sarif", "graph", "bom"} {
		request := workflowHTTPSeed(t, "from [{id:1}] | eval broken =")
		request.Format = format
		payload, report, err := workflow.AssessOutput(request)
		if err != nil {
			t.Fatal(err)
		}
		if report.CIExitCode != 1 {
			t.Fatal("lost CI status")
		}
		code, mime, body := workflowHTTPPost(t, server, "assess", workflowHTTPJSON(t, request), "application/json; charset=utf-8", false)
		if code != 200 {
			t.Fatalf("%s %d %s", format, code, body)
		}
		if format == "text" {
			if mime != "text/plain; charset=utf-8" || string(body) != payload.(string) {
				t.Fatalf("text mismatch MIME=%s", mime)
			}
		} else {
			var got, want any
			json.Unmarshal(body, &got)
			json.Unmarshal(workflowHTTPJSON(t, payload), &want)
			if mime != "application/json" || !reflect.DeepEqual(got, want) {
				t.Fatalf("%s JSON mismatch MIME=%s", format, mime)
			}
		}
	}
}
func TestWorkflowHTTPStrictAdmissionAndMethods(t *testing.T) {
	server := httptest.NewServer(NewServer().mux)
	defer server.Close()
	for _, route := range []string{"assess", "compare", "evidence", "recheck"} {
		for _, raw := range [][]byte{[]byte(`{`), []byte(`{"schema_version":1,"schema_version":1}`), []byte(`{"schema_version":1,"unexpected":true}`), []byte(`{"schema_version":1,"bad":"\ud800"}`), append([]byte(`{"bad":"`), 0xff, '"', '}')} {
			code, _, body := workflowHTTPPost(t, server, route, raw, "application/json", false)
			var detail workflow.RequestErrorDetail
			if code != 400 || json.Unmarshal(body, &detail) != nil || detail.Code != "request_invalid" {
				t.Fatalf("%s code=%d body=%s", route, code, body)
			}
		}
		for _, mime := range []string{"", "text/plain", "application/json; invalid"} {
			code, _, _ := workflowHTTPPost(t, server, route, []byte(`{}`), mime, false)
			if code != 400 {
				t.Fatalf("%s MIME %q code=%d", route, mime, code)
			}
		}
		request, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/workflow/"+route, nil)
		if err != nil {
			t.Fatal(err)
		}
		response, err := server.Client().Do(request)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(io.Discard, response.Body)
		closeErr := response.Body.Close()
		if err != nil || closeErr != nil || response.StatusCode != 405 {
			t.Fatalf("method response %d read=%v close=%v", response.StatusCode, err, closeErr)
		}
	}
	request := workflowHTTPSeed(t, "from [{id:1}]")
	request.Format = "unsupported"
	code, _, body := workflowHTTPPost(t, server, "assess", workflowHTTPJSON(t, request), "application/json", false)
	var detail workflow.RequestErrorDetail
	json.Unmarshal(body, &detail)
	if code != 400 || detail.Path != "/format" {
		t.Fatalf("format error %d %s", code, body)
	}
}
func TestWorkflowHTTPBodyLimitsKnownLengthAndChunked(t *testing.T) {
	server := httptest.NewServer(NewServer().mux)
	defer server.Close()
	request := workflowHTTPSeed(t, "from [{id:1}]")
	report, err := workflow.Assess(request)
	if err != nil {
		t.Fatal(err)
	}
	document := request.Documents[0].Document
	operations := map[string]any{"assess": request, "compare": workflow.CompareRequest{SchemaVersion: 1, Before: *report, After: *report}, "evidence": workflow.EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}}, "recheck": workflow.RecheckRequest{SchemaVersion: 1, Context: workflow.RecheckContext{Original: request.Documents[0], Snapshot: request.Settings.Snapshot, SchemaBundle: request.Settings.SchemaBundle}, Proposal: workflow.Proposal{Document: &document, Settings: request.Settings.Entries[0]}}}
	for route, operation := range operations {
		raw := workflowHTTPJSON(t, operation)
		exact := append(raw, bytes.Repeat([]byte(" "), (8<<20)-len(raw))...)
		for _, chunked := range []bool{false, true} {
			code, _, body := workflowHTTPPost(t, server, route, exact, "application/json", chunked)
			if code != 200 {
				t.Fatalf("%s exact chunked=%t code=%d body=%s", route, chunked, code, body)
			}
			code, _, body = workflowHTTPPost(t, server, route, append(exact, ' '), "application/json", chunked)
			if code != 400 || !strings.Contains(string(body), "request body exceeds 8 MiB limit") {
				t.Fatalf("%s over chunked=%t code=%d body=%s", route, chunked, code, body)
			}
		}
	}
}

func TestWorkflowHTTPRetainsEntryConfigurationFailures(t *testing.T) {
	server := httptest.NewServer(NewServer().mux)
	defer server.Close()
	request := workflowHTTPSeed(t, "from [{id:1}]", "from [{id:1}] | eval broken =")
	request.Settings.Entries[0].Compatibility.QueryScope.App.Values = []string{"resolution", "resolution"}
	report, err := workflow.Assess(request)
	if err != nil {
		t.Fatal(err)
	}
	if report.CIExitCode != 2 || report.Entries[0].Failure == nil || report.Entries[1].Compatibility == nil {
		t.Fatalf("failure setup %+v", report)
	}
	document := request.Documents[0].Document
	operations := map[string]any{"assess": request, "compare": workflow.CompareRequest{SchemaVersion: 1, Before: *report, After: *report}, "evidence": workflow.EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}}, "recheck": workflow.RecheckRequest{SchemaVersion: 1, Context: workflow.RecheckContext{Original: request.Documents[0], Snapshot: request.Settings.Snapshot, SchemaBundle: request.Settings.SchemaBundle}, Proposal: workflow.Proposal{Document: &document, Settings: request.Settings.Entries[0]}}}
	for route, operation := range operations {
		code, _, body := workflowHTTPPost(t, server, route, workflowHTTPJSON(t, operation), "application/json", false)
		if code != 200 || !json.Valid(body) {
			t.Fatalf("%s dropped retained failure code=%d body=%s", route, code, body)
		}
		if route == "evidence" {
			var got workflow.EvidenceReport
			json.Unmarshal(body, &got)
			if got.SourceCIExitCode != 2 {
				t.Fatalf("sourceCI %+v", got)
			}
		}
	}
}
