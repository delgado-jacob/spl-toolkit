package resolution_test

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

func acceptanceNamed(t *testing.T, name string) resolution.Request {
	t.Helper()
	for _, c := range readAcceptanceCases(t) {
		if c.ID == name {
			return fixtureRequest(t, c)
		}
	}
	t.Fatal("unknown fixture", name)
	return resolution.Request{}
}
func TestResolveDetachmentAndConcurrentCalls(t *testing.T) {
	r := acceptanceNamed(t, "repeated-marker-scopes")
	p, err := resolution.Prepare(r.Compatibility.Snapshot, r.Compatibility.SchemaBundle)
	if err != nil {
		t.Fatal(err)
	}
	request := resolution.PreparedRequest{SchemaVersion: 1, Document: r.Document, Resolutions: r.Resolutions, Compatibility: compatibility.ResolutionAssessment{QueryScope: r.Compatibility.QueryScope, InputBindings: r.Compatibility.InputBindings}}
	report, err := p.Resolve(request)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := json.Marshal(report)
	r.Compatibility.Snapshot.Objects[0].Name = "caller mutation"
	r.Compatibility.SchemaBundle.Schemas[0].Catalog[0] = '!'
	report.Resolutions[0].Values[0] = "report mutation"
	report.Original.Placeholders[0].OriginalInputIDs[0] = "mutation"
	report.Variants[0].Selection[0].Value = "mutation"
	report.Variants[0].Proof.Roles[0].CandidateInput.Occurrences[0].ID = "mutation"
	report.Variants[0].Compatibility.InputBindings[0].Expected.Name = "mutation"
	*report.Variants[0].ResolvedQuery = "mutation"
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fresh, err := p.Resolve(request)
			if err != nil {
				t.Error(err)
				return
			}
			got, _ := json.Marshal(fresh)
			if !bytes.Equal(want, got) {
				t.Error("shared prepared state or report mutation")
			}
		}()
	}
	wg.Wait()
}
func TestResolveStrictJSONAndStableDigests(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	raw, _ := json.Marshal(r)
	report, err := resolution.ResolveJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if report.Provenance.QueryDigest == report.Variants[0].Provenance.QueryDigest {
		t.Fatal("candidate digest must identify exact changed text")
	}
	next, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(report)
	b, _ := json.Marshal(next)
	if !bytes.Equal(a, b) {
		t.Fatal("JSON and typed calls disagree")
	}
	_, err = resolution.ResolveJSON(bytes.Replace(raw, []byte(`"schema_version":1`), []byte(`"schema_version":1,"schema_version":1`), 1))
	if _, ok := resolution.RequestErrorDetails(err); !ok {
		t.Fatal("duplicate wire property accepted")
	}
}
func TestResolveResidualMarkersAreAtomicAndUnpublished(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	r.Resolutions[0].Values = []string{"$other"}
	value := "$other"
	r.Compatibility.InputBindings[0].ResolvedValue = &value
	r.Compatibility.InputBindings[0].Expected.Name = value
	selectedObject := r.Compatibility.InputBindings[0].ObjectID
	for i := range r.Compatibility.Snapshot.Objects {
		if r.Compatibility.Snapshot.Objects[i].ID == selectedObject {
			r.Compatibility.Snapshot.Objects[i].Name = value
		}
	}
	for i := range r.Compatibility.SchemaBundle.Bindings {
		if r.Compatibility.SchemaBundle.Bindings[i].ObjectID == selectedObject {
			r.Compatibility.SchemaBundle.Bindings[i].Expected.Name = value
		}
	}
	report, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	v := report.Variants[0]
	if v.ResolvedQuery != nil || v.Outcome != "incomplete" || v.CandidateText == nil || v.CandidateAnalysis.Inputs[0].Identity.Value != "$other" || len(v.Changes) != 1 || !v.Proof.Proven || v.Compatibility == nil || v.Compatibility.Outcome != "satisfied" {
		t.Fatalf("residual publication: %+v", v)
	}
}
func TestResolveAdmissionAndPartialDiscovery(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	r.Resolutions = []resolution.Resolution{}
	if report, err := resolution.Resolve(r); err == nil || report != nil {
		t.Fatal("missing eligible resolution admitted")
	}
	r = acceptanceNamed(t, "single-value")
	r.Document.Text = `FROM $events | append [FROM {kind: $kind, properties: {name: "main"}}]`
	// Select the original eligible role using canonical analysis of this document.
	a, err := analysis.Analyze(r.Document)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range a.Inputs {
		if i.Identity.Value == "$events" {
			r.Compatibility.InputBindings[0].OriginalInputID = i.ID
		}
	}
	report, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Original.Placeholders) != 1 || report.Counts.Incomplete != 1 || report.Variants[0].ResolvedQuery != nil || report.Original.Coverage.State != "partial" {
		t.Fatalf("partial discovery: counts=%+v coverage=%+v placeholders=%+v", report.Counts, report.Original.Coverage, report.Original.Placeholders)
	}
}
func TestResolveUnsupportedRenderingRetainsAssessableSibling(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	r.Document = analysis.QueryDocument{Language: "spl", Text: `search index="$events"`}
	r.Resolutions[0].Kind = "index"
	r.Resolutions[0].Values = []string{"main", "main*", "main\nbad"}
	r.Compatibility.InputBindings = []compatibility.ResolutionBinding{}
	r.Compatibility.SchemaBundle = nil
	base := r.Compatibility.Snapshot.Objects[0]
	base.ID = "index-main"
	base.Kind = "index"
	base.Name = "main"
	base.Namespace = ""
	base.App = ""
	base.Owner = ""
	r.Compatibility.Snapshot.Objects = append(r.Compatibility.Snapshot.Objects, base)
	a, err := analysis.Analyze(r.Document)
	if err != nil {
		t.Fatal(err)
	}
	for i := range r.Compatibility.Snapshot.Capabilities {
		r.Compatibility.Snapshot.Capabilities[i].ID = "language:spl:profile:splunkd"
	}
	_ = a
	report, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	if report.GeneratedCount != 3 || report.Counts.Incomplete != 1 || report.Counts.Failed != 2 || report.Variants[1].CandidateText == nil || report.Variants[1].ResolvedQuery != nil || len(report.Variants[1].Proof.Limitations) == 0 {
		t.Fatalf("unsupported sibling: %+v", report.Counts)
	}
	if report.Variants[1].ID == report.Variants[2].ID || *report.Variants[1].CandidateText != *report.Variants[2].CandidateText || report.Variants[1].Selection[0].Value == report.Variants[2].Selection[0].Value {
		t.Fatal("equal-text selections collapsed")
	}
}

