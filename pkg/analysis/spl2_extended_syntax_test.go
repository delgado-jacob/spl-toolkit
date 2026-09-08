package analysis

import "testing"

func TestSPL2ExtendedSyntaxContracts(t *testing.T) {
	for _, tt := range []struct {
		query         string
		invalid, held bool
	}{
		{`FROM main | join left=L right=R where L.id=R.id [FROM other]`, false, false},
		{`FROM main | join left=L where L.id=R.id [FROM other]`, true, false},
		{`FROM main | join type=full left=L right=R where L.id=R.id [FROM other]`, true, false},
		{`union main`, true, false},
		{`FROM main | union other`, false, false},
		{`FROM main | spath output=name`, true, false},
		{`FROM main | spath output=name path="a.b"`, false, false},
		{`FROM main | bin span=0.5log10 bytes`, true, false},
		{`FROM main | bin span=2log1 bytes`, true, false},
		{`FROM main | bin span=2log10 bytes`, false, false},
		{`FROM main | bin span=1w _time`, false, true},
		{`FROM main | bin span=hour _time`, false, true},
		{`FROM main | timechart agg=sum count()`, false, true},
		{`FROM main | timechart avg(size),max(size)`, false, true},
		{`FROM main | timechart avg(size) max(size)`, false, true},
		{`FROM main | timechart span=w@w1 count()`, false, false},
		{`FROM main | timechart span=w@ count()`, true, false},
		{`FROM main | timewrap week align=start`, true, false},
		{`FROM main | rex field=payload /(?<tag>\w+)/`, false, false},
		{`FROM main | rex field=payload /(?<tag>\w+)`, true, false},
		{`FROM main | rex field=payload /(?<tag>red|blue)/`, false, true},
		{`FROM main | spl1 "stats count by host"`, false, true},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			invalid := false
			for _, d := range p.diagnostics {
				if d.Severity == "error" {
					invalid = true
				}
				if d.Location.Start.Offset < 0 || d.Location.End.Offset > len(tt.query) {
					t.Fatal("unlocated finding")
				}
			}
			if invalid != tt.invalid || tt.held && p.syntaxComplete {
				t.Fatalf("invalid=%v syntax=%v diagnostics=%+v", invalid, p.syntaxComplete, p.diagnostics)
			}
			if p.semanticComplete {
				t.Fatal("unmodeled effects became complete")
			}
		})
	}
}

func TestSPL2ExtendedSyntaxOptionBoundaries(t *testing.T) {
	for _, tt := range []struct {
		query         string
		invalid, held bool
	}{
		{`FROM main | bin mystery=1 bytes`, false, true},
		{`FROM main | spath mystery=1`, false, true},
		{`FROM main | makemv mystery=1 labels`, false, true},
		{`FROM main | mvexpand mystery=1 labels`, false, true},
		{`FROM main | mvcombine mystery=1 labels`, false, true},
		{`FROM main | fillnull mystery=1 labels`, false, true},
		{`FROM main | timechart mystery=1 count()`, false, true},
		{`tstats aggregates=[count()] mystery=1`, false, true},
		{`mstats aggregates=[count()] datamodel_name='A.B'`, false, true},
		{`FROM main | append maxtime=30 [FROM archive]`, true, false},
		{`FROM main | appendcols override=true [FROM archive]`, true, false},
		{`FROM main | makemv allowempty=true labels`, true, false},
		{`FROM main | makemv setsv=true labels`, true, false},
		{`FROM main | timechart usenull=false count() BY host`, true, false},
		{`FROM main | timechart avg('byte*')`, true, false},
		{`FROM main | bin span=1log2 bytes`, false, false},
		{`FROM main | bin span=log bytes`, false, false},
		{`FROM main | bin span=10log10 bytes`, true, false},
		{`FROM main | bin span=2.5sec _time`, true, false},
		{`FROM main | timewrap 2.5day`, true, false},
		{`FROM main | timewrap fortnight`, false, true},
		{`FROM main | timechart span=fortnight count()`, false, true},
		{`FROM main | join type=left max=2.5 left=A right=B where A.id=B.id [FROM other]`, true, false},
		{`FROM main | join type= left=A right=B where A.id=B.id [FROM other]`, true, false},
		{`FROM main | append [FROM other | rex /(?<n>[0-9]+)/] | fields n`, false, false},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			bad := false
			for _, d := range p.diagnostics {
				bad = bad || d.Severity == "error"
			}
			if bad != tt.invalid || tt.held && p.syntaxComplete {
				t.Fatalf("invalid=%v syntax=%v diagnostics %+v", bad, p.syntaxComplete, p.diagnostics)
			}
		})
	}
}

