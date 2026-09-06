package mapper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// MappingConfig represents the complete configuration for field mappings
type MappingConfig struct {
	Version     string                 `json:"version"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Mappings    []FieldMapping         `json:"mappings"`
	Rules       []ConditionalRule      `json:"rules,omitempty"`
	DataModels  []DataModelMapping     `json:"datamodels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConditionalRule represents a conditional mapping rule for Phase 2
type ConditionalRule struct {
	ID          string         `json:"id"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Conditions  []Condition    `json:"conditions"`
	Mappings    []FieldMapping `json:"mappings"`
	Priority    int            `json:"priority"`
	Enabled     bool           `json:"enabled"`
}

// Condition represents a condition for conditional mapping
type Condition struct {
	Type     string      `json:"type"` // "field_value", "field_exists", "sourcetype", "source", "combination"
	Field    string      `json:"field,omitempty"`
	Operator string      `json:"operator,omitempty"` // "equals", "contains", "regex", "exists", "not_exists"
	Value    interface{} `json:"value,omitempty"`
	Children []Condition `json:"children,omitempty"` // For combination conditions (AND/OR)
}

// DataModelMapping represents mapping between datamodels for Phase 2
type DataModelMapping struct {
	SourceDataModel     string                  `json:"source_datamodel"`
	TargetDataModel     string                  `json:"target_datamodel"`
	FieldMappings       []DataModelFieldMapping `json:"field_mappings"`
	ConditionalMappings []ConditionalRule       `json:"conditional_mappings,omitempty"`
}

// DataModelFieldMapping represents field mapping within datamodels
type DataModelFieldMapping struct {
	SourceField string `json:"source_field"`
	TargetField string `json:"target_field"`
	SourcePath  string `json:"source_path,omitempty"` // For nested datamodel fields
	TargetPath  string `json:"target_path,omitempty"`
}

// TranslationRule represents query translation rules for Phase 3
type TranslationRule struct {
	ID         string            `json:"id"`
	Name       string            `json:"name,omitempty"`
	SourceType string            `json:"source_type"` // "raw", "datamodel", "tstats"
	TargetType string            `json:"target_type"`
	Mappings   []FieldMapping    `json:"mappings"`
	Conditions []Condition       `json:"conditions,omitempty"`
	Templates  map[string]string `json:"templates,omitempty"` // Query templates
}

// ValidationResult represents the result of schema validation
type ValidationResult struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors,omitempty"`
}

// LoadMappingConfig loads and validates a mapping configuration from JSON
func LoadMappingConfig(jsonData []byte) (*MappingConfig, error) {
	var config MappingConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mapping config: %w", err)
	}

	// Validate the configuration
	if result := config.Validate(); !result.Valid {
		return nil, fmt.Errorf("invalid mapping config: %v", result.Errors)
	}

	return &config, nil
}

