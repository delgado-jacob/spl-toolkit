package workflow

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func disclosureMarkers(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/workflow/disclosure.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err = json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}
func markedDisclosureReport(t *testing.T) (*Report, map[string]string) {
	t.Helper()
	m := disclosureMarkers(t)
	req := boundComparisonRequest(t, `from events_good | fields `+m["requirement_names"]+` | eval note="`+m["query_text"]+`"`)
	req.Documents[0].Document.SourceID = m["source_identity"]
	req.Documents[0].ID = m["source_identity"]
	req.Settings.Entries[0].ID = m["source_identity"]
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	e := &r.Entries[0]
	e.Origin = corpus.Origin{Kind: "file", RelativePath: m["source_identity"] + ".spl", BaseURI: "https://" + m["source_identity"]}
	r.Provenance.EnvironmentDigest = m["artifact_identity"]
	r.Provenance.SchemaBundleDigest = m["artifact_identity"] + "_SCHEMA"
	r.Provenance.SettingsDigest = m["artifact_identity"] + "_SETTINGS"
	c := e.Compatibility
	c.Provenance.EnvironmentDigest = r.Provenance.EnvironmentDigest
	c.Provenance.SchemaBundleDigest = r.Provenance.SchemaBundleDigest
	c.Provenance.AssessmentIdentityDigest = m["artifact_identity"] + "_ASSESSMENT"
	c.Diagnostics = append(c.Diagnostics, environment.Diagnostic{Code: m["unknown_code"], Severity: "warning", Artifact: "snapshot", Path: m["source_identity"] + "/diagnostic", Message: m["diagnostic_details"]})
	c.Reasons = append(c.Reasons, compatibility.Reason{Code: m["unknown_code"], Message: m["diagnostic_details"], Locations: []analysis.Location{}, ReferenceIDs: []string{}, CandidateInputIDs: []string{}})
	c.Coverage = append(c.Coverage, compatibility.Coverage{Dimension: m["unknown_dimension"], State: m["unknown_state"], Reasons: []compatibility.Reason{}})
	for i := range c.Inputs {
		for j := range c.Inputs[i].Objects {
			if o := c.Inputs[i].Objects[j].Object; o != nil {
				o.Owner = m["environment_metadata"]
				o.Provenance.SourceID = m["source_identity"]
			}
		}
	}
	for i := range c.RequirementOutcomes {
		for j := range c.RequirementOutcomes[i].Objects {
			if o := c.RequirementOutcomes[i].Objects[j].Object; o != nil {
				o.Owner = m["environment_metadata"]
				o.Provenance.SourceID = m["source_identity"]
			}
		}
	}
	return r, m
}
func evidenceBytes(t *testing.T, q EvidenceRequest) (*EvidenceReport, string) {
	t.Helper()
	got, err := Evidence(q)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	return got, string(raw)
}
func TestDisclosureMinimalAndIndependentCategories(t *testing.T) {
	r, m := markedDisclosureReport(t)
	minimal, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
	for _, marker := range m {
		if strings.Contains(raw, marker) {
			t.Fatalf("default disclosed %s", marker)
		}
	}
	if !strings.Contains(raw, "unrecognized_code") || !strings.Contains(raw, "unrecognized") {
		t.Fatal("unknown vocabulary was silently dropped")
	}
	if len(minimal.Disclosure.Requested) != 0 || len(minimal.Disclosure.Emitted) != 0 || len(minimal.Disclosure.PotentiallySensitive) != 0 || len(minimal.Disclosure.Omitted) != 7 {
		t.Fatalf("incorrect default disclosure %+v", minimal.Disclosure)
	}
	for _, category := range []string{"query_text", "requirement_names", "source_identity", "environment_metadata", "diagnostic_details", "artifact_identity"} {
		t.Run(category, func(t *testing.T) {
			got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{category}})
			if !strings.Contains(raw, m[category]) {
				t.Fatalf("selected category missing %s", category)
			}
			for _, other := range []string{"query_text", "requirement_names", "source_identity", "environment_metadata", "artifact_identity"} {
				// Query text intentionally includes literal names used by the query.
				if other == category || (category == "query_text" && other == "requirement_names") {
					continue
				}
				if strings.Contains(raw, m[other]) {
					t.Fatalf("%s unexpectedly disclosed %s", category, other)
				}
			}
			if !reflect.DeepEqual(got.Disclosure.Requested, []string{category}) || !reflect.DeepEqual(got.Disclosure.Emitted, []string{category}) || !reflect.DeepEqual(got.Disclosure.PotentiallySensitive, []string{category}) {
				t.Fatalf("incorrect disclosure %+v", got.Disclosure)
			}
		})
	}
	q := EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{"source_identity", "artifact_identity", "diagnostic_details"}}
	got, first := evidenceBytes(t, q)
	_, second := evidenceBytes(t, q)
	if first != second {
		t.Fatal("projection is not deterministic")
	}
	for _, category := range q.Include {
		if !strings.Contains(first, m[category]) {
			t.Fatal("mixed disclosure lost category")
		}
	}
	for _, item := range got.Items {
		if strings.Contains(item.Pointer, "PRIVATE") || strings.Contains(item.Token, "PRIVATE") {
			t.Fatal("caller identity became pointer or token")
		}
	}
	r.Entries[0].Compatibility.Diagnostics[len(r.Entries[0].Compatibility.Diagnostics)-1].Message = "mutated source"
	after, _ := json.Marshal(got)
	if string(after) != first {
		t.Fatal("projection retained mutable input")
	}
	got.Items[0].Codes = append(got.Items[0].Codes, "mutated output")
	if strings.Contains(r.Entries[0].Compatibility.Diagnostics[0].Code, "mutated") {
		t.Fatal("output mutation affected source")
	}
}

