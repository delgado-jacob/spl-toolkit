package corpus

import (
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

// Scan evaluates already-acquired snapshots in selection order.
func (p *PreparedScan) Scan(input Input) (*Report, error) {
	if p == nil || p.mode == "" {
		return nil, inputError("scan is not prepared")
	}
	if input.Selection.Mode == "" {
		return nil, inputError("selection mode is required")
	}
	if input.Selection.Complete != (len(input.Selection.TraversalFailures) == 0) {
		return nil, inputError("selection completeness must match traversal failures")
	}
	if len(input.Entries) == 0 && input.Selection.Complete {
		return nil, inputError("selection must be nonempty")
	}
	seen := make(map[string]bool, len(input.Entries))
	revisions := make([]string, len(input.Entries))
	for i, entry := range input.Entries {
		if !utf8.ValidString(entry.ID) || strings.TrimSpace(entry.ID) == "" || seen[entry.ID] {
			return nil, inputError("entry %d has invalid or duplicate id", i)
		}
		seen[entry.ID] = true
		if (entry.Document == nil) == (entry.Failure == nil) {
			return nil, inputError("entry %d requires exactly one document or acquisition failure", i)
		}
		if entry.Document != nil && entry.SourceHash != "" && entry.SourceHash != SourceHash(entry.Document.Text) {
			return nil, inputError("entry %d source hash does not match document", i)
		}
		if entry.Document != nil {
			var err error
			revisions[i], err = AnalysisRevision(*entry.Document)
			if err != nil {
				return nil, err
			}
		}
	}
	r := newReport(input.Selection, p.mode)
	r.Counts.Selected = len(input.Entries)
	r.Counts.TraversalFailed = len(input.Selection.TraversalFailures)
	if !input.Selection.Complete || r.Counts.TraversalFailed != 0 {
		r.ExecutionComplete = false
		r.Status = analysis.Incomplete
	}
	for i, entry := range input.Entries {
		out := ReportEntry{ID: entry.ID, Origin: entry.Origin, SourceHash: entry.SourceHash}
		if entry.Failure != nil {
			failure := *entry.Failure
			out.Failure = &failure
			r.Counts.AcquisitionFailed++
			r.ExecutionComplete = false
			r.Status = reduceStatus(r.Status, analysis.Incomplete)
			r.Entries = append(r.Entries, out)
			continue
		}
		doc := *entry.Document
		out.SourceHash = SourceHash(doc.Text)
		var result *analysis.Result
		out.Evaluation = &Evaluation{Kind: p.mode}
		switch p.mode {
		case "analysis":
			var err error
			result, err = analysis.Analyze(doc)
			if err != nil {
				return nil, err
			}
			out.Evaluation.Analysis = result
		case "field_list":
			report, err := p.field.Validate(doc)
			if err != nil {
				return nil, err
			}
			out.Evaluation.FieldValidation = report
			result = report.Analysis
		case "json_schema", "ocsf":
			report, err := p.schema.Validate(doc)
			if err != nil {
				return nil, err
			}
			out.Evaluation.SchemaValidation = report
			result = report.Analysis
		default:
			return nil, inputError("unsupported prepared mode %q", p.mode)
		}
		out.AnalysisRevision = revisions[i]
		out.TargetDigest = p.digest
		r.Counts.Analyzed++
		addCoverage(&r.Coverage.Syntax, result.Coverage.SyntaxComplete)
		addCoverage(&r.Coverage.Semantic, result.Coverage.SemanticComplete)
		status := result.Status
		if p.mode == "field_list" {
			v := out.Evaluation.FieldValidation
			addCoverage(&r.Coverage.Schema, v.Coverage.SchemaComplete)
			status = v.Status
		} else if p.mode != "analysis" {
			v := out.Evaluation.SchemaValidation
			addCoverage(&r.Coverage.Schema, v.Coverage.SchemaComplete)
			status = v.Status
		}
		r.Status = reduceStatus(r.Status, status)
		addStatus(&r.StatusCounts, status)
		addCoverageReasons(r, result.Coverage.Reasons)
		if p.mode == "field_list" {
			addCoverageReasons(r, out.Evaluation.FieldValidation.Coverage.Reasons)
		}
		if p.mode != "field_list" && p.mode != "analysis" {
			addCoverageReasons(r, out.Evaluation.SchemaValidation.Coverage.Reasons)
		}
		if err := appendObservedCoverage(r, result); err != nil {
			return nil, err
		}
		appendDependencies(r, entry.ID, result)
		r.Entries = append(r.Entries, out)
	}
	sortDependencies(r)
	return r, nil
}

// Scan constructs an in-memory selection without invoking the local loader.
func Scan(request Request) (*Report, error) {
	if request.SchemaVersion != 1 || len(request.Documents) == 0 {
		return nil, inputError("request requires schema_version 1 and nonempty documents")
	}
	p, err := Prepare(ScanOptions{ValidationTarget: request.ValidationTarget})
	if err != nil {
		return nil, err
	}
	in := Input{Selection: Selection{Mode: "inline", Complete: true, IgnoredNames: []string{}, SkippedSymlinks: []string{}, TraversalFailures: []AcquisitionError{}}, Entries: make([]Entry, 0, len(request.Documents))}
	for _, item := range request.Documents {
		document := item.Document
		in.Entries = append(in.Entries, Entry{ID: item.ID, Origin: Origin{Kind: "inline"}, Document: &document})
	}
	return p.Scan(in)
}
