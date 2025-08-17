package mapper

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// Mapper represents the main SPL field mapping engine
type Mapper struct {
	fieldMappings map[string]string
	parser        *Parser
	config        *MappingConfig
}

// FieldMapping represents a source to target field mapping
type FieldMapping struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

// MappingRule represents a conditional mapping rule
type MappingRule struct {
	Condition string         `json:"condition"`
	Mappings  []FieldMapping `json:"mappings"`
	Priority  int            `json:"priority"`
}

// QueryInfo contains discovered information about a SPL query
type QueryInfo struct {
	DataModels  []string `json:"datamodels"`
	Datasets    []string `json:"datasets"` // Format: "DataModel.Dataset"
	Lookups     []string `json:"lookups"`
	Macros      []string `json:"macros"`
	Sources     []string `json:"sources"`
	SourceTypes []string `json:"sourcetypes"`
	InputFields []string `json:"input_fields"`
}

// New creates a new Mapper instance
func New() *Mapper {
	return &Mapper{
		fieldMappings: make(map[string]string),
		parser:        NewParser(),
		config:        nil,
	}
}

// NewWithConfig creates a new Mapper instance with a configuration
func NewWithConfig(config *MappingConfig) *Mapper {
	mapper := &Mapper{
		fieldMappings: make(map[string]string),
		parser:        NewParser(),
		config:        config,
	}

	// Load basic mappings from config
	for _, mapping := range config.Mappings {
		mapper.fieldMappings[mapping.Source] = mapping.Target
	}

	return mapper
}

// LoadMappings loads field mappings from JSON
func (m *Mapper) LoadMappings(jsonData []byte) error {
	var mappings []FieldMapping
	if err := json.Unmarshal(jsonData, &mappings); err != nil {
		return err
	}

	for _, mapping := range mappings {
		m.fieldMappings[mapping.Source] = mapping.Target
	}

	return nil
}

// MapQuery applies field mappings to a SPL query
func (m *Mapper) MapQuery(query string) (string, error) {
	// Get query context for conditional mappings
	context := m.extractQueryContextFromString(query)

	// Apply field mappings using token stream rewriting
	return m.mapQueryWithTokenRewriter(query, context)
}

// MapQueryWithContext applies field mappings with explicit context
func (m *Mapper) MapQueryWithContext(query string, context map[string]interface{}) (string, error) {
	// Apply field mappings using token stream rewriting
	return m.mapQueryWithTokenRewriter(query, context)
}

// DiscoverQuery analyzes a SPL query and returns discovered information
func (m *Mapper) DiscoverQuery(query string) (*QueryInfo, error) {
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// Use ANTLR to parse the query with listener pattern
	input := antlr.NewInputStream(query)
	lexer := parser.NewSPLLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	splParser := parser.NewSPLParser(stream)

	// Remove default error listeners to avoid console output
	splParser.RemoveErrorListeners()
	lexer.RemoveErrorListeners()

	// Add custom error listener to capture errors
	errorListener := &CustomErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		errors:               []string{},
	}
	splParser.AddErrorListener(errorListener)
	lexer.AddErrorListener(errorListener)

	// Extract macros using pattern matching first (since grammar doesn't support backticks yet)
	// We do this before parsing because macros may cause parse errors
	listener := NewFieldDiscoveryListener()
	macros := m.extractMacros(query)
	for _, macro := range macros {
		listener.addMacro(macro)
	}

	// Parse the query starting from the top-level rule
	tree := splParser.Query()

	// Check for parse errors - but still try to extract what we can
	if len(errorListener.errors) > 0 {
		// If parsing failed but we found macros, still return partial info
		// This handles cases where macros cause parsing to fail due to unsupported grammar
		if len(listener.Macros) > 0 {
			info := &QueryInfo{
				DataModels:  []string{},
				Datasets:    []string{},
				Lookups:     []string{},
				Macros:      listener.Macros,
				Sources:     []string{},
				SourceTypes: []string{},
				InputFields: []string{},
			}
			return info, nil
		}
		return nil, fmt.Errorf("parse errors: %s", strings.Join(errorListener.errors, "; "))
	}

	// Create and walk with the discovery listener
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Convert listener results to QueryInfo
	info := &QueryInfo{
		DataModels:  listener.DataModels,
		Datasets:    listener.Datasets,
		Lookups:     listener.Lookups,
		Macros:      listener.Macros,
		Sources:     listener.Sources,
		SourceTypes: listener.SourceTypes,
		InputFields: listener.InputFields,
	}

	return info, nil
}

// GetInputFields returns all input fields required for a query
func (m *Mapper) GetInputFields(query string) ([]string, error) {
	info, err := m.DiscoverQuery(query)
	if err != nil {
		return nil, err
	}

	return info.InputFields, nil
}

// ValidateQuery validates the syntax of a SPL query
func (m *Mapper) ValidateQuery(query string) error {
	return m.parser.ValidateQuery(query)
}

func (m *Mapper) getEffectiveMappings(context map[string]interface{}) map[string]string {
	// Start with basic mappings
	result := make(map[string]string)
	for k, v := range m.fieldMappings {
		result[k] = v
	}

	// Add conditional mappings if config is available
	if m.config != nil && context != nil {
		conditionalMappings := m.config.GetMappingsForConditions(context)
		for _, mapping := range conditionalMappings {
			result[mapping.Source] = mapping.Target
		}
	}

	return result
}

