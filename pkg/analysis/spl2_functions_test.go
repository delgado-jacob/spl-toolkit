package analysis

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

func TestSPL2CurrentFunctionPolicyCoversObservedForms(t *testing.T) {
	policy, ok := spl2TypedPolicyFor("splunkd", "current")
	if !ok {
		t.Fatal("missing private splunkd/current typed policy")
	}
	want := []string{"abs", "any", "avg", "cidrmatch", "coalesce", "count", "dc", "distinct_count", "json", "json_array_to_mv", "like", "lower", "match", "max", "min", "mvindex", "round", "rtrim", "span", "sqrt", "stdev", "strftime", "sum", "tonumber", "values"}
	got := make([]string, 0, len(want))
	for _, name := range want {
		if _, exists := policy.functions[name]; exists {
			got = append(got, name)
		}
	}
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("observed function policy = %v, want %v", got, want)
	}
	for _, name := range []string{"where", "eval", "fields", "stats", "bin", "mvexpand"} {
		if _, ok := policy.commands[name]; !ok {
			t.Fatalf("missing selected command %q from private policy", name)
		}
	}
	for _, name := range []string{"avg", "count", "dc", "distinct_count", "max", "min", "stdev", "sum", "values"} {
		if policy.functions[name].position != spl2AggregateFunction || !policy.functions[name].signature.aggregate {
			t.Fatalf("aggregate policy for %q = %+v", name, policy.functions[name])
		}
	}
	if function := policy.functions["if"]; function.signature.min != 3 || function.signature.max != 3 || function.signature.aggregate {
		t.Fatalf("scalar wrapper policy = %+v", function)
	}
}

func TestSPL2FunctionNullabilityComesFromTypedPolicy(t *testing.T) {
	original := spl2TypedPolicies["splunkd/current"]
	modified := cloneSPL2TypedPolicyForTest(original)
	match := modified.functions["match"]
	match.nullability = spl2NullabilityAlways
	modified.functions["match"] = match
	spl2TypedPolicies["splunkd/current"] = modified
	t.Cleanup(func() { spl2TypedPolicies["splunkd/current"] = original })

	r := spl2AnalyzeTest(t, `FROM synthetic_events | eval synthetic_result=match("sample", "^sample")`)
	if r.Status != Valid || !r.Coverage.SemanticComplete {
		t.Fatalf("modified private policy was not selected: %+v", r)
	}
	for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
		if field.Name == "synthetic_result" {
			if field.Conditional {
				t.Fatalf("production evaluation ignored the policy nullability strategy: %+v", field)
			}
			return
		}
	}
	t.Fatalf("missing policy-controlled output: %+v", r)
}

func cloneSPL2TypedPolicyForTest(policy spl2TypedPolicy) spl2TypedPolicy {
	clone := spl2TypedPolicy{
		functions: make(map[string]spl2FunctionPolicy, len(policy.functions)),
		commands:  make(map[string]spl2CommandHandler, len(policy.commands)),
	}
	for name, function := range policy.functions {
		clone.functions[name] = function
	}
	for name, handler := range policy.commands {
		clone.commands[name] = handler
	}
	return clone
}

func TestSPL2MatchDoesNotProveNonnullOutput(t *testing.T) {
	for _, query := range []string{
		`FROM main | eval result=match(name,"^a")`,
		`FROM main | eval result=match(path,"/tmp/")`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Valid || !r.Coverage.SemanticComplete {
			t.Fatalf("match policy coverage: %+v", r)
		}
		fields := r.Lineage[len(r.Lineage)-1].After.Fields
		found := false
		for _, field := range fields {
			if field.Name == "result" {
				found = true
				if !field.Conditional {
					t.Fatalf("match result gained an unproved nonnull guarantee: %+v", field)
				}
			}
		}
		if !found {
			t.Fatalf("missing match result: %+v", r)
		}
	}
}

