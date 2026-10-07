package analysis_test

import (
	"encoding/json"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestResolutionProofExternalAuthority(t *testing.T) {
	evidence := analysis.ResolutionProofEvidence{Proven: true, Roles: []analysis.ResolutionRole{{OriginalInput: analysis.QueryInput{ID: "forged"}}}}
	raw, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	var proof analysis.ResolutionProof
	if err = json.Unmarshal(raw, &proof); err != nil {
		t.Fatal(err)
	}
	if _, _, _, authorized := proof.AssessmentEvidence(); authorized || proof.Evidence().Proven {
		t.Fatal("serialized reporting acquired authority")
	}
	var missing *analysis.ResolutionProof
	if _, _, _, authorized := missing.AssessmentEvidence(); authorized || missing.Evidence().Proven {
		t.Fatal("nil proof authorized")
	}
	s, err := analysis.PrepareResolution(analysis.QueryDocument{Language: "spl2", Text: `FROM $a | fields id`})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Render([]analysis.ResolutionChoice{{Placeholder: "$a", Kind: "dataset", Value: "events"}})
	if err != nil {
		t.Fatal(err)
	}
	document := r.CandidateDocument()
	document.Text += " | where injected=1"
	changes := r.Changes()
	changes[0].After = "injected"
	legitimate, err := s.Verify(r)
	if err != nil {
		t.Fatal(err)
	}
	if !legitimate.Evidence().Proven || r.CandidateDocument() == document || r.Changes()[0].After == "injected" {
		t.Fatal("detached reporting mutated authority")
	}
}
