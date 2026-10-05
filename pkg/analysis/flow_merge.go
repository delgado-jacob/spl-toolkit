package analysis

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type flowMergePath struct {
	Ordinal     int
	Environment *environment
	Reachable   bool
}

// composeFlowEnvironments combines simultaneous inputs into one matched row.
// Unlike mergeFlowEnvironments, neither input is an alternative path, so a
// field supplied by either side keeps its own presence and conditionality.
// The right trace must already be rebased onto the complete left trace; its
// suffix is then retained without changing requirement necessity.
func composeFlowEnvironments(left, right *environment) (*environment, []string, bool) {
	if left == nil || right == nil || !requirementTraceHasPrefix(left.requirements.trace, right.requirements.trace) {
		return nil, nil, false
	}
	trace := (*requirementTrace)(nil)
	if right.requirements.trace != nil {
		trace = right.requirements.trace.clone()
	}
	composed := left.cloneWithRequirementTrace(trace)
	composed.inputs = mergeInputFacts(left.inputs, right.inputs)
	composed.requirements.inputs = mergeInputFacts(left.requirements.inputs, right.requirements.inputs)
	composed.open = left.open || right.open
	composed.uncertain = left.uncertain || right.uncertain
	composed.requirements.open = left.requirements.open || right.requirements.open
	composed.requirements.uncertain = left.requirements.uncertain || right.requirements.uncertain
	composed.rewriteBarrier()

	collisions := []string{}
	seenCollision := map[string]bool{}
	recordCollision := func(name string) {
		if name != "" && !seenCollision[name] {
			seenCollision[name] = true
			collisions = append(collisions, name)
		}
	}
	leftNames := map[string]bool{}
	for _, key := range left.orderedFieldKeys() {
		if field, ok := left.fields[key]; ok {
			leftNames[field.Name] = true
		}
	}

	for _, key := range right.orderedFieldKeys() {
		rightField, ok := right.fields[key]
		if !ok {
			continue
		}
		if leftNames[rightField.Name] {
			recordCollision(rightField.Name)
		}
		if leftField, sameIdentity := composed.fields[key]; sameIdentity {
			leftField.ownerCollision = leftField.ownerCollision || rightField.ownerCollision || sourceOwnerKey(leftField.owners) != sourceOwnerKey(rightField.owners)
			leftField.owners = mergeSourceOwners(leftField.owners, rightField.owners)
			leftField.OriginReferenceIDs = uniqueIDs(leftField.OriginReferenceIDs, rightField.OriginReferenceIDs)
			leftField.Conditional = leftField.Conditional && rightField.Conditional
			leftField.source = leftField.source && rightField.source
			if leftField.valueState != rightField.valueState {
				leftField.valueState = fieldValueUnknown
			}
			composed.fields[key] = leftField
			delete(composed.removed, key)
			continue
		}
		composed.registerIdentityField(rightField)
		delete(composed.removed, key)
	}

	for key, rightField := range right.requirements.fields {
		if leftField, sameIdentity := composed.requirements.fields[key]; sameIdentity {
			leftField.ownerCollision = leftField.ownerCollision || rightField.ownerCollision || sourceOwnerKey(leftField.owners) != sourceOwnerKey(rightField.owners)
			leftField.owners = mergeSourceOwners(leftField.owners, rightField.owners)
			leftField.origins = uniqueIDs(leftField.origins, rightField.origins)
			leftField.conditional = leftField.conditional && rightField.conditional
			leftField.source = leftField.source && rightField.source
			leftField.unavailable = leftField.unavailable && rightField.unavailable
			composed.requirements.fields[key] = leftField
			delete(composed.requirements.removed, key)
			continue
		}
		composed.requirements.registerIdentityField(rightField)
		delete(composed.requirements.removed, key)
	}

	composeRemovedIdentities(composed, left, right)
	sort.Strings(collisions)
	return composed, collisions, true
}

func requirementTraceHasPrefix(prefix, trace *requirementTrace) bool {
	if prefix == nil || trace == nil {
		return prefix == nil && trace == nil
	}
	if len(trace.references) < len(prefix.references) || len(trace.diagnostics) < len(prefix.diagnostics) {
		return false
	}
	return reflect.DeepEqual(trace.references[:len(prefix.references)], prefix.references) &&
		reflect.DeepEqual(trace.diagnostics[:len(prefix.diagnostics)], prefix.diagnostics)
}

