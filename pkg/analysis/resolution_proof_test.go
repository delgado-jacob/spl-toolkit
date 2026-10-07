package analysis

import (
	"reflect"
	"testing"
)

func resolutionProofTest(t *testing.T, text string, values ...string) (*ResolutionSession, *ResolutionRendering, *ResolutionProof) {
	t.Helper()
	s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
	if err != nil {
		t.Fatal(err)
	}
	choices := []ResolutionChoice{}
	for i, p := range s.Evidence().Placeholders {
		choices = append(choices, ResolutionChoice{p.Placeholder, p.Kind, values[i]})
	}
	r, err := s.Render(choices)
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Verify(r)
	if err != nil {
		t.Fatal(err)
	}
	return s, r, p
}

func TestResolutionProofRoles(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		values     []string
		roles      int
		collapse   bool
	}{
		{"distinct", `FROM $a AS a JOIN $b AS b ON a.id=b.id | fields a.id`, []string{"events", "users"}, 2, false},
		{"equal", `from $a | join type=inner left=e right=u where e.id=u.id [from $b] | where id=1`, []string{"same", "same"}, 2, true},
		{"unchanged_collision", `FROM $a AS a JOIN events AS b ON a.id=b.id | fields a.id`, []string{"events"}, 2, true},
		{"local_views", `$v = FROM $a | fields id; $consumer = FROM $v | where id=1;`, []string{"events"}, 1, false},
		{"subquery", `FROM $a | append [FROM $b | fields id] | fields id`, []string{"events", "users"}, 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, r, p := resolutionProofTest(t, tc.text, tc.values...)
			evidence := p.Evidence()
			if !evidence.Proven {
				t.Fatalf("proof refused: %+v", evidence.Limitations)
			}
			original, candidate, roles, authorized := p.AssessmentEvidence()
			if !authorized || len(roles) != tc.roles {
				t.Fatalf("roles: %+v", roles)
			}
			ordinary, err := Analyze(r.CandidateDocument())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(candidate.Analysis, *ordinary) {
				t.Fatal("candidate differs from canonical analysis")
			}
			if tc.collapse && roles[0].CandidateInput.ID != roles[1].CandidateInput.ID {
				t.Fatal("expected canonical input collapse")
			}
			seen := map[string]bool{}
			n := 0
			for _, role := range roles {
				for _, pair := range role.Occurrences {
					if seen[pair.CandidateOccurrenceID] {
						t.Fatal("candidate occurrence reassigned")
					}
					seen[pair.CandidateOccurrenceID] = true
					n++
				}
				for _, pair := range role.Requirements {
					if pair.OriginalOccurrence.Necessity != pair.CandidateOccurrence.Necessity {
						t.Fatal("necessity changed")
					}
				}
			}
			want := 0
			for _, input := range original.Analysis.Inputs {
				want += len(input.Occurrences)
			}
			if n != want {
				t.Fatal("source occurrences missing")
			}
			for _, change := range r.Changes() {
				if change.CandidateLocation == nil || len(change.CandidateReferenceIDs) == 0 {
					t.Fatal("candidate edit not audited")
				}
				loc := *change.CandidateLocation
				if r.CandidateDocument().Text[loc.Start.Offset:loc.End.Offset] != change.After {
					t.Fatal("candidate edit coordinates wrong")
				}
			}
		})
	}
}

func TestResolutionProofAuthority(t *testing.T) {
	var zero ResolutionProof
	if _, _, _, ok := zero.AssessmentEvidence(); ok || zero.Evidence().Proven {
		t.Fatal("zero proof authorized")
	}
	s, r, p := resolutionProofTest(t, `FROM $a | fields id`, "events")
	other, err := PrepareResolution(s.Evidence().Analysis.Document)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = other.Verify(r); err == nil {
		t.Fatal("foreign rendering admitted")
	}
	evidence := p.Evidence()
	evidence.Proven = false
	evidence.Roles[0].CandidateInput.Name = "mutated"
	a, b, roles, ok := p.AssessmentEvidence()
	a.Analysis.Document.Text = "mutated"
	b.Analysis.Document.Text = "mutated"
	roles[0].Occurrences[0].CandidateInputID = "mutated"
	if !ok || !p.Evidence().Proven || p.Evidence().Roles[0].CandidateInput.Name != "events" {
		t.Fatal("reporting mutation changed proof")
	}
	freshA, freshB, freshRoles, ok := p.AssessmentEvidence()
	if !ok || freshA.Analysis.Document.Text == "mutated" || freshB.Analysis.Document.Text == "mutated" || freshRoles[0].Occurrences[0].CandidateInputID == "mutated" {
		t.Fatal("assessment alias")
	}
	r.candidate.Text += " | fields injected"
	if _, err = s.Verify(r); err == nil {
		t.Fatal("mutated private candidate admitted")
	}
}

func TestResolutionProofPreservesUnowned(t *testing.T) {
	_, _, p := resolutionProofTest(t, `from $a | join type=inner left=e right=u where e.id=u.id [from $b] | where id=1`, "same", "same")
	original, candidate, roles, ok := p.AssessmentEvidence()
	if !ok {
		t.Fatalf("%+v", p.Evidence())
	}
	var ambiguous RequirementItem
	for _, item := range original.Analysis.Requirements.Items {
		if item.Kind == "field" && item.Identity == "id" && item.InputID == "" {
			ambiguous = item
		}
	}
	if ambiguous.ID == "" {
		t.Fatal("missing original ambiguous field")
	}
	appearedProved := false
	for _, item := range candidate.Analysis.Requirements.Items {
		if item.Kind == "field" && item.Identity == "id" && item.InputID != "" {
			appearedProved = true
		}
	}
	if !appearedProved {
		t.Fatal("control does not exercise canonical owner collapse")
	}
	for _, role := range roles {
		if role.Coverage.State != "partial" {
			t.Fatal("original ambiguity lost")
		}
		for _, pair := range role.Requirements {
			if pair.OriginalRequirementID == ambiguous.ID {
				t.Fatal("unowned requirement assigned to role")
			}
		}
	}
}

