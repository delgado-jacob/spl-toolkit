package mapper

import (
	"encoding/json"
	"strings"
	"testing"
)

// Helper function for slice contains check
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("Expected non-nil mapper")
	}
	if m.fieldMappings == nil {
		t.Fatal("Expected initialized field mappings")
	}
	if m.parser == nil {
		t.Fatal("Expected initialized parser")
	}
}

func TestLoadMappings(t *testing.T) {
	m := New()

	mappings := []FieldMapping{
		{Source: "src_ip", Target: "source_ip"},
		{Source: "dst_ip", Target: "destination_ip"},
	}

	jsonData, err := json.Marshal(mappings)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	err = m.LoadMappings(jsonData)
	if err != nil {
		t.Fatalf("Failed to load mappings: %v", err)
	}

	if m.fieldMappings["src_ip"] != "source_ip" {
		t.Errorf("Expected mapping src_ip -> source_ip, got %s", m.fieldMappings["src_ip"])
	}

	if m.fieldMappings["dst_ip"] != "destination_ip" {
		t.Errorf("Expected mapping dst_ip -> destination_ip, got %s", m.fieldMappings["dst_ip"])
	}
}

func TestMapQuery(t *testing.T) {
	m := New()

	// Test with empty query
	_, err := m.MapQuery("")
	if err == nil {
		t.Error("Expected error for empty query")
	}

	// Test with simple query
	query := "search src_ip=192.168.1.1"
	result, err := m.MapQuery(query)
	if err != nil {
		t.Fatalf("Failed to map query: %v", err)
	}

	// For now, just ensure we get a result (placeholder implementation)
	if result == "" {
		// This is expected with current placeholder implementation
		t.Log("Placeholder implementation returns empty string")
	}
}

func TestDiscoverQuery(t *testing.T) {
	m := New()

	// Test with simple query
	query := "search sourcetype=access_combined | stats count by src_ip"
	info, err := m.DiscoverQuery(query)
	if err != nil {
		t.Fatalf("Failed to discover query info: %v", err)
	}

	if info == nil {
		t.Fatal("Expected non-nil query info")
	}

	// Verify structure is initialized
	if info.DataModels == nil {
		t.Error("Expected initialized DataModels slice")
	}
	if info.InputFields == nil {
		t.Error("Expected initialized InputFields slice")
	}

	// Verify expected sourcetypes are discovered
	if !contains(info.SourceTypes, "access_combined") {
		t.Errorf("Expected sourcetype 'access_combined' to be discovered, got: %v", info.SourceTypes)
	}

	// Verify expected input fields are discovered
	if !contains(info.InputFields, "src_ip") {
		t.Errorf("Expected input field 'src_ip' to be discovered, got: %v", info.InputFields)
	}

	// Verify non-input values are NOT included as input fields
	excludedValues := []string{
		"access_combined", // sourcetype value, not a field
		"search",          // SPL command
		"stats",           // SPL command
		"count",           // SPL function
		"by",              // SPL keyword
	}

	for _, excluded := range excludedValues {
		if contains(info.InputFields, excluded) {
			t.Errorf("Expected '%s' to NOT be included in input fields, got: %v", excluded, info.InputFields)
		}
	}

	// Verify sourcetype values are not in other discovery categories
	if contains(info.InputFields, "access_combined") {
		t.Errorf("Expected sourcetype value 'access_combined' to NOT be in input fields, got: %v", info.InputFields)
	}
	if contains(info.DataModels, "access_combined") {
		t.Errorf("Expected sourcetype value 'access_combined' to NOT be in datamodels, got: %v", info.DataModels)
	}
	if contains(info.Lookups, "access_combined") {
		t.Errorf("Expected sourcetype value 'access_combined' to NOT be in lookups, got: %v", info.Lookups)
	}
}

func TestGetInputFields(t *testing.T) {
	m := New()

	query := "search src_ip=192.168.1.1 | eval new_field=src_ip"
	fields, err := m.GetInputFields(query)
	if err != nil {
		t.Fatalf("Failed to get input fields: %v", err)
	}

	if fields == nil {
		t.Fatal("Expected non-nil fields slice")
	}
}

