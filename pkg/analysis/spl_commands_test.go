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

func TestFillnullSemantics(t *testing.T) {
	t.Run("exact target flows conditionally downstream", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search user=* | fillnull value="unknown" user | where user="svc"`)
		assertFieldCommandComplete(t, result, "fillnull")
		assertFieldCommandReference(t, result, "fillnull", "user", "read", "source", "exact")
		assertFieldCommandReference(t, result, "fillnull", "user", "create", "not_applicable", "exact")
		assertFieldCommandReference(t, result, "where", "user", "read", "indeterminate", "exact")
		assertFieldCommandTransition(t, result, "fillnull", "user", true)
	})

	t.Run("all fields preserves a closed environment", func(t *testing.T) {
		result := analyzeFieldCommand(t, `| eval user="known" | table user | fillnull value="unknown" | where user="svc"`)
		assertFieldCommandComplete(t, result, "fillnull")
		lineage := fieldCommandLineage(t, result, "fillnull")
		if !reflect.DeepEqual(lineage.Before, lineage.After) || lineage.After.Open || lineage.After.Uncertain {
			t.Fatalf("fillnull without targets changed the field environment: before=%+v after=%+v", lineage.Before, lineage.After)
		}
		assertFieldCommandReference(t, result, "where", "user", "read", "derived", "exact")
	})

	t.Run("dynamic selector is held without hiding known flow", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search user=* | fillnull 'user*' | where user="svc"`)
		assertFieldCommandIncomplete(t, result, "fillnull", `'user*'`)
		assertFieldCommandReference(t, result, "where", "user", "read", "source", "exact")
		if fieldCommandHasReference(result, "fillnull", "user*", "create") {
			t.Fatalf("dynamic fillnull selector became a literal output: %+v", result.References)
		}
	})
}

