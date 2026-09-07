// Package validation validates query field obligations against an offline catalog.
package validation

import "github.com/delgado-jacob/spl-toolkit/pkg/analysis"

// FieldCatalog declares concrete names, without event-presence or type claims.
type FieldCatalog struct {
	Fields         []string `json:"fields"`
	OptionalFields []string `json:"optional_fields"`
	Identity       string   `json:"identity"`
	Version        string   `json:"version"`
}
type Target struct {
	Kind string `json:"kind"`
	FieldCatalog
}
type Coverage struct {
	SyntaxComplete   bool     `json:"syntax_complete"`
	SemanticComplete bool     `json:"semantic_complete"`
	SchemaComplete   bool     `json:"schema_complete"`
	Reasons          []string `json:"reasons"`
}
type Match struct {
	Name    string `json:"name"`
	Binding string `json:"binding"`
	Outcome string `json:"outcome"`
}
type ReferenceOutcome struct {
	ReferenceID string  `json:"reference_id"`
	Outcome     string  `json:"outcome"`
	Matches     []Match `json:"matches"`
}
type Report struct {
	SchemaVersion int                   `json:"schema_version"`
	Target        Target                `json:"target"`
	Analysis      *analysis.Result      `json:"analysis"`
	Status        analysis.Status       `json:"status"`
	Coverage      Coverage              `json:"coverage"`
	Outcomes      []ReferenceOutcome    `json:"outcomes"`
	Diagnostics   []analysis.Diagnostic `json:"diagnostics"`
}
type BatchReport struct {
	SchemaVersion int             `json:"schema_version"`
	Status        analysis.Status `json:"status"`
	Reports       []*Report       `json:"reports"`
}
type Request struct {
	Document analysis.QueryDocument `json:"document"`
	Catalog  FieldCatalog           `json:"catalog"`
}
type BatchRequest struct {
	Documents []analysis.QueryDocument `json:"documents"`
	Catalog   FieldCatalog             `json:"catalog"`
}
