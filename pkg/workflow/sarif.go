package workflow

import (
	"fmt"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/sarif"
)

// ExportSARIF projects retained workflow evidence. Candidate findings stay
// unlocated: candidate byte ranges never describe the original artifact.
func ExportSARIF(source *Report) (*sarif.Log, error) {
	subjects, err := exportSubjects(source)
	if err != nil {
		return nil, err
	}
	adapter := &corpus.Report{SchemaVersion: 1, Status: source.Status, ExecutionComplete: true, Mode: "analysis", Selection: source.Selection, Entries: []corpus.ReportEntry{}}
	notifications := []sarif.Notification{}
	pointers := map[string]string{}
	failures := len(source.Selection.TraversalFailures)
	if failures > 0 {
		adapter.ExecutionComplete = false
	}
	for i, e := range source.Entries {
		pointers[e.ID] = fmt.Sprintf("/entries/%d/analysis", i)
		ce := corpus.ReportEntry{ID: e.ID, Origin: e.Origin, SourceHash: e.SourceHash}
		if e.Analysis != nil {
			revision, err := analysis.CapabilityRevisionFor(analysis.CapabilityOptions{Language: e.Analysis.Document.Language, Profile: e.Analysis.Document.Profile, Version: e.Analysis.Document.Version})
			if err != nil || revision != e.Analysis.Requirements.CapabilityRevision {
				return nil, fmt.Errorf("workflow: %s: recorded capability revision differs from current canonical capabilities", pointers[e.ID])
			}
			ce.AnalysisRevision, err = corpus.AnalysisRevision(e.Analysis.Document)
			if err != nil {
				return nil, err
			}
			ce.Evaluation = &corpus.Evaluation{Kind: "analysis", Analysis: e.Analysis}
		} else if e.Failure != nil {
			ce.Failure = &corpus.AcquisitionError{Code: e.Failure.Code, Phase: e.Failure.Phase, Message: e.Failure.Message}
			adapter.ExecutionComplete = false
		} else {
			return nil, fmt.Errorf("workflow: entry %q lacks analysis and failure", e.ID)
		}
		adapter.Entries = append(adapter.Entries, ce)
		if e.Failure != nil {
			failures++
			if e.Analysis != nil {
				notifications = append(notifications, sarif.Notification{Level: "error", Message: sarif.Message{Text: e.Failure.Message}, Properties: map[string]any{"document_id": e.ID, "code": e.Failure.Code, "phase": e.Failure.Phase, "detail": e.Failure.Detail, "evidence_pointer": fmt.Sprintf("/entries/%d/failure", i)}})
			}
		}
	}
	if source.ExecutionComplete != (failures == 0) {
		return nil, fmt.Errorf("workflow: execution completeness contradicts operational failures")
	}
	findings := []sarif.AdditionalFinding{}
	add := func(s exportSubject, code, level, message, pointer string, location *analysis.Location, properties map[string]any) {
		if properties == nil {
			properties = map[string]any{}
		}
		properties["domain"], properties["evidence_pointer"] = s.domain, pointer
		if s.domain == "candidate" {
			properties["variant_selection"] = s.selection
			if location != nil {
				properties["candidate_range"] = location
			}
			location = nil
		}
		findings = append(findings, sarif.AdditionalFinding{DocumentID: s.id, RuleID: code, Level: level, Message: message, Location: location, Properties: properties})
	}
	summary := func(s exportSubject, outcome, success, rule, pointer string, before int) {
		if outcome != success && len(findings) == before {
			add(s, rule, outcomeLevel(outcome), "Static evidence outcome: "+outcome, pointer, nil, map[string]any{"outcome": outcome})
		}
	}
	reason := func(s exportSubject, r compatibility.Reason, pointer, outcome string, query analysis.RequirementQueryIdentity) {
		loc := r.Location
		if loc == nil && len(r.Locations) == 1 {
			loc = &r.Locations[0]
		}
		// A reason's location is original only when its canonical occurrence agrees.
		if s.domain == "original" && !originalReason(s, query, r, loc) {
			loc = nil
		}
		add(s, r.Code, outcomeLevel(outcome), r.Message, pointer, loc, map[string]any{"reason": r, "outcome": outcome})
	}
	reasons := func(s exportSubject, rs []compatibility.Reason, pointer, outcome string, query analysis.RequirementQueryIdentity) {
		for i, r := range rs {
			reason(s, r, fmt.Sprintf("%s/%d", pointer, i), outcome, query)
		}
	}
	// Fallbacks are scoped to each nested entry; a sibling finding cannot
	// conceal a reasonless failed obligation or evidence coverage gap.
	requirementFinding := func(s exportSubject, o compatibility.RequirementOutcome, pointer string) {
		before := len(findings)
		reasons(s, o.Reasons, pointer+"/reasons", o.Outcome, o.Query)
		if len(findings) == before && o.Outcome != "satisfied" && o.Outcome != "not_applicable" && o.Applicability != "inapplicable" && o.Applicability != "not_applicable" {
			level := outcomeLevel(o.Outcome)
			if o.Outcome == "missing" && o.Applicability == "applicable" {
				level = "error"
			}
			add(s, "SPL_WORKFLOW_REQUIREMENT_OUTCOME", level, "Requirement evidence outcome: "+o.Outcome, pointer, nil, map[string]any{"outcome": o.Outcome, "applicability": o.Applicability, "requirement_id": o.RequirementID, "input_id": o.InputID})
		}
	}
	inputFinding := func(s exportSubject, o compatibility.InputOutcome, pointer string, query analysis.RequirementQueryIdentity) {
		before := len(findings)
		reasons(s, o.Reasons, pointer+"/reasons", o.Outcome, query)
		if len(findings) == before && o.Outcome != "satisfied" && o.Outcome != "not_applicable" {
			add(s, "SPL_WORKFLOW_INPUT_OUTCOME", outcomeLevel(o.Outcome), "Input evidence outcome: "+o.Outcome, pointer, nil, map[string]any{"outcome": o.Outcome, "input_id": o.InputID})
		}
	}
	coverageFinding := func(s exportSubject, cov compatibility.Coverage, pointer string, query analysis.RequirementQueryIdentity) {
		before := len(findings)
		reasons(s, cov.Reasons, pointer+"/reasons", "incomplete", query)
		if len(findings) == before && cov.State != "complete" && cov.State != "not_applicable" {
			add(s, "SPL_WORKFLOW_COVERAGE_INCOMPLETE", "warning", "Evidence coverage state: "+cov.State, pointer, nil, map[string]any{"coverage_state": cov.State, "dimension": cov.Dimension, "input_id": cov.InputID, "object_id": cov.ObjectID, "schema_id": cov.SchemaID})
		}
	}
	for _, s := range subjects {
		if s.domain == "candidate" && s.analysis != nil {
			before := len(findings)
			for i, d := range s.analysis.Diagnostics {
				add(s, d.Code, sarifLevel(d.Severity), d.Message, fmt.Sprintf("%s/diagnostics/%d", s.analysisPointer, i), &d.Location, map[string]any{"category": d.Category})
			}
			summary(s, string(s.analysis.Status), "valid", "SPL_WORKFLOW_CANDIDATE_INCOMPLETE", s.analysisPointer, before)
		}
		if s.closure != nil {
			before := len(findings)
			for i, d := range s.closure.Diagnostics {
				loc := originalInterval(s, s.closure.Query, d.Source)
				add(s, d.Diagnostic.Code, sarifLevel(d.Diagnostic.Severity), d.Diagnostic.Message, fmt.Sprintf("%s/diagnostics/%d", s.closurePointer, i), loc, map[string]any{"source_interval": d.Source, "origins": d.Origins, "candidate_range": d.Diagnostic.Location})
			}
			for i, g := range s.closure.Gaps {
				add(s, g.Code, "warning", "Dependency closure evidence gap: "+g.Code, fmt.Sprintf("%s/gaps/%d", s.closurePointer, i), originalInterval(s, s.closure.Query, g.Source), map[string]any{"source_interval": g.Source, "gap": g})
			}
			if !s.closure.Coverage.Complete {
				summary(s, "incomplete", "complete", "SPL_WORKFLOW_CLOSURE_INCOMPLETE", s.closurePointer, before)
			}
		}
	}
	for i, e := range source.Entries {
		s := subjectsOriginal(subjects, e.ID)
		base := fmt.Sprintf("/entries/%d", i)
		if c := e.Compatibility; c != nil {
			before := len(findings)
			reasons(s, c.Reasons, base+"/compatibility/reasons", c.Outcome, c.Requirements.Query)
			for j, o := range c.RequirementOutcomes {
				requirementFinding(s, o, fmt.Sprintf("%s/compatibility/requirement_outcomes/%d", base, j))
			}
			for j, o := range c.Inputs {
				inputFinding(s, o, fmt.Sprintf("%s/compatibility/inputs/%d", base, j), c.Requirements.Query)
			}
			for j, cov := range c.Coverage {
				coverageFinding(s, cov, fmt.Sprintf("%s/compatibility/coverage/%d", base, j), c.Requirements.Query)
			}
			for j, d := range c.Diagnostics {
				add(s, d.Code, sarifLevel(d.Severity), d.Message, fmt.Sprintf("%s/compatibility/diagnostics/%d", base, j), nil, map[string]any{"diagnostic": d})
			}
			summary(s, c.Outcome, "satisfied", "SPL_WORKFLOW_COMPATIBILITY_INCOMPLETE", base+"/compatibility", before)
		}
		if e.Resolution == nil {
			continue
		}
		for j, v := range e.Resolution.Variants {
			vp := fmt.Sprintf("%s/resolution/variants/%d", base, j)
			s := exportSubject{id: e.ID, domain: "candidate", selection: v.Selection, pointer: vp, analysis: v.CandidateAnalysis}
			before := len(findings)
			for _, f := range findings {
				if f.DocumentID == e.ID && f.Properties["domain"] == "candidate" {
					pointer, _ := f.Properties["evidence_pointer"].(string)
					if pointer == vp || strings.HasPrefix(pointer, vp+"/") {
						before--
					}
				}
			}
			for k, d := range v.Diagnostics {
				add(s, d.Code, outcomeLevel(v.Outcome), d.Message, fmt.Sprintf("%s/diagnostics/%d", vp, k), d.Location, map[string]any{"placeholder": d.Placeholder})
			}
			if c := v.Compatibility; c != nil {
				cbefore := len(findings)
				for k, r := range c.Reasons {
					reason(s, r.Evidence, fmt.Sprintf("%s/compatibility/reasons/%d/evidence", vp, k), c.Outcome, c.Requirements.Query)
				}
				for k, o := range c.RequirementOutcomes {
					requirementFinding(s, o.Evidence, fmt.Sprintf("%s/compatibility/requirement_outcomes/%d/evidence", vp, k))
				}
				for k, o := range c.Inputs {
					inputFinding(s, o.Evidence, fmt.Sprintf("%s/compatibility/inputs/%d/evidence", vp, k), c.Requirements.Query)
				}
				for k, cov := range c.Coverage {
					coverageFinding(s, cov.Evidence, fmt.Sprintf("%s/compatibility/coverage/%d/evidence", vp, k), c.Requirements.Query)
				}
				for k, d := range c.Diagnostics {
					add(s, d.Code, sarifLevel(d.Severity), d.Message, fmt.Sprintf("%s/compatibility/diagnostics/%d", vp, k), nil, map[string]any{"diagnostic": d})
				}
				summary(s, c.Outcome, "satisfied", "SPL_WORKFLOW_COMPATIBILITY_INCOMPLETE", vp+"/compatibility", cbefore)
			}
			summary(s, v.Outcome, "verified", "SPL_WORKFLOW_RESOLUTION_INCOMPLETE", vp, before)
		}
	}
	log, err := sarif.ExportAdditional(adapter, findings)
	if err != nil {
		return nil, err
	}
	run := &log.Runs[0]
	for i := range run.Results {
		r := &run.Results[i]
		if _, ok := r.Properties["domain"]; !ok {
			r.Properties["domain"] = "original"
			id, _ := r.Properties["document_id"].(string)
			r.Properties["evidence_pointer"] = pointers[id]
		}
	}
	run.Properties = map[string]any{"content_status": source.Status, "execution_complete": source.ExecutionComplete, "ci_exit_code": source.CIExitCode, "selection": source.Selection, "counts": source.Counts, "provenance": source.Provenance}
	// The corpus owner emits acquisition notifications; retain the workflow's
	// structured failure evidence on those notifications as well.
	for i := range run.Invocations[0].ToolExecutionNotifications {
		n := &run.Invocations[0].ToolExecutionNotifications[i]
		id, _ := n.Properties["document_id"].(string)
		for j, e := range source.Entries {
			if e.ID == id && e.Failure != nil {
				n.Properties["detail"] = e.Failure.Detail
				n.Properties["evidence_pointer"] = fmt.Sprintf("/entries/%d/failure", j)
			}
		}
	}
	run.Invocations[0].ExecutionSuccessful = source.ExecutionComplete
	run.Invocations[0].ToolExecutionNotifications = append(run.Invocations[0].ToolExecutionNotifications, notifications...)
	return detachedExport(log)
}

