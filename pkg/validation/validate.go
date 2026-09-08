package validation

import (
	"fmt"
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

const (
	CodeUnknownField       = "SPL_UNKNOWN_FIELD"
	CodeIndeterminateField = "SPL_INDETERMINATE_FIELD"
)

func Validate(document analysis.QueryDocument, catalog FieldCatalog) (*Report, error) {
	normalized, err := normalizeCatalog(catalog)
	if err != nil {
		return nil, err
	}
	document, err = normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	return validate(document, normalized)
}

// validate consumes kernel evidence; it never replays field transfers or repairs
// incomplete expansion evidence using catalog-side matching.
func validate(document analysis.QueryDocument, catalog FieldCatalog) (*Report, error) {
	fields := append(append([]string{}, catalog.Fields...), catalog.OptionalFields...)
	source, err := analysis.AnalyzeWithSourceFields(document, fields)
	if err != nil {
		return nil, err
	}
	result := source.Result
	report := &Report{SchemaVersion: 1, Target: Target{Kind: "field_list", FieldCatalog: catalog}, Analysis: result, Coverage: Coverage{SyntaxComplete: result.Coverage.SyntaxComplete, SemanticComplete: result.Coverage.SemanticComplete, SchemaComplete: result.Coverage.SyntaxComplete && result.Coverage.SemanticComplete, Reasons: []string{}}, Outcomes: []ReferenceOutcome{}, Diagnostics: append([]analysis.Diagnostic{}, result.Diagnostics...)}
	ordinary, optional := map[string]bool{}, map[string]bool{}
	for _, name := range catalog.Fields {
		ordinary[name] = true
	}
	for _, name := range catalog.OptionalFields {
		optional[name] = true
	}
	expansions := map[string]analysis.FieldExpansion{}
	for _, expansion := range source.Expansions {
		expansions[expansion.ReferenceID] = expansion
	}
	match := func(name, binding string) Match {
		outcome := "matching"
		if binding == "source" && optional[name] {
			outcome = "optional_equivalent"
		}
		return Match{Name: name, Binding: binding, Outcome: outcome}
	}
	for _, ref := range result.References {
		if ref.Role == "remove" || ref.Role == "null_test" || ref.Binding == "not_applicable" || ref.Kind != "field" {
			continue
		}
		item := ReferenceOutcome{ReferenceID: ref.ID, Matches: []Match{}}
		if ref.Binding == "unavailable" {
			item.Outcome = "unavailable"
			report.Outcomes = append(report.Outcomes, item)
			continue
		}
		reason := ""
		if ref.Resolution == "wildcard" {
			expansion, found := expansions[ref.ID]
			switch {
			case !found || !expansion.Complete:
				item.Outcome = "indeterminate"
				reason = "the kernel could not establish a complete set of fields for this selector"
			case len(expansion.Matches) == 0:
				item.Outcome = "missing"
			default:
				item.Outcome = "optional_equivalent"
				for _, member := range expansion.Matches {
					m := match(member.Name, member.Binding)
					item.Matches = append(item.Matches, m)
					if m.Outcome == "matching" {
						item.Outcome = "matching"
					}
				}
			}
		} else {
			switch ref.Binding {
			case "source":
				if ordinary[ref.NormalizedName] || optional[ref.NormalizedName] {
					m := match(ref.NormalizedName, "source")
					item.Outcome = m.Outcome
					item.Matches = append(item.Matches, m)
				} else {
					item.Outcome = "missing"
				}
			case "derived":
				item.Outcome = "matching"
				item.Matches = append(item.Matches, match(ref.NormalizedName, "derived"))
			default:
				item.Outcome = "indeterminate"
				reason = "the kernel could not prove whether this field is available from the source or earlier query stages"
			}
		}
		switch item.Outcome {
		case "missing":
			report.Diagnostics = append(report.Diagnostics, analysis.Diagnostic{Code: CodeUnknownField, Severity: "error", Category: "unknown_field", Message: fmt.Sprintf("Field %q has no matching declaration in the local catalog", ref.NormalizedName), Location: ref.Location, StageID: ref.StageID, ScopeID: ref.ScopeID})
		case "indeterminate":
			report.Coverage.SchemaComplete = false
			report.Diagnostics = append(report.Diagnostics, analysis.Diagnostic{Code: CodeIndeterminateField, Severity: "warning", Category: "schema_ambiguity", Message: fmt.Sprintf("Field %q is indeterminate: %s", ref.NormalizedName, reason), Location: ref.Location, StageID: ref.StageID, ScopeID: ref.ScopeID})
		}
		report.Outcomes = append(report.Outcomes, item)
	}
	finalizeReport(report)
	return report, nil
}

func finalizeReport(report *Report) {
	report.Diagnostics, report.Status = finalizeValidation(report.Diagnostics, &report.Coverage)
}

func finalizeValidation(diagnostics []analysis.Diagnostic, coverage *Coverage) ([]analysis.Diagnostic, analysis.Status) {
	sort.SliceStable(diagnostics, func(i, j int) bool {
		a, b := diagnostics[i], diagnostics[j]
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
	seen := map[analysis.Diagnostic]bool{}
	unique := make([]analysis.Diagnostic, 0, len(diagnostics))
	for _, d := range diagnostics {
		if !seen[d] {
			unique = append(unique, d)
			seen[d] = true
		}
	}
	diagnostics = unique
	reasons := map[string]bool{}
	for _, d := range diagnostics {
		if !reasons[d.Code] {
			coverage.Reasons = append(coverage.Reasons, d.Code)
			reasons[d.Code] = true
		}
	}
	status := analysis.Valid
	if !coverage.SyntaxComplete || !coverage.SemanticComplete || !coverage.SchemaComplete {
		status = analysis.Incomplete
	}
	for _, d := range diagnostics {
		if d.Severity == "error" {
			status = analysis.Invalid
		}
	}
	return diagnostics, status
}
