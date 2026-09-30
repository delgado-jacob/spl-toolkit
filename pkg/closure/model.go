// Package closure defines caller-supplied knowledge-object inputs for offline dependency closure.
package closure

import "github.com/delgado-jacob/spl-toolkit/pkg/analysis"

type Request struct {
	SchemaVersion int                    `json:"schema_version"`
	Document      analysis.QueryDocument `json:"document"`
	Bundle        DefinitionBundle       `json:"bundle"`
	Bindings      []Binding              `json:"bindings"`
}

type DefinitionBundle struct {
	SchemaVersion int          `json:"schema_version"`
	ScopeID       string       `json:"scope_id"`
	Collections   []Collection `json:"collections"`
	Objects       []Definition `json:"objects"`
}

type Collection struct {
	Kind     string `json:"kind"`
	Coverage string `json:"coverage"`
}

type Definition struct {
	ID         string                  `json:"id"`
	Kind       string                  `json:"kind"`
	Name       string                  `json:"name"`
	App        string                  `json:"app,omitempty"`
	Owner      string                  `json:"owner,omitempty"`
	Sharing    string                  `json:"sharing,omitempty"`
	SourceID   string                  `json:"source_id"`
	Document   *analysis.QueryDocument `json:"document,omitempty"`
	Arity      *int                    `json:"arity,omitempty"`
	Arguments  []string                `json:"arguments,omitempty"`
	EvalBased  *bool                   `json:"eval_based,omitempty"`
	Validation *string                 `json:"validation,omitempty"`
	Relations  []Relation              `json:"relations"`
}

type Relation struct {
	Kind     string  `json:"kind"`
	Name     string  `json:"name"`
	Start    *int    `json:"start,omitempty"`
	End      *int    `json:"end,omitempty"`
	Property *string `json:"property,omitempty"`
}

type Binding struct {
	DocumentDigest string `json:"document_digest"`
	Kind           string `json:"kind"`
	Start          int    `json:"start"`
	End            int    `json:"end"`
	ObjectID       string `json:"object_id"`
}
