package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

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
  analyze [query]   Analyze field availability, lineage, and coverage
  requirements [query]  Report direct external query requirements
  validate-fields   Validate fields against a local catalog
  validate-schema   Validate fields against local JSON Schema or OCSF
  rewrite           Safely preview or apply explicit rewrite rules
  capabilities      Show supported analysis commands and limitations
  demo              Run demonstration examples
  help              Show this help message

Options for map, discover, validate, analyze, requirements, validate-fields, validate-schema, rewrite, and capabilities:
  --format FORMAT   Output format: text or json (default: text)
  --output FILE     Write the result to a file (including analysis diagnostics)
  --help            Show this help message
  --query QUERY     Supply the query as an option (not capabilities)
  --config FILE     Mapping configuration (map and validate only)

Compatibility selectors (analyze, requirements, capabilities, and single-query validation):
  --language LANG                  spl or spl2 (default: spl)
  --profile PROFILE                Execution profile (default: splunkd)
  --compatibility-version VERSION  Compatibility version (default: current)
  --source-id ID                   Source identifier (document operations only)
analyze, requirements, and capabilities accept language/profile/version selectors.
Legacy map/discover/validate reject spl2; use analyze, structured validation, or rewrite where its capability form is supported.
Empty compatibility selectors use defaults; unknown selectors are input errors.

Additional validate-fields options:
  --fields FILE     Required local JSON field catalog (array or object; not -)

Shared validate-fields, validate-schema, and rewrite query inputs:
  --file FILE       Read one query verbatim; default source ID is FILE
  --stdin           Read one query verbatim; default source ID is <stdin>
  --batch FILE      Read a nonempty JSON document array (- reads stdin)
Choose exactly one positional/--query, --file, --stdin, or --batch source.
Batch documents carry their own options; global document options are rejected.

Additional validate-schema options:
  --schema FILE             Local inline JSON Schema (object or boolean; not -)
  --schema-base-uri URI     Explicit root identity; never fetched
  --schema-resources FILE   Local URI-to-inline-schema JSON object (not -)
  --ocsf-catalog FILE       Local official compiled OCSF catalog (not -)
  --ocsf-version VERSION    Exact catalog version (required for OCSF)
  --ocsf-class KEY_OR_UID   Concrete class key or nonnegative decimal UID
  --ocsf-category KEY_OR_UID  Concrete category key or nonnegative decimal UID
  --ocsf-profile NAME       Selected OCSF profile (repeatable)
  --ocsf-extension NAME     Full compiled extension set (repeatable)
Choose exactly one --schema or --ocsf-catalog; target options cannot be mixed.
OCSF requires exactly one --ocsf-class or --ocsf-category selector.

Additional rewrite options:
  --rules FILE      Required local versioned rewrite rule set (not -)
  --apply           Return a verified candidate when safe (default: preview)
  --fields FILE     Optional local field-list validation target
Rewrite also accepts the shared query inputs and optional schema/OCSF target options.
Choose at most one --fields, --schema, or --ocsf-catalog target family.

Analyze, requirements, validate-fields, validate-schema, and rewrite exit codes: 0 valid, 1 invalid content, 3 incomplete analysis or requirement coverage.
Request or output errors exit 2. Invalid and incomplete reports are still emitted.
`, buildinfo.Version)
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
