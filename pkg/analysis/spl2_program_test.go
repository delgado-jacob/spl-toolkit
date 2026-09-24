package analysis

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func spl2ProgramAnalyze(t *testing.T, text string) *Result {
	t.Helper()
	result, err := Analyze(QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current"})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func spl2ProgramReferences(result *Result, kind, name, role string) []Reference {
	var matches []Reference
	for _, reference := range result.References {
		if (kind == "" || reference.Kind == kind) && (name == "" || reference.NormalizedName == name) && (role == "" || reference.Role == role) {
			matches = append(matches, reference)
		}
	}
	return matches
}

func spl2ProgramDiagnostics(result *Result, code string) []Diagnostic {
	var matches []Diagnostic
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			matches = append(matches, diagnostic)
		}
	}
	return matches
}

func spl2ProgramContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func spl2ProgramRequireV1RequirementKinds(t *testing.T, result *Result) {
	t.Helper()
	if result.Requirements.SchemaVersion != 1 {
		t.Fatalf("requirement schema version = %d", result.Requirements.SchemaVersion)
	}
	for _, item := range result.Requirements.Items {
		switch item.Kind {
		case "module", "module_member", "function", "view", "symbol", "annotation":
			t.Fatalf("symbol binding invented a requirement kind: %+v", item)
		}
	}
}

func spl2ProgramStageContaining(t *testing.T, result *Result, text, command, fragment string) Stage {
	t.Helper()
	for _, stage := range result.Stages {
		if stage.Command != command {
			continue
		}
		source := text[stage.Location.Start.Offset:stage.Location.End.Offset]
		if strings.Contains(source, fragment) {
			return stage
		}
	}
	t.Fatalf("missing %q stage containing %q: %+v", command, fragment, result.Stages)
	return Stage{}
}

func spl2ProgramStageWithin(t *testing.T, result *Result, owner Stage, command string) Stage {
	t.Helper()
	for _, stage := range result.Stages {
		if stage.Command == command && stage.Location.Start.Offset >= owner.Location.Start.Offset && stage.Location.End.Offset <= owner.Location.End.Offset {
			return stage
		}
	}
	t.Fatalf("missing %q stage within %+v: %+v", command, owner, result.Stages)
	return Stage{}
}