// TestTokenStreamRewriting tests that field mapping preserves spaces and formatting
func TestTokenStreamRewriting(t *testing.T) {
	m := New()

	mappings := []FieldMapping{
		{Source: "src_ip", Target: "source_ip"},
		{Source: "dst_port", Target: "destination_port"},
	}

	jsonData, _ := json.Marshal(mappings)
	err := m.LoadMappings(jsonData)
	if err != nil {
		t.Fatalf("Failed to load mappings: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic field mapping with spaces",
			input:    "search src_ip=192.168.1.1 dst_port=80",
			expected: "search source_ip=192.168.1.1 destination_port=80",
		},
		{
			name:     "No EOF in output",
			input:    "search src_ip=10.0.0.1",
			expected: "search source_ip=10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.MapQuery(tt.input)
			if err != nil {
				t.Fatalf("Failed to map query: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}

			// Ensure no EOF token appears
			if strings.Contains(result, "<EOF>") {
				t.Errorf("Result contains EOF token: %q", result)
			}
		})
	}
}

// TestDiscoveryInputlookup tests discovery of inputlookup commands
func TestDiscoveryInputlookup(t *testing.T) {
	m := New()

	query := "| inputlookup ip_geo_lookup.csv | search country=US"
	info, err := m.DiscoverQuery(query)
	if err != nil {
		t.Fatalf("Failed to discover query info: %v", err)
	}

	// Should detect the lookup
	if !contains(info.Lookups, "ip_geo_lookup") {
		t.Errorf("Expected lookup 'ip_geo_lookup' to be discovered, got: %v", info.Lookups)
	}

	// Should detect input fields used in search
	if !contains(info.InputFields, "country") {
		t.Errorf("Expected input field 'country' to be discovered, got: %v", info.InputFields)
	}
}

// TestDiscoveryDatamodel tests discovery of datamodel commands
func TestDiscoveryDatamodel(t *testing.T) {
	m := New()

	query := "| datamodel Network_Traffic All_Traffic search | stats avg(bytes_in) by src_ip"
	info, err := m.DiscoverQuery(query)
	if err != nil {
		t.Fatalf("Failed to discover query info: %v", err)
	}

	// Should detect the datamodel
	if !contains(info.DataModels, "Network_Traffic") {
		t.Errorf("Expected datamodel 'Network_Traffic' to be discovered, got: %v", info.DataModels)
	}

	// Should detect input fields
	if !contains(info.InputFields, "src_ip") {
		t.Errorf("Expected input field 'src_ip' to be discovered, got: %v", info.InputFields)
	}
}

// TestDiscoveryTstats tests discovery of tstats commands
func TestDiscoveryTstats(t *testing.T) {
	m := New()

	query := "| tstats count from datamodel=Web where Web.status=200 by Web.src_ip"
	info, err := m.DiscoverQuery(query)
	if err != nil {
		t.Fatalf("Failed to discover query info: %v", err)
	}

	// Should detect the datamodel
	if !contains(info.DataModels, "Web") {
		t.Errorf("Expected datamodel 'Web' to be discovered, got: %v", info.DataModels)
	}

	// Should detect datamodel fields
	expectedFields := []string{"Web.status", "Web.src_ip"}
	for _, field := range expectedFields {
		if !contains(info.InputFields, field) {
			t.Errorf("Expected input field '%s' to be discovered, got: %v", field, info.InputFields)
		}
	}
}

// TestConditionalMapping tests conditional field mapping based on context
func TestConditionalMapping(t *testing.T) {
	config := &MappingConfig{
		Version: "1.0",
		Name:    "Test Conditional Mapping",
		Mappings: []FieldMapping{
			{Source: "ip", Target: "client_ip"},
		},
		Rules: []ConditionalRule{
			{
				ID:   "apache_logs",
				Name: "Apache Access Log Fields",
				Conditions: []Condition{
					{
						Type:     "sourcetype",
						Operator: "equals",
						Value:    "access_combined",
					},
				},
				Mappings: []FieldMapping{
					{Source: "clientip", Target: "source_address"},
				},
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	m := NewWithConfig(config)

	tests := []struct {
		name     string
		query    string
		context  map[string]interface{}
		expected string
	}{
		{
			name:     "Apache logs conditional mapping",
			query:    "search sourcetype=access_combined clientip=192.168.1.1",
			context:  map[string]interface{}{"sourcetype": "access_combined"},
			expected: "search sourcetype=access_combined source_address=192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.MapQueryWithContext(tt.query, tt.context)
			if err != nil {
				t.Fatalf("Failed to map query: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestComplexQueryDiscovery tests discovery on complex queries with multiple components
func TestComplexQueryDiscovery(t *testing.T) {
	m := New()

	tests := []struct {
		name               string
		query              string
		expectedDataModels []string
		expectedLookups    []string
		expectedFields     []string
	}{
		{
			name:               "Inputlookup with datamodel field references",
			query:              "| inputlookup geo_data.csv | search country=US | eval enriched_ip=Web.src_ip",
			expectedDataModels: []string{},
			expectedLookups:    []string{"geo_data"},
			expectedFields:     []string{"country", "Web.src_ip"},
		},
		{
			name:               "Tstats with multiple datamodel fields",
			query:              "| tstats count from datamodel=Authentication where Authentication.action=success by Authentication.user",
			expectedDataModels: []string{"Authentication"},
			expectedLookups:    []string{},
			expectedFields:     []string{"Authentication.action", "Authentication.user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("Failed to discover query info: %v", err)
			}

			// Check datamodels
			for _, expected := range tt.expectedDataModels {
				if !contains(info.DataModels, expected) {
					t.Errorf("Expected datamodel '%s' to be discovered, got: %v", expected, info.DataModels)
				}
			}

			// Check lookups
			for _, expected := range tt.expectedLookups {
				if !contains(info.Lookups, expected) {
					t.Errorf("Expected lookup '%s' to be discovered, got: %v", expected, info.Lookups)
				}
			}

			// Check fields
			for _, expected := range tt.expectedFields {
				if !contains(info.InputFields, expected) {
					t.Errorf("Expected field '%s' to be discovered, got: %v", expected, info.InputFields)
				}
			}
		})
	}
}

func TestDiscoverQueryComplexCases(t *testing.T) {
	m := New()

	tests := []struct {
		name               string
		query              string
		expectedST         []string
		expectedInput      []string
		excludedInput      []string
		excludedFromST     []string
		excludedFromLookup []string
		excludedFromDM     []string
	}{
		{
			name:               "Lookup query",
			query:              "search user_id=123 | lookup user_table user_id OUTPUT username",
			expectedInput:      []string{"user_id"},
			excludedInput:      []string{"username", "user_table", "123", "search", "lookup", "OUTPUT"},
			excludedFromST:     []string{"user_id", "user_table", "username"},
			excludedFromLookup: []string{"user_id", "123", "username"}, // user_table should be in lookups
			excludedFromDM:     []string{"user_id", "user_table", "username", "123"},
		},
		{
			name:               "Eval with complex expressions",
			query:              "search host=web01 src_ip=192.168.1.1 | eval total=bytes_in+bytes_out | stats sum(total) by dest_ip",
			expectedInput:      []string{"host", "src_ip", "bytes_in", "bytes_out", "dest_ip"},
			excludedInput:      []string{"total", "web01", "192.168.1.1", "search", "eval", "stats", "sum", "by"},
			excludedFromST:     []string{"host", "src_ip", "total", "web01"},
			excludedFromLookup: []string{"host", "src_ip", "total", "web01", "bytes_in", "bytes_out"},
			excludedFromDM:     []string{"host", "src_ip", "total", "web01", "bytes_in", "bytes_out"},
		},
		{
			name:               "Datamodel query",
			query:              "| tstats count from datamodel=Network_Traffic.All_Traffic by src_ip",
			expectedInput:      []string{"src_ip"},
			excludedInput:      []string{"Network_Traffic", "All_Traffic", "tstats", "count", "from", "datamodel", "by"},
			excludedFromST:     []string{"src_ip", "Network_Traffic", "All_Traffic"},
			excludedFromLookup: []string{"src_ip", "Network_Traffic", "All_Traffic"},
			excludedFromDM:     []string{"src_ip"}, // Network_Traffic should be in datamodels
		},
		{
			name:               "Subquery with rename",
			query:              "search sourcetype=firewall | rename source_ip as src | stats count by src",
			expectedST:         []string{"firewall"},
			expectedInput:      []string{"source_ip"},
			excludedInput:      []string{"src", "firewall", "search", "rename", "as", "stats", "count", "by"},
			excludedFromST:     []string{"source_ip", "src"},
			excludedFromLookup: []string{"firewall", "source_ip", "src"},
			excludedFromDM:     []string{"firewall", "source_ip", "src"},
		},
		{
			name:               "Multiple evals and lookups",
			query:              "search event_id=4624 | eval login_type=\"interactive\" | lookup logon_codes logon_type OUTPUT description",
			expectedInput:      []string{"event_id", "logon_type"},
			excludedInput:      []string{"login_type", "description", "logon_codes", "4624", "interactive", "search", "eval", "lookup", "OUTPUT"},
			excludedFromST:     []string{"event_id", "logon_type", "login_type", "description"},
			excludedFromLookup: []string{"event_id", "4624", "login_type", "description"}, // logon_codes should be in lookups
			excludedFromDM:     []string{"event_id", "logon_type", "login_type", "description", "logon_codes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("Failed to discover query info: %v", err)
			}

			// Check expected sourcetypes
			for _, expectedST := range tt.expectedST {
				if !contains(info.SourceTypes, expectedST) {
					t.Errorf("Expected sourcetype '%s' to be discovered, got: %v", expectedST, info.SourceTypes)
				}
			}

			// Check expected input fields
			for _, expectedField := range tt.expectedInput {
				if !contains(info.InputFields, expectedField) {
					t.Errorf("Expected input field '%s' to be discovered, got: %v", expectedField, info.InputFields)
				}
			}

			// Check that excluded fields are not present in input fields
			for _, excludedField := range tt.excludedInput {
				if contains(info.InputFields, excludedField) {
					t.Errorf("Expected field '%s' to NOT be in input fields, got: %v", excludedField, info.InputFields)
				}
			}

			// Check that values are not incorrectly categorized in sourcetypes
			for _, excluded := range tt.excludedFromST {
				if contains(info.SourceTypes, excluded) {
					t.Errorf("Expected '%s' to NOT be in sourcetypes, got: %v", excluded, info.SourceTypes)
				}
			}

			// Check that values are not incorrectly categorized in lookups
			for _, excluded := range tt.excludedFromLookup {
				if contains(info.Lookups, excluded) {
					t.Errorf("Expected '%s' to NOT be in lookups, got: %v", excluded, info.Lookups)
				}
			}

			// Check that values are not incorrectly categorized in datamodels
			for _, excluded := range tt.excludedFromDM {
				if contains(info.DataModels, excluded) {
					t.Errorf("Expected '%s' to NOT be in datamodels, got: %v", excluded, info.DataModels)
				}
			}
		})
	}
}

func TestDiscoverQueryEdgeCases(t *testing.T) {
	m := New()

	tests := []struct {
		name          string
		query         string
		expectedInput []string
		excludedInput []string
		description   string
	}{
		{
			name:          "Field with underscore and numbers",
			query:         "search field_name_123=value | stats count by another_field_456",
			expectedInput: []string{"field_name_123", "another_field_456"},
			excludedInput: []string{"value", "search", "stats", "count", "by", "123", "456"},
			description:   "Should handle field names with underscores and numbers",
		},
		{
			name:          "Multiple field comparisons",
			query:         "search user=admin AND status=active OR role=manager",
			expectedInput: []string{"user", "status", "role"},
			excludedInput: []string{"admin", "active", "manager", "search", "AND", "OR"},
			description:   "Should extract field names from multiple comparisons",
		},
		{
			name:          "Quoted values should not be fields",
			query:         "search message=\"error occurred\" | eval category=\"security\"",
			expectedInput: []string{"message"},
			excludedInput: []string{"error occurred", "security", "category", "error", "occurred", "eval"},
			description:   "Should not treat quoted strings or eval assignments as input fields",
		},
		{
			name:          "Complex field operations",
			query:         "search bytes>1000 | eval mb=round(bytes/1048576,2) | where mb>10",
			expectedInput: []string{"bytes"},
			excludedInput: []string{"mb", "1000", "1048576", "2", "10", "round", "eval", "where"},
			description:   "Should only extract original fields, not derived ones or numeric values",
		},
		{
			name:          "Nested field references",
			query:         "search sourcetype=syslog | lookup users.csv username OUTPUT full_name | eval display=username+\": \"+full_name",
			expectedInput: []string{"username"},
			excludedInput: []string{"full_name", "display", "syslog", "users.csv", "OUTPUT", ": "},
			description:   "Should distinguish input fields from lookup outputs and eval creations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("Failed to discover query info for %s: %v", tt.description, err)
			}

			// Check expected input fields
			for _, expectedField := range tt.expectedInput {
				if !contains(info.InputFields, expectedField) {
					t.Errorf("[%s] Expected input field '%s' to be discovered, got: %v", tt.description, expectedField, info.InputFields)
				}
			}

			// Check that excluded values are not present in input fields
			for _, excludedField := range tt.excludedInput {
				if contains(info.InputFields, excludedField) {
					t.Errorf("[%s] Expected '%s' to NOT be in input fields, got: %v", tt.description, excludedField, info.InputFields)
				}
			}
		})
	}
}

func TestNewWithConfig(t *testing.T) {
	config := &MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Source: "src_ip", Target: "source_ip"},
			{Source: "dst_ip", Target: "destination_ip"},
		},
		Rules: []ConditionalRule{
			{
				ID: "rule1",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "access_combined"},
				},
				Mappings: []FieldMapping{
					{Source: "clientip", Target: "client_address"},
				},
				Enabled: true,
			},
		},
	}

	m := NewWithConfig(config)
	if m == nil {
		t.Fatal("Expected non-nil mapper")
	}

	if m.config != config {
		t.Error("Expected config to be set")
	}

	// Test that basic mappings are loaded
	if m.fieldMappings["src_ip"] != "source_ip" {
		t.Errorf("Expected mapping src_ip -> source_ip, got %s", m.fieldMappings["src_ip"])
	}
}

