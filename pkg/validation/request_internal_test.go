package validation

import (
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestNormalizeDocumentDoesNotCloneCapabilityManifest(t *testing.T) {
	document := analysis.QueryDocument{Language: "spl2", Profile: "splunkd", Version: "current"}
	if _, err := normalizeDocument(document); err != nil {
		t.Fatal(err)
	}
	allocations := testing.AllocsPerRun(20, func() {
		if _, err := normalizeDocument(document); err != nil {
			t.Fatal(err)
		}
	})
	if allocations > 1 {
		t.Fatalf("selector-only normalization allocated %.0f objects; capability manifest was likely cloned", allocations)
	}
}
