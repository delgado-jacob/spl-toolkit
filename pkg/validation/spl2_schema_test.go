package validation

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestSPL2SchemaClosedSource(t *testing.T) {
	target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(
		`{"type":"object","properties":{"host":{"type":"string"}},"required":["host"],"additionalProperties":false}`)}
	report, err := ValidateSchema(analysis.QueryDocument{
		Text: "FROM main SELECT host", Language: "spl2", SourceID: "core-window",
	}, target)
	if err != nil || report.Status != analysis.Valid || !report.Coverage.SchemaComplete {
		t.Fatalf("closed exact source validation: %#v %v", report, err)
	}
	evidence := []SchemaEvidence{{ResourceURI: "urn:spl-toolkit:json-schema:root", Pointer: "/properties/host", Keyword: "properties", DeclarationBasis: "property", Requirement: "required"}}
	want := []SchemaReferenceOutcome{{ReferenceID: "ref-1", Outcome: "required", MatchesComplete: true, Matches: []SchemaMatch{{Name: "host", Binding: "source", Outcome: "required", Evidence: evidence}}, Evidence: evidence, SupportingClasses: []SchemaClass{}, MissingClasses: []SchemaClass{}, IndeterminateClasses: []SchemaClass{}}}
	if !reflect.DeepEqual(report.Outcomes, want) {
		t.Fatalf("full source evidence: %+v", report.Outcomes)
	}
	targetInfo := SchemaTargetInfo{Kind: "json_schema", Dialect: "https://json-schema.org/draft/2020-12/schema", BaseURI: "urn:spl-toolkit:json-schema:root", ResourceURIs: []string{"urn:spl-toolkit:json-schema:root"}, Members: []SchemaClass{}, Extensions: []SchemaExtension{}, Limitations: []string{"array_traversal"}}
	if !reflect.DeepEqual(report.Target, targetInfo) {
		t.Fatalf("owned target metadata: %+v", report.Target)
	}
	spl2SchemaWrapperParity(t, report.Analysis.Document, target, report)
}

func spl2SchemaWrapperParity(t *testing.T, doc analysis.QueryDocument, target SchemaTarget, want *SchemaReport) {
	t.Helper()
	raw, err := json.Marshal(SchemaRequest{Document: doc, Target: target})
	if err != nil {
		t.Fatal(err)
	}
	request, err := DecodeSchemaRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ValidateSchema(request.Document, request.Target)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("schema strict wrapper mismatch: %v", err)
	}
}

func TestSPL2SchemaCanonicalBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name, text, schema string
		status             analysis.Status
		complete           bool
		outcomes           []string
	}{
		{"absent host", `FROM main SELECT host`, `{"type":"object","additionalProperties":false}`, analysis.Invalid, true, []string{"missing"}},
		{"unknown output", `FROM main | mystery x=1 | table host`, `{"type":"object","properties":{"host":true},"additionalProperties":false}`, analysis.Incomplete, false, []string{"indeterminate"}},
		{"open source", `FROM main SELECT host`, `{"type":"object"}`, analysis.Valid, true, []string{"permitted_unspecified"}},
		{"conditional source", `FROM main SELECT host`, `{"anyOf":[{"type":"object","properties":{"host":true},"additionalProperties":false},{"type":"object","additionalProperties":false}]}`, analysis.Incomplete, false, []string{"conditional"}},
		{"quoted literal against nested only", `FROM main SELECT 'actor.name'`, `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true},"additionalProperties":false}},"additionalProperties":false}`, analysis.Incomplete, false, []string{"indeterminate"}},
		{"navigation against dotted literal only", `FROM main SELECT actor.name`, `{"type":"object","properties":{"actor.name":true},"additionalProperties":false}`, analysis.Invalid, false, []string{"missing", "indeterminate"}},
		{"navigation base survives", `FROM main SELECT actor.name`, `{"type":"object","properties":{"actor":{"type":"object","properties":{"name":true},"additionalProperties":false}},"additionalProperties":false}`, analysis.Incomplete, false, []string{"optional", "indeterminate"}},
		{"literal local source", `FROM [{host:"a"}] SELECT host`, `false`, analysis.Valid, true, []string{"matching"}},
		{"derived output", `FROM main SELECT 1 AS local | table local`, `false`, analysis.Valid, true, []string{"matching"}},
		{"conditional local key", `FROM [{host:"a"},{other:1}] SELECT host`, `false`, analysis.Incomplete, false, []string{"indeterminate"}},
		{"null inspection", `FROM main SELECT isnull(absent) AS answer`, `false`, analysis.Valid, true, []string{}},
		{"definite error and unknown", `FROM main | table absent | mystery x=1`, `{"type":"object","additionalProperties":false}`, analysis.Invalid, false, []string{"missing"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := analysis.QueryDocument{Text: tt.text, Language: "spl2", SourceID: "source\r\n"}
			target := SchemaTarget{Kind: "json_schema", Identity: tt.name, Schema: json.RawMessage(tt.schema)}
			r, err := ValidateSchema(doc, target)
			if err != nil {
				t.Fatal(err)
			}
			outcomes := []string{}
			for _, o := range r.Outcomes {
				outcomes = append(outcomes, o.Outcome)
				if o.Matches == nil || o.Evidence == nil || o.SupportingClasses == nil || o.MissingClasses == nil || o.IndeterminateClasses == nil {
					t.Fatal("nil schema outcome collections")
				}
				if o.Outcome == "indeterminate" || o.Outcome == "conditional" {
					if o.MatchesComplete || len(o.Matches) > 0 {
						t.Fatalf("guessed path/membership: %+v", o)
					}
				}
			}
			if r.Status != tt.status || r.Coverage.SchemaComplete != tt.complete || !reflect.DeepEqual(outcomes, tt.outcomes) {
				t.Fatalf("canonical: %+v outcomes %v", r, outcomes)
			}
			if r.Target.Identity != target.Identity || r.Analysis.Document.SourceID != doc.SourceID {
				t.Fatal("metadata changed")
			}
			spl2SchemaWrapperParity(t, doc, target, r)
		})
	}
}

