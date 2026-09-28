package analysis

import (
	"reflect"
	"testing"
)

func closedMergeEnvironment() *environment {
	env := newEnvironment()
	env.open = false
	env.requirements.open = false
	return env
}

func installMergeField(env *environment, identity fieldIdentity, origins []string) {
	env.installIdentity(identity, origins, false, false)
	env.requirements.installIdentity(identity, origins, false, false)
}

func TestMergeFlowEnvironmentsEqualOutputsRemainUnconditional(t *testing.T) {
	parent := closedMergeEnvironment()
	left := parent.clone()
	right := parent.clone()
	identity := atomicFieldIdentity("shared")
	installMergeField(left, identity, []string{"ref-1"})
	installMergeField(right, identity, []string{"ref-1"})

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: left, Reachable: true},
		{Ordinal: 1, Environment: right, Reachable: true},
	}, false)

	field, ok := merged.field(identity)
	if !ok || field.Conditional || merged.uncertain || !reflect.DeepEqual(field.OriginReferenceIDs, []string{"ref-1"}) {
		t.Fatalf("equal merge = field %+v known=%t uncertain=%t", field, ok, merged.uncertain)
	}
	requirementField, ok := merged.requirements.field(identity)
	if !ok || requirementField.conditional || merged.requirements.uncertain || !reflect.DeepEqual(requirementField.origins, []string{"ref-1"}) {
		t.Fatalf("equal requirement merge = field %+v known=%t uncertain=%t", requirementField, ok, merged.requirements.uncertain)
	}
}

func TestMergeFlowEnvironmentsSubsetPresenceIsConditional(t *testing.T) {
	parent := closedMergeEnvironment()
	left := parent.clone()
	right := parent.clone()
	identity := atomicFieldIdentity("left_only")
	installMergeField(left, identity, []string{"ref-2"})

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 1, Environment: right, Reachable: true},
		{Ordinal: 0, Environment: left, Reachable: true},
	}, false)

	field, ok := merged.field(identity)
	if !ok || !field.Conditional || merged.uncertain {
		t.Fatalf("subset merge = field %+v known=%t uncertain=%t", field, ok, merged.uncertain)
	}
	requirementField, ok := merged.requirements.field(identity)
	if !ok || !requirementField.conditional || merged.requirements.uncertain {
		t.Fatalf("subset requirement merge = field %+v known=%t uncertain=%t", requirementField, ok, merged.requirements.uncertain)
	}
}

func TestMergeFlowEnvironmentsRetainsRequirementOnlyIdentity(t *testing.T) {
	parent := closedMergeEnvironment()
	left := parent.clone()
	right := parent.clone()
	identity := atomicFieldIdentity("requirement_only")
	left.requirements.installIdentity(identity, []string{"ref-5"}, false, true)
	right.requirements.installIdentity(identity, []string{"ref-5"}, false, true)

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: left, Reachable: true},
		{Ordinal: 1, Environment: right, Reachable: true},
	}, false)
	field, ok := merged.requirements.field(identity)
	if !ok || field.conditional || !field.source || !reflect.DeepEqual(field.origins, []string{"ref-5"}) {
		t.Fatalf("requirement-only merge = field %+v known=%t", field, ok)
	}
}

func TestMergeFlowEnvironmentsConflictingOriginsAreStableAndUncertain(t *testing.T) {
	parent := closedMergeEnvironment()
	first := parent.clone()
	second := parent.clone()
	identity := atomicFieldIdentity("shared")
	installMergeField(first, identity, []string{"ref-9", "ref-3"})
	installMergeField(second, identity, []string{"ref-7"})

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 9, Environment: second, Reachable: true},
		{Ordinal: 2, Environment: first, Reachable: true},
	}, false)

	field, ok := merged.field(identity)
	if !ok || !merged.uncertain || !reflect.DeepEqual(field.OriginReferenceIDs, []string{"ref-3", "ref-9", "ref-7"}) {
		t.Fatalf("origin merge = field %+v known=%t uncertain=%t", field, ok, merged.uncertain)
	}
	requirementField, ok := merged.requirements.field(identity)
	if !ok || !merged.requirements.uncertain || !reflect.DeepEqual(requirementField.origins, []string{"ref-3", "ref-9", "ref-7"}) {
		t.Fatalf("requirement origin merge = field %+v known=%t uncertain=%t", requirementField, ok, merged.requirements.uncertain)
	}
}