func TestResolutionProofRefusedRetainsCanonical(t *testing.T) {
	s, r, p := resolutionProofTest(t, `FROM $a AS a JOIN $b AS b ON a.id=b.id | fields id`, "same", "same")
	original, candidate, roles, authorized := p.AssessmentEvidence()
	if authorized || p.Evidence().Proven || len(p.Evidence().Limitations) == 0 || len(roles) != 0 {
		t.Fatalf("unsafe SQL correspondence authorized: %+v", p.Evidence())
	}
	if original.Analysis.Document.Text != s.Evidence().Analysis.Document.Text || candidate.Analysis.Document != r.CandidateDocument() {
		t.Fatal("refused proof lost canonical snapshots")
	}
	ordinary, err := Analyze(r.CandidateDocument())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(candidate.Analysis, *ordinary) {
		t.Fatal("refused proof changed canonical analysis")
	}
	for _, change := range r.Changes() {
		if change.CandidateLocation != nil || len(change.CandidateReferenceIDs) > 0 {
			t.Fatal("refusal populated unproved edit correspondence")
		}
	}
}

func TestResolutionProofViewDeepCopies(t *testing.T) {
	_, _, proof := resolutionProofTest(t, `$base = FROM $a; $next = FROM $base; $consumer = FROM $next | union [FROM $base];`, "events")
	if !proof.Evidence().Proven {
		t.Fatalf("view proof: %+v", proof.Evidence())
	}
	before := proof.Evidence()
	copy := proof.Evidence()
	found := false
	for i := range copy.Roles {
		for j := range copy.Roles[i].OriginalInput.Occurrences {
			o := &copy.Roles[i].OriginalInput.Occurrences[j]
			if len(o.UseSiteLocations) > 0 {
				o.UseSiteLocations[0].Start.Offset = -1
				o.UseSiteReferenceIDs[0] = "mutated"
				found = true
			}
		}
	}
	if !found {
		t.Fatal("control has no situated use chain")
	}
	if !reflect.DeepEqual(before, proof.Evidence()) {
		t.Fatal("local view chains alias proof")
	}
}

func TestResolutionProofExactRenderControls(t *testing.T) {
	for _, mutation := range []struct {
		name  string
		apply func(*ResolutionRendering)
	}{
		{"effect", func(r *ResolutionRendering) { r.rewrite.effects[0].After = rewriteAtom("different") }},
		{"edit", func(r *ResolutionRendering) { r.rewrite.edits[0].After = "different" }},
		{"metadata", func(r *ResolutionRendering) { r.candidate.SourceID = "different" }},
		{"phase_text", func(r *ResolutionRendering) { r.candidate.Text += " | stats count" }},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			s, r, _ := resolutionProofTest(t, `FROM $a | eval copied=id | stats count(copied) AS n | where n>0`, "events")
			mutation.apply(r)
			if _, err := s.Verify(r); err == nil {
				t.Fatal("mutated rendering acquired authority")
			}
		})
	}
}

func TestResolutionProofClassicSources(t *testing.T) {
	for _, tc := range []struct {
		text    string
		choices []ResolutionChoice
	}{
		{`search index="$a" | fields id`, []ResolutionChoice{{"$a", "index", "events"}}},
		{`search index="$a" source="$s" sourcetype="$t" | lookup "$l" id OUTPUT name`, []ResolutionChoice{{"$a", "index", "events"}, {"$s", "source", "logs"}, {"$t", "sourcetype", "json"}, {"$l", "lookup", "people"}}},
	} {
		s, err := PrepareResolution(QueryDocument{Language: "spl", Text: tc.text})
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.Render(tc.choices)
		if err != nil {
			t.Fatal(err)
		}
		p, err := s.Verify(r)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Evidence().Proven {
			t.Fatalf("classic proof: %+v", p.Evidence())
		}
		_, candidate, _, ok := p.AssessmentEvidence()
		ordinary, err := Analyze(r.CandidateDocument())
		if err != nil {
			t.Fatal(err)
		}
		if !ok || !reflect.DeepEqual(candidate.Analysis, *ordinary) {
			t.Fatal("classic candidate identity drift")
		}
	}
}

func TestResolutionProofOccurrenceNecessity(t *testing.T) {
	_, _, p := resolutionProofTest(t, `$required = from $a | where id=1; $conditional = from seed | branch (guard=true) [from $b | where id=1], (guard=false) [from $b | where other=2];`, "same", "same")
	if !p.Evidence().Proven {
		t.Fatalf("branch correspondence: %+v", p.Evidence())
	}
	_, candidate, roles, ok := p.AssessmentEvidence()
	if !ok {
		t.Fatal("proof unauthorized")
	}
	strongest := map[string]string{}
	for _, item := range candidate.Analysis.Requirements.Items {
		strongest[item.ID] = item.Necessity
	}
	retainedConditional := false
	for _, role := range roles {
		for _, pair := range role.Requirements {
			if strongest[pair.CandidateRequirementID] == "required" && pair.OriginalOccurrence.Necessity == "conditional" {
				if pair.CandidateOccurrence.Necessity != "conditional" {
					t.Fatal("group necessity overwrote occurrence")
				}
				retainedConditional = true
			}
		}
	}
	if !retainedConditional {
		t.Fatalf("control has no conditional occurrence in required group: items %+v roles %+v", candidate.Analysis.Requirements.Items, roles)
	}
}
