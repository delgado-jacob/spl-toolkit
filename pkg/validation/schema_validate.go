package validation

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// ValidateSchema checks canonical field obligations against an offline target.
// Schema failures are InputErrors; query syntax failures remain invalid reports.
func ValidateSchema(document analysis.QueryDocument, target SchemaTarget) (*SchemaReport, error) {
	document, err := normalizeDocument(document)
	if err != nil {
		return nil, err
	}
	prepared, err := prepareSchemaTarget(target)
	if err != nil {
		return nil, err
	}
	return validatePreparedSchema(document, prepared)
}

// validatePreparedSchema consumes canonical bindings and final reference IDs.
// Target metadata never repairs or expands the kernel's membership evidence.
func validatePreparedSchema(document analysis.QueryDocument, prepared preparedSchemaTarget) (*SchemaReport, error) {
	universe := prepared.universe()
	source, err := analysis.AnalyzeWithSourceUniverse(document, universe)
	if err != nil {
		return nil, err
	}
	result := source.Result
	report := &SchemaReport{SchemaVersion: 1, Target: ownedSchemaInfo(prepared.info()), Analysis: result,
		Coverage: Coverage{SyntaxComplete: result.Coverage.SyntaxComplete, SemanticComplete: result.Coverage.SemanticComplete, SchemaComplete: result.Coverage.SyntaxComplete && result.Coverage.SemanticComplete, Reasons: []string{}},
		Outcomes: []SchemaReferenceOutcome{}, Diagnostics: append([]analysis.Diagnostic{}, result.Diagnostics...)}
	expansions := map[string]analysis.FieldExpansion{}
	for _, expansion := range source.Expansions {
		expansions[expansion.ReferenceID] = expansion
	}
	for _, ref := range result.References {
		if ref.Role == "remove" || ref.Role == "null_test" || ref.Binding == "not_applicable" || ref.Kind != "field" {
			continue
		}
		item := SchemaReferenceOutcome{ReferenceID: ref.ID, Matches: []SchemaMatch{}, Evidence: []SchemaEvidence{}, SupportingClasses: []SchemaClass{}, MissingClasses: []SchemaClass{}, IndeterminateClasses: []SchemaClass{}}
		missing := false
		addProjection := func(name, binding string) string {
			p := newProjection(analysis.SourceFieldAdmitted, "matching", SchemaEvidence{DeclarationBasis: "derived"})
			if binding == "source" {
				p = prepared.project(name)
			}
			item.Evidence = append(item.Evidence, p.Evidence...)
			item.SupportingClasses = append(item.SupportingClasses, p.SupportingClasses...)
			item.MissingClasses = append(item.MissingClasses, p.MissingClasses...)
			item.IndeterminateClasses = append(item.IndeterminateClasses, p.IndeterminateClasses...)
			if p.Admission == analysis.SourceFieldAdmitted {
				item.Matches = append(item.Matches, SchemaMatch{Name: name, Binding: binding, Outcome: p.Outcome, Evidence: canonicalSchemaEvidence(p.Evidence)})
			}
			missing = missing || p.Outcome == "missing"
			return p.Outcome
		}
		switch {
		case ref.Binding == "unavailable":
			item.Outcome = "unavailable"
		case ref.Resolution == "wildcard":
			expansion, found := expansions[ref.ID]
			item.MatchesComplete = found && expansion.Complete
			item.Outcome = "matching"
			for _, member := range expansion.Matches {
				item.Outcome = summarizeSchemaOutcome(item.Outcome, addProjection(member.Name, member.Binding))
			}
			if !item.MatchesComplete {
				item.Outcome = "indeterminate"
				if !universe.Complete {
					for _, limitation := range report.Target.Limitations {
						switch limitation {
						case "partial_name_universe", "enumeration_budget", "unrepresentable_source_name":
							item.Evidence = append(item.Evidence, SchemaEvidence{DeclarationBasis: "unknown", Requirement: "unknown", Reason: limitation})
						}
					}
				}
			} else if len(expansion.Matches) == 0 {
				item.Outcome = "missing"
				missing = true
			}
		case ref.Binding == "source":
			item.Outcome = addProjection(ref.NormalizedName, "source")
			item.MatchesComplete = item.Outcome != "conditional" && (len(item.Matches) > 0 || item.Outcome == "missing")
		case ref.Binding == "derived":
			item.Outcome = addProjection(ref.NormalizedName, "derived")
			item.MatchesComplete = true
		default:
			item.Outcome = "indeterminate"
		}
		if missing {
			report.Diagnostics = append(report.Diagnostics, analysis.Diagnostic{Code: CodeUnknownField, Severity: "error", Category: "unknown_field", Message: fmt.Sprintf("Field %q has no matching declaration in the local schema target", ref.NormalizedName), Location: ref.Location, StageID: ref.StageID, ScopeID: ref.ScopeID})
		}
		if item.Outcome == "indeterminate" || item.Outcome == "conditional" {
			report.Coverage.SchemaComplete = false
			reason := "the kernel could not establish complete field membership and source binding"
			if ref.Binding == "source" && ref.Resolution != "wildcard" {
				reason = "the schema target could not establish conclusive membership or requiredness"
			}
			report.Diagnostics = append(report.Diagnostics, analysis.Diagnostic{Code: CodeIndeterminateField, Severity: "warning", Category: "schema_ambiguity", Message: fmt.Sprintf("Field %q is %s: %s", ref.NormalizedName, item.Outcome, reason), Location: ref.Location, StageID: ref.StageID, ScopeID: ref.ScopeID})
		}
		item.Evidence = canonicalSchemaEvidence(item.Evidence)
		item.SupportingClasses = canonicalSchemaClasses(item.SupportingClasses)
		item.MissingClasses = canonicalSchemaClasses(item.MissingClasses)
		item.IndeterminateClasses = canonicalSchemaClasses(item.IndeterminateClasses)
		sort.Slice(item.Matches, func(i, j int) bool { return item.Matches[i].Name < item.Matches[j].Name })
		report.Outcomes = append(report.Outcomes, item)
	}
	report.Diagnostics, report.Status = finalizeValidation(report.Diagnostics, &report.Coverage)
	return report, nil
}

