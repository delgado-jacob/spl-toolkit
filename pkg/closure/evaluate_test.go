package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func evalRequest(text string, collections []Collection, objects ...Definition) Request {
	return Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: text, Language: "spl", SourceID: "root.spl"}, Bundle: DefinitionBundle{SchemaVersion: 1, ScopeID: "test", Collections: collections, Objects: objects}, Bindings: []Binding{}}
}
func evalDef(id, kind, name, body string, relations ...Relation) Definition {
	d := Definition{ID: id, Kind: kind, Name: name, SourceID: id + ".conf", Relations: relations}
	if body != "" {
		d.Document = &analysis.QueryDocument{Text: body, Language: "spl", SourceID: id + ".conf"}
	}
	return d
}
func evalDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}
func edgeWith(report *Report, kind, name, target string) bool {
	for _, e := range report.Traversal {
		if e.Kind == kind && e.Name == name && e.ToObjectID == target {
			return true
		}
	}
	return false
}

func TestEvaluateMacroEffectiveFlowAndLookup(t *testing.T) {
	macro := macroDef("m", "enrich", "lookup users user OUTPUT role")
	lookup := evalDef("l", "lookup", "users", "")
	req := evalRequest("`enrich` | eval access=role", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, lookup)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Valid || !report.Coverage.Complete || report.EffectiveAnalysis.Document.Text != "lookup users user OUTPUT role | eval access=role" {
		t.Fatalf("closure: status=%s coverage=%+v effective=%q gaps=%+v", report.Status, report.Coverage, report.EffectiveAnalysis.Document.Text, report.Gaps)
	}
	if report.DirectAnalysis.Requirements.Coverage.Complete {
		t.Fatal("direct macro warning unexpectedly disappeared")
	}
	if !edgeWith(report, "macro", "enrich", "m") || !edgeWith(report, "lookup", "users", "l") {
		t.Fatalf("missing traversal edges: %+v", report.Traversal)
	}
	if len(report.Provenance) == 0 || report.EffectiveAnalysis.Status != analysis.Valid {
		t.Fatalf("effective analysis/provenance: %+v %+v", report.EffectiveAnalysis, report.Provenance)
	}
}
func TestEvaluateDiagnosticPreservesCrossSegmentOrigins(t *testing.T) {
	query := "search index=main | `x` z"
	macro := macroDef("x", "x", "lookup users")
	report, err := Evaluate(evalRequest(query, []Collection{{Kind: "macro", Coverage: "complete"}}, macro))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range report.Diagnostics {
		if d.Diagnostic.Code != "SPL_UNSUPPORTED_SEMANTICS" {
			continue
		}
		if len(d.Origins) != 2 {
			t.Fatalf("cross-segment diagnostic lost original contributors: %+v", d)
		}
		if got := d.Origins[0]; got.Kind != "definition" || got.ObjectID != "x" || got.SourceID != "x.conf" || got.Start != 7 || got.End != 12 {
			t.Fatalf("macro contributor: %+v", got)
		}
		if got := d.Origins[1]; got.Kind != "query" || got.SourceID != "root.spl" || got.Start != strings.Index(query, " z") || got.End != len(query) {
			t.Fatalf("query contributor: %+v", got)
		}
		if d.Source != d.Origins[0] {
			t.Fatalf("convenience source must be first contributor: %+v", d)
		}
		return
	}
	t.Fatalf("expected crossing diagnostic: %+v", report.Diagnostics)
}

