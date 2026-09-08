package validation

import (
	"encoding/json"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type OCSFSelection struct {
	Version     string   `json:"version"`
	Class       string   `json:"class,omitempty"`
	ClassUID    *int64   `json:"class_uid,omitempty"`
	Category    string   `json:"category,omitempty"`
	CategoryUID *int64   `json:"category_uid,omitempty"`
	Profiles    []string `json:"profiles"`
	Extensions  []string `json:"extensions"`
}
type SchemaTarget struct {
	Kind      string                     `json:"kind"`
	Identity  string                     `json:"identity,omitempty"`
	Schema    json.RawMessage            `json:"schema,omitempty"`
	BaseURI   string                     `json:"base_uri,omitempty"`
	Resources map[string]json.RawMessage `json:"resources,omitempty"`
	Catalog   json.RawMessage            `json:"catalog,omitempty"`
	Selection *OCSFSelection             `json:"selection,omitempty"`
}
type SchemaClass struct {
	Key string `json:"key"`
	UID int64  `json:"uid"`
}
type SchemaExtension struct {
	Key     string `json:"key"`
	UID     int64  `json:"uid"`
	Version string `json:"version"`
}
type SchemaTargetInfo struct {
	Kind           string            `json:"kind"`
	Identity       string            `json:"identity,omitempty"`
	Version        string            `json:"version,omitempty"`
	Dialect        string            `json:"dialect,omitempty"`
	BaseURI        string            `json:"base_uri,omitempty"`
	ResourceURIs   []string          `json:"resource_uris"`
	CompileVersion int               `json:"compile_version,omitempty"`
	Selection      *OCSFSelection    `json:"selection,omitempty"`
	Members        []SchemaClass     `json:"members"`
	Extensions     []SchemaExtension `json:"extensions"`
	Limitations    []string          `json:"limitations"`
}
type SchemaEvidence struct {
	ResourceURI      string `json:"resource_uri,omitempty"`
	Pointer          string `json:"pointer,omitempty"`
	Keyword          string `json:"keyword,omitempty"`
	Operator         string `json:"operator,omitempty"`
	Branch           string `json:"branch,omitempty"`
	ClassKey         string `json:"class_key,omitempty"`
	ClassUID         *int64 `json:"class_uid,omitempty"`
	DeclarationBasis string `json:"declaration_basis,omitempty"`
	Requirement      string `json:"requirement,omitempty"`
	Reason           string `json:"reason,omitempty"`
}
type fieldProjection struct {
	Admission            analysis.SourceFieldAdmission
	Outcome              string
	Evidence             []SchemaEvidence
	SupportingClasses    []SchemaClass
	MissingClasses       []SchemaClass
	IndeterminateClasses []SchemaClass
}
type preparedSchemaTarget interface {
	project(string) fieldProjection
	universe() analysis.SourceUniverse
	info() SchemaTargetInfo
}
