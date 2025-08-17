package mapper

import (
	"encoding/json"
	"testing"
)

func TestLoadMappingConfig(t *testing.T) {
	// Valid configuration
	validConfig := `{
		"version": "1.0",
		"name": "Test Config",
		"mappings": [
			{"source": "src_ip", "target": "source_ip"},
			{"source": "dst_ip", "target": "destination_ip"}
		],
		"rules": [
			{
				"id": "rule1",
				"conditions": [
					{
						"type": "sourcetype",
						"operator": "equals",
						"value": "access_combined"
					}
				],
				"mappings": [
					{"source": "clientip", "target": "src_ip"}
				],
				"priority": 1,
				"enabled": true
			}
		]
	}`

	config, err := LoadMappingConfig([]byte(validConfig))
	if err != nil {
		t.Fatalf("Failed to load valid config: %v", err)
	}

	if config.Version != "1.0" {
		t.Errorf("Expected version '1.0', got %s", config.Version)
	}

	if len(config.Mappings) != 2 {
		t.Errorf("Expected 2 mappings, got %d", len(config.Mappings))
	}

	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}
}

func TestMappingConfigValidate(t *testing.T) {
	// Test missing version
	config := MappingConfig{
		Mappings: []FieldMapping{
			{Source: "test", Target: "test2"},
		},
	}

	result := config.Validate()
	if result.Valid {
		t.Error("Expected validation to fail for missing version")
	}

	// Test missing source field
	config = MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Target: "test2"},
		},
	}

	result = config.Validate()
	if result.Valid {
		t.Error("Expected validation to fail for missing source field")
	}

	// Test valid configuration
	config = MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Source: "test", Target: "test2"},
		},
	}

	result = config.Validate()
	if !result.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
	}
}

func TestConditionalRuleValidation(t *testing.T) {
	config := MappingConfig{
		Version: "1.0",
		Rules: []ConditionalRule{
			{
				// Missing ID
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "test"},
				},
				Mappings: []FieldMapping{
					{Source: "test", Target: "test2"},
				},
			},
		},
	}

	result := config.Validate()
	if result.Valid {
		t.Error("Expected validation to fail for missing rule ID")
	}
}

func TestConditionValidation(t *testing.T) {
	// Test invalid condition type
	condition := Condition{
		Type:     "invalid_type",
		Operator: "equals",
		Value:    "test",
	}

	err := validateCondition(condition)
	if err == nil {
		t.Error("Expected validation to fail for invalid condition type")
	}

	// Test invalid operator
	condition = Condition{
		Type:     "field_value",
		Field:    "test_field",
		Operator: "invalid_operator",
		Value:    "test",
	}

	err = validateCondition(condition)
	if err == nil {
		t.Error("Expected validation to fail for invalid operator")
	}

	// Test missing field for field_value condition
	condition = Condition{
		Type:     "field_value",
		Operator: "equals",
		Value:    "test",
	}

	err = validateCondition(condition)
	if err == nil {
		t.Error("Expected validation to fail for missing field")
	}

	// Test valid condition
	condition = Condition{
		Type:     "field_value",
		Field:    "test_field",
		Operator: "equals",
		Value:    "test",
	}

	err = validateCondition(condition)
	if err != nil {
		t.Errorf("Expected valid condition to pass, got error: %v", err)
	}
}

func TestCombinationCondition(t *testing.T) {
	// Test combination with insufficient children
	condition := Condition{
		Type:     "combination",
		Operator: "and",
		Children: []Condition{
			{Type: "field_value", Field: "test", Operator: "equals", Value: "test"},
		},
	}

	err := validateCondition(condition)
	if err == nil {
		t.Error("Expected validation to fail for combination with insufficient children")
	}

	// Test valid combination
	condition = Condition{
		Type:     "combination",
		Operator: "and",
		Children: []Condition{
			{Type: "field_value", Field: "test1", Operator: "equals", Value: "value1"},
			{Type: "field_value", Field: "test2", Operator: "equals", Value: "value2"},
		},
	}

	err = validateCondition(condition)
	if err != nil {
		t.Errorf("Expected valid combination to pass, got error: %v", err)
	}
}

func TestGetMappingsForConditions(t *testing.T) {
	config := MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Source: "base_field", Target: "mapped_field"},
		},
		Rules: []ConditionalRule{
			{
				ID: "rule1",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "access_combined"},
				},
				Mappings: []FieldMapping{
					{Source: "conditional_field", Target: "conditional_mapped"},
				},
				Enabled: true,
			},
			{
				ID: "rule2",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "other_sourcetype"},
				},
				Mappings: []FieldMapping{
					{Source: "other_field", Target: "other_mapped"},
				},
				Enabled: true,
			},
			{
				ID: "rule3_disabled",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "access_combined"},
				},
				Mappings: []FieldMapping{
					{Source: "disabled_field", Target: "disabled_mapped"},
				},
				Enabled: false,
			},
		},
	}

	// Test matching conditions
	conditions := map[string]interface{}{
		"sourcetype": "access_combined",
	}

	mappings := config.GetMappingsForConditions(conditions)

	// Should include base mapping + rule1 (rule3 is disabled)
	expectedCount := 2
	if len(mappings) != expectedCount {
		t.Errorf("Expected %d mappings, got %d", expectedCount, len(mappings))
	}

	// Check that correct mappings are included
	found := map[string]bool{
		"base_field":        false,
		"conditional_field": false,
	}

	for _, mapping := range mappings {
		if _, exists := found[mapping.Source]; exists {
			found[mapping.Source] = true
		}
	}

	for source, wasFound := range found {
		if !wasFound {
			t.Errorf("Expected mapping for source '%s' was not found", source)
		}
	}
}

