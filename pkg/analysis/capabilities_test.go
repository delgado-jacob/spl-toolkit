package analysis

import "testing"

// Catches unsupported functions/forms being advertised or treated as complete.
func TestCapabilitiesFunctionForms(t *testing.T) {
	for _, q := range []string{"eval x=abs(a)", "eval x=round(a,2)", "eval x=ceil(a)", "eval x=ceiling(a)", "eval x=floor(a)", "eval x=len(a)", "eval x=lower(a)", "eval x=upper(a)", "eval x=trim(a)", "eval x=ltrim(a,\"x\")", "eval x=rtrim(a)", "eval x=substr(a,1,2)", "eval x=replace(a,\"x\",\"y\")", "eval x=coalesce(a,b)", "eval x=if(a,b,c)", "eval x=case(a,b,c,d)", "eval x=isnull(a)", "eval x=isnotnull(a)", "eval x=tonumber(a)", "eval x=tostring(a)", "eval x=mvcount(a)", "eval x=split(a,\",\")", "eval x=match(a,\"x\")", "stats count sum(a) avg(a) min(a) max(a) values(a) list(a) dc(a) distinct_count(a) first(a) last(a)"} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Valid {
			t.Errorf("%s: %s %+v", q, r.Status, r.Diagnostics)
		}
	}
	for _, q := range []string{"eval x=unknown(a)", "eval x=lower()", "eval x=if(a,b)", "eval x=case(a,b,c)", "eval x=searchmatch(\"a=b\")", "stats sum(*)", "fields a*", "search a=1 | mystery a | where z=2"} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Incomplete {
			t.Errorf("%s: %s %+v", q, r.Status, r.Diagnostics)
		}
	}
	m := Capabilities()
	if len(m.Functions) == 0 {
		t.Fatal("missing function manifest")
	}
	m.Functions[0].Name = "mutated"
	m.Functions[0].Limitations[0] = "mutated"
	m.Commands[0].Limitations[0] = "mutated"
	n := Capabilities()
	if n.Functions[0].Name == "mutated" || n.Functions[0].Limitations[0] == "mutated" || n.Commands[0].Limitations[0] == "mutated" {
		t.Fatal("manifest aliases caller memory")
	}
}

func TestCapabilitiesUnsupportedArgumentForms(t *testing.T) {
	for _, q := range []string{`stats sum("x")`, `stats sum(a+b)`, `stats sum(a,b) AS n`, `head 1.5`, `tail 2.5`, `dedup 0 a`, `sort 1.5 a`} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Incomplete {
			t.Errorf("%s: %s %+v", q, r.Status, r.Diagnostics)
		}
	}
	for _, tc := range []struct{ q, code string }{{`eval x=custom(a)`, CodeUnsupportedFunction}, {`eval x=searchmatch("a=b")`, CodeDynamicReference}, {`fields a*`, CodeUnresolvedWildcard}} {
		r, _ := Analyze(QueryDocument{Text: tc.q})
		found := false
		for _, d := range r.Diagnostics {
			if d.Code == tc.code {
				found = true
			}
		}
		if !found {
			t.Error(tc.q, "missing", tc.code)
		}
	}
}

// Command options must not become grouping keys, aggregates, or source requirements.
func TestCapabilitiesUnsupportedStatsOptions(t *testing.T) {
	for _, q := range []string{`search a=1 | streamstats window=5 sum(a) AS n`, `search a=1 | eventstats allnum=true avg(a) AS n`, `search a=1 | stats partitions=2 count`} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Incomplete || !r.Coverage.SyntaxComplete {
			t.Errorf("%s: %s %+v", q, r.Status, r.Diagnostics)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && (ref.NormalizedName == "window" || ref.NormalizedName == "allnum" || ref.NormalizedName == "partitions" || ref.NormalizedName == "true") {
				t.Fatal("option became a field", ref)
			}
		}
	}
}