func TestSPL2SchemaWildcardPartialEvidence(t *testing.T) {
	// These controls preserve the accepted shared fields/internal-retention
	// policy. Schema admission is separate from source environment conditionality:
	// an partial unobserved source loses certainty before candidate expansion, while
	// a prior exact read and the local derived output retain independent evidence.
	// A closed root with generic true leaves still has unbounded descendant names.
	for _, tt := range []struct {
		closed, observed, scalar bool
		matches                  []string
	}{
		{true, false, true, []string{"host", "local"}},
		{true, false, false, []string{"local"}},
		{false, false, false, []string{"local"}},
		{false, true, false, []string{"host", "local"}},
	} {
		schema := `{"type":"object","properties":{"host":true,"actor.name":true}}`
		if tt.closed {
			schema = `{"type":"object","properties":{"host":true,"actor.name":true},"additionalProperties":false}`
		}
		if tt.scalar {
			schema = `{"type":"object","properties":{"host":{"type":"string"},"actor.name":{"type":"string"}},"additionalProperties":false}`
		}
		text := `FROM main | eval local=1 | fields '*'`
		if tt.observed {
			text = `FROM main | where host=true | eval local=1 | fields '*'`
		}
		target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(schema)}
		doc := analysis.QueryDocument{Text: text, Language: "spl2"}
		r, err := ValidateSchema(doc, target)
		if err != nil {
			t.Fatal(err)
		}
		if r.Status != analysis.Incomplete || r.Coverage.SchemaComplete {
			t.Fatalf("partial wildcard %+v", r)
		}
		o := r.Outcomes[len(r.Outcomes)-1]
		names := []string{}
		for _, m := range o.Matches {
			names = append(names, m.Name)
			if m.Name == "host" && (m.Binding != "source" || len(m.Evidence) == 0) {
				t.Fatalf("flat source evidence %+v", m)
			}
			if m.Name == "local" && (m.Binding != "derived" || !reflect.DeepEqual(m.Evidence, []SchemaEvidence{{DeclarationBasis: "derived"}})) {
				t.Fatalf("derived evidence %+v", m)
			}
		}
		if o.Outcome != "indeterminate" || o.MatchesComplete || !reflect.DeepEqual(names, tt.matches) {
			t.Fatalf("lost partial evidence closed=%v observed=%v %+v", tt.closed, tt.observed, o)
		}
		for _, f := range r.Analysis.Lineage[len(r.Analysis.Lineage)-1].After.Fields {
			if f.Name == "actor.name" {
				t.Fatal("ambiguous candidate materialized")
			}
		}
		spl2SchemaWrapperParity(t, doc, target, r)
	}
}

