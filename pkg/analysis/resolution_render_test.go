package analysis

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolutionRenderTypedAtoms(t *testing.T) {
	cases := []struct{ name, language, text, marker, kind, value, want string }{
		{"bare dataset", "spl2", `FROM $events | fields id`, "$events", "dataset", "events", `FROM events | fields id`},
		{"quoted dataset", "spl2", `FROM '$events' | fields id`, "$events", "dataset", "event stream", `FROM 'event stream' | fields id`},
		{"quoted selector", "spl", `search index="$events" | fields id`, "$events", "index", "event stream", `search index="event stream" | fields id`},
		{"lookup", "spl", `search index=main | lookup "$people" id OUTPUT name`, "$people", "lookup", "new people", `search index=main | lookup "new people" id OUTPUT name`},
		{"model", "spl2", `tstats aggregates=[count()] datamodel_name='$model'`, "$model", "data_model", "Network", `tstats aggregates=[count()] datamodel_name='Network'`},
		{"unicode", "spl2", `FROM $events | eval note="é雪" | fields note`, "$events", "dataset", "é雪", `FROM 'é雪' | eval note="é雪" | fields note`},
		{"unicode before", "spl", `search source="é雪" index="$events" | fields id`, "$events", "index", "é雪", `search source="é雪" index="é雪" | fields id`},
		{"escape quotes", "spl2", `FROM '$events' | fields id`, "$events", "dataset", "a'b", `FROM 'a\'b' | fields id`},
		{"pipe", "spl2", `FROM $events | fields id`, "$events", "dataset", "a | fields secret", `FROM 'a | fields secret' | fields id`},
		{"repeated", "spl2", `FROM $events | join left=e right=u where e.id=u.id [FROM $events | fields id]`, "$events", "dataset", "events", `FROM events | join left=e right=u where e.id=u.id [FROM events | fields id]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			document := QueryDocument{Text: tc.text, Language: tc.language, Profile: "splunkd", Version: "current", SourceID: "source"}
			s, err := PrepareResolution(document)
			if err != nil {
				t.Fatal(err)
			}
			r, err := s.Render([]ResolutionChoice{{tc.marker, tc.kind, tc.value}})
			if err != nil {
				t.Fatal(err)
			}
			if r.CandidateDocument().Text != tc.want {
				t.Fatalf("got %q want %q limitations=%+v", r.CandidateDocument().Text, tc.want, r.limitations)
			}
			document.Text = tc.want
			if !reflect.DeepEqual(document, r.CandidateDocument()) {
				t.Fatal("metadata changed")
			}
			edits := []RewriteTextEdit{}
			for _, c := range r.Changes() {
				if tc.text[c.OriginalLocation.Start.Offset:c.OriginalLocation.End.Offset] != c.Before {
					t.Fatal("incorrect original audit slice")
				}
				if c.CandidateLocation != nil || len(c.CandidateReferenceIDs) != 0 {
					t.Fatal("unverified candidate correspondence")
				}
				if len(c.OriginalReferenceIDs) == 0 {
					t.Fatal("missing original reference IDs")
				}
				edits = append(edits, RewriteTextEdit{Location: c.OriginalLocation, Before: c.Before, After: c.After})
			}
			if rewriteCandidateText(tc.text, edits) != tc.want {
				t.Fatal("audits do not recreate candidate")
			}
			changes := r.Changes()
			changes[0].After = "mutated"
			changes[0].OriginalReferenceIDs[0] = "mutated"
			if reflect.DeepEqual(changes, r.Changes()) {
				t.Fatal("audits not detached")
			}
		})
	}
}

func TestResolutionRenderSimultaneous(t *testing.T) {
	text := `FROM $left | join left=e right=u where e.id=u.id [FROM $right | fields id] | eval note="$left"`
	s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Render([]ResolutionChoice{{"$right", "dataset", "events"}, {"$left", "dataset", "$right"}})
	if err != nil {
		t.Fatal(err)
	}
	want := `FROM '\u0024right' | join left=e right=u where e.id=u.id [FROM events | fields id] | eval note="$left"`
	if r.CandidateDocument().Text != want {
		t.Fatalf("got %q want %q", r.CandidateDocument().Text, want)
	}
}

func TestResolutionRenderAdmissionAndLimits(t *testing.T) {
	s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: `FROM $events | fields id`})
	if err != nil {
		t.Fatal(err)
	}
	for _, choices := range [][]ResolutionChoice{nil, {{"$missing", "dataset", "events"}}, {{"$events", "index", "events"}}, {{"$events", "dataset", ""}}, {{"$events", "dataset", string([]byte{255})}}, {{"$events", "dataset", "a"}, {"$events", "dataset", "b"}}} {
		if _, err := s.Render(choices); err == nil {
			t.Fatalf("accepted invalid choices %+v", choices)
		}
	}
	var zero ResolutionSession
	if _, err := zero.Render(nil); err == nil {
		t.Fatal("zero session accepted")
	}
	conflict, _ := PrepareResolution(QueryDocument{Language: "spl", Text: `search index="$same" | lookup "$same" id OUTPUT name`})
	if _, err := conflict.Render([]ResolutionChoice{{"$same", "index", "events"}, {"$same", "lookup", "people"}}); err == nil {
		t.Fatal("conflicting kinds accepted")
	}
	limited, _ := PrepareResolution(QueryDocument{Language: "spl", Text: `search index="$events" | fields id`})
	r, err := limited.Render([]ResolutionChoice{{"$events", "index", "events*"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.limitations) == 0 || r.limitations[0].Location.Start.Offset == r.limitations[0].Location.End.Offset || !strings.Contains(r.CandidateDocument().Text, "$events") {
		t.Fatalf("unsafe content not limited: %+v", r)
	}
}

func TestResolutionRenderPreservesOtherRoles(t *testing.T) {
	text := `$base = FROM $events; $consumer = FROM $base AS e | eval note="$events" | fields e.id, note;`
	s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Render([]ResolutionChoice{{"$events", "dataset", "archive"}})
	if err != nil {
		t.Fatal(err)
	}
	want := `$base = FROM archive; $consumer = FROM $base AS e | eval note="$events" | fields e.id, note;`
	if r.CandidateDocument().Text != want || len(r.Changes()) != 1 {
		t.Fatalf("other roles changed: %q %+v", r.CandidateDocument().Text, r.Changes())
	}
	classic := `search index="$events" | eval note="$events" /* $events */ | fields note`
	s, err = PrepareResolution(QueryDocument{Language: "spl", Text: classic})
	if err != nil {
		t.Fatal(err)
	}
	r, err = s.Render([]ResolutionChoice{{"$events", "index", "archive"}})
	if err != nil {
		t.Fatal(err)
	}
	if r.CandidateDocument().Text != strings.Replace(classic, `index="$events"`, `index="archive"`, 1) {
		t.Fatalf("comments or literals changed: %q limitations=%+v", r.CandidateDocument().Text, r.limitations)
	}
}

func TestResolutionRenderRetainsContentLimitation(t *testing.T) {
	text := `search index="$events" source="$source" | fields id`
	s, err := PrepareResolution(QueryDocument{Language: "spl", Text: text})
	if err != nil {
		t.Fatal(err)
	}
	choices := []ResolutionChoice{{"$events", "index", "events*"}, {"$source", "source", "a | b"}}
	r, err := s.Render(choices)
	if err != nil {
		t.Fatal(err)
	}
	if len(r.limitations) != 1 || len(r.choices) != 2 || len(r.Changes()) != 1 {
		t.Fatalf("choice or limitation lost: %+v", r)
	}
	if r.CandidateDocument().Text != `search index="$events" source="a | b" | fields id` {
		t.Fatalf("unsafe choice rewritten: %q", r.CandidateDocument().Text)
	}
	r, err = s.Render([]ResolutionChoice{{"$events", "index", "a\n b"}, {"$source", "source", "source"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.limitations) != 1 {
		t.Fatal("non-renderable newline not limited")
	}
}

func TestResolutionRenderSelfMarkerIsAtomic(t *testing.T) {
	for _, text := range []string{`FROM $events | fields id`, `FROM '$events' | fields id`} {
		s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
		if err != nil {
			t.Fatal(err)
		}
		r, err := s.Render([]ResolutionChoice{{"$events", "dataset", "$events"}})
		if err != nil {
			t.Fatal(err)
		}
		ordinary := rewriteTestSession(t, "spl2", text)
		site := rewriteFind(t, ordinary, "dataset", "$events", 0)
		unchanged, err := ordinary.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteAtom("$events")}})
		if err != nil || len(unchanged.Edits()) != 0 {
			t.Fatalf("ordinary rewrite equal-identity behavior changed: %+v %v", unchanged, err)
		}
		want := `FROM '\u0024events' | fields id`
		if r.CandidateDocument().Text != want {
			t.Fatalf("self marker retained variable syntax: %q", r.CandidateDocument().Text)
		}
		if decoded, ok := spl2DecodeKey(`'\u0024events'`); !ok || decoded != "$events" {
			t.Fatalf("decoded atom=%q exact=%v", decoded, ok)
		}
		changes := r.Changes()
		if len(changes) != 1 {
			t.Fatalf("self marker missing audit: %+v", changes)
		}
		c := changes[0]
		if text[c.OriginalLocation.Start.Offset:c.OriginalLocation.End.Offset] != c.Before {
			t.Fatal("original audit slice mismatch")
		}
		if rewriteCandidateText(text, []RewriteTextEdit{{Location: c.OriginalLocation, Before: c.Before, After: c.After}}) != want {
			t.Fatal("self marker audit does not reconstruct candidate")
		}
	}
}