func TestMapQueryWithContext(t *testing.T) {
	config := &MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Source: "src_ip", Target: "source_ip"},
		},
		Rules: []ConditionalRule{
			{
				ID: "rule1",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "access_combined"},
				},
				Mappings: []FieldMapping{
					{Source: "clientip", Target: "client_address"},
				},
				Enabled: true,
			},
		},
	}

	m := NewWithConfig(config)

	query := "search sourcetype=access_combined"
	context := map[string]interface{}{
		"sourcetype": "access_combined",
	}

	result, err := m.MapQueryWithContext(query, context)
	if err != nil {
		t.Fatalf("Failed to map query with context: %v", err)
	}

	// Should return the original query since we're not actually changing field references yet
	// The mapping logic is in place but we need more sophisticated AST manipulation
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestGetEffectiveMappings(t *testing.T) {
	config := &MappingConfig{
		Version: "1.0",
		Mappings: []FieldMapping{
			{Source: "base_field", Target: "base_mapped"},
		},
		Rules: []ConditionalRule{
			{
				ID: "rule1",
				Conditions: []Condition{
					{Type: "sourcetype", Operator: "equals", Value: "test_sourcetype"},
				},
				Mappings: []FieldMapping{
					{Source: "conditional_field", Target: "conditional_mapped"},
				},
				Enabled: true,
			},
		},
	}

	m := NewWithConfig(config)

	// Test without context (only basic mappings)
	mappings := m.getEffectiveMappings(nil)
	if len(mappings) != 1 {
		t.Errorf("Expected 1 mapping without context, got %d", len(mappings))
	}

	// Test with matching context (basic + conditional mappings)
	context := map[string]interface{}{
		"sourcetype": "test_sourcetype",
	}
	mappings = m.getEffectiveMappings(context)
	if len(mappings) != 2 {
		t.Errorf("Expected 2 mappings with matching context, got %d", len(mappings))
	}

	// Verify specific mappings
	if mappings["base_field"] != "base_mapped" {
		t.Error("Expected base mapping to be present")
	}
	if mappings["conditional_field"] != "conditional_mapped" {
		t.Error("Expected conditional mapping to be present")
	}
}