func TestSPL2ProgramForwardDeclarationsAndMetadata(t *testing.T) {
	query := `@owned("synthetic")
$consumer = FROM $base | eval normalized=normalize(value);
function normalize($input) { return lower($input); }
@source()
$base = FROM synthetic_events | fields value;
export {normalize, base as source_events};`
	result := spl2ProgramAnalyze(t, query)
	if result.Status != Valid || !result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
		t.Fatalf("forward-bound module = status %q coverage %+v requirements %+v diagnostics %+v", result.Status, result.Coverage, result.Requirements.Coverage, result.Diagnostics)
	}
	spl2ProgramRequireV1RequirementKinds(t, result)
	if spl2HasCode(result, "SPL_UNSUPPORTED_MODULE") {
		t.Fatalf("typed module retained the opaque module diagnostic: %+v", result.Diagnostics)
	}
	if got := spl2ProgramReferences(result, "view", "$base", "read"); len(got) != 1 {
		t.Fatalf("local view references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "function", "normalize", "call"); len(got) != 1 {
		t.Fatalf("local function references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "annotation", "owned", "read"); len(got) != 1 {
		t.Fatalf("owned annotation references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "annotation", "source", "read"); len(got) != 1 {
		t.Fatalf("source annotation references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "symbol", "normalize", "export"); len(got) != 1 {
		t.Fatalf("export references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "symbol", "source_events", "export"); len(got) != 1 || got[0].OriginalName != "source_events" {
		t.Fatalf("view export references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "dataset", "$base", "read"); len(got) != 0 || spl2ProgramContains(result.Dependencies.Datasets, "$base") {
		t.Fatalf("local view leaked as an external dataset: refs=%+v deps=%+v", got, result.Dependencies.Datasets)
	}
	for _, forbidden := range []string{"normalize", "$input", "$base"} {
		if got := spl2ProgramReferences(result, "field", forbidden, ""); len(got) != 0 {
			t.Fatalf("symbol %q leaked into the field namespace: %+v", forbidden, got)
		}
	}
	if !reflect.DeepEqual(result.Dependencies.Datasets, []string{"synthetic_events"}) {
		t.Fatalf("datasets = %+v", result.Dependencies.Datasets)
	}
	for i := 1; i < len(result.Stages); i++ {
		if result.Stages[i-1].Location.Start.Offset > result.Stages[i].Location.Start.Offset {
			t.Fatalf("stages are not stable by source location: %+v", result.Stages)
		}
	}
	create := spl2ProgramReferences(result, "field", "normalized", "create")
	reads := spl2ProgramReferences(result, "field", "value", "read")
	if len(create) != 1 || len(reads) == 0 || len(create[0].OriginReferenceIDs) == 0 {
		t.Fatalf("parameter substitution lost field origins: create=%+v reads=%+v", create, reads)
	}
}

func TestSPL2SymbolsDuplicatesAndUnresolvedExports(t *testing.T) {
	query := `$same = FROM first;
$same = FROM second;
function convert($value) { return $value; }
function convert($other) { return $other; }
import remote as convert from vendor/security;
export missing;`
	result := spl2ProgramAnalyze(t, query)
	duplicates := spl2ProgramDiagnostics(result, CodeDuplicateSymbol)
	unresolved := spl2ProgramDiagnostics(result, CodeUnresolvedSymbol)
	if result.Status != Invalid || len(duplicates) != 3 || len(unresolved) != 1 {
		t.Fatalf("duplicate/unresolved result = status %q duplicates %+v unresolved %+v all %+v", result.Status, duplicates, unresolved, result.Diagnostics)
	}
	for _, diagnostic := range append(duplicates, unresolved...) {
		if diagnostic.Severity != "error" || diagnostic.Category != "contract" || diagnostic.StageID == "" || diagnostic.ScopeID != "scope-0" {
			t.Fatalf("diagnostic ownership = %+v", diagnostic)
		}
	}
	if result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete || result.Requirements.QueryStatus != Invalid {
		t.Fatalf("duplicate/unresolved completeness = analysis %+v requirements %+v", result.Coverage, result.Requirements)
	}
}

func TestSPL2DuplicateSymbolsAreUnusableByDependents(t *testing.T) {
	t.Run("view", func(t *testing.T) {
		query := `$same = FROM first | eval retained=first_value;
$same = FROM second | eval retained=second_value;
$consumer = FROM $same | fields retained;`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Invalid || len(spl2ProgramDiagnostics(result, CodeDuplicateSymbol)) != 1 {
			t.Fatalf("duplicate view result = %+v", result)
		}
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $same")
		if consumer.SemanticComplete {
			t.Fatalf("duplicate view remained usable by its dependent: %+v", consumer)
		}
		if got := spl2ProgramReferences(result, "dataset", "$same", "read"); len(got) != 0 {
			t.Fatalf("duplicate local view leaked as an external dataset: %+v", got)
		}
		for _, dataset := range []string{"first", "second"} {
			if spl2ProgramContains(result.Dependencies.Datasets, dataset) {
				t.Fatalf("unusable duplicate view body %q was analyzed: %+v", dataset, result.Dependencies.Datasets)
			}
		}
	})

	t.Run("function", func(t *testing.T) {
		query := `function same($value) { return lower($value); }
function same($value) { return upper($value); }
function dependent($value) { return same($value); }
$output = FROM synthetic_events | eval normalized=dependent(name);`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Invalid || len(spl2ProgramDiagnostics(result, CodeDuplicateSymbol)) != 1 || len(spl2ProgramDiagnostics(result, CodeInvalidFunctionCall)) == 0 {
			t.Fatalf("duplicate function result = %+v", result)
		}
		dependent := spl2ProgramStageContaining(t, result, query, "function", "function dependent")
		if dependent.SemanticComplete {
			t.Fatalf("duplicate function remained usable transitively: %+v", dependent)
		}
	})
}

func TestSPL2ProgramSourceOrderIsDeterministic(t *testing.T) {
	query := `$consumer = FROM $base | eval normalized=normalize(value);
function normalize($input) { return lower($input); }
$base = FROM synthetic_events | fields value;`
	want, err := json.Marshal(spl2ProgramAnalyze(t, query))
	if err != nil {
		t.Fatal(err)
	}
	for run := 0; run < 5; run++ {
		got, marshalErr := json.Marshal(spl2ProgramAnalyze(t, query))
		if marshalErr != nil {
			t.Fatal(marshalErr)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d changed the bound report\nwant %s\n got %s", run+1, want, got)
		}
	}
}

func TestSPL2ViewSummaryReuseOriginsAndNestedOwnership(t *testing.T) {
	query := `$first = FROM $base | eval first_copy=value;
$second = FROM $base | eval second_copy=value;
$base = FROM synthetic_events | fields value;`
	result := spl2ProgramAnalyze(t, query)
	if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
		t.Fatalf("view reuse result = %+v", result)
	}
	if got := spl2ProgramReferences(result, "dataset", "synthetic_events", "read"); len(got) != 1 {
		t.Fatalf("view body was replayed instead of summarized: %+v", got)
	}
	if got := spl2ProgramReferences(result, "view", "$base", "read"); len(got) != 2 {
		t.Fatalf("view consumer references = %+v", got)
	}
	for _, output := range []string{"first_copy", "second_copy"} {
		created := spl2ProgramReferences(result, "field", output, "create")
		if len(created) != 1 || len(created[0].OriginReferenceIDs) == 0 {
			t.Fatalf("%s origins = %+v", output, created)
		}
	}
	viewScopes := 0
	for _, scope := range result.Scopes {
		switch scope.Kind {
		case "view":
			viewScopes++
			if scope.ParentID != "scope-0" || scope.StageID == "" {
				t.Fatalf("view scope ownership = %+v", scope)
			}
		}
	}
	if viewScopes != 3 {
		t.Fatalf("view scope count = %d, scopes=%+v", viewScopes, result.Scopes)
	}
}

func TestSPL2ViewSummaryFeedsSQLSourceScheduling(t *testing.T) {
	query := `$base = FROM synthetic_events | eval normalized=value;
$consumer = FROM $base WHERE normalized != "" SELECT normalized;`
	result := spl2ProgramAnalyze(t, query)
	if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
		t.Fatalf("SQL local-view result = %+v", result)
	}
	if got := spl2ProgramReferences(result, "view", "$base", "read"); len(got) != 1 {
		t.Fatalf("SQL local-view references = %+v", got)
	}
	if got := spl2ProgramReferences(result, "dataset", "$base", "read"); len(got) != 0 || spl2ProgramContains(result.Dependencies.Datasets, "$base") {
		t.Fatalf("SQL local view leaked as an external dataset: refs=%+v deps=%+v", got, result.Dependencies.Datasets)
	}
	created := spl2ProgramReferences(result, "field", "normalized", "create")
	reads := spl2ProgramReferences(result, "field", "normalized", "read")
	if len(created) != 1 || len(reads) != 2 {
		t.Fatalf("SQL local-view summary references = create %+v reads %+v", created, reads)
	}
	for _, read := range reads {
		if !spl2ProgramContains(read.OriginReferenceIDs, created[0].ID) {
			t.Fatalf("SQL local-view summary lost origin %q: %+v", created[0].ID, read)
		}
	}
}

func TestSPL2ViewCyclesKeepUnrelatedDeclarationsAnalyzable(t *testing.T) {
	query := `$left = FROM $right;
$right = FROM $left;
$ok = FROM synthetic_events | fields retained;`
	result := spl2ProgramAnalyze(t, query)
	cycles := spl2ProgramDiagnostics(result, CodeDeclarationCycle)
	if result.Status != Invalid || len(cycles) != 2 || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete {
		t.Fatalf("view cycle result = status %q cycles %+v coverage %+v requirements %+v", result.Status, cycles, result.Coverage, result.Requirements.Coverage)
	}
	locations := []string{}
	for _, diagnostic := range cycles {
		locations = append(locations, query[diagnostic.Location.Start.Offset:diagnostic.Location.End.Offset])
	}
	sort.Strings(locations)
	if !reflect.DeepEqual(locations, []string{"$left = FROM $right;", "$right = FROM $left;"}) {
		t.Fatalf("cycle owners = %+v", locations)
	}
	if got := spl2ProgramReferences(result, "dataset", "synthetic_events", "read"); len(got) != 1 {
		t.Fatalf("unrelated view was not analyzed: %+v", got)
	}
}

func TestSPL2ViewFailuresPropagateToDependents(t *testing.T) {
	t.Run("uncertain summary", func(t *testing.T) {
		query := `import remote as source from vendor/security;
$base = FROM source | fields source_value;
$consumer = FROM $base | fields source_value;`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Incomplete || !spl2HasCode(result, CodeUnresolvedModule) {
			t.Fatalf("uncertain view result = %+v", result)
		}
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $base")
		if consumer.SemanticComplete {
			t.Fatalf("uncertain view summary reported a complete consumer: %+v", consumer)
		}
	})

	t.Run("transitive cycle", func(t *testing.T) {
		query := `$left = FROM $right;
$right = FROM $left;
$dependent = FROM $left;
$consumer = FROM $dependent | fields retained;`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Invalid || len(spl2ProgramDiagnostics(result, CodeDeclarationCycle)) != 2 {
			t.Fatalf("transitive view cycle result = %+v", result)
		}
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $dependent")
		if dependent.SemanticComplete || consumer.SemanticComplete {
			t.Fatalf("view-cycle dependency remained complete: dependent=%+v consumer=%+v", dependent, consumer)
		}
	})
}

func TestSPL2LocalFunctionBindingAndFailures(t *testing.T) {
	t.Run("forward calls substitute parameters", func(t *testing.T) {
		query := `function outer($input) { return inner($input); }
function inner($value) { return lower($value); }
$output = FROM synthetic_events | eval normalized=outer(name);`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
			t.Fatalf("forward function result = %+v", result)
		}
		for _, local := range []string{"$input", "$value", "outer", "inner"} {
			if got := spl2ProgramReferences(result, "field", local, ""); len(got) != 0 {
				t.Fatalf("local function symbol %q leaked as a field: %+v", local, got)
			}
		}
		created := spl2ProgramReferences(result, "field", "normalized", "create")
		nameReads := spl2ProgramReferences(result, "field", "name", "read")
		if len(created) != 1 || len(nameReads) != 1 || !spl2ProgramContains(created[0].OriginReferenceIDs, nameReads[0].ID) {
			t.Fatalf("substituted origins = create %+v reads %+v", created, nameReads)
		}
	})

	for _, test := range []struct {
		name, query, code string
		count             int
	}{
		{"arity", `function one($value) { return $value; }
$output = FROM synthetic_events | eval bad=one(first, second);`, CodeInvalidFunctionCall, 1},
		{"duplicate parameters", `function same($value, $value) { return $value; }
$output = FROM synthetic_events | eval bad=same(input, input);`, CodeDuplicateSymbol, 1},
		{"invalid body", `function bad($value) { return free_field + $value; }
$output = FROM synthetic_events | eval bad=bad(input);`, CodeUnresolvedSymbol, 1},
		{"unresolved call", `function bad($value) { return missing($value); }
$output = FROM synthetic_events | eval bad=bad(input);`, CodeUnresolvedSymbol, 1},
		{"aggregate position", `function one($value) { return $value; }
$output = FROM synthetic_events | stats one(bytes) AS total;`, CodeInvalidFunctionCall, 1},
		{"recursive cycle", `function loop($value) { return loop($value); }
$output = FROM synthetic_events | eval bad=loop(input);`, CodeDeclarationCycle, 1},
		{"indirect cycle", `function left($value) { return right($value); }
function right($value) { return left($value); }
$output = FROM synthetic_events | eval bad=left(input);`, CodeDeclarationCycle, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := spl2ProgramAnalyze(t, test.query)
			if result.Status != Invalid || len(spl2ProgramDiagnostics(result, test.code)) != test.count || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete {
				t.Fatalf("%s result = %+v", test.name, result)
			}
			for _, diagnostic := range spl2ProgramDiagnostics(result, test.code) {
				if diagnostic.Severity != "error" || diagnostic.Category != "contract" || diagnostic.StageID == "" || diagnostic.ScopeID == "" {
					t.Fatalf("%s diagnostic ownership = %+v", test.name, diagnostic)
				}
			}
		})
	}
}

func TestSPL2LocalFunctionSummaryOwnsDiagnosticsAndParameterRoles(t *testing.T) {
	t.Run("selected shape diagnostic is declaration owned", func(t *testing.T) {
		query := `function bad($items) { return any($items, 1); }
$output = FROM synthetic_events | eval first=bad(items), second=bad(other_items);`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		function := spl2ProgramStageContaining(t, result, query, "function", "function bad")
		if result.Status != Invalid || len(diagnostics) != 1 {
			t.Fatalf("selected function shape result = %+v", result)
		}
		if diagnostics[0].StageID != function.ID || diagnostics[0].ScopeID != function.ScopeID || function.SemanticComplete {
			t.Fatalf("selected function shape ownership: diagnostic=%+v function=%+v", diagnostics[0], function)
		}
		if got := query[diagnostics[0].Location.Start.Offset:diagnostics[0].Location.End.Offset]; got != "1" {
			t.Fatalf("selected function shape location = %q", got)
		}
	})

	t.Run("parameter role survives substitution", func(t *testing.T) {
		query := `function absent($value) { return isnull($value); }
$output = FROM synthetic_events | eval missing=absent(candidate);`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
			t.Fatalf("parameter role result = %+v", result)
		}
		if got := spl2ProgramReferences(result, "field", "candidate", "null_test"); len(got) != 1 {
			t.Fatalf("substituted null-test references = %+v", got)
		}
		if got := spl2ProgramReferences(result, "field", "candidate", "read"); len(got) != 0 {
			t.Fatalf("substituted parameter degraded to a read: %+v", got)
		}
		for _, item := range result.Requirements.Items {
			if item.Kind == "field" && item.Identity == "candidate" {
				t.Fatalf("substituted null-test became a required read: %+v", result.Requirements.Items)
			}
		}
	})

	t.Run("null-test role applies only to direct arguments", func(t *testing.T) {
		query := `function absent($value) { return isnull($value); }
$output = FROM synthetic_events | eval missing=absent(lower(candidate));`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
			t.Fatalf("nested parameter argument result = %+v", result)
		}
		if got := spl2ProgramReferences(result, "field", "candidate", "read"); len(got) != 1 {
			t.Fatalf("nested parameter argument reads = %+v", got)
		}
		if got := spl2ProgramReferences(result, "field", "candidate", "null_test"); len(got) != 0 {
			t.Fatalf("nested parameter argument was misclassified as a null test: %+v", got)
		}
	})

	t.Run("collapsed effective roles evaluate a complex argument once", func(t *testing.T) {
		query := `function inspect($value) { return isnull($value) OR $value == ""; }
$output = FROM synthetic_events | eval flagged=inspect(lower(candidate));`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
			t.Fatalf("collapsed parameter roles result = %+v", result)
		}
		reads := spl2ProgramReferences(result, "field", "candidate", "read")
		if len(reads) != 1 || len(spl2ProgramReferences(result, "field", "candidate", "null_test")) != 0 {
			t.Fatalf("collapsed parameter references = %+v", result.References)
		}
		item := requirementItem(result.Requirements, "field", "candidate", "read")
		if item == nil || len(item.Occurrences) != 1 {
			t.Fatalf("collapsed parameter requirement = %+v", result.Requirements.Items)
		}
		created := spl2ProgramReferences(result, "field", "flagged", "create")
		if len(created) != 1 || !reflect.DeepEqual(created[0].OriginReferenceIDs, []string{reads[0].ID}) {
			t.Fatalf("collapsed parameter origins: created=%+v read=%+v", created, reads[0])
		}
	})
}

