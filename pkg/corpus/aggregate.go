package corpus

import (
	"sort"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func newReport(selection Selection, mode string) *Report {
	selection.IgnoredNames = append([]string{}, selection.IgnoredNames...)
	selection.SkippedSymlinks = append([]string{}, selection.SkippedSymlinks...)
	selection.TraversalFailures = append([]AcquisitionError{}, selection.TraversalFailures...)
	r := &Report{SchemaVersion: 1, Status: analysis.Valid, ExecutionComplete: true, Mode: mode, Selection: selection, CoverageReasons: []string{}, CommandCoverage: []CommandCoverage{}, ObservedReferenceForms: []ObservedReferenceForm{}, Dependencies: []DependencySummary{}, Entries: []ReportEntry{}}
	if mode == "analysis" {
		r.Coverage.Schema.NotRequested = true
	}
	return r
}

func reduceStatus(current, next analysis.Status) analysis.Status {
	if current == analysis.Invalid || next == analysis.Invalid {
		return analysis.Invalid
	}
	if current == analysis.Incomplete || next == analysis.Incomplete {
		return analysis.Incomplete
	}
	return analysis.Valid
}

func addCoverage(count *CoverageCount, complete bool) {
	count.Denominator++
	if complete {
		count.Complete++
	} else {
		count.Incomplete++
	}
}

func addStatus(counts *StatusCounts, status analysis.Status) {
	switch status {
	case analysis.Valid:
		counts.Valid++
	case analysis.Invalid:
		counts.Invalid++
	case analysis.Incomplete:
		counts.Incomplete++
	}
}

func addCoverageReasons(report *Report, reasons ...[]string) {
	for _, group := range reasons {
		for _, reason := range group {
			found := false
			for _, current := range report.CoverageReasons {
				if current == reason {
					found = true
					break
				}
			}
			if !found {
				report.CoverageReasons = append(report.CoverageReasons, reason)
			}
		}
	}
}

func appendObservedCoverage(report *Report, result *analysis.Result) error {
	manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: result.Document.Language, Profile: result.Document.Profile, Version: result.Document.Version})
	if err != nil {
		return err
	}
	for _, stage := range result.Stages {
		index := -1
		for i, item := range report.CommandCoverage {
			if item.Language == manifest.Language && item.Profile == manifest.Profile && item.Version == manifest.Version && item.Command == stage.Command {
				index = i
				break
			}
		}
		if index < 0 {
			item := CommandCoverage{Language: manifest.Language, Profile: manifest.Profile, Version: manifest.Version, Command: stage.Command, Limitations: []string{}}
			for _, capability := range manifest.Commands {
				if capability.Name == stage.Command {
					item.Declared = true
					item.SyntaxSupported = capability.SyntaxSupported
					item.SemanticSupported = capability.SemanticSupported
					item.Limitations = append([]string{}, capability.Limitations...)
					break
				}
			}
			report.CommandCoverage = append(report.CommandCoverage, item)
			index = len(report.CommandCoverage) - 1
		}
		report.CommandCoverage[index].Count++
	}
	for _, ref := range result.References {
		index := -1
		for i, item := range report.ObservedReferenceForms {
			if item.Kind == ref.Kind && item.Role == ref.Role && item.Resolution == ref.Resolution {
				index = i
				break
			}
		}
		if index < 0 {
			report.ObservedReferenceForms = append(report.ObservedReferenceForms, ObservedReferenceForm{Kind: ref.Kind, Role: ref.Role, Resolution: ref.Resolution})
			index = len(report.ObservedReferenceForms) - 1
		}
		report.ObservedReferenceForms[index].Count++
	}
	return nil
}

func appendDependencies(report *Report, documentID string, result *analysis.Result) {
	known := map[string]map[string]bool{
		"index":      stringSet(result.Dependencies.Indexes),
		"source":     stringSet(result.Dependencies.Sources),
		"sourcetype": stringSet(result.Dependencies.SourceTypes),
		"dataset":    stringSet(result.Dependencies.Datasets),
		"lookup":     stringSet(result.Dependencies.Lookups),
		"data_model": stringSet(result.Dependencies.DataModels),
		"macro":      stringSet(result.Dependencies.Macros),
	}
	for _, ref := range result.References {
		if !known[ref.Kind][ref.NormalizedName] || ref.Resolution != "exact" {
			continue
		}
		index := -1
		for i := range report.Dependencies {
			if report.Dependencies[i].Kind == ref.Kind && report.Dependencies[i].Name == ref.NormalizedName {
				index = i
				break
			}
		}
		if index < 0 {
			report.Dependencies = append(report.Dependencies, DependencySummary{Kind: ref.Kind, Name: ref.NormalizedName, DocumentIDs: []string{}, ReferenceIDs: []string{}})
			index = len(report.Dependencies) - 1
		}
		d := &report.Dependencies[index]
		if len(d.DocumentIDs) == 0 || d.DocumentIDs[len(d.DocumentIDs)-1] != documentID {
			d.DocumentIDs = append(d.DocumentIDs, documentID)
		}
		d.ReferenceIDs = append(d.ReferenceIDs, ref.ID)
	}
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func sortDependencies(report *Report) {
	sort.Strings(report.CoverageReasons)
	sort.Slice(report.CommandCoverage, func(i, j int) bool {
		a, b := report.CommandCoverage[i], report.CommandCoverage[j]
		if a.Language != b.Language {
			return a.Language < b.Language
		}
		if a.Profile != b.Profile {
			return a.Profile < b.Profile
		}
		if a.Version != b.Version {
			return a.Version < b.Version
		}
		return a.Command < b.Command
	})
	sort.Slice(report.ObservedReferenceForms, func(i, j int) bool {
		a, b := report.ObservedReferenceForms[i], report.ObservedReferenceForms[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Role != b.Role {
			return a.Role < b.Role
		}
		return a.Resolution < b.Resolution
	})
	sort.Slice(report.Dependencies, func(i, j int) bool {
		a, b := report.Dependencies[i], report.Dependencies[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return a.Name < b.Name
	})
}
