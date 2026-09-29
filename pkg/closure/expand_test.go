package closure

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func macroDef(id, name, body string, args ...string) Definition {
	n := len(args)
	return Definition{ID: id, Kind: "macro", Name: name, SourceID: id + ".conf", Document: &analysis.QueryDocument{Text: body, Language: "spl"}, Arity: &n, Arguments: args}
}
func macroRequest(text string, defs ...Definition) Request {
	return Request{Document: analysis.QueryDocument{Text: text, Language: "spl", SourceID: "query"}, Bundle: DefinitionBundle{Objects: defs}}
}
func bindingFor(text string, start, end int, id string) Binding {
	sum := sha256.Sum256([]byte(text))
	return Binding{DocumentDigest: "sha256:" + hex.EncodeToString(sum[:]), Kind: "macro", Start: start, End: end, ObjectID: id}
}
func assertCovered(t *testing.T, got expansion) {
	t.Helper()
	at := 0
	for _, seg := range got.Segments {
		if seg.EffectiveStart != at || seg.EffectiveEnd <= at || seg.EffectiveEnd > len(got.Text) {
			t.Fatalf("bad coverage at %d: %+v for %q", at, seg, got.Text)
		}
		at = seg.EffectiveEnd
	}
	if at != len(got.Text) {
		t.Fatalf("coverage ends at %d, text length %d", at, len(got.Text))
	}
}
func TestExpandFieldFlowAndZeroArg(t *testing.T) {
	req := macroRequest("`make` | eval total=created+1", macroDef("m", "make", "eval created=1"))
	got := expandMacros(req)
	if got.Text != "eval created=1 | eval total=created+1" || len(got.Gaps) != 0 {
		t.Fatalf("expanded: %+v", got)
	}
	assertCovered(t, got)
	analyzed, err := analysis.Analyze(analysis.QueryDocument{Text: got.Text, Language: "spl"})
	if err != nil || analyzed.Status != analysis.Valid {
		t.Fatalf("field flow: %v, %+v", err, analyzed)
	}
}
func TestExpandArgumentsOverloadAndRepeatedReplacement(t *testing.T) {
	req := macroRequest("`m(α)` | `m(y=2,x=1)`", macroDef("zero", "m", "wrong"), macroDef("one", "m", "eval v=\"$x$:$x$\"", "x"), macroDef("two", "m", "eval v=$x$+$y$", "x", "y"))
	got := expandMacros(req)
	if got.Text != "eval v=\"α:α\" | eval v=1+2" || len(got.Gaps) != 0 {
		t.Fatalf("expanded: %+v", got)
	}
	assertCovered(t, got)
	if got.Segments[0].Source.ObjectID != "one" {
		t.Fatalf("first source: %+v", got.Segments[0])
	}
	var replacements int
	for _, seg := range got.Segments {
		if seg.Placeholder != nil {
			replacements++
			if seg.Placeholder.ObjectID != "one" && seg.Placeholder.ObjectID != "two" {
				t.Fatalf("placeholder: %+v", seg)
			}
		}
	}
	if replacements != 4 {
		t.Fatalf("replacements=%d", replacements)
	}
}
func TestExpandNestedAndRepeatedObjectInstances(t *testing.T) {
	req := macroRequest("`outer(1)` | `outer(2)`", macroDef("outer", "outer", "eval n=$x$ | `inner($x$)`", "x"), macroDef("inner", "inner", "eval z=$y$", "y"))
	got := expandMacros(req)
	if got.Text != "eval n=1 | eval z=1 | eval n=2 | eval z=2" || len(got.Gaps) != 0 {
		t.Fatalf("expanded: %+v", got)
	}
	assertCovered(t, got)
	ids := map[string]bool{}
	for _, seg := range got.Segments {
		for _, frame := range seg.InvocationChain {
			if frame.ObjectID == "outer" {
				ids[frame.InstanceID] = true
			}
		}
	}
	if len(ids) != 2 {
		t.Fatalf("outer instances = %v", ids)
	}
	var nested bool
	for _, seg := range got.Segments {
		if seg.Source.ObjectID == "inner" {
			nested = true
			if len(seg.InvocationChain) != 2 || seg.InvocationChain[0].ObjectID != "outer" || seg.InvocationChain[1].ObjectID != "inner" {
				t.Fatalf("nested chain: %+v", seg.InvocationChain)
			}
		}
	}
	if !nested {
		t.Fatal("missing nested provenance")
	}
}
func TestExpandAmbiguousBindingAndOpaqueSites(t *testing.T) {
	a := macroDef("a", "same", "eval x=1")
	b := macroDef("b", "same", "eval x=2")
	req := macroRequest("`same` | `same`", a, b)
	req.Bindings = []Binding{bindingFor(req.Document.Text, 0, 6, "b")}
	got := expandMacros(req)
	if got.Text != "eval x=2 | `same`" || len(got.Gaps) != 1 || got.Gaps[0].Reason != "ambiguous" {
		t.Fatalf("binding: %+v", got)
	}
	assertCovered(t, got)
	cases := []struct {
		name   string
		def    Definition
		reason string
	}{
		{"missing", macroDef("other", "other", "x"), "missing"},
		{"eval", func() Definition { d := macroDef("m", "m", "x"); v := true; d.EvalBased = &v; return d }(), "eval_based"},
		{"validation", func() Definition { d := macroDef("m", "m", "x"); v := "isnum($x$)"; d.Validation = &v; return d }(), "validation_held"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := expandMacros(macroRequest("`m`", tc.def))
			if out.Text != "`m`" || len(out.Gaps) != 1 || out.Gaps[0].Reason != tc.reason {
				t.Fatalf("opaque: %+v", out)
			}
			assertCovered(t, out)
		})
	}
}
func TestMacroCycleAndEmptyBody(t *testing.T) {
	got := expandMacros(macroRequest("`a` | `empty`", macroDef("a", "a", "eval x=1 | `a`"), macroDef("empty", "empty", "")))
	if got.Text != "eval x=1 | `a` | " || len(got.Gaps) != 1 || got.Gaps[0].Reason != "cycle" {
		t.Fatalf("cycle: %+v", got)
	}
	assertCovered(t, got)
}
func TestExpansionLimitDepthAndBytes(t *testing.T) {
	defs := make([]Definition, 34)
	for i := range defs {
		name := strings.Repeat("a", i+1)
		body := "eval x=1"
		if i+1 < len(defs) {
			body = "eval x=1 | `" + strings.Repeat("a", i+2) + "`"
		}
		defs[i] = macroDef(name, name, body)
	}
	got := expandMacros(macroRequest("`a`", defs...))
	if len(got.Gaps) != 1 || got.Gaps[0].Reason != "limit" || !strings.HasPrefix(got.Text, "eval x=1 | ") {
		t.Fatalf("depth: %+v", got)
	}
	assertCovered(t, got)
	huge := strings.Repeat("x", 1<<20)
	bytes := expandMacros(macroRequest("`large`", macroDef("large", "large", huge+"x")))
	if len(bytes.Text) != 1<<20 || len(bytes.Gaps) != 1 || bytes.Gaps[0].Reason != "limit" {
		t.Fatalf("byte limit: text=%d gaps=%+v", len(bytes.Text), bytes.Gaps)
	}
	assertCovered(t, bytes)
}
func TestExpansionLimitCalls(t *testing.T) {
	input := strings.Repeat("`empty` ", 4097)
	got := expandMacros(macroRequest(input, macroDef("empty", "empty", "")))
	if len(got.Gaps) != 1 || got.Gaps[0].Reason != "limit" || !strings.Contains(got.Text, "`empty`") {
		t.Fatalf("call limit: text=%q gaps=%+v", got.Text, got.Gaps)
	}
	assertCovered(t, got)
}