func TestRexSemantics(t *testing.T) {
	t.Run("exact input offset and captures flow downstream", func(t *testing.T) {
		query := `search payload=* | rex field=payload max_match=2 offset_field=positions "(?<user>[^ ]+)(?P<host>.+)(?<user>x)" | where user="svc" AND host="web" AND positions>0`
		result := analyzeFieldCommand(t, query)
		assertFieldCommandComplete(t, result, "rex")
		assertFieldCommandReference(t, result, "rex", "payload", "read", "source", "exact")
		for _, name := range []string{"positions", "user", "host"} {
			assertFieldCommandReference(t, result, "rex", name, "output", "not_applicable", "exact")
			assertFieldCommandReference(t, result, "where", name, "read", "indeterminate", "exact")
			assertFieldCommandTransition(t, result, "rex", name, true)
		}
		if got := fieldCommandReferenceCount(result, "rex", "user", "output"); got != 1 {
			t.Fatalf("deduplicated rex capture count = %d, want 1: %+v", got, result.References)
		}
	})

	t.Run("escaped and class-contained openers are ignored", func(t *testing.T) {
		query := `search payload=* | rex field=payload "\(?<escaped>x)[(?<class>x)](?<kept>.+)" | where kept="x"`
		result := analyzeFieldCommand(t, query)
		assertFieldCommandComplete(t, result, "rex")
		assertFieldCommandReference(t, result, "rex", "kept", "output", "not_applicable", "exact")
		if fieldCommandHasReference(result, "rex", "escaped", "output") || fieldCommandHasReference(result, "rex", "class", "output") {
			t.Fatalf("rex scanner accepted escaped or class-contained openers: %+v", result.References)
		}
		assertFieldCommandReference(t, result, "where", "kept", "read", "indeterminate", "exact")
	})

	t.Run("literal closing bracket stays in the character class", func(t *testing.T) {
		query := `search payload=* | rex field=payload "[](?<fake>)](?<real>x)" | where real="x"`
		result := analyzeFieldCommand(t, query)
		assertFieldCommandComplete(t, result, "rex")
		assertFieldCommandReference(t, result, "rex", "real", "output", "not_applicable", "exact")
		assertFieldCommandReferenceLocation(t, result, "rex", "real", "output", "real")
		if fieldCommandHasReference(result, "rex", "fake", "output") {
			t.Fatalf("rex scanner treated a leading literal ] as the class close: %+v", result.References)
		}
		assertFieldCommandReference(t, result, "where", "real", "read", "indeterminate", "exact")
	})

	t.Run("sed mode keeps the input identity and is held", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | rex mode=sed field=payload "s/a/b/g" | where payload="b"`)
		assertFieldCommandIncomplete(t, result, "rex", "mode=sed")
		assertFieldCommandReference(t, result, "rex", "payload", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "payload", "read", "source", "exact")
	})

	t.Run("malformed capture opener is held at the expression", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | rex field=payload "(?<bad-name>.+)" | where payload="x"`)
		assertFieldCommandIncomplete(t, result, "rex", `"(?<bad-name>.+)"`)
		assertFieldCommandReference(t, result, "where", "payload", "read", "source", "exact")
	})

	t.Run("incomplete Python capture opener is held at the expression", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | rex field=payload "(?P" | where payload="x"`)
		assertFieldCommandIncomplete(t, result, "rex", `"(?P"`)
		assertFieldCommandReference(t, result, "rex", "payload", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "payload", "read", "source", "exact")
	})
}

func TestRexNamedCaptures(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    []string
		ok      bool
	}{
		{name: "incomplete Python opener", pattern: `(?P`, ok: false},
		{name: "leading literal closing bracket", pattern: `[](?<fake>)](?<real>x)`, want: []string{"real"}, ok: true},
		{name: "second caret is a class member", pattern: `[^^](?<real>x)`, want: []string{"real"}, ok: true},
		{name: "escaped and class-contained openers", pattern: `\(?<escaped>x)[(?P<class>x)](?<real>x)`, want: []string{"real"}, ok: true},
		{name: "source order and deduplication", pattern: `(?P<two>x)(?<one>y)(?<two>z)`, want: []string{"two", "one"}, ok: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := rexNamedCaptures(tc.pattern)
			if ok != tc.ok || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("rexNamedCaptures(%q) = (%v, %t), want (%v, %t)", tc.pattern, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestSpathSemantics(t *testing.T) {
	t.Run("exact input path and output flow downstream", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | spath input=payload path="event.id" output=event_id | where event_id="42"`)
		assertFieldCommandComplete(t, result, "spath")
		assertFieldCommandReference(t, result, "spath", "payload", "read", "source", "exact")
		assertFieldCommandReference(t, result, "spath", "event_id", "output", "not_applicable", "exact")
		assertFieldCommandReference(t, result, "where", "event_id", "read", "indeterminate", "exact")
		assertFieldCommandTransition(t, result, "spath", "event_id", true)
	})

	t.Run("auto extraction is held after retaining the input read", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | spath input=payload | where payload="x"`)
		assertFieldCommandIncomplete(t, result, "spath", "input=payload")
		assertFieldCommandReference(t, result, "spath", "payload", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "payload", "read", "source", "exact")
	})

	t.Run("dynamic path is held without inventing an output", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search payload=* | spath input=payload path=event_path output=event_id | where payload="x"`)
		assertFieldCommandIncomplete(t, result, "spath", "path=event_path")
		assertFieldCommandReference(t, result, "spath", "payload", "read", "source", "exact")
		if fieldCommandHasReference(result, "spath", "event_id", "output") {
			t.Fatalf("spath with dynamic path invented an output: %+v", result.References)
		}
		assertFieldCommandReference(t, result, "where", "payload", "read", "source", "exact")
	})
}

