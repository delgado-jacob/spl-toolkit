package analysis

import (
	"reflect"
	"strings"
	"testing"
)

// A damaged stage must neither swallow its original adjacent stages nor turn
// delimiters inside typed strings, objects, arrays, comments or regex into pipes.
func TestSPL2CanonicalRecoveryAdjacentStages(t *testing.T) {
	for _, text := range []string{
		`FROM main | eval broken= | where host=bytes`,
		`FROM main | eval broken= | eval also= | where host=bytes`,
		`FROM main | eval a=["x|y",{key:"[|]"}], broken= | where host=bytes`,
		`FROM main | eval broken= /* | where invented=1 */ | where host=bytes`,
		`FROM main | rex /a|b/ | eval broken= | where host=bytes`,
	} {
		t.Run(text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, text)
			if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("recovery truth: %+v", r)
			}
			for _, name := range []string{"host", "bytes"} {
				ref := spl2Ref(t, r, name, "read")
				if ref.OriginalName != name || !strings.HasSuffix(text[ref.Location.Start.Offset:], name) && ref.Location.Start.Offset < strings.LastIndex(text, "where") {
					t.Fatalf("lost adjacent original read: %+v", ref)
				}
			}
			for _, ref := range r.References {
				if strings.Contains(ref.OriginalName, "missing") || ref.NormalizedName == "invented" {
					t.Fatalf("invented reference: %+v", ref)
				}
			}
		})
	}
}

func TestSPL2CanonicalUnknownAndExclusions(t *testing.T) {
	for _, tt := range []struct {
		text   string
		status Status
		code   string
		read   bool
	}{
		{`mystery host=1`, Incomplete, CodeUnsupportedSemantics, false},
		{`FROM main | mystery host=1 | where bytes>0`, Incomplete, CodeUnsupportedSemantics, true},
		{`FROM main | mystery host=1 | eval x= | where bytes>0`, Invalid, CodeSyntaxError, true},
		{`FROM main | branch [where host=1 | into output] | where bytes>0`, Incomplete, CodeUnsupportedSemantics, true},
		{`FROM main | thru output | where bytes>0`, Incomplete, CodeUnsupportedSemantics, true},
		{`FROM main | into output`, Incomplete, CodeUnsupportedSemantics, false},
		{`FROM main | route output`, Invalid, "SPL_PROFILE_MISMATCH", false},
		{`import foo`, Invalid, "SPL_UNSUPPORTED_MODULE", false},
		{`FROM main; FROM other`, Invalid, "SPL_UNSUPPORTED_MODULE", false},
	} {
		t.Run(tt.text, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tt.text)
			if r.Status != tt.status || !spl2HasCode(r, tt.code) || r.Coverage.SemanticComplete {
				t.Fatalf("classification: %+v", r)
			}
			if tt.read {
				spl2Ref(t, r, "bytes", "read")
			}
			if strings.Contains(tt.text, "branch") && spl2HasCode(r, "SPL_UNSUPPORTED_MODULE") {
				t.Fatal("native child became module")
			}
		})
	}
}

func TestSPL2CanonicalDamagedChildCannotEscape(t *testing.T) {
	for _, q := range []string{`FROM main | append [FROM child | eval x= | where secret=1`, `FROM main | append [FROM child | eval x=] | where host>0`} {
		r := spl2AnalyzeTest(t, q)
		if r.Status != Invalid || len(r.Scopes) != 1 {
			t.Fatalf("damaged child ownership: %+v", r)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "secret" || ref.NormalizedName == "child" {
				t.Fatalf("damaged child read escaped: %+v", ref)
			}
		}
		if strings.HasSuffix(q, "host>0") {
			spl2Ref(t, r, "host", "read")
		}
	}
}