func composeRemovedIdentities(composed, left, right *environment) {
	keys := map[fieldIdentityKey]bool{}
	for key := range left.removed {
		keys[key] = true
	}
	for key := range right.removed {
		keys[key] = true
	}
	orderedKeys := make([]string, 0, len(keys))
	for key := range keys {
		orderedKeys = append(orderedKeys, string(key))
	}
	sort.Strings(orderedKeys)
	for _, encoded := range orderedKeys {
		key := fieldIdentityKey(encoded)
		if _, present := composed.fields[key]; present {
			delete(composed.removed, key)
			continue
		}
		leftRemoved, rightRemoved := left.removed[key], right.removed[key]
		provedAbsent := leftRemoved && rightRemoved || leftRemoved && !right.open || rightRemoved && !left.open
		if provedAbsent {
			if !composed.removed[key] {
				composed.removedOrder = append(composed.removedOrder, key)
			}
			composed.removed[key] = true
			if identity, ok := right.identities[key]; ok {
				composed.identities[key] = identity.clone()
			}
		} else {
			delete(composed.removed, key)
			composed.uncertain = true
		}

		if _, present := composed.requirements.fields[key]; present {
			delete(composed.requirements.removed, key)
			continue
		}
		leftRequirementRemoved, rightRequirementRemoved := left.requirements.removed[key], right.requirements.removed[key]
		requirementAbsent := leftRequirementRemoved && rightRequirementRemoved || leftRequirementRemoved && !right.requirements.open || rightRequirementRemoved && !left.requirements.open
		if requirementAbsent {
			composed.requirements.removed[key] = true
		} else {
			delete(composed.requirements.removed, key)
			if leftRequirementRemoved || rightRequirementRemoved {
				composed.requirements.uncertain = true
			}
		}
	}
}

func mergeFlowEnvironments(parent *environment, paths []flowMergePath, includeParent bool) *environment {
	reachable := make([]flowMergePath, 0, len(paths)+1)
	maxOrdinal := -1
	for _, path := range paths {
		if !path.Reachable || path.Environment == nil {
			continue
		}
		reachable = append(reachable, path)
		if path.Ordinal > maxOrdinal {
			maxOrdinal = path.Ordinal
		}
	}
	if includeParent && parent != nil {
		reachable = append(reachable, flowMergePath{Ordinal: maxOrdinal + 1, Environment: parent, Reachable: true})
	}
	sort.SliceStable(reachable, func(i, j int) bool { return reachable[i].Ordinal < reachable[j].Ordinal })
	if len(reachable) == 0 {
		if parent == nil {
			return newEnvironment()
		}
		return parent.clone()
	}

	var mergedTrace *requirementTrace
	if parent != nil && parent.requirements.trace != nil {
		tracePaths := make([]requirementTracePath, 0, len(reachable))
		for _, path := range reachable {
			tracePaths = append(tracePaths, requirementTracePath{Ordinal: path.Ordinal, Trace: path.Environment.requirements.trace, Reachable: true})
		}
		mergedTrace = mergeRequirementTraces(parent.requirements.trace, tracePaths)
	}
	merged := newEnvironmentWithRequirementTrace(mergedTrace)
	for _, path := range reachable {
		merged.inputs = mergeInputFacts(merged.inputs, path.Environment.inputs)
		merged.requirements.inputs = mergeInputFacts(merged.requirements.inputs, path.Environment.requirements.inputs)
	}
	merged.open = false
	merged.requirements.open = false
	if parent != nil {
		merged.rewrite = parent.rewrite.clone()
		merged.rewriteBarrier()
	}

	for _, path := range reachable {
		merged.open = merged.open || path.Environment.open
		merged.uncertain = merged.uncertain || path.Environment.uncertain
		merged.requirements.open = merged.requirements.open || path.Environment.requirements.open
		merged.requirements.uncertain = merged.requirements.uncertain || path.Environment.requirements.uncertain
	}

	keys := orderedMergeKeys(reachable)
	for _, key := range keys {
		mergeTrackedIdentity(merged, reachable, key)
		mergeRequirementIdentity(&merged.requirements, reachable, key)
	}
	for name := range merged.ambiguous {
		merged.detectCollision(name)
	}
	for name := range merged.requirements.ambiguous {
		merged.requirements.detectCollision(name)
	}
	return merged
}

func orderedMergeKeys(paths []flowMergePath) []fieldIdentityKey {
	keys := []fieldIdentityKey{}
	seen := map[fieldIdentityKey]bool{}
	for _, path := range paths {
		pathKeys := make([]string, 0, len(path.Environment.fields)+len(path.Environment.removed)+len(path.Environment.requirements.fields)+len(path.Environment.requirements.removed))
		for key := range path.Environment.fields {
			pathKeys = append(pathKeys, string(key))
		}
		for key := range path.Environment.removed {
			pathKeys = append(pathKeys, string(key))
		}
		for key := range path.Environment.requirements.fields {
			pathKeys = append(pathKeys, string(key))
		}
		for key := range path.Environment.requirements.removed {
			pathKeys = append(pathKeys, string(key))
		}
		sort.Strings(pathKeys)
		for _, encoded := range pathKeys {
			key := fieldIdentityKey(encoded)
			if seen[key] {
				continue
			}
			seen[key] = true
			keys = append(keys, key)
		}
	}
	return keys
}

