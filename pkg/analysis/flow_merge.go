package analysis

import (
	"sort"
	"strconv"
	"strings"
)

type flowMergePath struct {
	Ordinal     int
	Environment *environment
	Reachable   bool
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
			if present == 1 {
				field = candidate
			}
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
	var firstOrigins []string
	originsDiffer := false
	allSource := true
	for _, path := range paths {
		candidate, known := path.Environment.requirements.fields[key]
		if known {
			present++
			if present == 1 {
				field = candidate
			}
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
