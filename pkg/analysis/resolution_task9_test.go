package analysis

import "testing"

func TestResolutionDiscoveryUnsupportedWholeMarkerIsDiagnostic(t *testing.T) {
	for _, text := range []string{`FROM {kind: $kind, properties: {name: "main"}}`, `FROM $events | append [FROM {kind: $kind, properties: {name: "main"}}]`, `FROM $events | append [FROM {kind: $events, properties: {name: "main"}}]`} {
		s, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
		if err != nil {
			t.Fatal(err)
		}
		e := s.Evidence()
		want := 0
		if text != `FROM {kind: $kind, properties: {name: "main"}}` {
			want = 1
		}
		if len(e.Placeholders) != want || e.Coverage.State != "partial" {
			t.Fatalf("unsupported owner entered admission: text=%s placeholders=%+v coverage=%+v", text, e.Placeholders, e.Coverage)
		}
		choices := []ResolutionChoice{}
		if want == 1 {
			choices = append(choices, ResolutionChoice{"$events", "dataset", "events"})
			if _, err := s.Render(nil); err == nil {
				t.Fatal("missing eligible choice accepted")
			}
		}
		if _, err := s.Render(choices); err != nil {
			t.Fatal(err)
		}
	}
}
