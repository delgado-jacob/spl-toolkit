package analysis

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func spl2AnalyzeTest(t *testing.T, text string) *Result {
	t.Helper()
	r, err := Analyze(QueryDocument{Text: text, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestSPL2LowerSelectedCommandTransfers(t *testing.T) {
	t.Run("where eval and fields closure", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | where synthetic_enabled=true | eval synthetic_scaled=synthetic_value*2 | fields synthetic_scaled | where synthetic_scaled>0`)
		if r.Status != Valid || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || len(r.Diagnostics) != 0 {
			t.Fatalf("selected pipeline must be complete: %+v", r)
		}
		created := spl2Ref(t, r, "synthetic_scaled", "create")
		read := spl2Ref(t, r, "synthetic_value", "read")
		if !reflect.DeepEqual(created.OriginReferenceIDs, []string{read.ID}) {
			t.Fatalf("eval origin = %v, want %s", created.OriginReferenceIDs, read.ID)
		}
		last := r.Lineage[len(r.Lineage)-2].After
		if last.Open || len(last.Fields) != 1 || last.Fields[0].Name != "synthetic_scaled" {
			t.Fatalf("fields include must close to exact selectors: %+v", last)
		}
		removed := spl2AnalyzeTest(t, `FROM synthetic_events | fields synthetic_region | where synthetic_value>0`)
		if removed.Status != Invalid || spl2Ref(t, removed, "synthetic_value", "read").Binding != "unavailable" || !spl2HasCode(removed, CodeUnavailableField) {
			t.Fatalf("closed fields projection retained an omitted source: %+v", removed)
		}
	})

	t.Run("stats emits only groups and aggregate aliases", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | stats sum(synthetic_value) AS synthetic_total BY synthetic_region | where synthetic_total>0 AND synthetic_region!=""`)
		if r.Status != Valid || !r.Coverage.SemanticComplete || len(r.Lineage[1].After.Fields) != 2 || r.Lineage[1].After.Open {
			t.Fatalf("selected stats output: %+v", r)
		}
		for _, name := range []string{"synthetic_region", "synthetic_total"} {
			if spl2Ref(t, r, name, "read").Binding == "unavailable" {
				t.Fatalf("stats output %q not installed: %+v", name, r)
			}
		}
		discarded := spl2AnalyzeTest(t, `FROM synthetic_events | stats sum(synthetic_value) AS synthetic_total BY synthetic_region | where synthetic_value>0`)
		unavailable := false
		for _, ref := range discarded.References {
			unavailable = unavailable || ref.NormalizedName == "synthetic_value" && ref.Role == "read" && ref.Binding == "unavailable"
		}
		if discarded.Status != Invalid || !unavailable || !spl2HasCode(discarded, CodeUnavailableField) {
			t.Fatalf("stats retained a discarded aggregate input: %+v", discarded)
		}
	})

	t.Run("selected stats group identity collision", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | stats count() AS synthetic_count BY 'actor.name', actor.name`)
		if r.Status != Incomplete || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete {
			t.Fatalf("colliding selected groups received complete credit: %+v", r)
		}
		groupIDs := []string{}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.Role == "group" && ref.NormalizedName == "actor.name" {
				groupIDs = append(groupIDs, ref.ID)
			}
		}
		if len(groupIDs) != 2 {
			t.Fatalf("colliding group reads = %+v", r.References)
		}
		ambiguities := 0
		for _, diagnostic := range r.Diagnostics {
			if diagnostic.Code == CodeAmbiguousField {
				ambiguities++
			}
		}
		if ambiguities != 1 {
			t.Fatalf("ambiguity diagnostics = %d: %+v", ambiguities, r.Diagnostics)
		}
		state := r.Lineage[len(r.Lineage)-1].After
		bindings := []FieldBinding{}
		for _, field := range state.Fields {
			if field.Name == "actor.name" {
				bindings = append(bindings, field)
			}
		}
		if !state.Uncertain || len(bindings) != 1 || !reflect.DeepEqual(bindings[0].OriginReferenceIDs, groupIDs) {
			t.Fatalf("colliding group state = %+v, want combined origins %v", state, groupIDs)
		}
		for _, transition := range r.Lineage[len(r.Lineage)-1].Transitions {
			if transition.Operation == "project" && transition.Output == "actor.name" {
				t.Fatalf("ambiguous group emitted an identity-specific transition: %+v", transition)
			}
		}
		assertRequirementGap(t, r.Requirements.Gaps, CodeAmbiguousField, groupIDs, []string{CodeAmbiguousField})
	})

	t.Run("selected stats group and aggregate collision", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_dataset | stats count() AS 'actor.name' BY actor.name`)
		if r.Status != Incomplete || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete {
			t.Fatalf("group/output collision received complete credit: %+v", r)
		}
		group := spl2Ref(t, r, "actor.name", "group")
		output := spl2Ref(t, r, "actor.name", "output")
		ambiguities := 0
		for _, diagnostic := range r.Diagnostics {
			if diagnostic.Code == CodeAmbiguousField {
				ambiguities++
				if diagnostic.Category != "unsupported_semantics" {
					t.Fatalf("ambiguity category = %q", diagnostic.Category)
				}
			}
		}
		if ambiguities != 1 {
			t.Fatalf("ambiguity diagnostics = %d: %+v", ambiguities, r.Diagnostics)
		}
		state := r.Lineage[len(r.Lineage)-1].After
		if !state.Uncertain || len(state.Fields) != 1 || state.Fields[0].Name != "actor.name" || !reflect.DeepEqual(state.Fields[0].OriginReferenceIDs, []string{group.ID, output.ID}) {
			t.Fatalf("collided public binding = %+v, want origins %s/%s", state, group.ID, output.ID)
		}
		for _, transition := range r.Lineage[len(r.Lineage)-1].Transitions {
			if transition.Output == "actor.name" {
				t.Fatalf("collided identity emitted an unrepresentable transition: %+v", transition)
			}
		}
		assertRequirementGap(t, r.Requirements.Gaps, CodeAmbiguousField, []string{group.ID, output.ID}, []string{CodeAmbiguousField})
	})

	t.Run("selected stats collision preserves other transition order", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_dataset | stats sum(synthetic_value) AS synthetic_total, count() AS 'actor.name' BY synthetic_region, actor.name`)
		transitions := r.Lineage[len(r.Lineage)-1].Transitions
		if len(transitions) != 2 || transitions[0].Operation != "project" || transitions[0].Output != "synthetic_region" || transitions[1].Operation != "aggregate" || transitions[1].Output != "synthetic_total" {
			t.Fatalf("noncolliding transition order changed: %+v", transitions)
		}
	})

	t.Run("sequential selected stats retain installed outputs", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM [{synthetic_value:1,synthetic_region:"region"}] | stats sum(synthetic_value) AS synthetic_total BY synthetic_region | stats max(synthetic_total) AS synthetic_max BY synthetic_region`)
		if r.Status != Valid || !r.Coverage.SemanticComplete || len(r.Diagnostics) != 0 {
			t.Fatalf("sequential stats lost complete transfer state: %+v", r)
		}
		state := r.Lineage[len(r.Lineage)-1].After
		fields := map[string]bool{}
		for _, field := range state.Fields {
			fields[field.Name] = true
		}
		if state.Open || len(state.Fields) != 2 || !fields["synthetic_region"] || !fields["synthetic_max"] {
			t.Fatalf("sequential stats output = %+v", state)
		}
		transitions := r.Lineage[len(r.Lineage)-1].Transitions
		if len(transitions) != 2 || transitions[0].Operation != "project" || transitions[0].Output != "synthetic_region" || transitions[1].Operation != "aggregate" || transitions[1].Output != "synthetic_max" {
			t.Fatalf("sequential stats transitions = %+v", transitions)
		}
	})

	t.Run("bin installs exact identity", func(t *testing.T) {
		for _, query := range []string{
			`FROM synthetic_events | bin span=5m synthetic_time | where synthetic_time>0`,
			`FROM synthetic_events | bin span=5m synthetic_time AS synthetic_bucket | where synthetic_bucket>0`,
		} {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Valid || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || len(r.Diagnostics) != 0 {
				t.Fatalf("selected bin must be complete: %+v", r)
			}
			input := spl2Ref(t, r, "synthetic_time", "read")
			transition := r.Lineage[1].Transitions[0]
			if transition.Conditional || !reflect.DeepEqual(transition.InputReferenceIDs, []string{input.ID}) {
				t.Fatalf("bin origin transfer: %+v", transition)
			}
		}
	})

	t.Run("bin preserves input availability", func(t *testing.T) {
		conditional := spl2AnalyzeTest(t, `FROM [{synthetic_time:1},{synthetic_other:2}] | bin synthetic_time`)
		input := spl2Ref(t, conditional, "synthetic_time", "read")
		if input.Binding != "indeterminate" {
			t.Fatalf("conditional bin input binding = %+v", input)
		}
		state := conditional.Lineage[len(conditional.Lineage)-1].After
		if len(state.Fields) != 2 {
			t.Fatalf("conditional bin shape = %+v", state)
		}
		foundConditional := false
		for _, field := range state.Fields {
			if field.Name == "synthetic_time" {
				foundConditional = field.Conditional
			}
		}
		if !foundConditional || len(conditional.Lineage[1].Transitions) != 1 || !conditional.Lineage[1].Transitions[0].Conditional {
			t.Fatalf("bin erased conditionality: %+v", conditional.Lineage[1])
		}

		unavailable := spl2AnalyzeTest(t, `FROM [{synthetic_present:1}] | bin synthetic_missing`)
		if ref := spl2Ref(t, unavailable, "synthetic_missing", "read"); ref.Binding != "unavailable" {
			t.Fatalf("closed-source bin input = %+v", ref)
		}
		for _, field := range unavailable.Lineage[len(unavailable.Lineage)-1].After.Fields {
			if field.Name == "synthetic_missing" {
				t.Fatalf("unavailable bin input installed an output: %+v", unavailable)
			}
		}
		for _, transition := range unavailable.Lineage[len(unavailable.Lineage)-1].Transitions {
			if transition.Output == "synthetic_missing" {
				t.Fatalf("unavailable bin input emitted a transition: %+v", transition)
			}
		}
	})

	t.Run("mvexpand preserves identity", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM synthetic_events | mvexpand synthetic_values | where synthetic_values>0`)
		if r.Status != Valid || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || len(r.Diagnostics) != 0 || len(r.Lineage[1].Transitions) != 0 {
			t.Fatalf("selected mvexpand must preserve exact identity: %+v", r)
		}
		refs := []Reference{}
		for _, ref := range r.References {
			if ref.NormalizedName == "synthetic_values" && ref.Role == "read" {
				refs = append(refs, ref)
			}
		}
		if len(refs) != 2 || refs[0].Binding != "source" || refs[1].Binding != "source" || !reflect.DeepEqual(refs[1].OriginReferenceIDs, []string{refs[0].ID}) {
			t.Fatalf("mvexpand identity chain: %+v", refs)
		}

		unavailable := spl2AnalyzeTest(t, `FROM [{synthetic_present:1}] | mvexpand synthetic_missing`)
		if unavailable.Status != Invalid || unavailable.Coverage.SemanticComplete || unavailable.Requirements.Coverage.Complete {
			t.Fatalf("unavailable mvexpand received semantic credit: %+v", unavailable)
		}
		if ref := spl2Ref(t, unavailable, "synthetic_missing", "read"); ref.Binding != "unavailable" {
			t.Fatalf("closed-source mvexpand input = %+v", ref)
		}
		for _, field := range unavailable.Lineage[len(unavailable.Lineage)-1].After.Fields {
			if field.Name == "synthetic_missing" {
				t.Fatalf("unavailable mvexpand installed a field: %+v", unavailable)
			}
		}
		if len(unavailable.Lineage[len(unavailable.Lineage)-1].Transitions) != 0 {
			t.Fatalf("unavailable mvexpand emitted a transition: %+v", unavailable.Lineage[len(unavailable.Lineage)-1].Transitions)
		}
	})

	for _, query := range []string{
		`FROM synthetic_events | bin synthetic_unknown=1 synthetic_time AS synthetic_bucket`,
		`FROM synthetic_events | mvexpand synthetic_unknown=1 synthetic_values`,
	} {
		t.Run("unknown option "+query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, query)
			if r.Status != Incomplete || r.Coverage.SemanticComplete || len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != CodeUnsupportedSemantics {
				t.Fatalf("unknown option boundary: %+v", r)
			}
			for _, ref := range r.References {
				if ref.NormalizedName == "synthetic_unknown" || ref.NormalizedName == "synthetic_bucket" {
					t.Fatalf("unknown option guessed field output: %+v", ref)
				}
			}
			if strings.Contains(query, "bin ") {
				spl2Ref(t, r, "synthetic_time", "read")
			} else {
				spl2Ref(t, r, "synthetic_values", "read")
			}
		})
	}
}

func TestSPL2SelectedCommandDispatchUsesTypedPolicy(t *testing.T) {
	policy, ok := spl2TypedPolicyFor("splunkd", "current")
	if !ok {
		t.Fatal("missing private splunkd/current typed policy")
	}
	want := map[string]spl2CommandHandler{
		"where":    spl2WhereCommandHandler,
		"eval":     spl2EvalCommandHandler,
		"fields":   spl2FieldsCommandHandler,
		"stats":    spl2StatsCommandHandler,
		"bin":      spl2BinCommandHandler,
		"mvexpand": spl2MvexpandCommandHandler,
	}
	if !reflect.DeepEqual(policy.commands, want) {
		t.Fatalf("selected command handlers = %+v, want %+v", policy.commands, want)
	}
	for _, name := range []string{"search", "table", "eventstats", "synthetic_unknown"} {
		if _, selected := policy.commands[name]; selected {
			t.Fatalf("parser-only or neighboring command %q resolved through selected policy", name)
		}
	}

	original := spl2TypedPolicies["splunkd/current"]
	modified := cloneSPL2TypedPolicyForTest(original)
	delete(modified.commands, "where")
	spl2TypedPolicies["splunkd/current"] = modified
	t.Cleanup(func() { spl2TypedPolicies["splunkd/current"] = original })
	r := spl2AnalyzeTest(t, `FROM synthetic_events | where synthetic_value > 0`)
	if r.Status != Incomplete || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
		t.Fatalf("production dispatch ignored selected command policy: %+v", r)
	}
	for _, ref := range r.References {
		if ref.NormalizedName == "synthetic_value" {
			t.Fatalf("unselected command still executed selected transfer: %+v", ref)
		}
	}
}

func TestSPL2RejectedSelectedCommandEffects(t *testing.T) {
	for _, tc := range []struct {
		name, query, input string
	}{
		{"bin suffix option", `FROM main | bin synthetic_time span=15m`, "synthetic_time"},
		{"bin fractional bins", `FROM main | bin bins=2.5 synthetic_value`, "synthetic_value"},
		{"bin malformed alignment", `FROM main | bin span=12h aligntime=@d+ synthetic_time`, ""},
		{"bin unknown option", `FROM main | bin synthetic_unknown=1 synthetic_value`, "synthetic_value"},
		{"mvexpand suffix option", `FROM main | mvexpand synthetic_values limit=2`, "synthetic_values"},
		{"mvexpand unknown option", `FROM main | mvexpand synthetic_unknown=1 synthetic_values`, "synthetic_values"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status == Valid || r.Coverage.SemanticComplete {
				t.Fatalf("rejected command was promoted: %+v", r)
			}
			if tc.input != "" {
				spl2Ref(t, r, tc.input, "read")
			}
			for _, ref := range r.References {
				if ref.Role == "output" {
					t.Fatalf("rejected command guessed output: %+v", ref)
				}
			}
			for _, lineage := range r.Lineage {
				if len(lineage.Transitions) != 0 {
					t.Fatalf("rejected command installed transition: %+v", lineage.Transitions)
				}
			}
		})
	}
}

func TestSPL2RejectedSelectedCommandsPreserveConservativeState(t *testing.T) {
	for _, tc := range []struct {
		name, query, input, absent string
	}{
		{"bin unknown option", `FROM [{synthetic_value:1,synthetic_other:2}] | bin synthetic_unknown=1 synthetic_value AS synthetic_bucket`, "synthetic_value", "synthetic_bucket"},
		{"bin malformed suffix", `FROM [{synthetic_value:1,synthetic_other:2}] | bin synthetic_value span=15m`, "synthetic_value", ""},
		{"mvexpand unknown option", `FROM [{synthetic_values:[1,2],synthetic_other:2}] | mvexpand synthetic_unknown=1 synthetic_values`, "synthetic_values", ""},
		{"mvexpand malformed suffix", `FROM [{synthetic_values:[1,2],synthetic_other:2}] | mvexpand synthetic_values limit=2`, "synthetic_values", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status == Valid || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete {
				t.Fatalf("rejected command received complete credit: %+v", r)
			}
			input := spl2Ref(t, r, tc.input, "read")
			state := r.Lineage[len(r.Lineage)-1].After
			foundInput, foundOther := false, false
			for _, field := range state.Fields {
				switch field.Name {
				case tc.input:
					foundInput = true
					if !field.Conditional {
						t.Fatalf("affected input remained unconditional: %+v", field)
					}
				case "synthetic_other":
					foundOther = true
					if field.Conditional {
						t.Fatalf("unrelated identity became conditional: %+v", field)
					}
				case tc.absent:
					if tc.absent != "" {
						t.Fatalf("rejected command guessed output identity: %+v", field)
					}
				}
			}
			if !foundInput || !foundOther {
				t.Fatalf("rejected command lost known identities: %+v", state)
			}
			if transitions := r.Lineage[len(r.Lineage)-1].Transitions; len(transitions) != 0 {
				t.Fatalf("rejected command emitted transitions: %+v", transitions)
			}
			ownedGap := false
			for _, gap := range r.Requirements.Gaps {
				if gap.Code == CodeUnsupportedSemantics && reflect.DeepEqual(gap.ReferenceIDs, []string{input.ID}) {
					ownedGap = true
				}
			}
			if !ownedGap {
				t.Fatalf("rejected effect gap does not own input %s: %+v", input.ID, r.Requirements.Gaps)
			}
		})
	}
}

func TestSPL2LookupImplicitOutputUncertainty(t *testing.T) {
	for _, pair := range []struct{ match, local string }{{"uid AS user", "user"}, {"id AS account", "account"}} {
		r := spl2AnalyzeTest(t, "FROM main | lookup users "+pair.match+" | table "+pair.local)
		var reads []Reference
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == pair.local {
				reads = append(reads, ref)
			}
		}
		if r.Status != Incomplete || len(reads) != 2 || reads[0].Binding != "source" || reads[1].Binding != "indeterminate" || !reflect.DeepEqual(reads[1].OriginReferenceIDs, []string{reads[0].ID}) {
			t.Fatalf("unknown lookup outputs may overwrite local alias: %+v", r)
		}
		if f := r.Lineage[1].After.Fields; len(f) != 1 || f[0].Name != pair.local || !f[0].Conditional || !reflect.DeepEqual(f[0].OriginReferenceIDs, []string{reads[0].ID}) {
			t.Fatalf("post-lookup field evidence: %+v", r)
		}
	}
	for _, suffix := range []string{"uid | table uid", "uid AS user OUTPUT name | table user", "uid AS user OUTPUTNEW name | table user"} {
		r := spl2AnalyzeTest(t, "FROM main | lookup users "+suffix)
		if r.References[len(r.References)-1].Binding != "source" {
			t.Fatalf("protected key/explicit output changed: %+v", r)
		}
	}
	r := spl2AnalyzeTest(t, "FROM main | eval gone=null | lookup users uid AS user | table gone")
	if r.Status != Incomplete || spl2Ref(t, r, "gone", "read").Binding != "indeterminate" || spl2HasCode(r, CodeUnavailableField) {
		t.Fatalf("unknown output cannot prove absence: %+v", r)
	}
}

func TestSPL2ConflictingAllnumOptions(t *testing.T) {
	for _, command := range []string{"stats", "eventstats"} {
		for _, values := range []string{"true allnum=false", "false allnum=true"} {
			r := spl2AnalyzeTest(t, "FROM main | "+command+" allnum="+values+" count()")
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
				t.Fatalf("conflicting option precedence is unproved: %+v", r)
			}
			located := false
			for _, d := range r.Diagnostics {
				if d.Code == CodeUnsupportedSemantics && strings.HasPrefix(r.Document.Text[d.Location.Start.Offset:d.Location.End.Offset], "allnum=") {
					located = true
				}
			}
			if !located {
				t.Fatalf("missing located conflicting option diagnostic: %+v", r)
			}
		}
		for _, value := range []string{"true", "false"} {
			r := spl2AnalyzeTest(t, "FROM main | "+command+" allnum="+value+" allnum="+value+" count()")
			if r.Status != Valid || !r.Coverage.SemanticComplete {
				t.Fatalf("identical repeated value changed: %+v", r)
			}
		}
	}
}

func TestSPL2RejectedAssignmentEffects(t *testing.T) {
	for _, tc := range []struct {
		expression string
		intact     bool
	}{
		{`@"She said ""yes""`, false},
		{`coalesce(values:primary,backup)`, false},
		{`[[1,2],bytes+1`, false},
		{`map(items,($x) ->)`, false},
		{`round(precision:2,bytes)`, false},
		{`coalesce(values:a,b)`, false},
		{`{a:1,a:2}`, true},
		{`{name:user,"name":host}`, true},
		{`{'display name':user,"display name":role}`, true},
		{`{a:1,a:2}`, true},
		{`{name:user,"name":host}`, true},
	} {
		t.Run(tc.expression, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | eval rejected=`+tc.expression)
			if r.Status != Invalid || r.Coverage.SyntaxComplete != tc.intact || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeSyntaxError) {
				t.Fatalf("rejected assignment coverage: %+v", r)
			}
			for _, ref := range r.References {
				if ref.NormalizedName == "rejected" {
					t.Fatalf("damaged target reference: %+v", ref)
				}
			}
			for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
				if field.Name == "rejected" {
					t.Fatalf("damaged target state: %+v", field)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, expression := range []string{`{a:user,b:host}`, `[[1,2],bytes+1]`, `mystery(bytes)`} {
		r := spl2AnalyzeTest(t, `FROM main | eval kept=`+expression)
		spl2Ref(t, r, "kept", "create")
		if expression == `mystery(bytes)` && !r.Lineage[len(r.Lineage)-1].After.Fields[len(r.Lineage[len(r.Lineage)-1].After.Fields)-1].Conditional {
			t.Fatalf("unknown intact output promoted: %+v", r)
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | eval first=1, rejected={a:user,"a":host}, last=2`)
	spl2Ref(t, r, "first", "create")
	spl2Ref(t, r, "last", "create")
	spl2Ref(t, r, "user", "read")
	spl2Ref(t, r, "host", "read")
	for _, ref := range r.References {
		if ref.NormalizedName == "rejected" {
			t.Fatalf("invalid constructor target survived: %+v", r)
		}
	}
}

