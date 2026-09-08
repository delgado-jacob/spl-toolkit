package rewrite

import (
	"bytes"
	"encoding/json"
	"reflect"
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

func TestRewriteModelCandidateFieldListProjection(t *testing.T) {
	decode := func(raw []byte) map[string]any {
		t.Helper()
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		var object map[string]any
		if err := decoder.Decode(&object); err != nil {
			t.Fatal(err)
		}
		return object
	}
	for _, tc := range []struct{ name, kind, identity, version string }{
		{"populated metadata", "field_list", " destination catalog ", "v7"},
		{"empty metadata remains present", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, err := validation.Validate(analysis.QueryDocument{Text: "search user=alice | table user extra missing", SourceID: "candidate.spl"}, validation.FieldCatalog{
				Fields: []string{"user"}, OptionalFields: []string{"extra"}, Identity: tc.identity, Version: tc.version,
			})
			if err != nil {
				t.Fatal(err)
			}
			report.Target.Kind = tc.kind
			if report.Status != analysis.Invalid || len(report.Outcomes) == 0 || len(report.Diagnostics) == 0 || report.Analysis == nil || len(report.Analysis.References) == 0 {
				t.Fatalf("fixture lacks canonical report evidence: %#v", report)
			}
			canonical, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			var snapshot validation.Report
			if err := json.Unmarshal(canonical, &snapshot); err != nil {
				t.Fatal(err)
			}
			candidate := CandidateValidation{Kind: "field_list", FieldList: report}
			raw, err := json.Marshal(candidate)
			if err != nil {
				t.Fatal(err)
			}
			got := decode(raw)
			if len(got) != 2 || got["kind"] != "field_list" {
				t.Fatalf("candidate validation envelope changed: %s", raw)
			}
			fieldList, ok := got["field_list"].(map[string]any)
			if !ok {
				t.Fatalf("field-list report missing: %s", raw)
			}
			target, ok := fieldList["target"].(map[string]any)
			if !ok {
				t.Fatalf("target metadata missing: %s", raw)
			}
			for _, key := range []string{"fields", "optional_fields"} {
				if _, exists := target[key]; exists {
					t.Errorf("rewrite candidate repeats target.%s: %s", key, raw)
				}
			}
			if !reflect.DeepEqual(target, map[string]any{"kind": tc.kind, "identity": tc.identity, "version": tc.version}) {
				t.Errorf("target metadata changed or omitted: %#v", target)
			}
			want := decode(canonical)
			canonicalTarget := want["target"].(map[string]any)
			if !reflect.DeepEqual(canonicalTarget["fields"], []any{"user"}) || !reflect.DeepEqual(canonicalTarget["optional_fields"], []any{"extra"}) {
				t.Fatalf("standalone canonical catalog arrays changed: %s", canonical)
			}
			delete(canonicalTarget, "fields")
			delete(canonicalTarget, "optional_fields")
			if !reflect.DeepEqual(fieldList, want) {
				t.Error("rewrite projection changed report values beyond the two target catalog keys")
			}
			if candidate.FieldList != report || !reflect.DeepEqual(report, &snapshot) {
				t.Error("rewrite serialization mutated the original Go report")
			}
			after, err := json.Marshal(report)
			if err != nil || !bytes.Equal(after, canonical) {
				t.Errorf("standalone canonical serialization changed after rewrite marshal: %s, %v", after, err)
			}
		})
	}
}

func TestRewriteModelCandidateSchemaSerialization(t *testing.T) {
	for _, kind := range []string{"json_schema", "ocsf"} {
		t.Run(kind, func(t *testing.T) {
			uid := int64(9007199254740993)
			report := &validation.SchemaReport{
				SchemaVersion: 1, Status: analysis.Incomplete,
				Target:      validation.SchemaTargetInfo{Kind: kind, Identity: "original schema", ResourceURIs: []string{"urn:local:target"}, Members: []validation.SchemaClass{{Key: "event", UID: uid}}, Limitations: []string{"partial_name_universe"}},
				Analysis:    &analysis.Result{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "search user=alice", SourceID: "schema-candidate.spl"}, Status: analysis.Valid},
				Coverage:    validation.Coverage{SyntaxComplete: true, SemanticComplete: true, Reasons: []string{"schema_ambiguity"}},
				Outcomes:    []validation.SchemaReferenceOutcome{{ReferenceID: "ref1", Outcome: "indeterminate", Evidence: []validation.SchemaEvidence{{Pointer: "/properties/user", ClassUID: &uid, Reason: "unknown"}}}},
				Diagnostics: []analysis.Diagnostic{{Code: "specific", Message: "original schema evidence", Severity: "warning"}},
			}
			canonical, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(CandidateValidation{Kind: kind, Schema: report})
			if err != nil {
				t.Fatal(err)
			}
			var got map[string]json.RawMessage
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 || string(got["kind"]) != `"`+kind+`"` || !bytes.Equal(got["schema"], canonical) {
				t.Fatalf("canonical schema report serialization changed: %s", raw)
			}
		})
	}
}
