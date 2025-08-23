package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// Version will be set at build time via ldflags
var Version = "dev"

func main() {
	if len(os.Args) < 2 {
		showHelp()
		return
	}

	command := os.Args[1]

	switch command {
	case "version", "--version", "-v":
		fmt.Printf("spl-toolkit %s\n", Version)
	case "map":
		mapCommand()
	case "discover":
		discoverCommand()
	case "validate":
		validateCommand()
	case "demo":
		runDemo()
	case "help", "--help", "-h":
		showHelp()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Printf("SPL Toolkit %s\n", Version)
	fmt.Println("Usage: spl-toolkit <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  version           Show version information")
	fmt.Println("  map <query>       Map fields in SPL query")
	fmt.Println("  discover <query>  Discover query information")
	fmt.Println("  validate <query>  Validate SPL query syntax")
	fmt.Println("  demo              Run demonstration examples")
	fmt.Println("  help              Show this help message")
}

func mapCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: spl-toolkit map <query>")
		os.Exit(1)
	}

	query := os.Args[2]
	m := mapper.New()

	result, err := m.MapQuery(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result)
}

func discoverCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: spl-toolkit discover <query>")
		os.Exit(1)
	}

	query := os.Args[2]
	m := mapper.New()

	info, err := m.DiscoverQuery(query)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	result, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(result))
}

func validateCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: spl-toolkit validate <query>")
		os.Exit(1)
	}

	query := os.Args[2]
	p := mapper.NewParser()

	err := p.ValidateQuery(query)
	if err != nil {
		fmt.Printf("Invalid: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Valid")
}

func runDemo() {
	fmt.Println("SPL Toolkit Library - Demo")
	fmt.Println("================================")

	// Example 1: Basic field mapping
	fmt.Println("\n1. Basic Field Mapping:")
	basicMappingDemo()

	// Example 2: Conditional mapping with configuration
	fmt.Println("\n2. Conditional Field Mapping:")
	conditionalMappingDemo()

	// Example 3: Query discovery
	fmt.Println("\n3. Query Discovery:")
	queryDiscoveryDemo()
}

func basicMappingDemo() {
	// Create basic field mappings
	mappings := []mapper.FieldMapping{
		{Source: "src_ip", Target: "source_ip"},
		{Source: "dst_ip", Target: "destination_ip"},
		{Source: "src_port", Target: "source_port"},
	}

	mappingsJSON, _ := json.Marshal(mappings)

	// Initialize mapper and load mappings
	m := mapper.New()
	if err := m.LoadMappings(mappingsJSON); err != nil {
		log.Fatal("Failed to load mappings:", err)
	}

	// Test query
	query := "search src_ip=192.168.1.1 dst_port=80"

	fmt.Printf("Original query: %s\n", query)

	// Map the query
	mappedQuery, err := m.MapQuery(query)
	if err != nil {
		log.Printf("Error mapping query: %v", err)
		return
	}

	fmt.Printf("Mapped query: %s\n", mappedQuery)
}

func conditionalMappingDemo() {
	// Create configuration with conditional rules
	config := &mapper.MappingConfig{
		Version: "1.0",
		Name:    "Web Access Logs Mapping",
		Mappings: []mapper.FieldMapping{
			{Source: "ip", Target: "client_ip"},
		},
		Rules: []mapper.ConditionalRule{
			{
				ID:   "apache_logs",
				Name: "Apache Access Log Fields",
				Conditions: []mapper.Condition{
					{
						Type:     "sourcetype",
						Operator: "equals",
						Value:    "access_combined",
					},
				},
				Mappings: []mapper.FieldMapping{
					{Source: "clientip", Target: "source_address"},
					{Source: "status", Target: "http_status_code"},
				},
				Priority: 1,
				Enabled:  true,
			},
			{
				ID:   "nginx_logs",
				Name: "Nginx Access Log Fields",
				Conditions: []mapper.Condition{
					{
						Type:     "sourcetype",
						Operator: "equals",
						Value:    "nginx_access",
					},
				},
				Mappings: []mapper.FieldMapping{
					{Source: "remote_addr", Target: "source_address"},
					{Source: "request_status", Target: "http_status_code"},
				},
				Priority: 1,
				Enabled:  true,
			},
		},
	}

	// Initialize mapper with configuration
	m := mapper.NewWithConfig(config)

	// Test queries with different contexts
	testCases := []struct {
		query   string
		context map[string]interface{}
	}{
		{
			query:   "search sourcetype=access_combined clientip=192.168.1.1",
			context: map[string]interface{}{"sourcetype": "access_combined"},
		},
		{
			query:   "search sourcetype=nginx_access remote_addr=10.0.0.1",
			context: map[string]interface{}{"sourcetype": "nginx_access"},
		},
	}

	for i, tc := range testCases {
		fmt.Printf("\nTest case %d:\n", i+1)
		fmt.Printf("Original: %s\n", tc.query)
		fmt.Printf("Context: %v\n", tc.context)

		mapped, err := m.MapQueryWithContext(tc.query, tc.context)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}
		fmt.Printf("Mapped: %s\n", mapped)
	}
}

func queryDiscoveryDemo() {
	m := mapper.New()

	// Test different types of queries
	queries := []string{
		"search sourcetype=access_combined | stats count by src_ip",
		"| inputlookup ip_geo_lookup.csv | search country=US",
		"| datamodel Network_Traffic All_Traffic search | stats avg(bytes_in) by src_ip",
		"| tstats count from datamodel=Web where Web.status=200 by Web.src_ip",
	}

	for i, query := range queries {
		fmt.Printf("\nQuery %d: %s\n", i+1, query)

		info, err := m.DiscoverQuery(query)
		if err != nil {
			log.Printf("Error discovering query: %v", err)
			continue
		}

		fmt.Printf("  Data Models: %v\n", info.DataModels)
		fmt.Printf("  Datasets: %v\n", info.Datasets)
		fmt.Printf("  Lookups: %v\n", info.Lookups)
		fmt.Printf("  Source Types: %v\n", info.SourceTypes)
		fmt.Printf("  Sources: %v\n", info.Sources)
		fmt.Printf("  Input Fields: %v\n", info.InputFields)
	}
}