func TestMergeFlowEnvironmentsRemovalRequiresEveryPath(t *testing.T) {
	parent := closedMergeEnvironment()
	identity := atomicFieldIdentity("retained")
	installMergeField(parent, identity, []string{"ref-1"})

	removedLeft := parent.clone()
	removedLeft.removeIdentity(identity)
	removedLeft.requirements.removeIdentity(identity)
	removedRight := parent.clone()
	removedRight.removeIdentity(identity)
	removedRight.requirements.removeIdentity(identity)

	allRemoved := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: removedLeft, Reachable: true},
		{Ordinal: 1, Environment: removedRight, Reachable: true},
	}, false)
	key, _ := identity.privateKey()
	if _, ok := allRemoved.field(identity); ok || !allRemoved.removed[key] || !allRemoved.requirements.removed[key] {
		t.Fatalf("all-removed merge = %+v requirements=%+v", allRemoved.snapshot(), allRemoved.requirements)
	}

	retained := parent.clone()
	someRemoved := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: removedLeft, Reachable: true},
		{Ordinal: 1, Environment: retained, Reachable: true},
	}, false)
	field, ok := someRemoved.field(identity)
	if !ok || !field.Conditional || !someRemoved.uncertain || someRemoved.removed[key] {
		t.Fatalf("some-removed merge = field %+v known=%t state=%+v", field, ok, someRemoved.snapshot())
	}
	requirementField, ok := someRemoved.requirements.field(identity)
	if !ok || !requirementField.conditional || !someRemoved.requirements.uncertain || someRemoved.requirements.removed[key] {
		t.Fatalf("some-removed requirement merge = field %+v known=%t env=%+v", requirementField, ok, someRemoved.requirements)
	}
}

func TestMergeFlowEnvironmentsPropagatesOpenAndUncertainInputs(t *testing.T) {
	parent := closedMergeEnvironment()
	closed := parent.clone()
	open := parent.clone()
	open.open = true
	open.requirements.open = true
	uncertain := parent.clone()
	uncertain.uncertain = true
	uncertain.requirements.uncertain = true

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: closed, Reachable: true},
		{Ordinal: 1, Environment: open, Reachable: true},
		{Ordinal: 2, Environment: uncertain, Reachable: true},
	}, false)
	if !merged.open || !merged.uncertain || !merged.requirements.open || !merged.requirements.uncertain {
		t.Fatalf("open/uncertain merge = public open=%t uncertain=%t requirement open=%t uncertain=%t", merged.open, merged.uncertain, merged.requirements.open, merged.requirements.uncertain)
	}
}

func TestMergeFlowEnvironmentsOmittedElseIncludesParentPath(t *testing.T) {
	parent := closedMergeEnvironment()
	branch := parent.clone()
	identity := atomicFieldIdentity("branch_only")
	installMergeField(branch, identity, []string{"ref-4"})

	merged := mergeFlowEnvironments(parent, []flowMergePath{{Ordinal: 0, Environment: branch, Reachable: true}}, true)
	field, ok := merged.field(identity)
	if !ok || !field.Conditional {
		t.Fatalf("omitted-else merge = field %+v known=%t", field, ok)
	}
}

func TestMergeFlowEnvironmentsPreservesAtomicStructuralIdentities(t *testing.T) {
	parent := closedMergeEnvironment()
	atomic := parent.clone()
	structural := parent.clone()
	installMergeField(atomic, atomicFieldIdentity("actor.name"), []string{"ref-2"})
	installMergeField(structural, pathFieldIdentity("", []string{"actor", "name"}), []string{"ref-8"})

	merged := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: atomic, Reachable: true},
		{Ordinal: 1, Environment: structural, Reachable: true},
	}, false)
	state := merged.snapshot()
	if len(state.Fields) != 2 || state.Uncertain || state.Fields[0].FieldIdentity.Kind != "atomic" || state.Fields[1].FieldIdentity.Kind != "path" || !state.Fields[0].Conditional || !state.Fields[1].Conditional || !reflect.DeepEqual(state.Fields[0].OriginReferenceIDs, []string{"ref-2"}) || !reflect.DeepEqual(state.Fields[1].OriginReferenceIDs, []string{"ref-8"}) {
		t.Fatalf("identity merge = %+v", state)
	}
	if !merged.requirements.ambiguous["actor.name"] || !merged.requirements.uncertain {
		t.Fatalf("requirement collision merge = %+v", merged.requirements)
	}
}

