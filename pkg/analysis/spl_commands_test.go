package analysis

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestTstatsSemantics(t *testing.T) {
	empty := func() Dependencies {
		return Dependencies{Indexes: []string{}, Sources: []string{}, SourceTypes: []string{}, Datasets: []string{}, Lookups: []string{}, DataModels: []string{}, Macros: []string{}}
	}
	tests := []struct {
		name         string
		query        string
		references   []string
		dependencies Dependencies
		fields       []string
		transitions  []string
	}{
		{
			name:       "exact model only",
			query:      `| tstats count FROM datamodel=Authentication`,
			references: []string{"count=>count:field:output:not_applicable:exact", "Authentication=>Authentication:data_model:read:not_applicable:exact"},
			dependencies: func() Dependencies {
				d := empty()
				d.DataModels = []string{"Authentication"}
				return d
			}(),
			fields:      []string{"count:false:ref-0"},
			transitions: []string{"aggregate:count::ref-0:false"},
		},
		{
			name:       "qualified model and dataset",
			query:      `| tstats count FROM datamodel=Network_Traffic.All_Traffic`,
			references: []string{"count=>count:field:output:not_applicable:exact", "Network_Traffic=>Network_Traffic:data_model:read:not_applicable:exact", "Network_Traffic.All_Traffic=>Network_Traffic.All_Traffic:dataset:read:not_applicable:exact"},
			dependencies: func() Dependencies {
				d := empty()
				d.DataModels = []string{"Network_Traffic"}
				d.Datasets = []string{"Network_Traffic.All_Traffic"}
				return d
			}(),
			fields:      []string{"count:false:ref-0"},
			transitions: []string{"aggregate:count::ref-0:false"},
		},
		{
			name:       "search dependencies",
			query:      `| tstats count WHERE index=main source="/var/log/a" sourcetype=syslog`,
			references: []string{"count=>count:field:output:not_applicable:exact", "main=>main:index:read:not_applicable:exact", `"/var/log/a"=>/var/log/a:source:read:not_applicable:exact`, "syslog=>syslog:sourcetype:read:not_applicable:exact"},
			dependencies: func() Dependencies {
				d := empty()
				d.Indexes = []string{"main"}
				d.Sources = []string{"/var/log/a"}
				d.SourceTypes = []string{"syslog"}
				return d
			}(),
			fields:      []string{"count:false:ref-0"},
			transitions: []string{"aggregate:count::ref-0:false"},
		},
		{
			name:         "ordinary predicate fields",
			query:        `| tstats count WHERE All_Traffic.action="allowed" nodename=All_Traffic.All_Traffic`,
			references:   []string{"count=>count:field:output:not_applicable:exact", "All_Traffic.action=>All_Traffic.action:field:filter:source:exact", "nodename=>nodename:field:filter:source:exact"},
			dependencies: empty(),
			fields:       []string{"count:false:ref-0"},
			transitions:  []string{"aggregate:count::ref-0:false"},
		},
		{
			name:         "aliased and implicit aggregate outputs",
			query:        `| tstats summariesonly=true sum(bytes) AS total count avg(duration)`,
			references:   []string{"bytes=>bytes:field:read:source:exact", "total=>total:field:output:not_applicable:exact", "count=>count:field:output:not_applicable:exact", "avg(duration)=>avg(duration):field:output:not_applicable:exact", "duration=>duration:field:read:source:exact"},
			dependencies: empty(),
			fields:       []string{"avg(duration):false:ref-3,ref-4", "count:false:ref-2", "total:false:ref-1,ref-0"},
			transitions:  []string{"aggregate:total:ref-0:ref-1:false", "aggregate:count::ref-2:false", "aggregate:avg(duration):ref-4:ref-3:false"},
		},
		{
			name:         "supported literal options",
			query:        `| tstats summariesonly=true local=false include_reduced_buckets=true allow_old_summaries=false chunk_size=1000 fillnull_value="NULL" prestats=false append=false count`,
			references:   []string{"count=>count:field:output:not_applicable:exact"},
			dependencies: empty(),
			fields:       []string{"count:false:ref-0"},
			transitions:  []string{"aggregate:count::ref-0:false"},
		},
		{
			name:         "case insensitive boolean options",
			query:        `| tstats SuMmArIeSoNlY=TRUE LOCAL=FaLsE include_reduced_buckets=tRuE allow_old_summaries=FALSE prestats=FaLsE append=fAlSe count`,
			references:   []string{"count=>count:field:output:not_applicable:exact"},
			dependencies: empty(),
			fields:       []string{"count:false:ref-0"},
			transitions:  []string{"aggregate:count::ref-0:false"},
		},
		{
			name:         "grouping",
			query:        `| tstats sum(bytes) AS total BY host user`,
			references:   []string{"bytes=>bytes:field:read:source:exact", "total=>total:field:output:not_applicable:exact", "host=>host:field:group:source:exact", "user=>user:field:group:source:exact"},
			dependencies: empty(),
			fields:       []string{"host:false:ref-2", "total:false:ref-1,ref-0", "user:false:ref-3"},
			transitions:  []string{"project:host:ref-2::false", "project:user:ref-3::false", "aggregate:total:ref-0:ref-1:false"},
		},
		{
			name:         "_time span",
			query:        `| tstats count BY _time span=5m host`,
			references:   []string{"count=>count:field:output:not_applicable:exact", "_time=>_time:field:group:source:exact", "host=>host:field:group:source:exact"},
			dependencies: empty(),
			fields:       []string{"_time:false:ref-1", "count:false:ref-0", "host:false:ref-2"},
			transitions:  []string{"project:_time:ref-1::false", "project:host:ref-2::false", "aggregate:count::ref-0:false"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Valid || !result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete || len(result.Stages) != 1 || !result.Stages[0].SemanticComplete {
				t.Fatalf("tstats completeness = status %q coverage %+v stages %+v diagnostics %+v", result.Status, result.Coverage, result.Stages, result.Diagnostics)
			}
			if got := tstatsReferenceFacts(result.References); !reflect.DeepEqual(got, tc.references) {
				t.Fatalf("references\n got: %q\nwant: %q", got, tc.references)
			}
			if !reflect.DeepEqual(result.Dependencies, tc.dependencies) {
				t.Fatalf("dependencies\n got: %+v\nwant: %+v", result.Dependencies, tc.dependencies)
			}
			lineage := result.Lineage[0]
			if lineage.After.Open || lineage.After.Uncertain {
				t.Fatalf("final environment was not exact and closed: %+v", lineage.After)
			}
			if got := tstatsFieldFacts(lineage.After.Fields); !reflect.DeepEqual(got, tc.fields) {
				t.Fatalf("final fields\n got: %q\nwant: %q", got, tc.fields)
			}
			if got := tstatsTransitionFacts(lineage.Transitions); !reflect.DeepEqual(got, tc.transitions) {
				t.Fatalf("transitions\n got: %q\nwant: %q", got, tc.transitions)
			}
		})
	}
}

