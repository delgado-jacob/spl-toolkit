package compatibility

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

func TestPrepareBindingReuseAndDetach(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	prepared, err := Prepare(r.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Snapshot.Objects[0].Name = "caller changed"
	*r.Snapshot.CaptureScope.App.All = false
	r.QueryScope = secondRequest(t).QueryScope
	first, err := prepared.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	if first.Inputs[0].Objects[0].Object.Name != "events" {
		t.Fatal("prepared snapshot aliases caller")
	}
	r.InputBindings[0] = InputBinding{InputID: r.Requirements.Inputs[0].ID, ObjectID: "users", Expected: environment.ObjectIdentity{Kind: "dataset", Name: "users", Namespace: "search", App: "app", Owner: "nobody"}}
	// QueryScope shares the caller's original all selector pointer; fix the
	// intentionally mutated configuration for the second independently bound call.
	yes := true
	r.QueryScope.App = environment.Selector{All: &yes}
	second, err := prepared.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	if first.Inputs[0].Objects[0].ObjectID != "events" || second.Inputs[0].Objects[0].ObjectID != "users" {
		t.Fatal("query bindings leaked between prepared calls")
	}
	r.Requirements.Inputs[0].Occurrences[0].UseSiteReferenceIDs = append(r.Requirements.Inputs[0].Occurrences[0].UseSiteReferenceIDs, "caller")
	r.Requirements.Items[0].Ownership.CandidateInputIDs[0] = "caller"
	r.InputBindings[0].Expected.Name = "caller"
	if first.Requirements.Items[0].Ownership.CandidateInputIDs[0] == "caller" || second.Inputs[0].Objects[0].Expected.Name == "caller" {
		t.Fatal("report aliases assessment caller")
	}
	first.Inputs[0].Objects[0].Object.Name = "report changed"
	third, err := prepared.Check(secondRequest(t))
	if err != nil {
		t.Fatal(err)
	}
	if third.Inputs[0].Objects[0].Object.Name != "events" {
		t.Fatal("report mutation changed prepared evidence")
	}
}
func secondRequest(t *testing.T) AssessmentRequest {
	return requestFixture(t, "from $events | fields id").assessment()
}
func TestPrepareBoundAbsentAndSchemaLink(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	r.Snapshot.Objects = []environment.Object{}
	r.SchemaBundle = &environment.SchemaBundle{SchemaVersion: 1, BundleID: "fields", Provenance: environment.Provenance{SourceKind: "fixture", SourceID: "fixture", ObservedAt: "2026-10-01T12:02:00Z"}, Schemas: []environment.SchemaEntry{{ID: "fields", Kind: "field_list", Catalog: json.RawMessage(`{"fields":["id"],"optional_fields":[],"identity":"fixture","version":"1"}`), Provenance: environment.Provenance{SourceKind: "fixture", SourceID: "fixture", ObservedAt: "2026-10-01T12:02:00Z"}}}, Bindings: []environment.SchemaBinding{{SchemaID: "fields", ObjectID: "events", Expected: r.InputBindings[0].Expected, SourceCoverage: "complete"}}}
	r.InputBindings[0].SchemaID = "fields"
	prepared, err := Prepare(r.Snapshot, r.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.env.Bindings("events")) != 1 {
		t.Fatal("absent schema binding was discarded")
	}
	r.SchemaBundle.Schemas[0].Catalog[0] = 'X'
	r.SchemaBundle.Bindings[0].Expected.Name = "changed"
	report, err := prepared.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Inputs[0].Objects) != 1 || report.Inputs[0].Objects[0].Object != nil {
		t.Fatal("bound absence is not valid detached configuration")
	}
	r.InputBindings[0].SchemaID = "other"
	_, err = prepared.Check(r.assessment())
	requireRequestError(t, err, "binding_invalid", "/input_bindings/0/schema_id")
}
func TestPrepareExplicitBindingMatches(t *testing.T) {
	for _, tt := range []struct {
		text, kind, name string
		matches          int
		mapped           bool
	}{
		{"from events | fields id", "dataset", "events", 1, true},
		{"from app.events | fields id", "dataset", "app.events", 1, true},
		{"from events | fields id", "index", "events", 0, true},
		{`from {kind:"index",properties:{name:"main"}} | fields id`, "index", "main", 1, true},
		{`from {kind:"index",properties:{name:"main",enabled:true}} | fields id`, "index", "main", 0, false},
	} {
		t.Run(tt.text+tt.kind, func(t *testing.T) {
			r := requestFixture(t, tt.text)
			r.InputBindings = []InputBinding{}
			r.Snapshot.Objects[0].Kind = tt.kind
			r.Snapshot.Objects[0].Name = tt.name
			if tt.kind == "index" {
				r.Snapshot.Objects[0].Namespace = ""
				r.Snapshot.Objects[0].App = ""
				r.Snapshot.Objects[0].Owner = ""
			}
			p, err := Prepare(r.Snapshot, nil)
			if err != nil {
				t.Fatal(err)
			}
			resolved, err := p.resolveInputs(r.Requirements.Inputs, r.QueryScope, r.InputBindings)
			if err != nil {
				t.Fatal(err)
			}
			got := resolved[r.Requirements.Inputs[0].ID]
			if len(got.objects) != tt.matches || got.identityMapped != tt.mapped {
				t.Fatalf("resolution %#v", got)
			}
		})
	}
	r := requestFixture(t, "from events | fields id")
	r.Snapshot.Objects = append(r.Snapshot.Objects, r.Snapshot.Objects[0])
	r.Snapshot.Objects[2].ID = "events-other"
	r.Snapshot.Objects[2].App = "other"
	p, err := Prepare(r.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := p.resolveInputs(r.Requirements.Inputs, r.QueryScope, r.InputBindings)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved[r.Requirements.Inputs[0].ID].objects) != 2 {
		t.Fatal("scoped ambiguity lost")
	}
	r.QueryScope.App = environment.Selector{Values: []string{"app"}}
	resolved, err = p.resolveInputs(r.Requirements.Inputs, r.QueryScope, r.InputBindings)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved[r.Requirements.Inputs[0].ID].objects) != 1 {
		t.Fatal("scope was ignored")
	}
	r.InputBindings = []InputBinding{{InputID: r.Requirements.Inputs[0].ID, ObjectID: "absent", Expected: environment.ObjectIdentity{Kind: "index", Name: "events"}}}
	_, err = p.Check(r.assessment())
	requireRequestError(t, err, "binding_invalid", "/input_bindings/0/expected")
}
func TestPrepareMalformedOptionalSchema(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	if _, err := Check(r); err != nil {
		t.Fatal(err)
	}
	r.SchemaBundle = &environment.SchemaBundle{SchemaVersion: 1}
	_, err := Check(r)
	if detail, ok := RequestErrorDetails(err); !ok || detail.Code != "schema_bundle_invalid" {
		t.Fatalf("%v", err)
	}
}
func TestRequestCanonicalProducerEvidence(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/compatibility/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name     string                 `json:"name"`
		Document analysis.QueryDocument `json:"document"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			requirements, err := analysis.Requirements(fixture.Document)
			if err != nil {
				t.Fatal(err)
			}
			if err := validateRequirements(*requirements); err != nil {
				t.Fatal(err)
			}
			r := requestFixture(t, "from $events")
			r.Requirements = *requirements
			r.InputBindings = []InputBinding{}
			_, err = DecodeRequest(requestRaw(t, r))
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
func TestRequestCorrelationContradictions(t *testing.T) {
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	if len(r.Requirements.Correlation.Edges) != 1 {
		t.Fatal("fixture lost proved edge")
	}
	edge := &r.Requirements.Correlation.Edges[0]
	edge.Keys[0].Right.InputID = edge.Left.InputID
	_, err := Check(r)
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.Code != "requirements_inconsistent" || !strings.Contains(detail.Path, "/keys/0/right") {
		t.Fatalf("%v", err)
	}
}

func TestPrepareContradictoryBindingReuse(t *testing.T) {
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	r.InputBindings[1].Expected.Name = "users"
	_, err := Check(r)
	requireRequestError(t, err, "binding_invalid", "/input_bindings/1/object_id")
}
func TestRequestBindingRequiredBeforeInvalidQueryAggregation(t *testing.T) {
	r := requestFixture(t, "from $events | where )")
	if len(r.Requirements.Inputs) != 1 || r.Requirements.QueryStatus != analysis.Invalid {
		t.Fatal("fixture must retain invalid query and discovered placeholder")
	}
	r.InputBindings = []InputBinding{}
	_, err := Check(r)
	requireRequestError(t, err, "missing_input_binding", "/input_bindings")
}
func TestRequestStrictNestedSchemaAdmission(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	base := strings.TrimSuffix(string(requestRaw(t, r)), "}")
	for _, tt := range []struct{ schema, path string }{
		{`{"schema_version":1,"bundle_id":"a","schemas":[],"bindings":[]}`, "/schema_bundle/provenance"},
		{`{"schema_version":1,"schema_version":1,"bundle_id":"a","provenance":{},"schemas":[],"bindings":[]}`, "/schema_bundle/schema_version"},
		{`{"schema_version":1,"bundle_id":"a","provenance":{},"schemas":null,"bindings":[]}`, "/schema_bundle/schemas"},
	} {
		_, err := DecodeRequest([]byte(base + `,"schema_bundle":` + tt.schema + `}`))
		requireRequestError(t, err, "schema_bundle_invalid", tt.path)
	}
}
func TestRequestRequiredWireArrays(t *testing.T) {
	for _, path := range [][]string{{"input_coverage", "reasons"}, {"field_attribution_coverage", "reasons"}, {"correlation", "nodes"}, {"correlation", "edges"}, {"correlation", "components"}, {"coverage", "reasons"}} {
		t.Run(strings.Join(path, "/"), func(t *testing.T) {
			r := requestFixture(t, "from $events | fields id")
			var root map[string]any
			_ = json.Unmarshal(requestRaw(t, r), &root)
			requirements := root["requirements"].(map[string]any)
			part := requirements[path[0]].(map[string]any)
			delete(part, path[1])
			raw, _ := json.Marshal(root)
			_, err := DecodeRequest(raw)
			requireRequestError(t, err, "requirements_refresh_required", "/requirements/"+strings.Join(path, "/"))
		})
	}
}

func TestRequestGraphComponentContradiction(t *testing.T) {
	r := requestFixture(t, "from $events | append [from $users]")
	if len(r.Requirements.Correlation.Nodes) != 2 {
		t.Fatal("missing sources")
	}
	nodes := r.Requirements.Correlation.Nodes
	r.Requirements.Correlation.Components = [][]string{{nodes[0].OccurrenceID, nodes[1].OccurrenceID}}
	_, err := Check(r)
	requireRequestError(t, err, "requirements_inconsistent", "/requirements/correlation/components")
}

func TestPrepareConcurrentBindingIsolation(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	p, err := Prepare(r.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	first := r.assessment()
	second := detach(first)
	second.InputBindings[0].ObjectID = "users"
	second.InputBindings[0].Expected.Name = "users"
	var group sync.WaitGroup
	for i := 0; i < 20; i++ {
		request := first
		if i%2 != 0 {
			request = second
		}
		group.Add(1)
		go func() {
			defer group.Done()
			report, err := p.Check(request)
			if err != nil {
				t.Error(err)
				return
			}
			if report.Inputs[0].Objects[0].ObjectID != request.InputBindings[0].ObjectID {
				t.Error("concurrent bindings leaked")
			}
		}()
	}
	group.Wait()
}

func TestRequestCorrelationKnownReferenceEvidence(t *testing.T) {
	for _, test := range []struct {
		name, path string
		mutate     func(*analysis.CorrelationEdge)
	}{
		{"known field identity", "/requirements/correlation/edges/0/left/field_identity", func(e *analysis.CorrelationEdge) {
			e.Left.FieldIdentity.Segments = []string{"contradiction"}
			e.Keys[0].Left.FieldIdentity.Segments = []string{"contradiction"}
		}},
		{"known source location", "/requirements/correlation/edges/0/left/location", func(e *analysis.CorrelationEdge) {
			e.Left.Location.Start.Offset++
			e.Keys[0].Left.Location.Start.Offset++
		}},
		{"known non-field reference", "/requirements/correlation/edges/0/left/reference_ids", func(e *analysis.CorrelationEdge) {
			e.Left.ReferenceIDs = []string{"ref-0"}
			e.Keys[0].Left.ReferenceIDs = []string{"ref-0"}
		}},
		{"first key unknown reference", "/requirements/correlation/edges/0/keys/0/left", func(e *analysis.CorrelationEdge) { e.Keys[0].Left.ReferenceIDs = []string{"unknown-derived-reference"} }},
		{"first key unknown field", "/requirements/correlation/edges/0/keys/0/left", func(e *analysis.CorrelationEdge) {
			e.Left.ReferenceIDs = []string{"unknown-derived-reference"}
			e.Keys[0].Left.ReferenceIDs = []string{"unknown-derived-reference"}
			e.Keys[0].Left.FieldIdentity.Segments = []string{"different"}
		}},
		{"first key unknown location", "/requirements/correlation/edges/0/keys/0/left", func(e *analysis.CorrelationEdge) {
			e.Left.ReferenceIDs = []string{"unknown-derived-reference"}
			e.Keys[0].Left.ReferenceIDs = []string{"unknown-derived-reference"}
			e.Keys[0].Left.Location.Start.Offset++
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
			test.mutate(&r.Requirements.Correlation.Edges[0])
			requireBothRequestErrors(t, r, "requirements_inconsistent", test.path)
		})
	}
}
func TestRequestCorrelationAvailableReferenceAuthority(t *testing.T) {
	for _, text := range []string{
		`from $events | join left=e right=u where e.id=u.uid AND u.region=e.region [from $users]`,
		`from $events | join left=e right=u where e.id=u.asset_key [from $users | rename id AS asset_key]`,
		`from $events | eval key=id+region | join left=e right=u where e.key=u.uid [from $users]`,
		`from $events | eval k=1 | join left=e right=u where e.k=u.uid [from $users]`,
		`$left = from $events; $right = from $users; $pair = from $left | join left=e right=u where e.id=u.uid [from $right]; $out = from $pair | union [from $pair];`,
	} {
		t.Run(text, func(t *testing.T) {
			r := requestFixture(t, text)
			if _, err := Check(r); err != nil {
				t.Fatal(err)
			}
			if _, err := CheckJSON(requestRaw(t, r)); err != nil {
				t.Fatal(err)
			}
		})
	}
	// A requirement may retain a conditional destination identity without proving
	// a source identity. Its exact reference location still remains authoritative.
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	for i := range r.Requirements.Items {
		item := &r.Requirements.Items[i]
		for _, o := range item.Occurrences {
			if o.ReferenceID == r.Requirements.Correlation.Edges[0].Left.ReferenceIDs[0] {
				item.InputID = ""
				item.Ownership.State = "unproved"
				item.Necessity = "conditional"
				item.Identity = "destination"
				item.FieldIdentity.Segments = []string{"destination"}
				item.Occurrences[0].Binding = "indeterminate"
				item.Occurrences[0].Necessity = "conditional"
			}
		}
	}
	if _, err := Check(r); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckJSON(requestRaw(t, r)); err != nil {
		t.Fatal(err)
	}
	conditional := detach(r)
	conditional.Requirements.Correlation.Edges[0].Left.Location.Start.Offset++
	conditional.Requirements.Correlation.Edges[0].Keys[0].Left.Location.Start.Offset++
	requireBothRequestErrors(t, conditional, "requirements_inconsistent", "/requirements/correlation/edges/0/left/location")
	// Unknown derived references are intentionally outside the partial Items inventory.
	r = requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	e := &r.Requirements.Correlation.Edges[0]
	e.Left.ReferenceIDs = []string{"unknown-derived-reference"}
	e.Keys[0].Left.ReferenceIDs = []string{"unknown-derived-reference"}
	if _, err := Check(r); err != nil {
		t.Fatal(err)
	}
	if _, err := CheckJSON(requestRaw(t, r)); err != nil {
		t.Fatal(err)
	}
}
func TestRequestCorrelationAtomicDotDistinctFromPath(t *testing.T) {
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	e := &r.Requirements.Correlation.Edges[0]
	for i := range r.Requirements.Items {
		item := &r.Requirements.Items[i]
		for _, o := range item.Occurrences {
			if o.ReferenceID == e.Left.ReferenceIDs[0] {
				item.Identity = "actor.id"
				item.FieldIdentity = &analysis.FieldIdentity{Kind: "atomic", Segments: []string{"actor.id"}}
			}
		}
	}
	e.Left.FieldIdentity = analysis.FieldIdentity{Kind: "path", Segments: []string{"actor", "id"}}
	e.Keys[0].Left.FieldIdentity = detach(e.Left.FieldIdentity)
	requireBothRequestErrors(t, r, "requirements_inconsistent", "/requirements/correlation/edges/0/left/field_identity")
}

func requireBothRequestErrors(t *testing.T, r Request, code, path string) {
	t.Helper()
	t.Run("typed", func(t *testing.T) { _, err := Check(r); requireRequestError(t, err, code, path) })
	t.Run("json", func(t *testing.T) { _, err := CheckJSON(requestRaw(t, r)); requireRequestError(t, err, code, path) })
}
func TestRequestMissingInputBindingCode(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	r.InputBindings = []InputBinding{}
	requireBothRequestErrors(t, r, "missing_input_binding", "/input_bindings")
}

func TestPrepareAutomaticObjectEvidenceDetached(t *testing.T) {
	r := requestFixture(t, "from events | fields id")
	r.Snapshot.Objects[0].Document = &analysis.QueryDocument{Text: "search"}
	start, end := 0, 6
	r.Snapshot.Objects[0].Relations = []closure.Relation{{Kind: "lookup", Name: "table", Start: &start, End: &end}}
	p, err := Prepare(r.Snapshot, nil)
	if err != nil {
		t.Fatal(err)
	}
	first, err := p.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	object := first.Inputs[0].Objects[0].Object
	object.Document.Text = "report changed"
	*object.Relations[0].Start = 1
	object.Relations[0].Name = "report changed"
	second, err := p.Check(r.assessment())
	if err != nil {
		t.Fatal(err)
	}
	object = second.Inputs[0].Objects[0].Object
	if object.Document.Text != "search" || *object.Relations[0].Start != 0 || object.Relations[0].Name != "table" {
		t.Fatalf("returned automatic object evidence aliases prepared snapshot: %#v %#v", object.Document, object.Relations)
	}
}

func TestRequestCorrelationAuthoredPositiveChecks(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/compatibility/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name         string                    `json:"name"`
		Document     analysis.QueryDocument    `json:"document"`
		Snapshot     environment.Snapshot      `json:"snapshot"`
		SchemaBundle *environment.SchemaBundle `json:"schema_bundle"`
		QueryScope   environment.CaptureScope  `json:"query_scope"`
		Bindings     map[string]struct {
			ObjectID string                     `json:"object_id"`
			Expected environment.ObjectIdentity `json:"expected"`
			SchemaID string                     `json:"schema_id"`
		} `json:"bindings"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, fixture := range fixtures {
		if fixture.Name != "pipeline-three-input-renamed-key" && fixture.Name != "same-name-key-chain-proved-rename" {
			continue
		}
		checked++
		t.Run(fixture.Name, func(t *testing.T) {
			requirements, err := analysis.Requirements(fixture.Document)
			if err != nil {
				t.Fatal(err)
			}
			r := Request{SchemaVersion: 1, Requirements: *requirements, QueryScope: fixture.QueryScope, InputBindings: []InputBinding{}, Snapshot: fixture.Snapshot, SchemaBundle: fixture.SchemaBundle}
			for _, input := range requirements.Inputs {
				if binding, found := fixture.Bindings[input.Kind+":"+input.Name]; found {
					r.InputBindings = append(r.InputBindings, InputBinding{InputID: input.ID, ObjectID: binding.ObjectID, Expected: binding.Expected, SchemaID: binding.SchemaID})
				}
			}
			if _, err := Check(r); err != nil {
				t.Fatal(err)
			}
			if _, err := CheckJSON(requestRaw(t, r)); err != nil {
				t.Fatal(err)
			}
		})
	}
	if checked != 2 {
		t.Fatalf("missing authored positive fixtures: %d", checked)
	}
}
func TestRequestViewReferenceCoordinates(t *testing.T) {
	text := `$left = from $events; $right = from $users; $pair = from $left | join left=e right=u where e.id=u.uid [from $right]; $out = from $pair | union [from $pair];`
	for _, test := range []struct {
		name, path string
		mutate     func(*Request)
	}{
		{"original source", "/requirements/inputs/0/occurrences/1/original_reference_id", func(r *Request) { r.Requirements.Inputs[0].Occurrences[1].Location.Start.Offset++ }},
		{"terminal source", "/requirements/inputs/1/occurrences/0/use_site_locations/1", func(r *Request) { r.Requirements.Inputs[1].Occurrences[0].UseSiteLocations[1].Start.Offset++ }},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := requestFixture(t, text)
			test.mutate(&r)
			requireBothRequestErrors(t, r, "requirements_inconsistent", test.path)
		})
	}
}

