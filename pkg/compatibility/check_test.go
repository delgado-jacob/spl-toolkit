package compatibility

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func assessmentFixture(t *testing.T, text string) Request {
	t.Helper()
	r := requestFixture(t, text)
	r.Snapshot.Capabilities = []environment.Capability{{ID: "language:" + r.Requirements.Query.Language + ":profile:" + r.Requirements.Query.Profile, Version: r.Requirements.Query.Version, State: "available", Provenance: r.Snapshot.Objects[0].Provenance}}
	bindings := []environment.SchemaBinding{}
	for i := range r.InputBindings {
		r.InputBindings[i].SchemaID = "fields"
		b := r.InputBindings[i]
		bindings = append(bindings, environment.SchemaBinding{SchemaID: "fields", ObjectID: b.ObjectID, Expected: b.Expected, SourceCoverage: "complete"})
	}
	if len(bindings) > 1 {
		bindings = bindings[:1]
	}
	r.SchemaBundle = &environment.SchemaBundle{SchemaVersion: 1, BundleID: "fields", Provenance: r.Snapshot.Objects[0].Provenance, Schemas: []environment.SchemaEntry{{ID: "fields", Kind: "field_list", Catalog: json.RawMessage(`{"fields":["id","other"],"optional_fields":[],"identity":"fields"}`), Provenance: r.Snapshot.Objects[0].Provenance}}, Bindings: bindings}
	return r
}
func checked(t *testing.T, r Request) *Report {
	t.Helper()
	report, err := Check(r)
	if err != nil {
		t.Fatal(err)
	}
	return report
}
func fieldOutcome(t *testing.T, report *Report, id string) RequirementOutcome {
	t.Helper()
	for _, out := range report.RequirementOutcomes {
		for _, item := range report.Requirements.Items {
			if item.Kind == "field" && item.Identity == id && item.ID == out.RequirementID {
				return out
			}
		}
	}
	t.Fatalf("missing field outcome %s", id)
	return RequirementOutcome{}
}
func hasReason(report *Report, code string) bool {
	for _, r := range report.Reasons {
		if r.Code == code {
			return true
		}
	}
	return false
}
func TestCheckEvidenceMatrix(t *testing.T) {
	tests := []struct {
		name, outcome, field, code string
		change                     func(*Request)
	}{
		{"complete positive", "satisfied", "satisfied", "", func(r *Request) {}},
		{"partial object positive", "satisfied", "satisfied", "", func(r *Request) {
			r.Snapshot.Collections[0].Coverage = "partial"
			r.Snapshot.Collections[0].Reason = "bounded"
		}},
		{"partial schema positive", "satisfied", "satisfied", "", func(r *Request) {
			r.SchemaBundle.Bindings[0].SourceCoverage = "partial"
			r.SchemaBundle.Bindings[0].Reason = "bounded"
		}},
		{"complete object missing", "unsatisfied", "satisfied", "", func(r *Request) { r.Snapshot.Objects = []environment.Object{} }},
		{"partial object absent", "incomplete", "satisfied", "environment_collection_partial", func(r *Request) {
			r.Snapshot.Objects = []environment.Object{}
			r.Snapshot.Collections[0].Coverage = "partial"
			r.Snapshot.Collections[0].Reason = "bounded"
		}},
		{"unavailable object absent", "incomplete", "satisfied", "environment_collection_unavailable", func(r *Request) {
			r.Snapshot.Objects = []environment.Object{}
			r.Snapshot.Collections[0].Coverage = "unavailable"
			r.Snapshot.Collections[0].Reason = "not acquired"
		}},
		{"out of scope absent", "incomplete", "satisfied", "environment_scope_not_covered", func(r *Request) {
			r.Snapshot.Objects = []environment.Object{}
			r.Snapshot.CaptureScope.App = environment.Selector{Values: []string{"elsewhere"}}
			r.Snapshot.Collections = r.Snapshot.Collections[:1]
		}},
		{"complete field missing", "unsatisfied", "missing", "", func(r *Request) {
			r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"optional_fields":[],"identity":"fields"}`)
		}},
		{"partial field absent", "incomplete", "indeterminate", "schema_partial", func(r *Request) {
			r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"optional_fields":[],"identity":"fields"}`)
			r.SchemaBundle.Bindings[0].SourceCoverage = "partial"
			r.SchemaBundle.Bindings[0].Reason = "bounded"
		}},
		{"no schema", "incomplete", "indeterminate", "schema_not_supplied", func(r *Request) { r.SchemaBundle = nil; r.InputBindings[0].SchemaID = "" }},
		{"optional declaration", "satisfied", "satisfied", "", func(r *Request) {
			r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"optional_fields":["id"],"identity":"fields"}`)
		}},
		{"open schema", "incomplete", "indeterminate", "schema_projection_indeterminate", func(r *Request) {
			r.SchemaBundle.Schemas[0].Kind = "json_schema"
			r.SchemaBundle.Schemas[0].Catalog = nil
			r.SchemaBundle.Schemas[0].Target = json.RawMessage(`{"kind":"json_schema","identity":"open","schema":{"type":"object","additionalProperties":true}}`)
		}},
		{"missing and unknown", "unsatisfied", "indeterminate", "schema_not_supplied", func(r *Request) {
			r.Snapshot.Objects = []environment.Object{}
			r.SchemaBundle = nil
			r.InputBindings[0].SchemaID = ""
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := assessmentFixture(t, "from $events | fields id")
			tt.change(&r)
			report := checked(t, r)
			if report.Outcome != tt.outcome || fieldOutcome(t, report, "id").Outcome != tt.field {
				t.Fatalf("got aggregate %s field %s; want %s %s\n%s", report.Outcome, fieldOutcome(t, report, "id").Outcome, tt.outcome, tt.field, assessmentSummary(report))
			}
			if tt.code != "" && !hasReason(report, tt.code) {
				t.Fatalf("missing reason %s", tt.code)
			}
			if tt.name == "optional declaration" && fieldOutcome(t, report, "id").FieldProjection.Outcome != "optional" {
				t.Fatal("optionality lost")
			}
		})
	}
}
func TestCheckCapabilityObligation(t *testing.T) {
	for _, tt := range []struct{ name, state, version, outcome, code string }{
		{"available", "available", "current", "satisfied", ""}, {"unavailable", "unavailable", "current", "unsatisfied", ""}, {"unknown", "unknown", "current", "incomplete", "environment_capability_unknown"}, {"absent", "", "current", "incomplete", "environment_capability_not_supplied"}, {"different version", "available", "other", "incomplete", "environment_capability_version_unproven"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := assessmentFixture(t, "from $events | fields id")
			if tt.state == "" {
				r.Snapshot.Capabilities = []environment.Capability{}
			} else {
				r.Snapshot.Capabilities[0].State = tt.state
				r.Snapshot.Capabilities[0].Version = tt.version
			}
			r.Snapshot.Capabilities = append(r.Snapshot.Capabilities, environment.Capability{ID: "unused", Version: "1", State: "unavailable", Provenance: r.Snapshot.Objects[0].Provenance})
			report := checked(t, r)
			if report.Outcome != tt.outcome {
				t.Fatalf("%s", assessmentSummary(report))
			}
			if tt.code != "" && !hasReason(report, tt.code) {
				t.Fatal("missing capability reason")
			}
			count := 0
			for _, out := range report.RequirementOutcomes {
				if strings.HasPrefix(out.RequirementID, "capability:") {
					count++
					if out.RequirementID != "capability:language:spl2:profile:splunkd" || out.Query != r.Requirements.Query || out.InputID != "" {
						t.Fatal("fabricated capability identity")
					}
				}
			}
			if count != 1 {
				t.Fatal("must assess only one language/profile obligation")
			}
		})
	}
}
func TestCheckConditionalMissingKeepsFact(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	for i := range r.Requirements.Items {
		if r.Requirements.Items[i].Kind == "field" {
			r.Requirements.Items[i].Necessity = "conditional"
			for j := range r.Requirements.Items[i].Occurrences {
				r.Requirements.Items[i].Occurrences[j].Necessity = "conditional"
			}
		}
	}
	r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["other"],"optional_fields":[],"identity":"fields"}`)
	report := checked(t, r)
	out := fieldOutcome(t, report, "id")
	if report.Outcome != "incomplete" || out.Outcome != "missing" || out.Applicability != "indeterminate" || !hasReason(report, "conditional_applicability_unproven") {
		t.Fatalf("%s", assessmentSummary(report))
	}
	r.SchemaBundle.Schemas[0].Catalog = json.RawMessage(`{"fields":["id"],"optional_fields":[],"identity":"fields"}`)
	if report := checked(t, r); report.Outcome != "satisfied" {
		t.Fatalf("conditional positive: %s", assessmentSummary(report))
	}
}
func TestCheckSelectedSchemaIsolationAndSharedEvidence(t *testing.T) {
	r := assessmentFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	for i := range r.InputBindings {
		if r.Requirements.Inputs[i].Name == "$users" {
			r.InputBindings[i].ObjectID = "users"
			r.InputBindings[i].Expected.Name = "users"
			r.InputBindings[i].SchemaID = "users-fields"
		}
	}
	r.SchemaBundle.Schemas = append(r.SchemaBundle.Schemas, environment.SchemaEntry{ID: "users-fields", Kind: "field_list", Catalog: json.RawMessage(`{"fields":["other"],"optional_fields":[],"identity":"users-fields"}`), Provenance: r.SchemaBundle.Provenance})
	for _, b := range r.InputBindings {
		if b.ObjectID == "users" {
			r.SchemaBundle.Bindings = append(r.SchemaBundle.Bindings, environment.SchemaBinding{SchemaID: b.SchemaID, ObjectID: b.ObjectID, Expected: b.Expected, SourceCoverage: "complete"})
		}
	}
	report := checked(t, r)
	if report.Outcome != "unsatisfied" {
		t.Fatalf("%s", assessmentSummary(report))
	}
	states := map[string]string{}
	for _, out := range report.RequirementOutcomes {
		if out.InputID != "" {
			for _, item := range report.Requirements.Items {
				if item.ID == out.RequirementID && item.Kind == "field" {
					states[out.InputID] = out.Outcome
				}
			}
		}
	}
	for _, input := range report.Inputs {
		if states[input.InputID] != input.Outcome {
			t.Fatal("per-input field outcome not retained")
		}
	}
	if len(states) != 2 {
		t.Fatal("inputs collapsed")
	}
	for _, input := range r.Requirements.Inputs {
		want := "satisfied"
		if input.Name == "$users" {
			want = "missing"
		}
		if states[input.ID] != want {
			t.Fatalf("schema evidence crossed input %s: %s", input.Name, states[input.ID])
		}
	}
	for i := range r.InputBindings {
		r.InputBindings[i] = InputBinding{InputID: r.InputBindings[i].InputID, ObjectID: "events", Expected: r.InputBindings[0].Expected, SchemaID: "fields"}
		r.InputBindings[i].Expected.Name = "events"
	}
	report = checked(t, r)
	if report.Outcome != "satisfied" {
		t.Fatalf("shared selection: %s", assessmentSummary(report))
	}
	if len(report.Inputs) != 2 {
		t.Fatal("shared evidence collapsed inputs")
	}
	for _, input := range report.Inputs {
		if len(input.RequirementIDs) != 2 {
			t.Fatal("lost distinct owned requirement")
		}
	}
}
func TestCheckNoObligationsAndContentStatus(t *testing.T) {
	for _, tt := range []struct{ text, want string }{{"FROM [{id:1}] | fields id", "satisfied"}, {"from $events | fields (", "not assessed"}, {"from $events | unsupportedcommand", "incomplete"}} {
		t.Run(tt.text, func(t *testing.T) {
			r := assessmentFixture(t, tt.text)
			report := checked(t, r)
			if report.Outcome != tt.want {
				t.Fatalf("want %s: %s", tt.want, assessmentSummary(report))
			}
			if tt.want == "incomplete" && len(report.Inputs) == 0 {
				t.Fatal("useful known input lost")
			}
		})
	}
	r := assessmentFixture(t, "FROM [{id:1}] | fields id")
	r.Requirements.QueryStatus = analysis.Incomplete
	r.Requirements.InputCoverage.State = "partial"
	r.Requirements.Coverage.Complete = false
	r.Requirements.Gaps = append(r.Requirements.Gaps, analysis.RequirementGap{Code: analysis.CodeAnalysisResourceLimit, Message: "analysis stopped", ReferenceIDs: []string{}, DiagnosticCodes: []string{}})
	if report := checked(t, r); report.Outcome != "not assessed" {
		t.Fatalf("aborted empty: %s", assessmentSummary(report))
	}
}
func TestCheckIndependentCorrelation(t *testing.T) {
	r := assessmentFixture(t, "from $events | union [from $users]")
	report := checked(t, r)
	if report.Outcome != "satisfied" || report.Correlation.Outcome != "disconnected" {
		t.Fatalf("%s", assessmentSummary(report))
	}
	r.Requirements.Correlation.Outcome = "indeterminate"
	r.Requirements.Correlation.Coverage.State = "partial"
	r.Requirements.Correlation.Coverage.Reasons = append(r.Requirements.Correlation.Coverage.Reasons, analysis.InputReason{Code: "correlation_unproved", Message: "unsupported correlation", ReferenceIDs: []string{}})
	if report := checked(t, r); report.Outcome != "satisfied" {
		t.Fatalf("correlation-only uncertainty: %s", assessmentSummary(report))
	}
}
func TestCheckPreparedConcurrentCanonicalReports(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	p, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	expected := checked(t, r)
	raw, _ := json.Marshal(expected)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			report, err := p.Check(r.assessment())
			if err != nil {
				t.Error(err)
				return
			}
			got, _ := json.Marshal(report)
			if string(got) != string(raw) {
				t.Error("prepared report differs")
			}
			report.InputBindings[0].Expected.Name = "mutation"
			report.RequirementOutcomes[0].Reasons = append(report.RequirementOutcomes[0].Reasons, newReason("caller", "caller", "caller"))
		}()
	}
	wg.Wait()
	alternate := r.assessment()
	alternate.InputBindings[0].SchemaID = ""
	other, err := p.Check(alternate)
	if err != nil {
		t.Fatal(err)
	}
	if other.Provenance.AssessmentIdentityDigest == expected.Provenance.AssessmentIdentityDigest {
		t.Fatal("distinct schema selections share assessment identity")
	}
	if expected.Provenance.SchemaBundleDigest == "" {
		t.Fatal("supplied schema digest omitted")
	}
	r.SchemaBundle = nil
	r.InputBindings[0].SchemaID = ""
	if checked(t, r).Provenance.SchemaBundleDigest != "" {
		t.Fatal("fabricated absent schema digest")
	}
}