func TestSPL2SchemaMixedBatchAndOwnedOCSF(t *testing.T) {
	target := SchemaTarget{Kind: "ocsf", Identity: "local-test", Catalog: edgeOCSF(t), Selection: &OCSFSelection{Version: "1.6.0", Category: "test", Profiles: []string{}, Extensions: []string{}}}
	docs := []analysis.QueryDocument{{Text: "table time", SourceID: "spl"}, {Text: "FROM main SELECT time", Language: "spl2", SourceID: "spl2"}, {Text: "FROM main SELECT recommended", Language: "spl2", SourceID: "conditional"}, {Text: "FROM main | eval n=1 | appendpipe [table n]", Language: "spl2", SourceID: "child"}}
	calls := 0
	got, err := validateSchemaBatchWith(docs, target, func(t SchemaTarget) (preparedSchemaTarget, error) { calls++; return prepareSchemaTarget(t) })
	if err != nil || calls != 1 || got.Status != analysis.Incomplete || len(got.Reports) != 4 {
		t.Fatalf("mixed one prepare: %+v %v calls %d", got, err, calls)
	}
	for i, doc := range docs {
		single, e := ValidateSchema(doc, target)
		if e != nil || !reflect.DeepEqual(single, got.Reports[i]) || got.Reports[i].Analysis.Document.SourceID != doc.SourceID {
			t.Fatalf("mixed report %d %v", i, e)
		}
	}
	classes := []SchemaClass{{Key: "a", UID: 1001}, {Key: "b", UID: 1002}}
	o := got.Reports[1].Outcomes[0]
	if !o.MatchesComplete || !reflect.DeepEqual(o.SupportingClasses, classes) || len(o.Evidence) != 2 || len(o.Matches) != 1 || len(o.Matches[0].Evidence) != 2 {
		t.Fatalf("OCSF class evidence %+v", o)
	}
	conditional := got.Reports[2].Outcomes[0]
	if conditional.MatchesComplete || conditional.Outcome != "conditional" || !reflect.DeepEqual(conditional.SupportingClasses, classes[:1]) || !reflect.DeepEqual(conditional.MissingClasses, classes[1:]) {
		t.Fatalf("OCSF conditional classes %+v", conditional)
	}
	raw, _ := json.Marshal(SchemaBatchRequest{Documents: docs, Target: target})
	decoded, err := DecodeSchemaBatchRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	wrapped, err := ValidateSchemaBatch(decoded.Documents, decoded.Target)
	if err != nil || !reflect.DeepEqual(wrapped, got) {
		t.Fatal("mixed schema wrapper mismatch")
	}
	got.Reports[0].Target.Members[0].Key = "mutated"
	got.Reports[0].Target.Selection.Profiles = append(got.Reports[0].Target.Selection.Profiles, "mutated")
	*got.Reports[0].Outcomes[0].Evidence[0].ClassUID = 99
	if !reflect.DeepEqual(got.Reports[1].Target.Members, classes) || len(got.Reports[1].Target.Selection.Profiles) != 0 || *got.Reports[1].Outcomes[0].Evidence[0].ClassUID == 99 {
		t.Fatal("schema metadata or evidence shared across reports")
	}
	for _, bad := range []analysis.QueryDocument{{Text: "FROM main", Language: "sql"}, {Text: "FROM main", Language: "spl2", Profile: "edge"}, {Text: "FROM main", Language: "spl2", Version: "next"}, {Text: string([]byte{255}), Language: "spl2"}, {Text: "FROM main", Language: "spl2", SourceID: string([]byte{255})}} {
		if report, e := ValidateSchema(bad, target); report != nil || !IsInputError(e) {
			t.Fatalf("schema input error: %+v %v", report, e)
		}
		count := 0
		batch, e := validateSchemaBatchWith(append(append([]analysis.QueryDocument{}, docs...), bad), target, func(t SchemaTarget) (preparedSchemaTarget, error) { count++; return prepareSchemaTarget(t) })
		if batch != nil || !IsInputError(e) || count != 0 {
			t.Fatalf("invalid batch partially prepared: %+v %v count %d", batch, e, count)
		}
	}
	for _, raw := range []string{`{"document":{"text":"FROM main","language":"SPL2"},"target":{"kind":"json_schema","schema":true}}`, `{"document":{"text":"FROM main","text":"FROM other"},"target":{"kind":"json_schema","schema":true}}`, `{"document":{"text":"\ud800"},"target":{"kind":"json_schema","schema":true}}`} {
		if _, e := DecodeSchemaRequest([]byte(raw)); !IsInputError(e) {
			t.Fatalf("strict schema single accepted %s", raw)
		}
	}
	if _, e := DecodeSchemaBatchRequest([]byte(`{"documents":[{"text":"FROM main","language":"spl2"},{"text":"x","language":"sql"}],"target":{"kind":"json_schema","schema":true}}`)); !IsInputError(e) {
		t.Fatal("strict batch accepted bad document")
	}
}