func TestEvaluateDiagnosticRetainsPlaceholderOrigin(t *testing.T) {
	query := "search index=main | `m(users)` z"
	body := "lookup $name$"
	macro := macroDef("m", "m", body, "name")
	report, err := Evaluate(evalRequest(query, []Collection{{Kind: "macro", Coverage: "complete"}}, macro))
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range report.Diagnostics {
		if d.Diagnostic.Code != "SPL_UNSUPPORTED_SEMANTICS" {
			continue
		}
		argument := SourceInterval{Kind: "query", SourceID: "root.spl", Start: strings.Index(query, "users"), End: strings.Index(query, "users") + len("users")}
		placeholder := SourceInterval{Kind: "definition", SourceID: "m.conf", ObjectID: "m", Start: strings.Index(body, "$name$"), End: len(body)}
		trailing := SourceInterval{Kind: "query", SourceID: "root.spl", Start: strings.Index(query, " z"), End: len(query)}
		if len(d.Origins) != 3 || d.Origins[0] != argument || d.Origins[1] != placeholder || d.Origins[2] != trailing {
			t.Fatalf("substitution and trailing contributors: %+v", d)
		}
		return
	}
	t.Fatalf("expected substituted crossing diagnostic: %+v", report.Diagnostics)
}

func TestEvaluateSavedSearchAndEventTypeTransitive(t *testing.T) {
	leaf := evalDef("l", "lookup", "users", "")
	child := evalDef("child", "event_type", "Child", "lookup users user OUTPUT role")
	daily := evalDef("daily", "saved_search", "Daily", "search index=main", Relation{Kind: "event_type", Name: "Child", Property: stringPointer("/eventtype")})
	req := evalRequest("| from savedsearch:Daily", []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "event_type", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, daily, child, leaf)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Incomplete || report.Coverage.EffectiveQuery {
		t.Fatalf("from field effects must remain incomplete: status=%s coverage=%+v gaps=%+v", report.Status, report.Coverage, report.Gaps)
	}
	if !edgeWith(report, "saved_search", "Daily", "daily") || !edgeWith(report, "event_type", "Child", "child") || !edgeWith(report, "lookup", "users", "l") {
		t.Fatalf("transitive traversal: %+v", report.Traversal)
	}
	if len(report.DefinitionAnalyses) != 2 {
		t.Fatalf("reachable definitions only: %+v", report.DefinitionAnalyses)
	}
	for _, e := range report.Traversal {
		if e.Kind == "dataset" && e.Name == "savedsearch:Daily" {
			t.Fatalf("legacy dataset projection traversed: %+v", e)
		}
	}
}
func stringPointer(s string) *string { return &s }
func TestEvaluateCycleAndKnownEdges(t *testing.T) {
	a := evalDef("a", "saved_search", "A", "search index=main", Relation{Kind: "saved_search", Name: "B", Property: stringPointer("/next")})
	b := evalDef("b", "saved_search", "B", "search index=main", Relation{Kind: "saved_search", Name: "A", Property: stringPointer("/next")})
	report, err := Evaluate(evalRequest("| from savedsearch:A", []Collection{{Kind: "saved_search", Coverage: "complete"}}, a, b))
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Incomplete || !edgeWith(report, "saved_search", "A", "a") || !edgeWith(report, "saved_search", "B", "b") {
		t.Fatalf("cycle lost edges: status=%s traversal=%+v", report.Status, report.Traversal)
	}
	found := false
	for _, e := range report.Traversal {
		if len(e.CyclePath) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("cycle path absent: %+v", report.Traversal)
	}
}
func TestEvaluateCoverageKnownAbsentAndUnknown(t *testing.T) {
	query := "| lookup missing user OUTPUT role"
	for _, tc := range []struct{ name, coverage, want string }{{"complete", "complete", "missing"}, {"partial", "partial", "unknown"}, {"unavailable", "unavailable", "unknown"}} {
		t.Run(tc.name, func(t *testing.T) {
			r, e := Evaluate(evalRequest(query, []Collection{{Kind: "lookup", Coverage: tc.coverage}}))
			if e != nil {
				t.Fatal(e)
			}
			if r.Status != analysis.Incomplete || len(r.Gaps) == 0 || r.Traversal[0].Resolution != tc.want {
				t.Fatalf("status=%s gaps=%+v edges=%+v", r.Status, r.Gaps, r.Traversal)
			}
		})
	}
	l := evalDef("l", "lookup", "missing", "")
	r, e := Evaluate(evalRequest(query, []Collection{{Kind: "lookup", Coverage: "partial"}, {Kind: "macro", Coverage: "unavailable"}}, l))
	if e != nil {
		t.Fatal(e)
	}
	if r.Status != analysis.Incomplete || r.Coverage.Collections || !edgeWith(r, "lookup", "missing", "l") {
		t.Fatalf("supplied lookup in partial collection: status=%s coverage=%+v edges=%+v gaps=%+v", r.Status, r.Coverage, r.Traversal, r.Gaps)
	}
}
func TestEvaluateBindingValidationAndAmbiguity(t *testing.T) {
	a := evalDef("a", "lookup", "users", "")
	b := evalDef("b", "lookup", "users", "")
	query := "| lookup users user OUTPUT role"
	req := evalRequest(query, []Collection{{Kind: "lookup", Coverage: "complete"}}, a, b)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Incomplete || report.Traversal[0].Resolution != "ambiguous" {
		t.Fatalf("ambiguity: %+v", report.Traversal)
	}
	start := strings.Index(query, "users")
	req.Bindings = []Binding{{DocumentDigest: evalDigest(query), Kind: "lookup", Start: start, End: start + len("users"), ObjectID: "b"}}
	report, err = Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Valid || !edgeWith(report, "lookup", "users", "b") {
		t.Fatalf("bound: status=%s edges=%+v gaps=%+v", report.Status, report.Traversal, report.Gaps)
	}
	req.Bindings[0].Start++
	if report, err = Evaluate(req); !IsInputError(err) || report != nil {
		t.Fatalf("stale binding: report=%+v err=%v", report, err)
	}
	nested := evalRequest("`outer(`inner`)`", []Collection{{Kind: "macro", Coverage: "complete"}}, macroDef("outer", "outer", "eval x=$a$", "a"), macroDef("inner", "inner", "1"))
	site := strings.Index(nested.Document.Text, "`inner`")
	nested.Bindings = []Binding{{DocumentDigest: evalDigest(nested.Document.Text), Kind: "macro", Start: site, End: site + len("`inner`"), ObjectID: "inner"}}
	if report, err = Evaluate(nested); err != nil || report == nil {
		t.Fatalf("nested binding: %+v %v", report, err)
	}
}
func TestEvaluateInvalidEffectiveAndDynamicExpansion(t *testing.T) {
	bad := evalRequest("`bad`", []Collection{{Kind: "macro", Coverage: "complete"}}, macroDef("bad", "bad", "eval ="))
	report, err := Evaluate(bad)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Invalid {
		t.Fatalf("definite invalid content: %+v", report)
	}
	held := macroDef("held", "held", "eval x=1")
	v := true
	held.EvalBased = &v
	report, err = Evaluate(evalRequest("`held`", []Collection{{Kind: "macro", Coverage: "complete"}}, held))
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Incomplete || report.Coverage.Expansion || len(report.Gaps) == 0 {
		t.Fatalf("dynamic macro: status=%s coverage=%+v gaps=%+v", report.Status, report.Coverage, report.Gaps)
	}
}