// Authored corpus facts are checked before any adapter parity comparison. The
// corpus remains independent of implementation output and is never regenerated.
func TestCheckAuthoredCorpusAssessment(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/compatibility/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name          string                    `json:"name"`
		Document      analysis.QueryDocument    `json:"document"`
		Snapshot      environment.Snapshot      `json:"snapshot"`
		SchemaBundle  *environment.SchemaBundle `json:"schema_bundle"`
		QueryScope    environment.CaptureScope  `json:"query_scope"`
		InputBindings map[string]struct {
			Input    string                     `json:"input"`
			ObjectID string                     `json:"object_id"`
			Expected environment.ObjectIdentity `json:"expected"`
			SchemaID string                     `json:"schema_id"`
		} `json:"bindings"`
		Expected struct {
			RequestError  any `json:"request_error"`
			Compatibility struct {
				Outcome     string   `json:"outcome"`
				ReasonCodes []string `json:"reason_codes"`
			} `json:"compatibility"`
			RequirementOutcomes []struct {
				Kind, Identity, Input, Necessity, Outcome string
				FieldIdentity                             *analysis.FieldIdentity `json:"field_identity"`
			} `json:"requirement_outcomes"`
		} `json:"expected"`
	}
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if c.Expected.RequestError != nil {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			set, err := analysis.Requirements(c.Document)
			if err != nil {
				t.Fatal(err)
			}
			r := Request{SchemaVersion: 1, Requirements: *set, Snapshot: c.Snapshot, SchemaBundle: c.SchemaBundle, QueryScope: c.QueryScope, InputBindings: []InputBinding{}}
			for key, binding := range c.InputBindings {
				for _, input := range set.Inputs {
					if input.Kind+":"+input.Name == key {
						r.InputBindings = append(r.InputBindings, InputBinding{InputID: input.ID, ObjectID: binding.ObjectID, Expected: binding.Expected, SchemaID: binding.SchemaID})
					}
				}
			}
			report := checked(t, r)
			if report.Outcome != c.Expected.Compatibility.Outcome {
				t.Fatalf("got %s want %s\n%s", report.Outcome, c.Expected.Compatibility.Outcome, assessmentSummary(report))
			}
			for _, code := range c.Expected.Compatibility.ReasonCodes {
				if !hasReason(report, code) {
					t.Errorf("missing authored reason %s", code)
				}
			}
			for _, want := range c.Expected.RequirementOutcomes {
				found := false
				for _, item := range set.Items {
					input := ""
					for _, source := range set.Inputs {
						if source.ID == item.InputID {
							input = source.Kind + ":" + source.Name
						}
					}
					if item.Kind != want.Kind || item.Identity != want.Identity || (want.Input != "" && input != want.Input) || (want.Input == "" && want.Kind == "field" && item.InputID != "") || (want.Necessity != "" && item.Necessity != want.Necessity) || (want.FieldIdentity != nil && !reflect.DeepEqual(item.FieldIdentity, want.FieldIdentity)) {
						continue
					}
					for _, out := range report.RequirementOutcomes {
						if out.RequirementID == item.ID {
							found = true
							if out.Outcome != want.Outcome {
								t.Errorf("%s %s got %s want %s", item.Kind, item.Identity, out.Outcome, want.Outcome)
							}
						}
					}
				}
				if !found {
					t.Errorf("authored requirement not found: %+v", want)
				}
			}
		})
	}
}
func TestCheckBindingIdentityCanonicalOrder(t *testing.T) {
	r := assessmentFixture(t, "from $events | union [from $users]")
	first := checked(t, r)
	sort.Slice(r.InputBindings, func(i, j int) bool { return r.InputBindings[i].InputID > r.InputBindings[j].InputID })
	second := checked(t, r)
	if first.Provenance.AssessmentIdentityDigest != second.Provenance.AssessmentIdentityDigest {
		t.Fatal("binding order changed assessment identity")
	}
}