// Derived matches are neutral. A missing constituent remains an error even
// when an uncertain constituent takes precedence in the descriptive summary.
func summarizeSchemaOutcome(a, b string) string {
	rank := func(s string) int {
		switch s {
		case "indeterminate":
			return 7
		case "conditional":
			return 6
		case "missing":
			return 5
		case "permitted_unspecified":
			return 4
		case "optional":
			return 3
		case "required":
			return 2
		default:
			return 1
		}
	}
	if rank(b) > rank(a) {
		return b
	}
	return a
}

func canonicalSchemaClasses(classes []SchemaClass) []SchemaClass {
	result := append([]SchemaClass{}, classes...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Key != result[j].Key {
			return result[i].Key < result[j].Key
		}
		return result[i].UID < result[j].UID
	})
	out := result[:0]
	for _, c := range result {
		if len(out) == 0 || out[len(out)-1] != c {
			out = append(out, c)
		}
	}
	return out
}

// Compare UID values, never pointer identity; every published pointer is owned.
func canonicalSchemaEvidence(evidence []SchemaEvidence) []SchemaEvidence {
	key := func(e SchemaEvidence) [10]string {
		uid := ""
		if e.ClassUID != nil {
			uid = strconv.FormatInt(*e.ClassUID, 10)
		}
		return [10]string{e.ResourceURI, e.Pointer, e.Operator, e.Branch, e.ClassKey, uid, e.Reason, e.Keyword, e.DeclarationBasis, e.Requirement}
	}
	result := append([]SchemaEvidence{}, evidence...)
	sort.SliceStable(result, func(i, j int) bool {
		a, b := key(result[i]), key(result[j])
		for n := range a {
			if a[n] != b[n] {
				return a[n] < b[n]
			}
		}
		return false
	})
	out := make([]SchemaEvidence, 0, len(result))
	var previous [10]string
	for _, e := range result {
		k := key(e)
		if len(out) > 0 && k == previous {
			continue
		}
		previous = k
		if e.ClassUID != nil {
			uid := *e.ClassUID
			e.ClassUID = &uid
		}
		out = append(out, e)
	}
	return out
}

func ownedSchemaInfo(info SchemaTargetInfo) SchemaTargetInfo {
	info.ResourceURIs = append([]string{}, info.ResourceURIs...)
	info.Limitations = append([]string{}, info.Limitations...)
	info.Members = canonicalSchemaClasses(info.Members)
	info.Extensions = append([]SchemaExtension{}, info.Extensions...)
	if info.Selection != nil {
		s := *info.Selection
		s.Profiles = append([]string{}, s.Profiles...)
		s.Extensions = append([]string{}, s.Extensions...)
		if s.ClassUID != nil {
			uid := *s.ClassUID
			s.ClassUID = &uid
		}
		if s.CategoryUID != nil {
			uid := *s.CategoryUID
			s.CategoryUID = &uid
		}
		info.Selection = &s
	}
	return info
}
