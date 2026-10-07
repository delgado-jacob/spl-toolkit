package workflow

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"testing"
)

func TestWorkflowSARIFCandidateFindingsDoNotPointIntoOriginal(t *testing.T) {
	r, err := Assess(seedResolutionRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	log, err := ExportSARIF(r)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, result := range log.Runs[0].Results {
		if result.Properties["domain"] == "candidate" {
			count++
			if len(result.Locations) != 0 {
				t.Fatal("candidate coordinates attached to original artifact")
			}
		}
	}
	if count == 0 {
		t.Fatal("unsuccessful sibling has no candidate findings")
	}
}

func TestWorkflowSARIFRetainsFailureAndOriginalDiagnostics(t *testing.T) {
	for _, phase := range []string{"configuration", "internal"} {
		t.Run(phase, func(t *testing.T) {
			req := seedResolutionRequest(t)
			req.Documents[0].Document.Text = "from $events | eval broken ="
			maximum := uint64(1)
			req.Settings.Entries[0].Resolution.MaxVariants = &maximum
			r, err := Assess(req)
			if err != nil {
				t.Fatal(err)
			}
			if r.Entries[0].Failure == nil || r.Entries[0].Analysis.Status != analysis.Invalid {
				t.Fatal("failure fixture missing retained invalid analysis")
			}
			// Internal failure uses the same retained-analysis boundary as configuration.
			if phase == "internal" {
				r.Entries[0].Failure.Phase = phase
				r.Counts.ConfigurationFailed = 0
				r.Counts.InternalFailed = 1
			}
			before, _ := json.Marshal(r)
			log, err := ExportSARIF(r)
			if err != nil {
				t.Fatal(err)
			}
			run := log.Runs[0]
			if run.Invocations[0].ExecutionSuccessful || len(run.Invocations[0].ToolExecutionNotifications) != 1 {
				t.Fatalf("invocation: %+v", run.Invocations)
			}
			if run.Invocations[0].ToolExecutionNotifications[0].Properties["phase"] != phase {
				t.Fatal("failure phase lost")
			}
			found := false
			for _, result := range run.Results {
				if result.Level == "error" && result.Properties["domain"] == "original" {
					found = true
				}
			}
			if !found {
				t.Fatal("configuration failure hid original content defect")
			}
			run.Invocations[0].ToolExecutionNotifications[0].Properties["detail"] = "changed"
			after, _ := json.Marshal(r)
			if string(before) != string(after) {
				t.Fatal("report mutated")
			}
		})
	}
}

func TestWorkflowSARIFPartialSelectionAndAcquisitionFailures(t *testing.T) {
	req := seedRequest(t, "from [{id:1}]")
	req.Documents = append(req.Documents, corpus.RequestDocument{ID: "failed", Document: req.Documents[0].Document})
	req.Settings.Entries = append(req.Settings.Entries, EntrySettings{ID: "failed", Compatibility: detach(req.Settings.Entries[0].Compatibility)})
	p, err := Prepare(req.Settings)
	if err != nil {
		t.Fatal(err)
	}
	input := inlineInput(req)
	input.Entries[1].Document = nil
	input.Entries[1].Failure = &corpus.AcquisitionError{Code: "read_failed", Phase: "acquisition", Message: "cannot read"}
	input.Selection.Complete = false
	input.Selection.TraversalFailures = []corpus.AcquisitionError{{Code: "walk_failed", Phase: "traversal", Message: "cannot walk"}}
	r, err := p.Assess(input)
	if err != nil {
		t.Fatal(err)
	}
	log, err := ExportSARIF(r)
	if err != nil {
		t.Fatal(err)
	}
	run := log.Runs[0]
	if run.Invocations[0].ExecutionSuccessful || len(run.Invocations[0].ToolExecutionNotifications) != 2 {
		t.Fatalf("invocation: %+v", run.Invocations)
	}
	if run.Properties["counts"].(map[string]any)["selected"] != float64(2) || run.Properties["selection"].(map[string]any)["complete"] != false {
		t.Fatal("selection denominator lost")
	}
	run.Properties["selection"].(map[string]any)["traversal_failures"].([]any)[0].(map[string]any)["message"] = "changed"
	if r.Selection.TraversalFailures[0].Message != "cannot walk" {
		t.Fatal("selection aliases report")
	}
}

func TestWorkflowSARIFContentOutcomeRulesAndPointers(t *testing.T) {
	for _, req := range []Request{seedResolutionRequest(t), seedRequest(t, "from [{id:1}] | eval broken =")} {
		r, err := Assess(req)
		if err != nil {
			t.Fatal(err)
		}
		log, err := ExportSARIF(r)
		if err != nil {
			t.Fatal(err)
		}
		if !log.Runs[0].Invocations[0].ExecutionSuccessful {
			t.Fatal("content defect became execution failure")
		}
		for _, result := range log.Runs[0].Results {
			if result.RuleIndex < 0 || result.RuleIndex >= len(log.Runs[0].Tool.Driver.Rules) || log.Runs[0].Tool.Driver.Rules[result.RuleIndex].ID != result.RuleID {
				t.Fatal("rule index mismatch")
			}
			pointer, _ := result.Properties["evidence_pointer"].(string)
			workflowPointer(t, r, pointer)
		}
		r.Entries[0].SourceHash = "tampered"
		if _, err := ExportSARIF(r); err == nil {
			t.Fatal("tampered original source accepted")
		}
	}
}

func TestWorkflowSARIFClosureLocationsStayInTheirSourceDomain(t *testing.T) {
	r := closureExportFixture(t, "`m` | `missing`", []closure.Definition{macroExportDefinition("m", "m", "lookup absent user OUTPUT role")})
	log, err := ExportSARIF(r)
	if err != nil {
		t.Fatal(err)
	}
	direct, definition := false, false
	for _, result := range log.Runs[0].Results {
		pointer, _ := result.Properties["evidence_pointer"].(string)
		workflowPointer(t, r, pointer)
		interval, ok := result.Properties["source_interval"].(map[string]any)
		if !ok {
			continue
		}
		switch interval["kind"] {
		case "query":
			if len(result.Locations) > 0 {
				direct = true
			}
		case "definition":
			definition = true
			if len(result.Locations) > 0 {
				t.Fatal("definition coordinates attached to original")
			}
		}
	}
	if !direct || !definition {
		t.Fatal("fixture failed to exercise both closure source domains")
	}
}

func TestWorkflowSARIFReasonlessNestedFindingsRemainVisible(t *testing.T) {
	for _, domain := range []string{"original", "candidate"} {
		t.Run(domain, func(t *testing.T) {
			requirements := []compatibility.RequirementOutcome{
				{Outcome: "missing", Applicability: "applicable", Reasons: []compatibility.Reason{}},
				{Outcome: "indeterminate", Applicability: "inapplicable", Reasons: []compatibility.Reason{}},
				{Outcome: "satisfied", Applicability: "applicable", Reasons: []compatibility.Reason{}},
				{Outcome: "missing", Applicability: "applicable", Reasons: []compatibility.Reason{{Code: "HAS_REASON", Message: "already explained"}}},
			}
			inputs := []compatibility.InputOutcome{{Outcome: "indeterminate", Reasons: []compatibility.Reason{}}, {Outcome: "satisfied", Reasons: []compatibility.Reason{}}, {Outcome: "missing", Reasons: []compatibility.Reason{{Code: "HAS_REASON", Message: "already explained"}}}}
			coverage := []compatibility.Coverage{{Dimension: "field_schema", State: "partial", Reasons: []compatibility.Reason{}}, {Dimension: "query_semantics", State: "complete", Reasons: []compatibility.Reason{}}, {Dimension: "dependency_closure", State: "not_applicable", Reasons: []compatibility.Reason{}}, {Dimension: "field_schema", State: "unavailable", Reasons: []compatibility.Reason{{Code: "HAS_REASON", Message: "already explained"}}}}
			req := seedRequest(t, "from [{id:1}]")
			if domain == "candidate" {
				req = seedResolutionRequest(t)
			}
			r, err := Assess(req)
			if err != nil {
				t.Fatal(err)
			}
			pointer := "/entries/0/compatibility"
			unrelated := compatibility.Reason{Code: "UNRELATED", Message: "unrelated finding"}
			if domain == "original" {
				c := r.Entries[0].Compatibility
				c.Outcome = "incomplete"
				c.Reasons = []compatibility.Reason{unrelated}
				c.RequirementOutcomes, c.Inputs, c.Coverage = requirements, inputs, coverage
			} else {
				pointer = "/entries/0/resolution/variants/0/compatibility"
				c := r.Entries[0].Resolution.Variants[0].Compatibility
				c.Outcome = "incomplete"
				c.Reasons = []compatibility.ResolutionReason{{Evidence: unrelated}}
				c.RequirementOutcomes = []compatibility.ResolutionRequirementOutcome{}
				for _, o := range requirements {
					c.RequirementOutcomes = append(c.RequirementOutcomes, compatibility.ResolutionRequirementOutcome{Evidence: o})
				}
				c.Inputs = []compatibility.ResolutionInputOutcome{}
				for _, o := range inputs {
					c.Inputs = append(c.Inputs, compatibility.ResolutionInputOutcome{Evidence: o})
				}
				c.Coverage = []compatibility.ResolutionCoverage{}
				for _, cov := range coverage {
					c.Coverage = append(c.Coverage, compatibility.ResolutionCoverage{Evidence: cov})
				}
			}
			log, err := ExportSARIF(r)
			if err != nil {
				t.Fatal(err)
			}
			suffix := ""
			if domain == "candidate" {
				suffix = "/evidence"
			}
			expected := map[string]string{pointer + "/requirement_outcomes/0" + suffix: "missing", pointer + "/inputs/0" + suffix: "indeterminate", pointer + "/coverage/0" + suffix: "partial"}
			summaries := 0
			for _, result := range log.Runs[0].Results {
				if result.RuleID != "SPL_WORKFLOW_REQUIREMENT_OUTCOME" && result.RuleID != "SPL_WORKFLOW_INPUT_OUTCOME" && result.RuleID != "SPL_WORKFLOW_COVERAGE_INCOMPLETE" {
					continue
				}
				ptr, _ := result.Properties["evidence_pointer"].(string)
				state, ok := expected[ptr]
				if !ok {
					t.Fatalf("unexpected nested fallback %s", ptr)
				}
				if result.Properties["outcome"] != state && result.Properties["coverage_state"] != state {
					t.Fatalf("missing nested state: %+v", result)
				}
				if result.Properties["domain"] != domain || len(result.Locations) != 0 {
					t.Fatalf("domain/location: %+v", result)
				}
				if domain == "candidate" && len(result.Properties["variant_selection"].([]any)) == 0 {
					t.Fatal("candidate selection lost")
				}
				workflowPointer(t, r, ptr)
				summaries++
				delete(expected, ptr)
			}
			if summaries != 3 || len(expected) != 0 {
				t.Fatalf("nested evidence omitted alongside unrelated finding: %v", expected)
			}
		})
	}
}
