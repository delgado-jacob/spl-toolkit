package validation

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestValidateSchemaInvalidWinsHistoricalIncomplete(t *testing.T) {
	target := SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(`{"type":"object","properties":{"host":true},"additionalProperties":false}`)}
	report, err := ValidateSchema(analysis.QueryDocument{Text: "search absent=x | mystery host"}, target)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Invalid || report.Coverage.SemanticComplete {
		t.Fatalf("report=%+v", report)
	}
	found := false
	for _, outcome := range report.Outcomes {
		found = found || outcome.Outcome == "missing"
	}
	if !found {
		t.Fatalf("no definite missing obligation: %+v", report.Outcomes)
	}
}

func TestSchemaCanonicalOutcomes(t *testing.T) {
	for _, tc := range []struct {
		name, schema, text, outcomes string
		status                       analysis.Status
		complete                     bool
	}{
		{"closed", `{"type":"object","properties":{"host":true},"required":["host"],"additionalProperties":false}`, "table host absent", "required,missing", analysis.Invalid, true},
		{"partial", `{"type":"object","properties":{"host":true}}`, "eval label=1 | table *", "indeterminate", analysis.Incomplete, false},
		{"derived", `false`, "eval label=1 | table label", "matching", analysis.Valid, true},
		{"removal", `false`, "fields -absent*", "", analysis.Valid, true},
		{"unavailable", `{"type":"object","properties":{"host":true},"additionalProperties":false}`, "fields -host | table host", "unavailable", analysis.Invalid, true},
		{"conditional", `{"anyOf":[{"type":"object","properties":{"host":true},"additionalProperties":false},{"type":"object","additionalProperties":false}]}`, "table host", "conditional", analysis.Incomplete, false},
		{"recovery", `{"type":"object","properties":{"host":true}}`, "table * | stats count AS total | table total*", "indeterminate,matching", analysis.Incomplete, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, e := ValidateSchema(analysis.QueryDocument{Text: tc.text}, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(tc.schema)})
			if e != nil {
				t.Fatal(e)
			}
			got := ""
			for i, o := range r.Outcomes {
				if i > 0 {
					got += ","
				}
				got += o.Outcome
				if o.Matches == nil || o.Evidence == nil || o.SupportingClasses == nil || o.MissingClasses == nil || o.IndeterminateClasses == nil {
					t.Fatal("null outcome arrays")
				}
			}
			if got != tc.outcomes || r.Status != tc.status || r.Coverage.SchemaComplete != tc.complete {
				t.Fatalf("outcomes=%s report=%+v", got, r)
			}
			if tc.name == "partial" {
				o := r.Outcomes[0]
				if o.MatchesComplete || len(o.Matches) != 2 || o.Matches[0].Name != "host" || o.Matches[0].Binding != "source" || o.Matches[1].Name != "label" || o.Matches[1].Binding != "derived" {
					t.Fatalf("lost partial matches: %+v", o)
				}
			}
			if tc.name == "conditional" && (r.Outcomes[0].MatchesComplete || len(r.Outcomes[0].Matches) != 0) {
				t.Fatal("conditional fabricated membership")
			}
		})
	}
}

// Frozen full reports are paired with independently authored semantic summaries.
// No test regenerates an expected report.
type schemaCorpusCase struct {
	ID             string                 `json:"id"`
	Target         SchemaTarget           `json:"target"`
	CatalogFixture string                 `json:"catalog_fixture,omitempty"`
	Document       analysis.QueryDocument `json:"document"`
	Outcomes       string                 `json:"outcomes"`
	Status         analysis.Status        `json:"status"`
	SchemaComplete bool                   `json:"schema_complete"`
	Expected       json.RawMessage        `json:"expected"`
}