func TestSPL2ExtendedSyntaxAdditionalContracts(t *testing.T) {
	for _, tt := range []struct {
		query         string
		invalid, held bool
	}{
		{`FROM main | JOIN left=A right=B where A.id=B.id [FROM other]`, false, true},
		{`FROM main | join type=INNER left=A right=B where A.id=B.id [FROM other]`, false, true},
		{`FROM main | makemv _raw`, true, false},
		{`FROM main | makemv '_time'`, true, false},
		{`FROM main | mvcombine _raw`, true, false},
		{`FROM main | mvcombine 'ordinary'`, false, false},
		{`FROM main | timechart span=2.5log10 count()`, false, false},
		{`FROM main | bin span=log2.5 bytes`, false, false},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			bad := false
			for _, d := range p.diagnostics {
				bad = bad || d.Severity == "error"
			}
			if bad != tt.invalid || tt.held && p.syntaxComplete {
				t.Fatalf("invalid %v syntax %v diagnostics %+v", bad, p.syntaxComplete, p.diagnostics)
			}
		})
	}
}

func TestSPL2ExtendedSyntaxSourceGuards(t *testing.T) {
	for _, tt := range []struct {
		query         string
		invalid, held bool
	}{
		{`FROM main | timechart avg('${"x*"}')`, false, false},
		{`FROM main | timechart count() BY host mystery=1`, false, true},
		{`FROM main | rex mystery=1 "(?<n>[0-9]+)"`, false, true},
		{`FROM main | appendpipe mystery=1 [fields n]`, false, true},
		{`makeresults mystery=1`, false, true},
		{`loadjob mystery=1 1780123000.5`, false, true},
		{`loadjob ignore_running=true 1780123000.5`, true, false},
		{`tstats aggregates=[count()] byfields=[host span=5min]`, false, true},
		{`FROM main | bin span=2.5 bytes`, true, false},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			bad := false
			for _, d := range p.diagnostics {
				bad = bad || d.Severity == "error"
			}
			if bad != tt.invalid || tt.held && p.syntaxComplete {
				t.Fatalf("invalid %v syntax %v diagnostics %+v", bad, p.syntaxComplete, p.diagnostics)
			}
		})
	}
}

func TestSPL2ExtendedSyntaxRemovedOptionsAndOwners(t *testing.T) {
	for _, tt := range []struct {
		query   string
		invalid bool
	}{
		{`mstats aggregates=[count()] backfill=true`, true},
		{`mstats aggregates=[count()] chart=true`, true},
		{`mstats aggregates=[count()] update_period=5`, true},
		{`mstats aggregates=[count()] span=1`, true},
		{`mstats aggregates=[count()] local=true`, false},
		{`tstats aggregates=[count()] chart=true`, false},
		{`FROM main | append [tstats aggregates=[count()] local=true]`, true},
		{`FROM main | append [FROM other | spath maxtime=1]`, false},
		{`FROM main | append [FROM other | stats mode="edge" count()]`, true},
	} {
		t.Run(tt.query, func(t *testing.T) {
			p := parseSPL2Document(tt.query)
			bad := false
			for _, d := range p.diagnostics {
				bad = bad || d.Severity == "error"
			}
			if bad != tt.invalid || p.syntaxComplete {
				t.Fatalf("invalid %v syntax %v diagnostics %+v", bad, p.syntaxComplete, p.diagnostics)
			}
		})
	}
}

func TestSPL2ExtendedSyntaxExactLogDomains(t *testing.T) {
	for _, query := range []string{`FROM main | bin span=log1.00000000000000000001 bytes`, `FROM main | bin span=1.00000000000000000001log1.00000000000000000002 bytes`} {
		t.Run(query, func(t *testing.T) {
			p := parseSPL2Document(query)
			for _, d := range p.diagnostics {
				if d.Severity == "error" {
					t.Fatalf("rounded a valid literal domain: %+v", d)
				}
			}
		})
	}
}

func TestSPL2ExtendedSyntaxMinimumSpanForms(t *testing.T) {
	for _, query := range []string{`FROM main | bin minspan=log2 bytes`, `FROM main | timechart minspan=log2 count()`, `FROM main | timechart minspan=w@w1 count()`} {
		t.Run(query, func(t *testing.T) {
			p := parseSPL2Document(query)
			if p.syntaxComplete {
				t.Fatal("minspan inherited an unproved span alternative")
			}
			for _, d := range p.diagnostics {
				if d.Severity == "error" {
					t.Fatalf("unproved alternative became invalid: %+v", d)
				}
			}
		})
	}
}