func TestDisclosureDefinitionBodiesAreSeparate(t *testing.T) {
	m := disclosureMarkers(t)
	r, err := Assess(macroComparisonRequest(t, `| makeresults | eval note="`+m["definitions"]+`"`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Entries[0].Compatibility.Closure.EffectiveAnalysis.Document.Text, m["definitions"]) {
		t.Fatal("fixture has no effective body marker")
	}
	for _, include := range [][]string{{}, {"query_text"}, {"environment_metadata"}, {"diagnostic_details"}} {
		_, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: include})
		if strings.Contains(raw, m["definitions"]) {
			t.Fatalf("definition body disclosed with %v", include)
		}
	}
	got, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{"definitions"}})
	if !strings.Contains(raw, m["definitions"]) || !reflect.DeepEqual(got.Disclosure.Emitted, []string{"definitions"}) {
		t.Fatal("definitions opt-in lost reachable body")
	}
	c, err := Compare(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
	if err != nil {
		t.Fatal(err)
	}
	_, raw = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{"query_text"}})
	if strings.Contains(raw, m["definitions"]) {
		t.Fatal("comparison query text disclosed definition body")
	}
}

func TestDisclosureComparisonDefaultRetainedSensitiveEvidence(t *testing.T) {
	r, m := markedDisclosureReport(t)
	a, err := exportCopy(*r)
	if err != nil {
		t.Fatal(err)
	}
	a.Entries[0].Compatibility.Diagnostics[0].Message = m["diagnostic_details"] + "_AFTER"
	c, err := Compare(CompareRequest{SchemaVersion: 1, Before: *r, After: a})
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Entries[0].Deltas) == 0 {
		t.Fatal("fixture has no delta")
	}
	_, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
	for _, marker := range m {
		if strings.Contains(raw, marker) {
			t.Fatalf("comparison default disclosed %s", marker)
		}
	}
	_, raw = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{"diagnostic_details"}})
	if !strings.Contains(raw, m["diagnostic_details"]) {
		t.Fatal("comparison opt-in lost diagnostics")
	}
	if strings.Contains(raw, "before_pointer") || strings.Contains(raw, "after_pointer") || strings.Contains(raw, `"delta":`) {
		t.Fatal("copied raw delta or alignment into Details")
	}
}