func TestExpandNestedArgumentAndDefinitionBinding(t *testing.T) {
	inner := macroDef("inner", "inner", "2")
	outer := macroDef("outer", "outer", "eval n=$x$", "x")
	arg := expandMacros(macroRequest("`outer(`inner`)`", inner, outer))
	if arg.Text != "eval n=2" || len(arg.Gaps) != 0 {
		t.Fatalf("nested argument: %+v", arg)
	}
	a := macroDef("a", "same", "eval v=1")
	b := macroDef("b", "same", "eval v=2")
	wrapper := macroDef("wrapper", "wrapper", "`same` | `same`")
	req := macroRequest("`wrapper`", a, b, wrapper)
	req.Bindings = []Binding{bindingFor(wrapper.Document.Text, 0, 6, "b")}
	out := expandMacros(req)
	if out.Text != "eval v=2 | `same`" || len(out.Gaps) != 1 || out.Gaps[0].Reason != "ambiguous" {
		t.Fatalf("definition binding: %+v", out)
	}
	assertCovered(t, out)
}

func TestExpandAmbiguousDefinitionBindingAcrossSubstitution(t *testing.T) {
	a := macroDef("a", "same", "eval v=$x$", "x")
	b := macroDef("b", "same", "eval w=$x$", "x")
	wrapper := macroDef("wrapper", "wrapper", "`same($x$)`", "x")
	req := macroRequest("`wrapper(9)`", a, b, wrapper)
	req.Bindings = []Binding{bindingFor(wrapper.Document.Text, 0, len(wrapper.Document.Text), "b")}
	got := expandMacros(req)
	if got.Text != "eval w=9" || len(got.Gaps) != 0 {
		t.Fatalf("definition binding across substitution: %+v", got)
	}
	assertCovered(t, got)
}

func TestExpandInvalidDefinitionRemainsDefinite(t *testing.T) {
	got := expandMacros(macroRequest("`bad`", macroDef("bad", "bad", "eval =")))
	if got.Text != "eval =" || len(got.Gaps) != 0 {
		t.Fatalf("expanded: %+v", got)
	}
	analyzed, err := analysis.Analyze(analysis.QueryDocument{Text: got.Text, Language: "spl"})
	if err != nil || analyzed.Status != analysis.Invalid {
		t.Fatalf("invalid content: %v, status=%q diagnostics=%+v", err, analyzed.Status, analyzed.Diagnostics)
	}
}
func TestExpansionLimitPreservesUTF8(t *testing.T) {
	source := strings.Repeat("x", maxExpansionBytes-1) + "é"
	got := expandMacros(macroRequest("`m`", macroDef("m", "m", source)))
	if len(got.Text) != maxExpansionBytes-1 || !utf8.ValidString(got.Text) || len(got.Gaps) != 1 || got.Gaps[0].Reason != "limit" {
		t.Fatalf("UTF-8 limit: bytes=%d gaps=%+v", len(got.Text), got.Gaps)
	}
	assertCovered(t, got)
}

func TestExpandLiteralDollarAndNamedArgumentOrder(t *testing.T) {
	def := macroDef("m", "m", "eval price=\"$unknown$\", v=$x$+$y$", "x", "y")
	req := macroRequest("`m(y=`two`,x=`one`)`", def, macroDef("one", "one", "1"), macroDef("two", "two", "2"))
	got := expandMacros(req)
	if got.Text != "eval price=\"$unknown$\", v=1+2" || len(got.Gaps) != 0 {
		t.Fatalf("dollar: %+v", got)
	}
	assertCovered(t, got)
}