func loadSchemaCorpus(t *testing.T) []schemaCorpusCase {
	t.Helper()
	raw, e := os.ReadFile("../../testdata/schemas/cases.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		SchemaVersion int                `json:"schema_version"`
		Cases         []schemaCorpusCase `json:"cases"`
	}
	if e = json.Unmarshal(raw, &corpus); e != nil {
		t.Fatal(e)
	}
	if corpus.SchemaVersion != 1 || len(corpus.Cases) < 20 {
		t.Fatal("incomplete schema corpus")
	}
	for i := range corpus.Cases {
		tc := &corpus.Cases[i]
		if tc.CatalogFixture != "" {
			raw, e := os.ReadFile(filepath.Join("../../testdata/schemas", tc.CatalogFixture))
			if e != nil {
				t.Fatal(e)
			}
			if strings.HasSuffix(tc.CatalogFixture, ".gz") {
				z, e := gzip.NewReader(bytes.NewReader(raw))
				if e != nil {
					t.Fatal(e)
				}
				raw, e = io.ReadAll(z)
				z.Close()
				if e != nil {
					t.Fatal(e)
				}
			}
			tc.Target.Catalog = raw
		}
	}
	return corpus.Cases
}
func checkSchemaCorpusSemantics(t *testing.T, tc schemaCorpusCase, r *SchemaReport) {
	t.Helper()
	var parts []string
	refs := map[string]analysis.Reference{}
	for _, ref := range r.Analysis.References {
		refs[ref.ID] = ref
		if tc.Document.Text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
			t.Fatalf("bad original slice %+v", ref)
		}
	}
	for _, o := range r.Outcomes {
		ref, ok := refs[o.ReferenceID]
		if !ok || ref.Role == "remove" || ref.Binding == "not_applicable" || ref.Kind != "field" {
			t.Fatalf("noncanonical obligation %+v", o)
		}
		parts = append(parts, ref.NormalizedName+":"+o.Outcome)
		if o.Matches == nil || o.Evidence == nil || o.SupportingClasses == nil || o.MissingClasses == nil || o.IndeterminateClasses == nil {
			t.Fatal("null arrays")
		}
	}
	if got := strings.Join(parts, ","); got != tc.Outcomes || r.Status != tc.Status || r.Coverage.SchemaComplete != tc.SchemaComplete {
		t.Fatalf("%s got %s %s complete=%t; want %s %s complete=%t", tc.ID, got, r.Status, r.Coverage.SchemaComplete, tc.Outcomes, tc.Status, tc.SchemaComplete)
	}
	for _, d := range r.Diagnostics {
		if d.Code == CodeUnknownField || d.Code == CodeIndeterminateField {
			found := false
			for _, ref := range refs {
				found = found || (d.Location == ref.Location && d.StageID == ref.StageID && d.ScopeID == ref.ScopeID)
			}
			if !found {
				t.Fatalf("unlocated schema finding %+v", d)
			}
		}
	}
}
func TestSchemaFrozenCorpus(t *testing.T) {
	for _, tc := range loadSchemaCorpus(t) {
		t.Run(tc.ID, func(t *testing.T) {
			r, e := ValidateSchema(tc.Document, tc.Target)
			if e != nil {
				t.Fatal(e)
			}
			checkSchemaCorpusSemantics(t, tc, r)
			actual, e := json.Marshal(r)
			if e != nil {
				t.Fatal(e)
			}
			var frozen SchemaReport
			if e = json.Unmarshal(tc.Expected, &frozen); e != nil {
				t.Fatal(e)
			}
			checkSchemaCorpusSemantics(t, tc, &frozen)
			expected, _ := json.Marshal(&frozen)
			if !bytes.Equal(actual, expected) {
				t.Fatalf("full canonical report changed for %s", tc.ID)
			}
			batch, e := ValidateSchemaBatch([]analysis.QueryDocument{tc.Document, tc.Document}, tc.Target)
			if e != nil {
				t.Fatal(e)
			}
			for _, r := range batch.Reports {
				got, _ := json.Marshal(r)
				if !bytes.Equal(got, expected) {
					t.Fatal("batch report differs")
				}
			}
		})
	}
}
func TestSchemaConcurrentPreparedAndOwnership(t *testing.T) {
	uid := int64(1001)
	target := SchemaTarget{Kind: "ocsf", Catalog: edgeOCSF(t), Selection: &OCSFSelection{Version: "1.6.0", ClassUID: &uid, Profiles: []string{"p"}}}
	prepared, e := prepareSchemaTarget(target)
	if e != nil {
		t.Fatal(e)
	}
	doc := analysis.QueryDocument{Text: "table time profiled recommended"}
	baseline, e := validatePreparedSchema(doc, prepared)
	if e != nil {
		t.Fatal(e)
	}
	frozen, _ := json.Marshal(baseline)
	// Changing caller-owned raw data and pointer/array values cannot affect the prepared target.
	target.Catalog[0] = '!'
	uid = 999
	target.Selection.Profiles[0] = "changed"
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, e := validatePreparedSchema(doc, prepared)
			if e != nil {
				t.Error(e)
				return
			}
			got, _ := json.Marshal(r)
			if !bytes.Equal(got, frozen) {
				t.Error("prepared result changed")
			}
		}()
	}
	wg.Wait()
	batch, e := validateSchemaBatchPrepared([]analysis.QueryDocument{doc, doc}, prepared)
	if e != nil {
		t.Fatal(e)
	}
	first := batch.Reports[0]
	first.Target.Members[0].Key = "changed"
	first.Target.Selection.Profiles[0] = "changed"
	if first.Target.Selection.ClassUID != nil {
		*first.Target.Selection.ClassUID = 999
	}
	*first.Outcomes[0].Evidence[0].ClassUID = 999
	*first.Outcomes[0].Matches[0].Evidence[0].ClassUID = 888
	got, _ := json.Marshal(batch.Reports[1])
	if !bytes.Equal(got, frozen) {
		t.Fatal("batch reports alias")
	}
	r, e := validatePreparedSchema(doc, prepared)
	if e != nil {
		t.Fatal(e)
	}
	got, _ = json.Marshal(r)
	if !bytes.Equal(got, frozen) {
		t.Fatal("report aliases prepared target")
	}
}
func TestM3FinalizationByteParity(t *testing.T) {
	raw, e := os.ReadFile("../../testdata/validation/cases.json")
	if e != nil {
		t.Fatal(e)
	}
	var corpus struct {
		Cases []struct {
			Document analysis.QueryDocument `json:"document"`
			Catalog  FieldCatalog           `json:"catalog"`
			Expected Report                 `json:"expected"`
		} `json:"cases"`
	}
	if e = json.Unmarshal(raw, &corpus); e != nil {
		t.Fatal(e)
	}
	for i, tc := range corpus.Cases {
		r, e := Validate(tc.Document, tc.Catalog)
		if e != nil {
			t.Fatal(e)
		}
		got, _ := json.Marshal(r)
		want, _ := json.Marshal(tc.Expected)
		if !bytes.Equal(got, want) {
			t.Fatalf("M3 bytes changed case %d", i)
		}
	}
}