func TestSPL2DeferredCommandOriginalRoles(t *testing.T) {
	for _, tc := range []struct {
		query                  string
		reads, groups, outputs []string
	}{
		{`FROM main | rex field=message offset_field=offsets @"(?<digits>\d+)"`, []string{"message"}, nil, []string{"offsets"}},
		{`FROM main | spath input=payload output=user_name path="actor.name"`, []string{"payload"}, nil, []string{"user_name"}},
		{`tstats aggregates=[sum(bytes)] datamodel_name='Traffic.All' predicate=(port=443) byfields=[host,source]`, []string{"bytes", "port"}, []string{"host", "source"}, nil},
		{`mstats aggregates=[avg('cpu.load')] predicate=(index="metrics") byfields=[host]`, []string{"cpu.load"}, []string{"host"}, nil},
		{`FROM main | timechart agg=(sum(bytes)) avg(size) BY host`, []string{"bytes", "size"}, []string{"host"}, nil},
		{`FROM main | timechart agg=(max(bytes) AS peak) count() BY host`, []string{"bytes"}, []string{"host"}, []string{"peak"}},
		{`FROM main | makemv delim=":" labels`, []string{"labels"}, nil, nil},
		{`FROM main | mvcombine delim=";" account`, []string{"account"}, nil, nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Incomplete || !r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
				t.Fatalf("deferred effects changed: %+v", r)
			}
			for _, names := range []struct {
				role  string
				names []string
			}{{"read", tc.reads}, {"group", tc.groups}, {"output", tc.outputs}} {
				for _, name := range names.names {
					ref := spl2Ref(t, r, name, names.role)
					if names.role == "output" && ref.Binding != "not_applicable" {
						t.Fatalf("candidate binding: %+v", ref)
					}
				}
			}
			if strings.HasPrefix(tc.query, "tstats") {
				found := false
				for _, ref := range r.References {
					if ref.Kind == "data_model" && ref.NormalizedName == "Traffic.All" && ref.OriginalName == "'Traffic.All'" {
						found = true
					}
				}
				if !found || !reflect.DeepEqual(r.Dependencies.DataModels, []string{"Traffic.All"}) || len(r.Dependencies.Datasets) != 0 {
					t.Fatalf("atomic model dependency: %+v", r)
				}
			}
			if strings.HasPrefix(tc.query, "mstats") && !reflect.DeepEqual(r.Dependencies.Indexes, []string{"metrics"}) {
				t.Fatalf("typed metrics index selector: %+v", r)
			}
			allowed := map[string]bool{}
			for _, name := range append(append(append([]string{}, tc.reads...), tc.groups...), tc.outputs...) {
				allowed[name] = true
			}
			for _, ref := range r.References {
				if ref.Kind == "field" && !allowed[ref.NormalizedName] {
					t.Fatalf("invented field role %+v", ref)
				}
			}
			for _, name := range tc.outputs {
				for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
					if field.Name == name {
						t.Fatalf("candidate proved output %+v", r)
					}
				}
			}
			for _, lineage := range r.Lineage {
				if len(lineage.Transitions) != 0 {
					t.Fatalf("deferred effect invented transition: %+v", r)
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
}

func TestSPL2DeferredGeneratingInputsDoNotProveOutput(t *testing.T) {
	for _, query := range []string{`tstats aggregates=[sum(bytes)] | table bytes`, `mstats aggregates=[avg(bytes)] | table bytes`, `FROM main | timechart sum(bytes) | table bytes`} {
		r := spl2AnalyzeTest(t, query)
		var refs []Reference
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "bytes" {
				refs = append(refs, ref)
			}
		}
		if len(refs) != 2 || refs[0].Binding != "source" || refs[1].Binding != "indeterminate" {
			t.Fatalf("input presence leaked through generator: %+v", r)
		}
		if !r.Lineage[len(r.Lineage)-2].After.Open || !r.Lineage[len(r.Lineage)-2].After.Uncertain {
			t.Fatalf("unproved generating output closed: %+v", r)
		}
	}
}

func TestSPL2DeferredOverwriteBoundaries(t *testing.T) {
	for _, tc := range []struct {
		command                             string
		firstConditional, secondConditional bool
	}{
		{`rex mode=sed field=first "s/x/y/g"`, true, false},
		{`rex field=first "(?<unknown>.*)"`, true, true},
		{`spath input=first output=second path="name"`, false, true},
		{`spath input=first`, true, true},
		{`makemv delim=":" first`, true, false},
	} {
		r := spl2AnalyzeTest(t, `FROM main | eval first="x", second="y" | `+tc.command)
		if r.Status != Incomplete {
			t.Fatalf("deferred control: %+v", r)
		}
		last := r.Lineage[len(r.Lineage)-1]
		if len(last.After.Fields) != 2 || last.After.Fields[0].Name != "first" || last.After.Fields[0].Conditional != tc.firstConditional || last.After.Fields[1].Name != "second" || last.After.Fields[1].Conditional != tc.secondConditional || len(last.Transitions) != 0 {
			t.Fatalf("overwrite boundary: %+v", r)
		}
		for i, f := range last.Before.Fields {
			if !reflect.DeepEqual(f.OriginReferenceIDs, last.After.Fields[i].OriginReferenceIDs) {
				t.Fatalf("origin changed: %+v", r)
			}
		}
	}
}

func TestSPL2ExternalJobAndJoinIntentions(t *testing.T) {
	for _, sid := range []string{`1780123000.5`, `"1780123000.5"`, `'saved_job'`} {
		t.Run(sid, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "loadjob "+sid)
			if r.Status != Incomplete || len(r.References) != 1 || r.References[0].Kind != "search_job" || r.References[0].Role != "read" || r.References[0].Binding != "not_applicable" || r.References[0].OriginalName != sid || r.References[0].Location.Start.Offset != len("loadjob ") || len(r.Dependencies.Indexes) != 0 || len(r.Lineage[0].After.Fields) != 0 {
				t.Fatalf("static job intention: %+v", r)
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`loadjob "${job}"`, `loadjob "broken`, `loadjob`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "search_job" {
				t.Fatalf("unproved job identity: %+v", r)
			}
		}
	}
	r := spl2AnalyzeTest(t, `FROM main | join left=L right=R where L.id=R.uid [FROM other | table uid]`)
	if r.Status != Incomplete || len(r.Scopes) != 2 {
		t.Fatalf("join scope: %+v", r)
	}
	for _, name := range []string{"id", "uid"} {
		ref := spl2Ref(t, r, name, "read")
		if ref.Binding != "source" || ref.Resolution != "exact" || ref.ScopeID != "scope-0" || len(ref.OriginReferenceIDs) != 0 {
			t.Fatalf("qualified join input: %+v", ref)
		}
	}
	var parentAfter *FieldState
	for i := range r.Lineage {
		if r.Lineage[i].StageID == "stage-1" && r.Lineage[i].ScopeID == "scope-0" {
			parentAfter = &r.Lineage[i].After
		}
	}
	if parentAfter == nil || len(parentAfter.Fields) != 2 || parentAfter.Fields[0].Name != "id" || parentAfter.Fields[1].Name != "uid" || !parentAfter.Uncertain {
		t.Fatalf("join qualified inputs: %+v", parentAfter)
	}
	childUID := Reference{}
	for _, ref := range r.References {
		if ref.OriginalName == "uid" && ref.Role == "read" {
			childUID = ref
		}
	}
	if childUID.ScopeID != "scope-1" {
		t.Fatalf("child scope lost: %+v", r)
	}
	assertCorpusIntegrity(t, r)
}

func TestSPL2IfMergesSelectedBranches(t *testing.T) {
	query := `FROM [{base:1, guard:true}] | if (guard=true) [eval shared=base, left_only=1] else [eval shared=base, right_only=1]`
	r := spl2AnalyzeTest(t, query)
	if r.Status != Valid || !r.Coverage.SemanticComplete {
		t.Fatalf("selected if coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
	}
	final := r.Lineage[len(r.Lineage)-1].After
	want := map[string]bool{"base": false, "guard": false, "shared": false, "left_only": true, "right_only": true}
	if len(final.Fields) != len(want) || !final.Uncertain {
		t.Fatalf("selected if state = %+v", final)
	}
	for _, field := range final.Fields {
		conditional, ok := want[field.Name]
		if !ok || field.Conditional != conditional {
			t.Errorf("selected if field = %+v want conditional=%t", field, conditional)
		}
	}
	guard := spl2Ref(t, r, "guard", "read")
	if guard.ScopeID != "scope-0" || guard.StageID != r.Lineage[len(r.Lineage)-1].StageID {
		t.Errorf("guard ownership = %+v", guard)
	}
	childReads := 0
	for _, ref := range r.References {
		if ref.NormalizedName == "base" && ref.Role == "read" {
			childReads++
			if ref.ScopeID == "scope-0" {
				t.Errorf("branch read escaped child scope: %+v", ref)
			}
		}
	}
	if childReads != 2 || len(r.Scopes) != 3 {
		t.Fatalf("branch-local evidence = reads %d scopes %+v", childReads, r.Scopes)
	}
}

func TestSPL2IfOmittedElseAndNestedConditional(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		field       string
	}{
		{name: "omitted else", query: `FROM [{guard:true}] | if (guard=true) [eval branch_only=1]`, field: "branch_only"},
		{name: "nested", query: `FROM [{outer_flag:true, inner_flag:true}] | if (outer_flag=true) [if (inner_flag=true) [eval nested_only=1] else [eval nested_only=2]]`, field: "nested_only"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Valid || !r.Coverage.SemanticComplete {
				t.Fatalf("coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
			}
			final := r.Lineage[len(r.Lineage)-1].After
			found := false
			for _, field := range final.Fields {
				if field.Name == tc.field {
					found = true
					if !field.Conditional {
						t.Errorf("%s should be conditional after unchanged outer path: %+v", tc.field, field)
					}
				}
			}
			if !found {
				t.Fatalf("missing %s in %+v", tc.field, final)
			}
		})
	}
}

func TestSPL2BranchMergesGuardedArms(t *testing.T) {
	query := `FROM [{status:"ok", value:1}] | branch (status="ok") [eval common=value, ok_only=1], (status!="ok") [eval common=value, failed_only=1]`
	r := spl2AnalyzeTest(t, query)
	if r.Status != Valid || !r.Coverage.SemanticComplete {
		t.Fatalf("guarded branch coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
	}
	final := r.Lineage[len(r.Lineage)-1].After
	want := map[string]bool{"status": false, "value": false, "common": false, "ok_only": true, "failed_only": true}
	for _, field := range final.Fields {
		conditional, ok := want[field.Name]
		if !ok || field.Conditional != conditional {
			t.Errorf("guarded branch field = %+v want conditional=%t", field, conditional)
		}
		delete(want, field.Name)
	}
	if len(want) != 0 || len(r.Scopes) != 3 {
		t.Fatalf("guarded branch state = %+v scopes %+v missing %v", final, r.Scopes, want)
	}
}

func TestSPL2BranchRejectsUnguardedLayout(t *testing.T) {
	r := spl2AnalyzeTest(t, `FROM [{status:"ok"}] | branch [eval leaked=1], (status="ok") [eval selected=1]`)
	if r.Status == Valid || r.Coverage.SemanticComplete || !spl2HasCode(r, CodeUnsupportedSemantics) {
		t.Fatalf("unguarded branch received selected semantics: %+v", r)
	}
	for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
		if field.Name == "leaked" || field.Name == "selected" {
			t.Fatalf("unguarded branch installed child output: %+v", r)
		}
	}
}

func TestSPL2UnionMergesExactDatasetsAndIndependentQueries(t *testing.T) {
	query := `union synthetic_left, $synthetic_right, [FROM [{shared:1, child_only:1}]]`
	r := spl2AnalyzeTest(t, query)
	if r.Status != Valid || !r.Coverage.SemanticComplete {
		t.Fatalf("selected union coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
	}
	if !reflect.DeepEqual(r.Dependencies.Datasets, []string{"$synthetic_right", "synthetic_left"}) {
		t.Fatalf("union dependencies = %v", r.Dependencies.Datasets)
	}
	final := r.Lineage[len(r.Lineage)-1].After
	if !final.Open || !final.Uncertain {
		t.Fatalf("union state = %+v", final)
	}
	for _, name := range []string{"shared", "child_only"} {
		found := false
		for _, field := range final.Fields {
			if field.Name == name {
				found = true
				if !field.Conditional {
					t.Errorf("union child field should be conditional: %+v", field)
				}
			}
		}
		if !found {
			t.Errorf("union output missing %s: %+v", name, final)
		}
	}
	if len(r.Scopes) != 2 || r.Scopes[1].Kind != "search" {
		t.Fatalf("union child scope = %+v", r.Scopes)
	}
}

func TestSPL2UnionDynamicOperandRetainsSiblingEvidence(t *testing.T) {
	query := `union {kind:lower("synthetic")}, [FROM [{proved:1}]]`
	r := spl2AnalyzeTest(t, query)
	if r.Status != Incomplete || !spl2HasCode(r, CodeDynamicReference) {
		t.Fatalf("dynamic union boundary = status %s diagnostics %+v", r.Status, r.Diagnostics)
	}
	proved := false
	for _, field := range r.Lineage[len(r.Lineage)-1].After.Fields {
		proved = proved || field.Name == "proved"
	}
	if !proved || len(r.Scopes) != 2 {
		t.Fatalf("dynamic union erased sibling evidence: %+v", r)
	}
}

func TestSPL2UnionResolvesLocalViewOperand(t *testing.T) {
	r := spl2ProgramAnalyze(t, `$out = union $base, [FROM [{b:1}]]; $base = FROM [{a:1}];`)
	if r.Status != Valid || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete {
		t.Fatalf("local-view union = %+v", r)
	}
	for _, ref := range r.References {
		if ref.Kind == "dataset" && ref.NormalizedName == "$base" {
			t.Fatalf("local view leaked as dataset requirement: %+v", ref)
		}
	}
	if spl2ProgramContains(r.Dependencies.Datasets, "$base") {
		t.Fatalf("local view leaked as dependency: %+v", r.Dependencies)
	}
	foundView := false
	for _, ref := range r.References {
		foundView = foundView || ref.Kind == "view" && ref.NormalizedName == "$base" && ref.Role == "read"
	}
	if !foundView {
		t.Fatalf("local view reference missing: %+v", r.References)
	}
	unionStageID := ""
	for _, stage := range r.Stages {
		if stage.Command == "union" {
			unionStageID = stage.ID
		}
	}
	unionFields := map[string]bool{}
	for _, lineage := range r.Lineage {
		if lineage.StageID != unionStageID {
			continue
		}
		for _, field := range lineage.After.Fields {
			unionFields[field.Name] = field.Conditional
		}
	}
	if !reflect.DeepEqual(unionFields, map[string]bool{"a": true, "b": true}) {
		t.Fatalf("local-view union fields = %+v", unionFields)
	}
}

func TestSPL2UnionMergesOverlappingAndDisjointIndependentFields(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		wantScopes  int
	}{
		{name: "generating", query: `union [FROM [{shared:1, left_only:1}]], [FROM [{shared:2, right_only:1}]]`, wantScopes: 3},
		{name: "pipelined", query: `FROM [{shared:1, left_only:1}] | union [FROM [{shared:2, right_only:1}]]`, wantScopes: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Valid || !r.Coverage.SemanticComplete {
				t.Fatalf("independent union coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
			}
			final := r.Lineage[len(r.Lineage)-1].After
			want := map[string]bool{"shared": false, "left_only": true, "right_only": true}
			if len(final.Fields) != len(want) || !final.Uncertain {
				t.Fatalf("independent union state = %+v", final)
			}
			for _, field := range final.Fields {
				conditional, ok := want[field.Name]
				if !ok || field.Conditional != conditional {
					t.Errorf("independent union field = %+v want conditional=%t", field, conditional)
				}
				delete(want, field.Name)
			}
			if len(want) != 0 || len(r.Scopes) != tc.wantScopes {
				t.Fatalf("independent union missing fields %v scopes %+v", want, r.Scopes)
			}
		})
	}
}

func TestSPL2PipelineJoinSelectedTypes(t *testing.T) {
	for _, tc := range []struct {
		joinType         string
		leftConditional  bool
		rightConditional bool
	}{
		{joinType: "inner"},
		{joinType: "left", rightConditional: true},
		{joinType: "outer", leftConditional: true, rightConditional: true},
	} {
		t.Run(tc.joinType, func(t *testing.T) {
			query := `FROM [{left_key:1, left_value:2}] | join type=` + tc.joinType + ` left=L right=R where L.left_key=R.right_key [FROM [{right_key:1, right_value:3}]]`
			r := spl2AnalyzeTest(t, query)
			if r.Status != Valid || !r.Coverage.SemanticComplete {
				t.Fatalf("join coverage = status %s coverage %+v diagnostics %+v", r.Status, r.Coverage, r.Diagnostics)
			}
			want := map[string]bool{
				"left_key": tc.leftConditional, "left_value": tc.leftConditional,
				"right_key": tc.rightConditional, "right_value": tc.rightConditional,
			}
			final := r.Lineage[len(r.Lineage)-1].After
			if len(final.Fields) != len(want) || final.Uncertain {
				t.Fatalf("join state = %+v", final)
			}
			for _, field := range final.Fields {
				conditional, ok := want[field.Name]
				if !ok || field.Conditional != conditional {
					t.Errorf("join field = %+v want conditional=%t", field, conditional)
				}
			}
			for _, identity := range []string{"left_key", "right_key"} {
				ref := spl2Ref(t, r, identity, "read")
				if ref.Binding == "unavailable" || ref.Resolution != "exact" || len(ref.OriginReferenceIDs) == 0 {
					t.Errorf("join key %s = %+v", identity, ref)
				}
			}
			if len(r.Scopes) != 2 || r.Scopes[1].Kind != "search" {
				t.Fatalf("join right scope = %+v", r.Scopes)
			}
		})
	}
}

func TestSPL2PipelineJoinRejectsUnselectedLayouts(t *testing.T) {
	for _, query := range []string{
		`FROM [{id:1}] | join left=L right=R where L.id=R.id [FROM [{id:1}]]`,
		`FROM [{id:1}] | join type=left left=L where L.id=R.id [FROM [{id:1}]]`,
		`FROM [{id:1}] | join type=right left=L right=R where L.id=R.id [FROM [{id:1}]]`,
		`FROM [{id:1}] | join type=inner left=L right=R max=2 where L.id=R.id [FROM [{id:1}]]`,
	} {
		r := spl2AnalyzeTest(t, query)
		if r.Status == Valid || r.Coverage.SemanticComplete {
			t.Errorf("unselected join layout became complete for %q: %+v", query, r)
		}
	}
}

func TestSPL2PipelineJoinRejectsUnresolvedKeysAndOutputCollisions(t *testing.T) {
	unavailable := spl2AnalyzeTest(t, `FROM [{left_key:1}] | join type=inner left=L right=R where L.missing=R.right_key [FROM [{right_key:1}]]`)
	if unavailable.Status != Invalid || !spl2HasCode(unavailable, CodeUnavailableField) {
		t.Fatalf("unavailable join key = %+v", unavailable)
	}

	unresolved := spl2AnalyzeTest(t, `FROM [{left_key:1}] | join type=inner left=L right=R where L.left_key=X.right_key [FROM [{right_key:1}]]`)
	if unresolved.Status == Valid || unresolved.Coverage.SemanticComplete || !spl2HasCode(unresolved, CodeUnsupportedSemantics) {
		t.Fatalf("unresolved join qualifier = %+v", unresolved)
	}

	collision := spl2AnalyzeTest(t, `FROM [{id:1, left_value:2}] | join type=inner left=L right=R where L.id=R.id [FROM [{id:1, right_value:3}]]`)
	if collision.Status != Incomplete || collision.Coverage.SemanticComplete || !spl2HasCode(collision, CodeAmbiguousField) {
		t.Fatalf("join output collision = %+v", collision)
	}
	state := collision.Lineage[len(collision.Lineage)-1].After
	fields := map[string]FieldBinding{}
	for _, field := range state.Fields {
		fields[field.Name] = field
	}
	id := fields["id"]
	if !state.Uncertain || len(fields) != 3 || len(id.OriginReferenceIDs) != 2 || fields["right_value"].Name == "" {
		t.Fatalf("join collision state = %+v", state)
	}
	assertRequirementGap(t, collision.Requirements.Gaps, CodeAmbiguousField, id.OriginReferenceIDs, []string{CodeAmbiguousField})
}

func TestSPL2PipelineJoinOpenSourcesInstallSideOwnedKeys(t *testing.T) {
	query := `FROM main | join type=left left=L right=R where L.id=R.uid [FROM other] | where uid>0`
	r := spl2AnalyzeTest(t, query)
	var joinAfter *FieldState
	for i := range r.Lineage {
		stage := r.Lineage[i]
		if stage.ScopeID == "scope-0" && r.Stages[1].ID == stage.StageID {
			joinAfter = &stage.After
		}
	}
	if joinAfter == nil {
		t.Fatalf("missing join lineage: %+v", r.Lineage)
	}
	uidConditional := false
	for _, field := range joinAfter.Fields {
		uidConditional = uidConditional || field.Name == "uid" && field.Conditional
	}
	if !uidConditional {
		t.Fatalf("right key did not become a conditional left-join output: %+v", joinAfter)
	}
	downstream := Reference{}
	for _, ref := range r.References {
		if ref.OriginalName == "uid" && ref.Role == "read" {
			downstream = ref
		}
	}
	if downstream.Binding != "indeterminate" || len(downstream.OriginReferenceIDs) == 0 {
		t.Fatalf("downstream right key = %+v", downstream)
	}

	collision := spl2AnalyzeTest(t, `FROM main | join type=inner left=L right=R where L.id=R.id [FROM other]`)
	if collision.Status != Incomplete || !spl2HasCode(collision, CodeAmbiguousField) {
		t.Fatalf("open-source join collision = %+v", collision)
	}
}

func TestSPL2PipelineJoinDoesNotReownInputAmbiguity(t *testing.T) {
	query := `FROM main | where actor.name="x" AND 'actor.name'="x" | join type=inner left=L right=R where L.id=R.uid [FROM other]`
	r := spl2AnalyzeTest(t, query)
	ambiguities := 0
	for _, diagnostic := range r.Diagnostics {
		if diagnostic.Code != CodeAmbiguousField {
			continue
		}
		ambiguities++
		if diagnostic.Location.Start.Offset >= strings.Index(query, " | join") {
			t.Errorf("join re-owned input ambiguity: %+v", diagnostic)
		}
	}
	gaps := 0
	for _, gap := range r.Requirements.Gaps {
		if gap.Code == CodeAmbiguousField {
			gaps++
		}
	}
	if ambiguities != 1 || gaps != 1 {
		t.Fatalf("input ambiguity cardinality = diagnostics %d gaps %d: %+v", ambiguities, gaps, r)
	}
}

func TestSPL2SelectedFlowRejectsDeferredHeldParserDiagnostic(t *testing.T) {
	r := spl2ProgramAnalyze(t, `$out = FROM [{left_id:1}] | join type=INNER left=L right=R where L.left_id=R.right_id [FROM [{right_id:1, right_value:2}]];`)
	if r.Status == Valid || r.Coverage.SemanticComplete {
		t.Fatalf("held parser diagnostic received selected join effects: %+v", r)
	}
	joinStageID := ""
	for _, stage := range r.Stages {
		if stage.Command == "join" {
			joinStageID = stage.ID
		}
	}
	for _, lineage := range r.Lineage {
		if lineage.StageID != joinStageID {
			continue
		}
		for _, field := range lineage.After.Fields {
			if field.Name == "right_id" || field.Name == "right_value" {
				t.Fatalf("held join installed right output in parent lineage: %+v", lineage)
			}
		}
	}
}

func TestSPL2SelectedAlternativeMergesOwnPrivateIdentityCollisions(t *testing.T) {
	for _, tc := range []struct {
		name, query string
	}{
		{name: "if", query: `FROM main | if (guard=true) [where actor.name="x"] else [where 'actor.name'="x"]`},
		{name: "branch", query: `FROM main | branch (guard=true) [where actor.name="x"], (guard=false) [where 'actor.name'="x"]`},
		{name: "union", query: `union [FROM main | where actor.name="x"], [FROM other | where 'actor.name'="x"]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			ambiguities := 0
			for _, diagnostic := range r.Diagnostics {
				if diagnostic.Code == CodeAmbiguousField {
					ambiguities++
				}
			}
			if r.Status != Incomplete || r.Coverage.SemanticComplete || r.Requirements.Coverage.Complete || ambiguities != 1 {
				t.Fatalf("selected merge collision = %+v", r)
			}
			state := r.Lineage[len(r.Lineage)-1].After
			actorBindings := []FieldBinding{}
			for _, field := range state.Fields {
				if field.Name == "actor.name" {
					actorBindings = append(actorBindings, field)
				}
			}
			if !state.Uncertain || len(actorBindings) != 1 || len(actorBindings[0].OriginReferenceIDs) != 2 {
				t.Fatalf("selected merge collision state = %+v", state)
			}
			assertRequirementGap(t, r.Requirements.Gaps, CodeAmbiguousField, actorBindings[0].OriginReferenceIDs, []string{CodeAmbiguousField})
		})
	}
}

func TestSPL2SelectedFlowForwardViewAlternatives(t *testing.T) {
	for _, tc := range []struct {
		name, command, query string
	}{
		{
			name:    "if",
			command: "if",
			query: `$out = FROM synthetic_seed | if (guard=true) [FROM $left] else [FROM $right];
$left = FROM synthetic_left | fields a;
$right = FROM synthetic_right | fields b;`,
		},
		{
			name:    "branch",
			command: "branch",
			query: `$out = FROM synthetic_seed | branch (guard=true) [FROM $left], (guard=false) [FROM $right];
$left = FROM synthetic_left | fields a;
$right = FROM synthetic_right | fields b;`,
		},
		{
			name:    "union",
			command: "union",
			query: `$out = union $left, $right;
$left = FROM synthetic_left | fields a;
$right = FROM synthetic_right | fields b;`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			trace.assertReferences(result.References)
			if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
				t.Fatalf("forward-view %s = %+v", tc.command, result)
			}
			stageID := ""
			for _, stage := range result.Stages {
				if stage.Command == tc.command {
					stageID = stage.ID
					break
				}
			}
			var state *FieldState
			for i := range result.Lineage {
				if result.Lineage[i].StageID == stageID {
					state = &result.Lineage[i].After
				}
			}
			if state == nil || state.Open || len(state.Fields) != 2 {
				t.Fatalf("forward-view %s state = %+v", tc.command, state)
			}
			fields := map[string]FieldBinding{}
			for _, field := range state.Fields {
				fields[field.Name] = field
			}
			for _, name := range []string{"a", "b"} {
				field, ok := fields[name]
				if !ok || !field.Conditional || len(field.OriginReferenceIDs) != 1 {
					t.Errorf("forward-view %s field %s = %+v", tc.command, name, field)
				}
			}
			for _, dataset := range []string{"synthetic_left", "synthetic_right"} {
				item := requirementItem(result.Requirements, "dataset", dataset, "read")
				if item == nil || item.Necessity != "required" || len(item.Occurrences) != 1 {
					t.Errorf("forward-view %s dataset %s = %+v", tc.command, dataset, item)
				}
			}
			if tc.command != "union" {
				guard := requirementItem(result.Requirements, "field", "guard", "read")
				if guard == nil || guard.Necessity != "required" {
					t.Errorf("forward-view %s guard = %+v", tc.command, guard)
				}
			}
		})
	}
}

func TestSPL2PipelineJoinForwardViewsKeepTraceSuffixOnce(t *testing.T) {
	for _, tc := range []struct {
		name, options string
		valid         bool
	}{
		{name: "selected", options: "type=inner ", valid: true},
		{name: "rejected", options: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query := `$out = FROM $left | join ` + tc.options + `left=L right=R where L.id=R.uid [FROM $right];
$left = FROM synthetic_left | fields id;
$right = FROM synthetic_right | fields uid, name;`
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			trace.assertReferences(result.References)
			if tc.valid && (result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete) {
				t.Fatalf("forward join = %+v", result)
			}
			if !tc.valid && (result.Status == Valid || result.Coverage.SemanticComplete) {
				t.Fatalf("rejected forward join was promoted: %+v", result)
			}
			for _, dataset := range []string{"synthetic_left", "synthetic_right"} {
				refs := 0
				for _, ref := range result.References {
					if ref.Kind == "dataset" && ref.Role == "read" && ref.NormalizedName == dataset {
						refs++
					}
				}
				if refs != 1 {
					t.Errorf("forward join dataset %s references = %d", dataset, refs)
				}
			}
			for _, identity := range []string{"synthetic_left", "synthetic_right"} {
				item := requirementItem(result.Requirements, "dataset", identity, "read")
				if item == nil || item.Necessity != "required" || len(item.Occurrences) != 1 {
					t.Errorf("forward join dataset %s = %+v", identity, item)
				}
			}
		})
	}
}

func TestSPL2SelectedFlowRecoveryKeepsAlternativeRequirementsConditional(t *testing.T) {
	for _, tc := range []struct {
		name, query string
		recovered   []string
		lost        []string
	}{
		{
			name:      "if",
			query:     `FROM main | if (guard=true) [where left>0] elseif (oops=) [where bad>0] else [where right>0]`,
			recovered: []string{"left"},
			lost:      []string{"right"},
		},
		{
			name:      "branch",
			query:     `FROM main | branch (guard=true) [where left>0], (oops=) [where bad>0], (guard=false) [where right>0]`,
			recovered: []string{"left", "right"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query, Language: "spl2"}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			trace.assertReferences(result.References)
			if result.Status == Valid || result.Coverage.SyntaxComplete || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete {
				t.Fatalf("recovered selected flow received complete credit: %+v", result)
			}
			guard := requirementItem(result.Requirements, "field", "guard", "read")
			if guard == nil || guard.Necessity != "required" {
				t.Errorf("recovered %s guard = %+v", tc.name, guard)
			}
			for _, identity := range tc.recovered {
				item := requirementItem(result.Requirements, "field", identity, "read")
				if item == nil || item.Necessity != "conditional" {
					t.Errorf("recovered %s alternative %s = %+v", tc.name, identity, item)
				}
			}
			for _, identity := range tc.lost {
				if item := requirementItem(result.Requirements, "field", identity, "read"); item != nil {
					t.Errorf("unrepresented %s alternative %s was invented: %+v", tc.name, identity, item)
				}
			}
			for _, ref := range result.References {
				if (ref.NormalizedName == "left" || ref.NormalizedName == "right") && ref.ScopeID == "scope-0" {
					t.Errorf("recovered child read escaped child scope: %+v", ref)
				}
			}
			gap := false
			for _, candidate := range result.Requirements.Gaps {
				gap = gap || candidate.Code == CodeUnsupportedSemantics
			}
			if !gap {
				t.Errorf("recovered %s lost-arm evidence has no command gap: %+v", tc.name, result.Requirements.Gaps)
			}
		})
	}
}

func TestSPL2MetricsSelectorTypedPosition(t *testing.T) {
	for _, predicate := range []string{`index="metrics"`, `index=metrics`, `index="metrics" AND port=443`} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(`+predicate+`) byfields=[index]`)
		if !reflect.DeepEqual(r.Dependencies.Indexes, []string{"metrics"}) {
			t.Fatalf("selector lost: %+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && (ref.NormalizedName == "metrics" || (ref.NormalizedName == "index" && ref.Role != "group")) {
				t.Fatalf("selector became input: %+v", ref)
			}
		}
		spl2Ref(t, r, "index", "group")
	}
	for _, predicate := range []string{`'index'="metrics"`, `index=lower("metrics")`, `index="metrics*"`, `index=actor.name`, `index="${catalog}"`} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(`+predicate+`)`)
		if len(r.Dependencies.Indexes) != 0 || r.Status != Incomplete {
			t.Fatalf("unproved exact source membership: %+v", r)
		}
	}
	for _, query := range []string{`FROM main | rex "plain"`, `FROM main | spath`} {
		r := spl2AnalyzeTest(t, query)
		for _, ref := range r.References {
			if ref.NormalizedName == "_raw" {
				t.Fatalf("implicit span fabricated: %+v", r)
			}
		}
	}
}

func TestSPL2UnresolvedMetricsSelectorKeepsItsRole(t *testing.T) {
	for _, tc := range []struct {
		rhs, field string
	}{
		{rhs: `"${catalog}"`, field: "catalog"},
		{rhs: `lower(catalog)`, field: "catalog"},
		{rhs: `catalog.name`, field: "catalog.name"},
	} {
		r := spl2AnalyzeTest(t, `mstats aggregates=[count()] predicate=(index=`+tc.rhs+`)`)
		if r.Status != Incomplete || len(r.Dependencies.Indexes) != 0 {
			t.Fatalf("unresolved selector promoted: %+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "index" {
				t.Fatalf("selector label became field: %+v", r)
			}
		}
		ref := spl2Ref(t, r, tc.field, "read")
		if tc.field == "catalog.name" && (ref.Resolution != "exact" || ref.Binding != "source") {
			t.Fatalf("structural selector operand: %+v", ref)
		}
	}
}

func TestSPL2RecoveredInputAndSelectorSoundness(t *testing.T) {
	for _, tail := range []string{`severity=`, `status IN (401,,403)`, `"unterminated`, `a=1 AND`, `earliest=`, `_index_latest=-`} {
		t.Run(tail, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `search index=app `+tail)
			if r.Status != Invalid || !reflect.DeepEqual(r.Dependencies.Indexes, []string{"app"}) {
				t.Fatalf("intact index lost: %+v", r)
			}
			for _, ref := range r.References {
				if ref.Kind == "field" && (ref.NormalizedName == "index" || ref.NormalizedName == "_index_latest" || ref.NormalizedName == "earliest") {
					t.Fatalf("modifier label became field: %+v", r)
				}
			}
		})
	}
	for _, command := range []string{`table server account`, `eval x=1 y=2`, `reverse a`, `head null=true`, `stats count() sum(bytes) AS total`, `sort -host +num(bytes)`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command)
			if r.Status != Invalid || r.Stages[1].SemanticComplete {
				t.Fatalf("malformed owner promoted: %+v", r)
			}
		})
	}
	for _, command := range []string{`stats`, `eventstats`, `streamstats`} {
		t.Run(command+" alias", func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command+` count() as`)
			for _, ref := range r.References {
				if ref.Kind == "field" && ref.NormalizedName == "count" {
					t.Fatalf("damaged alias fell back: %+v", r)
				}
			}
		})
	}
	for _, command := range []string{`table "host"`, `table "host${suffix}"`, `table "${if(flag,"*","host")}"`, `fields host*`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, `FROM main | `+command)
			for _, ref := range r.References {
				if ref.NormalizedName == "host${suffix}" || ref.OriginalName == `"host"` {
					t.Fatalf("unproved static selector: %+v", r)
				}
			}
			for _, l := range r.Lineage {
				for _, f := range l.After.Fields {
					if f.Name == "" {
						t.Fatalf("empty field: %+v", r)
					}
				}
			}
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`search index=app earliest=unproved`, `search index=app _index_latest=-`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "field" && (ref.NormalizedName == "earliest" || ref.NormalizedName == "_index_latest") {
				t.Fatalf("time modifier is not field: %+v", r)
			}
		}
	}
}
func spl2HasCode(r *Result, code string) bool {
	for _, d := range r.Diagnostics {
		if d.Code == code {
			return true
		}
	}
	return false
}
func spl2Ref(t *testing.T, r *Result, name, role string) Reference {
	t.Helper()
	for _, ref := range r.References {
		if ref.NormalizedName == name && ref.Role == role {
			return ref
		}
	}
	t.Fatalf("missing %s %s in %+v", role, name, r.References)
	return Reference{}
}
func TestSPL2SequentialFullState(t *testing.T) {
	text := "FROM main | eval x=bytes, y=x+1 | table y"
	got := spl2AnalyzeTest(t, text)
	loc := func(start, end int) Location {
		return Location{Position{start, 1, start + 1}, Position{end, 1, end + 1}}
	}
	ref := func(id, name, role, stage, bind string, start, end int, origins ...string) Reference {
		kind := "field"
		if name == "main" {
			kind = "dataset"
		}
		return Reference{ID: id, OriginalName: text[start:end], NormalizedName: name, Kind: kind, Role: role, StageID: stage, ScopeID: "scope-0", Location: loc(start, end), Resolution: "exact", Binding: bind, OriginReferenceIDs: append([]string{}, origins...)}
	}
	field := func(name string, ids ...string) FieldBinding {
		return FieldBinding{Name: name, OriginReferenceIDs: ids}
	}
	before := FieldState{Fields: []FieldBinding{}, Removed: []string{}, Open: true}
	assigned := FieldState{Fields: []FieldBinding{field("bytes", "ref-2"), field("x", "ref-1", "ref-2"), field("y", "ref-3", "ref-4", "ref-1", "ref-2")}, Removed: []string{}, Open: true}
	after := FieldState{Fields: []FieldBinding{field("y", "ref-3", "ref-4", "ref-1", "ref-2")}, Removed: []string{}}
	want := newResult(QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current"})
	want.Status = Valid
	want.Scopes = []Scope{{ID: "scope-0", Kind: "root", Location: loc(0, len(text))}}
	want.Stages = []Stage{{"stage-0", "from", 0, "scope-0", loc(0, 9), true}, {"stage-1", "eval", 1, "scope-0", loc(12, 31), true}, {"stage-2", "table", 2, "scope-0", loc(34, 41), true}}
	want.Dependencies.Datasets = []string{"main"}
	want.References = []Reference{ref("ref-0", "main", "read", "stage-0", "not_applicable", 5, 9), ref("ref-1", "x", "create", "stage-1", "not_applicable", 17, 18, "ref-2"), ref("ref-2", "bytes", "read", "stage-1", "source", 19, 24), ref("ref-3", "y", "create", "stage-1", "not_applicable", 26, 27, "ref-4", "ref-1", "ref-2"), ref("ref-4", "x", "read", "stage-1", "derived", 28, 29, "ref-1", "ref-2"), ref("ref-5", "y", "read", "stage-2", "derived", 40, 41, "ref-3", "ref-4", "ref-1", "ref-2")}
	want.Lineage = []Lineage{{StageID: "stage-0", ScopeID: "scope-0", Before: before, After: before, Transitions: []Transition{}}, {StageID: "stage-1", ScopeID: "scope-0", Before: before, After: assigned, Transitions: []Transition{{"create", "x", []string{"ref-2"}, "ref-1", false}, {"create", "y", []string{"ref-4"}, "ref-3", false}}}, {StageID: "stage-2", ScopeID: "scope-0", Before: assigned, After: after, Transitions: []Transition{{"project", "y", []string{"ref-5"}, "", false}}}}
	want.Requirements = RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			Language: "spl2", Profile: "splunkd", Version: "current",
			QueryDigest: "sha256:159f5b7d4683ac55dd0efac524ba3ea30133c495a91e81af6c695b70d1e16afa",
		},
		CapabilityRevision: "sha256:69b166318f99909d0ffbad378f0369fd9377a1f56945e2c3c0b69eaa32c03e95",
		QueryStatus:        Valid,
		Coverage:           RequirementCoverage{Complete: true, Reasons: []string{}},
		Items: []RequirementItem{
			{ID: "req-1", Kind: "dataset", Identity: "main", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-0", OriginalName: "main", Binding: "not_applicable", StageID: "stage-0", ScopeID: "scope-0", Location: loc(5, 9)}}},
			{ID: "req-2", Kind: "field", Identity: "bytes", Role: "read", Necessity: "required", Origin: "direct", Resolution: "exact", Occurrences: []RequirementOccurrence{{ReferenceID: "ref-2", OriginalName: "bytes", Binding: "source", StageID: "stage-1", ScopeID: "scope-0", Location: loc(19, 24)}}},
		},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
	}
	if !reflect.DeepEqual(got, want) {
		g, _ := json.MarshalIndent(got, "", "  ")
		w, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("got %s\nwant %s", g, w)
	}
	for _, u := range []SourceUniverse{{Fields: []string{"bytes", "_time"}, Complete: true}, {Fields: []string{"bytes"}}} {
		s, e := AnalyzeWithSourceUniverse(want.Document, u)
		if e != nil || !reflect.DeepEqual(s, &SourceAnalysis{Result: want, Expansions: []FieldExpansion{}}) {
			t.Fatalf("source report %+v %v", s, e)
		}
	}
}
func TestSPL2OrdinaryReadAndTransferBoundaries(t *testing.T) {
	t.Run("search literals", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `search index=main code=bytes status IN (true,false) flag="-1"`)
		if r.Status != Valid || !reflect.DeepEqual(r.Dependencies.Indexes, []string{"main"}) {
			t.Fatalf("%+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName != "code" && ref.NormalizedName != "status" && ref.NormalizedName != "flag" {
				t.Fatalf("literal read %+v", ref)
			}
		}
	})
	t.Run("where reads", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | where bytes=rate`)
		if r.Status != Valid {
			t.Fatalf("%+v", r)
		}
		spl2Ref(t, r, "bytes", "read")
		spl2Ref(t, r, "rate", "read")
	})
	t.Run("null removes", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval x=bytes, x=null | where x>0`)
		if r.Status != Invalid || spl2Ref(t, r, "x", "remove").Binding != "not_applicable" || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
			t.Fatalf("%+v", r)
		}
		last := r.Lineage[1].Transitions[1]
		if last.Operation != "remove" || len(last.InputReferenceIDs) != 0 {
			t.Fatalf("%+v", last)
		}
	})
	t.Run("fields open", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval x=bytes, y=x+1 | fields y`)
		if r.Status != Valid || !r.Coverage.SemanticComplete || !r.Requirements.Coverage.Complete || spl2Ref(t, r, "x", "read").Binding != "derived" {
			t.Fatalf("%+v", r)
		}
		last := r.Lineage[len(r.Lineage)-1].After
		if last.Open || last.Uncertain || len(last.Fields) != 1 || last.Fields[0].Name != "y" {
			t.Fatalf("fields closure: %+v", last)
		}
	})
	t.Run("internal removal", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields - _time | eval y=bytes | fields y`, Language: "spl2"}, SourceUniverse{Fields: []string{"bytes", "_time", "_raw"}, Complete: true})
		if e != nil {
			t.Fatal(e)
		}
		last := r.Result.Lineage[len(r.Result.Lineage)-1].After
		if r.Result.Status != Valid || len(last.Fields) != 1 || last.Fields[0].Name != "y" || !reflect.DeepEqual(last.Removed, []string{"_time"}) {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("independent rename", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | rename a AS x, b AS y | table x,y`)
		if r.Status != Valid || spl2Ref(t, r, "x", "read").Binding != "derived" {
			t.Fatalf("%+v", r)
		}
	})
	for _, tail := range []string{"a AS x, a AS y", "a AS x, b AS x", "a AS b, b AS c", "a AS b, b AS a"} {
		t.Run(tail, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | rename "+tail)
			if r.Status != Invalid {
				t.Fatalf("%+v", r)
			}
		})
	}
	t.Run("aggregate output", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | stats count(), sum(bytes), dc(action) BY host`)
		if r.Status != Valid {
			t.Fatalf("%+v", r)
		}
		if r.Lineage[1].After.Open || len(r.Lineage[1].After.Fields) != 4 {
			t.Fatalf("%+v", r.Lineage)
		}
		for _, name := range []string{"count", "sum(bytes)", "dc(action)"} {
			spl2Ref(t, r, name, "output")
		}
	})
	t.Run("lookup local", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | eval role="local" | lookup users uid AS user OUTPUTNEW role, label AS name | sort name | dedup user | head 5 | reverse`)
		if r.Status != Valid || len(r.Lineage[2].Transitions) != 1 || r.Lineage[2].Transitions[0].Output != "name" {
			t.Fatalf("%+v", r)
		}
		spl2Ref(t, r, "user", "read")
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName == "uid" {
				t.Fatal("catalog column became event read")
			}
		}
	})
}
func TestSPL2MixedDialectConcurrent(t *testing.T) {
	splText := `search user=* | eval x=coalesce(user)`
	baseline, e := Analyze(QueryDocument{Text: splText})
	if e != nil {
		t.Fatal(e)
	}
	expected, _ := json.Marshal(baseline)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 8; j++ {
				s, e := Analyze(QueryDocument{Text: splText})
				b, _ := json.Marshal(s)
				if e != nil || string(b) != string(expected) {
					t.Errorf("SPL changed %s %v", b, e)
				}
				r, e := Analyze(QueryDocument{Text: `FROM main | eval x=coalesce(user) | table x`, Language: "spl2"})
				if e != nil || r.Status != Valid {
					t.Errorf("SPL2 %+v %v", r, e)
				}
				universe := SourceUniverse{Fields: []string{"host", "host.child"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}
				legacy, err := AnalyzeWithSourceUniverse(QueryDocument{Text: "fields host*"}, universe)
				if err != nil || len(legacy.Expansions) != 1 || !legacy.Expansions[0].Complete || len(legacy.Expansions[0].Matches) != 2 {
					t.Errorf("SPL resolver isolation: %+v %v", legacy, err)
				}
				selected, err := AnalyzeWithSourceUniverse(QueryDocument{Text: "FROM main | fields 'host*'", Language: "spl2"}, universe)
				if err != nil || len(selected.Expansions) != 1 || selected.Expansions[0].Complete || len(selected.Expansions[0].Matches) != 1 || selected.Expansions[0].Matches[0].Name != "host" {
					t.Errorf("SPL2 resolver isolation: %+v %v", selected, err)
				}
			}
		}()
	}
	wg.Wait()
}
func TestSPL2Capabilities(t *testing.T) {
	m, e := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if e != nil || m.Language != "spl2" {
		t.Fatalf("%+v %v", m, e)
	}
	found := false
	for i, c := range m.Functions {
		if c.Name == "coalesce" {
			found = true
			if !c.SemanticSupported || !strings.Contains(c.Limitations[0], "1") {
				t.Fatalf("%+v", c)
			}
			m.Functions[i].Limitations[0] = "mutated"
		}
	}
	if !found {
		t.Fatal("missing coalesce")
	}
	n, _ := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	for _, c := range n.Functions {
		if c.Limitations[0] == "mutated" {
			t.Fatal("shared mutable capability")
		}
	}
}

func TestSPL2QuotedDotsAcrossSharedTransfers(t *testing.T) {
	for _, tail := range []string{`table 'actor.name'`, `fields 'actor.name'`, `rename 'actor.name' AS name`, `stats count() BY 'actor.name'`} {
		document := QueryDocument{Text: "FROM main | " + tail, Language: "spl2"}
		for _, universe := range []SourceUniverse{{Fields: []string{"actor.name"}, Complete: true}, {Fields: []string{"actor.name"}}, {Fields: []string{"actor.name"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }}} {
			source, err := AnalyzeWithSourceUniverse(document, universe)
			if err != nil {
				t.Fatal(err)
			}
			expected := "source"
			if universe.Resolve != nil {
				expected = "indeterminate"
			}
			found := false
			for _, ref := range source.Result.References {
				if ref.NormalizedName == "actor.name" {
					found = true
					if ref.Binding != expected {
						t.Fatalf("%s %+v", tail, ref)
					}
				}
			}
			if !found {
				t.Fatal("missing dotted source")
			}
		}
		r := spl2AnalyzeTest(t, document.Text)
		for _, ref := range r.References {
			if ref.NormalizedName == "actor.name" && ref.Binding != "source" && !strings.HasPrefix(tail, "fields") {
				t.Fatalf("plain name lost %+v", ref)
			}
		}
	}
}

func TestSPL2CommandSourceEvidence(t *testing.T) {
	t.Run("finite wildcard and internals", func(t *testing.T) {
		text := `FROM main | fields - _time | fields 'b*'`
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text, Language: "spl2"}, SourceUniverse{Fields: []string{"bytes", "bits", "_raw", "_time"}, Complete: true})
		if e != nil {
			t.Fatal(e)
		}
		if r.Result.Status != Valid || !reflect.DeepEqual(r.Expansions, []FieldExpansion{{ReferenceID: "ref-2", Complete: true, Matches: []ExpandedField{{Name: "bits", Binding: "source"}, {Name: "bytes", Binding: "source"}}}}) {
			t.Fatalf("%+v", r)
		}
		want := FieldState{Fields: []FieldBinding{{Name: "bits", OriginReferenceIDs: []string{"ref-2"}}, {Name: "bytes", OriginReferenceIDs: []string{"ref-2"}}}, Removed: []string{"_time"}}
		if !reflect.DeepEqual(r.Result.Lineage[2].After, want) {
			t.Fatalf("%+v", r.Result.Lineage)
		}
		if len(r.Result.References) != 3 {
			t.Fatalf("fabricated internals %+v", r.Result.References)
		}
	})
	t.Run("partial wildcard retains admitted member", func(t *testing.T) {
		text := `FROM synthetic_events | fields 'synthetic_host*'`
		r, err := AnalyzeWithSourceUniverse(QueryDocument{Text: text, Language: "spl2"}, SourceUniverse{
			Fields:   []string{"synthetic_host"},
			Complete: false,
			Resolve: func(name string) SourceFieldAdmission {
				if name == "synthetic_host" {
					return SourceFieldAdmitted
				}
				return SourceFieldIndeterminate
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if r.Result.Status != Incomplete || len(r.Expansions) != 1 || r.Expansions[0].Complete || !reflect.DeepEqual(r.Expansions[0].Matches, []ExpandedField{{Name: "synthetic_host", Binding: "source"}}) {
			t.Fatalf("partial wildcard evidence = %+v", r)
		}
		last := r.Result.Lineage[len(r.Result.Lineage)-1].After
		if !last.Open || !last.Uncertain || len(last.Fields) != 1 || last.Fields[0].Name != "synthetic_host" {
			t.Fatalf("partial wildcard state = %+v", last)
		}
	})
	for _, command := range []string{`eventstats count() AS n`, `streamstats BY host current=true reset before code=500 window=3 count() AS n`, `stats allnum=true count() AS n`, `streamstats current=false count() AS n`} {
		t.Run(command, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval prior=1 | "+command)
			if r.Status != Valid {
				t.Fatalf("%+v", r)
			}
			last := r.Lineage[2].After
			preserve := !strings.HasPrefix(command, "stats")
			seenPrior := false
			for _, f := range last.Fields {
				seenPrior = seenPrior || f.Name == "prior"
			}
			if preserve != seenPrior {
				t.Fatalf("%+v", last)
			}
			conditional := strings.Contains(command, "allnum=true") || strings.Contains(command, "current=false")
			if r.Lineage[2].Transitions[len(r.Lineage[2].Transitions)-1].Conditional != conditional {
				t.Fatalf("%+v", r.Lineage)
			}
			if strings.Contains(command, "code") {
				spl2Ref(t, r, "code", "read")
			}
		})
	}
	for _, query := range []string{`search index=main code=-1`, `search index=main code=-bytes`, `search index=main code=-(1+2)`} {
		r := spl2AnalyzeTest(t, query)
		if r.Status != Incomplete {
			t.Fatalf("%+v", r)
		}
		for _, ref := range r.References {
			if ref.Kind == "field" && ref.NormalizedName != "code" {
				t.Fatalf("search literal evaluated %+v", ref)
			}
		}
	}
	t.Run("source reset", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM first | eval local=1 | FROM second | where local>0`)
		if spl2Ref(t, r, "local", "read").Binding != "source" || len(r.Lineage[2].After.Fields) != 0 {
			t.Fatalf("%+v", r)
		}
	})
	t.Run("UTF8 CRLF caller data", func(t *testing.T) {
		text := "FROM 'données'\r\n| eval 'métrique'=bytes | where 'métrique'>0"
		r, e := Analyze(QueryDocument{Text: text, Language: "spl2", SourceID: " exact\r\n "})
		if e != nil || r.Status != Valid || r.Document.Text != text || r.Document.SourceID != " exact\r\n " {
			t.Fatalf("%+v %v", r, e)
		}
		ref := spl2Ref(t, r, "métrique", "create")
		if ref.OriginalName != "'métrique'" || ref.Location.Start.Line != 2 || ref.Location.Start.Column != 8 || text[ref.Location.Start.Offset:ref.Location.End.Offset] != ref.OriginalName {
			t.Fatalf("%+v", ref)
		}
	})
}

