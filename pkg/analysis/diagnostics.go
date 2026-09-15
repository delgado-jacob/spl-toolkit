package analysis

// Diagnostic codes are stable across all analysis adapters.
const (
	CodeUnavailableField      = "SPL_UNAVAILABLE_FIELD"
	CodeUnsupportedFunction   = "SPL_UNSUPPORTED_FUNCTION"
	CodeDynamicReference      = "SPL_DYNAMIC_REFERENCE"
	CodeUnresolvedWildcard    = "SPL_UNRESOLVED_WILDCARD"
	CodeSyntaxError           = "SPL_SYNTAX_ERROR"
	CodeUnsupportedCommand    = "SPL_UNSUPPORTED_COMMAND"
	CodeUnsupportedSemantics  = "SPL_UNSUPPORTED_SEMANTICS"
	CodeAnalysisResourceLimit = "SPL_ANALYSIS_RESOURCE_LIMIT"

	CodeRequirementIndeterminate      = "SPL_REQUIREMENT_INDETERMINATE"
	CodeRequirementDynamic            = "SPL_REQUIREMENT_DYNAMIC"
	CodeRequirementCoverageIncomplete = "SPL_REQUIREMENT_COVERAGE_INCOMPLETE"
)
