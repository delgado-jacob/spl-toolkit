package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
)

func TestCapabilitiesPublishCanonicalLedger(t *testing.T) {
	for _, options := range []CapabilityOptions{{}, {Language: "spl2"}} {
		manifest, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		if manifest.ToolkitVersion != buildinfo.Version {
			t.Errorf("toolkit version = %q, want %q", manifest.ToolkitVersion, buildinfo.Version)
		}
		gotRecords := manifest.Records
		gotEvidence := manifest.Evidence
		if len(gotRecords) == 0 || len(gotEvidence) == 0 {
			t.Fatalf("selected ledger is empty: records=%d evidence=%d", len(gotRecords), len(gotEvidence))
		}
		if got := manifest.Summary; !reflect.DeepEqual(got, summarizeCapabilityRecords(gotRecords)) {
			t.Errorf("summary is stale: got=%+v recomputed=%+v", got, summarizeCapabilityRecords(gotRecords))
		}
		evidenceIDs := make(map[string]struct{}, len(gotEvidence))
		for _, item := range gotEvidence {
			evidenceIDs[item.ID] = struct{}{}
		}
		for _, record := range gotRecords {
			for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
				for _, id := range dimension.claim.EvidenceIDs {
					if _, found := evidenceIDs[id]; !found {
						t.Errorf("record %q references evidence %q outside selected manifest", record.ID, id)
					}
				}
			}
		}
	}
}

func TestCapabilitiesSurviveJSONRoundTrip(t *testing.T) {
	for _, options := range []CapabilityOptions{{}, {Language: "spl2"}} {
		manifest, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		var decoded CapabilityManifest
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(decoded, manifest) {
			t.Fatal("capability manifest changed across its JSON wire representation")
		}
	}
}

func TestCapabilitiesPreserveLegacyProjection(t *testing.T) {
	spl2, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		got  CapabilityManifest
		want CapabilityManifest
	}{
		{name: "spl", got: Capabilities(), want: legacySPLCapabilities()},
		{name: "spl2", got: spl2, want: legacySPL2Capabilities()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.got.Commands, tc.want.Commands) {
				t.Fatalf("commands changed:\n got: %+v\nwant: %+v", tc.got.Commands, tc.want.Commands)
			}
			if !reflect.DeepEqual(tc.got.Functions, tc.want.Functions) {
				t.Fatalf("functions changed:\n got: %+v\nwant: %+v", tc.got.Functions, tc.want.Functions)
			}
		})
	}
}

func TestCapabilitiesReturnDetachedLedger(t *testing.T) {
	for _, options := range []CapabilityOptions{{}, {Language: "spl2"}} {
		first, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		second, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		before, err := json.Marshal(second)
		if err != nil {
			t.Fatal(err)
		}

		mutatedClaim := false
		for i := range first.Records {
			if len(first.Records[i].Dimensions.Syntax.EvidenceIDs) != 0 {
				first.Records[i].Dimensions.Syntax.EvidenceIDs[0] = "mutated"
				mutatedClaim = true
				break
			}
		}
		if !mutatedClaim {
			t.Fatal("selected records have no nested evidence IDs to test ownership")
		}
		mutatedObservation := false
		for i := range first.Evidence {
			semantics := first.Evidence[i].Observations.Semantics
			if semantics != nil && len(semantics.Stages) != 0 {
				semantics.Stages[0].Command = "mutated"
				mutatedObservation = true
				break
			}
		}
		if !mutatedObservation {
			t.Fatal("selected evidence has no nested semantic stage to test ownership")
		}
		mutatedRaw := false
		for i := range first.Evidence {
			if len(first.Evidence[i].RewriteRequest) != 0 {
				first.Evidence[i].RewriteRequest[0] = '['
				mutatedRaw = true
				break
			}
		}
		if !mutatedRaw {
			t.Fatal("selected evidence has no rewrite request to test ownership")
		}
		first.Summary.Syntax.Supported = -1
		first.Rewrite.Forms[0].IdentityForms[0] = "mutated"
		first.Rewrite.Forms[0].Limitations[0] = "mutated"

		afterManifest, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		after, err := json.Marshal(afterManifest)
		if err != nil {
			t.Fatal(err)
		}
		if string(after) != string(before) {
			t.Fatalf("manifest mutation escaped into later call:\n before: %s\n  after: %s", before, after)
		}
	}
}