func TestResolveMarkerAdmissionCodes(t *testing.T) {
	for _, tc := range []struct {
		name, code string
		change     func(*resolution.Request)
	}{
		{"missing", "resolution_missing", func(r *resolution.Request) { r.Resolutions = []resolution.Resolution{} }},
		{"unused", "resolution_unused", func(r *resolution.Request) {
			r.Resolutions = append(r.Resolutions, resolution.Resolution{Placeholder: "$unused", Kind: "dataset", Values: []string{"events"}})
		}},
		{"kind", "resolution_kind_mismatch", func(r *resolution.Request) { r.Resolutions[0].Kind = "index" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := acceptanceNamed(t, "single-value")
			tc.change(&r)
			report, err := resolution.Resolve(r)
			detail, ok := resolution.RequestErrorDetails(err)
			if report != nil || !ok || detail.Code != tc.code {
				t.Fatalf("admission: %+v %v", report, err)
			}
		})
	}
}
func TestResolvePartialUnknownChoiceKeepsEvidence(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	r.Document.Text = `FROM $events | append [FROM {kind: $kind, properties: {name: "main"}}]`
	r.Resolutions = append(r.Resolutions, resolution.Resolution{Placeholder: "$kind", Kind: "dataset", Values: []string{"index"}})
	report, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	v := report.Variants[0]
	if v.Outcome != "incomplete" || v.ResolvedQuery != nil || len(v.Selection) != 2 || len(v.Changes) != 1 || v.CandidateAnalysis == nil {
		t.Fatalf("partial unknown evidence lost: %+v", v)
	}
}
func TestResolveSemanticArtifactErrorsAreStructured(t *testing.T) {
	for _, which := range []string{"snapshot", "schema"} {
		t.Run(which, func(t *testing.T) {
			r := acceptanceNamed(t, "single-value")
			if which == "snapshot" {
				r.Compatibility.Snapshot.Objects[0].ID = ""
			} else {
				r.Compatibility.SchemaBundle.Bindings[0].SchemaID = "does-not-exist"
			}
			report, err := resolution.Resolve(r)
			detail, ok := resolution.RequestErrorDetails(err)
			if report != nil || !ok || detail.Path == "" {
				t.Fatalf("typed artifact detail lost: report=%v err=%v", report, err)
			}
			raw, _ := json.Marshal(r)
			report, err = resolution.ResolveJSON(raw)
			detail, ok = resolution.RequestErrorDetails(err)
			if report != nil || !ok || detail.Path == "" {
				t.Fatalf("wire artifact detail lost: report=%v err=%v", report, err)
			}
		})
	}
}
func TestResolveAssessmentDigestIgnoresBindingOrder(t *testing.T) {
	r := acceptanceNamed(t, "equal-name-distinct-schemas")
	first, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	r.Compatibility.InputBindings[0], r.Compatibility.InputBindings[1] = r.Compatibility.InputBindings[1], r.Compatibility.InputBindings[0]
	next, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	if first.Provenance.AssessmentInputDigest != next.Provenance.AssessmentInputDigest {
		t.Fatal("assessment digest depends on binding array order")
	}
}

func TestResolveVariantIdentityIncludesOriginalQueryIdentity(t *testing.T) {
	r := acceptanceNamed(t, "single-value")
	first, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	r.Document.SourceID = "different-original-document"
	next, err := resolution.Resolve(r)
	if err != nil {
		t.Fatal(err)
	}
	if first.Variants[0].ID == next.Variants[0].ID || first.Variants[0].Provenance.QueryDigest != next.Variants[0].Provenance.QueryDigest {
		t.Fatal("variant identity must distinguish documents while query digest identifies exact bytes")
	}
}
