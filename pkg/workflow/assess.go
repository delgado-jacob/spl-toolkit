package workflow

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func AssessJSON(raw []byte) (*Report, error) {
	request, err := DecodeRequest(raw)
	if err != nil {
		return nil, err
	}
	return Assess(request)
}

func Assess(request Request) (*Report, error) {
	// Prepare global artifacts before admitting any selected document.
	prepared, err := Prepare(request.Settings)
	if err != nil {
		return nil, err
	}
	admitted, err := normalizeRequest(request)
	if err != nil {
		return nil, err
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true, IgnoredNames: []string{}, SkippedSymlinks: []string{}, TraversalFailures: []corpus.AcquisitionError{}}, Entries: make([]corpus.Entry, 0, len(admitted.Documents))}
	for _, entry := range admitted.Documents {
		doc := entry.Document
		input.Entries = append(input.Entries, corpus.Entry{ID: entry.ID, Origin: corpus.Origin{Kind: "inline"}, SourceHash: corpus.SourceHash(doc.Text), Document: &doc})
	}
	return prepared.Assess(input)
}

func (p *Prepared) admit(input corpus.Input) (corpus.Input, error) {
	if !validUTF8(reflect.ValueOf(input)) {
		return corpus.Input{}, requestErrorAt("request_invalid", "/selection", "input contains invalid UTF-8")
	}
	if strings.TrimSpace(input.Selection.Mode) == "" {
		return corpus.Input{}, requestErrorAt("request_invalid", "/selection/mode", "selection mode is required")
	}
	if input.Selection.Complete != (len(input.Selection.TraversalFailures) == 0) {
		return corpus.Input{}, requestErrorAt("request_invalid", "/selection/complete", "selection completeness must match traversal failures")
	}
	if len(input.Entries) == 0 && input.Selection.Complete {
		return corpus.Input{}, requestErrorAt("request_invalid", "/entries", "complete selection must be nonempty")
	}
	out := detach(input)
	seen := map[string]bool{}
	for i := range out.Entries {
		entry := &out.Entries[i]
		path := fmt.Sprintf("/entries/%d", i)
		if strings.TrimSpace(entry.ID) == "" || seen[entry.ID] {
			return corpus.Input{}, requestErrorAt("request_invalid", path+"/id", "id must be unique and nonblank")
		}
		seen[entry.ID] = true
		if (entry.Document == nil) == (entry.Failure == nil) {
			return corpus.Input{}, requestErrorAt("request_invalid", path, "exactly one document or acquisition failure is required")
		}
		if _, ok := p.entries[entry.ID]; !ok {
			return corpus.Input{}, requestErrorAt("request_invalid", path+"/id", "missing entry settings")
		}
		if entry.Document != nil {
			if entry.SourceHash != "" && entry.SourceHash != corpus.SourceHash(entry.Document.Text) {
				return corpus.Input{}, requestErrorAt("request_invalid", path+"/source_hash", "source hash does not match document")
			}
			selected, err := capabilityselector.Normalize(entry.Document.Language, entry.Document.Profile, entry.Document.Version)
			if err != nil {
				return corpus.Input{}, requestErrorAt("request_invalid", path+"/document", err.Error())
			}
			entry.Document.Language, entry.Document.Profile, entry.Document.Version = selected.Language, selected.Profile, selected.Version
		}
	}
	for _, entry := range p.settings.Entries {
		if !seen[entry.ID] {
			return corpus.Input{}, requestErrorAt("request_invalid", "/settings/entries", "settings id is not selected: "+entry.ID)
		}
	}
	return out, nil
}

// Assess retains acquired failures and evaluates admitted documents in selection order.
func (p *Prepared) Assess(input corpus.Input) (*Report, error) {
	if p == nil || p.compatibility == nil {
		return nil, requestErrorAt("request_invalid", "", "workflow is not prepared")
	}
	admitted, err := p.admit(input)
	if err != nil {
		return nil, err
	}
	report := &Report{SchemaVersion: 1, Status: analysis.Valid, ExecutionComplete: admitted.Selection.Complete, Selection: admitted.Selection, Entries: make([]ReportEntry, 0, len(admitted.Entries)), Provenance: detach(p.provenance)}
	report.Counts.Selected, report.Counts.TraversalFailed = len(admitted.Entries), len(admitted.Selection.TraversalFailures)
	for _, entry := range admitted.Entries {
		setting := p.entries[entry.ID]
		out := ReportEntry{ID: entry.ID, Origin: entry.Origin, SourceHash: entry.SourceHash, Mode: "compatibility", Status: analysis.Incomplete}
		if setting.Resolution != nil {
			out.Mode = "resolution"
		}
		if entry.Failure != nil {
			detail, _ := json.Marshal(entry.Failure)
			out.Failure = &Failure{Phase: "acquisition", Code: entry.Failure.Code, Message: entry.Failure.Message, Detail: detail}
		} else {
			out.SourceHash = corpus.SourceHash(entry.Document.Text)
			out.Analysis, err = analysis.Analyze(*entry.Document)
			if err != nil {
				out.Failure = operationFailure(err, "analysis")
			} else if setting.Compatibility == nil {
				out.Failure = &Failure{Phase: "configuration", Code: "mode_unavailable", Message: "resolution orchestration is not available"}
			} else {
				check := detach(*setting.Compatibility)
				out.Compatibility, err = p.compatibility.Check(compatibility.AssessmentRequest{SchemaVersion: 1, Requirements: out.Analysis.Requirements, Document: entry.Document, QueryScope: check.QueryScope, InputBindings: check.InputBindings, DependencyBindings: check.DependencyBindings})
				if err != nil {
					out.Failure = operationFailure(err, "compatibility")
				} else {
					out.Status = compatibilityStatus(out.Analysis, out.Compatibility)
				}
			}
		}
		report.Entries = append(report.Entries, out)
	}
	finalize(report)
	return report, nil
}

func operationFailure(err error, operation string) *Failure {
	phase, code := "internal", "operation_failed"
	if validation.IsInputError(err) {
		phase, code = "configuration", "configuration_invalid"
	}
	failure := &Failure{Phase: phase, Code: code, Message: operation + ": " + err.Error()}
	if detail, ok := compatibility.RequestErrorDetails(err); ok {
		failure.Code, failure.Message = detail.Code, detail.Message
		failure.Detail, _ = json.Marshal(detail)
	}
	return failure
}
