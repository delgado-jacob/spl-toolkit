package impact

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

type PreparedSchemaComparison struct {
	before       *corpus.PreparedScan
	after        *corpus.PreparedScan
	beforeDigest string
	afterDigest  string
}

// PrepareSchemas compiles both canonical targets before a caller acquires files.
func PrepareSchemas(before, after corpus.ValidationTarget) (*PreparedSchemaComparison, error) {
	b, err := corpus.Prepare(corpus.ScanOptions{ValidationTarget: &before})
	if err != nil {
		return nil, inputError("before_target: %v", err)
	}
	a, err := corpus.Prepare(corpus.ScanOptions{ValidationTarget: &after})
	if err != nil {
		return nil, inputError("after_target: %v", err)
	}
	bd, err := corpus.TargetDigest(&before)
	if err != nil {
		return nil, inputError("before_target: %v", err)
	}
	ad, err := corpus.TargetDigest(&after)
	if err != nil {
		return nil, inputError("after_target: %v", err)
	}
	return &PreparedSchemaComparison{before: b, after: a, beforeDigest: bd, afterDigest: ad}, nil
}

func CompareSchemas(request SchemaRequest) (*Report, error) {
	if request.SchemaVersion != 1 {
		return nil, inputError("schema_version must be integer 1")
	}
	p, err := PrepareSchemas(request.BeforeTarget, request.AfterTarget)
	if err != nil {
		return nil, err
	}
	return p.Compare(request.Input)
}

func (p *PreparedSchemaComparison) Compare(input corpus.Input) (*Report, error) {
	if p == nil || p.before == nil || p.after == nil {
		return nil, inputError("schema comparison is not prepared")
	}
	snapshot := copyInput(input)
	before, err := p.before.Scan(snapshot)
	if err != nil {
		return nil, classifyInputError(err)
	}
	after, err := p.after.Scan(snapshot)
	if err != nil {
		return nil, classifyInputError(err)
	}
	r := &Report{SchemaVersion: 1, Kind: "schema", ExecutionComplete: before.ExecutionComplete && after.ExecutionComplete,
		Selection: copySelection(snapshot.Selection), BeforeDigest: p.beforeDigest, AfterDigest: p.afterDigest,
		Counts: Counts{Selected: len(snapshot.Entries), Denominator: len(snapshot.Entries), TraversalFailed: len(snapshot.Selection.TraversalFailures)}, Entries: make([]ImpactEntry, 0, len(snapshot.Entries))}
	for i, source := range snapshot.Entries {
		e := emptyEntry(source)
		b, a := before.Entries[i], after.Entries[i]
		if source.Failure != nil {
			failure := *source.Failure
			e.Failure = &failure
			e.Classification = Failed
			e.Reasons = append(e.Reasons, "acquisition_failed")
			r.append(e)
			continue
		}
		e.SourceHash, e.AnalysisRevision = b.SourceHash, b.AnalysisRevision
		e.Before, e.After = b.Evaluation, a.Evaluation
		bv, av := evidenceFromEvaluation(e.Before), evidenceFromEvaluation(e.After)
		e.BeforeStatus, e.AfterStatus = bv.status, av.status
		e.Alignment = alignOriginal(bv.analysis, av.analysis, b.AnalysisRevision, a.AnalysisRevision)
		e.Deltas = compareEvidence(bv, av, e.Alignment)
		e.Classification, e.Reasons = classify(e.Deltas, bv.complete && av.complete, false, e.Alignment)
		r.append(e)
	}
	return r, nil
}
