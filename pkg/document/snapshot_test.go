package document

import (
	"errors"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestSnapshotDetachedMutation(t *testing.T) {
	result, err := analysis.Analyze(analysis.QueryDocument{Text: "search host=web | eval output=host"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := New(result, testRevisionContext("target-a"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(result, testRevisionContext("target-a"))
	if err != nil {
		t.Fatal(err)
	}

	first.References[0].OriginalName = "changed"
	if result.References[0].OriginalName == "changed" || second.References[0].OriginalName == "changed" {
		t.Fatal("snapshot mutation leaked into canonical or another snapshot")
	}
	returned, err := second.ReferencesIn(0, len(second.Document.Text))
	if err != nil || len(returned) == 0 {
		t.Fatalf("returned references = %+v, %v", returned, err)
	}
	returned[0].OriginalName = "returned-slice-change"
	if second.References[0].OriginalName == "returned-slice-change" || result.References[0].OriginalName == "returned-slice-change" {
		t.Fatal("returned reference slice mutated snapshot or canonical analysis")
	}

	first.References = []analysis.Reference{{ID: "caller-added", OriginalName: "caller", Location: analysis.Location{Start: analysis.Position{Offset: 0}, End: analysis.Position{Offset: 1}}}}
	got, found := first.Reference("caller-added")
	if !found || got.OriginalName != "caller" {
		t.Fatalf("lookup retained a stale index: %+v, %v", got, found)
	}
	got.OriginalName = "returned-copy"
	again, found := first.Reference("caller-added")
	if !found || again.OriginalName != "caller" {
		t.Fatalf("lookup returned mutable internal evidence: %+v, %v", again, found)
	}
	if _, found := first.Reference("missing"); found {
		t.Fatal("unknown reference was found")
	}
}

func TestSnapshotRequirementSetDetached(t *testing.T) {
	document := analysis.QueryDocument{Text: "search host=x | table host* | `expand_me`", SourceID: "requirements"}
	result, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	want, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := New(result, RevisionContext{ToolVersion: "test", ContractVersion: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(snapshot.Requirements, result.Requirements) {
		t.Fatalf("snapshot requirements differ: got %+v want %+v", snapshot.Requirements, result.Requirements)
	}
	if len(result.Requirements.Items) == 0 || len(result.Requirements.Items[0].Occurrences) == 0 || len(result.Requirements.Gaps) == 0 || len(result.Requirements.Gaps[0].DiagnosticCodes) == 0 || len(result.Requirements.Diagnostics) == 0 {
		t.Fatalf("detachment fixture lacks nested requirement evidence: %+v", result.Requirements)
	}

	mutateSnapshotRequirements(&result.Requirements)
	if !reflect.DeepEqual(snapshot.Requirements, want.Requirements) {
		t.Fatal("analysis requirement mutation leaked into snapshot")
	}

	fresh, err := analysis.Analyze(document)
	if err != nil {
		t.Fatal(err)
	}
	freshSnapshot, err := New(fresh, RevisionContext{ToolVersion: "test", ContractVersion: "v1"})
	if err != nil {
		t.Fatal(err)
	}
	mutateSnapshotRequirements(&freshSnapshot.Requirements)
	if !reflect.DeepEqual(fresh.Requirements, want.Requirements) {
		t.Fatal("snapshot requirement mutation leaked into analysis")
	}
}

func mutateSnapshotRequirements(set *analysis.RequirementSet) {
	set.Coverage.Reasons[0] = "mutated"
	set.Items[0].Identity = "mutated"
	set.Items[0].Occurrences[0].OriginalName = "mutated"
	set.Gaps[0].Code = "mutated"
	set.Gaps[0].ReferenceIDs = append(set.Gaps[0].ReferenceIDs, "mutated")
	set.Gaps[0].DiagnosticCodes[0] = "mutated"
	set.Diagnostics[0].Code = "mutated"
}

func TestOverlappingReferenceLookup(t *testing.T) {
	view, err := New(&analysis.Result{
		Document: analysis.QueryDocument{Text: "abc"},
		References: []analysis.Reference{
			{ID: "first", Role: "read", Location: location(0, 2)},
			{ID: "second", Role: "write", Location: location(1, 2)},
			{ID: "adjacent", Role: "read", Location: location(2, 3)},
			{ID: "eof", Role: "recovery", Location: location(3, 3)},
		},
	}, testRevisionContext("target-a"))
	if err != nil {
		t.Fatal(err)
	}

	matches, err := view.ReferencesIn(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got := referenceIDs(matches); !reflect.DeepEqual(got, []string{"first", "second"}) {
		t.Fatalf("overlaps or canonical order lost: %v", got)
	}
	atBoundary, err := view.ReferencesAt(2)
	if err != nil {
		t.Fatal(err)
	}
	if got := referenceIDs(atBoundary); !reflect.DeepEqual(got, []string{"adjacent"}) {
		t.Fatalf("half-open boundary selected a preceding reference: %v", got)
	}
	atEOF, err := view.ReferencesAt(3)
	if err != nil {
		t.Fatal(err)
	}
	if got := referenceIDs(atEOF); !reflect.DeepEqual(got, []string{"eof"}) {
		t.Fatalf("EOF attached to preceding reference or missed exact zero-width reference: %v", got)
	}
	between, err := view.ReferencesIn(3, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(between) != 0 {
		t.Fatalf("zero-width range matched references: %+v", between)
	}
}

func TestSourceRangeBoundaries(t *testing.T) {
	view, err := New(&analysis.Result{Document: analysis.QueryDocument{Text: "a😀b"}}, testRevisionContext("target-a"))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := view.Source(1, 5); err != nil || got != "😀" {
		t.Fatalf("source range = %q, %v", got, err)
	}
	for _, r := range [][2]int{{-1, 0}, {2, 5}, {5, 2}, {0, 7}} {
		if _, err := view.Source(r[0], r[1]); !errors.Is(err, ErrInvalidRange) {
			t.Fatalf("range %+v error = %v, want ErrInvalidRange", r, err)
		}
	}
	if got, err := view.Source(len(view.Document.Text), len(view.Document.Text)); err != nil || got != "" {
		t.Fatalf("EOF range = %q, %v", got, err)
	}
}

func TestSQLPhasePreservation(t *testing.T) {
	result, err := analysis.Analyze(analysis.QueryDocument{Text: "SELECT host FROM main WHERE host=web", Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	view, err := New(result, testRevisionContext("target-a"))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(view.Lineage, result.Lineage) {
		t.Fatalf("lineage facts changed: got %#v want %#v", view.Lineage, result.Lineage)
	}
	if len(view.Lineage) == 0 || view.Lineage[0].ExecutionOrder == nil {
		t.Fatalf("SQL phase/evaluation order absent: %#v", view.Lineage)
	}
	originalOrder := *result.Lineage[0].ExecutionOrder
	*view.Lineage[0].ExecutionOrder = originalOrder + 100
	if *result.Lineage[0].ExecutionOrder != originalOrder {
		t.Fatal("execution order pointer leaked into canonical analysis")
	}
	for i := range view.Lineage {
		for j := range view.Lineage[i].After.Fields {
			if len(view.Lineage[i].After.Fields[j].OriginReferenceIDs) == 0 {
				continue
			}
			original := result.Lineage[i].After.Fields[j].OriginReferenceIDs[0]
			view.Lineage[i].After.Fields[j].OriginReferenceIDs[0] = "changed"
			if result.Lineage[i].After.Fields[j].OriginReferenceIDs[0] != original {
				t.Fatal("nested origin links leaked into canonical analysis")
			}
			return
		}
	}
	t.Fatal("SQL fixture did not provide origin evidence")
}

func TestSnapshotRevisionContextSeparatesChangedTarget(t *testing.T) {
	result, err := analysis.Analyze(analysis.QueryDocument{Text: "search host=web"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := New(result, testRevisionContext("field-list:host"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := New(result, testRevisionContext("field-list:host,web"))
	if err != nil {
		t.Fatal(err)
	}
	if first.SourceHash != second.SourceHash || first.Revision.TargetDigest == second.Revision.TargetDigest {
		t.Fatalf("same source did not retain distinct target context: %+v %+v", first, second)
	}
}

func location(start, end int) analysis.Location {
	return analysis.Location{Start: analysis.Position{Offset: start}, End: analysis.Position{Offset: end}}
}

func referenceIDs(references []analysis.Reference) []string {
	ids := make([]string, len(references))
	for i, reference := range references {
		ids[i] = reference.ID
	}
	return ids
}

func testRevisionContext(targetDigest string) RevisionContext {
	return RevisionContext{ToolVersion: "0.1.1", ContractVersion: "analysis-report-v1", TargetDigest: targetDigest}
}