func TestEvaluateDefinitionBindingAndExplicitRelationGap(t *testing.T) {
	a := evalDef("a", "lookup", "users", "")
	b := evalDef("b", "lookup", "users", "")
	body := "lookup users user OUTPUT role"
	saved := evalDef("s", "saved_search", "Daily", body, Relation{Kind: "lookup", Name: "absent", Property: stringPointer("/lookup")})
	req := evalRequest("| from savedsearch:Daily", []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, a, b, saved)
	start := strings.Index(body, "users")
	req.Bindings = []Binding{{DocumentDigest: evalDigest(body), Kind: "lookup", Start: start, End: start + len("users"), ObjectID: "b"}}
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if !edgeWith(report, "lookup", "users", "b") {
		t.Fatalf("definition binding: %+v", report.Traversal)
	}
	found := false
	for _, edge := range report.Traversal {
		if edge.Kind == "lookup" && edge.Name == "absent" && edge.Property == "/lookup" && edge.Resolution == "missing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("explicit absent relation not retained: %+v", report.Traversal)
	}
	req.Bindings[0].DocumentDigest = evalDigest("wrong")
	if got, err := Evaluate(req); !IsInputError(err) || got != nil {
		t.Fatalf("stale definition binding: %+v %v", got, err)
	}
}
func TestEvaluateRepeatedMacroInstancesAndCyclePath(t *testing.T) {
	m := macroDef("m", "m", "lookup $name$ user OUTPUT role", "name")
	a := evalDef("a", "lookup", "one", "")
	b := evalDef("b", "lookup", "two", "")
	report, err := Evaluate(evalRequest("`m(one)` | `m(two)`", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, m, a, b))
	if err != nil {
		t.Fatal(err)
	}
	if !edgeWith(report, "lookup", "one", "a") || !edgeWith(report, "lookup", "two", "b") {
		t.Fatalf("argument contexts: %+v", report.Traversal)
	}
	var macroEdges int
	for _, edge := range report.Traversal {
		if edge.Kind == "macro" && edge.Name == "m" {
			macroEdges++
		}
	}
	if macroEdges != 2 {
		t.Fatalf("macro occurrences=%d: %+v", macroEdges, report.Traversal)
	}
	loop := macroDef("loop", "loop", "`loop`")
	cycle, err := Evaluate(evalRequest("`loop`", []Collection{{Kind: "macro", Coverage: "complete"}}, loop))
	if err != nil {
		t.Fatal(err)
	}
	if cycle.Status != analysis.Incomplete {
		t.Fatalf("macro cycle became %s", cycle.Status)
	}
	found := false
	for _, edge := range cycle.Traversal {
		if len(edge.CyclePath) == 2 && edge.CyclePath[0] == "loop" && edge.CyclePath[1] == "loop" {
			found = true
		}
	}
	if !found {
		t.Fatalf("macro cycle path absent: %+v", cycle.Traversal)
	}
}
func TestEvaluateMissingQueryBodyAndEmptyArrays(t *testing.T) {
	missing := evalDef("s", "saved_search", "Daily", "")
	r, err := Evaluate(evalRequest("| from savedsearch:Daily", []Collection{{Kind: "saved_search", Coverage: "complete"}}, missing))
	if err != nil {
		t.Fatal(err)
	}
	if r.Coverage.TraversedDefinitions || r.Status != analysis.Incomplete {
		t.Fatalf("missing body: %+v %+v", r.Coverage, r.Gaps)
	}
	clean, err := Evaluate(evalRequest("search index=main", nil))
	if err != nil {
		t.Fatal(err)
	}
	if clean.DefinitionAnalyses == nil || clean.Provenance == nil || clean.Gaps == nil || clean.Diagnostics == nil || clean.Traversal == nil || clean.Coverage.Reasons == nil {
		t.Fatalf("nil report arrays: %+v", clean)
	}
}

