package analysis

import (
	"reflect"
	"strings"
	"testing"
)

func TestPipelineJoinCorrelationProof(t *testing.T) {
	for _, options := range []string{"", "type=inner", "max=0", "max=2", "type=left", "type=outer"} {
		t.Run(options, func(t *testing.T) {
			query := `from $events | join ` + options + ` left=e right=u where e.id=u.uid AND u.region=e.region [from $users]`
			r := spl2AnalyzeTest(t, query)
			if len(r.Correlation.Nodes) != 2 || len(r.Correlation.Edges) != 1 || len(r.Correlation.Components) != 1 || r.Correlation.Outcome != "connected" || r.Correlation.Coverage.State != "complete" {
				t.Fatalf("graph: %+v diagnostics: %+v", r.Correlation, r.Diagnostics)
			}
			edge := r.Correlation.Edges[0]
			if len(edge.Keys) != 2 {
				t.Fatalf("key evidence: %+v", edge)
			}
			for i, key := range edge.Keys {
				want := []string{"e.id=u.uid", "u.region=e.region"}[i]
				if query[key.Location.Start.Offset:key.Location.End.Offset] != want || len(key.Left.ReferenceIDs) != 1 || len(key.Right.ReferenceIDs) != 1 {
					t.Fatalf("key proof: %+v", key)
				}
			}
			if r.FieldAttributionCoverage.State != "complete" {
				t.Fatalf("attribution: %+v", r.FieldAttributionCoverage)
			}
		})
	}
}
func TestPipelineJoinCorrelationOccurrences(t *testing.T) {
	for _, tc := range []struct {
		name, query, outcome     string
		nodes, edges, components int
	}{
		{"self", `from $events | join left=e right=u where e.id=u.id [from $events]`, "connected", 2, 1, 1},
		{"independent reuse", `from $events | join left=e right=u where e.id=u.id [from $events] | union [from $events]`, "disconnected", 3, 1, 2},
		{"chain", `from $events | fields id | join left=e right=u where e.id=u.uid [from $users | fields uid] | join left=p right=a where p.uid=a.aid [from $accounts | fields aid]`, "connected", 3, 2, 1},
		{"view reuse", `$pair = from $events | join left=e right=u where e.id=u.uid [from $users]; $out = from $pair | union [from $pair];`, "disconnected", 4, 2, 2},
		{"nested views", `$left = from $events; $right = from $users; $pair = from $left | join left=e right=u where e.id=u.uid [from $right]; $out = from $pair | union [from $pair];`, "disconnected", 4, 2, 2},
		{"union", `from $events | union [from $users]`, "disconnected", 2, 0, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Correlation.Outcome != tc.outcome || len(r.Correlation.Nodes) != tc.nodes || len(r.Correlation.Edges) != tc.edges || len(r.Correlation.Components) != tc.components {
				t.Fatalf("graph %+v diagnostics %+v", r.Correlation, r.Diagnostics)
			}
			if !reflect.DeepEqual(r.Correlation, r.Requirements.Correlation) {
				t.Fatal("projections differ")
			}
		})
	}
}
func TestPipelineJoinCorrelationUnprovedOccurrence(t *testing.T) {
	for _, query := range []string{
		`from $events | union [from $events] | join left=e right=u where e.id=u.uid [from $users]`,
		`from $events | join left=e right=u where e.actor.id=u.id [from $users]`,
		`from $events | join left=e right=u max=-1 where e.id=u.uid [from $users]`,
		`from $events | join left=e right=u max=1 max=2 where e.id=u.uid [from $users]`,
		`from $events | join left=e right=u type=left type=inner where e.id=u.uid [from $users]`,
		`from $events | join left=e left=x right=u where e.id=u.uid [from $users]`,
		`from $events | join left=e right=e where e.id=e.uid [from $users]`,
		`from $events | join left=e right=u max=1.5 where e.id=u.uid [from $users]`,
		`from $events | join left=e right=u max=$limit where e.id=u.uid [from $users]`,
		`from $events | join left=e right=u where e.id=u.uid OR e.region=u.region [from $users]`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Correlation.Outcome != "indeterminate" || len(r.Correlation.Edges) != 0 {
			t.Fatalf("unproved graph for %q: %+v", query, r.Correlation)
		}
	}
}
func TestPipelineJoinCorrelationDerivedKey(t *testing.T) {
	query := `from $events | eval key=id+region | join left=e right=u where e.key=u.uid [from $users]`
	r := spl2AnalyzeTest(t, query)
	if r.Correlation.Outcome != "connected" || len(r.Correlation.Edges) != 1 {
		t.Fatalf("derived supplier graph: %+v diagnostics %+v", r.Correlation, r.Diagnostics)
	}
	for _, item := range r.Requirements.Items {
		if item.Kind == "field" && strings.Contains(item.Identity, "key") {
			t.Fatalf("derived key created destination requirement: %+v", item)
		}
	}
}

