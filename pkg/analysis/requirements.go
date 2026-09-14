package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type RequirementQueryIdentity struct {
	SourceID    string `json:"source_id"`
	Language    string `json:"language"`
	Profile     string `json:"profile"`
	Version     string `json:"version"`
	QueryDigest string `json:"query_digest"`
}

type RequirementCoverage struct {
	Complete bool     `json:"complete"`
	Reasons  []string `json:"reasons"`
}

type RequirementOccurrence struct {
	ReferenceID  string   `json:"reference_id"`
	OriginalName string   `json:"original_name"`
	Binding      string   `json:"binding"`
	StageID      string   `json:"stage_id"`
	ScopeID      string   `json:"scope_id"`
	Location     Location `json:"location"`
}

type RequirementItem struct {
	ID          string                  `json:"id"`
	Kind        string                  `json:"kind"`
	Identity    string                  `json:"identity"`
	Role        string                  `json:"role"`
	Necessity   string                  `json:"necessity"`
	Origin      string                  `json:"origin"`
	Resolution  string                  `json:"resolution"`
	Occurrences []RequirementOccurrence `json:"occurrences"`
}

type RequirementGap struct {
	Code            string   `json:"code"`
	Message         string   `json:"message"`
	ReferenceIDs    []string `json:"reference_ids"`
	DiagnosticCodes []string `json:"diagnostic_codes"`
}

type RequirementSet struct {
	SchemaVersion      int                      `json:"schema_version"`
	Query              RequirementQueryIdentity `json:"query"`
	CapabilityRevision string                   `json:"capability_revision"`
	QueryStatus        Status                   `json:"query_status"`
	Coverage           RequirementCoverage      `json:"coverage"`
	Items              []RequirementItem        `json:"items"`
	Gaps               []RequirementGap         `json:"gaps"`
	Diagnostics        []Diagnostic             `json:"diagnostics"`
}

func queryDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func capabilityRevision(document QueryDocument) (string, error) {
	manifest, err := CapabilitiesFor(CapabilityOptions{
		Language: document.Language,
		Profile:  document.Profile,
		Version:  document.Version,
	})
	if err != nil {
		return "", err
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func cloneRequirementSet(in RequirementSet) RequirementSet {
	out := in
	out.Coverage.Reasons = append([]string{}, in.Coverage.Reasons...)
	out.Items = append([]RequirementItem{}, in.Items...)
	for i := range out.Items {
		out.Items[i].Occurrences = append([]RequirementOccurrence{}, in.Items[i].Occurrences...)
	}
	out.Gaps = append([]RequirementGap{}, in.Gaps...)
	for i := range out.Gaps {
		out.Gaps[i].ReferenceIDs = append([]string{}, in.Gaps[i].ReferenceIDs...)
		out.Gaps[i].DiagnosticCodes = append([]string{}, in.Gaps[i].DiagnosticCodes...)
	}
	out.Diagnostics = append([]Diagnostic{}, in.Diagnostics...)
	return out
}

// Requirements analyzes a query once and returns its detached direct external obligations.
func Requirements(document QueryDocument) (*RequirementSet, error) {
	result, err := Analyze(document)
	if err != nil {
		return nil, err
	}
	set := cloneRequirementSet(result.Requirements)
	return &set, nil
}

func projectRequirements(document QueryDocument, trace *requirementTrace) (RequirementSet, error) {
	revision, err := capabilityRevision(document)
	if err != nil {
		return RequirementSet{}, err
	}
	set := RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			SourceID:    document.SourceID,
			Language:    document.Language,
			Profile:     document.Profile,
			Version:     document.Version,
			QueryDigest: queryDigest(document.Text),
		},
		CapabilityRevision: revision,
		QueryStatus:        Valid,
		Coverage:           RequirementCoverage{Complete: true, Reasons: []string{}},
		Items:              []RequirementItem{},
		Gaps:               []RequirementGap{},
		Diagnostics:        []Diagnostic{},
	}

	type groupKey struct {
		kind, identity, role, resolution string
	}
	type gapCandidate struct {
		gap          RequirementGap
		eventOrdinal int
	}
	groups := map[groupKey]int{}
	gapCandidates := []gapCandidate{}
	references := append([]requirementTraceReference{}, trace.references...)
	sort.SliceStable(references, func(i, j int) bool {
		a, b := references[i].reference, references[j].reference
		if ai, aok := canonicalReferenceOrdinal(a.ID); aok {
			if bi, bok := canonicalReferenceOrdinal(b.ID); bok && ai != bi {
				return ai < bi
			}
		}
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		if a.Location.End.Offset != b.Location.End.Offset {
			return a.Location.End.Offset < b.Location.End.Offset
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Role != b.Role {
			return a.Role < b.Role
		}
		if a.NormalizedName != b.NormalizedName {
			return a.NormalizedName < b.NormalizedName
		}
		return a.Resolution < b.Resolution
	})
	for _, entry := range references {
		if !entry.directExternal && !entry.conditional {
			continue
		}
		reference := entry.reference
		if reference.NormalizedName != "" {
			key := groupKey{reference.Kind, reference.NormalizedName, reference.Role, reference.Resolution}
			index, found := groups[key]
			if !found {
				necessity := "required"
				if entry.conditional {
					necessity = "conditional"
				}
				set.Items = append(set.Items, RequirementItem{
					Kind:        reference.Kind,
					Identity:    reference.NormalizedName,
					Role:        reference.Role,
					Necessity:   necessity,
					Origin:      "direct",
					Resolution:  reference.Resolution,
					Occurrences: []RequirementOccurrence{},
				})
				index = len(set.Items) - 1
				groups[key] = index
			} else if entry.directExternal {
				set.Items[index].Necessity = "required"
			}
			set.Items[index].Occurrences = append(set.Items[index].Occurrences, RequirementOccurrence{
				ReferenceID:  reference.ID,
				OriginalName: reference.OriginalName,
				Binding:      reference.Binding,
				StageID:      reference.StageID,
				ScopeID:      reference.ScopeID,
				Location:     reference.Location,
			})
		}
		if entry.conditional {
			code := CodeRequirementIndeterminate
			message := "requirement origin is indeterminate"
			if reference.Resolution == "wildcard" || reference.Resolution == "dynamic" {
				code = CodeRequirementDynamic
				message = "dynamic requirement identity is unresolved"
				if traceHasOwnedDiagnostic(trace, reference.ID, CodeUnresolvedWildcard) {
					continue
				}
			}
			gapCandidates = append(gapCandidates, gapCandidate{RequirementGap{
				Code: code, Message: message, ReferenceIDs: []string{reference.ID}, DiagnosticCodes: []string{},
			}, entry.eventOrdinal})
		}
	}
	for i := range set.Items {
		set.Items[i].ID = fmt.Sprintf("req-%d", i+1)
	}
	for _, entry := range trace.diagnostics {
		set.Diagnostics = append(set.Diagnostics, entry.diagnostic)
		if entry.diagnostic.Severity == "error" {
			set.QueryStatus = Invalid
		}
		if !entry.incomplete {
			continue
		}
		gapCandidates = append(gapCandidates, gapCandidate{RequirementGap{
			Code:            entry.diagnostic.Code,
			Message:         entry.diagnostic.Message,
			ReferenceIDs:    append([]string{}, entry.pendingReferenceIDs...),
			DiagnosticCodes: []string{entry.diagnostic.Code},
		}, entry.eventOrdinal})
	}
	sort.SliceStable(set.Diagnostics, func(i, j int) bool {
		a, b := set.Diagnostics[i], set.Diagnostics[j]
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	sort.SliceStable(gapCandidates, func(i, j int) bool { return gapCandidates[i].eventOrdinal < gapCandidates[j].eventOrdinal })
	for _, candidate := range gapCandidates {
		appendRequirementGap(&set, candidate.gap)
	}
	if (!trace.syntaxComplete || !trace.semanticComplete) && len(set.Gaps) == 0 {
		appendRequirementGap(&set, RequirementGap{
			Code:            CodeRequirementCoverageIncomplete,
			Message:         "requirement coverage is incomplete",
			ReferenceIDs:    []string{},
			DiagnosticCodes: []string{},
		})
	}
	if set.QueryStatus != Invalid && (!trace.syntaxComplete || !trace.semanticComplete) {
		set.QueryStatus = Incomplete
	}
	return set, nil
}

func traceHasOwnedDiagnostic(trace *requirementTrace, referenceID, code string) bool {
	for _, diagnostic := range trace.diagnostics {
		if !diagnostic.incomplete || diagnostic.diagnostic.Code != code {
			continue
		}
		for _, owner := range diagnostic.pendingReferenceIDs {
			if owner == referenceID {
				return true
			}
		}
	}
	return false
}

func appendRequirementGap(set *RequirementSet, gap RequirementGap) {
	for _, existing := range set.Gaps {
		if existing.Code == gap.Code && stringSlicesEqual(existing.ReferenceIDs, gap.ReferenceIDs) && stringSlicesEqual(existing.DiagnosticCodes, gap.DiagnosticCodes) {
			return
		}
	}
	set.Gaps = append(set.Gaps, gap)
	set.Coverage.Complete = false
	for _, reason := range set.Coverage.Reasons {
		if reason == gap.Code {
			return
		}
	}
	set.Coverage.Reasons = append(set.Coverage.Reasons, gap.Code)
}

func stringSlicesEqual(left, right []string) bool {
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

func canonicalReferenceOrdinal(id string) (int, bool) {
	value, err := strconv.Atoi(strings.TrimPrefix(id, "ref-"))
	return value, err == nil && strings.HasPrefix(id, "ref-")
}