func TestRequestRepeatedReferenceTypedSourceIdentity(t *testing.T) {
	for _, test := range []struct {
		name          string
		first, second analysis.FieldIdentity
	}{
		{"atomic dot versus path", analysis.FieldIdentity{Kind: "atomic", Segments: []string{"actor.id"}}, analysis.FieldIdentity{Kind: "path", Segments: []string{"actor", "id"}}},
		{"different path qualifier", analysis.FieldIdentity{Kind: "path", Qualifier: "first", Segments: []string{"actor", "id"}}, analysis.FieldIdentity{Kind: "path", Qualifier: "second", Segments: []string{"actor", "id"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := requestFixture(t, "from $events | fields id")
			item := &r.Requirements.Items[1]
			item.Identity = "actor.id"
			item.FieldIdentity = &test.first
			second := detach(*item)
			second.ID = "req-typed-reuse"
			second.FieldIdentity = &test.second
			r.Requirements.Items = append(r.Requirements.Items, second)
			requireBothRequestErrors(t, r, "requirements_inconsistent", "/requirements/items/2/occurrences/0/reference_id")
		})
	}
}
func TestRequestRepeatedReferenceAuthorityControls(t *testing.T) {
	for _, unproved := range []bool{false, true} {
		name := "identical proved source"
		if unproved {
			name = "unproved destination"
		}
		t.Run(name, func(t *testing.T) {
			r := requestFixture(t, "from $events | fields id")
			item := &r.Requirements.Items[1]
			item.Identity = "actor.id"
			item.FieldIdentity = &analysis.FieldIdentity{Kind: "atomic", Segments: []string{"actor.id"}}
			if unproved {
				item.InputID = ""
				item.Ownership.State = "unproved"
				item.Necessity = "conditional"
				item.Occurrences[0].Binding = "indeterminate"
				item.Occurrences[0].Necessity = "conditional"
			}
			second := detach(*item)
			second.ID = "req-typed-reuse"
			if unproved {
				second.FieldIdentity = &analysis.FieldIdentity{Kind: "path", Segments: []string{"actor", "id"}}
			}
			r.Requirements.Items = append(r.Requirements.Items, second)
			if _, err := Check(r); err != nil {
				t.Fatal(err)
			}
			if _, err := CheckJSON(requestRaw(t, r)); err != nil {
				t.Fatal(err)
			}
		})
	}
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	second := detach(r.Requirements.Items[1])
	second.ID = "req-typed-reuse"
	second.InputID = r.Requirements.Inputs[1].ID
	second.Ownership.CandidateInputIDs = []string{second.InputID}
	second.Occurrences[0].InputOccurrenceIDs = []string{r.Requirements.Inputs[1].Occurrences[0].ID}
	r.Requirements.Items = append(r.Requirements.Items, second)
	requireBothRequestErrors(t, r, "requirements_inconsistent", "/requirements/items/4/occurrences/0/reference_id")
}
func TestRequestRepeatedReferenceRetainsSourceAuthority(t *testing.T) {
	r := requestFixture(t, "from $events | join left=e right=u where e.id=u.id [from $users]")
	second := detach(r.Requirements.Items[1])
	second.ID = "req-weaker-reuse"
	second.Resolution = "wildcard"
	second.FieldIdentity = nil
	r.Requirements.Items = append(r.Requirements.Items, second)
	e := &r.Requirements.Correlation.Edges[0]
	e.Left.FieldIdentity.Segments = []string{"contradiction"}
	e.Keys[0].Left.FieldIdentity.Segments = []string{"contradiction"}
	requireBothRequestErrors(t, r, "requirements_inconsistent", "/requirements/correlation/edges/0/left/field_identity")
}