func TestBinAndBucketSemantics(t *testing.T) {
	t.Run("bin replaces its exact input", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search _time=* | bin span=5m _time | where _time>0`)
		assertFieldCommandComplete(t, result, "bin")
		assertFieldCommandReference(t, result, "bin", "_time", "read", "source", "exact")
		assertFieldCommandReference(t, result, "bin", "_time", "create", "not_applicable", "exact")
		assertFieldCommandReference(t, result, "where", "_time", "read", "derived", "exact")
		assertFieldCommandTransition(t, result, "bin", "_time", false)
	})

	t.Run("bucket alias leaves its source present", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search bytes=* | bucket bins=10 bytes AS band | where bytes>0 AND band>0`)
		assertFieldCommandComplete(t, result, "bucket")
		assertFieldCommandReference(t, result, "bucket", "bytes", "read", "source", "exact")
		assertFieldCommandReference(t, result, "bucket", "band", "create", "not_applicable", "exact")
		assertFieldCommandReference(t, result, "where", "bytes", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "band", "read", "derived", "exact")
		assertFieldCommandTransition(t, result, "bucket", "band", false)
	})

	t.Run("unsupported option retains the exact source fact", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search bytes=* | bin mystery=5 bytes | where bytes>0`)
		assertFieldCommandIncomplete(t, result, "bin", "mystery=5")
		assertFieldCommandReference(t, result, "bin", "bytes", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "bytes", "read", "source", "exact")
	})

	t.Run("dynamic bin span retains the exact source fact", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search bytes=* | bin span=foo bytes | where bytes>0`)
		assertFieldCommandIncomplete(t, result, "bin", "span=foo")
		assertFieldCommandReference(t, result, "bin", "bytes", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "bytes", "read", "source", "exact")
		if fieldCommandHasReference(result, "bin", "bytes", "create") {
			t.Fatalf("dynamic bin span applied a false assignment: %+v", result.References)
		}
	})

	t.Run("dynamic bucket minspan retains source and suppresses alias", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search bytes=* | bucket minspan=dynamic bytes AS band | where bytes>0`)
		assertFieldCommandIncomplete(t, result, "bucket", "minspan=dynamic")
		assertFieldCommandReference(t, result, "bucket", "bytes", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "bytes", "read", "source", "exact")
		if fieldCommandHasReference(result, "bucket", "band", "create") {
			t.Fatalf("dynamic bucket minspan applied a false alias: %+v", result.References)
		}
	})
}

func TestRegexSemantics(t *testing.T) {
	t.Run("exact field comparison is a filter read", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search user=* | regex user!="^svc_" | where user="alice"`)
		assertFieldCommandComplete(t, result, "regex")
		assertFieldCommandReference(t, result, "regex", "user", "filter", "source", "exact")
		assertFieldCommandReference(t, result, "where", "user", "read", "source", "exact")
		if len(fieldCommandLineage(t, result, "regex").Transitions) != 0 {
			t.Fatalf("regex changed field shape: %+v", fieldCommandLineage(t, result, "regex"))
		}
	})

	t.Run("quoted-only form reads the implicit raw field", func(t *testing.T) {
		result := analyzeFieldCommand(t, `| regex "error" | where _raw="error"`)
		assertFieldCommandComplete(t, result, "regex")
		assertFieldCommandReference(t, result, "regex", "_raw", "filter", "source", "exact")
		assertFieldCommandReference(t, result, "where", "_raw", "read", "source", "exact")
	})

	t.Run("dynamic field is held without changing known shape", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search user=* | regex 'user*'="x" | where user="alice"`)
		assertFieldCommandIncomplete(t, result, "regex", `'user*'`)
		assertFieldCommandReference(t, result, "regex", "user*", "filter", "source", "dynamic")
		assertFieldCommandReference(t, result, "where", "user", "read", "source", "exact")
	})
}

func TestMvexpandSemantics(t *testing.T) {
	t.Run("exact field identity flows downstream", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search values=* | mvexpand values | where values="x"`)
		assertFieldCommandComplete(t, result, "mvexpand")
		assertFieldCommandReference(t, result, "mvexpand", "values", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "values", "read", "source", "exact")
		if len(fieldCommandLineage(t, result, "mvexpand").Transitions) != 0 {
			t.Fatalf("mvexpand changed field identity: %+v", fieldCommandLineage(t, result, "mvexpand"))
		}
	})

	t.Run("unsupported option retains the exact read", func(t *testing.T) {
		result := analyzeFieldCommand(t, `search values=* | mvexpand values limit=3 | where values="x"`)
		assertFieldCommandIncomplete(t, result, "mvexpand", "limit=3")
		assertFieldCommandReference(t, result, "mvexpand", "values", "read", "source", "exact")
		assertFieldCommandReference(t, result, "where", "values", "read", "source", "exact")
	})
}

