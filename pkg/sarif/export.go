package sarif

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

const schemaURI = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"
const completenessRule = "SPL_TOOLING_INCOMPLETE"

// Export projects an accepted corpus report into a detached SARIF 2.1.0 log.
// It does not re-analyze sources or infer locations for unlocated findings.
func Export(report *corpus.Report) (*Log, error) {
	if report == nil || report.SchemaVersion != 1 {
		return nil, fmt.Errorf("sarif: corpus report v1 is required")
	}
	run := Run{Tool: Tool{Driver: Driver{Name: "spl-toolkit", Version: buildinfo.Version, Rules: []Rule{}}}, ColumnKind: "unicodeCodePoints", NewlineSequences: []string{"\r\n", "\r", "\n"}, Artifacts: []Artifact{}, Results: []Result{}, Invocations: []Invocation{{ExecutionSuccessful: true}}, Properties: map[string]any{
		"content_status": report.Status, "execution_complete": report.ExecutionComplete, "mode": report.Mode,
		"coverage": report.Coverage, "coverage_reasons": append([]string{}, report.CoverageReasons...), "counts": report.Counts,
	}}
	artifacts := map[string]int{}
	rules := map[string]bool{}
	seenIDs := map[string]bool{}
	root := ""
	failures := 0
	for i, entry := range report.Entries {
		if entry.ID == "" || seenIDs[entry.ID] {
			return nil, fmt.Errorf("sarif: duplicate or empty document ID at entry %d", i)
		}
		seenIDs[entry.ID] = true
		location, base, err := originLocation(entry.ID, entry.Origin)
		if err != nil {
			return nil, fmt.Errorf("sarif: entry %q: %w", entry.ID, err)
		}
		if base != "" {
			if root != "" && root != base {
				return nil, fmt.Errorf("sarif: multiple file roots in one run")
			}
			root = base
		}
		if entry.Failure != nil {
			if entry.Evaluation != nil {
				return nil, fmt.Errorf("sarif: entry %q has failure and evaluation", entry.ID)
			}
			failures++
			run.Invocations[0].ToolExecutionNotifications = append(run.Invocations[0].ToolExecutionNotifications, failureNotification(*entry.Failure, entry.ID))
			continue
		}
		source, diagnostics, status, coverage, err := evaluated(entry.Evaluation)
		if err != nil {
			return nil, fmt.Errorf("sarif: entry %q: %w", entry.ID, err)
		}
		if !utf8.ValidString(source.Document.Text) || !utf8.ValidString(source.Document.SourceID) {
			return nil, fmt.Errorf("sarif: entry %q has invalid UTF-8 source metadata", entry.ID)
		}
		if entry.SourceHash != corpus.SourceHash(source.Document.Text) {
			return nil, fmt.Errorf("sarif: entry %q source hash mismatch", entry.ID)
		}
		key := location.URIBaseID + "\x00" + location.URI
		artifactIndex, exists := artifacts[key]
		if exists {
			if run.Artifacts[artifactIndex].Hashes["sha-256"] != entry.SourceHash {
				return nil, fmt.Errorf("sarif: conflicting snapshots for artifact %q", location.URI)
			}
		} else {
			artifactIndex = len(run.Artifacts)
			artifacts[key] = artifactIndex
			artifact := Artifact{Location: location, MIMEType: "text/plain", Encoding: "utf-8", SourceLanguage: source.Document.Language, Hashes: map[string]string{"sha-256": entry.SourceHash}, Properties: map[string]any{}}
			if entry.Origin.Kind == "inline" {
				artifact.Contents = &ArtifactContent{Text: source.Document.Text}
			}
			run.Artifacts = append(run.Artifacts, artifact)
		}
		if source.Document.SourceID != "" {
			artifact := &run.Artifacts[artifactIndex]
			ids, _ := artifact.Properties["source_ids"].([]string)
			found := false
			for _, id := range ids {
				if id == source.Document.SourceID {
					found = true
					break
				}
			}
			if !found {
				ids = append(ids, source.Document.SourceID)
				sort.Strings(ids)
			}
			artifact.Properties["source_ids"] = ids
			if len(ids) == 1 {
				artifact.Properties["source_id"] = ids[0]
			} else {
				delete(artifact.Properties, "source_id")
			}
		}
		location.Index = &artifactIndex
		locatedWarning := false
		locatedWarningCodes := map[string]bool{}
		positions := sourceCoordinates(source.Document.Text, diagnostics)
		for _, d := range diagnostics {
			if d.Code == "" || d.Message == "" {
				return nil, fmt.Errorf("sarif: entry %q has incomplete diagnostic identity", entry.ID)
			}
			level, err := diagnosticLevel(d.Severity)
			if err != nil {
				return nil, fmt.Errorf("sarif: entry %q: %w", entry.ID, err)
			}
			region, err := sourceRegion(positions, d.Location)
			if err != nil {
				return nil, fmt.Errorf("sarif: entry %q diagnostic %q: %w", entry.ID, d.Code, err)
			}
			result := Result{RuleID: d.Code, Level: level, Message: Message{Text: d.Message}, Properties: resultProperties(entry, source.Document.SourceID, d.Category, d.StageID, d.ScopeID)}
			if region != nil {
				result.Locations = []Location{{PhysicalLocation: PhysicalLocation{ArtifactLocation: location, Region: region}}}
				if level == "warning" || level == "note" {
					locatedWarning = true
					locatedWarningCodes[d.Code] = true
				}
			}
			rules[d.Code] = true
			run.Results = append(run.Results, result)
		}
		reasons := append([]string{}, source.Coverage.Reasons...)
		if coverage != nil {
			reasons = append(reasons, coverage.Reasons...)
		}
		incomplete := status == analysis.Incomplete || !source.Coverage.SyntaxComplete || !source.Coverage.SemanticComplete || (coverage != nil && !coverage.SchemaComplete)
		sort.Strings(reasons)
		reasons = uniqueStrings(reasons)
		unlocatedReasons := []string{}
		for _, reason := range reasons {
			if !locatedWarningCodes[reason] {
				unlocatedReasons = append(unlocatedReasons, reason)
			}
		}
		if incomplete && (!locatedWarning || len(unlocatedReasons) != 0) {
			msg := "Static analysis is incomplete for document " + entry.ID
			if len(unlocatedReasons) != 0 {
				msg += ": " + strings.Join(unlocatedReasons, ", ")
			}
			result := Result{RuleID: completenessRule, Level: "warning", Message: Message{Text: msg}, Properties: resultProperties(entry, source.Document.SourceID, "tooling_completeness", "", "")}
			result.Properties["coverage_reasons"] = unlocatedReasons
			run.Results = append(run.Results, result)
			rules[completenessRule] = true
		}
	}
	for _, failure := range report.Selection.TraversalFailures {
		failures++
		run.Invocations[0].ToolExecutionNotifications = append(run.Invocations[0].ToolExecutionNotifications, failureNotification(failure, failure.ID))
	}
	if (failures == 0) != report.ExecutionComplete {
		return nil, fmt.Errorf("sarif: execution completeness contradicts operational failures")
	}
	run.Invocations[0].ExecutionSuccessful = failures == 0
	if root != "" {
		run.OriginalURIBases = map[string]ArtifactLocation{sourceRootID: {URI: root}}
	}
	ids := make([]string, 0, len(rules))
	for id := range rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	indices := make(map[string]int, len(ids))
	for i, id := range ids {
		indices[id] = i
		run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, Rule{ID: id})
	}
	for i := range run.Results {
		run.Results[i].RuleIndex = indices[run.Results[i].RuleID]
	}
	return &Log{Schema: schemaURI, Version: "2.1.0", Runs: []Run{run}}, nil
}