func TestSPL2CanonicalScopesEnvironments(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM main | eval parent=1 | append [FROM child | eval own=parent] | appendpipe [eval inherited=parent] | if (flag=true) [eval a=parent] else [eval b=parent] | where own>0`)
	if r.Status != Incomplete || len(r.Scopes) != 5 {
		t.Fatalf("children: %+v", r)
	}
	bindings := []string{}
	for _, ref := range r.References {
		if ref.NormalizedName == "parent" && ref.Role == "read" {
			bindings = append(bindings, ref.Binding)
		}
	}
	if !reflect.DeepEqual(bindings, []string{"source", "derived", "derived", "derived"}) {
		t.Fatalf("independent/inherited reads: %+v", r.References)
	}
	if spl2Ref(t, r, "own", "read").Binding != "indeterminate" {
		t.Fatal("child output leaked as certain parent field")
	}
	for _, scope := range r.Scopes[1:] {
		if scope.ParentID != "scope-0" || scope.StageID == "" {
			t.Fatalf("scope owner: %+v", scope)
		}
	}
	spl2Ref(t, r, "flag", "read")
	spl2AssertAllScopeIDs(t, r)
}

func spl2AssertAllScopeIDs(t *testing.T, r *Result) {
	t.Helper()
	refs := map[string]bool{}
	stages := map[string]Stage{}
	positions := map[string]map[int]bool{}
	for i, stage := range r.Stages {
		if i > 0 && r.Stages[i-1].Location.Start.Offset > stage.Location.Start.Offset {
			t.Fatal("stages are not lexical")
		}
		stages[stage.ID] = stage
		if positions[stage.ScopeID] == nil {
			positions[stage.ScopeID] = map[int]bool{}
		}
		if positions[stage.ScopeID][stage.Position] {
			t.Fatal("duplicate per-scope stage position")
		}
		positions[stage.ScopeID][stage.Position] = true
	}
	for _, ref := range r.References {
		refs[ref.ID] = true
		if !strings.HasPrefix(ref.ID, "ref-") || stages[ref.StageID].ScopeID != ref.ScopeID {
			t.Fatalf("reference ownership: %+v", ref)
		}
	}
	check := func(ids []string) {
		for _, id := range ids {
			if !refs[id] {
				t.Fatalf("unfinalized reference %s", id)
			}
		}
	}
	for _, ref := range r.References {
		check(ref.OriginReferenceIDs)
	}
	for i, line := range r.Lineage {
		if line.ExecutionOrder == nil || *line.ExecutionOrder != i {
			t.Fatalf("execution order %+v", line)
		}
		for _, state := range []FieldState{line.Before, line.After} {
			for _, f := range state.Fields {
				check(f.OriginReferenceIDs)
			}
		}
		for _, tr := range line.Transitions {
			check(tr.InputReferenceIDs)
			if tr.OutputReferenceID != "" {
				check([]string{tr.OutputReferenceID})
			}
		}
	}
}

func TestSPL2CanonicalSQLNestedScopeFinalization(t *testing.T) {
	r := spl2AnalyzeTest(t, `SELECT 1 AS parent FROM main AS m WHERE EXISTS(SELECT c.id FROM child AS c WHERE c.id=m.id) | appendpipe [eval copy=parent | append [SELECT host AS child_name FROM other] | table copy] | table parent`)
	if r.Status != Incomplete || len(r.Scopes) != 4 {
		t.Fatalf("SQL children %+v", r)
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "c.id" || ref.NormalizedName == "m.id" {
			if ref.Binding != "indeterminate" {
				t.Fatalf("qualified read %+v", ref)
			}
		}
	}
	if spl2Ref(t, r, "parent", "read").Binding != "derived" {
		t.Fatal("SQL derived parent lost in child")
	}
	spl2AssertAllScopeIDs(t, r)
}

func TestSPL2CanonicalDatasetFields(t *testing.T) {
	for _, q := range []string{`FROM [{host:"a",n:1},{host:"b",n:2}] | table host,n`, `FROM [{host:"a",n:1},{host:"b",n:2}] SELECT host,n`} {
		r := spl2AnalyzeTest(t, q)
		if r.Status != Valid || spl2Ref(t, r, "host", "read").Binding != "derived" || spl2Ref(t, r, "n", "read").Binding != "derived" || len(r.Dependencies.Datasets) != 0 {
			t.Fatalf("literal sources: %+v", r)
		}
	}
	r := spl2AnalyzeTest(t, `FROM [{host:"a",optional:1},{host:"b"}] | where optional>0`)
	if r.Status != Valid || spl2Ref(t, r, "optional", "read").Binding != "indeterminate" {
		t.Fatalf("conditional literal key: %+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM [{host:"a"}] | where absent>0`)
	if r.Status != Invalid || spl2Ref(t, r, "absent", "read").Binding != "unavailable" {
		t.Fatalf("literal closedness: %+v", r)
	}
}

func TestSPL2CanonicalDamagedAncestorNoPromotedScope(t *testing.T) {
	q := `FROM main | append [FROM child | append [FROM nested | table secret] | eval x=] | where host>0`
	r := spl2AnalyzeTest(t, q)
	if r.Status != Invalid || len(r.Scopes) != 1 {
		t.Fatalf("promoted child through damaged ownership: %+v", r)
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "secret" || ref.NormalizedName == "nested" {
			t.Fatalf("child escaped: %+v", ref)
		}
	}
	spl2Ref(t, r, "host", "read")
}

