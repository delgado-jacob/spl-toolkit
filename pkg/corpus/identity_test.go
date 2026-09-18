package corpus

import (
	"testing"

	"github.com/delgado-jacob/spl-toolkit/internal/buildinfo"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestAnalysisRevisionMatchesCanonicalManifestPayload(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		document := analysis.QueryDocument{Text: "search user=alice", Language: language, Profile: "splunkd", Version: "current"}
		manifest, err := analysis.CapabilitiesFor(analysis.CapabilityOptions{Language: language, Profile: "splunkd", Version: "current"})
		if err != nil {
			t.Fatal(err)
		}
		want, err := digestJSON([]any{"analysis-revision-v1", SourceHash(document.Text), manifest.Language, manifest.Profile, manifest.Version, buildinfo.Version, "analysis-report-v1", manifest})
		if err != nil {
			t.Fatal(err)
		}
		got, err := AnalysisRevision(document)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%s analysis revision = %q, want %q", language, got, want)
		}
	}
}

func TestAnalysisRevisionDoesNotCloneCapabilityManifest(t *testing.T) {
	document := analysis.QueryDocument{Text: "search user=alice", Language: "spl2", Profile: "splunkd", Version: "current"}
	if _, err := AnalysisRevision(document); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(20, func() {
		if _, err := AnalysisRevision(document); err != nil {
			panic(err)
		}
	})
	if allocations > 50 {
		t.Fatalf("AnalysisRevision allocated %.0f objects; capability manifests must not be cloned per document", allocations)
	}
}

func TestRevisionIncludesTargetContent(t *testing.T) {
	doc := analysis.QueryDocument{Text: "search user=alice"}
	base, err := AnalysisRevision(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, changed := range []analysis.QueryDocument{
		{Text: "search user=bob"},
		{Text: doc.Text, Language: "spl2"},
	} {
		got, err := AnalysisRevision(changed)
		if err != nil || got == base {
			t.Fatalf("revision did not change: %q, %v", got, err)
		}
	}
	defaulted, err := AnalysisRevision(analysis.QueryDocument{Text: doc.Text, Language: "spl", Profile: "splunkd", Version: "current"})
	if err != nil || defaulted != base {
		t.Fatalf("normalized selector revision: %q, %v", defaulted, err)
	}

	left := []byte(`{"kind":"json_schema","identity":"same","schema":{"type":"object","properties":{"user":{"type":"string"}}}}`)
	right := []byte(`{"kind":"json_schema","identity":"same","schema":{"type":"object","properties":{"user":{"type":"number"}}}}`)
	a, err := DecodeValidationTarget(left)
	if err != nil {
		t.Fatal(err)
	}
	b, err := DecodeValidationTarget(right)
	if err != nil {
		t.Fatal(err)
	}
	ad, err := TargetDigest(a)
	if err != nil {
		t.Fatal(err)
	}
	bd, err := TargetDigest(b)
	if err != nil || ad == bd {
		t.Fatalf("content change retained digest: %q, %q, %v", ad, bd, err)
	}
	if SourceHash("a\r\n") == SourceHash("a\n") {
		t.Fatal("source hash normalized line endings")
	}
	if got := SourceHash("a\r\n"); got != "8e4621379786ef42a4fec155cd525c291dd7db3c1fde3478522f4f61c03fd1bd" {
		t.Fatalf("source hash is not SHA-256 of exact bytes: %s", got)
	}
	firstSelection, err := DecodeValidationTarget([]byte(`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"a","profiles":["z","b"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	sameSelection, err := DecodeValidationTarget([]byte(`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"a","profiles":["b","z"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	otherSelection, err := DecodeValidationTarget([]byte(`{"kind":"ocsf","catalog":{},"selection":{"version":"1","class":"b","profiles":["b","z"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	firstDigest, err := TargetDigest(firstSelection)
	if err != nil {
		t.Fatal(err)
	}
	sameDigest, err := TargetDigest(sameSelection)
	if err != nil || sameDigest != firstDigest {
		t.Fatalf("normalized selection changed digest: %q, %q, %v", firstDigest, sameDigest, err)
	}
	otherDigest, err := TargetDigest(otherSelection)
	if err != nil || otherDigest == firstDigest {
		t.Fatalf("changed selection retained digest: %q, %q, %v", firstDigest, otherDigest, err)
	}
	resourceA, err := DecodeValidationTarget([]byte(`{"kind":"json_schema","identity":"same","schema":{"$ref":"https://example.test/child"},"resources":{"https://example.test/child":{"type":"string"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	resourceB, err := DecodeValidationTarget([]byte(`{"kind":"json_schema","identity":"same","schema":{"$ref":"https://example.test/child"},"resources":{"https://example.test/child":{"type":"number"}}}`))
	if err != nil {
		t.Fatal(err)
	}
	resourceDigestA, err := TargetDigest(resourceA)
	if err != nil {
		t.Fatal(err)
	}
	resourceDigestB, err := TargetDigest(resourceB)
	if err != nil || resourceDigestA == resourceDigestB {
		t.Fatalf("changed local resource retained digest: %q, %q, %v", resourceDigestA, resourceDigestB, err)
	}
}