func TestConditionEvaluation(t *testing.T) {
	config := MappingConfig{}

	// Test field_exists condition - exists
	condition := Condition{
		Type:     "field_exists",
		Field:    "test_field",
		Operator: "exists",
	}

	context := map[string]interface{}{
		"test_field": "some_value",
	}

	result := config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected field_exists condition to evaluate to true")
	}

	// Test field_exists condition - not_exists
	condition = Condition{
		Type:     "field_exists",
		Field:    "missing_field",
		Operator: "not_exists",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected field_exists not_exists condition to evaluate to true")
	}

	// Test field_value condition - equals
	condition = Condition{
		Type:     "field_value",
		Field:    "test_field",
		Operator: "equals",
		Value:    "expected_value",
	}

	context = map[string]interface{}{
		"test_field": "expected_value",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected field_value equals condition to evaluate to true")
	}

	// Test field_value condition - contains (this tests the bug fix)
	condition = Condition{
		Type:     "field_value",
		Field:    "test_field",
		Operator: "contains",
		Value:    "partial",
	}

	context = map[string]interface{}{
		"test_field": "this contains partial text",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected field_value contains condition to evaluate to true")
	}

	// Test field_value condition - contains negative case
	condition = Condition{
		Type:     "field_value",
		Field:    "test_field",
		Operator: "contains",
		Value:    "missing",
	}

	context = map[string]interface{}{
		"test_field": "this does not have the substring",
	}

	result = config.evaluateCondition(condition, context)
	if result {
		t.Error("Expected field_value contains condition to evaluate to false")
	}

	// Test sourcetype condition - equals
	condition = Condition{
		Type:     "sourcetype",
		Operator: "equals",
		Value:    "access_combined",
	}

	context = map[string]interface{}{
		"sourcetype": "access_combined",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected sourcetype equals condition to evaluate to true")
	}

	// Test sourcetype condition - contains (this tests the bug fix)
	condition = Condition{
		Type:     "sourcetype",
		Operator: "contains",
		Value:    "access",
	}

	context = map[string]interface{}{
		"sourcetype": "access_combined_extended",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected sourcetype contains condition to evaluate to true")
	}

	// Test source condition - contains
	condition = Condition{
		Type:     "source",
		Operator: "contains",
		Value:    "app",
	}

	context = map[string]interface{}{
		"source": "/var/log/app/server.log",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected source contains condition to evaluate to true")
	}

	// Test combination condition (AND)
	condition = Condition{
		Type:     "combination",
		Operator: "and",
		Children: []Condition{
			{Type: "field_value", Field: "field1", Operator: "equals", Value: "value1"},
			{Type: "field_value", Field: "field2", Operator: "equals", Value: "value2"},
		},
	}

	context = map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected AND combination condition to evaluate to true")
	}

	// Test combination condition (OR) with one false
	condition = Condition{
		Type:     "combination",
		Operator: "or",
		Children: []Condition{
			{Type: "field_value", Field: "field1", Operator: "equals", Value: "value1"},
			{Type: "field_value", Field: "field2", Operator: "equals", Value: "wrong_value"},
		},
	}

	context = map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected OR combination condition to evaluate to true")
	}
}

func TestMultipleContextValueHandling(t *testing.T) {
	config := MappingConfig{}

	// Test sourcetype condition with array values (any-match semantics)
	condition := Condition{
		Type:     "sourcetype",
		Operator: "equals",
		Value:    "access_combined",
	}

	// Test with array containing matching value
	context := map[string]interface{}{
		"sourcetype": []string{"firewall", "access_combined", "syslog"},
	}

	result := config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected sourcetype array condition (equals) to evaluate to true when any element matches")
	}

	// Test with array not containing matching value
	context = map[string]interface{}{
		"sourcetype": []string{"firewall", "nginx_access", "syslog"},
	}

	result = config.evaluateCondition(condition, context)
	if result {
		t.Error("Expected sourcetype array condition (equals) to evaluate to false when no element matches")
	}

	// Test sourcetype contains with array
	condition = Condition{
		Type:     "sourcetype",
		Operator: "contains",
		Value:    "access",
	}

	context = map[string]interface{}{
		"sourcetype": []string{"firewall", "nginx_access_log", "syslog"},
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected sourcetype array condition (contains) to evaluate to true when any element contains substring")
	}

	// Test source condition with array values
	condition = Condition{
		Type:     "source",
		Operator: "contains",
		Value:    "app",
	}

	context = map[string]interface{}{
		"source": []string{"/var/log/system.log", "/var/log/app/server.log", "/var/log/auth.log"},
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected source array condition (contains) to evaluate to true when any element contains substring")
	}

	// Test backward compatibility with single values
	condition = Condition{
		Type:     "sourcetype",
		Operator: "equals",
		Value:    "access_combined",
	}

	context = map[string]interface{}{
		"sourcetype": "access_combined",
	}

	result = config.evaluateCondition(condition, context)
	if !result {
		t.Error("Expected single value sourcetype condition to still work (backward compatibility)")
	}
}

func TestToJSON(t *testing.T) {
	config := MappingConfig{
		Version: "1.0",
		Name:    "Test Config",
		Mappings: []FieldMapping{
			{Source: "test", Target: "test_mapped"},
		},
	}

	jsonData, err := config.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	// Test that we can deserialize it back
	var restored MappingConfig
	err = json.Unmarshal(jsonData, &restored)
	if err != nil {
		t.Fatalf("Failed to deserialize JSON: %v", err)
	}

	if restored.Version != config.Version {
		t.Errorf("Version mismatch after JSON round-trip: expected %s, got %s", config.Version, restored.Version)
	}

	if restored.Name != config.Name {
		t.Errorf("Name mismatch after JSON round-trip: expected %s, got %s", config.Name, restored.Name)
	}
}
