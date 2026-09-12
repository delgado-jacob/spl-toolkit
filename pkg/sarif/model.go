// Package sarif projects canonical corpus findings into SARIF 2.1.0.
package sarif

// Log and its children intentionally use SARIF's standardized member names.
// Tool-specific status, provenance, and coverage stay in property bags.
type Log struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

type Run struct {
	Tool             Tool                        `json:"tool"`
	ColumnKind       string                      `json:"columnKind"`
	NewlineSequences []string                    `json:"newlineSequences"`
	OriginalURIBases map[string]ArtifactLocation `json:"originalUriBaseIds,omitempty"`
	Artifacts        []Artifact                  `json:"artifacts"`
	Results          []Result                    `json:"results"`
	Invocations      []Invocation                `json:"invocations"`
	Properties       map[string]any              `json:"properties,omitempty"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}
type Driver struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	Rules   []Rule `json:"rules"`
}
type Rule struct {
	ID string `json:"id"`
}
type Message struct {
	Text string `json:"text"`
}
type ArtifactContent struct {
	Text string `json:"text"`
}
type ArtifactLocation struct {
	URI       string `json:"uri"`
	URIBaseID string `json:"uriBaseId,omitempty"`
	Index     *int   `json:"index,omitempty"`
}
type Artifact struct {
	Location       ArtifactLocation  `json:"location"`
	MIMEType       string            `json:"mimeType,omitempty"`
	Encoding       string            `json:"encoding,omitempty"`
	SourceLanguage string            `json:"sourceLanguage,omitempty"`
	Contents       *ArtifactContent  `json:"contents,omitempty"`
	Hashes         map[string]string `json:"hashes,omitempty"`
	Properties     map[string]any    `json:"properties,omitempty"`
}
type Region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
}
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
	Region           *Region          `json:"region,omitempty"`
}
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}
type Result struct {
	RuleID     string         `json:"ruleId"`
	RuleIndex  int            `json:"ruleIndex"`
	Level      string         `json:"level"`
	Message    Message        `json:"message"`
	Locations  []Location     `json:"locations,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}
type Notification struct {
	Level      string         `json:"level"`
	Message    Message        `json:"message"`
	Properties map[string]any `json:"properties,omitempty"`
}
type Invocation struct {
	ExecutionSuccessful        bool           `json:"executionSuccessful"`
	ToolExecutionNotifications []Notification `json:"toolExecutionNotifications,omitempty"`
}