func TestSPL2NullTestReadsDoNotEstablishPresence(t *testing.T) {
	for _, name := range []string{"isnull", "isnotnull"} {
		t.Run(name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | eval answer="+name+"((absent))")
			ref := spl2Ref(t, r, "absent", "null_test")
			if ref.Binding != "source" || r.Status != Valid || len(r.Lineage[1].After.Fields) != 1 || r.Lineage[1].After.Fields[0].Name != "answer" {
				t.Fatalf("%+v", r)
			}
		})
	}
	r := spl2AnalyzeTest(t, `FROM main | eval x=null, answer=isnull(x)`)
	if r.Status != Valid || spl2Ref(t, r, "x", "null_test").Binding != "unavailable" || len(r.Lineage[1].After.Removed) != 1 {
		t.Fatalf("%+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM main | eval x=null, answer=isnull(x) | where x>0`)
	if r.Status != Invalid || spl2Ref(t, r, "x", "read").Binding != "unavailable" {
		t.Fatalf("%+v", r)
	}
	r = spl2AnalyzeTest(t, `FROM main | eval x=tonumber("17"), answer=isnotnull(x)`)
	if r.Status != Valid || spl2Ref(t, r, "x", "null_test").Binding != "indeterminate" || !r.Lineage[1].After.Fields[1].Conditional {
		t.Fatalf("%+v", r)
	}
	for _, query := range []string{`FROM main | eval x=isnull(a+1)`, `FROM main | eval x=isnotnull(abs(a))`, `FROM main | eval x=isnull(a,b)`, `FROM main | eval x=isnull(value:a)`} {
		r := spl2AnalyzeTest(t, query)
		spl2Ref(t, r, "a", "read")
		for _, ref := range r.References {
			if ref.NormalizedName == "a" && ref.Role == "null_test" {
				t.Fatalf("arbitrary subtree exempted %+v", r)
			}
		}
	}
}

func TestSPL2ResolverWildcardCandidateIdentity(t *testing.T) {
	for _, pattern := range []string{"actor.*", "actor*", "*"} {
		for _, resolver := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/resolver=%v", pattern, resolver), func(t *testing.T) {
				text := `FROM main | eval 'actor.local'=1 | fields '` + pattern + `'`
				u := SourceUniverse{Fields: []string{"actor.name", "host"}, Complete: true}
				if resolver {
					u.Resolve = func(string) SourceFieldAdmission { return SourceFieldAdmitted }
				}
				r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text, Language: "spl2"}, u)
				if e != nil {
					t.Fatal(e)
				}
				if len(r.Expansions) != 1 {
					t.Fatalf("%+v", r)
				}
				exp := r.Expansions[0]
				if exp.Complete == resolver {
					t.Fatalf("bad completeness %+v", r)
				}
				matches := map[string]string{}
				for _, m := range exp.Matches {
					matches[m.Name] = m.Binding
				}
				if matches["actor.local"] != "derived" {
					t.Fatalf("lost derived %+v", r)
				}
				if (matches["actor.name"] != "") == resolver {
					t.Fatalf("ambiguous candidate proof %+v", r)
				}
				if pattern == "*" && matches["host"] != "source" {
					t.Fatalf("lost flat proof %+v", r)
				}
				if resolver {
					for _, f := range r.Result.Lineage[2].After.Fields {
						if f.Name == "actor.name" {
							t.Fatalf("materialized withheld source %+v", r)
						}
					}
					used, e := AnalyzeWithSourceUniverse(QueryDocument{Text: text + ` | where 'actor.local'>0 AND 'actor.name'>0`, Language: "spl2"}, u)
					if e != nil {
						t.Fatal(e)
					}
					if spl2Ref(t, used.Result, "actor.local", "read").Binding != "derived" || spl2Ref(t, used.Result, "actor.name", "read").Binding != "indeterminate" {
						t.Fatalf("%+v", used)
					}
				}
			})
		}
	}
	t.Run("implicit internals", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields host`, Language: "spl2"}, SourceUniverse{Fields: []string{"_object.name", "host"}, Complete: true, Resolve: func(string) SourceFieldAdmission { return SourceFieldAdmitted }})
		if e != nil {
			t.Fatal(e)
		}
		for _, f := range r.Result.Lineage[1].After.Fields {
			if f.Name == "_object.name" {
				t.Fatalf("implicit ambiguous internal %+v", r)
			}
		}
	})
	t.Run("partial flat removal", func(t *testing.T) {
		r, e := AnalyzeWithSourceUniverse(QueryDocument{Text: `FROM main | fields - 'actor*'`, Language: "spl2"}, SourceUniverse{Fields: []string{"actor.name"}})
		if e != nil {
			t.Fatal(e)
		}
		if len(r.Expansions) != 1 || r.Expansions[0].Complete || !reflect.DeepEqual(r.Expansions[0].Matches, []ExpandedField{{Name: "actor.name", Binding: "source"}}) {
			t.Fatalf("%+v", r)
		}
	})
}

func TestSPL2DeferredDestinationAndPrerequisiteBoundaries(t *testing.T) {
	t.Run("fillnull intentions", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | fillnull value="0" host,user`)
		for _, name := range []string{"host", "user"} {
			ref := spl2Ref(t, r, name, "output")
			if ref.Binding != "not_applicable" {
				t.Fatalf("destination binding: %+v", ref)
			}
		}
		if len(r.Lineage[len(r.Lineage)-1].After.Fields) != 0 {
			t.Fatalf("destination invented state: %+v", r)
		}
	})
	t.Run("dynamic SID", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `loadjob 'job_${sid}'`)
		for _, ref := range r.References {
			if ref.Kind == "search_job" {
				t.Fatalf("dynamic SID promoted: %+v", r)
			}
		}
	})
	t.Run("timewrap predecessor", func(t *testing.T) {
		r := spl2AnalyzeTest(t, `FROM main | timewrap fortnight`)
		if r.Status != Invalid || !spl2HasCode(r, CodeSyntaxError) || !spl2HasCode(r, CodeUnsupportedSemantics) {
			t.Fatalf("missing prerequisite: %+v", r)
		}
		for _, q := range []string{`FROM main | timechart count() | timewrap fortnight`, `unknown | timewrap fortnight`} {
			r = spl2AnalyzeTest(t, q)
			if r.Status != Incomplete || spl2HasCode(r, CodeSyntaxError) {
				t.Fatalf("unproved/available predecessor: %+v", r)
			}
		}
	})
}

