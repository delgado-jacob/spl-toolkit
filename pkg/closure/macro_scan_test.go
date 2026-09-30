package closure

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMacroScanFixtureAndByteRanges(t *testing.T) {
	source := "search café=1 `base()` | eval literal=\"skip `false()` and \\\"`also_false`\\\"\" | `field_filter(field=\"user\")`"
	calls := scanMacroInvocations(source)
	if len(calls) != 2 {
		t.Fatalf("calls = %+v", calls)
	}
	for _, c := range calls {
		for _, span := range []macroSpan{c.Span, c.NameSpan} {
			if span.Start < 0 || span.End > len(source) || span.End <= span.Start || !utf8.ValidString(source[:span.Start]) || !utf8.ValidString(source[:span.End]) {
				t.Fatalf("invalid byte span %+v", span)
			}
		}
		if source[c.NameSpan.Start:c.NameSpan.End] != c.Name || c.Unsupported != "" {
			t.Fatalf("call = %+v", c)
		}
	}
	if got := source[calls[0].Span.Start:calls[0].Span.End]; got != "`base()`" || calls[0].Position != "fragment" || len(calls[0].Arguments) != 0 {
		t.Fatalf("first call = %+v, text %q", calls[0], got)
	}
	if got := source[calls[1].Span.Start:calls[1].Span.End]; got != "`field_filter(field=\"user\")`" || calls[1].Position != "stage" {
		t.Fatalf("second call = %+v, text %q", calls[1], got)
	}
	arg := calls[1].Arguments[0]
	if arg.Name != "field" || source[arg.Span.Start:arg.Span.End] != `field="user"` || source[arg.NameSpan.Start:arg.NameSpan.End] != "field" || source[arg.ValueSpan.Start:arg.ValueSpan.End] != `"user"` {
		t.Fatalf("argument = %+v", arg)
	}
}

