package analysis

import (
	"encoding/json"
	"os"
	"testing"
)

// Durable examples exercise the canonical facade, including held forms whose
// original syntax coverage is intentionally incomplete.
func TestRewriteForms(t *testing.T) {
	data, err := os.ReadFile("../../testdata/rewrite/forms.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SchemaVersion int `json:"schema_version"`
		Forms         []struct {
			ID, Language, Query, Kind, Name, Role, Target, Candidate string
			Eligible                                                 bool
		}
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, form := range fixture.Forms {
		t.Run(form.ID, func(t *testing.T) {
			s, err := PrepareRewrite(QueryDocument{Language: form.Language, Text: form.Query}, nil)
			if err != nil {
				t.Fatal(err)
			}
			ordinary, err := Analyze(QueryDocument{Language: form.Language, Text: form.Query})
			if err != nil {
				t.Fatal(err)
			}
			a, _ := json.Marshal(ordinary)
			b, _ := json.Marshal(s.Evidence().Analysis)
			if string(a) != string(b) {
				t.Fatal("ordinary analysis changed")
			}
			site := rewriteFind(t, s, form.Kind, form.Name, 0)
			if site.Role != form.Role || (site.Eligibility == "eligible") != form.Eligible {
				t.Fatalf("form evidence: %+v", site)
			}
			if !form.Eligible {
				if len(site.Limitations) == 0 {
					t.Fatal("missing refusal")
				}
				return
			}
			r, err := s.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName(form.Target)}})
			if err != nil {
				t.Fatal(err)
			}
			candidate := rewriteApply(form.Query, r.Edits())
			if candidate != form.Candidate {
				t.Fatalf("candidate %q; want %q", candidate, form.Candidate)
			}
			next, err := PrepareRewrite(QueryDocument{Language: form.Language, Text: candidate}, nil)
			if err != nil {
				t.Fatal(err)
			}
			if proof := s.Verify(next, r); !proof.Proven {
				t.Fatalf("form correspondence: %+v", proof)
			}
		})
	}
}