func TestEvaluateDefinitionMacroBindingAcrossArgumentSubstitution(t *testing.T) {
	a := macroDef("a", "same", "eval v=$x$", "x")
	b := macroDef("b", "same", "eval w=$x$", "x")
	wrapper := macroDef("wrapper", "wrapper", "`same($x$)`", "x")
	req := evalRequest("`wrapper(9)`", []Collection{{Kind: "macro", Coverage: "complete"}}, a, b, wrapper)
	req.Bindings = []Binding{{DocumentDigest: evalDigest(wrapper.Document.Text), Kind: "macro", Start: 0, End: len(wrapper.Document.Text), ObjectID: "b"}}
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Valid || !report.Coverage.Complete || report.EffectiveAnalysis.Document.Text != "eval w=9" {
		t.Fatalf("context-bound macro: status=%s coverage=%+v effective=%q edges=%+v gaps=%+v", report.Status, report.Coverage, report.EffectiveAnalysis.Document.Text, report.Traversal, report.Gaps)
	}
}

func TestEvaluateDatasetLeafAndModuleBodyBoundary(t *testing.T) {
	dataset := evalDef("d", "dataset", "Model.Data", "")
	model := evalDef("model", "data_model", "Model", "")
	query := "| tstats count FROM datamodel=Model.Data"
	complete, err := Evaluate(evalRequest(query, []Collection{{Kind: "dataset", Coverage: "complete"}, {Kind: "data_model", Coverage: "complete"}}, dataset, model))
	if err != nil {
		t.Fatal(err)
	}
	if complete.Status != analysis.Valid || !edgeWith(complete, "dataset", "Model.Data", "d") || !complete.Coverage.TraversedDefinitions || !complete.Coverage.Collections {
		t.Fatalf("dataset leaf: status=%s coverage=%+v edges=%+v gaps=%+v", complete.Status, complete.Coverage, complete.Traversal, complete.Gaps)
	}
	partial, err := Evaluate(evalRequest(query, []Collection{{Kind: "dataset", Coverage: "partial"}, {Kind: "data_model", Coverage: "complete"}}, dataset, model))
	if err != nil {
		t.Fatal(err)
	}
	if partial.Status != analysis.Incomplete || !edgeWith(partial, "dataset", "Model.Data", "d") || partial.Coverage.Collections {
		t.Fatalf("partial dataset leaf: %+v %+v", partial.Coverage, partial.Traversal)
	}
	module := evalDef("mod", "module", "m", "")
	saved := evalDef("s", "saved_search", "Daily", "search index=main", Relation{Kind: "module", Name: "m", Property: stringPointer("/module")})
	report, err := Evaluate(evalRequest("| from savedsearch:Daily", []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "module", Coverage: "complete"}}, saved, module))
	if err != nil {
		t.Fatal(err)
	}
	if report.Coverage.TraversedDefinitions || !edgeWith(report, "module", "m", "mod") {
		t.Fatalf("module body boundary: %+v %+v", report.Coverage, report.Traversal)
	}
}

