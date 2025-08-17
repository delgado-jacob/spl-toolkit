package mapper

import (
	"encoding/json"
	"fmt"
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
		if mapping.Source == "" {
			errors = append(errors, fmt.Sprintf("mapping[%d]: source field is required", i))
		}
		if mapping.Target == "" {
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

		// Validate conditions
		for j, condition := range rule.Conditions {
			if err := validateCondition(condition); err != nil {
				errors = append(errors, fmt.Sprintf("rule[%d].condition[%d]: %s", i, j, err.Error()))
			}
		}
	}

	// Validate datamodel mappings
	for i, dm := range mc.DataModels {
		if dm.SourceDataModel == "" {
			errors = append(errors, fmt.Sprintf("datamodel[%d]: source_datamodel is required", i))
		}
		if dm.TargetDataModel == "" {
			errors = append(errors, fmt.Sprintf("datamodel[%d]: target_datamodel is required", i))
		}
	}

	return ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

func validateCondition(condition Condition) error {
	validTypes := []string{"field_value", "field_exists", "sourcetype", "source", "combination"}
	validOperators := []string{"equals", "contains", "regex", "exists", "not_exists", "and", "or"}

	// Check type
	isValidType := false
	for _, vt := range validTypes {
		if condition.Type == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("invalid condition type: %s", condition.Type)
	}

	// Check operator if present
	if condition.Operator != "" {
		isValidOperator := false
		for _, vo := range validOperators {
			if condition.Operator == vo {
				isValidOperator = true
				break
			}
		}
		if !isValidOperator {
			return fmt.Errorf("invalid operator: %s", condition.Operator)
		}
	}

	// Type-specific validation
	switch condition.Type {
	case "field_value", "field_exists":
		if condition.Field == "" {
			return fmt.Errorf("field is required for type %s", condition.Type)
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
	}

	return nil
}

// ToJSON serializes the mapping configuration to JSON
func (mc *MappingConfig) ToJSON() ([]byte, error) {
	return json.MarshalIndent(mc, "", "  ")
}

// GetMappingsForConditions returns mappings that match the given conditions
func (mc *MappingConfig) GetMappingsForConditions(conditions map[string]interface{}) []FieldMapping {
	var result []FieldMapping

	// Add basic mappings
	result = append(result, mc.Mappings...)

	// Check conditional rules
	for _, rule := range mc.Rules {
		if !rule.Enabled {
			continue
		}

		if mc.evaluateConditions(rule.Conditions, conditions) {
			result = append(result, rule.Mappings...)
		}
	}

	return result
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
			return value == condition.Value
		case "contains":
			if str, ok := value.(string); ok {
				if condStr, ok := condition.Value.(string); ok {
					return strings.Contains(str, condStr)
				}
			}
		}

	case "sourcetype", "source":
		value, exists := context[condition.Type]
		if !exists {
			return false
		}

		// Handle both single values and arrays (any-match semantics)
		switch condition.Operator {
		case "equals":
			if arr, ok := value.([]string); ok {
				// Array case - check if any element matches
				for _, str := range arr {
					if str == condition.Value {
						return true
					}
				}
				return false
			}
			// Single value case
			return value == condition.Value
		case "contains":
			if condStr, ok := condition.Value.(string); ok {
				if arr, ok := value.([]string); ok {
					// Array case - check if any element contains the condition
					for _, str := range arr {
						if strings.Contains(str, condStr) {
							return true
						}
					}
					return false
				}
				// Single value case
				if str, ok := value.(string); ok {
					return strings.Contains(str, condStr)
				}
			}
		}

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
