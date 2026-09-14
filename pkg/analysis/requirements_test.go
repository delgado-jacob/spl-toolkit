package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

func TestRequirementTypesJSONShape(t *testing.T) {
	set := RequirementSet{
		SchemaVersion: 1,
		Query: RequirementQueryIdentity{
			SourceID:    "queries/example.spl",
			Language:    "spl",
			Profile:     "splunkd",
			Version:     "current",
			QueryDigest: "sha256:query",
		},
		CapabilityRevision: "sha256:capability",
		QueryStatus:        Incomplete,
		Coverage: RequirementCoverage{
			Complete: false,
			Reasons:  []string{"SPL_REQUIREMENT_DYNAMIC"},
		},
		Items: []RequirementItem{{
			ID:         "req-1",
			Kind:       "field",
			Identity:   "host",
			Role:       "read",
			Necessity:  "required",
			Origin:     "direct",
			Resolution: "exact",
			Occurrences: []RequirementOccurrence{{
				ReferenceID:  "ref-1",
				OriginalName: "Host",
				Binding:      "source",
				StageID:      "stage-0",
				ScopeID:      "scope-0",
				Location: Location{
					Start: Position{Offset: 7, Line: 1, Column: 8},
					End:   Position{Offset: 11, Line: 1, Column: 12},
				},
			}},
		}},
		Gaps: []RequirementGap{{
			Code:            "SPL_REQUIREMENT_DYNAMIC",
			Message:         "dynamic requirement",
			ReferenceIDs:    []string{"ref-1"},
			DiagnosticCodes: []string{"SPL_DYNAMIC_REFERENCE"},
		}},
		Diagnostics: []Diagnostic{{
			Code:     "SPL_DYNAMIC_REFERENCE",
			Severity: "warning",
			Category: "semantic",
			Message:  "dynamic reference",
			Location: Location{
				Start: Position{Offset: 7, Line: 1, Column: 8},
				End:   Position{Offset: 11, Line: 1, Column: 12},
			},
			StageID: "stage-0",
			ScopeID: "scope-0",
		}},
	}

	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var got any
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	const expected = `{"schema_version":1,"query":{"source_id":"queries/example.spl","language":"spl","profile":"splunkd","version":"current","query_digest":"sha256:query"},"capability_revision":"sha256:capability","query_status":"incomplete","coverage":{"complete":false,"reasons":["SPL_REQUIREMENT_DYNAMIC"]},"items":[{"id":"req-1","kind":"field","identity":"host","role":"read","necessity":"required","origin":"direct","resolution":"exact","occurrences":[{"reference_id":"ref-1","original_name":"Host","binding":"source","stage_id":"stage-0","scope_id":"scope-0","location":{"start":{"offset":7,"line":1,"column":8},"end":{"offset":11,"line":1,"column":12}}}]}],"gaps":[{"code":"SPL_REQUIREMENT_DYNAMIC","message":"dynamic requirement","reference_ids":["ref-1"],"diagnostic_codes":["SPL_DYNAMIC_REFERENCE"]}],"diagnostics":[{"code":"SPL_DYNAMIC_REFERENCE","severity":"warning","category":"semantic","message":"dynamic reference","location":{"start":{"offset":7,"line":1,"column":8},"end":{"offset":11,"line":1,"column":12}},"stage_id":"stage-0","scope_id":"scope-0"}]}`
	var want any
	if err := json.Unmarshal([]byte(expected), &want); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requirement JSON shape mismatch:\n got: %s\nwant: %s", encoded, expected)
	}
}

func TestQueryDigestExactBytes(t *testing.T) {
	for _, text := range []string{"", "café 😀", "a\nb", "a\r\nb"} {
		sum := sha256.Sum256([]byte(text))
		want := "sha256:" + hex.EncodeToString(sum[:])
		if got := queryDigest(text); got != want {
			t.Errorf("queryDigest(%q) = %q, want %q", text, got, want)
		}
	}

	left := QueryDocument{Text: "search café=*", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "left.spl"}
	right := QueryDocument{Text: left.Text, Language: "spl2", Profile: "unused", Version: "future", SourceID: "right.spl"}
	if queryDigest(left.Text) != queryDigest(right.Text) {
		t.Fatal("selector-only changes affected the query digest")
	}
}

