package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

// Version will be set at build time via ldflags.
var Version = "dev"

func main() {
	os.Exit(runCLI(os.Args[1:], os.Stdout, os.Stderr))
}

func showHelp(w io.Writer) error {
	_, err := fmt.Fprintf(w, `SPL Toolkit %s
Usage: spl-toolkit <command> [options]

Commands:
  version           Show version information
  map [query]       Map fields using --config
  discover [query]  Discover query information
  validate [query]  Validate query syntax or --config
  demo              Run demonstration examples
  help              Show this help message

Options for map, discover, and validate:
  --query QUERY     Supply the query as an option
  --config FILE     Mapping configuration (map and validate only)
  --format FORMAT   Output format: text or json (default: text)
  --output FILE     Write a successful result to a file
  --help            Show this help message
`, Version)
	return err
}

func runDemo(w io.Writer) error {
	if _, err := fmt.Fprintln(w, "SPL Toolkit Library - Demo"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "================================"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "\n1. Basic Field Mapping:"); err != nil {
		return err
	}
	if err := basicMappingDemo(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "\n2. Conditional Field Mapping:"); err != nil {
		return err
	}
	if err := conditionalMappingDemo(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "\n3. Query Discovery:"); err != nil {
		return err
	}
	return queryDiscoveryDemo(w)
}

func basicMappingDemo(w io.Writer) error {
	mappings := []mapper.FieldMapping{
		{Source: "src_ip", Target: "source_ip"},
		{Source: "dst_ip", Target: "destination_ip"},
		{Source: "src_port", Target: "source_port"},
	}
	mappingsJSON, err := json.Marshal(mappings)
	if err != nil {
		return err
	}
	m := mapper.New()
	if err := m.LoadMappings(mappingsJSON); err != nil {
		return fmt.Errorf("load basic mappings: %w", err)
	}
	query := "search src_ip=192.168.1.1 dst_port=80"
	if _, err := fmt.Fprintf(w, "Original query: %s\n", query); err != nil {
		return err
	}
	mappedQuery, err := m.MapQuery(query)
	if err != nil {
		return fmt.Errorf("map basic query: %w", err)
	}
	_, err = fmt.Fprintf(w, "Mapped query: %s\n", mappedQuery)
	return err
}

func conditionalMappingDemo(w io.Writer) error {
	config := &mapper.MappingConfig{
		Version: "1.0",
		Name:    "Web Access Logs Mapping",
		Mappings: []mapper.FieldMapping{
			{Source: "ip", Target: "client_ip"},
		},
		Rules: []mapper.ConditionalRule{
			{
				ID:         "apache_logs",
				Name:       "Apache Access Log Fields",
				Conditions: []mapper.Condition{{Type: "sourcetype", Operator: "equals", Value: "access_combined"}},
				Mappings: []mapper.FieldMapping{
					{Source: "clientip", Target: "source_address"},
					{Source: "status", Target: "http_status_code"},
				},
				Priority: 1,
				Enabled:  true,
			},
			{
				ID:         "nginx_logs",
				Name:       "Nginx Access Log Fields",
				Conditions: []mapper.Condition{{Type: "sourcetype", Operator: "equals", Value: "nginx_access"}},
				Mappings: []mapper.FieldMapping{
					{Source: "remote_addr", Target: "source_address"},
					{Source: "request_status", Target: "http_status_code"},
				},
				Priority: 1,
				Enabled:  true,
			},
		},
	}
	m := mapper.NewWithConfig(config)
	testCases := []struct {
		query   string
		context map[string]interface{}
	}{
		{query: "search sourcetype=access_combined clientip=192.168.1.1", context: map[string]interface{}{"sourcetype": "access_combined"}},
		{query: "search sourcetype=nginx_access remote_addr=10.0.0.1", context: map[string]interface{}{"sourcetype": "nginx_access"}},
	}
	for i, testCase := range testCases {
		if _, err := fmt.Fprintf(w, "\nTest case %d:\nOriginal: %s\nContext: %v\n", i+1, testCase.query, testCase.context); err != nil {
			return err
		}
		mapped, err := m.MapQueryWithContext(testCase.query, testCase.context)
		if err != nil {
			return fmt.Errorf("map conditional query %d: %w", i+1, err)
		}
		if _, err := fmt.Fprintf(w, "Mapped: %s\n", mapped); err != nil {
			return err
		}
	}
	return nil
}

func queryDiscoveryDemo(w io.Writer) error {
	m := mapper.New()
	queries := []string{
		"search sourcetype=access_combined | stats count by src_ip",
		"| inputlookup ip_geo_lookup.csv | search country=US",
		"| datamodel Network_Traffic All_Traffic search | stats avg(bytes_in) by src_ip",
		"| tstats count from datamodel=Web where Web.status=200 by Web.src_ip",
	}
	for i, query := range queries {
		if _, err := fmt.Fprintf(w, "\nQuery %d: %s\n", i+1, query); err != nil {
			return err
		}
		info, err := m.DiscoverQuery(query)
		if err != nil {
			return fmt.Errorf("discover query %d: %w", i+1, err)
		}
		if _, err := fmt.Fprintf(w, "  Data Models: %v\n  Datasets: %v\n  Lookups: %v\n  Macros: %v\n  Sources: %v\n  Source Types: %v\n  Input Fields: %v\n", info.DataModels, info.Datasets, info.Lookups, info.Macros, info.Sources, info.SourceTypes, info.InputFields); err != nil {
			return err
		}
	}
	return nil
}