func TestMacroScanMultilineNestedAndRepeated(t *testing.T) {
	source := "`f(α, concat(\"a,b)\", (x+y)), key=(one(two)))`\n| `f(a)` | `f(b)`"
	calls := scanMacroInvocations(source)
	if len(calls) != 3 {
		t.Fatalf("calls = %+v", calls)
	}
	if calls[0].Name != "f" || calls[0].Unsupported != "" || len(calls[0].Arguments) != 3 || calls[0].Position != "stage" {
		t.Fatalf("nested = %+v", calls[0])
	}
	for i, want := range []string{"α", `concat("a,b)", (x+y))`, "key=(one(two))"} {
		arg := calls[0].Arguments[i]
		if got := source[arg.Span.Start:arg.Span.End]; got != want || !utf8.ValidString(source[:arg.Span.Start]) || !utf8.ValidString(source[:arg.Span.End]) {
			t.Fatalf("arg %d = %+v, text %q", i, arg, got)
		}
	}
	if calls[0].Arguments[2].Name != "key" || calls[1].Name != "f" || calls[2].Name != "f" {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestMacroScanUnsupportedForms(t *testing.T) {
	for _, source := range []string{
		"`bad(a,,b)`", "`bad(a=)`", "`bad(a=(b)`", "`bad(a, b`", "`bad(a=b=c)`", "`bad(a b)`", "`bad(?)`", "`.`", "`:bad`", "prefix`bad()`suffix",
	} {
		t.Run(source, func(t *testing.T) {
			calls := scanMacroInvocations(source)
			if len(calls) != 1 || calls[0].Unsupported == "" {
				t.Fatalf("calls = %+v", calls)
			}
			if calls[0].Span.Start < 0 || calls[0].Span.End > len(source) || calls[0].Span.End <= calls[0].Span.Start {
				t.Fatalf("span = %+v", calls[0].Span)
			}
		})
	}
}

func TestMacroScanStageFragmentAndQuotes(t *testing.T) {
	source := `"` + "`quoted`" + `" 'ignored ` + "`also_ignored`" + `' | ` + "`stage`" + ` | search host=` + "`value(x)`" + ` OR ` + "`term`" + ` | eval x=if(a,` + "`inside()`" + `,b)`
	calls := scanMacroInvocations(source)
	if len(calls) != 4 {
		t.Fatalf("calls = %+v", calls)
	}
	for i, want := range []string{"stage", "value", "term", "inside"} {
		if calls[i].Name != want || calls[i].Unsupported != "" {
			t.Fatalf("call %d = %+v", i, calls[i])
		}
	}
	if calls[0].Position != "stage" || calls[1].Position != "fragment" || calls[2].Position != "fragment" || calls[3].Position != "fragment" {
		t.Fatalf("positions = %+v", calls)
	}
	if strings.Contains(source[calls[0].Span.Start:calls[0].Span.End], "quoted") {
		t.Fatal(calls[0])
	}
}

func TestMacroScanEscapedArgumentQuotesAndBareName(t *testing.T) {
	source := "`bare` | search `quoted(\"a,\\\"b\\\"`c\", nested=one(two,three))`"
	calls := scanMacroInvocations(source)
	if len(calls) != 2 || calls[0].Unsupported != "" || calls[0].Name != "bare" || len(calls[0].Arguments) != 0 {
		t.Fatalf("bare = %+v", calls)
	}
	if calls[1].Unsupported != "" || len(calls[1].Arguments) != 2 {
		t.Fatalf("quoted = %+v", calls[1])
	}
	if got := source[calls[1].Arguments[0].Span.Start:calls[1].Arguments[0].Span.End]; got != "\"a,\\\"b\\\"`c\"" {
		t.Fatalf("quoted argument = %q", got)
	}
}

func TestMacroScanCommentsDoNotCreateInvocations(t *testing.T) {
	source := "search * // `line` |\n| `real()` /* `block` | */ | stats count"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Name != "real" || calls[0].Position != "stage" || calls[0].Unsupported != "" {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestMacroScanQuotedName(t *testing.T) {
	source := "| `'a\\'b'()` |"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Unsupported != "" || calls[0].Name != "a'b" || calls[0].Position != "stage" {
		t.Fatalf("calls = %+v", calls)
	}
	if got := source[calls[0].NameSpan.Start:calls[0].NameSpan.End]; got != `'a\'b'` {
		t.Fatalf("name span = %q", got)
	}
}

func TestMacroScanRecoversAfterMalformedDelimiter(t *testing.T) {
	source := "search `bad(a | `good()` | eval literal=\"`quoted()`\""
	calls := scanMacroInvocations(source)
	if len(calls) != 2 {
		t.Fatalf("calls = %+v", calls)
	}
	if calls[0].Unsupported == "" || calls[0].Span.Start != strings.Index(source, "`bad") || calls[0].Span.End != strings.Index(source, "|") {
		t.Fatalf("malformed call = %+v", calls[0])
	}
	if calls[1].Name != "good" || calls[1].Unsupported != "" || source[calls[1].Span.Start:calls[1].Span.End] != "`good()`" {
		t.Fatalf("recovered call = %+v", calls[1])
	}
}

func TestMacroScanComparisonArgumentsRemainPositional(t *testing.T) {
	source := "`filter(score>=10,status!=0,age<=18,field=\"user\")`"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Unsupported != "" || len(calls[0].Arguments) != 4 {
		t.Fatalf("calls = %+v", calls)
	}
	for i, want := range []string{"score>=10", "status!=0", "age<=18"} {
		arg := calls[0].Arguments[i]
		if arg.Name != "" || arg.NameSpan != (macroSpan{}) || source[arg.Span.Start:arg.Span.End] != want || arg.ValueSpan != arg.Span {
			t.Fatalf("comparison argument %d = %+v", i, arg)
		}
	}
	if arg := calls[0].Arguments[3]; arg.Name != "field" || source[arg.ValueSpan.Start:arg.ValueSpan.End] != `"user"` {
		t.Fatalf("named argument = %+v", arg)
	}
}

func TestMacroScanPipeInsideArgumentIsNotRecoveryBoundary(t *testing.T) {
	source := "`filter([ search host=web | stats count ], \"a|b\")` | `next()`"
	calls := scanMacroInvocations(source)
	if len(calls) != 2 || calls[0].Unsupported != "" || len(calls[0].Arguments) != 2 || calls[1].Name != "next" {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestMacroScanNestedInvocation(t *testing.T) {
	source := "`outer(`inner()`, \"quoted `false()`\")`"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Name != "outer" || calls[0].Unsupported != "" || len(calls[0].Arguments) != 2 {
		t.Fatalf("outer = %+v", calls)
	}
	outer := calls[0]
	if outer.Span != (macroSpan{0, len(source)}) || len(outer.Nested) != 1 || outer.Nested[0].Name != "inner" || outer.Nested[0].Unsupported != "" {
		t.Fatalf("nested calls = %+v", outer)
	}
	inner := outer.Nested[0]
	if got := source[inner.Span.Start:inner.Span.End]; got != "`inner()`" || inner.Span != outer.Arguments[0].Span || inner.Position != "fragment" {
		t.Fatalf("inner = %+v, text %q", inner, got)
	}
	if source[outer.Arguments[1].Span.Start:outer.Arguments[1].Span.End] != "\"quoted `false()`\"" {
		t.Fatalf("quoted argument = %+v", outer.Arguments[1])
	}
}

func TestMacroScanSameStageMalformedRecovery(t *testing.T) {
	source := "search `bad(a `good()` \"quoted `false()`\""
	calls := scanMacroInvocations(source)
	if len(calls) != 2 || calls[0].Unsupported == "" || calls[1].Unsupported != "" || calls[1].Name != "good" {
		t.Fatalf("calls = %+v", calls)
	}
	if calls[0].Span.Start != strings.Index(source, "`bad") || calls[0].Span.End != strings.Index(source, "`good") ||
		source[calls[1].Span.Start:calls[1].Span.End] != "`good()`" || len(calls[0].Nested) != 0 {
		t.Fatalf("ranges = %+v", calls)
	}
}

func TestMacroScanSubqueryStagePosition(t *testing.T) {
	for _, source := range []string{
		"search [ search host=web | `inner()` ]",
		"search [ `inner()` | stats count ]",
	} {
		calls := scanMacroInvocations(source)
		if len(calls) != 1 || calls[0].Name != "inner" || calls[0].Unsupported != "" || calls[0].Position != "stage" {
			t.Fatalf("source %q: calls = %+v", source, calls)
		}
	}
	fragment := scanMacroInvocations("search [ search host=`inner()` ]")
	if len(fragment) != 1 || fragment[0].Unsupported != "" || fragment[0].Position != "fragment" {
		t.Fatalf("subquery fragment = %+v", fragment)
	}
}

func TestMacroScanNestedArgumentSubqueryStagePosition(t *testing.T) {
	source := "`outer([ search host=web | `inner()` ])`"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Unsupported != "" || calls[0].Position != "stage" || len(calls[0].Nested) != 1 {
		t.Fatalf("outer = %+v", calls)
	}
	inner := calls[0].Nested[0]
	if inner.Name != "inner" || inner.Unsupported != "" || inner.Position != "stage" || source[inner.Span.Start:inner.Span.End] != "`inner()`" {
		t.Fatalf("inner = %+v", inner)
	}
}

func TestMacroScanRootStageAfterSubquery(t *testing.T) {
	source := "search [ search host=web | stats count ] | `root()`"
	calls := scanMacroInvocations(source)
	if len(calls) != 1 || calls[0].Name != "root" || calls[0].Unsupported != "" || calls[0].Position != "stage" {
		t.Fatalf("calls = %+v", calls)
	}
}

func TestMacroScanCommentsInArgumentSyntax(t *testing.T) {
	for _, tc := range []struct {
		source, value, name string
		arity               int
	}{
		{"`m(1/*,*/+2)`", "1/*,*/+2", "", 1},
		{"`m(x/*comment*/=1)`", "1", "x", 1},
		{"`m(x/*=*/=1)`", "1", "x", 1},
		{"`m((1/*)*/+2))`", "(1/*)*/+2)", "", 1},
		{"`m(1//,\n+2)`", "1//,\n+2", "", 1},
		{"`m(x//comment\n=1)`", "1", "x", 1},
		{"`m(/*,=*/)`", "", "", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			calls := scanMacroInvocations(tc.source)
			if len(calls) != 1 || calls[0].Unsupported != "" || len(calls[0].Arguments) != tc.arity {
				t.Fatalf("calls = %+v", calls)
			}
			if tc.arity == 0 {
				return
			}
			arg := calls[0].Arguments[0]
			if arg.Name != tc.name || tc.source[arg.ValueSpan.Start:arg.ValueSpan.End] != tc.value {
				t.Fatalf("argument = %+v", arg)
			}
			if tc.name != "" && tc.source[arg.NameSpan.Start:arg.NameSpan.End] != tc.name {
				t.Fatalf("name span = %+v", arg.NameSpan)
			}
		})
	}
}

func TestMacroScanRecoversAfterMissingCloseInSameStage(t *testing.T) {
	for _, source := range []string{
		"search `bad(a) `good()`",
		"search `bad(a)/*gap*/`good()`",
	} {
		t.Run(source, func(t *testing.T) {
			calls := scanMacroInvocations(source)
			if len(calls) != 2 || calls[0].Unsupported == "" || calls[1].Unsupported != "" || calls[1].Name != "good" {
				t.Fatalf("calls = %+v", calls)
			}
			if calls[0].Span.End != strings.Index(source, "`good") || source[calls[1].Span.Start:calls[1].Span.End] != "`good()`" {
				t.Fatalf("ranges = %+v", calls)
			}
		})
	}
}