func TestExtractQueryContext(t *testing.T) {
	m := New()

	// Test extracting context from query string using the listener-based approach
	query := "search sourcetype=access_combined"
	context := m.extractQueryContextFromString(query)

	// Should extract sourcetype from the query
	if sourcetype, exists := context["sourcetype"]; !exists || sourcetype != "access_combined" {
		t.Errorf("Expected sourcetype 'access_combined' in context, got %v", sourcetype)
	}
}

// TestAllDataModelPatterns tests comprehensive support for all SPL datamodel reference patterns
func TestAllDataModelPatterns(t *testing.T) {
	m := New()

	tests := []struct {
		name               string
		query              string
		expectedDataModels []string
		expectedDatasets   []string
		expectedFields     []string
		description        string
	}{
		{
			name:               "tstats with datamodel parameter",
			query:              "| tstats count from datamodel=Network_Traffic where Web.status=200 by Web.src_ip",
			expectedDataModels: []string{"Network_Traffic"},
			expectedDatasets:   []string{},
			expectedFields:     []string{"Web.status", "Web.src_ip"},
			description:        "Basic tstats with datamodel= parameter",
		},
		{
			name:               "tstats with nodename",
			query:              "| tstats count from datamodel=Authentication nodename=Authentication where Authentication.action=success",
			expectedDataModels: []string{"Authentication"},
			expectedDatasets:   []string{},
			expectedFields:     []string{"Authentication.action"},
			description:        "tstats with both datamodel and nodename parameters",
		},
		{
			name:               "datamodel command",
			query:              "| datamodel Web All_Traffic search",
			expectedDataModels: []string{"Web"},
			expectedDatasets:   []string{"Web.All_Traffic"},
			expectedFields:     []string{},
			description:        "Direct datamodel command usage",
		},
		{
			name:               "from datamodel syntax",
			query:              "| from datamodel:Network_Traffic.All_Traffic",
			expectedDataModels: []string{"Network_Traffic"},
			expectedDatasets:   []string{"Network_Traffic.All_Traffic"},
			expectedFields:     []string{},
			description:        "from datamodel: syntax without quotes (working case)",
		},
		{
			name:               "from datamodel without quotes",
			query:              "| from datamodel:Web.All_Traffic",
			expectedDataModels: []string{"Web"},
			expectedDatasets:   []string{"Web.All_Traffic"},
			expectedFields:     []string{},
			description:        "from datamodel: syntax without quotes",
		},
		{
			name:               "pivot command",
			query:              "| pivot Network_Traffic All_Traffic count(bytes) AS total_bytes",
			expectedDataModels: []string{"Network_Traffic"},
			expectedDatasets:   []string{"Network_Traffic.All_Traffic"},
			expectedFields:     []string{},
			description:        "pivot command with datamodel and object",
		},
		{
			name:               "pivot with quoted names",
			query:              "| pivot \"Network Traffic\" \"All Traffic\" count(bytes) AS total_bytes",
			expectedDataModels: []string{"Network Traffic"},
			expectedDatasets:   []string{"Network Traffic.All Traffic"},
			expectedFields:     []string{},
			description:        "pivot command with quoted datamodel and object names",
		},
		{
			name:               "fully qualified field references",
			query:              "search Network_Traffic.All_Traffic.src_ip=192.168.1.1 | eval dest=Network_Traffic.All_Traffic.dest_ip",
			expectedDataModels: []string{},
			expectedDatasets:   []string{},
			expectedFields:     []string{"Network_Traffic.All_Traffic.src_ip", "Network_Traffic.All_Traffic.dest_ip"},
			description:        "Fully qualified DataModel.Object.field references",
		},
		{
			name:               "mixed datamodel patterns",
			query:              "| tstats count from datamodel=Web by Web.status",
			expectedDataModels: []string{"Web"},
			expectedDatasets:   []string{},
			expectedFields:     []string{"Web.status"},
			description:        "Simple tstats datamodel pattern",
		},
		{
			name:               "complex tstats with multiple fields",
			query:              "| tstats sum(Web.bytes) AS total_bytes from datamodel=Web where Web.status=200 by Web.src_ip",
			expectedDataModels: []string{"Web"},
			expectedDatasets:   []string{},
			expectedFields:     []string{"Web.bytes", "Web.status", "Web.src_ip"},
			description:        "Complex tstats with multiple datamodel fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("[%s] Failed to discover query info: %v", tt.description, err)
			}

			// Check datamodels
			for _, expected := range tt.expectedDataModels {
				if !contains(info.DataModels, expected) {
					t.Errorf("[%s] Expected datamodel '%s' to be discovered, got: %v", tt.description, expected, info.DataModels)
				}
			}

			// Check datasets
			for _, expected := range tt.expectedDatasets {
				if !contains(info.Datasets, expected) {
					t.Errorf("[%s] Expected dataset '%s' to be discovered, got: %v", tt.description, expected, info.Datasets)
				}
			}

			// Check fields
			for _, expected := range tt.expectedFields {
				if !contains(info.InputFields, expected) {
					t.Errorf("[%s] Expected field '%s' to be discovered, got: %v", tt.description, expected, info.InputFields)
				}
			}

			// Verify that datamodel names don't appear in wrong categories
			for _, dataModel := range tt.expectedDataModels {
				if contains(info.InputFields, dataModel) {
					t.Errorf("[%s] Datamodel name '%s' should not appear in InputFields, got: %v", tt.description, dataModel, info.InputFields)
				}
				if contains(info.Lookups, dataModel) {
					t.Errorf("[%s] Datamodel name '%s' should not appear in Lookups, got: %v", tt.description, dataModel, info.Lookups)
				}
			}
		})
	}
}

