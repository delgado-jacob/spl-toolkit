package analysis

type ResolutionChoice struct {
	Placeholder string `json:"placeholder"`
	Kind        string `json:"kind"`
	Value       string `json:"value"`
}
type ResolutionPlaceholder struct {
	Placeholder      string     `json:"placeholder"`
	Kind             string     `json:"kind"`
	ReferenceIDs     []string   `json:"reference_ids"`
	OriginalInputIDs []string   `json:"original_input_ids"`
	Locations        []Location `json:"locations"`
}
type ResolutionEvidence struct {
	Analysis     Result                  `json:"analysis"`
	Placeholders []ResolutionPlaceholder `json:"placeholders"`
	Coverage     InputCoverage           `json:"coverage"`
}
type ResolutionChange struct {
	Placeholder           string    `json:"placeholder"`
	Kind                  string    `json:"kind"`
	Value                 string    `json:"value"`
	OriginalReferenceIDs  []string  `json:"original_reference_ids"`
	CandidateReferenceIDs []string  `json:"candidate_reference_ids"`
	OriginalLocation      Location  `json:"original_location"`
	CandidateLocation     *Location `json:"candidate_location,omitempty"`
	Before                string    `json:"before"`
	After                 string    `json:"after"`
}
type ResolutionOccurrencePair struct {
	OriginalInputID       string `json:"original_input_id"`
	CandidateInputID      string `json:"candidate_input_id"`
	OriginalOccurrenceID  string `json:"original_occurrence_id"`
	CandidateOccurrenceID string `json:"candidate_occurrence_id"`
}
type ResolutionRequirementPair struct {
	OriginalRequirementID  string                `json:"original_requirement_id"`
	CandidateRequirementID string                `json:"candidate_requirement_id"`
	OriginalOccurrence     RequirementOccurrence `json:"original_occurrence"`
	CandidateOccurrence    RequirementOccurrence `json:"candidate_occurrence"`
}
type ResolutionRole struct {
	OriginalInput  QueryInput                  `json:"original_input"`
	CandidateInput QueryInput                  `json:"candidate_input"`
	Occurrences    []ResolutionOccurrencePair  `json:"occurrences"`
	Requirements   []ResolutionRequirementPair `json:"requirements"`
	Coverage       InputCoverage               `json:"coverage"`
}
type ResolutionReferencePair struct {
	OriginalID  string `json:"original_id"`
	CandidateID string `json:"candidate_id"`
}
type ResolutionLimitation struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Location Location `json:"location"`
}
type ResolutionProofEvidence struct {
	Proven      bool                      `json:"proven"`
	References  []ResolutionReferencePair `json:"references"`
	Roles       []ResolutionRole          `json:"roles"`
	Limitations []ResolutionLimitation    `json:"limitations"`
}