func TestEvaluateLeavesRequestDetached(t *testing.T) {
	macro := macroDef("m", "m", "lookup users user OUTPUT role")
	lookup := evalDef("l", "lookup", "users", "")
	req := evalRequest("`m`", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, lookup)
	before, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	after, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("request mutated: before=%s after=%s", before, after)
	}
	report.DirectAnalysis.Document.Text = "changed"
	report.DefinitionAnalyses[0].DirectAnalysis.Document.Text = "changed"
	if req.Document.Text != "`m`" || req.Bundle.Objects[0].Document.Text != "lookup users user OUTPUT role" {
		t.Fatalf("report aliases request: %+v", req)
	}
}

func TestEvaluateDynamicSavedSearchKeepsUnknownEdge(t *testing.T) {
	report, err := Evaluate(evalRequest("| from \"savedsearch:$daily$\"", []Collection{{Kind: "saved_search", Coverage: "complete"}}))
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != analysis.Incomplete {
		t.Fatalf("dynamic search became %s", report.Status)
	}
	found := false
	for _, edge := range report.Traversal {
		if edge.Kind == "saved_search" && edge.Name == "$daily$" && edge.Resolution == "dynamic" {
			found = true
		}
	}
	if !found {
		t.Fatalf("dynamic edge missing: %+v", report.Traversal)
	}
}