func TestMergeFlowEnvironmentsRetainsElementSelectionOnlyWhenEveryPathAgrees(t *testing.T) {
	parent := closedMergeEnvironment()
	identity := atomicFieldIdentity("values")
	installMergeField(parent, identity, []string{"ref-1"})
	selectedLeft := parent.clone()
	selectedRight := parent.clone()
	selectedLeft.selectElement(identity)
	selectedRight.selectElement(identity)

	agreed := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: selectedLeft, Reachable: true},
		{Ordinal: 1, Environment: selectedRight, Reachable: true},
	}, false)
	field, _ := agreed.field(identity)
	if field.valueState != fieldValueElementSelected {
		t.Fatalf("agreed value state = %v", field.valueState)
	}
	if got, want := agreed.snapshot(), parent.snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("agreed private state changed public schema: got %+v want %+v", got, want)
	}

	disagreed := mergeFlowEnvironments(parent, []flowMergePath{
		{Ordinal: 0, Environment: selectedLeft, Reachable: true},
		{Ordinal: 1, Environment: parent.clone(), Reachable: true},
	}, false)
	field, _ = disagreed.field(identity)
	if field.valueState != fieldValueUnknown {
		t.Fatalf("disagreed value state = %v", field.valueState)
	}
	if got, want := disagreed.snapshot(), parent.snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("disagreed private state changed public schema: got %+v want %+v", got, want)
	}
}

func TestComposeFlowEnvironmentsDisjointFieldsAndRequiredTrace(t *testing.T) {
	baseTrace := newRequirementTrace()
	left := newEnvironmentWithRequirementTrace(baseTrace)
	left.open = false
	left.requirements.open = false
	installMergeField(left, atomicFieldIdentity("left"), []string{"ref-1"})

	rightTrace := baseTrace.forkBranch()
	appendMergeTraceReference(rightTrace, "pending-right", "right_source", 10)
	right := newEnvironmentWithRequirementTrace(rightTrace)
	right.open = false
	right.requirements.open = false
	installMergeField(right, atomicFieldIdentity("right"), []string{"ref-2"})

	composed, collisions, ok := composeFlowEnvironments(left, right)
	if !ok || len(collisions) != 0 {
		t.Fatalf("composition = ok %t collisions %v", ok, collisions)
	}
	for _, identity := range []fieldIdentity{atomicFieldIdentity("left"), atomicFieldIdentity("right")} {
		field, known := composed.field(identity)
		if !known || field.Conditional {
			t.Errorf("disjoint field %s = %+v known=%t", identity.PublicName, field, known)
		}
	}
	if composed.requirements.trace == rightTrace || len(composed.requirements.trace.references) != 1 || !composed.requirements.trace.references[0].directExternal || composed.requirements.trace.references[0].conditional {
		t.Fatalf("composed trace = %+v", composed.requirements.trace)
	}
}

func TestComposeFlowEnvironmentsCombinesSameIdentityOrigins(t *testing.T) {
	left := closedMergeEnvironment()
	right := closedMergeEnvironment()
	identity := atomicFieldIdentity("shared")
	installMergeField(left, identity, []string{"ref-1"})
	installMergeField(right, identity, []string{"ref-2"})

	composed, collisions, ok := composeFlowEnvironments(left, right)
	field, known := composed.field(identity)
	if !ok || !known || field.Conditional || composed.uncertain || !reflect.DeepEqual(field.OriginReferenceIDs, []string{"ref-1", "ref-2"}) || !reflect.DeepEqual(collisions, []string{"shared"}) {
		t.Fatalf("same identity composition = field %+v known=%t uncertain=%t collisions=%v ok=%t", field, known, composed.uncertain, collisions, ok)
	}
	requirement, known := composed.requirements.field(identity)
	if !known || requirement.conditional || !reflect.DeepEqual(requirement.origins, []string{"ref-1", "ref-2"}) {
		t.Fatalf("same identity requirement = %+v known=%t", requirement, known)
	}
}