func TestSPL2DeferredRecoveredInputsAndUnionIntentions(t *testing.T) {
	for _, tc := range []struct{ query, name, role string }{
		{`tstats aggregates=[sum(bytes)] byfields=[host source]`, `bytes`, `read`},
		{`FROM main | timechart count() BY host usenull=1`, `host`, `group`},
		{`FROM main | timechart eval(avg(bytes)*avg(duration))`, `bytes`, `read`},
		{`FROM main | timechart eval(avg(bytes)*avg(duration))`, `duration`, `read`},
		{`FROM main | timechart bins= avg(bytes)`, `bytes`, `read`},
	} {
		t.Run(tc.query+tc.name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, tc.query)
			if r.Status != Invalid {
				t.Fatalf("damage promoted: %+v", r)
			}
			spl2Ref(t, r, tc.name, tc.role)
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`mstats aggregates=[avg('cpu.load')] byfields=['host*']`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			ref := spl2Ref(t, r, "host*", "group")
			if ref.Binding != "indeterminate" || ref.Resolution != "wildcard" {
				t.Fatalf("wildcard membership invented: %+v", r)
			}
			for _, l := range r.Lineage {
				for _, f := range l.After.Fields {
					if f.Name == "host*" {
						t.Fatalf("pattern installed as field: %+v", r)
					}
				}
			}
		})
	}
	for _, q := range []string{`union main, archive`, `union main`, `FROM main | union other, [FROM archive]`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if len(r.Dependencies.Datasets) == 0 {
				t.Fatalf("named union inputs lost: %+v", r)
			}
			spl2Ref(t, r, "main", "read")
			assertCorpusIntegrity(t, r)
		})
	}
	for _, q := range []string{`FROM main | append [search index=audit | stats count()`, `FROM main | appendcols [FROM other | eval x=1`} {
		t.Run(q, func(t *testing.T) {
			r := spl2AnalyzeTest(t, q)
			if r.Status != Invalid || len(r.Scopes) != 1 {
				t.Fatalf("unclosed original child promoted: %+v", r)
			}
			spl2Ref(t, r, "main", "read")
			assertCorpusIntegrity(t, r)
		})
	}
	control := spl2AnalyzeTest(t, `FROM main | append [FROM child | eval x=1]`)
	if len(control.Scopes) != 2 {
		t.Fatalf("intact child lost: %+v", control)
	}
	spl2Ref(t, control, "x", "create")
}

