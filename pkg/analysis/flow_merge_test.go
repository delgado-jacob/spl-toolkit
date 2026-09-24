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

func TestMergeFlowEnvironmentsCollapsesAtomicStructuralCollision(t *testing.T) {
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
	if len(state.Fields) != 1 || state.Fields[0].Name != "actor.name" || !state.Fields[0].Conditional || !state.Uncertain || !reflect.DeepEqual(state.Fields[0].OriginReferenceIDs, []string{"ref-2", "ref-8"}) {
		t.Fatalf("collision merge = %+v", state)
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
