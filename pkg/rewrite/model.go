// Package rewrite defines safe, explicit query rewrite requests and reports.
package rewrite

import (
	"encoding/json"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

type Mode string

const (
	Preview Mode = "preview"
	Apply   Mode = "apply"
)

// Identity is a logical atom or an exact field path, never an SPL fragment.
// A dotted Name and a multi-segment Path are distinct identities.
type Identity struct {
	Name *string  `json:"name,omitempty"`
	Path []string `json:"path,omitempty"`
}

// Condition contains exactly one combinator or fact. Value retains the exact
// JSON scalar: nil means missing, while RawMessage("null") is present data.
type Condition struct {
	All      []Condition     `json:"all,omitempty"`
	Any      []Condition     `json:"any,omitempty"`
	Fact     string          `json:"fact,omitempty"`
	Kind     string          `json:"kind,omitempty"`
	Identity *Identity       `json:"identity,omitempty"`
	Operator string          `json:"operator,omitempty"`
	Value    json.RawMessage `json:"value,omitempty"`
}

type Rule struct {
	ID     string     `json:"id"`
	Kind   string     `json:"kind"`
	Source Identity   `json:"source"`
	Target Identity   `json:"target"`
	When   *Condition `json:"when,omitempty"`
}

type Request struct {
	SchemaVersion    int                    `json:"schema_version"`
	Mode             Mode                   `json:"mode"`
	Document         analysis.QueryDocument `json:"document"`
	Rules            []Rule                 `json:"rules"`
	ValidationTarget *ValidationTarget      `json:"validation_target,omitempty"`
}

type BatchRequest struct {
	SchemaVersion    int                      `json:"schema_version"`
	Mode             Mode                     `json:"mode"`
	Documents        []analysis.QueryDocument `json:"documents"`
	Rules            []Rule                   `json:"rules"`
	ValidationTarget *ValidationTarget        `json:"validation_target,omitempty"`
}

// RuleSet is the CLI file contract. It cannot select apply mode or a document.
type RuleSet struct {
	SchemaVersion int    `json:"schema_version"`
	Rules         []Rule `json:"rules"`
}

type Coverage struct {
	SyntaxComplete   bool                 `json:"syntax_complete"`
	SemanticComplete bool                 `json:"semantic_complete"`
	RewriteComplete  bool                 `json:"rewrite_complete"`
	Validation       *validation.Coverage `json:"validation,omitempty"`
	Reasons          []string             `json:"reasons"`
}

// Change describes candidate inclusion independently from committed text.
// Locations are present only when an actual reference supplies the location.
type Change struct {
	Outcome               string             `json:"outcome"`
	Reason                string             `json:"reason"`
	GroupID               string             `json:"group_id"`
	RuleIDs               []string           `json:"rule_ids"`
	OriginalReferenceIDs  []string           `json:"original_reference_ids"`
	CandidateReferenceIDs []string           `json:"candidate_reference_ids"`
	OriginalLocation      *analysis.Location `json:"original_location,omitempty"`
	CandidateLocation     *analysis.Location `json:"candidate_location,omitempty"`
	OldText               string             `json:"old_text"`
	NewText               string             `json:"new_text"`
	CandidateApplied      bool               `json:"candidate_applied"`
	Committed             bool               `json:"committed"`
}

// ConditionEvaluation retains every child, even when another child determines
// the all/any result. State is true, false, or unknown; reference IDs are real.
type ConditionEvaluation struct {
	State        string                `json:"state"`
	Reason       string                `json:"reason"`
	ReferenceIDs []string              `json:"reference_ids"`
	Children     []ConditionEvaluation `json:"children"`
}

// RuleEvaluation can explain no_match without inventing a source reference.
type RuleEvaluation struct {
	RuleID       string               `json:"rule_id"`
	Outcome      string               `json:"outcome"`
	Reason       string               `json:"reason"`
	ReferenceIDs []string             `json:"reference_ids"`
	Location     *analysis.Location   `json:"location,omitempty"`
	Condition    *ConditionEvaluation `json:"condition,omitempty"`
}

type Result struct {
	SchemaVersion       int                    `json:"schema_version"`
	Document            analysis.QueryDocument `json:"document"`
	Mode                Mode                   `json:"mode"`
	Status              analysis.Status        `json:"status"`
	Coverage            Coverage               `json:"coverage"`
	OriginalText        string                 `json:"original_text"`
	CandidateText       string                 `json:"candidate_text"`
	Text                string                 `json:"text"`
	Committed           bool                   `json:"committed"`
	Changes             []Change               `json:"changes"`
	RuleEvaluations     []RuleEvaluation       `json:"rule_evaluations"`
	OriginalAnalysis    *analysis.Result       `json:"original_analysis"`
	CandidateAnalysis   *analysis.Result       `json:"candidate_analysis"`
	CandidateValidation *CandidateValidation   `json:"candidate_validation,omitempty"`
}

type BatchResult struct {
	SchemaVersion int             `json:"schema_version"`
	Status        analysis.Status `json:"status"`
	Reports       []*Result       `json:"reports"`
}

// Marshal methods keep public collections as arrays without mutating callers.
func (r Request) MarshalJSON() ([]byte, error) {
	type wire Request
	r.Rules = append([]Rule{}, r.Rules...)
	return json.Marshal(wire(r))
}
func (r BatchRequest) MarshalJSON() ([]byte, error) {
	type wire BatchRequest
	r.Rules = append([]Rule{}, r.Rules...)
	r.Documents = append([]analysis.QueryDocument{}, r.Documents...)
	return json.Marshal(wire(r))
}
func (r RuleSet) MarshalJSON() ([]byte, error) {
	type wire RuleSet
	r.Rules = append([]Rule{}, r.Rules...)
	return json.Marshal(wire(r))
}
func (c Coverage) MarshalJSON() ([]byte, error) {
	type wire Coverage
	c.Reasons = append([]string{}, c.Reasons...)
	return json.Marshal(wire(c))
}
func (c Change) MarshalJSON() ([]byte, error) {
	type wire Change
	c.RuleIDs = append([]string{}, c.RuleIDs...)
	c.OriginalReferenceIDs = append([]string{}, c.OriginalReferenceIDs...)
	c.CandidateReferenceIDs = append([]string{}, c.CandidateReferenceIDs...)
	return json.Marshal(wire(c))
}
func (c ConditionEvaluation) MarshalJSON() ([]byte, error) {
	type wire ConditionEvaluation
	c.ReferenceIDs = append([]string{}, c.ReferenceIDs...)
	c.Children = append([]ConditionEvaluation{}, c.Children...)
	return json.Marshal(wire(c))
}
func (r RuleEvaluation) MarshalJSON() ([]byte, error) {
	type wire RuleEvaluation
	r.ReferenceIDs = append([]string{}, r.ReferenceIDs...)
	return json.Marshal(wire(r))
}
func (r Result) MarshalJSON() ([]byte, error) {
	type wire Result
	r.Changes = append([]Change{}, r.Changes...)
	r.RuleEvaluations = append([]RuleEvaluation{}, r.RuleEvaluations...)
	return json.Marshal(wire(r))
}
func (r BatchResult) MarshalJSON() ([]byte, error) {
	type wire BatchResult
	r.Reports = append([]*Result{}, r.Reports...)
	return json.Marshal(wire(r))
}