func subjectsOriginal(subjects []exportSubject, id string) exportSubject {
	for _, s := range subjects {
		if s.id == id && s.domain == "original" {
			return s
		}
	}
	return exportSubject{}
}
func outcomeLevel(outcome string) string {
	switch outcome {
	case "failed", "unsatisfied", "invalid":
		return "error"
	default:
		return "warning"
	}
}
func sarifLevel(severity string) string {
	if severity == "info" {
		return "note"
	}
	return severity
}
func originalInterval(s exportSubject, query analysis.RequirementQueryIdentity, interval closure.SourceInterval) *analysis.Location {
	if s.domain != "original" || s.analysis == nil || query != s.analysis.Requirements.Query || interval.Kind != "query" || interval.ObjectID != "" || interval.SourceID != s.analysis.Document.SourceID {
		return nil
	}
	// Coordinates are computed by the SARIF owner from these checked byte offsets.
	return &analysis.Location{Start: analysis.Position{Offset: interval.Start, Line: 1, Column: 1}, End: analysis.Position{Offset: interval.End, Line: 1, Column: 1}}
}

func originalReason(s exportSubject, query analysis.RequirementQueryIdentity, reason compatibility.Reason, location *analysis.Location) bool {
	if s.analysis == nil || location == nil || query != s.analysis.Requirements.Query {
		return false
	}
	for _, item := range s.analysis.Requirements.Items {
		if item.ID == reason.RequirementID {
			for _, occurrence := range item.Occurrences {
				if occurrence.Location == *location {
					return true
				}
			}
		}
	}
	for _, input := range s.analysis.Inputs {
		if input.ID == reason.InputID {
			for _, occurrence := range input.Occurrences {
				if occurrence.Location == *location {
					return true
				}
			}
		}
	}
	for _, reference := range s.analysis.References {
		for _, id := range reason.ReferenceIDs {
			if reference.ID == id && reference.Location == *location {
				return true
			}
		}
	}
	return false
}
