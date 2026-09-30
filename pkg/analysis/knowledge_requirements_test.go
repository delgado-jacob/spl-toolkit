package analysis

import (
	"reflect"
	"testing"
)

func TestKnowledgeRequirementsExactClassicReferences(t *testing.T) {
	for _, tc := range []struct {
		query    string
		kind     string
		name     string
		spelling string
	}{
		{`| from savedsearch:Daily`, "saved_search", "Daily", "Daily"},
		{`| from "savedsearch:Daily Report"`, "saved_search", "Daily Report", "Daily Report"},
	} {
		t.Run(tc.query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: tc.query})
			if err != nil {
				t.Fatal(err)
			}
			var found *Reference
			for i := range result.References {
				if result.References[i].Kind == tc.kind {
					found = &result.References[i]
				}
			}
			if found == nil || found.NormalizedName != tc.name || found.OriginalName != tc.spelling || found.Resolution != "exact" || tc.query[found.Location.Start.Offset:found.Location.End.Offset] != tc.spelling {
				t.Fatalf("reference = %+v, all references = %+v", found, result.References)
			}
			var item *RequirementItem
			for i := range result.Requirements.Items {
				if result.Requirements.Items[i].Kind == tc.kind {
					item = &result.Requirements.Items[i]
				}
			}
			if item == nil || item.Identity != tc.name || item.Necessity != "required" || len(item.Occurrences) != 1 || item.Occurrences[0].ReferenceID != found.ID {
				t.Fatalf("requirement = %+v, all requirements = %+v", item, result.Requirements.Items)
			}
		})
	}
}

func TestKnowledgeRequirementsImportedSPL2Symbols(t *testing.T) {
	query := `import {normalize as norm} from vendor/security;
$output = FROM synthetic_events | eval result=norm(value);`
	result := spl2ProgramAnalyze(t, query)
	want := map[string]string{"module": "vendor/security", "function": "vendor/security.normalize"}
	for kind, name := range want {
		found := false
		for _, item := range result.Requirements.Items {
			if item.Kind == kind && item.Identity == name && item.Resolution == "exact" && len(item.Occurrences) == 1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing %s %s: %+v", kind, name, result.Requirements.Items)
		}
	}
	for _, item := range result.Requirements.Items {
		if item.Kind == "function" && item.Identity == "norm" {
			t.Fatalf("alias became external identity: %+v", item)
		}
	}
}

func TestKnowledgeRequirementsImportedCallInsideLocalFunction(t *testing.T) {
	query := `import {normalize as norm} from vendor/security;
function wrap($x) { return norm($x); }
$output = FROM synthetic_events | eval result=wrap(value);`
	result := spl2ProgramAnalyze(t, query)
	var reference *Reference
	for i := range result.References {
		candidate := &result.References[i]
		if candidate.Kind == "function" && candidate.NormalizedName == "vendor/security.normalize" {
			reference = candidate
		}
	}
	if reference == nil || reference.Resolution != "exact" || query[reference.Location.Start.Offset:reference.Location.End.Offset] != "norm" {
		t.Fatalf("imported call reference = %+v, all references = %+v", reference, result.References)
	}
	found := false
	for _, item := range result.Requirements.Items {
		if item.Kind == "function" && item.Identity == "vendor/security.normalize" && item.Resolution == "exact" && len(item.Occurrences) == 1 && item.Occurrences[0].ReferenceID == reference.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("imported call requirement missing: %+v", result.Requirements.Items)
	}
}

func TestKnowledgeRequirementsLocalAndBuiltinSPL2Functions(t *testing.T) {
	query := `function normalize($input) { return lower($input); }
$output = FROM synthetic_events | eval result=normalize(value);`
	result := spl2ProgramAnalyze(t, query)
	if !result.Coverage.SyntaxComplete || !result.Coverage.SemanticComplete {
		t.Fatalf("local and built-in functions were not analyzed: %+v", result.Diagnostics)
	}
	for _, item := range result.Requirements.Items {
		if item.Kind == "function" || item.Kind == "module" {
			t.Fatalf("local or built-in function became external: %+v", item)
		}
	}
}

func TestKnowledgeRequirementsNoFalseClassicReferences(t *testing.T) {
	for _, query := range []string{`eventtype=authentication`, `tag=identity`, `eventtype=auth*`, `tag="id*"`, `eventtype_name=authentication`, `tagged=identity`, `"eventtype"=authentication`, `"tag"=identity`, `search message="eventtype=authentication"`, `| eval marker="tag=identity"`} {
		t.Run(query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: query})
			if err != nil {
				t.Fatal(err)
			}
			for _, reference := range result.References {
				if reference.Kind == "event_type" || reference.Kind == "tag" {
					t.Fatalf("non-knowledge form became a knowledge reference: %+v", reference)
				}
			}
		})
	}
}

func TestKnowledgeRequirementsDynamicSavedSearchIsIncomplete(t *testing.T) {
	for _, query := range []string{`| from savedsearch:Daily*`, `| from "savedsearch:$daily$"`} {
		t.Run(query, func(t *testing.T) {
			result, err := Analyze(QueryDocument{Text: query})
			if err != nil {
				t.Fatal(err)
			}
			if result.Requirements.Coverage.Complete {
				t.Fatalf("dynamic saved search became complete: %+v", result.Requirements)
			}
			for _, item := range result.Requirements.Items {
				if (item.Kind == "saved_search" || item.Kind == "dataset") && item.Resolution == "exact" {
					t.Fatalf("dynamic saved search became exact: %+v", item)
				}
			}
		})
	}
}

func TestKnowledgeRequirementsSavedSearchKeepsDatasetProjection(t *testing.T) {
	result, err := Analyze(QueryDocument{Text: `| from savedsearch:Daily`})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Dependencies.Datasets, []string{"savedsearch:Daily"}) {
		t.Fatalf("dataset compatibility projection = %+v", result.Dependencies)
	}
}

func TestKnowledgeRequirementsWildcardImportAliasIsNotFunction(t *testing.T) {
	query := `import * as external from vendor/security;
$output = FROM synthetic_events | eval result=external(value);`
	result := spl2ProgramAnalyze(t, query)
	if result.Requirements.Coverage.Complete || result.Coverage.SemanticComplete {
		t.Fatalf("wildcard import alias call became complete: status %s coverage %+v requirements %+v", result.Status, result.Coverage, result.Requirements.Coverage)
	}
	for _, reference := range result.References {
		if reference.Kind == "function" && reference.NormalizedName == "vendor/security" {
			t.Fatalf("wildcard import alias became an external function: %+v", reference)
		}
	}
	module := false
	for _, item := range result.Requirements.Items {
		if item.Kind == "function" {
			t.Fatalf("wildcard import alias became a function requirement: %+v", item)
		}
		module = module || item.Kind == "module" && item.Identity == "vendor/security"
	}
	if !module {
		t.Fatalf("external module obligation was lost: %+v", result.Requirements.Items)
	}
}
