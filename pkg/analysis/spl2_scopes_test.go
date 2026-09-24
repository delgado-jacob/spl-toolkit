package analysis

import (
	"reflect"
	"strings"
	"testing"
)

func TestSPL2ScopesTypedOwnership(t *testing.T) {
	query := `FROM main | eval a=[{key:"[|]"}], b=($x)->{return $x;} | append [FROM other | appendpipe [eval nested=1]] | if (flag=true) [fields a] else [where flag=false]`
	p := parseSPL2Document(query)
	if !p.syntaxComplete {
		t.Fatalf("syntax: %+v", p.diagnostics)
	}
	// Typed child contexts, not lexical bracket counting, establish four scopes.
	if got := len(spl2Nodes(p.syntax, "independentSearch")); got != 1 {
		t.Fatalf("independent children %d", got)
	}
	if got := len(spl2Nodes(p.syntax, "inheritedSubpipe")); got != 3 {
		t.Fatalf("inherited children %d", got)
	}
}

func TestSPL2ScopesDescriptors(t *testing.T) {
	query := "FROM [{name:\"é\"}]\r\n | append [FROM other | appendpipe [eval n=1]] | if (true) [fields n] else [where n=1]"
	p := parseSPL2Document(query)
	scopes := spl2ChildScopes(p)
	want := []struct {
		text, input string
		parent      int
	}{
		{`[FROM other | appendpipe [eval n=1]]`, "independent", -1},
		{`[eval n=1]`, "inherited", 0},
		{`[fields n]`, "inherited", -1},
		{`[where n=1]`, "inherited", -1},
	}
	if len(scopes) != len(want) {
		t.Fatalf("scopes: %+v diagnostics %+v", scopes, p.diagnostics)
	}
	for i, s := range scopes {
		if query[s.location.Start.Offset:s.location.End.Offset] != want[i].text || s.input != want[i].input || s.parent != want[i].parent || s.owner == nil || s.body == nil {
			t.Fatalf("scope %d %+v", i, s)
		}
		if s.location.Start.Line != 2 {
			t.Fatal("lost CRLF location")
		}
	}
}

func TestSPL2ScopesSQLChildAndBracketRoles(t *testing.T) {
	query := `FROM main AS m WHERE EXISTS(SELECT c.id FROM child AS c WHERE c.id=m.id) | union [{name:"[FROM fake]"}], [FROM other | eval a=[1,2], b=($x)->{return $x;} ]`
	p := parseSPL2Document(query)
	scopes := spl2ChildScopes(p)
	if len(scopes) != 2 || scopes[0].kind != "exists" || scopes[0].input != "correlated" || scopes[1].kind != "search" {
		t.Fatalf("scopes %+v diagnostics %+v", scopes, p.diagnostics)
	}
	if query[scopes[0].location.Start.Offset:scopes[0].location.End.Offset] != `EXISTS(SELECT c.id FROM child AS c WHERE c.id=m.id)` {
		t.Fatal("SQL child scope rewritten")
	}
	if len(spl2Nodes(p.syntax, "array")) != 2 || len(spl2Nodes(p.syntax, "lambdaBlock")) != 1 {
		t.Fatal("compound literal scopes confused")
	}
}

func TestSPL2ScopesRecovery(t *testing.T) {
	for _, tt := range []struct {
		query string
		count int
	}{
		{`FROM main | append [FROM other | eval x=]`, 0},
		{`FROM main | append [FROM other] | where`, 1},
		{`FROM main | if (true) [fields]`, 0},
		{`FROM main | eval a=["[FROM x]"], b={key:"[eval x=1]"}`, 0},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			scopes := spl2ChildScopes(p)
			if len(scopes) != tt.count {
				t.Fatalf("scopes %+v diagnostics %+v", scopes, p.diagnostics)
			}
			for _, scope := range scopes {
				if scope.owner.GetStart().GetTokenIndex() < 0 || scope.owner.GetStop().GetTokenIndex() < 0 {
					t.Fatal("invented scope token")
				}
			}
		})
	}
}

