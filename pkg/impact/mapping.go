package impact

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

type PreparedMappingComparison struct {
	before       MappingSide
	after        MappingSide
	beforeDigest string
	afterDigest  string
}

func PrepareMappings(before, after MappingSide) (*PreparedMappingComparison, error) {
	b, bd, err := prepareMappingSide(before)
	if err != nil {
		return nil, inputError("before: %v", err)
	}
	a, ad, err := prepareMappingSide(after)
	if err != nil {
		return nil, inputError("after: %v", err)
	}
	return &PreparedMappingComparison{before: b, after: a, beforeDigest: bd, afterDigest: ad}, nil
}

func prepareMappingSide(side MappingSide) (MappingSide, string, error) {
	if side.Rules.SchemaVersion != 1 {
		return MappingSide{}, "", inputError("rules schema_version must be integer 1")
	}
	if err := checkRuleStrings(side.Rules.Rules); err != nil {
		return MappingSide{}, "", err
	}
	raw, err := json.Marshal(side.Rules)
	if err != nil {
		return MappingSide{}, "", inputError("rules: %v", err)
	}
	rules, err := rewrite.DecodeRuleSet(raw)
	if err != nil {
		return MappingSide{}, "", inputError("rules: %v", err)
	}
	var target *rewrite.ValidationTarget
	if side.ValidationTarget != nil {
		raw, err := json.Marshal(side.ValidationTarget)
		if err != nil {
			return MappingSide{}, "", inputError("validation target: %v", err)
		}
		target, err = corpus.DecodeValidationTarget(raw)
		if err != nil {
			return MappingSide{}, "", inputError("validation target: %v", err)
		}
		if _, err := corpus.Prepare(corpus.ScanOptions{ValidationTarget: target}); err != nil {
			return MappingSide{}, "", inputError("validation target: %v", err)
		}
	}
	prepared := MappingSide{Rules: rules, ValidationTarget: target}
	raw, err = json.Marshal(struct {
		Rules  rewrite.RuleSet           `json:"rules"`
		Target *rewrite.ValidationTarget `json:"validation_target"`
	}{rules, target})
	if err != nil {
		return MappingSide{}, "", inputError("mapping side: %v", err)
	}
	sum := sha256.Sum256(append([]byte("mapping-side-v1:"), raw...))
	return prepared, hex.EncodeToString(sum[:]), nil
}

func checkRuleStrings(rules []rewrite.Rule) error {
	for _, rule := range rules {
		for _, s := range []string{rule.ID, rule.Kind} {
			if !utf8.ValidString(s) {
				return inputError("rule strings must be valid UTF-8")
			}
		}
		if err := checkIdentityStrings(rule.Source); err != nil {
			return err
		}
		if err := checkIdentityStrings(rule.Target); err != nil {
			return err
		}
		if rule.When != nil {
			if err := checkConditionStrings(*rule.When); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkIdentityStrings(id rewrite.Identity) error {
	if id.Name != nil && !utf8.ValidString(*id.Name) {
		return inputError("identity name must be valid UTF-8")
	}
	for _, s := range id.Path {
		if !utf8.ValidString(s) {
			return inputError("identity path must be valid UTF-8")
		}
	}
	return nil
}

func checkConditionStrings(c rewrite.Condition) error {
	for _, s := range []string{c.Fact, c.Kind, c.Operator} {
		if !utf8.ValidString(s) {
			return inputError("condition strings must be valid UTF-8")
		}
	}
	if c.Identity != nil {
		if err := checkIdentityStrings(*c.Identity); err != nil {
			return err
		}
	}
	if c.Value != nil {
		if err := jsoninput.ValidateUnicode(c.Value); err != nil {
			return inputError("condition value: %v", err)
		}
	}
	for _, child := range c.All {
		if err := checkConditionStrings(child); err != nil {
			return err
		}
	}
	for _, child := range c.Any {
		if err := checkConditionStrings(child); err != nil {
			return err
		}
	}
	return nil
}

func CompareMappings(request MappingRequest) (*Report, error) {
	if request.SchemaVersion != 1 {
		return nil, inputError("schema_version must be integer 1")
	}
	p, err := PrepareMappings(request.Before, request.After)
	if err != nil {
		return nil, err
	}
	return p.Compare(request.Input)
}

func (p *PreparedMappingComparison) Compare(input corpus.Input) (*Report, error) {
	if p == nil || p.before.Rules.SchemaVersion != 1 || p.after.Rules.SchemaVersion != 1 {
		return nil, inputError("mapping comparison is not prepared")
	}
	snapshot := copyInput(input)
	base, err := corpus.Prepare(corpus.ScanOptions{})
	if err != nil {
		return nil, err
	}
	baseline, err := base.Scan(snapshot)
	if err != nil {
		return nil, classifyInputError(err)
	}
	documents := make([]analysis.QueryDocument, 0, baseline.Counts.Analyzed)
	for _, entry := range snapshot.Entries {
		if entry.Document != nil {
			documents = append(documents, *entry.Document)
		}
	}
	var before, after *rewrite.BatchResult
	if len(documents) != 0 {
		before, err = rewrite.RewriteBatch(rewrite.BatchRequest{SchemaVersion: 1, Mode: rewrite.Preview, Documents: documents, Rules: p.before.Rules.Rules, ValidationTarget: p.before.ValidationTarget})
		if err != nil {
			return nil, classifyInputError(err)
		}
		after, err = rewrite.RewriteBatch(rewrite.BatchRequest{SchemaVersion: 1, Mode: rewrite.Preview, Documents: documents, Rules: p.after.Rules.Rules, ValidationTarget: p.after.ValidationTarget})
		if err != nil {
			return nil, classifyInputError(err)
		}
	}
	r := &Report{SchemaVersion: 1, Kind: "mapping", ExecutionComplete: baseline.ExecutionComplete, Selection: copySelection(snapshot.Selection),
		BeforeDigest: p.beforeDigest, AfterDigest: p.afterDigest, Counts: Counts{Selected: len(snapshot.Entries), Denominator: len(snapshot.Entries), TraversalFailed: len(snapshot.Selection.TraversalFailures)}, Entries: make([]ImpactEntry, 0, len(snapshot.Entries))}
	index := 0
	for i, source := range snapshot.Entries {
		e := emptyEntry(source)
		if source.Failure != nil {
			failure := *source.Failure
			e.Failure = &failure
			e.Classification = Failed
			e.Reasons = append(e.Reasons, "acquisition_failed")
			r.append(e)
			continue
		}
		e.SourceHash, e.AnalysisRevision = baseline.Entries[i].SourceHash, baseline.Entries[i].AnalysisRevision
		e.BeforeRewrite, e.AfterRewrite = before.Reports[index], after.Reports[index]
		index++
		bv, av := evidenceFromRewrite(e.BeforeRewrite), evidenceFromRewrite(e.AfterRewrite)
		e.BeforeStatus, e.AfterStatus = bv.status, av.status
		e.Alignment = alignMapping(e.BeforeRewrite, e.AfterRewrite, e.AnalysisRevision)
		e.Deltas = compareEvidence(bv, av, e.Alignment)
		e.Classification, e.Reasons = classify(e.Deltas, bv.complete && av.complete, e.BeforeRewrite.CandidateText == e.AfterRewrite.CandidateText && (!bv.complete || !av.complete), e.Alignment)
		r.append(e)
	}
	return r, nil
}