func mergeTrackedIdentity(merged *environment, paths []flowMergePath, key fieldIdentityKey) {
	present, removed := 0, 0
	var field trackedField
	origins := []string{}
	owners := []sourceOwner{}
	var firstOrigins []string
	originsDiffer := false
	allSource := true
	valueState := fieldValueUnknown
	valueStateSet := false
	valueStateAgrees := true
	for _, path := range paths {
		candidate, known := path.Environment.fields[key]
		if known {
			present++
			owners = mergeSourceOwners(owners, candidate.owners)
			if present == 1 {
				field = candidate
			}
			field.ownerCollision = field.ownerCollision || candidate.ownerCollision
			orderedOrigins := orderedReferenceIDs(candidate.OriginReferenceIDs)
			if firstOrigins == nil {
				firstOrigins = orderedOrigins
			} else if !sameStrings(firstOrigins, orderedOrigins) {
				originsDiffer = true
			}
			origins = uniqueIDs(origins, orderedOrigins)
			allSource = allSource && candidate.source
			if !valueStateSet {
				valueState = candidate.valueState
				valueStateSet = true
			} else if valueState != candidate.valueState {
				valueStateAgrees = false
			}
			continue
		}
		allSource = false
		valueStateAgrees = false
		if path.Environment.removed[key] {
			removed++
		}
	}
	if present == 0 {
		if removed == len(paths) {
			identity := mergeIdentity(paths, key)
			merged.identities[key] = identity
			merged.removed[key] = true
			merged.removedOrder = append(merged.removedOrder, key)
		} else if removed > 0 {
			merged.uncertain = true
		}
		return
	}
	field.owners = owners
	field.identity = mergeIdentity(paths, key)
	field.Name = field.identity.PublicName
	field.OriginReferenceIDs = origins
	field.Conditional = field.Conditional || present != len(paths)
	field.source = allSource
	if !valueStateAgrees {
		field.valueState = fieldValueUnknown
	} else {
		field.valueState = valueState
	}
	for _, path := range paths {
		candidate, known := path.Environment.fields[key]
		if known && candidate.Conditional {
			field.Conditional = true
		}
		if !known && path.Environment.open {
			merged.uncertain = true
		}
	}
	if originsDiffer || removed > 0 && removed != len(paths) {
		merged.uncertain = true
	}
	collision, _ := merged.registerIdentityField(field)
	if collision {
		merged.uncertain = true
	}
}

func mergeRequirementIdentity(merged *requirementEnvironment, paths []flowMergePath, key fieldIdentityKey) {
	present, removed := 0, 0
	var field requirementField
	origins := []string{}
	owners := []sourceOwner{}
	var firstOrigins []string
	originsDiffer := false
	allSource := true
	for _, path := range paths {
		candidate, known := path.Environment.requirements.fields[key]
		if known {
			present++
			owners = mergeSourceOwners(owners, candidate.owners)
			if present == 1 {
				field = candidate
			}
			field.ownerCollision = field.ownerCollision || candidate.ownerCollision
			orderedOrigins := orderedReferenceIDs(candidate.origins)
			if firstOrigins == nil {
				firstOrigins = orderedOrigins
			} else if !sameStrings(firstOrigins, orderedOrigins) {
				originsDiffer = true
			}
			origins = uniqueIDs(origins, orderedOrigins)
			allSource = allSource && candidate.source
			continue
		}
		allSource = false
		if path.Environment.requirements.removed[key] {
			removed++
		}
	}
	if present == 0 {
		if removed == len(paths) {
			merged.removed[key] = true
		} else if removed > 0 {
			merged.uncertain = true
		}
		return
	}
	field.owners = owners
	field.identity = mergeIdentity(paths, key)
	field.origins = origins
	field.conditional = field.conditional || present != len(paths)
	field.source = allSource
	for _, path := range paths {
		candidate, known := path.Environment.requirements.fields[key]
		if known && candidate.conditional {
			field.conditional = true
		}
		if !known && path.Environment.requirements.open {
			merged.uncertain = true
		}
	}
	if originsDiffer || removed > 0 && removed != len(paths) {
		merged.uncertain = true
	}
	merged.registerIdentityField(field)
}

func mergeIdentity(paths []flowMergePath, key fieldIdentityKey) fieldIdentity {
	for _, path := range paths {
		if field, ok := path.Environment.fields[key]; ok {
			return field.identity.clone()
		}
		if field, ok := path.Environment.requirements.fields[key]; ok {
			return field.identity.clone()
		}
		if identity, ok := path.Environment.identities[key]; ok {
			return identity.clone()
		}
	}
	return atomicFieldIdentity(string(key))
}

func orderedReferenceIDs(ids []string) []string {
	out := append([]string{}, ids...)
	sort.SliceStable(out, func(i, j int) bool {
		leftPrefix, leftOrdinal := referenceIDOrder(out[i])
		rightPrefix, rightOrdinal := referenceIDOrder(out[j])
		if leftPrefix != rightPrefix {
			return leftPrefix < rightPrefix
		}
		if leftOrdinal != rightOrdinal {
			return leftOrdinal < rightOrdinal
		}
		return out[i] < out[j]
	})
	return out
}

func referenceIDOrder(id string) (string, int) {
	index := strings.LastIndexByte(id, '-')
	if index < 0 {
		return id, -1
	}
	ordinal, err := strconv.Atoi(id[index+1:])
	if err != nil {
		return id, -1
	}
	return id[:index], ordinal
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