func TestSPL2LocalFunctionParameterShadowsViewOnlyInsideBody(t *testing.T) {
	query := `$value = FROM synthetic_events | fields source_value;
function identity($value) { return $value; }
$output = FROM synthetic_events | eval copied=identity(source_value);`
	result := spl2ProgramAnalyze(t, query)
	if result.Status != Valid || !result.Coverage.SemanticComplete || !result.Requirements.Coverage.Complete {
		t.Fatalf("shadowed parameter result = %+v", result)
	}
	if got := spl2ProgramReferences(result, "view", "$value", "read"); len(got) != 0 {
		t.Fatalf("function parameter resolved as the same-spelling view: %+v", got)
	}
	if got := spl2ProgramReferences(result, "field", "$value", ""); len(got) != 0 {
		t.Fatalf("function parameter leaked as a field: %+v", got)
	}
	created := spl2ProgramReferences(result, "field", "copied", "create")
	reads := spl2ProgramReferences(result, "field", "source_value", "read")
	if len(created) != 1 || len(reads) != 2 || !spl2ProgramContains(created[0].OriginReferenceIDs, reads[1].ID) {
		t.Fatalf("shadowed parameter substitution = create %+v reads %+v", created, reads)
	}
}

func TestSPL2FunctionCycleFailurePropagatesTransitively(t *testing.T) {
	query := `function left($value) { return right($value); }
function right($value) { return left($value); }
function dependent($value) { return left($value); }
$output = FROM synthetic_events | eval bad=dependent(input);`
	result := spl2ProgramAnalyze(t, query)
	if result.Status != Invalid || len(spl2ProgramDiagnostics(result, CodeDeclarationCycle)) != 2 || len(spl2ProgramDiagnostics(result, CodeInvalidFunctionCall)) == 0 {
		t.Fatalf("transitive function cycle result = %+v", result)
	}
	dependent := spl2ProgramStageContaining(t, result, query, "function", "function dependent")
	if dependent.SemanticComplete {
		t.Fatalf("function-cycle dependency remained complete: %+v", dependent)
	}
}