// mapQueryWithTokenRewriter uses ANTLR token stream rewriting for proper field mapping
func (m *Mapper) mapQueryWithTokenRewriter(query string, context map[string]interface{}) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty query")
	}

	// Create input stream
	input := antlr.NewInputStream(query)
	lexer := parser.NewSPLLexer(input)

	// Create token stream
	stream := antlr.NewCommonTokenStream(lexer, 0)

	// Create parser
	splParser := parser.NewSPLParser(stream)

	// Remove default error listeners to avoid console output
	splParser.RemoveErrorListeners()
	lexer.RemoveErrorListeners()

	// Add custom error listener to capture errors
	errorListener := &CustomErrorListener{
		DefaultErrorListener: antlr.NewDefaultErrorListener(),
		errors:               []string{},
	}
	splParser.AddErrorListener(errorListener)
	lexer.AddErrorListener(errorListener)

	// Parse the query
	tree := splParser.Query()

	// Check for parse errors
	if len(errorListener.errors) > 0 {
		return "", fmt.Errorf("parse errors: %s", strings.Join(errorListener.errors, "; "))
	}

	// Create and configure the mapping listener
	effectiveMappings := m.getEffectiveMappings(context)
	listener := NewFieldMappingListener(stream, effectiveMappings)

	// Walk the tree to apply mappings
	antlr.ParseTreeWalkerDefault.Walk(listener, tree)

	// Get the rewritten query
	return listener.GetRewrittenText(), nil
}

// extractQueryContextFromString extracts context from query string for conditional mappings
func (m *Mapper) extractQueryContextFromString(query string) map[string]interface{} {
	context := make(map[string]interface{})

	// Use the ANTLR listener to extract context information
	info, err := m.DiscoverQuery(query)
	if err != nil {
		return context // Return empty context on error
	}

	// Convert discovered information to context map
	// Store arrays to support multiple values and any-match evaluation
	if len(info.SourceTypes) > 0 {
		if len(info.SourceTypes) == 1 {
			context["sourcetype"] = info.SourceTypes[0]
		} else {
			context["sourcetype"] = info.SourceTypes
		}
	}
	if len(info.Sources) > 0 {
		if len(info.Sources) == 1 {
			context["source"] = info.Sources[0]
		} else {
			context["source"] = info.Sources
		}
	}
	if len(info.DataModels) > 0 {
		if len(info.DataModels) == 1 {
			context["datamodel"] = info.DataModels[0]
		} else {
			context["datamodel"] = info.DataModels
		}
	}

	return context
}

// isValidInputField determines if a field name is likely a valid input field
func (m *Mapper) isValidInputField(fieldName string) bool {
	if fieldName == "" {
		return false
	}

	// Skip obvious non-field values
	if strings.Contains(fieldName, "=") ||
		strings.Contains(fieldName, "|") ||
		strings.Contains(fieldName, "(") ||
		strings.Contains(fieldName, ")") ||
		strings.Contains(fieldName, "+") ||
		strings.Contains(fieldName, "-") ||
		strings.Contains(fieldName, "*") ||
		strings.Contains(fieldName, "/") {
		return false
	}

	// Skip quoted strings
	if (strings.HasPrefix(fieldName, "\"") && strings.HasSuffix(fieldName, "\"")) ||
		(strings.HasPrefix(fieldName, "'") && strings.HasSuffix(fieldName, "'")) {
		return false
	}

	// Skip numbers
	if strings.ContainsAny(fieldName, "0123456789") && !strings.ContainsAny(fieldName, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_") {
		return false
	}

	// Skip SPL commands and functions
	commands := []string{"search", "stats", "eval", "where", "by", "as", "count", "sum", "avg", "max", "min",
		"lookup", "inputlookup", "rename", "sort", "head", "tail", "dedup", "table", "fields", "tstats",
		"datamodel", "from", "case", "if", "OUTPUT", "output"}
	for _, cmd := range commands {
		if strings.EqualFold(fieldName, cmd) {
			return false
		}
	}

	// Must contain at least one letter and start with letter or underscore
	if !strings.ContainsAny(fieldName, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return false
	}

	firstChar := fieldName[0]
	if !((firstChar >= 'a' && firstChar <= 'z') ||
		(firstChar >= 'A' && firstChar <= 'Z') ||
		firstChar == '_') {
		return false
	}

	return true
}

// extractMacros extracts macro invocations from the query using regex pattern matching
// This is a temporary solution until the ANTLR grammar supports backtick syntax
func (m *Mapper) extractMacros(query string) []string {
	var macros []string

	// Pattern for macro invocations: `macro_name(args)` or `macro_name`
	// This regex matches backtick-enclosed content that looks like a macro
	macroRegex := regexp.MustCompile("`([a-zA-Z_][a-zA-Z0-9_]*)(\\([^)]*\\))?`")

	matches := macroRegex.FindAllStringSubmatch(query, -1)

	for _, match := range matches {
		if len(match) > 1 {
			macroName := match[1]
			// Only add unique macros
			found := false
			for _, existing := range macros {
				if existing == macroName {
					found = true
					break
				}
			}
			if !found {
				macros = append(macros, macroName)
			}
		}
	}

	return macros
}
