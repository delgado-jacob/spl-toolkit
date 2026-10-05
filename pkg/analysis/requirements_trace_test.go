package analysis

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var _ func(*Result, *sourceRefinement, *requirementTrace) = finalizeReferences

var _ map[string]int = newRequirementTrace().pendingReferenceIndexes
var _ map[string]struct{} = newRequirementTrace().incompleteStageIDs

func TestRequirementTraceIndexesStaySynchronized(t *testing.T) {
	trace := newRequirementTrace()
	reference := Reference{ID: "pending-0", StageID: "stage-pending", OriginReferenceIDs: []string{}}
	trace.recordReference(reference, true, false, trace.nextEvent())
	if got := trace.reference("pending-0"); got != &trace.references[0] {
		t.Fatalf("pending reference index returned %p, want %p", got, &trace.references[0])
	}

	diagnostics := []Diagnostic{{Code: CodeUnsupportedSemantics}, {Code: CodeSyntaxError}}
	for _, diagnostic := range diagnostics {
		trace.recordDiagnostic(diagnostic, true, []string{"pending-0"}, trace.nextEvent())
	}
	if environment := newRequirementEnvironment(trace); environment.stageIncomplete("stage-pending") {
		t.Fatal("stage was indexed before parser diagnostic synchronization")
	}
	indexPointer := reflect.ValueOf(trace.incompleteStageIDs).Pointer()
	diagnosticCount := len(trace.diagnostics)
	nextOrdinal := trace.nextOrdinal
	eventOrdinals := make([]int, len(trace.diagnostics))
	for i := range trace.diagnostics {
		eventOrdinals[i] = trace.diagnostics[i].eventOrdinal
	}
	assertStable := func() {
		t.Helper()
		if len(trace.diagnostics) != diagnosticCount || trace.nextOrdinal != nextOrdinal {
			t.Fatalf("parser diagnostic synchronization changed trace cardinality or ordinal state: diagnostics=%d want=%d next=%d want=%d", len(trace.diagnostics), diagnosticCount, trace.nextOrdinal, nextOrdinal)
		}
		if got := reflect.ValueOf(trace.incompleteStageIDs).Pointer(); got != indexPointer {
			t.Fatalf("parser diagnostic synchronization replaced the incomplete-stage index: got %x want %x", got, indexPointer)
		}
		for i := range trace.diagnostics {
			if trace.diagnostics[i].eventOrdinal != eventOrdinals[i] {
				t.Fatalf("parser diagnostic synchronization changed event ordinal %d from %d to %d", i, eventOrdinals[i], trace.diagnostics[i].eventOrdinal)
			}
		}
	}
	trace.syncParserDiagnostics([]Diagnostic{{Code: CodeUnsupportedSemantics, StageID: "stage-pending"}, {Code: CodeSyntaxError, StageID: "stage-pending"}})
	assertStable()
	environment := newRequirementEnvironment(trace)
	clone := environment.clone()
	if !environment.stageIncomplete("stage-pending") || !clone.stageIncomplete("stage-pending") {
		t.Fatal("environment clone lost shared incomplete-stage index")
	}
	trace.syncParserDiagnostics([]Diagnostic{{Code: CodeUnsupportedSemantics}, {Code: CodeSyntaxError, StageID: "stage-pending"}})
	assertStable()
	if !environment.stageIncomplete("stage-pending") || !clone.stageIncomplete("stage-pending") {
		t.Fatal("clearing one of two diagnostic owners removed a still-incomplete stage")
	}
	trace.syncParserDiagnostics(diagnostics)
	assertStable()
	if environment.stageIncomplete("stage-pending") || clone.stageIncomplete("stage-pending") {
		t.Fatal("clearing the final diagnostic owner retained a complete stage")
	}
	trace.syncParserDiagnostics([]Diagnostic{{Code: CodeUnsupportedSemantics, StageID: "stage-pending"}, {Code: CodeSyntaxError, StageID: "stage-pending"}})
	assertStable()

	trace.remapReferences(map[string]string{"pending-0": "ref-0"})
	if got := trace.reference("pending-0"); got.reference.ID != "ref-0" {
		t.Fatalf("pending reference index was not synchronized through remapping: %+v", got)
	}
	trace.remapStages(map[string]string{"stage-pending": "stage-0"})
	if environment.stageIncomplete("stage-pending") || !environment.stageIncomplete("stage-0") || !clone.stageIncomplete("stage-0") {
		t.Fatalf("incomplete-stage index was not remapped: %+v", trace.incompleteStageIDs)
	}
}

func TestRequirementTraceCloneOwnsExactIdentity(t *testing.T) {
	trace := newRequirementTrace()
	trace.recordReference(Reference{ID: "pending-0", Kind: "field", FieldIdentity: &FieldIdentity{Kind: "path", Segments: []string{"actor", "name"}}}, true, false, trace.nextEvent())
	trace.reference("pending-0").fieldIdentity = pathFieldIdentity("", []string{"actor", "name"})
	cloned := trace.clone()
	cloned.references[0].reference.FieldIdentity.Segments[0] = "changed"
	cloned.references[0].fieldIdentity.Segments[0] = "changed"
	if got := trace.references[0]; got.reference.FieldIdentity.Segments[0] != "actor" || got.fieldIdentity.Segments[0] != "actor" {
		t.Fatalf("trace clone aliases exact identity: %+v", got)
	}
}

func TestRequirementTraceKeyUsesPrivateIdentityOnlyForExactFields(t *testing.T) {
	identity := atomicFieldIdentity("actor[]")
	for _, reference := range []Reference{
		{Kind: "field", Resolution: "dynamic", NormalizedName: "actor[]"},
		{Kind: "dataset", Resolution: "exact", NormalizedName: "actor[]"},
	} {
		key := requirementTraceKey(requirementTraceReference{reference: reference, fieldIdentity: identity})
		if key.fieldKey != "" {
			t.Errorf("non-exact-field trace key used private identity: %+v", key)
		}
	}
	key := requirementTraceKey(requirementTraceReference{reference: Reference{Kind: "field", Resolution: "exact", NormalizedName: "actor[]"}, fieldIdentity: identity})
	if key.fieldKey == "" {
		t.Fatalf("exact-field trace key lost private identity: %+v", key)
	}
}

func TestRequirementTraceSPL2DatasetParameterProjection(t *testing.T) {
	for _, name := range []string{"$target_1", "$view"} {
		t.Run(name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: "FROM " + name, Language: "spl2", Profile: "splunkd", Version: "current"})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Valid || !reflect.DeepEqual(result.Dependencies.Datasets, []string{name}) {
				t.Fatalf("dataset parameter result = status %q dependencies %+v diagnostics %+v", result.Status, result.Dependencies, result.Diagnostics)
			}
			ref := spl2Ref(t, result, name, "read")
			if ref.Kind != "dataset" || ref.Resolution != "exact" || ref.Binding != "not_applicable" || ref.OriginalName != name {
				t.Fatalf("dataset parameter reference = %+v", ref)
			}
			item := requirementItem(result.Requirements, "dataset", name, "read")
			if item == nil || item.Necessity != "required" || item.Resolution != "exact" || len(item.Occurrences) != 1 || item.Occurrences[0].ReferenceID != ref.ID {
				t.Fatalf("dataset parameter requirement = %+v", item)
			}
		})
	}
}

