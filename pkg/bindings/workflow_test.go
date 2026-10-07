package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
)

func workflowNativeRequest(t *testing.T) workflow.Request {
	t.Helper()
	raw, err := os.ReadFile("../../examples/resolution/request.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := resolution.DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return workflow.Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "d1", Document: analysis.QueryDocument{Text: "from [{id:1}]", Language: "spl2", Profile: "splunkd", Version: "current", SourceID: "d1"}}}, Settings: workflow.Settings{SchemaVersion: 1, Snapshot: seed.Compatibility.Snapshot, SchemaBundle: seed.Compatibility.SchemaBundle, Entries: []workflow.EntrySettings{{ID: "d1", Compatibility: &workflow.CheckSettings{QueryScope: seed.Compatibility.QueryScope, InputBindings: []compatibility.InputBinding{}}}}}}
}

func TestWorkflowNativeOwnedResults(t *testing.T) {
	request := workflowNativeRequest(t)
	report, err := workflow.Assess(request)
	if err != nil {
		t.Fatal(err)
	}
	raw := func(value any) []byte {
		t.Helper()
		b, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	candidate := request.Documents[0].Document
	candidate.Text = "from [{id:1}] | eval broken ="
	text := request
	text.Format = "text"
	cases := []struct {
		name      string
		native    func(_Ctype_int, *_Ctype_char) *_Ctype_SPLResult
		canonical func([]byte) (any, error)
		request   []byte
	}{
		{"assess", spl_mapper_workflow_assess, workflow.AssessOutputJSON, raw(request)},
		{"text", spl_mapper_workflow_assess, workflow.AssessOutputJSON, raw(text)},
		{"compare", spl_mapper_workflow_compare, func(b []byte) (any, error) { return workflow.CompareJSON(b) }, raw(workflow.CompareRequest{SchemaVersion: 1, Before: *report, After: *report})},
		{"evidence", spl_mapper_workflow_evidence, func(b []byte) (any, error) { return workflow.EvidenceJSON(b) }, raw(workflow.EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}})},
		{"recheck", spl_mapper_workflow_recheck, func(b []byte) (any, error) { return workflow.RecheckJSON(b) }, raw(workflow.RecheckRequest{SchemaVersion: 1, Context: workflow.RecheckContext{Original: request.Documents[0], Snapshot: request.Settings.Snapshot, SchemaBundle: request.Settings.SchemaBundle}, Proposal: workflow.Proposal{Document: &candidate, Settings: request.Settings.Entries[0]}})},
	}
	handle := spl_mapper_new()
	defer spl_mapper_free(handle)
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want, err := c.canonical(c.request)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := marshalNativeJSON(want)
			if err != nil {
				t.Fatal(err)
			}
			// Preserve the full NUL-terminated buffer, including its terminator.
			input := append(bytes.Clone(c.request), 0)
			before := bytes.Clone(input)
			pointer := nativeTestCStringBytes(t, input[:len(input)-1])
			result := c.native(handle, pointer)
			if result == nil {
				t.Fatal("nil owned result")
			}
			defer spl_result_free(result)
			later := c.native(handle, nativeTestCString(t, "{"))
			if later == nil {
				t.Fatal("nil owned error")
			}
			spl_result_free(later)
			if result.error != nil || nativeTestGoString(result.result) != string(encoded) {
				t.Fatalf("error=%s result=%s", nativeTestGoString(result.error), nativeTestGoString(result.result))
			}
			if !bytes.Equal(input, before) || nativeTestGoString(pointer) != string(c.request) {
				t.Fatal("input buffer mutated")
			}
			for _, bad := range [][]byte{nil, []byte("{"), []byte(`{"id":"\ud800"}`), {0xff}} {
				failed := c.native(handle, nativeTestCStringBytes(t, bad))
				if failed == nil {
					t.Fatal("nil error")
				}
				if failed.error == nil || failed.result != nil {
					spl_result_free(failed)
					t.Fatal("invalid request admitted")
				}
				spl_result_free(failed)
			}
		})
	}
	spl_mapper_free(handle)
	for _, c := range cases {
		result := c.native(handle, nil)
		if result == nil {
			t.Fatal("nil closed result")
		}
		if nativeTestGoString(result.error) != "Mapper not found" || result.result != nil {
			spl_result_free(result)
			t.Fatal("closed mapper admitted")
		}
		spl_result_free(result)
	}
}
