package workflow

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWorkflowAlignment(t *testing.T) {
	q := comparisonSeed(t)
	q, _ = exportCopy(q)
	q.Before.Entries[0].Analysis.Requirements.CapabilityRevision = "sha256:" + strings.Repeat("a", 64)
	q.Before.Entries[0].Compatibility.Requirements.CapabilityRevision = "sha256:" + strings.Repeat("a", 64)
	q.Before.Entries[0].Compatibility.Provenance.CapabilityRevision = "sha256:" + strings.Repeat("a", 64)
	r, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Entries) != 1 || len(r.Entries[0].Pairs) == 0 {
		t.Fatalf("no original evidence correspondence: %+v", r)
	}
	q.After.Entries[0].Analysis.Document.Text += " "
	q.After.Entries[0].SourceHash = "bad"
	if _, err := Compare(q); err == nil {
		t.Fatal("accepted contradictory bytes")
	}
}

func TestWorkflowAlignmentUniqueAndAmbiguous(t *testing.T) {
	e := ComparisonEntry{Pairs: []EvidencePair{}, Unmatched: []string{}, Ambiguous: []string{}}
	b := []alignmentFact{{key: "same", pointer: "/before/0", id: "old1"}, {key: "same", pointer: "/before/1", id: "old2"}}
	a := []alignmentFact{{key: "same", pointer: "/after/0", id: "new1"}}
	pairFacts(&e, b, a, "test")
	if len(e.Pairs) != 0 || len(e.Ambiguous) != 3 {
		t.Fatalf("guessed many-to-one: %+v", e)
	}
}
func TestWorkflowAlignmentVariantOrder(t *testing.T) {
	req := seedResolutionRequest(t)
	r, err := Assess(req)
	if err != nil {
		t.Fatal(err)
	}
	q := CompareRequest{SchemaVersion: 1, Before: *r, After: *r}
	q, _ = exportCopy(q)
	q.After.Entries[0].Resolution.Variants[0], q.After.Entries[0].Resolution.Variants[1] = q.After.Entries[0].Resolution.Variants[1], q.After.Entries[0].Resolution.Variants[0]
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, p := range got.Entries[0].Pairs {
		if p.Basis == "canonical_resolution_selection" {
			n++
			workflowPointer(t, got, p.BeforePointer)
			workflowPointer(t, got, p.AfterPointer)
		}
	}
	if n != 2 {
		t.Fatalf("variant order changed correspondence: %+v", got.Entries[0])
	}
}

func TestWorkflowAlignmentChangedIDsAndOwnership(t *testing.T) {
	q := comparisonSeed(t)
	raw, _ := json.Marshal(q.Before)
	// This is supplied synthetic historical evidence, not output claimed from an old binary.
	raw = bytes.ReplaceAll(raw, []byte(`"ref-0"`), []byte(`"historical-ref"`))
	json.Unmarshal(raw, &q.Before)
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range got.Entries[0].Pairs {
		workflowPointer(t, got, p.BeforePointer)
		workflowPointer(t, got, p.AfterPointer)
	}
	if len(got.Entries[0].Pairs) < 5 {
		t.Fatal("unique original evidence failed to align across renamed IDs")
	}
	e := ComparisonEntry{Pairs: []EvidencePair{}, Unmatched: []string{}, Ambiguous: []string{}}
	b, _ := exportCopy(*q.Before.Entries[0].Analysis)
	a, _ := exportCopy(*q.After.Entries[0].Analysis)
	b.Requirements.Items[1].Ownership.State = "unproved"
	alignAnalysis(&e, &b, &a, "/before", "/after")
	for _, p := range e.Pairs {
		if p.BeforePointer == "/before/requirements/items/1" {
			t.Fatal("same name bypassed changed ownership")
		}
	}
}
func TestWorkflowAlignmentQueryMismatchAndEntryOrder(t *testing.T) {
	q := comparisonSeed(t)
	r, err := Assess(seedRequest(t, "from main | where host=\"b\""))
	if err != nil {
		t.Fatal(err)
	}
	q.After = *r
	got, err := Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries[0].Pairs) != 0 || got.Entries[0].Reasons[0] != "query_input_mismatch" {
		t.Fatalf("changed bytes were compared: %+v", got)
	}
	q = comparisonSeed(t)
	q, _ = exportCopy(q)
	second := q.Before.Entries[0]
	second.ID = "second"
	q.Before.Entries = append(q.Before.Entries, second)
	q.Before.Counts = Counts{Selected: 2}
	q.Before.Status = "valid"
	finalize(&q.Before)
	q.After, _ = exportCopy(q.Before)
	q.After.Entries[0], q.After.Entries[1] = q.After.Entries[1], q.After.Entries[0]
	got, err = Compare(q)
	if err != nil {
		t.Fatal(err)
	}
	if got.Entries[0].ID != "second" || got.Entries[1].ID != "d1" {
		t.Fatal("comparison lost after report order")
	}
}