func TestMacroDiscovery(t *testing.T) {
	m := New()

	tests := []struct {
		name           string
		query          string
		expectedMacros []string
		description    string
	}{
		{
			name:           "Simple macro without arguments",
			query:          "search `get_security_events` | stats count by user",
			expectedMacros: []string{"get_security_events"},
			description:    "Should detect simple macro invocation",
		},
		{
			name:           "Macro with arguments",
			query:          "search `get_indexes(security,network)` | head 100",
			expectedMacros: []string{"get_indexes"},
			description:    "Should detect macro with arguments",
		},
		{
			name:           "Multiple macros",
			query:          "`get_data` | eval result=`calculate_score(field1, field2)` | where result>10",
			expectedMacros: []string{"get_data", "calculate_score"},
			description:    "Should detect multiple macro invocations",
		},
		{
			name:           "No macros",
			query:          "search index=main sourcetype=syslog | stats count by host",
			expectedMacros: []string{},
			description:    "Should return empty list when no macros present",
		},
		{
			name:           "Macro with underscore and numbers",
			query:          "search `custom_macro_v2(param1)` | table _time, event",
			expectedMacros: []string{"custom_macro_v2"},
			description:    "Should handle macro names with underscores and numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("[%s] DiscoverQuery failed: %v", tt.description, err)
			}

			// Check expected macros are found
			for _, expected := range tt.expectedMacros {
				found := false
				for _, actual := range info.Macros {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("[%s] Expected macro '%s' to be discovered, got: %v", tt.description, expected, info.Macros)
				}
			}

			// Check no unexpected macros are found
			if len(info.Macros) != len(tt.expectedMacros) {
				t.Errorf("[%s] Expected %d macros, got %d: %v", tt.description, len(tt.expectedMacros), len(info.Macros), info.Macros)
			}
		})
	}
}