func TestSPL2CanonicalRecoveredAdjacentChild(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM main | mystery flag=1 | append [FROM child | where host>0] | where bytes>0`)
	if r.Status != Incomplete || len(r.Scopes) != 2 {
		t.Fatalf("adjacent child lost: %+v", r)
	}
	if ref := spl2Ref(t, r, "host", "read"); ref.ScopeID != "scope-1" || ref.Binding != "source" {
		t.Fatalf("independent recovered child: %+v", ref)
	}
	spl2Ref(t, r, "bytes", "read")
	spl2AssertAllScopeIDs(t, r)
}

func TestSPL2CanonicalLexicalDamageCannotAuthorizeNames(t *testing.T) {
	for _, q := range []string{`FROM main | where host☃>0 | where bytes>0`, `FROM main | eval value=host☃ | where bytes>0`} {
		r := spl2AnalyzeTest(t, q)
		if r.Status != Invalid {
			t.Fatalf("lost lexical error %+v", r)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "host" {
				t.Fatalf("lexically damaged operand became read: %+v", ref)
			}
		}
		spl2Ref(t, r, "bytes", "read")
	}
}

func TestSPL2CanonicalChildEvidenceAndDynamicReads(t *testing.T) {
	q := `FROM main | eval prior=null, stable=1 | appendpipe [eval missing=isnull(prior), copy=stable, dyn=record[key], tpl='field_${free}', fn=($x)->{return $x+outside;}, nullable=tonumber("7")]`
	r := spl2AnalyzeTest(t, q)
	if r.Status != Incomplete || len(r.Scopes) != 2 {
		t.Fatalf("child evidence %+v", r)
	}
	if spl2Ref(t, r, "prior", "null_test").Binding != "unavailable" || spl2Ref(t, r, "stable", "read").Binding != "derived" {
		t.Fatal("lost null_test/derived inherited evidence")
	}
	for _, name := range []string{"record", "key", "free", "outside"} {
		spl2Ref(t, r, name, "read")
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "x" || ref.NormalizedName == "$x" {
			t.Fatal("lambda local became source")
		}
	}
	spl2AssertAllScopeIDs(t, r)
}

func TestSPL2CanonicalUnknownCannotHideAdjacentMalformedChild(t *testing.T) {
	for _, head := range []string{`mystery flag=1`, `branch [where flag=1 | into output]`} {
		for _, child := range []string{`append [FROM child | eval broken=]`, `appendpipe [eval broken=]`} {
			r := spl2AnalyzeTest(t, `FROM main | `+head+` | `+child+` | where host>0`)
			if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeSyntaxError) || len(r.Scopes) != 1 {
				t.Fatalf("hidden child syntax error: %+v", r)
			}
			spl2Ref(t, r, "host", "read")
		}
	}
}

func TestSPL2CanonicalChildExpansionFinalization(t *testing.T) {
	result, err := AnalyzeWithSourceUniverse(QueryDocument{Text: `SELECT host AS alias FROM main | appendpipe [fields 'a*'] | table alias`, Language: "spl2"}, SourceUniverse{Fields: []string{"host"}, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Result.Status != Incomplete || len(result.Expansions) != 1 {
		t.Fatalf("child expansion %+v", result)
	}
	expansion := result.Expansions[0]
	if !expansion.Complete || !reflect.DeepEqual(expansion.Matches, []ExpandedField{{Name: "alias", Binding: "derived"}}) {
		t.Fatalf("child projection evidence %+v", expansion)
	}
	reference := spl2Ref(t, result.Result, "a*", "read")
	if reference.ID != expansion.ReferenceID || reference.ScopeID != "scope-1" || reference.OriginalName != "'a*'" {
		t.Fatalf("child expansion final IDs %+v %+v", reference, expansion)
	}
	spl2AssertAllScopeIDs(t, result.Result)
}

func TestSPL2CanonicalUnknownPreservesLexerErrors(t *testing.T) {
	for _, q := range []string{`FROM main | mystery ☃ flag=1 | where host>0`, `FROM main | branch [☃ into output] | where host>0`} {
		r := spl2AnalyzeTest(t, q)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
			t.Fatalf("unknown classification erased lexical damage %+v", r)
		}
		spl2Ref(t, r, "host", "read")
		if len(r.Scopes) != 1 {
			t.Fatal("unmodeled owner invented child scopes")
		}
	}
}

func TestSPL2CanonicalEmptyDatasetRetainsH09(t *testing.T) {
	for _, q := range []string{`FROM []`, `FROM [] SELECT host`, `FROM [] | table host`} {
		r := spl2AnalyzeTest(t, q)
		if r.Status != Incomplete || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
			t.Fatalf("empty dataset escaped H09: %+v", r)
		}
		for _, ref := range r.References {
			if ref.Binding == "unavailable" {
				t.Fatal("held empty dataset established certain absence")
			}
		}
	}
}