type evaluationCoverage struct {
	SchemaComplete bool
	Reasons        []string
}

func evaluated(e *corpus.Evaluation) (*analysis.Result, []analysis.Diagnostic, analysis.Status, *evaluationCoverage, error) {
	if e == nil {
		return nil, nil, "", nil, fmt.Errorf("evaluation is required")
	}
	switch e.Kind {
	case "analysis":
		if e.Analysis != nil {
			return e.Analysis, e.Analysis.Diagnostics, e.Analysis.Status, nil, nil
		}
	case "field_list":
		if e.FieldValidation != nil && e.FieldValidation.Analysis != nil {
			v := e.FieldValidation
			return v.Analysis, v.Diagnostics, v.Status, &evaluationCoverage{v.Coverage.SchemaComplete, v.Coverage.Reasons}, nil
		}
	case "json_schema", "ocsf":
		if e.SchemaValidation != nil && e.SchemaValidation.Analysis != nil {
			v := e.SchemaValidation
			return v.Analysis, v.Diagnostics, v.Status, &evaluationCoverage{v.Coverage.SchemaComplete, v.Coverage.Reasons}, nil
		}
	}
	return nil, nil, "", nil, fmt.Errorf("unrecognized or missing canonical evaluation %q", e.Kind)
}

func diagnosticLevel(severity string) (string, error) {
	switch severity {
	case "error":
		return "error", nil
	case "warning":
		return "warning", nil
	case "info", "note":
		return "note", nil
	default:
		return "", fmt.Errorf("unsupported diagnostic severity %q", severity)
	}
}

func resultProperties(entry corpus.ReportEntry, sourceID, category, stageID, scopeID string) map[string]any {
	p := map[string]any{"document_id": entry.ID, "source_hash": entry.SourceHash, "category": category}
	if sourceID != "" {
		p["source_id"] = sourceID
	}
	if entry.AnalysisRevision != "" {
		p["analysis_revision"] = entry.AnalysisRevision
	}
	if entry.TargetDigest != "" {
		p["target_digest"] = entry.TargetDigest
	}
	if stageID != "" {
		p["stage_id"] = stageID
	}
	if scopeID != "" {
		p["scope_id"] = scopeID
	}
	return p
}

func failureNotification(f corpus.AcquisitionError, documentID string) Notification {
	p := map[string]any{"code": f.Code, "phase": f.Phase}
	if documentID != "" {
		p["document_id"] = documentID
	}
	if f.Path != "" {
		p["path"] = f.Path
	}
	message := f.Message
	if message == "" {
		message = "Corpus operation failed"
		if f.Phase != "" {
			message += " during " + f.Phase
		}
		if documentID != "" {
			message += " for document " + documentID
		}
		if f.Code != "" {
			message += " (" + f.Code + ")"
		}
	}
	return Notification{Level: "error", Message: Message{Text: message}, Properties: p}
}

func uniqueStrings(values []string) []string {
	out := values[:0]
	for _, v := range values {
		if v != "" && (len(out) == 0 || out[len(out)-1] != v) {
			out = append(out, v)
		}
	}
	return out
}