func TestSPL2UnknownRuntimeFunctionMatchesStandalonePolicy(t *testing.T) {
	standalone := spl2ProgramAnalyze(t, `FROM main | eval x=map(a)`)
	module := spl2ProgramAnalyze(t, `$view = FROM main | eval x=map(a);`)
	for name, result := range map[string]*Result{"standalone": standalone, "module": module} {
		if result.Status != Incomplete || len(spl2ProgramDiagnostics(result, CodeUnsupportedFunction)) != 1 || spl2HasCode(result, CodeUnresolvedSymbol) {
			t.Fatalf("%s unknown runtime function policy = %+v", name, result)
		}
	}
}

func TestSPL2ProgramContractDiagnosticPreservesSyntaxCompleteness(t *testing.T) {
	standalone := spl2ProgramAnalyze(t, `FROM main | eval value={key: 1, key: 2}`)
	module := spl2ProgramAnalyze(t, `$output = FROM main | eval value={key: 1, key: 2};`)
	for name, result := range map[string]*Result{"standalone": standalone, "module": module} {
		if result.Status != Invalid || !result.Coverage.SyntaxComplete || len(spl2ProgramDiagnostics(result, CodeSyntaxError)) != 1 {
			t.Fatalf("%s contract diagnostic syntax coverage = %+v", name, result)
		}
	}
}

