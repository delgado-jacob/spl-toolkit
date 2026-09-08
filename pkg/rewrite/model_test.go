package rewrite

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func TestRewriteModelJSON(t *testing.T) {
	r := Result{SchemaVersion: 1, Mode: Preview, Status: analysis.Valid, OriginalText: "src", CandidateText: "user", Text: "src",
		Changes:         []Change{{Outcome: "applied", GroupID: "g1", RuleIDs: []string{"r1"}, OriginalReferenceIDs: []string{"ref1"}, CandidateReferenceIDs: []string{"ref2"}, OriginalLocation: &analysis.Location{Start: analysis.Position{Offset: 0, Line: 1, Column: 1}, End: analysis.Position{Offset: 3, Line: 1, Column: 4}}, OldText: "src", NewText: "user", CandidateApplied: true}},
		RuleEvaluations: []RuleEvaluation{{RuleID: "unmatched", Outcome: "skipped", Reason: ReasonNoMatch}},
	}
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"schema_version", "document", "mode", "status", "coverage", "original_text", "candidate_text", "text", "committed", "changes", "rule_evaluations", "original_analysis", "candidate_analysis"} {
		if _, ok := got[key]; !ok {
			t.Errorf("missing %s: %s", key, raw)
		}
	}
	if _, ok := got["candidate_validation"]; ok {
		t.Fatalf("absent validation included: %s", raw)
	}
	if string(got["text"]) != `"src"` || string(got["candidate_text"]) != `"user"` || string(got["committed"]) != "false" {
		t.Fatalf("preview distinction lost: %s", raw)
	}
	var changes []map[string]json.RawMessage
	if err := json.Unmarshal(got["changes"], &changes); err != nil {
		t.Fatal(err)
	}
	if string(changes[0]["candidate_applied"]) != "true" || string(changes[0]["committed"]) != "false" {
		t.Fatalf("audit distinction lost: %s", raw)
	}
	if _, ok := changes[0]["candidate_location"]; ok {
		t.Fatalf("invented candidate location: %s", raw)
	}
	var evaluations []map[string]json.RawMessage
	if err := json.Unmarshal(got["rule_evaluations"], &evaluations); err != nil {
		t.Fatal(err)
	}
	if string(evaluations[0]["reason"]) != `"no_match"` || string(evaluations[0]["reference_ids"]) != "[]" {
		t.Fatalf("unmatched evidence: %s", raw)
	}
	if _, ok := evaluations[0]["location"]; ok {
		t.Fatalf("invented unmatched location: %s", raw)
	}
	for _, value := range []any{Result{}, BatchResult{}, Coverage{}, Change{}, RuleEvaluation{}, ConditionEvaluation{}} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"reasons", "changes", "rule_evaluations", "reports", "rule_ids", "original_reference_ids", "candidate_reference_ids", "reference_ids", "children"} {
			if strings.Contains(string(data), `"`+key+`":null`) {
				t.Fatalf("null collection %s: %s", key, data)
			}
		}
	}
	data, err := json.Marshal(Request{SchemaVersion: 1})
	if err != nil || !strings.Contains(string(data), `"rules":[]`) {
		t.Fatalf("empty rules: %s, %v", data, err)
	}
	data, err = json.Marshal(Condition{Fact: "literal", Kind: "field", Operator: "equals", Value: json.RawMessage("null")})
	if err != nil || !strings.Contains(string(data), `"value":null`) {
		t.Fatalf("literal null omitted: %s, %v", data, err)
	}
}

func TestRewriteModelCandidateValidation(t *testing.T) {
	for _, value := range []CandidateValidation{
		{Kind: "field_list", FieldList: &validation.Report{SchemaVersion: 1, Status: analysis.Invalid, Diagnostics: []analysis.Diagnostic{{Code: "specific", Message: "original diagnostic"}}}},
		{Kind: "json_schema", Schema: &validation.SchemaReport{SchemaVersion: 1, Status: analysis.Incomplete, Target: validation.SchemaTargetInfo{Identity: "schema-identity"}, Outcomes: []validation.SchemaReferenceOutcome{{ReferenceID: "ref1", Evidence: []validation.SchemaEvidence{{Pointer: "/properties/user", Reason: "unknown"}}}}}},
	} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if value.Kind == "field_list" && (!strings.Contains(string(data), `"field_list":`) || !strings.Contains(string(data), "original diagnostic") || strings.Contains(string(data), `"schema":`)) {
			t.Fatalf("field report flattened/lost: %s", data)
		}
		if value.Kind == "json_schema" && (!strings.Contains(string(data), `"schema":`) || !strings.Contains(string(data), `/properties/user`) || strings.Contains(string(data), `"field_list":`)) {
			t.Fatalf("schema evidence lost: %s", data)
		}
	}
	for _, value := range []CandidateValidation{{Kind: "field_list"}, {Kind: "field_list", FieldList: &validation.Report{}, Schema: &validation.SchemaReport{}}, {Kind: "ocsf", FieldList: &validation.Report{}}, {Kind: "other", Schema: &validation.SchemaReport{}}} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("invalid validation union serialized: %#v", value)
		}
	}
}