func TestSchemaAdmittedUnknownRequiredness(t *testing.T) {
	r, e := ValidateSchema(analysis.QueryDocument{Text: "table profiled"}, SchemaTarget{Kind: "ocsf", Catalog: edgeOCSF(t), Selection: &OCSFSelection{Version: "1.6.0", Class: "a", Profiles: []string{"p"}}})
	if e != nil {
		t.Fatal(e)
	}
	o := r.Outcomes[0]
	if r.Status != analysis.Incomplete || r.Coverage.SchemaComplete || o.Outcome != "indeterminate" || !o.MatchesComplete || len(o.Matches) != 1 || o.Matches[0].Binding != "source" || o.Matches[0].Outcome != "indeterminate" {
		t.Fatalf("membership/requiredness conflated: %+v", r)
	}
	for _, evs := range [][]SchemaEvidence{o.Evidence, o.Matches[0].Evidence} {
		found := false
		for _, ev := range evs {
			found = found || (ev.Reason == "ocsf_profile_requirement" && ev.Requirement == "unknown" && ev.Pointer == "/classes/a/attributes/profiled")
		}
		if !found {
			t.Fatal("lost profile requiredness provenance")
		}
	}
}

func TestSchemaSelectorLimitationEvidence(t *testing.T) {
	for _, tc := range []struct {
		schema, text string
		want         bool
	}{
		{`{"type":"object","properties":{"host":{"type":"string"}},"additionalProperties":false}`, "| mystery | table *", false},
		{`{"type":"object","properties":{"host":{"type":"string"}}}`, "table *", true},
	} {
		r, e := ValidateSchema(analysis.QueryDocument{Text: tc.text}, SchemaTarget{Kind: "json_schema", Schema: json.RawMessage(tc.schema)})
		if e != nil {
			t.Fatal(e)
		}
		o := r.Outcomes[len(r.Outcomes)-1]
		found := false
		for _, ev := range o.Evidence {
			found = found || ev.Reason == "partial_name_universe"
			if ev.Reason == "array_traversal" {
				t.Fatal("unrelated descriptive limitation became causal evidence")
			}
		}
		if found != tc.want {
			t.Fatalf("selector evidence %+v", o)
		}
	}
}