func TestSPL2ScopeSchedulerReturnsChildEnvironmentAndTraceFork(t *testing.T) {
	query := `FROM main | appendpipe [eval child=1]`
	parsed := parseSPL2Document(query)
	if !parsed.syntaxComplete {
		t.Fatalf("syntax: %+v", parsed.diagnostics)
	}
	trace := newRequirementTrace()
	result := newResult(QueryDocument{Text: query, Language: "spl2", Profile: "splunkd", Version: "current"})
	result.Scopes = append(result.Scopes, Scope{ID: "scope-0", Kind: "root", Location: parsed.source.location(0, len(parsed.source.positions)-1)})
	scheduler := &spl2ScopeScheduler{
		result:   result,
		parsed:   parsed,
		trace:    trace,
		children: spl2ChildScopes(parsed),
		executed: map[int]bool{},
	}
	parent := closedMergeEnvironment()
	parent.requirements.trace = trace
	root := atomicFieldIdentity("root")
	installMergeField(parent, root, []string{"pending-root"})

	execution, ok := scheduler.executeChild(0, parent, map[string]bool{}, "scope-0", -1)
	if !ok || execution.Environment == nil || execution.Trace == nil {
		t.Fatalf("child execution = %+v ok=%t", execution, ok)
	}
	if execution.Trace == trace || len(trace.references) != 0 || len(execution.Trace.references) == 0 {
		t.Fatalf("child trace did not fork: parent=%p/%+v child=%p/%+v", trace, trace.references, execution.Trace, execution.Trace.references)
	}
	if _, ok := parent.field(atomicFieldIdentity("child")); ok {
		t.Fatal("child output escaped into parent environment")
	}
	if child, ok := execution.Environment.field(atomicFieldIdentity("child")); !ok || child.Conditional {
		t.Fatalf("returned child output = %+v known=%t", child, ok)
	}
	if inherited, ok := execution.Environment.field(root); !ok || inherited.Conditional {
		t.Fatalf("returned child lost inherited input = %+v known=%t", inherited, ok)
	}
}

func TestSPL2ScopeLocalViewChildPreservesFork(t *testing.T) {
	query := `$base = FROM synthetic_events | fields value; $out = FROM main | append [FROM $base | where value>0];`
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != Incomplete || result.Coverage.SemanticComplete {
		t.Fatalf("append boundary = status %s coverage %+v diagnostics %+v", result.Status, result.Coverage, result.Diagnostics)
	}
	trace.assertReferences(result.References)

	var summaryRead, childRead *Reference
	var viewRead *Reference
	for i := range result.References {
		reference := &result.References[i]
		switch {
		case reference.Kind == "view" && reference.NormalizedName == "$base":
			viewRead = reference
		case reference.Kind == "field" && reference.NormalizedName == "value" && reference.ScopeID == "scope-1":
			summaryRead = reference
		case reference.Kind == "field" && reference.NormalizedName == "value" && reference.ScopeID != "scope-1":
			childRead = reference
		}
	}
	if summaryRead == nil || viewRead == nil || childRead == nil {
		t.Fatalf("local-view child evidence missing: %+v", result.References)
	}
	if childRead.Binding != "source" || !reflect.DeepEqual(childRead.OriginReferenceIDs, []string{summaryRead.ID}) {
		t.Fatalf("child local-view origin = %+v want summary %s", childRead, summaryRead.ID)
	}
	entries := requirementTraceReferencesByID(trace)
	for _, reference := range []*Reference{viewRead, childRead} {
		entry, ok := entries[reference.ID]
		if !ok || entry.reference.ScopeID != reference.ScopeID || entry.reference.StageID != reference.StageID {
			t.Fatalf("child evidence missing from forked trace for %s: %+v", reference.ID, entry)
		}
	}
	foundSummary := false
	for _, lineage := range result.Lineage {
		for _, field := range lineage.After.Fields {
			if lineage.ScopeID == childRead.ScopeID && field.Name == "value" && reflect.DeepEqual(field.OriginReferenceIDs, []string{summaryRead.ID}) {
				foundSummary = true
			}
		}
	}
	if !foundSummary {
		t.Fatalf("child lineage lost local-view summary: %+v", result.Lineage)
	}
}

func TestSPL2ScopeForwardLocalViewPipelineChildPreservesFork(t *testing.T) {
	query := `$out = FROM main | append [FROM $base | where value>0]; $base = FROM synthetic_events | fields value;`
	result := assertSPL2ForwardLocalViewChild(t, query)
	if result.Status != Incomplete || result.Coverage.SemanticComplete {
		t.Fatalf("append boundary = status %s coverage %+v diagnostics %+v", result.Status, result.Coverage, result.Diagnostics)
	}
}