func analyzeFieldCommand(t *testing.T, query string) *Result {
	t.Helper()
	result, err := Analyze(QueryDocument{Text: query})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func assertFieldCommandComplete(t *testing.T, result *Result, command string) {
	t.Helper()
	for _, stage := range result.Stages {
		if stage.Command == command {
			if result.Status != Valid || !stage.SemanticComplete {
				t.Fatalf("%s completeness = status %q stage %+v diagnostics %+v", command, result.Status, stage, result.Diagnostics)
			}
			return
		}
	}
	t.Fatalf("missing %s stage: %+v", command, result.Stages)
}

func assertFieldCommandIncomplete(t *testing.T, result *Result, command, diagnosticSource string) {
	t.Helper()
	if result.Status != Incomplete {
		t.Fatalf("%s status = %q, want incomplete: %+v", command, result.Status, result.Diagnostics)
	}
	for _, stage := range result.Stages {
		if stage.Command == command && !stage.SemanticComplete {
			assertTstatsDiagnosticLocation(t, result, CodeUnsupportedSemantics, diagnosticSource)
			return
		}
	}
	t.Fatalf("missing incomplete %s stage: %+v", command, result.Stages)
}

func assertFieldCommandReference(t *testing.T, result *Result, command, name, role, binding, resolution string) {
	t.Helper()
	stageID := ""
	for _, stage := range result.Stages {
		if stage.Command == command {
			stageID = stage.ID
		}
	}
	for _, reference := range result.References {
		if reference.StageID == stageID && reference.NormalizedName == name && reference.Role == role {
			if reference.Binding != binding || reference.Resolution != resolution {
				t.Fatalf("%s %s/%s reference = %+v, want binding=%s resolution=%s", command, name, role, reference, binding, resolution)
			}
			return
		}
	}
	t.Fatalf("missing %s %s/%s reference: %+v", command, name, role, result.References)
}

func assertFieldCommandReferenceLocation(t *testing.T, result *Result, command, name, role, source string) {
	t.Helper()
	stageID := ""
	for _, stage := range result.Stages {
		if stage.Command == command {
			stageID = stage.ID
		}
	}
	for _, reference := range result.References {
		if reference.StageID != stageID || reference.NormalizedName != name || reference.Role != role {
			continue
		}
		if got := result.Document.Text[reference.Location.Start.Offset:reference.Location.End.Offset]; got != source {
			t.Fatalf("%s %s/%s location = %q, want %q: %+v", command, name, role, got, source, reference)
		}
		return
	}
	t.Fatalf("missing %s %s/%s reference: %+v", command, name, role, result.References)
}

func fieldCommandHasReference(result *Result, command, name, role string) bool {
	stageID := ""
	for _, stage := range result.Stages {
		if stage.Command == command {
			stageID = stage.ID
		}
	}
	for _, reference := range result.References {
		if reference.StageID == stageID && reference.NormalizedName == name && reference.Role == role {
			return true
		}
	}
	return false
}

func fieldCommandReferenceCount(result *Result, command, name, role string) int {
	count := 0
	stageID := ""
	for _, stage := range result.Stages {
		if stage.Command == command {
			stageID = stage.ID
		}
	}
	for _, reference := range result.References {
		if reference.StageID == stageID && reference.NormalizedName == name && reference.Role == role {
			count++
		}
	}
	return count
}

func fieldCommandLineage(t *testing.T, result *Result, command string) Lineage {
	t.Helper()
	stageID := ""
	for _, stage := range result.Stages {
		if stage.Command == command {
			stageID = stage.ID
		}
	}
	for _, lineage := range result.Lineage {
		if lineage.StageID == stageID {
			return lineage
		}
	}
	t.Fatalf("missing %s lineage: %+v", command, result.Lineage)
	return Lineage{}
}

func assertFieldCommandTransition(t *testing.T, result *Result, command, output string, conditional bool) {
	t.Helper()
	lineage := fieldCommandLineage(t, result, command)
	for _, transition := range lineage.Transitions {
		if transition.Output == output {
			if transition.Conditional != conditional || len(transition.InputReferenceIDs) != 1 {
				t.Fatalf("%s transition for %s = %+v, want one input and conditional=%t", command, output, transition, conditional)
			}
			return
		}
	}
	t.Fatalf("missing %s transition for %s: %+v", command, output, lineage.Transitions)
}