func TestComposeFlowEnvironmentsReportsJoinOutputCollision(t *testing.T) {
	left := closedMergeEnvironment()
	right := closedMergeEnvironment()
	installMergeField(left, atomicFieldIdentity("actor.name"), []string{"ref-1"})
	installMergeField(right, pathFieldIdentity("", []string{"actor", "name"}), []string{"ref-2"})

	composed, collisions, ok := composeFlowEnvironments(left, right)
	state := composed.snapshot()
	if !ok || !reflect.DeepEqual(collisions, []string{"actor.name"}) || len(state.Fields) != 2 || state.Uncertain || state.Fields[0].FieldIdentity.Kind != "atomic" || state.Fields[1].FieldIdentity.Kind != "path" || !reflect.DeepEqual(state.Fields[0].OriginReferenceIDs, []string{"ref-1"}) || !reflect.DeepEqual(state.Fields[1].OriginReferenceIDs, []string{"ref-2"}) {
		t.Fatalf("join composition = state %+v collisions=%v ok=%t", state, collisions, ok)
	}
}

func TestComposeFlowEnvironmentsPropagatesOpenUncertainAndRemovals(t *testing.T) {
	left := closedMergeEnvironment()
	right := closedMergeEnvironment()
	removed := atomicFieldIdentity("removed")
	left.removeIdentity(removed)
	left.requirements.removeIdentity(removed)
	right.removeIdentity(removed)
	right.requirements.removeIdentity(removed)
	right.open = true
	right.requirements.open = true
	right.uncertain = true
	right.requirements.uncertain = true

	composed, _, ok := composeFlowEnvironments(left, right)
	key, _ := removed.privateKey()
	if !ok || !composed.open || !composed.uncertain || !composed.requirements.open || !composed.requirements.uncertain || !composed.removed[key] || !composed.requirements.removed[key] {
		t.Fatalf("open/uncertain/removal composition = %+v requirements=%+v ok=%t", composed.snapshot(), composed.requirements, ok)
	}

	provider := closedMergeEnvironment()
	installMergeField(provider, removed, []string{"ref-3"})
	present, _, ok := composeFlowEnvironments(left, provider)
	if field, known := present.field(removed); !ok || !known || field.Conditional || present.removed[key] {
		t.Fatalf("present side did not override removal: field %+v known=%t state=%+v ok=%t", field, known, present.snapshot(), ok)
	}
}

func TestComposeFlowEnvironmentsPreservesStructuredIdentityAndPrivateValueState(t *testing.T) {
	left := closedMergeEnvironment()
	right := closedMergeEnvironment()
	structured := pathFieldIdentity("", []string{"actor", "name"})
	selected := atomicFieldIdentity("values")
	installMergeField(left, structured, []string{"ref-1"})
	installMergeField(left, selected, []string{"ref-2"})
	installMergeField(right, selected, []string{"ref-3"})
	left.selectElement(selected)
	right.selectElement(selected)

	composed, _, ok := composeFlowEnvironments(left, right)
	structuredField, structuredKnown := composed.field(structured)
	selectedField, selectedKnown := composed.field(selected)
	if !ok || !structuredKnown || structuredField.identity.Kind != fieldIdentityPath || !selectedKnown || selectedField.valueState != fieldValueElementSelected {
		t.Fatalf("structured/private composition = structured %+v/%t selected %+v/%t ok=%t", structuredField, structuredKnown, selectedField, selectedKnown, ok)
	}

	disagreed := closedMergeEnvironment()
	installMergeField(disagreed, selected, []string{"ref-4"})
	composed, _, ok = composeFlowEnvironments(left, disagreed)
	selectedField, _ = composed.field(selected)
	if !ok || selectedField.valueState != fieldValueUnknown {
		t.Fatalf("disagreed private state = %+v ok=%t", selectedField, ok)
	}
}

func TestComposeFlowEnvironmentsRejectsTraceWithoutCurrentPrefix(t *testing.T) {
	leftTrace := newRequirementTrace()
	appendMergeTraceReference(leftTrace, "pending-left", "left", 1)
	rightTrace := newRequirementTrace()
	appendMergeTraceReference(rightTrace, "pending-right", "right", 2)
	left := newEnvironmentWithRequirementTrace(leftTrace)
	right := newEnvironmentWithRequirementTrace(rightTrace)

	if composed, collisions, ok := composeFlowEnvironments(left, right); ok || composed != nil || len(collisions) != 0 {
		t.Fatalf("mismatched trace composition = env %+v collisions=%v ok=%t", composed, collisions, ok)
	}
}