func TestRequirementTraceSPL2StaticDescriptorProjection(t *testing.T) {
	tests := []struct {
		name, query, identity string
	}{
		{
			name:     "nested keys sort recursively",
			query:    `FROM {kind: "index", properties: {z: 1, a: {b: 2, a: 1}, labels: ["b", "a"]}}`,
			identity: `{"kind":"index","properties":{"a":{"a":1,"b":2},"labels":["b","a"],"z":1}}`,
		},
		{
			name:     "optional properties",
			query:    `FROM {kind: "index"}`,
			identity: `{"kind":"index"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Valid || !reflect.DeepEqual(result.Dependencies.Datasets, []string{tc.identity}) {
				t.Fatalf("descriptor result = status %q dependencies %+v diagnostics %+v", result.Status, result.Dependencies, result.Diagnostics)
			}
			ref := spl2Ref(t, result, tc.identity, "read")
			if ref.Kind != "dataset" || ref.OriginalName != tc.query[len("FROM "):] || ref.Resolution != "exact" {
				t.Fatalf("descriptor reference = %+v", ref)
			}
			item := requirementItem(result.Requirements, "dataset", tc.identity, "read")
			if item == nil || item.Necessity != "required" || len(item.Occurrences) != 1 || item.Occurrences[0].ReferenceID != ref.ID {
				t.Fatalf("descriptor requirement = %+v", item)
			}
		})
	}
}

func TestRequirementTraceSPL2DescriptorRejectsUnprovedIdentity(t *testing.T) {
	for _, query := range []string{
		`FROM {kind: "index", properties: {a: 1, "a": 2}}`,
		`FROM {kind: "index", properties: {nested: {a: 1, "a": 2}}}`,
		`FROM {properties: {name: "main"}}`,
		`FROM {kind: $kind, properties: {name: "main"}}`,
		`FROM {kind: "index", properties: {name: dataset_name}}`,
		`FROM {kind: "index", properties: {name: coalesce(dataset.name, fallback)}}`,
	} {
		t.Run(query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status == Valid || result.Requirements.Coverage.Complete || len(result.Dependencies.Datasets) != 0 {
				t.Fatalf("dynamic descriptor result = status %q dependencies %+v requirements %+v", result.Status, result.Dependencies, result.Requirements)
			}
			var dynamic *Reference
			for i := range result.References {
				ref := &result.References[i]
				if ref.Kind == "dataset" && ref.Resolution == "exact" {
					t.Fatalf("unproved descriptor gained exact identity: %+v", ref)
				}
				if ref.Kind == "dataset" && ref.Resolution == "dynamic" {
					if dynamic != nil {
						t.Fatalf("duplicate dynamic descriptor references: %+v", result.References)
					}
					dynamic = ref
				}
			}
			if dynamic == nil || dynamic.Location.Start.Offset != len("FROM ") || dynamic.Location.End.Offset != len(query) {
				t.Fatalf("missing located dynamic descriptor reference: %+v", result.References)
			}
			item := requirementItem(result.Requirements, "dataset", dynamic.NormalizedName, "read")
			if item == nil || item.Resolution != "dynamic" || item.Necessity != "conditional" || len(item.Occurrences) != 1 || item.Occurrences[0].ReferenceID != dynamic.ID {
				t.Fatalf("dynamic descriptor requirement = %+v", item)
			}
		})
	}
}

func TestRequirementTraceSPL2DynamicDescriptorRetainsChildEvidence(t *testing.T) {
	query := `FROM {kind: "index", properties: {name: mystery(dataset.name, fallback)}}`
	result, err := Analyze(QueryDocument{Text: query, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if result.Coverage.SyntaxComplete != true || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete || len(result.Dependencies.Datasets) != 0 {
		t.Fatalf("dynamic descriptor coverage = analysis %+v requirements %+v dependencies %+v", result.Coverage, result.Requirements.Coverage, result.Dependencies)
	}
	for _, original := range []string{"dataset.name", "fallback"} {
		found := false
		for _, ref := range result.References {
			if ref.Kind == "field" && ref.OriginalName == original {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing dynamic descriptor child %q: %+v", original, result.References)
		}
	}
	foundCall := false
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == CodeUnsupportedFunction {
			foundCall = true
			if got := query[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset]; got != "mystery(dataset.name, fallback)" {
				t.Fatalf("dynamic descriptor call diagnostic owns %q: %+v", got, diagnostic)
			}
		}
	}
	if !foundCall {
		t.Fatalf("dynamic descriptor lost independently parsed call evidence: %+v", result.Diagnostics)
	}
}

func TestSPL2ParserDamageSynchronizesRequirementStateBeforeDownstreamTransfers(t *testing.T) {
	tests := []struct {
		name            string
		query           string
		syntaxStage     string
		downstreamStage string
		nested          bool
	}{
		{
			name:            "recovered eventstats",
			query:           "FROM main | eventstats count() by host | where a=1",
			syntaxStage:     "stage-1",
			downstreamStage: "stage-2",
		},
		{
			name:            "recovered table field list",
			query:           "FROM main | table host user | where a=1",
			syntaxStage:     "stage-1",
			downstreamStage: "stage-2",
		},
		{
			name:            "nested scope stage remap",
			query:           "FROM main | append [FROM child] | table host user | where a=1",
			syntaxStage:     "stage-3",
			downstreamStage: "stage-4",
			nested:          true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			document := QueryDocument{Text: tc.query, Language: "spl2"}
			result, trace, err := analyzeRewriteWithTrace(document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Invalid || result.Requirements.QueryStatus != Invalid {
				t.Fatalf("status = analysis %q requirements %q, want invalid/invalid", result.Status, result.Requirements.QueryStatus)
			}

			var syntax *Diagnostic
			for i := range result.Diagnostics {
				diagnostic := &result.Diagnostics[i]
				if diagnostic.Code == CodeSyntaxError {
					if syntax != nil {
						t.Fatalf("multiple syntax diagnostics: %+v", result.Diagnostics)
					}
					syntax = diagnostic
				}
			}
			if syntax == nil || syntax.StageID != tc.syntaxStage || syntax.ScopeID != "scope-0" {
				t.Fatalf("syntax diagnostic = %+v, want stage=%q scope=scope-0", syntax, tc.syntaxStage)
			}
			if !reflect.DeepEqual(result.Requirements.Diagnostics, result.Diagnostics) {
				t.Fatalf("query-only diagnostics diverged from plain analysis:\nanalysis=%+v\nrequirements=%+v", result.Diagnostics, result.Requirements.Diagnostics)
			}
			for _, diagnostic := range result.Requirements.Diagnostics {
				if diagnostic.Code == CodeUnavailableField {
					t.Fatalf("parser damage produced a private unavailable-field diagnostic: %+v", diagnostic)
				}
			}

			var public *Reference
			for i := range result.References {
				reference := &result.References[i]
				if reference.Kind == "field" && reference.NormalizedName == "a" && reference.Role == "read" {
					public = reference
				}
			}
			private := requirementTraceReferenceByNameAndRole(t, trace, "a", "read")
			if public == nil || public.Binding != "indeterminate" || public.StageID != tc.downstreamStage || public.ScopeID != "scope-0" {
				t.Fatalf("downstream public reference = %+v, want indeterminate at %s/scope-0", public, tc.downstreamStage)
			}
			if private.reference.ID != public.ID || private.reference.Binding != "indeterminate" || private.reference.StageID != tc.downstreamStage || private.reference.ScopeID != "scope-0" || private.directExternal || !private.conditional {
				t.Fatalf("downstream query-only reference = %+v, want conditional indeterminate counterpart to %+v", private, public)
			}

			var item *RequirementItem
			for i := range result.Requirements.Items {
				candidate := &result.Requirements.Items[i]
				if candidate.Kind == "field" && candidate.Identity == "a" && candidate.Role == "read" {
					item = candidate
				}
			}
			if item == nil || item.Necessity != "conditional" || item.Resolution != "exact" || len(item.Occurrences) != 1 {
				t.Fatalf("downstream requirement item = %+v, want one conditional exact occurrence", item)
			}
			occurrence := item.Occurrences[0]
			if occurrence.ReferenceID != public.ID || occurrence.Binding != "indeterminate" || occurrence.StageID != tc.downstreamStage || occurrence.ScopeID != "scope-0" {
				t.Fatalf("downstream requirement occurrence = %+v, want final stage/scope and indeterminate binding", occurrence)
			}

			gapKeys := map[string]bool{}
			foundDownstreamGap := false
			for _, gap := range result.Requirements.Gaps {
				keyBytes, err := json.Marshal(struct {
					Code            string
					ReferenceIDs    []string
					DiagnosticCodes []string
				}{gap.Code, gap.ReferenceIDs, gap.DiagnosticCodes})
				if err != nil {
					t.Fatal(err)
				}
				key := string(keyBytes)
				if gapKeys[key] {
					t.Fatalf("duplicate requirement gap key %s: %+v", key, result.Requirements.Gaps)
				}
				gapKeys[key] = true
				if gap.Code == CodeRequirementIndeterminate && reflect.DeepEqual(gap.ReferenceIDs, []string{public.ID}) {
					foundDownstreamGap = true
				}
			}
			if !foundDownstreamGap {
				t.Fatalf("downstream conditional reference has no exact indeterminate gap: %+v", result.Requirements.Gaps)
			}

			if tc.nested {
				assertSPL2RequirementStageRemapIntegrity(t, result, trace)
				child := requirementTraceReferenceByNameAndRole(t, trace, "child", "read")
				if child.reference.Kind != "dataset" || child.reference.StageID != "stage-2" || child.reference.ScopeID != "scope-1" {
					t.Fatalf("nested dataset trace lost final stage/scope: %+v", child)
				}
				return
			}

			standalone, err := Requirements(document)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(standalone, &result.Requirements) {
				t.Fatalf("standalone requirements differ from plain analysis:\nstandalone=%+v\nplain=%+v", standalone, result.Requirements)
			}
			refined := map[string]func() (*Result, error){
				"field list": func() (*Result, error) {
					got, err := AnalyzeWithSourceFields(document, []string{"a", "host", "user"})
					if err != nil {
						return nil, err
					}
					return got.Result, nil
				},
				"complete universe": func() (*Result, error) {
					got, err := AnalyzeWithSourceUniverse(document, SourceUniverse{Fields: []string{"a", "host", "user"}, Complete: true})
					if err != nil {
						return nil, err
					}
					return got.Result, nil
				},
				"partial universe": func() (*Result, error) {
					got, err := AnalyzeWithSourceUniverse(document, SourceUniverse{Fields: []string{"host"}, Complete: false})
					if err != nil {
						return nil, err
					}
					return got.Result, nil
				},
			}
			for name, run := range refined {
				got, err := run()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got.Requirements, result.Requirements) {
					t.Errorf("requirements changed under %s refinement:\nplain=%+v\nrefined=%+v", name, result.Requirements, got.Requirements)
				}
			}
		})
	}
}

func assertSPL2RequirementStageRemapIntegrity(t *testing.T, result *Result, trace *requirementTrace) {
	t.Helper()
	stages := map[string]string{}
	for _, stage := range result.Stages {
		if stages[stage.ID] != "" {
			t.Fatalf("duplicate finalized stage ID %q: %+v", stage.ID, result.Stages)
		}
		stages[stage.ID] = stage.ScopeID
	}
	assertOwner := func(kind, id, scope string) {
		t.Helper()
		if id == "" || stages[id] != scope {
			t.Fatalf("%s owner %q/%q is not a finalized stage: %+v", kind, id, scope, result.Stages)
		}
	}
	for _, reference := range result.References {
		assertOwner("public reference", reference.StageID, reference.ScopeID)
	}
	for _, entry := range trace.references {
		assertOwner("query-only reference", entry.reference.StageID, entry.reference.ScopeID)
	}
	for _, diagnostic := range result.Requirements.Diagnostics {
		if diagnostic.StageID != "" {
			assertOwner("requirement diagnostic", diagnostic.StageID, diagnostic.ScopeID)
		}
	}
	for _, item := range result.Requirements.Items {
		for _, occurrence := range item.Occurrences {
			assertOwner("requirement occurrence", occurrence.StageID, occurrence.ScopeID)
		}
	}
}

func TestRequirementTraceClassifiesFieldOrigins(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "search src=* | eval derived=src | table src derived"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || trace == nil {
		t.Fatal("analysis did not return result and requirement trace")
	}

	want := []struct {
		pendingID    string
		name         string
		role         string
		binding      string
		direct       bool
		conditional  bool
		eventOrdinal int
	}{
		{"pending-0", "src", "filter", "source", true, false, 0},
		{"pending-1", "src", "read", "source", true, false, 1},
		{"pending-2", "derived", "create", "definition", false, false, 2},
		{"pending-3", "src", "read", "source", true, false, 3},
		{"pending-4", "derived", "read", "derived", false, false, 4},
	}
	if len(trace.references) != len(want) {
		t.Fatalf("trace references = %+v, want %d entries", trace.references, len(want))
	}
	for i, expected := range want {
		got := trace.references[i]
		if got.pendingID != expected.pendingID || got.reference.NormalizedName != expected.name || got.reference.Role != expected.role || got.reference.Binding != expected.binding || got.directExternal != expected.direct || got.conditional != expected.conditional || got.eventOrdinal != expected.eventOrdinal {
			t.Errorf("reference %d = %+v, want %+v", i, got, expected)
		}
	}

	for _, tc := range []struct {
		name          string
		document      QueryDocument
		pendingID     string
		referenceName string
		role          string
		binding       string
		direct        bool
		conditional   bool
		eventOrdinal  int
	}{
		{"rename source", QueryDocument{Text: "search host=x | rename host AS node"}, "pending-1", "host", "read", "source", true, false, 1},
		{"rename target", QueryDocument{Text: "search host=x | rename host AS node"}, "pending-2", "node", "rename", "definition", false, false, 2},
		{"removal", QueryDocument{Text: "search host=x | fields - host"}, "pending-1", "host", "remove", "not_applicable", false, false, 1},
		{"SPL null test", QueryDocument{Text: "| where isnull(absent)"}, "pending-0", "absent", "null_test", "source", false, false, 0},
		{"SPL non-null test", QueryDocument{Text: "| eval answer=isnotnull(absent)"}, "pending-0", "absent", "null_test", "source", false, false, 0},
		{"null test", QueryDocument{Text: "FROM main | eval answer=isnull(absent)", Language: "spl2"}, "pending-1", "absent", "null_test", "source", false, false, 1},
		{"unavailable local field", QueryDocument{Text: "search host=x | fields - host | table host"}, "pending-2", "host", "read", "unavailable", false, false, 2},
		{"wildcard read", QueryDocument{Text: "search src=* | table s*"}, "pending-1", "s*", "read", "indeterminate", false, true, 1},
		{"indeterminate source", QueryDocument{Text: "| mystery | table host"}, "pending-0", "host", "read", "indeterminate", false, true, 1},
		{"projection closes omitted source", QueryDocument{Text: "search host=x | table other | where host=x"}, "pending-2", "host", "read", "unavailable", false, false, 2},
		{"aggregation closes source", QueryDocument{Text: "search host=x | stats count AS total | table host"}, "pending-2", "host", "read", "unavailable", false, false, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, trace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			entry := trace.reference(tc.pendingID)
			if entry.reference.NormalizedName != tc.referenceName || entry.reference.Role != tc.role || entry.reference.Binding != tc.binding || entry.directExternal != tc.direct || entry.conditional != tc.conditional || entry.eventOrdinal != tc.eventOrdinal {
				t.Fatalf("trace reference = %+v", entry)
			}
		})
	}
}

func TestRequirementEnvironmentCloneOwnsState(t *testing.T) {
	trace := newRequirementTrace()
	env := newEnvironmentWithRequirementTrace(trace)
	hostIdentity := atomicFieldIdentity("host")
	hostKey, _ := hostIdentity.privateKey()
	oldKey, _ := atomicFieldIdentity("old").privateKey()
	newKey, _ := atomicFieldIdentity("new").privateKey()
	env.requirements.fields[hostKey] = requirementField{identity: hostIdentity, source: true, conditional: true, origins: []string{"pending-0"}}
	env.requirements.removed[oldKey] = true
	clone := env.clone()

	clonedField := clone.requirements.fields[hostKey]
	clonedField.origins[0] = "changed"
	clonedField.conditional = false
	clone.requirements.fields[hostKey] = clonedField
	clone.requirements.removed[newKey] = true
	clone.requirements.open = false
	clone.requirements.uncertain = true

	field := env.requirements.fields[hostKey]
	if !field.source || !field.conditional || !reflect.DeepEqual(field.origins, []string{"pending-0"}) || env.requirements.removed[newKey] || !env.requirements.open || env.requirements.uncertain {
		t.Fatalf("clone mutation changed source sidecar: %+v", env.requirements)
	}
	if clone.requirements.trace != trace {
		t.Fatal("clone detached the analysis-wide trace collector")
	}
	env.requirements.uncertain = true
	binding, direct, conditional := env.requirements.read(Reference{NormalizedName: "maybe", Kind: "field", Role: "null_test", Resolution: "exact"})
	if binding != "indeterminate" || direct || conditional {
		t.Fatalf("null test became an obligation: binding %s direct %t conditional %t", binding, direct, conditional)
	}
}

func TestRequirementEnvironmentExactProjectionClonesOrSynthesizesConditionalField(t *testing.T) {
	trace := newRequirementTrace()
	trace.recordReference(Reference{
		ID:                 "pending-origin",
		NormalizedName:     "host",
		Kind:               "field",
		Role:               "read",
		Resolution:         "exact",
		Binding:            "source",
		OriginReferenceIDs: []string{},
	}, true, false, trace.nextEvent())
	trace.recordReference(Reference{
		ID:                 "pending-select",
		NormalizedName:     "host",
		Kind:               "field",
		Role:               "read",
		Resolution:         "exact",
		Binding:            "indeterminate",
		OriginReferenceIDs: []string{"pending-origin"},
	}, false, true, trace.nextEvent())
	trace.reference("pending-select").fieldIdentity = atomicFieldIdentity("host")

	environment := newRequirementEnvironment(trace)
	existingIdentity := atomicFieldIdentity("existing")
	existingKey, _ := existingIdentity.privateKey()
	environment.fields[existingKey] = requirementField{identity: existingIdentity, source: true, origins: []string{"pending-origin"}}
	existing, ok := environment.exactProjection("existing", nil)
	if !ok || !existing.source || existing.conditional || !reflect.DeepEqual(existing.origins, []string{"pending-origin"}) {
		t.Fatalf("existing exact projection = %+v, %t", existing, ok)
	}
	existing.origins[0] = "changed"
	if got := environment.fields[existingKey].origins; !reflect.DeepEqual(got, []string{"pending-origin"}) {
		t.Fatalf("projected existing origins alias environment state: %v", got)
	}

	synthesized, ok := environment.exactProjection("host", []string{"pending-select"})
	if !ok || !synthesized.source || !synthesized.conditional || !reflect.DeepEqual(synthesized.origins, []string{"pending-select", "pending-origin"}) {
		t.Fatalf("synthesized exact projection = %+v, %t", synthesized, ok)
	}
	synthesized.origins[1] = "changed"
	if got := trace.reference("pending-select").reference.OriginReferenceIDs; !reflect.DeepEqual(got, []string{"pending-origin"}) {
		t.Fatalf("synthesized origins alias trace evidence: %v", got)
	}
	trace.recordReference(Reference{ID: "pending-other-identity", NormalizedName: "host", Kind: "field", Role: "read", Resolution: "exact", Binding: "indeterminate"}, false, true, trace.nextEvent())
	trace.reference("pending-other-identity").fieldIdentity = pathFieldIdentity("", []string{"host"})
	if field, ok := environment.exactProjection("host", []string{"pending-other-identity"}); ok {
		t.Fatalf("different exact identity synthesized atomic field: %+v", field)
	}

	for _, reference := range []Reference{
		{ID: "pending-wildcard", NormalizedName: "host", Kind: "field", Role: "read", Resolution: "wildcard", Binding: "indeterminate", OriginReferenceIDs: []string{}},
		{ID: "pending-dynamic", NormalizedName: "host", Kind: "field", Role: "read", Resolution: "dynamic", Binding: "indeterminate", OriginReferenceIDs: []string{}},
		{ID: "pending-unavailable", NormalizedName: "host", Kind: "field", Role: "read", Resolution: "exact", Binding: "unavailable", OriginReferenceIDs: []string{}},
		{ID: "pending-null-test", NormalizedName: "host", Kind: "field", Role: "null_test", Resolution: "exact", Binding: "indeterminate", OriginReferenceIDs: []string{}},
		{ID: "pending-unrelated", NormalizedName: "other", Kind: "field", Role: "read", Resolution: "exact", Binding: "indeterminate", OriginReferenceIDs: []string{}},
	} {
		trace.recordReference(reference, false, reference.Binding == "indeterminate", trace.nextEvent())
		if field, ok := environment.exactProjection("host", []string{reference.ID}); ok {
			t.Errorf("%s evidence synthesized field: %+v", reference.ID, field)
		}
	}

	environment.uncertain = true
	before := len(environment.fields)
	binding, direct, conditional := environment.read(Reference{ID: "pending-read", NormalizedName: "speculative", Kind: "field", Role: "read", Resolution: "exact"})
	if binding != "indeterminate" || direct || !conditional {
		t.Fatalf("uncertain read = %q, %t, %t", binding, direct, conditional)
	}
	if len(environment.fields) != before {
		t.Fatalf("uncertain read installed speculative field: %+v", environment.fields)
	}
}

func TestConditionalExactProjectionPreservesQueryOnlyFields(t *testing.T) {
	for _, tc := range []struct {
		name       string
		document   QueryDocument
		wantByRole map[string]int
	}{
		{
			name:       "SPL table then eval",
			document:   QueryDocument{Text: "search index=main | foobar | table host | eval y=host"},
			wantByRole: map[string]int{"read": 2},
		},
		{
			name:       "SPL aggregation then table",
			document:   QueryDocument{Text: "search index=main | foobar | stats count by host | table host"},
			wantByRole: map[string]int{"group": 1, "read": 1},
		},
		{
			name:       "SPL2 SQL select then eval",
			document:   QueryDocument{Text: "FROM main WHERE unknown_fn(a) SELECT host | eval y=host", Language: "spl2"},
			wantByRole: map[string]int{"read": 2},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Incomplete || result.Requirements.QueryStatus != Incomplete {
				t.Fatalf("status = analysis %q requirements %q, want incomplete/incomplete", result.Status, result.Requirements.QueryStatus)
			}
			for _, diagnostic := range result.Requirements.Diagnostics {
				if diagnostic.Code == CodeUnavailableField {
					t.Fatalf("conditional projected field produced private unavailable diagnostic: %+v", diagnostic)
				}
			}

			traceByID := map[string]requirementTraceReference{}
			for _, entry := range trace.references {
				traceByID[entry.reference.ID] = entry
			}
			for role, wantCount := range tc.wantByRole {
				wantReferenceIDs := []string{}
				for _, reference := range result.References {
					if reference.Kind != "field" || reference.NormalizedName != "host" || reference.Role != role {
						continue
					}
					wantReferenceIDs = append(wantReferenceIDs, reference.ID)
					if reference.Binding != "indeterminate" {
						t.Errorf("public %s reference = %+v, want indeterminate", role, reference)
					}
					entry := traceByID[reference.ID]
					if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
						t.Errorf("query-only %s reference = %+v, want conditional indeterminate", role, entry)
					}
				}
				if len(wantReferenceIDs) != wantCount {
					t.Fatalf("%s host references = %v, want %d", role, wantReferenceIDs, wantCount)
				}

				var item *RequirementItem
				for i := range result.Requirements.Items {
					candidate := &result.Requirements.Items[i]
					if candidate.Kind == "field" && candidate.Identity == "host" && candidate.Role == role {
						item = candidate
					}
				}
				if item == nil || item.Necessity != "conditional" || item.Resolution != "exact" || len(item.Occurrences) != wantCount {
					t.Fatalf("%s requirement item = %+v, want %d conditional exact occurrences", role, item, wantCount)
				}
				gotReferenceIDs := []string{}
				for _, occurrence := range item.Occurrences {
					gotReferenceIDs = append(gotReferenceIDs, occurrence.ReferenceID)
					if occurrence.Binding != "indeterminate" {
						t.Errorf("%s occurrence = %+v, want indeterminate", role, occurrence)
					}
				}
				if !reflect.DeepEqual(gotReferenceIDs, wantReferenceIDs) {
					t.Errorf("%s occurrence references = %v, want %v", role, gotReferenceIDs, wantReferenceIDs)
				}
			}

			for name, refined := range map[string]func() (*Result, error){
				"fields": func() (*Result, error) {
					got, err := AnalyzeWithSourceFields(tc.document, []string{"a", "host"})
					if err != nil {
						return nil, err
					}
					return got.Result, nil
				},
				"universe": func() (*Result, error) {
					got, err := AnalyzeWithSourceUniverse(tc.document, SourceUniverse{Fields: []string{"a", "host"}, Complete: true})
					if err != nil {
						return nil, err
					}
					return got.Result, nil
				},
			} {
				got, err := refined()
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(got.Requirements, result.Requirements) {
					t.Errorf("requirements changed under %s refinement:\nplain=%+v\nrefined=%+v", name, result.Requirements, got.Requirements)
				}
			}
		})
	}
}

func TestRequirementTraceKnowledgeObjects(t *testing.T) {
	for _, tc := range []struct {
		name     string
		document QueryDocument
		want     map[string]string
	}{
		{"search objects", QueryDocument{Text: `index=main source=access.log sourcetype=web`}, map[string]string{"index:main": "exact", "source:access.log": "exact", "sourcetype:web": "exact"}},
		{"lookup", QueryDocument{Text: `| inputlookup users`}, map[string]string{"lookup:users": "exact"}},
		{"data model", QueryDocument{Text: `| datamodel Web search`}, map[string]string{"data_model:Web": "exact"}},
		{"qualified data model", QueryDocument{Text: `| from datamodel:Network_Traffic.All_Traffic`}, map[string]string{"data_model:Network_Traffic": "exact", "dataset:Network_Traffic.All_Traffic": "exact"}},
		{"tstats objects", QueryDocument{Text: `| tstats count FROM datamodel=Network_Traffic.All_Traffic WHERE index=main source=access.log sourcetype=web`}, map[string]string{"data_model:Network_Traffic": "exact", "dataset:Network_Traffic.All_Traffic": "exact", "index:main": "exact", "source:access.log": "exact", "sourcetype:web": "exact"}},
		{"wildcard objects", QueryDocument{Text: `index=ma* source="access*"`}, map[string]string{"index:ma*": "wildcard", "source:access*": "wildcard"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, trace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			got := map[string]string{}
			for _, entry := range trace.references {
				if entry.reference.Kind == "field" {
					continue
				}
				key := entry.reference.Kind + ":" + entry.reference.NormalizedName
				got[key] = entry.reference.Resolution
				if entry.reference.Resolution == "exact" && (!entry.directExternal || entry.conditional) {
					t.Errorf("exact object is not direct and required: %+v", entry)
				}
				if entry.reference.Resolution != "exact" && (entry.directExternal || !entry.conditional) {
					t.Errorf("pattern object is not conditional: %+v", entry)
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("objects = %v, want %v", got, tc.want)
			}
		})
	}

	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "search host=web | `expand(host)` | stats count by user"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var macro *requirementTraceReference
	for i := range trace.references {
		if trace.references[i].reference.Kind == "macro" {
			macro = &trace.references[i]
			break
		}
	}
	if macro == nil || !macro.directExternal || macro.conditional {
		t.Fatalf("macro evidence = %+v", macro)
	}
	foundExpansion := false
	for _, diagnostic := range trace.diagnostics {
		if diagnostic.diagnostic.Code == CodeDynamicReference && reflect.DeepEqual(diagnostic.pendingReferenceIDs, []string{macro.reference.ID}) {
			foundExpansion = true
		}
	}
	if !foundExpansion {
		t.Fatalf("macro expansion diagnostic does not own %s: %+v", macro.reference.ID, trace.diagnostics)
	}
}

func TestTstatsOwnedFactsAreNotDuplicated(t *testing.T) {
	query := "| tstats `accelerated` sum(bytes) AS total FROM datamodel=Network_Traffic.All_Traffic WHERE index=main All_Traffic.action=allowed BY All_Traffic.src"
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantReferences := []string{
		"accelerated=>accelerated:macro:read:not_applicable:exact",
		"bytes=>bytes:field:read:source:exact",
		"total=>total:field:output:not_applicable:exact",
		"Network_Traffic=>Network_Traffic:data_model:read:not_applicable:exact",
		"Network_Traffic.All_Traffic=>Network_Traffic.All_Traffic:dataset:read:not_applicable:exact",
		"main=>main:index:read:not_applicable:exact",
		"All_Traffic.action=>All_Traffic.action:field:filter:source:exact",
		"All_Traffic.src=>All_Traffic.src:field:group:source:exact",
	}
	if got := tstatsReferenceFacts(result.References); !reflect.DeepEqual(got, wantReferences) {
		t.Fatalf("owned references\n got: %q\nwant: %q", got, wantReferences)
	}
	if len(trace.references) != len(result.References) {
		t.Fatalf("trace references = %d, public references = %d", len(trace.references), len(result.References))
	}
	seenTrace := map[string]int{}
	for _, entry := range trace.references {
		seenTrace[entry.reference.ID]++
	}
	for _, reference := range result.References {
		if seenTrace[reference.ID] != 1 {
			t.Errorf("reference %s trace occurrences = %d", reference.ID, seenTrace[reference.ID])
		}
	}
	wantRequirements := []string{
		"req-1:macro:accelerated:read:required:exact:ref-0:accelerated:not_applicable",
		"req-2:field:bytes:read:required:exact:ref-1:bytes:source",
		"req-3:data_model:Network_Traffic:read:required:exact:ref-3:Network_Traffic:not_applicable",
		"req-4:dataset:Network_Traffic.All_Traffic:read:required:exact:ref-4:Network_Traffic.All_Traffic:not_applicable",
		"req-5:index:main:read:required:exact:ref-5:main:not_applicable",
		"req-6:field:All_Traffic.action:filter:required:exact:ref-6:All_Traffic.action:source",
		"req-7:field:All_Traffic.src:group:required:exact:ref-7:All_Traffic.src:source",
	}
	if got := tstatsRequirementFacts(result.Requirements.Items); !reflect.DeepEqual(got, wantRequirements) {
		t.Fatalf("owned requirement occurrences\n got: %q\nwant: %q", got, wantRequirements)
	}
	if !reflect.DeepEqual(result.Dependencies.DataModels, []string{"Network_Traffic"}) || !reflect.DeepEqual(result.Dependencies.Datasets, []string{"Network_Traffic.All_Traffic"}) || !reflect.DeepEqual(result.Dependencies.Indexes, []string{"main"}) || !reflect.DeepEqual(result.Dependencies.Macros, []string{"accelerated"}) {
		t.Fatalf("owned dependencies were duplicated or omitted: %+v", result.Dependencies)
	}
}

func TestRequirementTraceDiagnosticOwnership(t *testing.T) {
	for _, tc := range []struct {
		name      string
		document  QueryDocument
		code      string
		ownedKind string
	}{
		{"wildcard", QueryDocument{Text: "table *"}, CodeUnresolvedWildcard, "field"},
		{"unknown command", QueryDocument{Text: "| mystery"}, CodeUnsupportedCommand, ""},
		{"unsupported semantics", QueryDocument{Text: "| datamodel Web search"}, CodeUnsupportedSemantics, ""},
		{"syntax error", QueryDocument{Text: "search host="}, CodeSyntaxError, ""},
		{"macro expansion", QueryDocument{Text: "| `expand()`"}, CodeDynamicReference, "macro"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, trace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			var got *requirementTraceDiagnostic
			for i := range trace.diagnostics {
				if trace.diagnostics[i].diagnostic.Code == tc.code {
					got = &trace.diagnostics[i]
					break
				}
			}
			if got == nil {
				t.Fatalf("missing %s in %+v", tc.code, trace.diagnostics)
			}
			if tc.ownedKind == "" {
				if len(got.pendingReferenceIDs) != 0 {
					t.Fatalf("unowned diagnostic has links %v", got.pendingReferenceIDs)
				}
			} else {
				if len(got.pendingReferenceIDs) != 1 {
					t.Fatalf("owned diagnostic links = %v", got.pendingReferenceIDs)
				}
				owned := false
				for _, entry := range trace.references {
					if entry.reference.ID == got.pendingReferenceIDs[0] && entry.reference.Kind == tc.ownedKind {
						owned = true
					}
				}
				if !owned {
					t.Fatalf("diagnostic owner %v is not a %s reference", got.pendingReferenceIDs, tc.ownedKind)
				}
			}
			for i := 1; i < len(trace.diagnostics); i++ {
				if trace.diagnostics[i-1].eventOrdinal >= trace.diagnostics[i].eventOrdinal {
					t.Fatalf("diagnostic event order is not increasing: %+v", trace.diagnostics)
				}
			}
		})
	}
}

func TestRequirementTraceRenameRemovesSource(t *testing.T) {
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "search host=x | rename host AS node | table host"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := trace.reference("pending-3")
	if entry.reference.NormalizedName != "host" || entry.reference.Role != "read" || entry.reference.Binding != "unavailable" || entry.directExternal || entry.conditional {
		t.Fatalf("post-rename source trace = %+v", entry)
	}
}

func TestRequirementTraceConflictingRenameInvalidatesQueryOnlyFields(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "eval a=1, b=2 | rename a AS b | table a"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	renameState := result.Lineage[1].After
	if len(renameState.Fields) != 0 || !renameState.Uncertain {
		t.Fatalf("public post-rename state = %+v, want no fields and uncertain", renameState)
	}
	var entry *requirementTraceReference
	for i := range trace.references {
		candidate := &trace.references[i]
		if candidate.reference.NormalizedName == "a" && candidate.reference.Role == "read" && candidate.reference.StageID == "stage-2" {
			entry = candidate
		}
	}
	if entry == nil {
		t.Fatalf("missing query-only post-conflict read: %+v", trace.references)
	}
	if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
		t.Fatalf("query-only post-conflict read = %+v, want conditional indeterminate evidence", entry)
	}
}

func TestRequirementTraceWildcardRenameInvalidatesConcreteTargets(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{
		Text:     `FROM [{'old_a':1,'new_a':2}] | rename 'old_*' AS 'new_*' | table new_a`,
		Language: "spl2",
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	renameState := result.Lineage[1].After
	for _, field := range renameState.Fields {
		if field.Name == "old_a" || field.Name == "new_a" {
			t.Fatalf("public post-rename state retained affected field: %+v", renameState)
		}
	}
	if !renameState.Uncertain {
		t.Fatalf("public post-rename state = %+v, want uncertain invalidation", renameState)
	}

	var publicRead *Reference
	for i := range result.References {
		candidate := &result.References[i]
		if candidate.NormalizedName == "new_a" && candidate.Role == "read" && candidate.StageID == "stage-2" {
			publicRead = candidate
		}
	}
	if publicRead == nil {
		t.Fatalf("missing public post-rename read: %+v", result.References)
	}
	var traceRead *requirementTraceReference
	for i := range trace.references {
		candidate := &trace.references[i]
		if candidate.reference.ID == publicRead.ID {
			traceRead = candidate
		}
	}
	if traceRead == nil {
		t.Fatalf("missing query-only counterpart for public read: public=%+v trace=%+v", publicRead, trace.references)
	}
	if publicRead.Binding != "indeterminate" || traceRead.reference.Binding != publicRead.Binding || traceRead.directExternal || !traceRead.conditional {
		t.Fatalf("post-rename reads diverged: public=%+v query-only=%+v", publicRead, traceRead)
	}
}

func TestRequirementTraceWildcardRenameInvalidatesStructuralMatches(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{
		Text:     `FROM main | eval x=actor.name | rename 'actor.*' AS 'changed.*' | eval y=actor.name`,
		Language: "spl2",
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	renameState := result.Lineage[2].After
	for _, field := range renameState.Fields {
		if field.Name == "actor.name" {
			t.Fatalf("wildcard rename retained structural binding: %+v", renameState)
		}
	}

	var later *Reference
	for i := range result.References {
		candidate := &result.References[i]
		if candidate.OriginalName == "actor.name" && candidate.Role == "read" && candidate.StageID == "stage-3" {
			later = candidate
		}
	}
	if later == nil || later.Binding != "indeterminate" {
		t.Fatalf("post-rename structural read = %+v, want indeterminate", later)
	}
	entry := requirementTraceReferencesByID(trace)[later.ID]
	if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
		t.Fatalf("post-rename structural requirement = %+v, want conditional indeterminate", entry)
	}

	final := result.Lineage[3]
	var y *FieldBinding
	for i := range final.After.Fields {
		if final.After.Fields[i].Name == "y" {
			y = &final.After.Fields[i]
		}
	}
	if y == nil || !y.Conditional {
		t.Fatalf("post-rename y binding = %+v, want conditional", y)
	}
	for _, transition := range final.Transitions {
		if transition.Output == "y" && !transition.Conditional {
			t.Fatalf("stale structural binding produced unconditional y: %+v", transition)
		}
	}
	item := requirementItem(result.Requirements, "field", "actor.name", "read")
	if item == nil {
		t.Fatalf("missing actor.name requirement: %+v", result.Requirements.Items)
	}
	foundLater := false
	for _, occurrence := range item.Occurrences {
		if occurrence.ReferenceID == later.ID {
			foundLater = true
			if occurrence.Binding != "indeterminate" {
				t.Fatalf("post-rename requirement occurrence = %+v", occurrence)
			}
		}
	}
	if !foundLater {
		t.Fatalf("missing post-rename occurrence in %+v", item)
	}
}

func TestRequirementTraceWildcardRenamePreservesClassicAtomicInvalidation(t *testing.T) {
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: `search a1=1 | rename a* AS b | eval y=a1`}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var later *Reference
	for i := range result.References {
		candidate := &result.References[i]
		if candidate.NormalizedName == "a1" && candidate.Role == "read" && candidate.StageID == "stage-2" {
			later = candidate
		}
	}
	if later == nil || later.Binding != "indeterminate" {
		t.Fatalf("classic post-rename read = %+v, want indeterminate", later)
	}
	entry := requirementTraceReferencesByID(trace)[later.ID]
	if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
		t.Fatalf("classic post-rename requirement = %+v, want conditional indeterminate", entry)
	}
	foundY := false
	for _, field := range result.Lineage[2].After.Fields {
		if field.Name == "y" {
			foundY = true
			if field.Conditional {
				t.Fatalf("classic wildcard rename assignment changed conditionality: %+v", field)
			}
		}
	}
	if !foundY {
		t.Fatalf("classic wildcard rename lost y: %+v", result.Lineage[2].After)
	}
}

func TestRequirementTraceDatasetLiteralClosesAbsentField(t *testing.T) {
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "FROM [{a:1}] | table b", Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := trace.reference("pending-1")
	if entry.reference.NormalizedName != "b" || entry.reference.Role != "read" || entry.reference.Binding != "unavailable" || entry.directExternal || entry.conditional {
		t.Fatalf("absent dataset field trace = %+v", entry)
	}
}

func TestRequirementTracePreservesConditionalFields(t *testing.T) {
	for _, tc := range []struct {
		name               string
		document           QueryDocument
		field, binding     string
		requireConditional bool
	}{
		{"SPL OUTPUTNEW", QueryDocument{Text: "search user=* | lookup users user OUTPUTNEW role | table role"}, "role", "indeterminate", true},
		{"SPL2 partial dataset field", QueryDocument{Text: "FROM [{a:1},{b:2}] | table a", Language: "spl2"}, "a", "indeterminate", true},
		{"SPL2 selected bin effect", QueryDocument{Text: "FROM main | eval a=host | bin a | table a", Language: "spl2"}, "a", "derived", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, trace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}

			var public *Reference
			var traced *requirementTraceReference
			for i := range result.References {
				reference := &result.References[i]
				if reference.NormalizedName == tc.field && reference.Role == "read" {
					public = reference
				}
			}
			for i := range trace.references {
				entry := &trace.references[i]
				if entry.reference.NormalizedName == tc.field && entry.reference.Role == "read" {
					traced = entry
				}
			}
			if public == nil || traced == nil {
				t.Fatalf("missing final %s read: public=%+v trace=%+v", tc.field, result.References, trace.references)
			}
			if public.Binding != tc.binding {
				t.Fatalf("public final read = %+v, want %s", public, tc.binding)
			}
			if traced.reference.Binding != tc.binding || traced.directExternal || traced.conditional != tc.requireConditional {
				t.Fatalf("query-only final read = %+v, want binding=%s conditional=%v", traced, tc.binding, tc.requireConditional)
			}
		})
	}
}

func TestRequirementTraceSQLProjectionClosesAbsentDownstreamField(t *testing.T) {
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "SELECT host FROM main | where bytes>0", Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := trace.reference("pending-2")
	if entry.reference.NormalizedName != "bytes" || entry.reference.Role != "read" || entry.reference.Binding != "unavailable" || entry.directExternal || entry.conditional {
		t.Fatalf("post-SELECT field trace = %+v", entry)
	}
}

func TestRequirementTraceSPLRecoveryMirrorsUncertainty(t *testing.T) {
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "search host= | table user"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := requirementTraceReferenceByNameAndRole(t, trace, "user", "read")
	if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
		t.Fatalf("post-recovery field trace = %+v", entry)
	}
	if trace.syntaxComplete || trace.semanticComplete {
		t.Fatalf("recovered SPL trace completeness = syntax %t semantic %t", trace.syntaxComplete, trace.semanticComplete)
	}
}

func TestSPL2RecoveryRequirementsRetainEveryQueryDiagnostic(t *testing.T) {
	const query = "FROM main | foobar | stats c=count() BY host | eval y=other"
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query, Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	wantDiagnostics := []struct {
		code, severity, category, message string
		start, end                        int
		stage                             string
	}{
		{CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Standalone command effects are unproved", 12, 18, "stage-1"},
		{CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Standalone command syntax and effects are unproved", 12, 18, "stage-1"},
		{CodeUnsupportedSemantics, "warning", "unsupported_semantics", "Standalone command effects are unproved", 21, 44, "stage-2"},
		{CodeUnsupportedSemantics, "warning", "unsupported", "Unproved command option retains incomplete syntax coverage", 27, 34, "stage-2"},
		{CodeSyntaxError, "error", "syntax", "", 34, 35, "stage-2"},
	}
	if result.Status != Invalid || result.Requirements.QueryStatus != Invalid {
		t.Fatalf("status = analysis %q requirements %q, want invalid/invalid", result.Status, result.Requirements.QueryStatus)
	}
	if len(result.Diagnostics) != len(wantDiagnostics) {
		t.Fatalf("analysis diagnostics = %d, want %d: %+v", len(result.Diagnostics), len(wantDiagnostics), result.Diagnostics)
	}
	for i, want := range wantDiagnostics {
		got := result.Diagnostics[i]
		if got.Code != want.code || got.Severity != want.severity || got.Category != want.category || got.Location != newSourceIndex(query).location(want.start, want.end) || got.StageID != want.stage || got.ScopeID != "scope-0" {
			t.Errorf("diagnostic %d = %+v, want code=%q severity=%q category=%q range=%d:%d stage=%q", i, got, want.code, want.severity, want.category, want.start, want.end, want.stage)
		}
		if want.message != "" && got.Message != want.message {
			t.Errorf("diagnostic %d message = %q, want %q", i, got.Message, want.message)
		}
		if want.code == CodeSyntaxError && !strings.HasPrefix(got.Message, "extraneous input '('") {
			t.Errorf("syntax diagnostic message = %q", got.Message)
		}
	}
	if !reflect.DeepEqual(result.Requirements.Diagnostics, result.Diagnostics) {
		t.Fatalf("requirement diagnostics differ from plain analysis:\nanalysis=%+v\nrequirements=%+v", result.Diagnostics, result.Requirements.Diagnostics)
	}
	standalone, err := Requirements(QueryDocument{Text: query, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(standalone, &result.Requirements) {
		t.Fatalf("standalone requirements differ from embedded requirements:\nstandalone=%+v\nembedded=%+v", standalone, result.Requirements)
	}
	wantGapCodes := []string{CodeUnsupportedSemantics, CodeSyntaxError, CodeRequirementIndeterminate}
	gotGapCodes := make([]string, len(result.Requirements.Gaps))
	for i, gap := range result.Requirements.Gaps {
		gotGapCodes[i] = gap.Code
	}
	if !reflect.DeepEqual(gotGapCodes, wantGapCodes) {
		t.Fatalf("requirement gaps = %q, want %q: %+v", gotGapCodes, wantGapCodes, result.Requirements.Gaps)
	}

	wantTraceOrder := []struct {
		code, message string
		event         int
	}{
		{CodeUnsupportedSemantics, "Standalone command syntax and effects are unproved", 0},
		{CodeSyntaxError, "", 1},
		{CodeUnsupportedSemantics, "Unproved command option retains incomplete syntax coverage", 2},
		{CodeUnsupportedSemantics, "Standalone command effects are unproved", 4},
		{CodeUnsupportedSemantics, "Standalone command effects are unproved", 5},
	}
	if len(trace.diagnostics) != len(wantTraceOrder) {
		t.Fatalf("trace diagnostics = %d, want %d: %+v", len(trace.diagnostics), len(wantTraceOrder), trace.diagnostics)
	}
	for i, want := range wantTraceOrder {
		got := trace.diagnostics[i]
		if got.diagnostic.Code != want.code || !got.incomplete || len(got.pendingReferenceIDs) != 0 || got.eventOrdinal != want.event {
			t.Errorf("trace diagnostic %d = %+v, want code=%q incomplete=true owners=[] event=%d", i, got, want.code, want.event)
		}
		if want.message != "" && got.diagnostic.Message != want.message {
			t.Errorf("trace diagnostic %d message = %q, want %q", i, got.diagnostic.Message, want.message)
		}
	}
}

func TestAggregationClosesQueryOnlyUncertaintyAtCurrentStage(t *testing.T) {
	const query = "search index=main | foobar | stats count by host | eval y=other"
	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	host := requirementTraceReferenceByNameAndRole(t, trace, "host", "group")
	if host.reference.Binding != "indeterminate" || host.directExternal || !host.conditional {
		t.Fatalf("group field trace = %+v, want conditional indeterminate", host)
	}
	other := requirementTraceReferenceByNameAndRole(t, trace, "other", "read")
	if other.reference.Binding != "unavailable" || other.directExternal || other.conditional {
		t.Fatalf("post-aggregation field trace = %+v, want unavailable", other)
	}
	if result.Status != Invalid || result.Requirements.QueryStatus != Invalid {
		t.Fatalf("status = analysis %q requirements %q, want invalid/invalid", result.Status, result.Requirements.QueryStatus)
	}
	standalone, err := Requirements(QueryDocument{Text: query})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(standalone, &result.Requirements) {
		t.Fatalf("standalone requirements differ from embedded requirements:\nstandalone=%+v\nembedded=%+v", standalone, result.Requirements)
	}
	foundUnavailable := false
	for _, diagnostic := range result.Requirements.Diagnostics {
		if diagnostic.Code == CodeUnavailableField && diagnostic.Location == newSourceIndex(query).location(58, 63) {
			foundUnavailable = true
		}
	}
	if !foundUnavailable {
		t.Fatalf("requirements omitted unavailable-field diagnostic: %+v", result.Requirements.Diagnostics)
	}
	for _, item := range result.Requirements.Items {
		if item.Kind == "field" && item.Identity == "other" {
			t.Fatalf("unavailable local field became an external requirement: %+v", item)
		}
	}
	for _, gap := range result.Requirements.Gaps {
		for _, referenceID := range gap.ReferenceIDs {
			if referenceID == other.reference.ID {
				t.Fatalf("unavailable local field became an indeterminate gap: %+v", gap)
			}
		}
	}
}

func TestRequirementTraceSQLHiddenFieldOwnsIncompleteEvidence(t *testing.T) {
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "SELECT marker FROM main GROUP BY marker HAVING hidden=1", Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	entry := requirementTraceReferenceByNameAndRole(t, trace, "hidden", "read")
	if entry.reference.Binding != "indeterminate" || entry.directExternal || !entry.conditional {
		t.Fatalf("hidden SQL field trace = %+v", entry)
	}
	if trace.semanticComplete {
		t.Fatal("hidden SQL field left the query-only trace semantically complete")
	}
	foundOwnedEvidence := false
	for _, diagnostic := range trace.diagnostics {
		if diagnostic.diagnostic.Code == CodeUnsupportedSemantics && diagnostic.incomplete && reflect.DeepEqual(diagnostic.pendingReferenceIDs, []string{entry.reference.ID}) {
			foundOwnedEvidence = true
		}
	}
	if !foundOwnedEvidence {
		t.Fatalf("hidden SQL field %q has no owned incomplete evidence: %+v", entry.reference.ID, trace.diagnostics)
	}
}

func requirementTraceReferenceByNameAndRole(t *testing.T, trace *requirementTrace, name, role string) *requirementTraceReference {
	t.Helper()
	for i := range trace.references {
		entry := &trace.references[i]
		if entry.reference.NormalizedName == name && entry.reference.Role == role {
			return entry
		}
	}
	t.Fatalf("missing trace reference %s/%s: %+v", name, role, trace.references)
	return nil
}

func assertFieldCommandRequirementTrace(t *testing.T, trace *requirementTrace, stageID, name, role, binding string, directExternal, conditional bool) {
	t.Helper()
	for i := range trace.references {
		entry := &trace.references[i]
		if entry.reference.StageID != stageID || entry.reference.NormalizedName != name || entry.reference.Role != role {
			continue
		}
		if entry.reference.Binding != binding || entry.directExternal != directExternal || entry.conditional != conditional {
			t.Fatalf("field-command requirement trace = %+v, want binding=%s direct=%t conditional=%t", entry, binding, directExternal, conditional)
		}
		return
	}
	t.Fatalf("missing field-command trace %s %s/%s: %+v", stageID, name, role, trace.references)
}

func TestRequirementTraceMacroDiagnosticsOwnExactReferences(t *testing.T) {
	const query = "| eval a=`one()`, b=`two()`"
	_, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: query}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	macroReferences := map[string]string{}
	for _, entry := range trace.references {
		if entry.reference.Kind == "macro" {
			macroReferences[entry.reference.NormalizedName] = entry.reference.ID
		}
	}
	wantOwners := map[string]string{"`one()`": "one", "`two()`": "two"}
	found := 0
	for _, entry := range trace.diagnostics {
		if entry.diagnostic.Code != CodeDynamicReference {
			continue
		}
		found++
		span := query[entry.diagnostic.Location.Start.Offset:entry.diagnostic.Location.End.Offset]
		name, ok := wantOwners[span]
		if !ok {
			t.Fatalf("unexpected macro diagnostic span %q", span)
		}
		if !reflect.DeepEqual(entry.pendingReferenceIDs, []string{macroReferences[name]}) {
			t.Fatalf("macro %q diagnostic owners = %v, want %q", name, entry.pendingReferenceIDs, macroReferences[name])
		}
	}
	if found != len(wantOwners) {
		t.Fatalf("macro diagnostic count = %d, want %d: %+v", found, len(wantOwners), trace.diagnostics)
	}

	for _, tc := range []struct{ name, query, macro string }{
		{"macro-only stage", "| `standalone()`", "standalone"},
		{"tstats inline macro", "| tstats `accelerated` count FROM datamodel=Authentication.Authentication BY Authentication.user", "accelerated"},
		{"macro next to stats", "search stable=* | `expand(invented)` | stats count by stable", "expand"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, siblingTrace, err := analyzeRewriteWithTrace(QueryDocument{Text: tc.query}, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			var macro *requirementTraceReference
			macroCount := 0
			for i := range siblingTrace.references {
				entry := &siblingTrace.references[i]
				if entry.reference.Kind != "macro" {
					continue
				}
				macro, macroCount = entry, macroCount+1
			}
			if macroCount != 1 || macro.reference.NormalizedName != tc.macro || !macro.directExternal || macro.conditional {
				t.Fatalf("macro trace = %+v count=%d", macro, macroCount)
			}
			if got := tc.query[macro.reference.Location.Start.Offset:macro.reference.Location.End.Offset]; got != tc.macro {
				t.Fatalf("macro trace source = %q, want %q", got, tc.macro)
			}
			owned := 0
			for _, diagnostic := range siblingTrace.diagnostics {
				if diagnostic.diagnostic.Code == CodeDynamicReference && diagnostic.incomplete && reflect.DeepEqual(diagnostic.pendingReferenceIDs, []string{macro.reference.ID}) {
					owned++
					if got := tc.query[diagnostic.diagnostic.Location.Start.Offset:diagnostic.diagnostic.Location.End.Offset]; !strings.Contains(got, tc.macro) {
						t.Fatalf("macro diagnostic source = %q", got)
					}
				}
			}
			if owned != 1 || len(result.Dependencies.Macros) != 1 {
				t.Fatalf("owned macro evidence = %d dependencies=%v diagnostics=%+v", owned, result.Dependencies.Macros, siblingTrace.diagnostics)
			}
		})
	}
}

func TestRequirementTraceRemapsEachPendingIDOnce(t *testing.T) {
	assertPanics := func(t *testing.T, run func()) {
		t.Helper()
		defer func() {
			if recover() == nil {
				t.Fatal("missing remap invariant did not panic")
			}
		}()
		run()
	}
	t.Run("duplicate pending ID", func(t *testing.T) {
		trace := newRequirementTrace()
		reference := Reference{ID: "pending-0", OriginReferenceIDs: []string{}}
		trace.recordReference(reference, false, false, trace.nextEvent())
		trace.recordReference(reference, false, false, trace.nextEvent())
		assertPanics(t, func() { trace.remapReferences(map[string]string{"pending-0": "ref-0"}) })
	})
	t.Run("dangling diagnostic owner", func(t *testing.T) {
		trace := newRequirementTrace()
		trace.recordDiagnostic(Diagnostic{Code: CodeUnsupportedSemantics}, true, []string{"pending-missing"}, trace.nextEvent())
		assertPanics(t, func() { trace.remapReferences(map[string]string{}) })
	})

	result, trace, err := analyzeRewriteWithTrace(QueryDocument{Text: "search [ table child ] host=x | eval copied=host"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	public := map[string]Reference{}
	for _, reference := range result.References {
		public[reference.ID] = reference
	}
	seenPending := map[string]bool{}
	for _, entry := range trace.references {
		if seenPending[entry.pendingID] {
			t.Fatalf("pending ID %q was remapped more than once", entry.pendingID)
		}
		seenPending[entry.pendingID] = true
		counterpart, ok := public[entry.reference.ID]
		if !ok {
			t.Fatalf("trace reference %q has no public counterpart", entry.reference.ID)
		}
		if counterpart.NormalizedName != entry.reference.NormalizedName || counterpart.Kind != entry.reference.Kind || counterpart.Role != entry.reference.Role || counterpart.StageID != entry.reference.StageID || counterpart.ScopeID != entry.reference.ScopeID || counterpart.Location != entry.reference.Location {
			t.Fatalf("trace reference differs from public counterpart: trace %+v public %+v", entry.reference, counterpart)
		}
	}
	if trace.references[0].pendingID != "pending-0" || trace.references[0].reference.ID == "ref-0" {
		t.Fatalf("test did not exercise pending versus source-order remapping: %+v", trace.references)
	}

	result, trace, err = analyzeRewriteWithTrace(QueryDocument{Text: "FROM orders AS o WHERE EXISTS (SELECT id FROM inventory WHERE id=o.item) SELECT item", Language: "spl2"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(trace.references) != len(result.References) {
		t.Fatalf("SPL2 child scope references were not all traced: trace %d public %d", len(trace.references), len(result.References))
	}
	for _, entry := range trace.references {
		if entry.reference.StageID == "" {
			t.Fatalf("SPL2 reference lost its finalized lexical stage: %+v", entry)
		}
	}
}

func TestRequirementTraceRefinementParity(t *testing.T) {
	admitted := func(string) SourceFieldAdmission { return SourceFieldAdmitted }
	for _, tc := range []struct {
		name       string
		document   QueryDocument
		refinement *sourceRefinement
		renameRefs bool
	}{
		{"finite field list", QueryDocument{Text: "fields host* | table hostname"}, testSourceRefinement([]string{"hostname"}, true, nil, true), false},
		{"partial source universe", QueryDocument{Text: "fields host* | table hostname"}, testSourceRefinement([]string{"hostname"}, false, nil, false), false},
		{"wildcard selectors", QueryDocument{Text: "eval label=host | table *"}, testSourceRefinement([]string{"host"}, true, nil, false), false},
		{"unsupported wildcard selector", QueryDocument{Text: "sort host*"}, testSourceRefinement([]string{"host"}, true, nil, false), false},
		{"wildcard rename conflicts", QueryDocument{Text: "search alpha=1 | rename a* AS b | table alpha"}, testSourceRefinement([]string{}, true, nil, false), true},
		{"rename existing destination", QueryDocument{Text: "search alpha=1 | fields a* | rename x AS alpha | table alpha"}, testSourceRefinement([]string{}, true, nil, false), true},
		{"dotted SPL2 names", QueryDocument{Text: `SELECT 'actor.name' FROM main`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 downstream derived", QueryDocument{Text: `FROM main | eval local='actor.name' | table local`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 rename output", QueryDocument{Text: `FROM main | rename 'actor.name' AS actor | table actor`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 aggregate output", QueryDocument{Text: `FROM main | stats count('actor.name') AS total | table total`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 lookup output", QueryDocument{Text: `FROM main | lookup users 'actor.name' OUTPUT role | table role`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 downstream SPL source", QueryDocument{Text: `FROM main | eval local='actor.name' | where other=1`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
		{"dotted SPL2 downstream SQL source", QueryDocument{Text: `FROM main WHERE 'actor.name'=1 SELECT other`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			plainResult, plainTrace, err := analyzeRewriteWithTrace(tc.document, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			refinedResult, refinedTrace, err := analyzeRewriteWithTrace(tc.document, tc.refinement, nil)
			if err != nil {
				t.Fatal(err)
			}
			plainTraceJSON := marshalRequirementTrace(t, plainTrace)
			refinedTraceJSON := marshalRequirementTrace(t, refinedTrace)
			if string(plainTraceJSON) != string(refinedTraceJSON) {
				t.Fatalf("trace changed under refinement:\n plain public: %+v\nrefined public: %+v\n plain: %s\nrefined: %s", plainResult.References, refinedResult.References, plainTraceJSON, refinedTraceJSON)
			}
			if tc.renameRefs {
				assertReferenceListParity(t, plainResult.References, refinedResult.References)
			}
			plainJSON, _ := json.Marshal(plainResult)
			refinedJSON, _ := json.Marshal(refinedResult)
			if string(plainJSON) == string(refinedJSON) {
				t.Fatal("refinement parity row did not change the public result")
			}
		})
	}
}

func assertReferenceListParity(t *testing.T, plain, refined []Reference) {
	t.Helper()
	if len(plain) != len(refined) {
		t.Fatalf("public reference counts differ: plain=%+v refined=%+v", plain, refined)
	}
	for i := range plain {
		a, b := plain[i], refined[i]
		if a.ID != b.ID || a.OriginalName != b.OriginalName || a.NormalizedName != b.NormalizedName || a.Kind != b.Kind || a.Role != b.Role || a.StageID != b.StageID || a.ScopeID != b.ScopeID || a.Location != b.Location || a.Resolution != b.Resolution {
			t.Fatalf("public reference %d changed shape: plain=%+v refined=%+v", i, a, b)
		}
	}
}

func testSourceRefinement(fields []string, complete bool, resolve func(string) SourceFieldAdmission, finite bool) *sourceRefinement {
	members := map[string]bool{}
	for _, field := range fields {
		members[field] = true
	}
	sorted := append([]string{}, fields...)
	sort.Strings(sorted)
	return &sourceRefinement{names: sorted, members: members, complete: complete, resolve: resolve, finiteCompatibility: finite, expansions: []FieldExpansion{}}
}

func marshalRequirementTrace(t *testing.T, trace *requirementTrace) []byte {
	t.Helper()
	type referenceJSON struct {
		PendingID       string    `json:"pending_id"`
		Reference       Reference `json:"reference"`
		DirectExternal  bool      `json:"direct_external"`
		Conditional     bool      `json:"conditional"`
		PathConditional bool      `json:"path_conditional"`
		EventOrdinal    int       `json:"event_ordinal"`
	}
	type diagnosticJSON struct {
		Diagnostic          Diagnostic `json:"diagnostic"`
		Incomplete          bool       `json:"incomplete"`
		PendingReferenceIDs []string   `json:"pending_reference_ids"`
		EventOrdinal        int        `json:"event_ordinal"`
	}
	wire := struct {
		References       []referenceJSON  `json:"references"`
		Diagnostics      []diagnosticJSON `json:"diagnostics"`
		SyntaxComplete   bool             `json:"syntax_complete"`
		SemanticComplete bool             `json:"semantic_complete"`
	}{
		References:       []referenceJSON{},
		Diagnostics:      []diagnosticJSON{},
		SyntaxComplete:   trace.syntaxComplete,
		SemanticComplete: trace.semanticComplete,
	}
	for _, entry := range trace.references {
		wire.References = append(wire.References, referenceJSON{entry.pendingID, entry.reference, entry.directExternal, entry.conditional, entry.pathConditional, entry.eventOrdinal})
	}
	for _, entry := range trace.diagnostics {
		wire.Diagnostics = append(wire.Diagnostics, diagnosticJSON{entry.diagnostic, entry.incomplete, entry.pendingReferenceIDs, entry.eventOrdinal})
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func mergeTraceReference(id, identity string, offset int) Reference {
	return Reference{
		ID:                 id,
		OriginalName:       identity,
		NormalizedName:     identity,
		Kind:               "dataset",
		Role:               "read",
		StageID:            "stage-0",
		ScopeID:            "scope-0",
		Location:           Location{Start: Position{Offset: offset}, End: Position{Offset: offset + 1}},
		Resolution:         "exact",
		Binding:            "not_applicable",
		OriginReferenceIDs: []string{},
	}
}

func appendMergeTraceReference(trace *requirementTrace, id, identity string, offset int) {
	trace.recordReference(mergeTraceReference(id, identity, offset), true, false, trace.nextEvent())
}

func TestRequirementTraceForkKeepsSharedPrefixImmutableAndSuffixesIndependent(t *testing.T) {
	base := newRequirementTrace()
	appendMergeTraceReference(base, "pending-prefix", "prefix", 1)
	left := base.forkBranch()
	right := base.forkBranch()
	appendMergeTraceReference(left, "pending-left", "left", 2)
	appendMergeTraceReference(right, "pending-right", "right", 3)

	if got := []string{base.references[0].pendingID}; len(base.references) != 1 || !reflect.DeepEqual(got, []string{"pending-prefix"}) {
		t.Fatalf("base trace changed after forks: %+v", base.references)
	}
	if len(left.references) != 2 || left.references[1].pendingID != "pending-left" || len(right.references) != 2 || right.references[1].pendingID != "pending-right" {
		t.Fatalf("fork suffixes leaked: left=%+v right=%+v", left.references, right.references)
	}
}

func TestRequirementTraceMergeMakesEveryPathObligationRequired(t *testing.T) {
	base := newRequirementTrace()
	left := base.forkBranch()
	right := base.forkBranch()
	appendMergeTraceReference(left, "pending-left", "shared", 10)
	appendMergeTraceReference(right, "pending-right", "shared", 20)

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: left, Reachable: true},
		{Ordinal: 1, Trace: right, Reachable: true},
	})
	if len(merged.references) != 2 {
		t.Fatalf("merged references = %+v", merged.references)
	}
	for _, entry := range merged.references {
		if !entry.directExternal || entry.conditional || entry.pathConditional {
			t.Fatalf("every-path obligation was not required: %+v", entry)
		}
	}
	projected := mustProjectRequirements(t, merged)
	if !projected.Coverage.Complete || len(projected.Gaps) != 0 || len(projected.Items) != 1 || projected.Items[0].Necessity != "required" || len(projected.Items[0].Occurrences) != 2 {
		t.Fatalf("every-path projection = coverage %+v items %+v gaps %+v", projected.Coverage, projected.Items, projected.Gaps)
	}
}

func TestRequirementTraceMergeMakesSubsetObligationConditional(t *testing.T) {
	base := newRequirementTrace()
	left := base.forkBranch()
	right := base.forkBranch()
	appendMergeTraceReference(left, "pending-left", "left-only", 10)

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: left, Reachable: true},
		{Ordinal: 1, Trace: right, Reachable: true},
	})
	if len(merged.references) != 1 || merged.references[0].directExternal || merged.references[0].conditional || !merged.references[0].pathConditional {
		t.Fatalf("subset obligation = %+v", merged.references)
	}
	projected := mustProjectRequirements(t, merged)
	if !projected.Coverage.Complete || len(projected.Gaps) != 0 || len(projected.Items) != 1 || projected.Items[0].Necessity != "conditional" || len(projected.Items[0].Occurrences) != 1 {
		t.Fatalf("subset projection = coverage %+v items %+v gaps %+v", projected.Coverage, projected.Items, projected.Gaps)
	}
}

func TestRequirementTraceMergeSupportsNestedForks(t *testing.T) {
	base := newRequirementTrace()
	outer := base.forkBranch()
	appendMergeTraceReference(outer, "pending-outer", "outer", 5)
	innerLeft := outer.forkBranch()
	innerRight := outer.forkBranch()
	appendMergeTraceReference(innerLeft, "pending-inner-left", "inner", 10)
	appendMergeTraceReference(innerRight, "pending-inner-right", "inner", 20)
	mergedInner := mergeRequirementTraces(outer, []requirementTracePath{
		{Ordinal: 0, Trace: innerLeft, Reachable: true},
		{Ordinal: 1, Trace: innerRight, Reachable: true},
	})
	otherOuter := base.forkBranch()
	appendMergeTraceReference(otherOuter, "pending-other", "other", 30)

	mergedOuter := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: mergedInner, Reachable: true},
		{Ordinal: 1, Trace: otherOuter, Reachable: true},
	})
	want := []string{"pending-outer", "pending-inner-left", "pending-inner-right", "pending-other"}
	got := make([]string, 0, len(mergedOuter.references))
	for _, entry := range mergedOuter.references {
		got = append(got, entry.pendingID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("nested merged order = %v want %v", got, want)
	}
	if mergedOuter.references[1].directExternal || mergedOuter.references[1].conditional || !mergedOuter.references[1].pathConditional || mergedOuter.references[2].directExternal || mergedOuter.references[2].conditional || !mergedOuter.references[2].pathConditional {
		t.Fatalf("nested subset necessity was lost: %+v", mergedOuter.references)
	}
}

func TestRequirementTraceMergeIgnoresUnreachablePaths(t *testing.T) {
	base := newRequirementTrace()
	reachable := base.forkBranch()
	unreachable := base.forkBranch()
	appendMergeTraceReference(reachable, "pending-live", "live", 10)
	appendMergeTraceReference(unreachable, "pending-dead", "dead", 1)

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: unreachable, Reachable: false},
		{Ordinal: 1, Trace: reachable, Reachable: true},
	})
	if len(merged.references) != 1 || merged.references[0].pendingID != "pending-live" || !merged.references[0].directExternal || merged.references[0].conditional || merged.references[0].pathConditional {
		t.Fatalf("unreachable path contributed requirements: %+v", merged.references)
	}
	projected := mustProjectRequirements(t, merged)
	if !projected.Coverage.Complete || len(projected.Items) != 1 || projected.Items[0].Identity != "live" || projected.Items[0].Necessity != "required" || len(projected.Gaps) != 0 {
		t.Fatalf("unreachable-path projection = %+v", projected)
	}
}

func TestRequirementTraceMergeKeepsOriginIndeterminacySeparateFromPathNecessity(t *testing.T) {
	base := newRequirementTrace()
	uncertain := base.forkBranch()
	reference := mergeTraceReference("pending-uncertain", "uncertain", 10)
	reference.Binding = "indeterminate"
	uncertain.recordReference(reference, false, true, uncertain.nextEvent())
	other := base.forkBranch()

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: uncertain, Reachable: true},
		{Ordinal: 1, Trace: other, Reachable: true},
	})
	if len(merged.references) != 1 || !merged.references[0].conditional || merged.references[0].directExternal {
		t.Fatalf("origin indeterminacy changed: %+v", merged.references)
	}
	projected := mustProjectRequirements(t, merged)
	if projected.Coverage.Complete || len(projected.Gaps) != 1 || projected.Gaps[0].Code != CodeRequirementIndeterminate || projected.Items[0].Necessity != "conditional" {
		t.Fatalf("origin-indeterminate projection = %+v", projected)
	}
}

func TestRequirementTraceMergeRetainsIncompleteReachableChild(t *testing.T) {
	base := newRequirementTrace()
	complete := base.forkBranch()
	incomplete := base.forkBranch()
	incomplete.recordDiagnostic(Diagnostic{Code: CodeUnsupportedSemantics, Severity: "warning", Category: "unsupported_semantics", Message: "child gap", StageID: "stage-child"}, true, nil, incomplete.nextEvent())

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 0, Trace: complete, Reachable: true},
		{Ordinal: 1, Trace: incomplete, Reachable: true},
	})
	projected := mustProjectRequirements(t, merged)
	if projected.Coverage.Complete || len(projected.Gaps) != 1 || projected.Gaps[0].Code != CodeUnsupportedSemantics {
		t.Fatalf("incomplete reachable child projection = %+v", projected)
	}
}

func TestRequirementTraceMergeOrdersOccurrencesByLocationBranchAndEvent(t *testing.T) {
	base := newRequirementTrace()
	left := base.forkBranch()
	right := base.forkBranch()
	appendMergeTraceReference(left, "pending-late", "shared", 30)
	appendMergeTraceReference(left, "pending-left-first", "shared", 10)
	appendMergeTraceReference(left, "pending-left-second", "shared", 10)
	appendMergeTraceReference(right, "pending-right", "shared", 10)

	merged := mergeRequirementTraces(base, []requirementTracePath{
		{Ordinal: 7, Trace: right, Reachable: true},
		{Ordinal: 3, Trace: left, Reachable: true},
	})
	want := []string{"pending-left-first", "pending-left-second", "pending-right", "pending-late"}
	got := make([]string, 0, len(merged.references))
	for _, entry := range merged.references {
		got = append(got, entry.pendingID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("merged occurrence order = %v want %v", got, want)
	}
}

func TestRequirementTraceMergeIsRepeatableAndDoesNotMutateInputs(t *testing.T) {
	base := newRequirementTrace()
	appendMergeTraceReference(base, "pending-prefix", "prefix", 1)
	left := base.forkBranch()
	right := base.forkBranch()
	appendMergeTraceReference(left, "pending-left", "shared", 10)
	appendMergeTraceReference(right, "pending-right", "shared", 20)
	baseBefore := marshalRequirementTrace(t, base)
	leftBefore := marshalRequirementTrace(t, left)
	rightBefore := marshalRequirementTrace(t, right)
	paths := []requirementTracePath{
		{Ordinal: 0, Trace: left, Reachable: true},
		{Ordinal: 1, Trace: right, Reachable: true},
	}

	first := mergeRequirementTraces(base, paths)
	second := mergeRequirementTraces(base, paths)
	if !reflect.DeepEqual(marshalRequirementTrace(t, first), marshalRequirementTrace(t, second)) {
		t.Fatalf("repeated trace merges differ: first=%+v second=%+v", first, second)
	}
	if !reflect.DeepEqual(marshalRequirementTrace(t, base), baseBefore) || !reflect.DeepEqual(marshalRequirementTrace(t, left), leftBefore) || !reflect.DeepEqual(marshalRequirementTrace(t, right), rightBefore) {
		t.Fatal("trace merge mutated an input")
	}
}

func TestRequirementTraceRebaseRetainsSuffixExactlyOnceWithoutMutation(t *testing.T) {
	oldBase := newRequirementTrace()
	appendMergeTraceReference(oldBase, "pending-prefix", "prefix", 1)
	branch := oldBase.forkBranch()
	appendMergeTraceReference(branch, "pending-child-first", "child", 5)
	appendMergeTraceReference(branch, "pending-child-second", "child", 6)
	newBase := oldBase.clone()
	appendMergeTraceReference(newBase, "pending-lazy", "lazy", 20)
	oldBefore := marshalRequirementTrace(t, oldBase)
	branchBefore := marshalRequirementTrace(t, branch)
	newBefore := marshalRequirementTrace(t, newBase)

	rebased := rebaseRequirementTrace(oldBase, newBase, branch)
	want := []string{"pending-prefix", "pending-lazy", "pending-child-first", "pending-child-second"}
	got := make([]string, 0, len(rebased.references))
	for _, entry := range rebased.references {
		got = append(got, entry.pendingID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rebased trace = %v want %v", got, want)
	}
	if !reflect.DeepEqual(marshalRequirementTrace(t, oldBase), oldBefore) || !reflect.DeepEqual(marshalRequirementTrace(t, branch), branchBefore) || !reflect.DeepEqual(marshalRequirementTrace(t, newBase), newBefore) {
		t.Fatal("trace rebase mutated an input")
	}
}

func TestRequirementTraceRebaseRejectsChangedPrefix(t *testing.T) {
	oldBase := newRequirementTrace()
	appendMergeTraceReference(oldBase, "pending-prefix", "prefix", 1)
	branch := oldBase.forkBranch()
	branch.references[0].reference.NormalizedName = "changed"

	defer func() {
		if recover() == nil {
			t.Fatal("trace rebase silently accepted a changed immutable prefix")
		}
	}()
	rebaseRequirementTrace(oldBase, oldBase, branch)
}

func TestRequirementTraceSourceOwnershipCloneAndRemap(t *testing.T) {
	trace := newRequirementTrace()
	trace.recordReference(Reference{ID: "pending-source", Kind: "dataset", NormalizedName: "$events", Role: "read", StageID: "stage-old"}, true, false, trace.nextEvent())
	trace.recordReference(Reference{ID: "pending-read", Kind: "field", NormalizedName: "id", Role: "read", StageID: "stage-old"}, true, false, trace.nextEvent())
	fact := inputFact{kind: "named_placeholder", identity: InputIdentity{Form: "parameter", Value: "$events"}, occurrence: InputOccurrence{ReferenceID: "pending-source", OriginalReferenceID: "pending-source", StageID: "stage-old", UseSiteLocations: []Location{{Start: Position{Offset: 5}}}, UseSiteReferenceIDs: []string{"pending-source"}}}
	trace.inputs = cloneInputFacts([]inputFact{fact})
	trace.reference("pending-read").owners = cloneSourceOwners([]sourceOwner{{input: fact, identity: atomicFieldIdentity("id")}})
	cloned := trace.clone()
	cloned.references[1].owners[0].identity.Segments[0] = "changed"
	cloned.references[1].owners[0].input.occurrence.UseSiteLocations[0].Start.Offset = 999
	cloned.references[1].owners[0].input.occurrence.UseSiteReferenceIDs[0] = "changed"
	owner := trace.references[1].owners[0]
	if owner.identity.Segments[0] != "id" || owner.input.occurrence.UseSiteLocations[0].Start.Offset != 5 || owner.input.occurrence.UseSiteReferenceIDs[0] != "pending-source" {
		t.Fatal("trace owner clone changed its immutable prefix")
	}
	trace.remapReferences(map[string]string{"pending-source": "ref-0", "pending-read": "ref-1"})
	trace.remapStages(map[string]string{"stage-old": "stage-0"})
	owner = trace.references[1].owners[0]
	if owner.input.occurrence.ReferenceID != "ref-0" || owner.input.occurrence.OriginalReferenceID != "ref-0" || owner.input.occurrence.UseSiteReferenceIDs[0] != "ref-0" || owner.input.occurrence.StageID != "stage-0" {
		t.Fatalf("owner context remap lost links: %+v", owner)
	}
}