func TestCapabilitiesRewriteFormsMatchCompatibilityFixture(t *testing.T) {
	var fixture struct {
		Checks []struct {
			Language  string `json:"language"`
			Kind      string `json:"kind"`
			Role      string `json:"role"`
			Supported bool   `json:"supported"`
		} `json:"capability_checks"`
	}
	raw, err := os.ReadFile("../../testdata/rewrite/forms.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	fixtureSupport := make(map[string]bool, len(fixture.Checks))
	for _, check := range fixture.Checks {
		key := check.Language + "/" + check.Kind + "/" + check.Role
		if _, duplicate := fixtureSupport[key]; duplicate {
			t.Fatalf("duplicate rewrite compatibility fixture %q", key)
		}
		fixtureSupport[key] = check.Supported
	}
	advertised := 0
	for _, options := range []CapabilityOptions{{}, {Language: "spl2"}} {
		manifest, err := CapabilitiesFor(options)
		if err != nil {
			t.Fatal(err)
		}
		if manifest.Rewrite == nil || len(manifest.Rewrite.Forms) == 0 {
			t.Fatal("rewrite compatibility projection is empty")
		}
		for _, form := range manifest.Rewrite.Forms {
			key := manifest.Language + "/" + form.Kind + "/" + form.Role
			supported, found := fixtureSupport[key]
			if !found {
				t.Errorf("advertised rewrite form %q has no compatibility fixture", key)
				continue
			}
			if supported != form.Supported {
				t.Errorf("advertised rewrite form %q support=%t, fixture support=%t", key, form.Supported, supported)
			}
			advertised++
		}
	}
	if advertised != len(fixtureSupport) {
		t.Errorf("advertised %d rewrite forms, found %d compatibility fixtures", advertised, len(fixtureSupport))
	}
}

func legacySPLCapabilities() CapabilityManifest {
	manifest := CapabilityManifest{Commands: []Capability{}, Functions: []Capability{}}
	for name, command := range commands {
		manifest.Commands = append(manifest.Commands, Capability{Name: name, SyntaxSupported: true, SemanticSupported: command.handle != nil, Limitations: []string{command.limitation}})
	}
	reviewedCommands := map[string]Capability{
		"append": {
			Name: "append", SyntaxSupported: true, SemanticSupported: false,
			Limitations: []string{"Branch merging is unmodeled.", "Branch merging remains unmodeled; direct child and parent requirements are retained."},
		},
		"appendpipe": {
			Name: "appendpipe", SyntaxSupported: true, SemanticSupported: false,
			Limitations: []string{"Branch merging is unmodeled.", "Branch merging remains unmodeled; direct child and parent requirements are retained."},
		},
		"bin": {
			Name: "bin", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact input and optional alias with supported literal binning options."},
		},
		"bucket": {
			Name: "bucket", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact input and optional alias with supported literal binning options."},
		},
		"fillnull": {
			Name: "fillnull", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact field targets with an optional literal fill value."},
		},
		"join": {
			Name: "join", SyntaxSupported: true, SemanticSupported: false,
			Limitations: []string{"Branch merging is unmodeled.", "Branch merging remains unmodeled; direct child and parent requirements are retained."},
		},
		"macro": {
			Name: "macro", SyntaxSupported: true, SemanticSupported: false,
			Limitations: []string{"Synthetic category for a macro-only stage; exact macro name dependencies, unresolved expansion. A literal command named macro remains unmodeled.", "The exact direct macro requirement is retained, but expansion requirements are unresolved."},
		},
		"mvexpand": {
			Name: "mvexpand", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact field form; row-count and memory effects are outside the static model."},
		},
		"regex": {
			Name: "regex", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact field comparison form."},
		},
		"rex": {
			Name: "rex", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact input fields and unambiguous (?<name>) or (?P<name>) captures."},
		},
		"spath": {
			Name: "spath", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{"Exact input, literal path, and exact output forms."},
		},
		"tstats": {
			Name: "tstats", SyntaxSupported: true, SemanticSupported: true,
			Limitations: []string{
				"Exact aggregates, a literal data-model source, exact WHERE selectors, exact groups, and supported literal options.",
				"Exact inline macro syntax is recognized; expansion remains unresolved.",
				"Macro expansion is unresolved, so the stage and output environment remain incomplete.",
				"The direct macro and sibling requirements are retained, but expansion requirements are unresolved.",
			},
		},
	}
	for i, command := range manifest.Commands {
		if reviewed, ok := reviewedCommands[command.Name]; ok {
			manifest.Commands[i] = reviewed
		}
	}
	for name, function := range functions {
		limit := fmt.Sprintf("%d to %d arguments", function.min, function.max)
		if function.max < 0 {
			limit = fmt.Sprintf("at least %d arguments", function.min)
		}
		if name == "case" {
			limit += " in condition/value pairs"
		}
		if function.aggregate {
			limit += "; aggregate context only"
		} else {
			limit += "; expression context only"
		}
		if function.dynamic {
			limit = "Dynamic query semantics are unresolved."
		}
		if name == "mvindex" {
			limit = "2 or 3 arguments; expression context only"
		}
		manifest.Functions = append(manifest.Functions, Capability{Name: name, SyntaxSupported: true, SemanticSupported: !function.dynamic, Limitations: []string{limit}})
	}
	sort.Slice(manifest.Commands, func(i, j int) bool { return manifest.Commands[i].Name < manifest.Commands[j].Name })
	sort.Slice(manifest.Functions, func(i, j int) bool { return manifest.Functions[i].Name < manifest.Functions[j].Name })
	return manifest
}

func legacySPL2Capabilities() CapabilityManifest {
	manifest := CapabilityManifest{Commands: []Capability{}, Functions: []Capability{}}
	for _, entry := range spl2CommandInventory {
		limitations := append([]string{entry.limitation}, spl2FormLimitations(entry.name)...)
		manifest.Commands = append(manifest.Commands, Capability{Name: entry.name, SyntaxSupported: entry.syntax, SemanticSupported: entry.semantic, Limitations: limitations})
	}
	for name, function := range spl2Functions {
		limit := fmt.Sprintf("%d to %d positional arguments", function.min, function.max)
		if function.max < 0 {
			limit = fmt.Sprintf("at least %d positional arguments", function.min)
		}
		if name == "case" {
			limit += " in condition/value pairs"
		}
		if function.aggregate {
			limit += "; aggregate context only"
		} else {
			limit += "; expression context only"
		}
		limit += "; conditional availability follows expression evidence; named arguments remain incomplete"
		manifest.Functions = append(manifest.Functions, Capability{Name: name, SyntaxSupported: true, SemanticSupported: true, Limitations: append([]string{limit}, spl2FormLimitations("function:"+name)...)})
	}
	sort.Slice(manifest.Commands, func(i, j int) bool { return manifest.Commands[i].Name < manifest.Commands[j].Name })
	sort.Slice(manifest.Functions, func(i, j int) bool { return manifest.Functions[i].Name < manifest.Functions[j].Name })
	return manifest
}

// Catches unsupported functions/forms being advertised or treated as complete.
func TestCapabilitiesFunctionForms(t *testing.T) {
	for _, q := range []string{"eval x=abs(a)", "eval x=round(a,2)", "eval x=ceil(a)", "eval x=ceiling(a)", "eval x=floor(a)", "eval x=len(a)", "eval x=lower(a)", "eval x=upper(a)", "eval x=trim(a)", "eval x=ltrim(a,\"x\")", "eval x=rtrim(a)", "eval x=substr(a,1,2)", "eval x=replace(a,\"x\",\"y\")", "eval x=coalesce(a,b)", "eval x=if(a,b,c)", "eval x=case(a,b,c,d)", "eval x=isnull(a)", "eval x=isnotnull(a)", "eval x=tonumber(a)", "eval x=tostring(a)", "eval x=mvcount(a)", "eval x=split(a,\",\")", "eval x=match(a,\"x\")", "eval x=mvindex(a,0)", "eval x=mvindex(a,0,1)", "eval x=mvfind(a,\"x\")", "eval x=true()", "eval x=null()", "eval x=now()", "eval x=relative_time(_time,\"-1h\")", "eval x=strftime(_time,\"%F\")", "eval x=like(host,\"web%\")", "stats count sum(a) avg(a) min(a) max(a) values(a) list(a) dc(a) distinct_count(a) first(a) last(a) earliest(_time) latest(_time) stdev(duration)"} {
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
	for _, tc := range []struct {
		query  string
		source string
	}{
		{`eval x=mvindex(a)`, `mvindex(a)`},
		{`eval x=mvindex(a,0,1,2)`, `mvindex(a,0,1,2)`},
		{`eval x=mvfind(a)`, `mvfind(a)`},
		{`eval x=true(a)`, `true(a)`},
		{`eval x=null(a)`, `null(a)`},
		{`eval x=now(a)`, `now(a)`},
		{`eval x=relative_time(a)`, `relative_time(a)`},
		{`eval x=strftime(a)`, `strftime(a)`},
		{`eval x=like(a)`, `like(a)`},
		{`eval x=earliest(a)`, `earliest(a)`},
		{`eval x=latest(a)`, `latest(a)`},
		{`eval x=stdev(a)`, `stdev(a)`},
		{`stats like(host,"web%")`, `like(host,"web%")`},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil {
				t.Fatal(err)
			}
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("unsupported function form = status %q coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
			}
			found := false
			for _, diagnostic := range r.Diagnostics {
				if diagnostic.Code != CodeUnsupportedSemantics {
					continue
				}
				if got := tc.query[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset]; got == tc.source {
					found = true
				}
			}
			if !found {
				t.Fatalf("missing source-located %s for %q: %+v", CodeUnsupportedSemantics, tc.source, r.Diagnostics)
			}
		})
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