func TestPipelineJoinHeldOutputOwnership(t *testing.T) {
	query := `from $events | join left=e right=u where e.actor.id=u.id [from $users] | fields id`
	r := spl2AnalyzeTest(t, query)
	if !r.Coverage.SyntaxComplete {
		t.Fatalf("intact deeper predicate lost syntax truth: %+v", r.Coverage)
	}
	for _, item := range r.Requirements.Items {
		if item.Kind == "field" && len(item.Occurrences) > 0 && item.Occurrences[len(item.Occurrences)-1].OriginalName == "id" {
			if item.InputID != "" || item.Ownership.State != "unproved" || len(item.Ownership.CandidateInputIDs) != 2 || r.FieldAttributionCoverage.State != "partial" {
				t.Fatalf("held join fabricated output owner: %+v coverage %+v", item, r.FieldAttributionCoverage)
			}
			return
		}
	}
	t.Fatal("held join dropped downstream field evidence")
}

func TestPipelineJoinLeftOuterRequirementParity(t *testing.T) {
	query := `from $events | fields id,region | join type=TYPE left=e right=u where e.id=u.uid [from $users | fields uid,team] | where team="security"`
	left := spl2AnalyzeTest(t, strings.Replace(query, "TYPE", "left ", 1))
	outer := spl2AnalyzeTest(t, strings.Replace(query, "TYPE", "outer", 1))
	// Padding keeps every evidence range equal; only the query digest differs.
	outer.Requirements.Query = left.Requirements.Query
	if !reflect.DeepEqual(left.Requirements, outer.Requirements) || !reflect.DeepEqual(left.Lineage, outer.Lineage) {
		t.Fatalf("left/outer parity: left %+v outer %+v", left.Requirements, outer.Requirements)
	}
}
func TestPipelineJoinCorrelationNestedRight(t *testing.T) {
	r := spl2AnalyzeTest(t, `from $events | fields id | join left=e right=p where e.id=p.uid [from $users | fields uid | join left=u right=a where u.uid=a.aid [from $accounts | fields aid]]`)
	if r.Correlation.Outcome != "connected" || len(r.Correlation.Edges) != 2 || len(r.Correlation.Nodes) != 3 {
		t.Fatalf("nested right correlation: %+v", r.Correlation)
	}
	for _, edge := range r.Correlation.Edges {
		if edge.Left.OccurrenceID == edge.Right.OccurrenceID {
			t.Fatal("fabricated self loop")
		}
	}
}
func TestPipelineJoinCorrelationRetainsKnownEdgesWhenIncomplete(t *testing.T) {
	query := `from $events | fields id | join left=e right=u where e.id=u.uid [from $users | fields uid] | join left=p right=a where p.actor.id=a.aid [from $accounts]`
	r := spl2AnalyzeTest(t, query)
	if !r.Coverage.SyntaxComplete || r.Correlation.Outcome != "indeterminate" || len(r.Correlation.Edges) != 1 || len(r.Correlation.Nodes) != 3 {
		t.Fatalf("partial graph lost known proof: %+v", r.Correlation)
	}
}
func TestPipelineJoinCorrelationSchemaCannotSelectOccurrence(t *testing.T) {
	query := `from $events | union [from $events] | join left=e right=u where e.id=u.uid [from $users]`
	r, err := AnalyzeWithSourceUniverse(QueryDocument{Text: query, Language: "spl2"}, SourceUniverse{Fields: []string{"id", "uid"}, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.Result.Correlation.Outcome != "indeterminate" || len(r.Result.Correlation.Edges) != 0 || r.Result.FieldAttributionCoverage.State != "complete" {
		t.Fatalf("schema selected occurrence or demoted proved logical ownership: %+v", r.Result)
	}
}
func TestPipelineJoinCorrelationCloneDetached(t *testing.T) {
	r := spl2AnalyzeTest(t, `from $events | join left=e right=u where e.id=u.uid [from $users]`)
	cloned := cloneRequirementSet(r.Requirements)
	cloned.Correlation.Edges[0].Keys[0].Left.ReferenceIDs[0] = "changed"
	cloned.Correlation.Edges[0].Keys[0].Left.FieldIdentity.Segments[0] = "changed"
	cloned.Correlation.Edges[0].Left.ReferenceIDs[0] = "changed"
	if r.Correlation.Edges[0].Keys[0].Left.ReferenceIDs[0] == "changed" || r.Requirements.Correlation.Edges[0].Keys[0].Left.FieldIdentity.Segments[0] == "changed" {
		t.Fatal("correlation key evidence aliases clone")
	}
}

func TestPipelineJoinCorrelationConstantKeys(t *testing.T) {
	for _, tc := range []struct {
		name, query, predicate, attribution string
		nodes, edges                        int
	}{
		{"left constant", `from $events | eval k=1 | join left=e right=u where e.k=u.uid [from $users]`, "e.k=u.uid", "complete", 2, 0},
		{"right constant", `from $events | join left=e right=u where e.id=u.k [from $users | eval k=1]`, "e.id=u.k", "complete", 2, 0},
		{"both constants", `from $events | where id>0 | eval k=1 | join left=e right=u where e.k=u.k [from $users | where uid>0 | eval k=1]`, "e.k=u.k", "complete", 2, 0},
		{"earlier edge", `from $events | fields id | join left=e right=u where e.id=u.uid [from $users | fields uid] | eval k=1 | join left=p right=a where p.k=a.aid [from $accounts]`, "p.k=a.aid", "complete", 3, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			graph := r.Correlation
			if graph.Outcome != "indeterminate" || graph.Coverage.State != "partial" || len(graph.Nodes) != tc.nodes || len(graph.Edges) != tc.edges || r.FieldAttributionCoverage.State != tc.attribution {
				t.Fatalf("constant-key proof: graph %+v attribution %+v", graph, r.FieldAttributionCoverage)
			}
			found := false
			for _, reason := range graph.Coverage.Reasons {
				if reason.Code != "correlation_incomplete" {
					continue
				}
				if tc.query[reason.Location.Start.Offset:reason.Location.End.Offset] != tc.predicate || len(reason.ReferenceIDs) != 2 || reason.StageID == "" || reason.ScopeID == "" {
					t.Fatalf("unlocated predicate uncertainty: %+v", reason)
				}
				for _, id := range reason.ReferenceIDs {
					ref := r.References[referenceOrdinal(t, id)]
					if ref.Kind != "field" || ref.Role != "read" || !strings.Contains(tc.predicate, ref.OriginalName) || tc.query[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
						t.Fatalf("lost key evidence: %+v", ref)
					}
				}
				found = true
			}
			if !found {
				t.Fatal("missing supplying-occurrence uncertainty")
			}
			for _, item := range r.Requirements.Items {
				if item.Kind == "field" && item.Identity == "k" {
					t.Fatalf("constant key became an external requirement: %+v", item)
				}
			}
		})
	}
}
func TestPipelineJoinCorrelationGeneratedRowsStayNotApplicable(t *testing.T) {
	for _, query := range []string{
		`from [{k:1}] | join left=e right=u where e.k=u.uid [from [{uid:1}]]`,
		`$unused = from $events; $out = from [{k:1}] | join left=e right=u where e.k=u.uid [from [{uid:1}]];`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Correlation.Outcome != "not applicable" || r.Correlation.Coverage.State != "not_applicable" || len(r.Correlation.Coverage.Reasons) != 0 || len(r.Correlation.Edges) != 0 {
			t.Fatalf("generated-only join gained source uncertainty: %+v", r.Correlation)
		}
	}
}