func TestEvaluateFoundObjectRequiresRelevantCollectionCoverage(t *testing.T) {
	lookup := evalDef("l", "lookup", "users", "")
	for _, tc := range []struct {
		name        string
		collections []Collection
		complete    bool
	}{
		{"complete", []Collection{{Kind: "lookup", Coverage: "complete"}, {Kind: "macro", Coverage: "partial"}}, true},
		{"partial", []Collection{{Kind: "lookup", Coverage: "partial"}}, false},
		{"unavailable", []Collection{{Kind: "lookup", Coverage: "unavailable"}}, false},
		{"omitted", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			report, err := Evaluate(evalRequest("| lookup users user OUTPUT role", tc.collections, lookup))
			if err != nil {
				t.Fatal(err)
			}
			if !edgeWith(report, "lookup", "users", "l") || report.Traversal[0].Resolution != "resolved" || report.Coverage.Collections != tc.complete || report.Coverage.Complete != tc.complete {
				t.Fatalf("coverage=%+v edges=%+v gaps=%+v", report.Coverage, report.Traversal, report.Gaps)
			}
			want := analysis.Incomplete
			if tc.complete {
				want = analysis.Valid
			}
			if report.Status != want {
				t.Fatalf("status=%s want=%s", report.Status, want)
			}
		})
	}
}
func TestEvaluateMacroOriginParentAndCycle(t *testing.T) {
	macro := macroDef("m", "m", "lookup $name$ user OUTPUT role", "name")
	lookup := evalDef("l", "lookup", "L", "", Relation{Kind: "macro", Name: "m", Property: stringPointer("/back")})
	req := evalRequest("`m(L)`", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, lookup)
	report, err := Evaluate(req)
	if err != nil {
		t.Fatal(err)
	}
	var lookupEdge *TraversalEdge
	for i := range report.Traversal {
		edge := &report.Traversal[i]
		if edge.Kind == "lookup" && edge.Name == "L" {
			lookupEdge = edge
			break
		}
	}
	if lookupEdge == nil {
		t.Fatalf("lookup edge absent: %+v", report.Traversal)
	}
	if lookupEdge.FromObjectID != "m" || len(lookupEdge.Path) != 1 || lookupEdge.Path[0] != "m" || len(lookupEdge.InvocationChain) != 1 || lookupEdge.InvocationChain[0].ObjectID != "m" || lookupEdge.InvocationChain[0].InstanceID == "" {
		t.Fatalf("semantic parent/instance: %+v", lookupEdge)
	}
	rootOrigin, placeholder := false, false
	for _, origin := range lookupEdge.Origins {
		rootOrigin = rootOrigin || origin.Kind == "query" && origin.SourceID == "root.spl" && origin.Start == 3 && origin.End == 4
		placeholder = placeholder || origin.Kind == "definition" && origin.ObjectID == "m" && origin.Start == 7 && origin.End == 13
	}
	if !rootOrigin || !placeholder {
		t.Fatalf("source argument and placeholder not retained: %+v", lookupEdge.Origins)
	}
	cycle := false
	for _, edge := range report.Traversal {
		if edge.Kind == "macro" && edge.Name == "m" && edge.FromObjectID == "l" && strings.Join(edge.CyclePath, ",") == "m,l,m" {
			cycle = true
		}
	}
	if !cycle || report.Status != analysis.Incomplete || report.Coverage.Complete {
		t.Fatalf("macro/lookup cycle: status=%s coverage=%+v edges=%+v", report.Status, report.Coverage, report.Traversal)
	}
}
func TestEvaluateRepeatedMacroFramesStayDistinct(t *testing.T) {
	macro := macroDef("m", "m", "lookup $name$ user OUTPUT role", "name")
	a := evalDef("a", "lookup", "A", "")
	b := evalDef("b", "lookup", "B", "")
	report, err := Evaluate(evalRequest("`m(A)` | `m(B)`", []Collection{{Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}, macro, a, b))
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, edge := range report.Traversal {
		if edge.Kind != "lookup" {
			continue
		}
		if edge.FromObjectID != "m" || len(edge.InvocationChain) != 1 {
			t.Fatalf("macro parent/frames: %+v", edge)
		}
		ids[edge.InvocationChain[0].InstanceID] = true
	}
	if len(ids) != 2 {
		t.Fatalf("distinct macro instances=%v edges=%+v", ids, report.Traversal)
	}
}

func TestEvaluateRelationOnlyMacroBodyKeepsKnownEdges(t *testing.T) {
	macro := macroDef("m", "m", "lookup secret user OUTPUT role")
	users := evalDef("users", "lookup", "users", "", Relation{Kind: "macro", Name: "m", Property: stringPointer("/macro")})
	report, err := Evaluate(evalRequest("| lookup users user OUTPUT role", []Collection{{Kind: "lookup", Coverage: "complete"}, {Kind: "macro", Coverage: "complete"}}, users, macro))
	if err != nil {
		t.Fatal(err)
	}
	if !edgeWith(report, "lookup", "users", "users") || !edgeWith(report, "macro", "m", "m") {
		t.Fatalf("reachable edges: %+v", report.Traversal)
	}
	missing, contextGap := false, false
	for _, edge := range report.Traversal {
		if edge.Kind == "lookup" && edge.Name == "secret" && edge.FromObjectID == "m" && edge.Resolution == "missing" {
			missing = true
		}
	}
	for _, gap := range report.Gaps {
		if gap.Code == "macro_context_missing" && gap.Kind == "macro" && gap.Name == "m" {
			contextGap = true
		}
	}
	if !missing || !contextGap || report.Status != analysis.Incomplete || report.Coverage.Complete {
		t.Fatalf("relation-only macro closure: status=%s coverage=%+v edges=%+v gaps=%+v", report.Status, report.Coverage, report.Traversal, report.Gaps)
	}
}

func TestEvaluateRepeatedDefinitionBodyKeepsOccurrencePaths(t *testing.T) {
	users := evalDef("users", "lookup", "users", "")
	child := evalDef("child", "saved_search", "Child", "lookup users user OUTPUT role")
	a := evalDef("a", "saved_search", "A", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	b := evalDef("b", "saved_search", "B", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	collections := []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}
	report, err := Evaluate(evalRequest("| from savedsearch:A | from savedsearch:B", collections, a, b, child, users))
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}
	ids := map[string]bool{}
	analysisCount := 0
	for _, edge := range report.Traversal {
		if edge.FromObjectID == "child" && edge.Kind == "lookup" && edge.Name == "users" {
			paths[strings.Join(edge.Path, "/")] = true
			ids[edge.ID] = true
		}
	}
	for _, body := range report.DefinitionAnalyses {
		if body.ObjectID == "child" {
			analysisCount++
		}
	}
	if !paths["a/child"] || !paths["b/child"] || len(ids) != 2 || analysisCount != 1 {
		t.Fatalf("DAG occurrences/body: paths=%v ids=%v analyses=%d edges=%+v", paths, ids, analysisCount, report.Traversal)
	}
}
func TestResolveBodyAnalysisMemo(t *testing.T) {
	def := evalDef("child", "saved_search", "Child", "lookup users user OUTPUT role")
	req := evalRequest("", []Collection{{Kind: "saved_search", Coverage: "complete"}}, def)
	e := &evaluator{req: req, directCache: map[string]*analysis.Result{}, bodyCache: map[string]bodyEvaluation{}, expandedByResult: map[*analysis.Result]expansion{}}
	first, err := e.bodyFor(def)
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.bodyFor(def)
	if err != nil {
		t.Fatal(err)
	}
	if first.direct != second.direct || first.effective != second.effective || len(e.bodyCache) != 1 {
		t.Fatalf("body analysis was not reused: first=%p/%p second=%p/%p cache=%d", first.direct, first.effective, second.direct, second.effective, len(e.bodyCache))
	}
}

func TestEvaluateCachedBodyRebasesMacroInstances(t *testing.T) {
	macro := macroDef("m", "m", "lookup users user OUTPUT role")
	users := evalDef("users", "lookup", "users", "")
	child := evalDef("child", "saved_search", "Child", "`m`")
	a := evalDef("a", "saved_search", "A", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	b := evalDef("b", "saved_search", "B", "search index=main", Relation{Kind: "saved_search", Name: "Child", Property: stringPointer("/child")})
	collections := []Collection{{Kind: "saved_search", Coverage: "complete"}, {Kind: "macro", Coverage: "complete"}, {Kind: "lookup", Coverage: "complete"}}
	report, err := Evaluate(evalRequest("| from savedsearch:A | from savedsearch:B", collections, a, b, child, macro, users))
	if err != nil {
		t.Fatal(err)
	}
	instances := map[string]bool{}
	for _, edge := range report.Traversal {
		if edge.FromObjectID != "m" || edge.Kind != "lookup" || edge.Name != "users" {
			continue
		}
		if len(edge.InvocationChain) != 1 {
			t.Fatalf("missing macro frame: %+v", edge)
		}
		instances[edge.InvocationChain[0].InstanceID] = true
	}
	if len(instances) != 2 {
		t.Fatalf("cached body reused expansion instance ID: %v edges=%+v", instances, report.Traversal)
	}
}