func TestSPL2ObservedFunctionFormsAreModeled(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		inputs []string
		role   string
	}{
		{"abs", `FROM synthetic_events | eval synthetic_out=abs(synthetic_value)`, []string{"synthetic_value"}, "read"},
		{"any", `FROM synthetic_events | eval synthetic_out=any(synthetic_values, $synthetic_member -> $synthetic_member > synthetic_floor)`, []string{"synthetic_values", "synthetic_floor"}, "read"},
		{"avg", `FROM synthetic_events | stats avg(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"cidrmatch", `FROM synthetic_events | eval synthetic_out=cidrmatch("198.51.100.0/24", synthetic_address)`, []string{"synthetic_address"}, "read"},
		{"coalesce", `FROM synthetic_events | eval synthetic_out=coalesce(synthetic_primary, synthetic_fallback)`, []string{"synthetic_primary", "synthetic_fallback"}, "read"},
		{"count", `FROM synthetic_events | stats count() AS synthetic_out`, nil, "read"},
		{"dc", `FROM synthetic_events | stats dc(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"distinct_count", `FROM synthetic_events | stats distinct_count(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"json", `FROM synthetic_events | eval synthetic_out=json(synthetic_object)`, []string{"synthetic_object"}, "read"},
		{"json_array_to_mv", `FROM synthetic_events | eval synthetic_out=json_array_to_mv(synthetic_json_array)`, []string{"synthetic_json_array"}, "read"},
		{"like", `FROM synthetic_events | eval synthetic_out=like(synthetic_label, "sample%")`, []string{"synthetic_label"}, "read"},
		{"lower", `FROM synthetic_events | eval synthetic_out=lower(synthetic_label)`, []string{"synthetic_label"}, "read"},
		{"match", `FROM synthetic_events | eval synthetic_out=match(synthetic_label, "^sample")`, []string{"synthetic_label"}, "read"},
		{"max", `FROM synthetic_events | stats max(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"min", `FROM synthetic_events | stats min(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"mvindex", `FROM synthetic_events | eval synthetic_out=mvindex(synthetic_values, 0)`, []string{"synthetic_values"}, "read"},
		{"round", `FROM synthetic_events | eval synthetic_out=round(synthetic_value, 2)`, []string{"synthetic_value"}, "read"},
		{"rtrim", `FROM synthetic_events | eval synthetic_out=rtrim(synthetic_label, ".")`, []string{"synthetic_label"}, "read"},
		{"span", `FROM synthetic_events | stats count() AS synthetic_count BY span(synthetic_time, 10m)`, []string{"synthetic_time"}, "group"},
		{"sqrt", `FROM synthetic_events | eval synthetic_out=sqrt(synthetic_value)`, []string{"synthetic_value"}, "read"},
		{"stdev", `FROM synthetic_events | stats stdev(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"strftime", `FROM synthetic_events | eval synthetic_out=strftime(synthetic_time, "%Y-%m-%d")`, []string{"synthetic_time"}, "read"},
		{"sum", `FROM synthetic_events | stats sum(synthetic_value) AS synthetic_out`, []string{"synthetic_value"}, "read"},
		{"tonumber", `FROM synthetic_events | eval synthetic_out=tonumber(synthetic_label)`, []string{"synthetic_label"}, "read"},
		{"values", `FROM synthetic_events | stats values(synthetic_label) AS synthetic_out`, []string{"synthetic_label"}, "read"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Valid || !r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || len(r.Diagnostics) != 0 {
				t.Fatalf("observed form must be fully modeled: %+v", r)
			}
			for _, input := range tc.inputs {
				ref := spl2Ref(t, r, input, tc.role)
				if ref.Binding != "source" || ref.Resolution != "exact" {
					t.Fatalf("input %q lost exact source ownership: %+v", input, ref)
				}
				if item := requirementItem(r.Requirements, "field", input, tc.role); item == nil || item.Necessity != "required" || item.Resolution != "exact" {
					t.Fatalf("input %q requirement = %+v", input, item)
				}
			}
			for _, ref := range r.References {
				if strings.Contains(ref.OriginalName, "$synthetic_member") || strings.Contains(ref.NormalizedName, "synthetic_member") {
					t.Fatalf("lambda local escaped into field references: %+v", ref)
				}
			}
			for _, item := range r.Requirements.Items {
				if strings.Contains(item.Identity, "synthetic_member") {
					t.Fatalf("lambda local escaped into requirements: %+v", item)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2ObservedFunctionBoundaries(t *testing.T) {
	for _, query := range []string{
		`FROM synthetic_events | eval synthetic_out=any(synthetic_values)`,
		`FROM synthetic_events | eval synthetic_out=cidrmatch(synthetic_address)`,
		`FROM synthetic_events | eval synthetic_out=json()`,
		`FROM synthetic_events | eval synthetic_out=json_array_to_mv(synthetic_json_array, 1)`,
		`FROM synthetic_events | eval synthetic_out=like(synthetic_label)`,
		`FROM synthetic_events | eval synthetic_out=mvindex(synthetic_values)`,
		`FROM synthetic_events | eval synthetic_out=sqrt(synthetic_value, 2)`,
		`FROM synthetic_events | eval synthetic_out=strftime(synthetic_time)`,
		`FROM synthetic_events | stats stdev() AS synthetic_out`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) {
			t.Fatalf("invalid selected signature accepted for %q: %+v", query, r)
		}
	}
	for _, query := range []string{
		`FROM synthetic_events | eval synthetic_out=stdev(synthetic_value)`,
		`FROM synthetic_events | stats sqrt(synthetic_value) AS synthetic_out`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) {
			t.Fatalf("invalid aggregate position accepted for %q: %+v", query, r)
		}
	}
	r := spl2AnalyzeTest(t, `FROM synthetic_events | eval synthetic_out=synthetic_unknown(synthetic_value)`)
	if r.Status != Incomplete || !spl2HasCode(r, CodeUnsupportedFunction) || spl2Ref(t, r, "synthetic_value", "read").Binding != "source" {
		t.Fatalf("unknown function boundary lost input evidence: %+v", r)
	}
}

func TestSPL2AnyLambdaShapeIsEnforced(t *testing.T) {
	t.Run("non-lambda argument", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | eval synthetic_out=any(synthetic_values, synthetic_member)`)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete {
			t.Fatalf("non-lambda any argument was accepted: %+v", r)
		}
		for _, name := range []string{"synthetic_values", "synthetic_member"} {
			if ref := spl2Ref(t, r, name, "read"); ref.Binding != "source" {
				t.Fatalf("rejected lambda shape lost parsed input %q: %+v", name, ref)
			}
		}
	})

	t.Run("unbound local", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | eval synthetic_out=any(synthetic_values, $synthetic_member -> $synthetic_missing > 0)`)
		if r.Status != Incomplete || !spl2HasCode(r, CodeUnsupportedSemantics) || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete {
			t.Fatalf("unbound lambda local was accepted: %+v", r)
		}
		for _, ref := range r.References {
			if strings.Contains(ref.OriginalName, "synthetic_missing") || strings.Contains(ref.NormalizedName, "synthetic_missing") {
				t.Fatalf("unbound lambda local escaped into field references: %+v", ref)
			}
		}
		for _, item := range r.Requirements.Items {
			if strings.Contains(item.Identity, "synthetic_missing") {
				t.Fatalf("unbound lambda local escaped into requirements: %+v", item)
			}
		}
	})
}

func TestSPL2NestedLenNumericPresence(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM main | eval n=abs(len("abc")) | table n`)
	if r.Status != Valid || !r.Coverage.SemanticComplete || spl2Ref(t, r, "n", "read").Binding != "derived" || r.Lineage[1].Transitions[0].Conditional {
		t.Fatalf("numeric len result must preserve supported abs presence: %+v", r)
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2FunctionResultDomains(t *testing.T) {
	for _, tc := range []struct {
		expression string
		status     Status
		binding    string
		input      string
	}{
		{`substr("abc",len("x"))`, Valid, "derived", ""},
		{`len(lower("ABC"))`, Valid, "derived", ""},
		{`split("a:b",":")`, Valid, "derived", ""},
		{`lower(split("a:b",":"))`, Valid, "indeterminate", ""},
		{`len(split("a:b",":"))`, Valid, "indeterminate", ""},
		{`lower(len("abc"))`, Valid, "indeterminate", ""},
		{`abs(len(null))`, Valid, "indeterminate", ""},
		{`abs(len(7))`, Valid, "indeterminate", ""},
		{`abs(len(value))`, Valid, "indeterminate", "value"},
		{`split(value,":")`, Valid, "indeterminate", "value"},
		{`split("a:b",null)`, Valid, "indeterminate", ""},
		{`mvcount(split("a:b",":"))`, Valid, "indeterminate", ""},
		{`abs(len(mystery(value)))`, Incomplete, "indeterminate", "value"},
		{`abs(len(value:input))`, Incomplete, "indeterminate", "input"},
		{`abs(len("abc",value))`, Invalid, "indeterminate", "value"},
		{`split(value)`, Invalid, "indeterminate", "value"},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval n="+tc.expression+" | table n")
			if r.Status != tc.status || spl2Ref(t, r, "n", "read").Binding != tc.binding || r.Lineage[1].Transitions[0].Conditional != (tc.binding == "indeterminate") {
				t.Fatalf("result presence/domain composition: %+v", r)
			}
			if tc.status == Valid && (!r.Coverage.SyntaxComplete || !r.Coverage.SemanticComplete || len(r.Diagnostics) != 0) {
				t.Fatalf("conditional presence must not change modeled analysis coverage: %+v", r)
			}
			if tc.input != "" && spl2Ref(t, r, tc.input, "read").Binding != "source" {
				t.Fatalf("original input read changed: %+v", r)
			}
			if strings.Contains(tc.expression, "value:") {
				for _, ref := range r.References {
					if ref.NormalizedName == "value" {
						t.Fatalf("named label became a read: %+v", ref)
					}
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

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