func TestDisclosureFailureDetailAndSafeNestedErrors(t *testing.T) {
	r, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	r.Entries[0].Failure = &Failure{Phase: "configuration", Code: "PRIVATE_FAILURE_CODE_17", Message: "PRIVATE_FAILURE_MESSAGE_17", Detail: json.RawMessage(`{"PRIVATE_RAW_SCHEMA_NAME_17":{"description":"PRIVATE_RAW_SCHEMA_METADATA_17","source":"PRIVATE_NESTED_SOURCE_17"}}`)}
	r.Entries[0].Status = analysis.Incomplete
	r.Entries[0].Compatibility = nil
	r.ExecutionComplete = false
	r.Counts = Counts{Selected: 1}
	finalize(r)
	q := EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}}
	got, raw := evidenceBytes(t, q)
	if strings.Contains(raw, "PRIVATE_") {
		t.Fatal("failure nested payload leaked")
	}
	found := false
	for _, item := range got.Items {
		if item.Kind == "failure" {
			found = true
			if item.Complete || !reflect.DeepEqual(item.Codes, []string{"unrecognized_code"}) {
				t.Fatal("unknown failure mishandled")
			}
		}
	}
	if !found {
		t.Fatal("missing failure")
	}
	q.Include = []string{"diagnostic_details"}
	_, raw = evidenceBytes(t, q)
	for _, marker := range []string{"PRIVATE_FAILURE_MESSAGE_17", "PRIVATE_FAILURE_CODE_17", "PRIVATE_RAW_SCHEMA_NAME_17", "PRIVATE_RAW_SCHEMA_METADATA_17"} {
		if !strings.Contains(raw, marker) {
			t.Fatal("diagnostic opt-in missing detail")
		}
	}
	r.Entries[0].Failure.Detail = json.RawMessage(`{"PRIVATE_DUPLICATE":1,"PRIVATE_DUPLICATE":2}`)
	_, err = Evidence(q)
	if err == nil || strings.Contains(err.Error(), "PRIVATE") {
		t.Fatalf("unsafe nested error %v", err)
	}
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.Path != "/report/entries/0/failure/detail" || detail.ByteOffset == nil {
		t.Fatalf("safe structured path/offset missing %+v", detail)
	}
}

func TestDisclosureRawSchemaAndRetainedProjectionMetadata(t *testing.T) {
	req := boundComparisonRequest(t, "from events_good | fields id")
	req.Settings.SchemaBundle.Schemas[0].Kind = "json_schema"
	req.Settings.SchemaBundle.Schemas[0].Catalog = nil
	req.Settings.SchemaBundle.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","schema":{"$id":"https://PRIVATE_SCHEMA_URI_17/schema","type":"object","properties":{"id":{"type":"number","description":"PRIVATE_RAW_SCHEMA_METADATA_17"},"PRIVATE_RAW_SCHEMA_NAME_17":{"type":"string"}},"required":["id"],"additionalProperties":false}}`)
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	var projection *validation.FieldProjection
	for i := range r.Entries[0].Compatibility.RequirementOutcomes {
		if f := r.Entries[0].Compatibility.RequirementOutcomes[i].FieldProjection; f != nil {
			projection = f
			break
		}
	}
	if projection == nil || len(projection.Evidence) == 0 {
		t.Fatal("fixture has no selected schema projection")
	}
	projection.Evidence[0].ClassKey = "PRIVATE_SCHEMA_CLASS_17"
	projection.Evidence[0].Pointer = "/PRIVATE_SCHEMA_POINTER_17"
	for _, include := range [][]string{{}, {"query_text"}, {"requirement_names"}, {"environment_metadata"}, {"source_identity"}} {
		_, raw := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: include})
		if strings.Contains(raw, "PRIVATE_RAW_SCHEMA_METADATA_17") || strings.Contains(raw, "PRIVATE_RAW_SCHEMA_NAME_17") {
			t.Fatal("full raw schema artifact was retained")
		}
		metadata := len(include) == 1 && include[0] == "environment_metadata"
		source := len(include) == 1 && include[0] == "source_identity"
		if strings.Contains(raw, "PRIVATE_SCHEMA_CLASS_17") != metadata {
			t.Fatalf("schema metadata category violated: %v", include)
		}
		if strings.Contains(raw, "PRIVATE_SCHEMA_URI_17") != source || strings.Contains(raw, "PRIVATE_SCHEMA_POINTER_17") != source {
			t.Fatalf("schema source category violated: %v", include)
		}
	}
}

