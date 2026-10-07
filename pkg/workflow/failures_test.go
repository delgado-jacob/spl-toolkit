package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestWorkflowResolutionLocalFailures(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*ResolveSettings)
		code   string
	}{
		{"over limit", func(s *ResolveSettings) { n := uint64(1); s.MaxVariants = &n }, "fanout_limit_exceeded"},
		{"missing selection", func(s *ResolveSettings) { s.Resolutions = []resolution.Resolution{} }, "resolution_missing"},
		{"duplicate scope", func(s *ResolveSettings) { s.Compatibility.QueryScope.App.Values = []string{"resolution", "resolution"} }, ""},
		{"blank scope", func(s *ResolveSettings) { s.Compatibility.QueryScope.App.Values = []string{""} }, ""},
		{"duplicate value", func(s *ResolveSettings) { s.Resolutions[0].Values = []string{"events_good", "events_good"} }, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := seedResolutionRequest(t)
			tc.change(req.Settings.Entries[0].Resolution)
			req.Documents = append(req.Documents, corpus.RequestDocument{ID: "good", Document: seedRequest(t, "from [{id:1}]").Documents[0].Document})
			req.Settings.Entries = append(req.Settings.Entries, EntrySettings{ID: "good", Compatibility: seedRequest(t, "from [{id:1}]").Settings.Entries[0].Compatibility})
			r, err := Assess(req)
			if err != nil {
				t.Fatal(err)
			}
			e := r.Entries[0]
			if r.CIExitCode != 2 || r.ExecutionComplete || r.Counts.ConfigurationFailed != 1 || r.Counts.Assessed != 1 || r.Counts.Valid != 1 || e.Analysis == nil || e.Resolution != nil || e.Failure == nil || e.Failure.Phase != "configuration" {
				t.Fatalf("report: %+v entry: %+v", r, e)
			}
			var detail resolution.RequestErrorDetail
			if err := json.Unmarshal(e.Failure.Detail, &detail); err != nil {
				t.Fatal(err)
			}
			if detail.Code == "" || detail.Path == "" || detail.Code != e.Failure.Code || (tc.code != "" && detail.Code != tc.code) {
				t.Fatalf("detail: %+v", detail)
			}
			if tc.name == "over limit" && (detail.TotalCombinations != "2" || detail.MaxVariants == nil || *detail.MaxVariants != 1) {
				t.Fatalf("fanout detail: %+v", detail)
			}
		})
	}
}

func TestWorkflowFailuresOverrideRetainedContent(t *testing.T) {
	req := seedResolutionRequest(t)
	n := uint64(1)
	req.Settings.Entries[0].Resolution.MaxVariants = &n
	req.Documents = append(req.Documents, corpus.RequestDocument{ID: "invalid", Document: seedRequest(t, "from [{id:1}] | eval broken =").Documents[0].Document})
	req.Settings.Entries = append(req.Settings.Entries, EntrySettings{ID: "invalid", Compatibility: seedRequest(t, "from [{id:1}]").Settings.Entries[0].Compatibility})
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	if r.CIExitCode != 2 || r.Counts.Invalid != 1 || r.Entries[1].Compatibility == nil {
		t.Fatalf("report: %+v", r)
	}
}

func TestWorkflowFailureRetainsFullWidthFanoutDetail(t *testing.T) {
	req := seedResolutionRequest(t)
	maximum := ^uint64(0)
	s := req.Settings.Entries[0].Resolution
	s.MaxVariants = &maximum
	s.Resolutions = []resolution.Resolution{}
	for i := 0; i < 64; i++ {
		s.Resolutions = append(s.Resolutions, resolution.Resolution{Placeholder: fmt.Sprintf("$role%d", i), Kind: "dataset", Values: []string{"one", "two"}})
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	var detail resolution.RequestErrorDetail
	if r.Entries[0].Failure == nil {
		t.Fatal("fanout admitted")
	}
	if err := json.Unmarshal(r.Entries[0].Failure.Detail, &detail); err != nil {
		t.Fatal(err)
	}
	if detail.TotalCombinations != "18446744073709551616" || detail.MaxVariants == nil || *detail.MaxVariants != maximum || !strings.Contains(string(r.Entries[0].Failure.Detail), "18446744073709551615") {
		t.Fatalf("detail: %+v", detail)
	}
}

func TestWorkflowOperationFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		err         error
		phase, code string
	}{
		{errors.New("local fault"), "internal", "operation_failed"},
		{&validation.InputError{Err: errors.New("invalid local operation")}, "configuration", "configuration_invalid"},
	} {
		failure := operationFailure(tc.err, "analysis")
		if failure.Phase != tc.phase || failure.Code != tc.code || !strings.Contains(failure.Message, tc.err.Error()) {
			t.Fatalf("failure: %+v", failure)
		}
	}
}