func TestSPL2ProgramParserDiagnosticsUseNarrowestDeclarationOwner(t *testing.T) {
	t.Run("nested command", func(t *testing.T) {
		query := `$output = FROM main | eval value={key: 1, key: 2};`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		eval := spl2ProgramStageContaining(t, result, query, "eval", "eval value")
		if len(diagnostics) != 1 || diagnostics[0].StageID != eval.ID || diagnostics[0].ScopeID != eval.ScopeID || eval.SemanticComplete {
			t.Fatalf("nested parser diagnostic ownership: diagnostics=%+v eval=%+v", diagnostics, eval)
		}
	})

	t.Run("attached annotation", func(t *testing.T) {
		query := `@meta({key: 1, key: 2})
$output = FROM main;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		view := spl2ProgramStageContaining(t, result, query, "view", "$output")
		if len(diagnostics) != 1 || diagnostics[0].StageID != view.ID || diagnostics[0].ScopeID != view.ScopeID || view.SemanticComplete {
			t.Fatalf("annotation parser diagnostic ownership: diagnostics=%+v view=%+v", diagnostics, view)
		}
	})
}

func TestSPL2ProgramParserDiagnosticFailuresPropagateToDependents(t *testing.T) {
	t.Run("view body", func(t *testing.T) {
		query := `$base = FROM main | eval value={key: 1, key: 2};
$dependent = FROM $base | fields value;
$output = FROM $dependent | fields value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		eval := spl2ProgramStageContaining(t, result, query, "eval", "eval value")
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $dependent")
		if result.Status != Invalid || len(diagnostics) != 1 || diagnostics[0].StageID != eval.ID {
			t.Fatalf("view parser failure ownership = %+v", result)
		}
		if eval.SemanticComplete || base.SemanticComplete || dependent.SemanticComplete || consumer.SemanticComplete {
			t.Fatalf("view parser failure did not propagate: eval=%+v base=%+v dependent=%+v consumer=%+v", eval, base, dependent, consumer)
		}
	})

	t.Run("function body", func(t *testing.T) {
		query := `function bad($value) { return {key: $value, key: 1}; }
function dependent($value) { return bad($value); }
$output = FROM synthetic_events | eval flagged=dependent(candidate);`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		bad := spl2ProgramStageContaining(t, result, query, "function", "function bad")
		dependent := spl2ProgramStageContaining(t, result, query, "function", "function dependent")
		caller := spl2ProgramStageContaining(t, result, query, "eval", "eval flagged")
		if result.Status != Invalid || len(diagnostics) != 1 || diagnostics[0].StageID != bad.ID {
			t.Fatalf("function parser failure ownership = %+v", result)
		}
		if bad.SemanticComplete || dependent.SemanticComplete || caller.SemanticComplete {
			t.Fatalf("function parser failure did not propagate: bad=%+v dependent=%+v caller=%+v", bad, dependent, caller)
		}
	})

	t.Run("attached annotation", func(t *testing.T) {
		query := `@meta({key: 1, key: 2})
$base = FROM main;
$dependent = FROM $base;
$output = FROM $dependent;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $dependent")
		if result.Status != Invalid || len(diagnostics) != 1 || diagnostics[0].StageID != base.ID {
			t.Fatalf("annotation parser failure ownership = %+v", result)
		}
		if base.SemanticComplete || dependent.SemanticComplete || consumer.SemanticComplete {
			t.Fatalf("annotation parser failure did not propagate: base=%+v dependent=%+v consumer=%+v", base, dependent, consumer)
		}
	})

	t.Run("SQL projection cannot clear view failure", func(t *testing.T) {
		query := `$base = FROM main WHERE {key: 1, key: 2} SELECT value;
$dependent = FROM $base SELECT value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		where := spl2ProgramStageContaining(t, result, query, "where", "WHERE {key")
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		source := spl2ProgramStageContaining(t, result, query, "from", "FROM main")
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $base")
		baseSelect, dependentSelect := result.Stages[0], result.Stages[0]
		for _, stage := range result.Stages {
			if stage.Command != "select" {
				continue
			}
			if stage.Location.Start.Offset >= base.Location.Start.Offset && stage.Location.End.Offset <= base.Location.End.Offset {
				baseSelect = stage
			}
			if stage.Location.Start.Offset >= dependent.Location.Start.Offset && stage.Location.End.Offset <= dependent.Location.End.Offset {
				dependentSelect = stage
			}
		}
		if len(diagnostics) != 1 || diagnostics[0].StageID != where.ID || where.SemanticComplete {
			t.Fatalf("SQL parser diagnostic ownership: diagnostics=%+v where=%+v", diagnostics, where)
		}
		if base.SemanticComplete || baseSelect.SemanticComplete || dependent.SemanticComplete || consumer.SemanticComplete || dependentSelect.SemanticComplete {
			t.Fatalf("SQL parser failure did not propagate: base=%+v base-select=%+v dependent=%+v source=%+v dependent-select=%+v", base, baseSelect, dependent, consumer, dependentSelect)
		}
		if !source.SemanticComplete {
			t.Fatalf("SQL parser failure contaminated the earlier source stage: %+v", source)
		}
		for _, lineage := range result.Lineage {
			if lineage.StageID == source.ID && lineage.After.Uncertain {
				t.Fatalf("SQL parser failure contaminated earlier lineage: %+v", lineage)
			}
		}
		for _, stageID := range []string{baseSelect.ID, consumer.ID, dependentSelect.ID} {
			found := false
			for _, lineage := range result.Lineage {
				if lineage.StageID != stageID {
					continue
				}
				found = true
				if !lineage.After.Uncertain || lineage.Phase == "project" && (len(lineage.After.Fields) != 1 || lineage.After.Fields[0].Name != "value") {
					t.Fatalf("SQL parser failure lineage lost precision or uncertainty: %+v", lineage)
				}
			}
			if !found {
				t.Fatalf("missing lineage for stage %q: %+v", stageID, result.Lineage)
			}
		}
	})

	t.Run("nested child scope taints owning view", func(t *testing.T) {
		query := `$base = FROM main AS m WHERE EXISTS(SELECT 1 AS found FROM child AS c WHERE c.child_id=m.parent_id AND {key: 1, key: 2}) SELECT value;
$dependent = FROM $base SELECT value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		consumer := spl2ProgramStageContaining(t, result, query, "from", "FROM $base")
		childWhere := result.Stages[0]
		if len(diagnostics) == 1 {
			for _, stage := range result.Stages {
				if stage.ID == diagnostics[0].StageID {
					childWhere = stage
				}
			}
		}
		dependentSelect := result.Stages[0]
		for _, stage := range result.Stages {
			if stage.Command == "select" && stage.ScopeID == consumer.ScopeID {
				dependentSelect = stage
			}
		}
		nestedOwner := false
		for _, scope := range result.Scopes {
			nestedOwner = nestedOwner || scope.ID == childWhere.ScopeID && scope.Kind == "exists"
		}
		if len(diagnostics) != 1 || diagnostics[0].StageID != childWhere.ID || !nestedOwner {
			t.Fatalf("nested parser diagnostic ownership: diagnostics=%+v child-where=%+v base=%+v", diagnostics, childWhere, base)
		}
		if base.SemanticComplete || dependent.SemanticComplete || consumer.SemanticComplete || dependentSelect.SemanticComplete {
			t.Fatalf("nested parser failure did not propagate: base=%+v dependent=%+v source=%+v select=%+v", base, dependent, consumer, dependentSelect)
		}
		projected := false
		for _, lineage := range result.Lineage {
			if lineage.StageID == dependentSelect.ID && lineage.Phase == "project" {
				projected = true
				if !lineage.After.Uncertain || len(lineage.After.Fields) != 1 || lineage.After.Fields[0].Name != "value" {
					t.Fatalf("nested parser failure projection lost precision or uncertainty: %+v", lineage)
				}
			}
		}
		if !projected {
			t.Fatalf("missing dependent projection lineage: %+v", result.Lineage)
		}
	})

	t.Run("annotated view taint survives exact projection chain", func(t *testing.T) {
		query := `@meta({key: 1, key: 2})
$base = FROM main;
$dependent = FROM $base SELECT value;
$output = FROM $dependent SELECT value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		dependent := spl2ProgramStageContaining(t, result, query, "view", "$dependent")
		output := spl2ProgramStageContaining(t, result, query, "view", "$output")
		dependentSource := spl2ProgramStageContaining(t, result, query, "from", "FROM $base")
		outputSource := spl2ProgramStageContaining(t, result, query, "from", "FROM $dependent")
		dependentSelect := spl2ProgramStageWithin(t, result, dependent, "select")
		outputSelect := spl2ProgramStageWithin(t, result, output, "select")
		if len(diagnostics) != 1 || diagnostics[0].StageID != base.ID {
			t.Fatalf("annotated view parser ownership = %+v", result)
		}
		for _, stage := range []Stage{base, dependent, dependentSource, dependentSelect, output, outputSource, outputSelect} {
			if stage.SemanticComplete {
				t.Fatalf("annotated view parser taint did not propagate: %+v", stage)
			}
		}
		for _, stageID := range []string{dependentSource.ID, dependentSelect.ID, outputSource.ID, outputSelect.ID} {
			found := false
			for _, lineage := range result.Lineage {
				if lineage.StageID != stageID {
					continue
				}
				found = true
				if !lineage.After.Uncertain || lineage.Phase == "project" && (len(lineage.After.Fields) != 1 || lineage.After.Fields[0].Name != "value") {
					t.Fatalf("annotated view lineage lost precision or uncertainty: %+v", lineage)
				}
			}
			if !found {
				t.Fatalf("missing annotated view lineage for stage %q: %+v", stageID, result.Lineage)
			}
		}
	})

	t.Run("parser tainted function taints projected view", func(t *testing.T) {
		query := `function broken($value) { return {key: $value, key: 1}; }
$base = FROM main SELECT broken(value) AS value;
$output = FROM $base SELECT value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeSyntaxError)
		function := spl2ProgramStageContaining(t, result, query, "function", "function broken")
		base := spl2ProgramStageContaining(t, result, query, "view", "$base")
		output := spl2ProgramStageContaining(t, result, query, "view", "$output")
		source := spl2ProgramStageContaining(t, result, query, "from", "FROM main")
		baseSelect := spl2ProgramStageWithin(t, result, base, "select")
		outputSource := spl2ProgramStageContaining(t, result, query, "from", "FROM $base")
		outputSelect := spl2ProgramStageWithin(t, result, output, "select")
		if len(diagnostics) != 1 || diagnostics[0].StageID != function.ID {
			t.Fatalf("parser-tainted function ownership = %+v", result)
		}
		for _, stage := range []Stage{function, base, baseSelect, output, outputSource, outputSelect} {
			if stage.SemanticComplete {
				t.Fatalf("parser-tainted function did not propagate: %+v", stage)
			}
		}
		if !source.SemanticComplete {
			t.Fatalf("parser-tainted function contaminated the earlier source: %+v", source)
		}
		for _, lineage := range result.Lineage {
			if lineage.StageID == source.ID && lineage.After.Uncertain {
				t.Fatalf("parser-tainted function contaminated earlier lineage: %+v", lineage)
			}
		}
		for _, stageID := range []string{baseSelect.ID, outputSource.ID, outputSelect.ID} {
			found := false
			for _, lineage := range result.Lineage {
				if lineage.StageID != stageID {
					continue
				}
				found = true
				if !lineage.After.Uncertain || lineage.Phase == "project" && (len(lineage.After.Fields) != 1 || lineage.After.Fields[0].Name != "value") {
					t.Fatalf("parser-tainted function lineage lost precision or uncertainty: %+v", lineage)
				}
			}
			if !found {
				t.Fatalf("missing parser-tainted function lineage for stage %q: %+v", stageID, result.Lineage)
			}
		}
	})
}

func TestSPL2ProgramCompatibilityDiagnosticPreservesSyntaxParity(t *testing.T) {
	standalone := spl2ProgramAnalyze(t, `FROM main | stats mode=fast count()`)
	module := spl2ProgramAnalyze(t, `$output = FROM main | stats mode=fast count();`)
	for name, result := range map[string]*Result{"standalone": standalone, "module": module} {
		if result.Status != Invalid || result.Coverage.SyntaxComplete || len(spl2ProgramDiagnostics(result, "SPL_PROFILE_MISMATCH")) != 1 {
			t.Fatalf("%s compatibility syntax coverage = %+v", name, result)
		}
	}
}

func TestSPL2ExportUnresolvedDiagnosticOwnsLocalName(t *testing.T) {
	query := `export missing as public;`
	result := spl2ProgramAnalyze(t, query)
	diagnostics := spl2ProgramDiagnostics(result, CodeUnresolvedSymbol)
	if result.Status != Invalid || len(diagnostics) != 1 {
		t.Fatalf("unresolved aliased export result = %+v", result)
	}
	if got := query[diagnostics[0].Location.Start.Offset:diagnostics[0].Location.End.Offset]; got != "missing" {
		t.Fatalf("unresolved export diagnostic location = %q", got)
	}
	references := spl2ProgramReferences(result, "symbol", "public", "export")
	if len(references) != 1 || references[0].OriginalName != "public" {
		t.Fatalf("aliased export reference = %+v", references)
	}
}

func TestSPL2QualifiedModuleNameDecodesSegments(t *testing.T) {
	query := `import remote as helper from 'vendor.name'/
security.
'module-name';`
	result := spl2ProgramAnalyze(t, query)
	references := spl2ProgramReferences(result, "module", "vendor.name/security.module-name", "read")
	if result.Status != Valid || len(references) != 1 {
		t.Fatalf("quoted qualified module result = %+v", result)
	}
	if references[0].OriginalName != "'vendor.name'/\nsecurity.\n'module-name'" {
		t.Fatalf("qualified module source spelling = %+v", references[0])
	}
}

func TestSPL2ImportResolutionCompletenessSplit(t *testing.T) {
	t.Run("unused import", func(t *testing.T) {
		query := `import remote as helper from vendor/security;
$output = FROM synthetic_events | fields value;`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeUnresolvedModule)
		if result.Status != Valid || !result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete || result.Requirements.QueryStatus != Incomplete || len(diagnostics) != 1 {
			t.Fatalf("unused import split = result %+v", result)
		}
		spl2ProgramRequireV1RequirementKinds(t, result)
		if diagnostics[0].Severity != "warning" || diagnostics[0].Category != "unsupported_semantics" || diagnostics[0].StageID == "" {
			t.Fatalf("unused import diagnostic = %+v", diagnostics[0])
		}
		if got := spl2ProgramReferences(result, "module", "vendor/security", "read"); len(got) != 1 {
			t.Fatalf("module references = %+v", got)
		}
	})

	t.Run("used member alias", func(t *testing.T) {
		query := `import {normalize as norm} from vendor/security;
$output = FROM synthetic_events | eval result=norm(value);`
		result := spl2ProgramAnalyze(t, query)
		diagnostics := spl2ProgramDiagnostics(result, CodeUnresolvedModule)
		if result.Status != Incomplete || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete || len(diagnostics) != 1 {
			t.Fatalf("used member result = %+v", result)
		}
		spl2ProgramRequireV1RequirementKinds(t, result)
		if got := spl2ProgramReferences(result, "module_member", "vendor/security.normalize", "read"); len(got) != 1 || got[0].OriginalName != "norm" {
			t.Fatalf("used imported member reference = %+v", got)
		}
		if got := spl2ProgramReferences(result, "field", "norm", ""); len(got) != 0 {
			t.Fatalf("import alias leaked as field: %+v", got)
		}
	})

	t.Run("used qualified container", func(t *testing.T) {
		query := `import * as external from vendor/security;
$output = FROM external.events | fields value;`
		result := spl2ProgramAnalyze(t, query)
		if result.Status != Incomplete || result.Coverage.SemanticComplete || result.Requirements.Coverage.Complete || len(spl2ProgramDiagnostics(result, CodeUnresolvedModule)) != 1 {
			t.Fatalf("used container result = %+v", result)
		}
		spl2ProgramRequireV1RequirementKinds(t, result)
		if got := spl2ProgramReferences(result, "module_member", "vendor/security.events", "read"); len(got) != 1 || got[0].OriginalName != "external.events" {
			t.Fatalf("qualified imported reference = %+v", got)
		}
		if spl2ProgramContains(result.Dependencies.Datasets, "external.events") {
			t.Fatalf("qualified import leaked as dataset: %+v", result.Dependencies.Datasets)
		}
		for _, reference := range result.References {
			if reference.Kind == "field" && strings.HasPrefix(reference.NormalizedName, "external") {
				t.Fatalf("qualified import leaked as field: %+v", reference)
			}
		}
	})
}