func TestSPL2ScopeForwardLocalViewSQLChildPreservesFork(t *testing.T) {
	query := `$out = FROM main AS m WHERE EXISTS(SELECT value FROM $base WHERE value>0) SELECT m.id; $base = FROM synthetic_events | fields value;`
	result := assertSPL2ForwardLocalViewChild(t, query)
	if result.Status != Invalid || !result.Coverage.SyntaxComplete {
		t.Fatalf("SQL child contract = status %s coverage %+v diagnostics %+v", result.Status, result.Coverage, result.Diagnostics)
	}
	foundCorrelationDiagnostic := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeSyntaxError && strings.Contains(diagnostic.Message, "requires equality correlation") {
			foundCorrelationDiagnostic = true
			break
		}
	}
	if !foundCorrelationDiagnostic {
		t.Fatalf("SQL child lost existing correlation diagnostic: %+v", result.Diagnostics)
	}
}

func TestSPL2ScopeForwardLocalViewRepeatedChildrenDoNotDuplicateSuffixes(t *testing.T) {
	query := `$out = FROM main | append [FROM $base | where value>0] | append [FROM $base | where value<10]; $base = FROM synthetic_events | fields value;`
	result := assertSPL2ForwardLocalViewChild(t, query)
	viewReads, childReads := 0, 0
	for _, reference := range result.References {
		if reference.Kind == "view" && reference.NormalizedName == "$base" {
			viewReads++
		}
		if reference.Kind == "field" && reference.NormalizedName == "value" && len(reference.OriginReferenceIDs) > 0 {
			childReads++
		}
	}
	if viewReads != 2 || childReads != 2 {
		t.Fatalf("repeated forward-view suffixes = view reads %d child reads %d: %+v", viewReads, childReads, result.References)
	}
}

func TestSPL2ScopeForwardLocalViewNestedChildDoesNotDuplicateSuffixes(t *testing.T) {
	query := `$out = FROM main | append [FROM other | append [FROM $base | where value>0]]; $base = FROM synthetic_events | fields value;`
	result := assertSPL2ForwardLocalViewChild(t, query)
	viewReads := 0
	for _, reference := range result.References {
		if reference.Kind == "view" && reference.NormalizedName == "$base" {
			viewReads++
		}
	}
	if viewReads != 1 {
		t.Fatalf("nested forward-view suffix duplicated %d times: %+v", viewReads, result.References)
	}
}

func assertSPL2ForwardLocalViewChild(t *testing.T, query string) *Result {
	t.Helper()
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	trace.assertReferences(result.References)
	seenTraceIDs := map[string]bool{}
	for _, entry := range trace.references {
		if seenTraceIDs[entry.reference.ID] {
			t.Fatalf("parent trace was mutated before child reconciliation: duplicate %s in %+v", entry.reference.ID, trace.references)
		}
		seenTraceIDs[entry.reference.ID] = true
	}

	stageCommands := map[string]string{}
	for _, stage := range result.Stages {
		stageCommands[stage.ID] = stage.Command
	}
	var summaryRead, viewRead *Reference
	childReads := []*Reference{}
	for i := range result.References {
		reference := &result.References[i]
		switch {
		case reference.Kind == "view" && reference.NormalizedName == "$base":
			viewRead = reference
		case reference.Kind == "field" && reference.NormalizedName == "value" && stageCommands[reference.StageID] == "fields":
			summaryRead = reference
		case reference.Kind == "field" && reference.NormalizedName == "value":
			childReads = append(childReads, reference)
		}
	}
	if summaryRead == nil || viewRead == nil || len(childReads) == 0 {
		t.Fatalf("forward local-view child evidence missing: %+v", result.References)
	}
	entries := requirementTraceReferencesByID(trace)
	if _, ok := entries[viewRead.ID]; !ok {
		t.Fatalf("view read %s missing from child trace", viewRead.ID)
	}
	for _, childRead := range childReads {
		if childRead.Binding != "source" || !reflect.DeepEqual(childRead.OriginReferenceIDs, []string{summaryRead.ID}) {
			t.Fatalf("forward child local-view origin = %+v want summary %s", childRead, summaryRead.ID)
		}
		if _, ok := entries[childRead.ID]; !ok {
			t.Fatalf("child field read %s missing from reconciled trace", childRead.ID)
		}
	}
	foundSummary := false
	for _, lineage := range result.Lineage {
		for _, field := range lineage.After.Fields {
			if field.Name == "value" && reflect.DeepEqual(field.OriginReferenceIDs, []string{summaryRead.ID}) {
				foundSummary = true
			}
		}
	}
	if !foundSummary {
		t.Fatalf("forward child lineage lost local-view summary: %+v", result.Lineage)
	}
	return result
}