// Validate validates the mapping configuration
func (mc *MappingConfig) Validate() ValidationResult {
	var errors []string

	// Check required fields
	if mc.Version == "" {
		errors = append(errors, "version is required")
	}

	// Validate basic mappings
	for i, mapping := range mc.Mappings {
		if strings.TrimSpace(mapping.Source) == "" {
			errors = append(errors, fmt.Sprintf("mapping[%d]: source field is required", i))
		}
		if strings.TrimSpace(mapping.Target) == "" {
			errors = append(errors, fmt.Sprintf("mapping[%d]: target field is required", i))
		}
	}

	// Validate conditional rules
	for i, rule := range mc.Rules {
		if rule.ID == "" {
			errors = append(errors, fmt.Sprintf("rule[%d]: id is required", i))
		}
		if len(rule.Conditions) == 0 {
			errors = append(errors, fmt.Sprintf("rule[%d]: at least one condition is required", i))
		}
		if len(rule.Mappings) == 0 {
			errors = append(errors, fmt.Sprintf("rule[%d]: at least one mapping is required", i))
		}
		for j, mapping := range rule.Mappings {
			if strings.TrimSpace(mapping.Source) == "" {
				errors = append(errors, fmt.Sprintf("rule[%d].mapping[%d]: source field is required", i, j))
			}
			if strings.TrimSpace(mapping.Target) == "" {
				errors = append(errors, fmt.Sprintf("rule[%d].mapping[%d]: target field is required", i, j))
			}
		}

		// Validate conditions
		for j, condition := range rule.Conditions {
			if err := validateCondition(condition); err != nil {
				errors = append(errors, fmt.Sprintf("rule[%d].condition[%d]: %s", i, j, err.Error()))
			}
		}
	}

	if len(mc.DataModels) > 0 {
		errors = append(errors, "datamodel rewrite mappings are not supported")
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func validateCondition(condition Condition) error {
	switch condition.Type {
	case "field_exists":
		if strings.TrimSpace(condition.Field) == "" {
			return fmt.Errorf("field is required for type field_exists")
		}
		if condition.Operator != "exists" && condition.Operator != "not_exists" {
			return fmt.Errorf("field_exists requires 'exists' or 'not_exists' operator")
		}
	case "field_value":
		if strings.TrimSpace(condition.Field) == "" {
			return fmt.Errorf("field is required for type field_value")
		}
		if err := validateValueOperator(condition.Operator, condition.Value); err != nil {
			return err
		}
	case "source", "sourcetype":
		expected, ok := condition.Value.(string)
		if !ok {
			return fmt.Errorf("%s requires a string value", condition.Type)
		}
		if err := validateStringOperator(condition.Operator, expected); err != nil {
			return err
		}
	case "combination":
		if len(condition.Children) < 2 {
			return fmt.Errorf("combination conditions require at least 2 children")
		}
		if condition.Operator != "and" && condition.Operator != "or" {
			return fmt.Errorf("combination conditions require 'and' or 'or' operator")
		}
		// Recursively validate children
		for i, child := range condition.Children {
			if err := validateCondition(child); err != nil {
				return fmt.Errorf("child[%d]: %s", i, err.Error())
			}
		}
	default:
		return fmt.Errorf("invalid condition type: %s", condition.Type)
	}

	return nil
}

func validateValueOperator(operator string, value interface{}) error {
	switch operator {
	case "equals":
		if !isScalar(value) {
			return fmt.Errorf("equals requires a scalar value")
		}
		return nil
	case "contains", "regex":
		expected, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s requires a string value", operator)
		}
		return validateStringOperator(operator, expected)
	default:
		return fmt.Errorf("field_value requires 'equals', 'contains', or 'regex' operator")
	}
}

func validateStringOperator(operator, expected string) error {
	switch operator {
	case "equals", "contains":
		return nil
	case "regex":
		if _, err := regexp.Compile(expected); err != nil {
			return fmt.Errorf("invalid regex: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("requires 'equals', 'contains', or 'regex' operator")
	}
}

func isScalar(value interface{}) bool {
	if value == nil {
		return true
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// ToJSON serializes the mapping configuration to JSON
func (mc *MappingConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(mc, "", "  ")
}

// GetMappingsForConditions returns mappings that match the given conditions
func (mc *MappingConfig) GetMappingsForConditions(conditions map[string]interface{}) []FieldMapping {
	result := append([]FieldMapping(nil), mc.Mappings...)
	result = append(result, mc.matchingRuleMappings(conditions)...)
	return result
}

func (mc *MappingConfig) matchingRuleMappings(context map[string]interface{}) []FieldMapping {
	rules := append([]ConditionalRule(nil), mc.Rules...)
	sort.SliceStable(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	for _, rule := range rules {
		if rule.Enabled && mc.evaluateConditions(rule.Conditions, context) {
			return rule.Mappings
		}
	}

	return nil
}

func (mc *MappingConfig) evaluateConditions(conditions []Condition, context map[string]interface{}) bool {
	for _, condition := range conditions {
		if !mc.evaluateCondition(condition, context) {
			return false // All conditions must be true
		}
	}
	return true
}

func (mc *MappingConfig) evaluateCondition(condition Condition, context map[string]interface{}) bool {
	switch condition.Type {
	case "field_exists":
		_, exists := context[condition.Field]
		return condition.Operator == "exists" && exists || condition.Operator == "not_exists" && !exists

	case "field_value":
		value, exists := context[condition.Field]
		if !exists {
			return false
		}

		switch condition.Operator {
		case "equals":
			return scalarEqual(value, condition.Value)
		case "contains", "regex":
			actual, actualOK := value.(string)
			expected, expectedOK := condition.Value.(string)
			return actualOK && expectedOK && matchString(actual, expected, condition.Operator)
		}

	case "sourcetype", "source":
		value, exists := context[condition.Type]
		if !exists {
			return false
		}

		expected, ok := condition.Value.(string)
		return ok && matchStringValue(value, expected, condition.Operator)

	case "combination":
		if condition.Operator == "and" {
			for _, child := range condition.Children {
				if !mc.evaluateCondition(child, context) {
					return false
				}
			}
			return true
		} else if condition.Operator == "or" {
			for _, child := range condition.Children {
				if mc.evaluateCondition(child, context) {
					return true
				}
			}
			return false
		}
	}

	return false
}

func scalarEqual(actual, expected interface{}) bool {
	return isScalar(actual) && isScalar(expected) && reflect.DeepEqual(actual, expected)
}

func matchStringValue(value interface{}, expected, operator string) bool {
	switch actual := value.(type) {
	case string:
		return matchString(actual, expected, operator)
	case []string:
		for _, item := range actual {
			if matchString(item, expected, operator) {
				return true
			}
		}
	case []interface{}:
		for _, item := range actual {
			if str, ok := item.(string); ok && matchString(str, expected, operator) {
				return true
			}
		}
	}
	return false
}

func matchString(actual, expected, operator string) bool {
	switch operator {
	case "equals":
		return actual == expected
	case "contains":
		return strings.Contains(actual, expected)
	case "regex":
		pattern, err := regexp.Compile(expected)
		return err == nil && pattern.MatchString(actual)
	default:
		return false
	}
}