func TestSPL2DeferredPatternNeverExpandsAndUnionDoesNotInstallRows(t *testing.T) {
	source, err := AnalyzeWithSourceUniverse(QueryDocument{Text: `mstats aggregates=[avg('cpu.load')] byfields=['host*']`, Language: "spl2"}, SourceUniverse{Fields: []string{"cpu.load", "host1"}, Complete: true})
	if err != nil {
		t.Fatal(err)
	}
	ref := spl2Ref(t, source.Result, "host*", "group")
	if ref.Binding != "indeterminate" || ref.Resolution != "wildcard" {
		t.Fatalf("forbidden group expanded: %+v", source)
	}
	for _, lineage := range source.Result.Lineage {
		for _, field := range lineage.After.Fields {
			if field.Name == "host*" || field.Name == "host1" {
				t.Fatalf("forbidden group membership: %+v", source)
			}
		}
	}
	for _, q := range []string{`union [{a:1}],[{b:2}]`, `FROM main | union [{a:1}],[{b:2}]`} {
		r := spl2AnalyzeTest(t, q)
		for _, ref := range r.References {
			if ref.Kind == "field" {
				t.Fatalf("union row shape leaked: %+v", r)
			}
		}
		if len(r.Scopes) != 1 || len(r.Lineage[len(r.Lineage)-1].After.Fields) != 0 {
			t.Fatalf("union row ownership invented: %+v", r)
		}
	}
}
