package analysis

import (
	"fmt"
	"strings"
	"testing"
)

func TestSPL2FiniteFunctionArities(t *testing.T) {
	signatures := []struct {
		names     string
		min, max  int
		aggregate bool
	}{{"abs ceil ceiling floor len lower upper isnull isnotnull mvcount", 1, 1, false}, {"round trim ltrim rtrim tonumber tostring", 1, 2, false}, {"substr", 2, 3, false}, {"replace if", 3, 3, false}, {"split match", 2, 2, false}, {"coalesce", 1, -1, false}, {"case", 2, -1, false}, {"count", 0, 1, true}, {"sum avg min max values list dc distinct_count first last", 1, 1, true}}
	for _, s := range signatures {
		for _, name := range strings.Fields(s.names) {
			maximum := s.max
			if maximum < 0 {
				maximum = 6
			}
			for n := 0; n <= maximum+1; n++ {
				t.Run(fmt.Sprintf("%s/%d", name, n), func(t *testing.T) {
					args := []string{}
					for j := 0; j < n; j++ {
						args = append(args, fmt.Sprintf("v%d", j))
					}
					call := name + "(" + strings.Join(args, ",") + ")"
					query := "FROM main | eval result=" + call
					if s.aggregate {
						query = "FROM main | stats " + call + " AS result"
					}
					r := spl2AnalyzeTest(t, query)
					valid := n >= s.min && (s.max < 0 || n <= s.max) && (name != "case" || n%2 == 0)
					if valid && (r.Status == Invalid || spl2HasCode(r, CodeUnsupportedFunction)) {
						t.Fatalf("valid arity rejected %+v", r)
					}
					if !valid && (r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError)) {
						t.Fatalf("invalid arity accepted %+v", r)
					}
					for j := 0; j < n; j++ {
						role := "read"
						if (name == "isnull" || name == "isnotnull") && n == 1 {
							role = "null_test"
						}
						spl2Ref(t, r, fmt.Sprintf("v%d", j), role)
					}
				})
			}
		}
	}
}
func TestSPL2FunctionBoundaries(t *testing.T) {
	for _, query := range []string{`FROM main | eval x=mystery(bytes)`, `FROM main | eval x=min(a,b)`, `FROM main | eval x=length(name)`, `FROM main | stats c() AS n`, `FROM main | stats count(bytes)`, `FROM main | stats sum(bytes+1)`} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Incomplete {
			t.Fatalf("%s: %+v", query, r)
		}
	}
	for _, query := range []string{`FROM main | eval x=sum(bytes)`, `FROM main | stats lower(name) AS n`} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) {
			t.Fatalf("%s: %+v", query, r)
		}
	}
	for _, query := range []string{`FROM main | eval x=round(precision:2,num:bytes)`, `FROM main | eval x=coalesce(values:[primary,backup])`, `FROM main | eval x=round(bytes,precision:2)`} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Incomplete {
			t.Fatalf("%+v", r)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "precision" || ref.NormalizedName == "num" || ref.NormalizedName == "values" {
				t.Fatalf("label read %+v", ref)
			}
		}
		if strings.Contains(query, "bytes") {
			spl2Ref(t, r, "bytes", "read")
		} else {
			spl2Ref(t, r, "primary", "read")
			spl2Ref(t, r, "backup", "read")
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | eval x=if(code=201,"created","other")`)
	if r.Status == Invalid {
		t.Fatalf("comparison became named argument: %+v", r)
	}
	spl2Ref(t, r, "code", "read")
}

func TestSPL2NullabilityEvidence(t *testing.T) {
	for _, expr := range []string{`coalesce(null,null)`, `if(code=201,null,null)`, `case(code=201,null)`, `case(code=201,null,true,null)`} {
		t.Run(expr, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval x="+expr+" | where x>0")
			if r.Status != Invalid || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
				t.Fatalf("%+v", r)
			}
			spl2Ref(t, r, "x", "remove")
			if strings.Contains(expr, "code") {
				spl2Ref(t, r, "code", "read")
			}
		})
	}
	for _, expr := range []string{`case(code=201,"a")`, `if(code=201,null,"a")`, `tonumber(value)`, `tostring(value)`, `mvcount(value)`, `coalesce(tonumber(value))`, `abs(value)`, `lower(value)`} {
		t.Run(expr, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval x="+expr)
			if r.Status != Valid || !r.Stages[1].SemanticComplete || !r.Lineage[1].Transitions[0].Conditional {
				t.Fatalf("%+v", r)
			}
			used := spl2AnalyzeTest(t, "FROM main | eval x="+expr+" | table x")
			if spl2Ref(t, used, "x", "read").Binding != "indeterminate" {
				t.Fatalf("%+v", used)
			}
		})
	}
	for _, expr := range []string{`coalesce(null,"fallback")`, `coalesce(tonumber(value),"fallback")`, `if(code=201,"a","b")`, `case(code=201,"a",true,"b")`, `abs(-7)`, `lower("ABC")`} {
		r := spl2AnalyzeTest(t, "FROM main | eval x="+expr+" | table x")
		if r.Status != Valid || spl2Ref(t, r, "x", "read").Binding != "derived" {
			t.Fatalf("%s %+v", expr, r)
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | fields - code | eval x=if(code=201,null,null) | where x>0`)
	if !spl2HasCode(r, CodeUnavailableField) || spl2Ref(t, r, "code", "read").Binding != "unavailable" || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
		t.Fatalf("%+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM main | eval x=if(mystery(code),null,null)`)
	if r.Status != Incomplete || spl2Ref(t, r, "x", "create").Binding != "not_applicable" {
		t.Fatalf("unknown condition promoted: %+v", r)
	}
	for _, name := range []string{"batch_id", "batch_time"} {
		r := spl2AnalyzeTest(t, "FROM main | eval x="+name+"()")
		if r.Status != Invalid || r.Coverage.SyntaxComplete || !spl2HasCode(r, "SPL_PROFILE_MISMATCH") || spl2HasCode(r, CodeUnsupportedFunction) {
			t.Fatalf("%+v", r)
		}
	}
}