func assessmentSummary(report *Report) string {
	return stableKey(struct {
		Outcome  string
		Inputs   []InputOutcome
		Outcomes []RequirementOutcome
		Reasons  []Reason
	}{report.Outcome, report.Inputs, report.RequirementOutcomes, report.Reasons})
}

func TestObjectObservationMeaningPreserved(t *testing.T) {
	for _, tt := range []struct{ name, text, want string }{{"observed positive", `from {kind:"sourcetype",properties:{name:"audit"}}`, "satisfied"}, {"not observed source", `from {kind:"source",properties:{name:"absent"}}`, "incomplete"}, {"not observed sourcetype", `from {kind:"sourcetype",properties:{name:"absent"}}`, "incomplete"}} {
		t.Run(tt.name, func(t *testing.T) {
			r := assessmentFixture(t, tt.text)
			r.SchemaBundle = nil
			r.InputBindings = []InputBinding{}
			provenance := r.Snapshot.Objects[0].Provenance
			r.Snapshot.SchemaVersion = 2
			r.Snapshot.Collections = append(r.Snapshot.Collections, environment.Collection{Kind: "source", Coverage: "complete"}, environment.Collection{Kind: "sourcetype", Coverage: "complete"})
			r.Snapshot.Objects = []environment.Object{{ID: "index-main", Kind: "index", Name: "main", Provenance: provenance}, {ID: "type-audit", Kind: "sourcetype", Name: "audit", Provenance: provenance}}
			yes := true
			r.Snapshot.Observation = &environment.ObservationScope{IndexSelection: environment.Selector{All: &yes}, Enumeration: environment.IndexEnumeration{Method: "distributed_rest", PeerScope: "configured_search_peers", Coverage: "complete", Provenance: provenance}, Indexes: []environment.ObservationIndex{{IndexID: "index-main", CatalogDatatypes: []string{"event"}, RequiredDatatypes: []string{"event"}}}, UnmatchedIndexes: []string{}, Method: "splunk_metadata", Visibility: "exporting_principal", Window: environment.ObservationWindow{Mode: "all_retained"}, TimePrecision: "bucket_overlap", AbsenceMeaning: "not_observed", Captures: []environment.ObservationCapture{{Kind: "source", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{}, Coverage: "complete", Provenance: provenance}, {Kind: "sourcetype", IndexID: "index-main", Datatype: "event", ObjectIDs: []string{"type-audit"}, Coverage: "complete", Provenance: provenance}}}
			report := checked(t, r)
			if report.Outcome != tt.want {
				t.Fatalf("%s", assessmentSummary(report))
			}
			if !reflect.DeepEqual(report.Observation, r.Snapshot.Observation) {
				t.Fatal("observation provenance or datatype associations lost")
			}
			if tt.want == "incomplete" && !hasReason(report, "observation_scope_insufficient") {
				t.Fatal("observed absence became global missing")
			}
			r.Snapshot.Observation.Captures[1].ObjectIDs[0] = "mutation"
			if report.Observation.Captures[1].ObjectIDs[0] != "type-audit" {
				t.Fatal("observation aliases caller")
			}
		})
	}
}
func TestObjectExactAmbiguityAndBinding(t *testing.T) {
	r := assessmentFixture(t, "from events")
	r.SchemaBundle = nil
	r.InputBindings = []InputBinding{}
	extra := r.Snapshot.Objects[0]
	extra.ID = "events-second"
	extra.App = "other"
	r.Snapshot.Objects = append(r.Snapshot.Objects, extra)
	report := checked(t, r)
	if report.Outcome != "incomplete" || report.Inputs[0].Outcome != "ambiguous" || len(report.Inputs[0].Objects) != 2 {
		t.Fatalf("%s", assessmentSummary(report))
	}
	r.InputBindings = []InputBinding{{InputID: r.Requirements.Inputs[0].ID, ObjectID: "events", Expected: objectIdentity(r.Snapshot.Objects[0])}}
	if report := checked(t, r); report.Outcome != "satisfied" {
		t.Fatalf("explicit selection: %s", assessmentSummary(report))
	}
	r.InputBindings[0].Expected.Name = "different"
	_, err := Check(r)
	requireRequestError(t, err, "binding_invalid", "/input_bindings/0/expected")
}
func TestObjectCanonicalCollectionsAndDynamicIdentity(t *testing.T) {
	r := assessmentFixture(t, "from $events")
	p, err := Prepare(r.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	item := r.Requirements.Items[0]
	item.InputID = ""
	item.Kind = "index"
	item.Identity = "absent"
	out := p.assessObject(item, r.Requirements.Query, r.QueryScope)
	if out.Outcome != "missing" {
		t.Fatalf("complete exact absence: %+v", out)
	}
	item.Resolution = "wildcard"
	out = p.assessObject(item, r.Requirements.Query, r.QueryScope)
	if out.Outcome != "indeterminate" {
		t.Fatal("wildcard became missing")
	}
	item.Resolution = "exact"
	item.Kind = "unregistered_kind"
	out = p.assessObject(item, r.Requirements.Query, r.QueryScope)
	if out.Outcome != "indeterminate" || out.Reasons[0].Code != "unsupported_semantics" {
		t.Fatal("arbitrary collection lookup")
	}
}
func TestFieldDynamicAndTypedProjectionBoundary(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	for i := range r.Requirements.Items {
		if r.Requirements.Items[i].Kind == "field" {
			r.Requirements.Items[i].Resolution = "dynamic"
		}
	}
	report := checked(t, r)
	if report.Outcome != "incomplete" || fieldOutcome(t, report, "id").Outcome != "indeterminate" {
		t.Fatal("dynamic requirement became exact blocker")
	}
	r = assessmentFixture(t, "from $events | fields payload.id")
	report = checked(t, r)
	if report.Outcome != "incomplete" || !hasReason(report, "schema_projection_indeterminate") {
		t.Fatal("flat field list reinterpreted structural path")
	}
	r = assessmentFixture(t, "from $events | fields id")
	for i := range r.Requirements.Items {
		if r.Requirements.Items[i].Kind == "field" {
			r.Requirements.Items[i].FieldIdentity.Kind = "path"
			r.Requirements.Items[i].FieldIdentity.Qualifier = "e"
		}
	}
	report = checked(t, r)
	if report.Outcome != "incomplete" || !hasReason(report, "schema_projection_indeterminate") {
		t.Fatal("caller-supplied qualifier silently stripped")
	}
}
func TestCheckConditionalObjectAbsenceDoesNotBlock(t *testing.T) {
	r := assessmentFixture(t, "from $events")
	r.SchemaBundle = nil
	r.InputBindings[0].SchemaID = ""
	r.Snapshot.Objects = []environment.Object{}
	for i := range r.Requirements.Items {
		r.Requirements.Items[i].Necessity = "conditional"
		for j := range r.Requirements.Items[i].Occurrences {
			r.Requirements.Items[i].Occurrences[j].Necessity = "conditional"
		}
	}
	report := checked(t, r)
	if report.Outcome != "incomplete" || !hasReason(report, "conditional_applicability_unproven") {
		t.Fatalf("conditional object became unconditional blocker: %s", assessmentSummary(report))
	}
}
func TestCheckOutputCollisionRelevance(t *testing.T) {
	for _, tt := range []struct{ text, want string }{{"from $events | join left=e right=u where e.id=u.id [from $events]", "satisfied"}, {"from $events | join left=e right=u where e.id=u.id [from $users] | fields id", "incomplete"}} {
		r := assessmentFixture(t, tt.text)
		report := checked(t, r)
		if report.Outcome != tt.want {
			t.Fatalf("%s", assessmentSummary(report))
		}
		if tt.want == "satisfied" {
			found := false
			for _, d := range report.Requirements.Diagnostics {
				if d.Code == analysis.CodeAmbiguousField {
					found = true
				}
			}
			if !found {
				t.Fatal("output collision diagnostic was lost")
			}
		} else {
			if !hasReason(report, "field_ownership_ambiguous") {
				t.Fatal("downstream ownership uncertainty was ignored")
			}
		}
	}
}
func TestCoveragePositiveEvidencePreservesPartialLimits(t *testing.T) {
	r := assessmentFixture(t, "from $events | fields id")
	r.Snapshot.Collections[0].Coverage = "partial"
	r.Snapshot.Collections[0].Reason = "capture collection limit"
	r.SchemaBundle.Bindings[0].SourceCoverage = "partial"
	r.SchemaBundle.Bindings[0].Reason = "source schema limit"
	report := checked(t, r)
	if report.Outcome != "satisfied" || len(report.Reasons) != 0 {
		t.Fatal("positive evidence demoted by coverage limits")
	}
	collection, schema := false, false
	for _, coverage := range report.Coverage {
		for _, reason := range coverage.Reasons {
			if reason.Message == "capture collection limit" && coverage.CollectionKind == "dataset" {
				collection = true
			}
			if reason.Message == "source schema limit" && coverage.SchemaID == "fields" {
				schema = true
			}
		}
	}
	if !collection || !schema {
		t.Fatal("supplied partial evidence limits disappeared")
	}
}
func TestObjectUnmappedStaticDescriptorReason(t *testing.T) {
	r := assessmentFixture(t, `from {kind:"index",properties:{name:"main",enabled:true}}`)
	r.SchemaBundle = nil
	r.InputBindings = []InputBinding{}
	report := checked(t, r)
	if report.Outcome != "incomplete" || !hasReason(report, "dataset_identity_unmapped") || !hasReason(report, "target_discovery_incomplete") {
		t.Fatalf("%s", assessmentSummary(report))
	}
}
func TestFieldOwnershipCandidatesRemainAmbiguous(t *testing.T) {
	r := assessmentFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users] | fields id")
	report := checked(t, r)
	if report.Outcome != "incomplete" {
		t.Fatal("ownership ambiguity was repaired by schemas")
	}
	found := false
	for _, out := range report.RequirementOutcomes {
		if out.Outcome == "ambiguous" {
			found = true
			if len(out.Schemas) != 0 || len(out.Reasons) == 0 || len(out.Reasons[0].CandidateInputIDs) != 2 || len(out.Reasons[0].Locations) == 0 {
				t.Fatal("ownership candidates or located evidence missing")
			}
		}
	}
	if !found {
		t.Fatal("unproved read lost")
	}
	for _, input := range report.Inputs {
		if input.Outcome != "ambiguous" {
			t.Fatal("candidate input claimed fully satisfied ownership")
		}
	}
}
