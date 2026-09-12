package document

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// New returns a caller-owned deep copy of result with its accepted canonical
// revision context. It identifies the supplied snapshot, not a cryptographic
// proof that an arbitrary caller-provided result came from the engine.
func New(result *analysis.Result, context RevisionContext) (*Snapshot, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is required")
	}
	if !utf8.ValidString(result.Document.Text) || !utf8.ValidString(result.Document.SourceID) {
		return nil, fmt.Errorf("analysis document must contain valid UTF-8")
	}
	if !utf8.ValidString(context.ToolVersion) || strings.TrimSpace(context.ToolVersion) == "" ||
		!utf8.ValidString(context.ContractVersion) || strings.TrimSpace(context.ContractVersion) == "" ||
		!utf8.ValidString(context.TargetDigest) {
		return nil, fmt.Errorf("valid tool and contract revision context is required")
	}

	view := &Snapshot{
		SchemaVersion: result.SchemaVersion,
		Revision:      context,
		SourceHash:    sourceHash(result.Document.Text),
		Document:      result.Document,
		Status:        result.Status,
		Coverage:      cloneCoverage(result.Coverage),
		Stages:        append([]analysis.Stage{}, result.Stages...),
		Scopes:        append([]analysis.Scope{}, result.Scopes...),
		References:    cloneReferences(result.References),
		Lineage:       cloneLineage(result.Lineage),
		Dependencies:  cloneDependencies(result.Dependencies),
		Diagnostics:   append([]analysis.Diagnostic{}, result.Diagnostics...),
	}
	return view, nil
}

func sourceHash(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func cloneCoverage(in analysis.Coverage) analysis.Coverage {
	in.Reasons = append([]string{}, in.Reasons...)
	return in
}

func cloneReferences(in []analysis.Reference) []analysis.Reference {
	out := make([]analysis.Reference, len(in))
	for i, reference := range in {
		out[i] = reference
		out[i].OriginReferenceIDs = append([]string{}, reference.OriginReferenceIDs...)
	}
	return out
}

func cloneLineage(in []analysis.Lineage) []analysis.Lineage {
	out := make([]analysis.Lineage, len(in))
	for i, lineage := range in {
		out[i] = lineage
		out[i].Before = cloneFieldState(lineage.Before)
		out[i].After = cloneFieldState(lineage.After)
		out[i].Transitions = make([]analysis.Transition, len(lineage.Transitions))
		for j, transition := range lineage.Transitions {
			out[i].Transitions[j] = transition
			out[i].Transitions[j].InputReferenceIDs = append([]string{}, transition.InputReferenceIDs...)
		}
		if lineage.ExecutionOrder != nil {
			order := *lineage.ExecutionOrder
			out[i].ExecutionOrder = &order
		}
	}
	return out
}

func cloneFieldState(in analysis.FieldState) analysis.FieldState {
	out := in
	out.Fields = make([]analysis.FieldBinding, len(in.Fields))
	for i, field := range in.Fields {
		out.Fields[i] = field
		out.Fields[i].OriginReferenceIDs = append([]string{}, field.OriginReferenceIDs...)
	}
	out.Removed = append([]string{}, in.Removed...)
	return out
}

func cloneDependencies(in analysis.Dependencies) analysis.Dependencies {
	return analysis.Dependencies{
		Indexes:     append([]string{}, in.Indexes...),
		Sources:     append([]string{}, in.Sources...),
		SourceTypes: append([]string{}, in.SourceTypes...),
		Datasets:    append([]string{}, in.Datasets...),
		Lookups:     append([]string{}, in.Lookups...),
		DataModels:  append([]string{}, in.DataModels...),
		Macros:      append([]string{}, in.Macros...),
	}
}