func TestDisclosureResolutionCandidateAndProvenance(t *testing.T) {
	req := seedResolutionRequest(t)
	rawReq, _ := json.Marshal(req)
	rawReq = []byte(strings.ReplaceAll(string(rawReq), "events_good", "PRIVATE_RESOLVED_TARGET_17"))
	if err := json.Unmarshal(rawReq, &req); err != nil {
		t.Fatal(err)
	}
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	rr := r.Entries[0].Resolution
	if rr == nil || len(rr.Variants) == 0 || rr.Variants[0].CandidateText == nil || rr.Variants[0].ResolvedQuery == nil {
		t.Fatal("fixture has no published candidate")
	}
	rawReport, _ := json.Marshal(r)
	for old, marker := range map[string]string{
		rr.Provenance.EnvironmentDigest:     "PRIVATE_RESOLUTION_ENV_DIGEST_17",
		rr.Provenance.SchemaBundleDigest:    "PRIVATE_RESOLUTION_SCHEMA_DIGEST_17",
		rr.Provenance.ResolutionInputDigest: "PRIVATE_RESOLUTION_INPUT_DIGEST_17",
		rr.Provenance.AssessmentInputDigest: "PRIVATE_RESOLUTION_ASSESSMENT_DIGEST_17",
	} {
		if old == "" {
			t.Fatal("fixture is missing digest")
		}
		rawReport = []byte(strings.ReplaceAll(string(rawReport), old, marker))
	}
	if err := json.Unmarshal(rawReport, r); err != nil {
		t.Fatal(err)
	}
	_, minimal := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{}})
	if strings.Contains(minimal, "PRIVATE_") {
		t.Fatal("resolution default disclosed private evidence")
	}
	_, text := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{"query_text"}})
	if !strings.Contains(text, "PRIVATE_RESOLVED_TARGET_17") || strings.Contains(text, "PRIVATE_RESOLUTION_INPUT_DIGEST_17") {
		t.Fatal("candidate text category violated")
	}
	_, names := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{"requirement_names"}})
	if !strings.Contains(names, "PRIVATE_RESOLVED_TARGET_17") {
		t.Fatal("resolution name opt-in lost selected value")
	}
	_, artifacts := evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Report: r, Include: []string{"artifact_identity"}})
	for _, marker := range []string{"PRIVATE_RESOLUTION_ENV_DIGEST_17", "PRIVATE_RESOLUTION_SCHEMA_DIGEST_17", "PRIVATE_RESOLUTION_INPUT_DIGEST_17", "PRIVATE_RESOLUTION_ASSESSMENT_DIGEST_17"} {
		if !strings.Contains(artifacts, marker) {
			t.Fatal("artifact opt-in lost provenance")
		}
	}
	if strings.Contains(artifacts, "PRIVATE_RESOLVED_TARGET_17") {
		t.Fatal("artifact identity disclosed resolved target")
	}
	c, err := Compare(CompareRequest{SchemaVersion: 1, Before: *r, After: *r})
	if err != nil {
		t.Fatal(err)
	}
	_, minimal = evidenceBytes(t, EvidenceRequest{SchemaVersion: 1, Comparison: c, Include: []string{}})
	if strings.Contains(minimal, "PRIVATE_") {
		t.Fatal("comparison resolution default disclosed private evidence")
	}
}