func TestTstatsPartialBoundaries(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		diagnosticCode    string
		diagnosticSource  string
		diagnosticMessage string
		references        []string
		fields            []string
		transitions       []string
		dataModels        []string
		datasets          []string
		macros            []string
	}{
		{
			name:             "inline macro",
			query:            "| tstats `summariesonly` count FROM datamodel=Authentication.Authentication BY Authentication.user",
			diagnosticCode:   CodeDynamicReference,
			diagnosticSource: "`summariesonly`",
			references:       []string{"summariesonly=>summariesonly:macro:read:not_applicable:exact", "count=>count:field:output:not_applicable:exact", "Authentication=>Authentication:data_model:read:not_applicable:exact", "Authentication.Authentication=>Authentication.Authentication:dataset:read:not_applicable:exact", "Authentication.user=>Authentication.user:field:group:source:exact"},
			fields:           []string{"Authentication.user:true:ref-4", "count:true:ref-1"},
			transitions:      []string{"project:Authentication.user:ref-4::true", "aggregate:count::ref-1:true"},
			dataModels:       []string{"Authentication"},
			datasets:         []string{"Authentication.Authentication"},
			macros:           []string{"summariesonly"},
		},
		{
			name:             "prestats=true",
			query:            `| tstats prestats=true sum(bytes) AS total FROM datamodel=Authentication.Authentication WHERE action=allowed BY host`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "prestats=true",
			references:       []string{"bytes=>bytes:field:read:source:exact", "Authentication=>Authentication:data_model:read:not_applicable:exact", "Authentication.Authentication=>Authentication.Authentication:dataset:read:not_applicable:exact", "action=>action:field:filter:source:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"host:true:ref-4"},
			transitions:      []string{"project:host:ref-4::true"},
			dataModels:       []string{"Authentication"},
			datasets:         []string{"Authentication.Authentication"},
		},
		{
			name:              "mixed-case prestats true",
			query:             `| tstats PrEsTaTs=TrUe sum(bytes) AS total WHERE action=allowed BY host`,
			diagnosticCode:    CodeUnsupportedSemantics,
			diagnosticSource:  "PrEsTaTs=TrUe",
			diagnosticMessage: "tstats result-shape option is unmodeled when true",
			references:        []string{"bytes=>bytes:field:read:source:exact", "action=>action:field:filter:source:exact", "host=>host:field:group:source:exact"},
			fields:            []string{"host:true:ref-2"},
			transitions:       []string{"project:host:ref-2::true"},
		},
		{
			name:             "wildcard grouping",
			query:            `| tstats sum(bytes) AS total count BY host 'host*'`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "'host*'",
			references:       []string{"bytes=>bytes:field:read:source:exact", "total=>total:field:output:not_applicable:exact", "count=>count:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"count:true:ref-2", "host:true:ref-3", "total:true:ref-1,ref-0"},
			transitions:      []string{"project:host:ref-3::true", "aggregate:total:ref-0:ref-1:true", "aggregate:count::ref-2:true"},
		},
		{
			name:             "dynamic catalog identities",
			query:            `| tstats count FROM datamodel="Authentication.*" BY user`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: `"Authentication.*"`,
			references:       []string{"count=>count:field:output:not_applicable:exact", "user=>user:field:group:source:exact"},
			fields:           []string{"count:true:ref-0", "user:true:ref-1"},
			transitions:      []string{"project:user:ref-1::true", "aggregate:count::ref-0:true"},
		},
		{
			name:             "unknown options",
			query:            `| tstats mystery=true sum(bytes) AS total BY host`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "mystery=true",
			references:       []string{"bytes=>bytes:field:read:source:exact", "total=>total:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"host:true:ref-2", "total:true:ref-1,ref-0"},
			transitions:      []string{"project:host:ref-2::true", "aggregate:total:ref-0:ref-1:true"},
		},
		{
			name:             "append=true",
			query:            `| tstats append=true count BY host`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "append=true",
			references:       []string{"count=>count:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"count:true:ref-0", "host:true:ref-1"},
			transitions:      []string{"project:host:ref-1::true", "aggregate:count::ref-0:true"},
		},
		{
			name:              "mixed-case append true",
			query:             `| tstats ApPeNd=TrUe count BY host`,
			diagnosticCode:    CodeUnsupportedSemantics,
			diagnosticSource:  "ApPeNd=TrUe",
			diagnosticMessage: "tstats result-shape option is unmodeled when true",
			references:        []string{"count=>count:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:            []string{"count:true:ref-0", "host:true:ref-1"},
			transitions:       []string{"project:host:ref-1::true", "aggregate:count::ref-0:true"},
		},
		{
			name:             "malformed options",
			query:            `| tstats summariesonly=1 count BY host`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "summariesonly=1",
			references:       []string{"count=>count:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"count:true:ref-0", "host:true:ref-1"},
			transitions:      []string{"project:host:ref-1::true", "aggregate:count::ref-0:true"},
		},
		{
			name:             "PREFIX and exact siblings",
			query:            `| tstats sum(bytes) AS total PREFIX(user) count BY host`,
			diagnosticCode:   CodeUnsupportedSemantics,
			diagnosticSource: "PREFIX(user)",
			references:       []string{"bytes=>bytes:field:read:source:exact", "total=>total:field:output:not_applicable:exact", "user=>user:field:read:source:exact", "count=>count:field:output:not_applicable:exact", "host=>host:field:group:source:exact"},
			fields:           []string{"count:true:ref-3", "host:true:ref-4", "total:true:ref-1,ref-0"},
			transitions:      []string{"project:host:ref-4::true", "aggregate:total:ref-0:ref-1:true", "aggregate:count::ref-3:true"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != Incomplete || !result.Coverage.SyntaxComplete || result.Coverage.SemanticComplete || len(result.Stages) != 1 || result.Stages[0].SemanticComplete {
				t.Fatalf("partial tstats completeness = status %q coverage %+v stages %+v diagnostics %+v", result.Status, result.Coverage, result.Stages, result.Diagnostics)
			}
			if got := tstatsReferenceFacts(result.References); !reflect.DeepEqual(got, tc.references) {
				t.Fatalf("references\n got: %q\nwant: %q", got, tc.references)
			}
			if got := tstatsFieldFacts(result.Lineage[0].After.Fields); !reflect.DeepEqual(got, tc.fields) || !result.Lineage[0].After.Open || !result.Lineage[0].After.Uncertain {
				t.Fatalf("partial final shape = fields %q state %+v, want fields %q open uncertain", got, result.Lineage[0].After, tc.fields)
			}
			if got := tstatsTransitionFacts(result.Lineage[0].Transitions); !reflect.DeepEqual(got, tc.transitions) {
				t.Fatalf("transitions\n got: %q\nwant: %q", got, tc.transitions)
			}
			if !reflect.DeepEqual(append([]string{}, result.Dependencies.DataModels...), append([]string{}, tc.dataModels...)) ||
				!reflect.DeepEqual(append([]string{}, result.Dependencies.Datasets...), append([]string{}, tc.datasets...)) ||
				!reflect.DeepEqual(append([]string{}, result.Dependencies.Macros...), append([]string{}, tc.macros...)) {
				t.Fatalf("dependencies = %+v", result.Dependencies)
			}
			assertTstatsDiagnosticLocation(t, result, tc.diagnosticCode, tc.diagnosticSource)
			if tc.diagnosticMessage != "" {
				assertTstatsDiagnosticMessage(t, result, tc.diagnosticCode, tc.diagnosticSource, tc.diagnosticMessage)
			}
		})
	}
}

func TestTstatsCombinedBoundariesRetainRequiredWhere(t *testing.T) {
	query := `| tstats mystery=true count FROM datamodel="Auth.*" WHERE action=allowed BY host`
	result, err := Analyze(QueryDocument{Text: query})
	if err != nil {
		t.Fatal(err)
	}
	wantReferences := []string{
		"count=>count:field:output:not_applicable:exact",
		"action=>action:field:filter:source:exact",
		"host=>host:field:group:source:exact",
	}
	if got := tstatsReferenceFacts(result.References); !reflect.DeepEqual(got, wantReferences) {
		t.Fatalf("combined-boundary references\n got: %q\nwant: %q", got, wantReferences)
	}
	wantRequirements := []string{
		"req-1:field:action:filter:required:exact:ref-1:action:source",
		"req-2:field:host:group:required:exact:ref-2:host:source",
	}
	if got := tstatsRequirementFacts(result.Requirements.Items); !reflect.DeepEqual(got, wantRequirements) {
		t.Fatalf("combined-boundary requirements\n got: %q\nwant: %q", got, wantRequirements)
	}
}

func tstatsReferenceFacts(references []Reference) []string {
	facts := make([]string, 0, len(references))
	for _, reference := range references {
		facts = append(facts, fmt.Sprintf("%s=>%s:%s:%s:%s:%s", reference.OriginalName, reference.NormalizedName, reference.Kind, reference.Role, reference.Binding, reference.Resolution))
	}
	return facts
}

func tstatsFieldFacts(fields []FieldBinding) []string {
	facts := make([]string, 0, len(fields))
	for _, field := range fields {
		facts = append(facts, fmt.Sprintf("%s:%t:%s", field.Name, field.Conditional, strings.Join(field.OriginReferenceIDs, ",")))
	}
	return facts
}

func tstatsTransitionFacts(transitions []Transition) []string {
	facts := make([]string, 0, len(transitions))
	for _, transition := range transitions {
		facts = append(facts, fmt.Sprintf("%s:%s:%s:%s:%t", transition.Operation, transition.Output, strings.Join(transition.InputReferenceIDs, ","), transition.OutputReferenceID, transition.Conditional))
	}
	return facts
}

func assertTstatsDiagnosticLocation(t *testing.T, result *Result, code, source string) {
	t.Helper()
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code != code {
			continue
		}
		if got := result.Document.Text[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset]; got != source {
			t.Fatalf("diagnostic %s location = %q, want %q: %+v", code, got, source, diagnostic)
		}
		return
	}
	t.Fatalf("missing diagnostic %s in %+v", code, result.Diagnostics)
}

func assertTstatsDiagnosticMessage(t *testing.T, result *Result, code, source, message string) {
	t.Helper()
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code != code || diagnostic.Message != message {
			continue
		}
		if got := result.Document.Text[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset]; got == source {
			return
		}
	}
	t.Fatalf("missing diagnostic %s with source %q and message %q in %+v", code, source, message, result.Diagnostics)
}