func TestSchemaEvidenceDistinctComponentsAndUIDOwnership(t *testing.T) {
	uid1, uid2 := int64(7), int64(7)
	input := []SchemaEvidence{{ResourceURI: "x\x00y", Pointer: "z", ClassUID: &uid1}, {ResourceURI: "x", Pointer: "y\x00z", ClassUID: &uid2}, {ResourceURI: "x", Pointer: "y\x00z", ClassUID: &uid1}}
	got := canonicalSchemaEvidence(input)
	if len(got) != 2 {
		t.Fatalf("lost distinct evidence: %+v", got)
	}
	*got[0].ClassUID = 99
	if uid1 != 7 || uid2 != 7 || *got[1].ClassUID != 7 {
		t.Fatal("UID evidence aliases input or sibling")
	}
}

func TestSchemaPublicUniverseLimitsDoNotPoisonExactReads(t *testing.T) {
	properties := map[string]any{"host": map[string]any{"type": "string"}}
	for i := 0; i < 5000; i++ {
		properties[fmt.Sprintf("field%04d", i)] = map[string]any{"type": "string"}
	}
	raw, e := json.Marshal(map[string]any{"type": "object", "properties": properties, "additionalProperties": false})
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		schema json.RawMessage
		reason string
	}{
		{raw, "enumeration_budget"},
		{json.RawMessage(`{"type":"object","properties":{"host":{"type":"string"}," ":{"type":"string"}},"additionalProperties":false}`), "unrepresentable_source_name"},
	} {
		target := SchemaTarget{Kind: "json_schema", Schema: tc.schema}
		r, e := ValidateSchema(analysis.QueryDocument{Text: "table *"}, target)
		if e != nil {
			t.Fatal(e)
		}
		o := r.Outcomes[0]
		if r.Status != analysis.Incomplete || o.MatchesComplete {
			t.Fatalf("claimed complete %+v", o)
		}
		found := false
		for _, ev := range o.Evidence {
			found = found || ev.Reason == tc.reason
		}
		if !found {
			t.Fatalf("lost %s evidence", tc.reason)
		}
		r, e = ValidateSchema(analysis.QueryDocument{Text: "table host"}, target)
		if e != nil {
			t.Fatal(e)
		}
		if r.Status != analysis.Valid || !r.Coverage.SchemaComplete || !r.Outcomes[0].MatchesComplete {
			t.Fatalf("unrelated exact obligation poisoned: %+v", r)
		}
	}
}