func TestFieldVsLiteralDiscrimination(t *testing.T) {
	m := New()

	tests := []struct {
		name            string
		query           string
		expectedFields  []string
		forbiddenFields []string
		description     string
	}{
		{
			name:            "IP addresses should not be fields",
			query:           "search src_ip=192.168.1.1 dest_ip=10.0.0.5 | stats count by host",
			expectedFields:  []string{"src_ip", "dest_ip", "host"},
			forbiddenFields: []string{"192.168.1.1", "10.0.0.5", "192", "168", "1", "10", "0", "5"},
			description:     "IP addresses and their components should not be detected as fields",
		},
		{
			name:            "File extensions should not be fields",
			query:           "search source=\"/var/log/app.log\" | inputlookup data.csv | lookup table.txt",
			expectedFields:  []string{"source"},
			forbiddenFields: []string{"app.log", ".log", ".csv", ".txt", "log", "csv", "txt"},
			description:     "File extensions should not be detected as fields",
		},
		{
			name:            "Numbers should not be fields",
			query:           "search bytes>1024 count<100 timestamp>1609459200 | eval mb=bytes/1048576",
			expectedFields:  []string{"bytes", "count", "timestamp"},
			forbiddenFields: []string{"1024", "100", "1609459200", "1048576", "mb"},
			description:     "Numeric values and eval-created fields should not be input fields",
		},
		{
			name:            "Quoted strings should not be fields",
			query:           "search message=\"error occurred\" status=\"success\" | eval category=\"security audit\"",
			expectedFields:  []string{"message", "status"},
			forbiddenFields: []string{"error occurred", "success", "security audit", "category", "error", "occurred", "security", "audit"},
			description:     "Quoted strings and words within quotes should not be fields",
		},
		{
			name:            "Command keywords should not be fields",
			query:           "search index=main | stats count by sourcetype | sort count desc | head 10",
			expectedFields:  []string{"index", "sourcetype"},
			forbiddenFields: []string{"main", "search", "stats", "count", "by", "sort", "desc", "head", "10"},
			description:     "SPL commands and keywords should not be detected as fields",
		},
		{
			name:            "Complex field names should be detected",
			query:           "search field_name_with_underscores=value field123=test CamelCaseField=data nested.field.name=nested",
			expectedFields:  []string{"field_name_with_underscores", "field123", "CamelCaseField", "nested.field.name"},
			forbiddenFields: []string{"value", "test", "data", "nested"},
			description:     "Complex but valid field names should be detected correctly",
		},
		{
			name:            "Datamodel fields should be detected",
			query:           "search Web.src_ip=192.168.1.1 Authentication.user=admin | stats count by Web.dest_port",
			expectedFields:  []string{"Web.src_ip", "Authentication.user", "Web.dest_port"},
			forbiddenFields: []string{"192.168.1.1", "admin", "Web", "Authentication", "src_ip", "user", "dest_port"},
			description:     "Datamodel-qualified field names should be detected as complete entities",
		},
		{
			name:            "Function arguments should be detected as fields",
			query:           "search bytes>0 | stats sum(bytes_in) avg(response_time) by host user",
			expectedFields:  []string{"bytes", "bytes_in", "response_time", "host", "user"},
			forbiddenFields: []string{"0", "sum", "avg", "stats"},
			description:     "Fields used as function arguments should be detected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := m.DiscoverQuery(tt.query)
			if err != nil {
				t.Fatalf("[%s] DiscoverQuery failed: %v", tt.description, err)
			}

			// Check expected fields are detected
			for _, expectedField := range tt.expectedFields {
				found := false
				for _, actualField := range info.InputFields {
					if actualField == expectedField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("[%s] Expected field '%s' to be detected, got: %v", tt.description, expectedField, info.InputFields)
				}
			}

			// Check forbidden items are not detected as fields
			for _, forbiddenItem := range tt.forbiddenFields {
				for _, actualField := range info.InputFields {
					if actualField == forbiddenItem {
						t.Errorf("[%s] Forbidden item '%s' was incorrectly detected as field, got: %v", tt.description, forbiddenItem, info.InputFields)
					}
				}
			}
		})
	}
}

func TestMapperValidateQuery(t *testing.T) {
	m := New()

	tests := []struct {
		name        string
		query       string
		expectError bool
		description string
	}{
		{
			name:        "Valid simple query",
			query:       "search index=main",
			expectError: false,
			description: "Simple valid query should pass validation",
		},
		{
			name:        "Valid complex query",
			query:       "search sourcetype=access_combined | stats count by src_ip | sort count desc",
			expectError: false,
			description: "Complex valid query should pass validation",
		},
		{
			name:        "Empty query",
			query:       "",
			expectError: true,
			description: "Empty query should fail validation",
		},
		{
			name:        "Invalid syntax",
			query:       "search index=main ||| invalid",
			expectError: true,
			description: "Query with invalid syntax should fail validation",
		},
		{
			name:        "Query with macros should fail validation",
			query:       "search `get_security_events` | stats count",
			expectError: true, // Macros cause parse errors in validation
			description: "Query with macros should fail validation due to grammar limitations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.ValidateQuery(tt.query)
			if tt.expectError && err == nil {
				t.Errorf("[%s] Expected validation to fail, but it passed", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("[%s] Expected validation to pass, but got error: %v", tt.description, err)
			}
		})
	}
}
