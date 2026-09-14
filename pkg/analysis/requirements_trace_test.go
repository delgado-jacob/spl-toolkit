package analysis

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

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
	env.requirements.fields["host"] = requirementField{source: true, conditional: true, origins: []string{"pending-0"}}
	env.requirements.removed["old"] = true
	clone := env.clone()

	clonedField := clone.requirements.fields["host"]
	clonedField.origins[0] = "changed"
	clonedField.conditional = false
	clone.requirements.fields["host"] = clonedField
	clone.requirements.removed["new"] = true
	clone.requirements.open = false
	clone.requirements.uncertain = true

	field := env.requirements.fields["host"]
	if !field.source || !field.conditional || !reflect.DeepEqual(field.origins, []string{"pending-0"}) || env.requirements.removed["new"] || !env.requirements.open || env.requirements.uncertain {
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
		name     string
		document QueryDocument
		field    string
	}{
		{"SPL OUTPUTNEW", QueryDocument{Text: "search user=* | lookup users user OUTPUTNEW role | table role"}, "role"},
		{"SPL2 partial dataset field", QueryDocument{Text: "FROM [{a:1},{b:2}] | table a", Language: "spl2"}, "a"},
		{"SPL2 deferred bin effect", QueryDocument{Text: "FROM main | eval a=host | bin a | table a", Language: "spl2"}, "a"},
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
			if public.Binding != "indeterminate" {
				t.Fatalf("public final read = %+v, want indeterminate", public)
			}
			if traced.reference.Binding != "indeterminate" || traced.directExternal || !traced.conditional {
				t.Fatalf("query-only final read = %+v, want conditional indeterminate evidence", traced)
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
	}{
		{"finite field list", QueryDocument{Text: "fields host* | table hostname"}, testSourceRefinement([]string{"hostname"}, true, nil, true)},
		{"partial source universe", QueryDocument{Text: "fields host* | table hostname"}, testSourceRefinement([]string{"hostname"}, false, nil, false)},
		{"wildcard selectors", QueryDocument{Text: "eval label=host | table *"}, testSourceRefinement([]string{"host"}, true, nil, false)},
		{"unsupported wildcard selector", QueryDocument{Text: "sort host*"}, testSourceRefinement([]string{"host"}, true, nil, false)},
		{"dotted SPL2 names", QueryDocument{Text: `SELECT 'actor.name' FROM main`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 downstream derived", QueryDocument{Text: `FROM main | eval local='actor.name' | table local`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 rename output", QueryDocument{Text: `FROM main | rename 'actor.name' AS actor | table actor`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 aggregate output", QueryDocument{Text: `FROM main | stats count('actor.name') AS total | table total`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 lookup output", QueryDocument{Text: `FROM main | lookup users 'actor.name' OUTPUT role | table role`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 downstream SPL source", QueryDocument{Text: `FROM main | eval local='actor.name' | where other=1`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
		{"dotted SPL2 downstream SQL source", QueryDocument{Text: `FROM main WHERE 'actor.name'=1 SELECT other`, Language: "spl2"}, testSourceRefinement([]string{"actor.name"}, true, admitted, false)},
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
				t.Fatalf("trace changed under refinement:\n plain: %s\nrefined: %s", plainTraceJSON, refinedTraceJSON)
			}
			plainJSON, _ := json.Marshal(plainResult)
			refinedJSON, _ := json.Marshal(refinedResult)
			if string(plainJSON) == string(refinedJSON) {
				t.Fatal("refinement parity row did not change the public result")
			}
		})
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
		PendingID      string    `json:"pending_id"`
		Reference      Reference `json:"reference"`
		DirectExternal bool      `json:"direct_external"`
		Conditional    bool      `json:"conditional"`
		EventOrdinal   int       `json:"event_ordinal"`
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
		wire.References = append(wire.References, referenceJSON{entry.pendingID, entry.reference, entry.directExternal, entry.conditional, entry.eventOrdinal})
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