func TestCapabilityRevisionUsesNormalizedTypedManifest(t *testing.T) {
	defaultDocument := QueryDocument{}
	explicitDocument := QueryDocument{Language: "spl", Profile: "splunkd", Version: "current"}

	defaultRevision, err := capabilityRevision(defaultDocument)
	if err != nil {
		t.Fatal(err)
	}
	explicitRevision, err := capabilityRevision(explicitDocument)
	if err != nil {
		t.Fatal(err)
	}
	if defaultRevision != explicitRevision {
		t.Fatalf("default revision %q differs from normalized explicit revision %q", defaultRevision, explicitRevision)
	}
	repeatedRevision, err := capabilityRevision(defaultDocument)
	if err != nil {
		t.Fatal(err)
	}
	if repeatedRevision != defaultRevision {
		t.Fatalf("repeated revision %q differs from first revision %q", repeatedRevision, defaultRevision)
	}

	spl2Document := QueryDocument{Language: "spl2", Profile: "splunkd", Version: "current"}
	spl2Revision, err := capabilityRevision(spl2Document)
	if err != nil {
		t.Fatal(err)
	}
	if spl2Revision == defaultRevision {
		t.Fatalf("SPL and SPL2 capability revisions are both %q", defaultRevision)
	}

	for _, document := range []QueryDocument{defaultDocument, explicitDocument, spl2Document} {
		manifest, err := CapabilitiesFor(CapabilityOptions{
			Language: document.Language,
			Profile:  document.Profile,
			Version:  document.Version,
		})
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(encoded)
		want := "sha256:" + hex.EncodeToString(sum[:])
		got, err := capabilityRevision(document)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("capabilityRevision(%+v) = %q, want %q", document, got, want)
		}
	}

	if got, err := capabilityRevision(QueryDocument{Language: "SPL"}); err == nil || got != "" {
		t.Fatalf("unsupported selectors returned revision %q and error %v", got, err)
	}
}

func TestCloneRequirementSetOwnsNestedSlices(t *testing.T) {
	source := RequirementSet{
		Coverage: RequirementCoverage{Reasons: []string{"reason"}},
		Items: []RequirementItem{{
			ID:          "req-1",
			Occurrences: []RequirementOccurrence{{ReferenceID: "ref-1"}},
		}},
		Gaps: []RequirementGap{{
			Code:            "gap",
			ReferenceIDs:    []string{"ref-1"},
			DiagnosticCodes: []string{"diagnostic"},
		}},
		Diagnostics: []Diagnostic{{Code: "diagnostic"}},
	}
	before, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}

	cloned := cloneRequirementSet(source)
	cloned.Coverage.Reasons[0] = "changed reason"
	cloned.Items[0].ID = "req-changed"
	cloned.Items[0].Occurrences[0].ReferenceID = "ref-changed"
	cloned.Gaps[0].Code = "changed gap"
	cloned.Gaps[0].ReferenceIDs[0] = "ref-changed"
	cloned.Gaps[0].DiagnosticCodes[0] = "changed diagnostic link"
	cloned.Diagnostics[0].Code = "changed diagnostic"

	after, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("clone mutations changed source:\n before: %s\n  after: %s", before, after)
	}
}

func TestRequirementSetEmptyCollectionsAreArrays(t *testing.T) {
	set := RequirementSet{
		Coverage:    RequirementCoverage{Reasons: []string{}},
		Items:       []RequirementItem{},
		Gaps:        []RequirementGap{},
		Diagnostics: []Diagnostic{},
	}
	encoded, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}

	coverage, ok := decoded["coverage"].(map[string]any)
	if !ok {
		t.Fatalf("coverage is %T, want object", decoded["coverage"])
	}
	for name, value := range map[string]any{
		"coverage.reasons": coverage["reasons"],
		"items":            decoded["items"],
		"gaps":             decoded["gaps"],
		"diagnostics":      decoded["diagnostics"],
	} {
		items, ok := value.([]any)
		if !ok || len(items) != 0 {
			t.Errorf("%s = %#v (%T), want []", name, value, value)
		}
	}
}

func TestRequirementDiagnosticCodes(t *testing.T) {
	want := []string{
		"SPL_REQUIREMENT_INDETERMINATE",
		"SPL_REQUIREMENT_DYNAMIC",
		"SPL_REQUIREMENT_COVERAGE_INCOMPLETE",
	}
	got := []string{
		CodeRequirementIndeterminate,
		CodeRequirementDynamic,
		CodeRequirementCoverageIncomplete,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("requirement diagnostic codes = %q, want %q", got, want)
	}
}