func TestSPL2UnmodeledOperandsCannotProveOutputs(t *testing.T) {
	for _, expr := range []string{`if([mystery(code)],null,null)`, `if("${mystery(code)}",null,null)`, `if({key:mystery(code)},null,null)`} {
		r := spl2AnalyzeTest(t, "FROM main | eval x="+expr)
		if r.Status != Incomplete {
			t.Fatalf("%+v", r)
		}
		spl2Ref(t, r, "x", "create")
		spl2Ref(t, r, "code", "read")
		for _, tr := range r.Lineage[1].Transitions {
			if tr.Operation == "remove" || !tr.Conditional {
				t.Fatalf("unknown child proved effect %+v", r)
			}
		}
	}
}

func TestSPL2StringTemplatePresence(t *testing.T) {
	for _, expression := range []string{`"${null}"`, `"delta=${abs(high-low)}"`, `"value=${tonumber(value)}"`, `"owner=${account}"`} {
		t.Run(expression, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval label="+expression+" | table label")
			if r.Status != Valid || spl2Ref(t, r, "label", "read").Binding != "derived" || r.Lineage[1].Transitions[0].Conditional {
				t.Fatalf("intact template must produce a present string: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	r := spl2AnalyzeTest(t, `FROM main | eval gone=null | eval label="value=${gone}" | table label`)
	if r.Status != Invalid || !spl2HasCode(r, CodeUnavailableField) || spl2Ref(t, r, "gone", "read").Binding != "unavailable" || spl2Ref(t, r, "label", "read").Binding != "derived" {
		t.Fatalf("template must retain unavailable input finding and string output: %+v", r)
	}
	for _, expression := range []string{`"${mystery(value)}"`, `"${value.`} {
		r := spl2AnalyzeTest(t, "FROM main | eval label="+expression)
		if r.Status == Valid || r.Coverage.SemanticComplete {
			t.Fatalf("unknown/damaged interpolation promoted: %+v", r)
		}
		for _, tr := range r.Lineage[len(r.Lineage)-1].Transitions {
			if tr.Output == "label" && !tr.Conditional {
				t.Fatalf("unproved template output presence: %+v", r)
			}
		}
	}
}

func TestSPL2HeldPrefixDoesNotProvePresence(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM main | eval x=nOt enabled`)
	if r.Status != Incomplete || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
		t.Fatalf("held prefix promoted: %+v", r)
	}
	spl2Ref(t, r, "enabled", "read")
	spl2Ref(t, r, "x", "create")
	for _, f := range r.Lineage[len(r.Lineage)-1].After.Fields {
		if f.Name == "x" && !f.Conditional {
			t.Fatalf("held operator proved output: %+v", r)
		}
	}
	for _, q := range []string{`FROM main | eval x=NOT enabled`, `FROM main | eval x=not[0]`} {
		control := spl2AnalyzeTest(t, q)
		if q == `FROM main | eval x=NOT enabled` && control.Status != Valid {
			t.Fatalf("documented prefix changed: %+v", control)
		}
		if q == `FROM main | eval x=not[0]` {
			spl2Ref(t, control, "not", "read")
		}
	}
}
func TestSPL2AggregateWildcardOperandsRemainIncomplete(t *testing.T) {
	for _, name := range []string{"sum", "count"} {
		r := spl2AnalyzeTest(t, "FROM main | stats "+name+"('bytes*') AS n")
		if r.Status != Incomplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
			t.Fatalf("%+v", r)
		}
		ref := spl2Ref(t, r, "bytes*", "read")
		if ref.Resolution != "wildcard" || ref.Binding != "indeterminate" {
			t.Fatalf("wildcard became exact %+v", ref)
		}
	}
}
