package analysis

// QueryDocument retains the original query and selects its compatibility contract.
type QueryDocument struct {
	Text     string `json:"text"`
	Language string `json:"language"`
	Profile  string `json:"profile"`
	Version  string `json:"version"`
	SourceID string `json:"source_id"`
}
type Status string

const (
	Valid      Status = "valid"
	Invalid    Status = "invalid"
	Incomplete Status = "incomplete"
)

type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}
type Location struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}
type Coverage struct {
	SyntaxComplete   bool     `json:"syntax_complete"`
	SemanticComplete bool     `json:"semantic_complete"`
	Reasons          []string `json:"reasons"`
}
type Stage struct {
	ID               string   `json:"id"`
	Command          string   `json:"command"`
	Position         int      `json:"position"`
	ScopeID          string   `json:"scope_id"`
	Location         Location `json:"location"`
	SemanticComplete bool     `json:"semantic_complete"`
}
type Scope struct {
	ID       string   `json:"id"`
	ParentID string   `json:"parent_id"`
	Kind     string   `json:"kind"`
	StageID  string   `json:"stage_id"`
	Location Location `json:"location"`
}
type Reference struct {
	ID                 string   `json:"id"`
	OriginalName       string   `json:"original_name"`
	NormalizedName     string   `json:"normalized_name"`
	Kind               string   `json:"kind"`
	Role               string   `json:"role"`
	StageID            string   `json:"stage_id"`
	ScopeID            string   `json:"scope_id"`
	Location           Location `json:"location"`
	Resolution         string   `json:"resolution"`
	Binding            string   `json:"binding"`
	OriginReferenceIDs []string `json:"origin_reference_ids"`
}
type FieldBinding struct {
	Name               string   `json:"name"`
	OriginReferenceIDs []string `json:"origin_reference_ids"`
	Conditional        bool     `json:"conditional"`
}
type FieldState struct {
	Fields    []FieldBinding `json:"fields"`
	Removed   []string       `json:"removed"`
	Open      bool           `json:"open"`
	Uncertain bool           `json:"uncertain"`
}
type Transition struct {
	Operation         string   `json:"operation"`
	Output            string   `json:"output"`
	InputReferenceIDs []string `json:"input_reference_ids"`
	OutputReferenceID string   `json:"output_reference_id"`
	Conditional       bool     `json:"conditional"`
}
type Lineage struct {
	StageID     string       `json:"stage_id"`
	ScopeID     string       `json:"scope_id"`
	Before      FieldState   `json:"before"`
	After       FieldState   `json:"after"`
	Transitions []Transition `json:"transitions"`
}
type Dependencies struct {
	Indexes     []string `json:"indexes"`
	Sources     []string `json:"sources"`
	SourceTypes []string `json:"source_types"`
	Datasets    []string `json:"datasets"`
	Lookups     []string `json:"lookups"`
	DataModels  []string `json:"data_models"`
	Macros      []string `json:"macros"`
}
type Diagnostic struct {
	Code     string   `json:"code"`
	Severity string   `json:"severity"`
	Category string   `json:"category"`
	Message  string   `json:"message"`
	Location Location `json:"location"`
	StageID  string   `json:"stage_id"`
	ScopeID  string   `json:"scope_id"`
}
type Result struct {
	SchemaVersion int           `json:"schema_version"`
	Document      QueryDocument `json:"document"`
	Status        Status        `json:"status"`
	Coverage      Coverage      `json:"coverage"`
	Stages        []Stage       `json:"stages"`
	Scopes        []Scope       `json:"scopes"`
	References    []Reference   `json:"references"`
	Lineage       []Lineage     `json:"lineage"`
	Dependencies  Dependencies  `json:"dependencies"`
	Diagnostics   []Diagnostic  `json:"diagnostics"`
}
type Capability struct {
	Name              string   `json:"name"`
	SyntaxSupported   bool     `json:"syntax_supported"`
	SemanticSupported bool     `json:"semantic_supported"`
	Limitations       []string `json:"limitations"`
}
type CapabilityManifest struct {
	SchemaVersion int          `json:"schema_version"`
	Language      string       `json:"language"`
	Profile       string       `json:"profile"`
	Version       string       `json:"version"`
	Commands      []Capability `json:"commands"`
	Functions     []Capability `json:"functions"`
}
