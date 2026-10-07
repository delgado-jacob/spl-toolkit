package analysis

import (
	"reflect"
	"strings"
	"testing"
)

func TestResolutionDiscoveryOwners(t *testing.T) {
	for _, tc := range []struct {
		name, language, text string
		markers              []string
	}{
		{"dataset", "spl2", `FROM $events | fields id`, []string{"$events:dataset"}},
		{"local chains", "spl2", `$base = FROM $events; $next = FROM $base; $consumer = FROM $next | fields id;`, []string{"$events:dataset"}},
		{"repeated subqueries", "spl2", `FROM $events | join left=e right=u where e.id=u.id [FROM $events | fields id]`, []string{"$events:dataset"}},
		{"index", "spl", `search index="$index" | fields id`, []string{"$index:index"}},
		{"lookup", "spl", `search index=main | lookup "$people" id OUTPUT name`, []string{"$people:lookup"}},
		{"model", "spl2", `tstats aggregates=[count()] datamodel_name='$model'`, []string{"$model:data_model"}},
		{"structured canonical", "spl2", `FROM $events | fields payload.foo`, []string{"$events:dataset"}},
		{"literal", "spl2", `FROM main | eval x="$events" | fields x`, []string{}},
		{"embedded", "spl2", `FROM prod_$region | fields id`, []string{}},
		{"interpolation", "spl2", `FROM "${expr}" | fields id`, []string{}},
		{"local shadow", "spl2", `$events = FROM main; $consumer = FROM $events | fields id;`, []string{}},
		{"incompatible kinds", "spl", `search index="$same" | lookup "$same" id OUTPUT name`, []string{"$same:index", "$same:lookup"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			doc := QueryDocument{Language: tc.language, Text: tc.text}
			s, err := PrepareResolution(doc)
			if err != nil {
				t.Fatal(err)
			}
			evidence := s.Evidence()
			got := []string{}
			for _, p := range evidence.Placeholders {
				got = append(got, p.Placeholder+":"+p.Kind)
				if len(p.Locations) == 0 || len(p.ReferenceIDs) == 0 {
					t.Fatalf("missing owner evidence: %+v", p)
				}
			}
			if !reflect.DeepEqual(got, tc.markers) {
				t.Fatalf("markers=%v want=%v analysis=%+v", got, tc.markers, evidence.Analysis)
			}
			plain, err := Analyze(doc)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(evidence.Analysis, *plain) {
				t.Fatalf("canonical analysis differs")
			}
		})
	}
}

func TestResolutionDiscoveryDetachedCanonical(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		text := `search index="$events" | eval x=payload.foo | mystery | fields x`
		if language == "spl2" {
			text = `$base = FROM $events; $next = FROM $base; $consumer = FROM $next | eval x=payload.foo | mystery | fields x;`
		}
		s, err := PrepareResolution(QueryDocument{Text: text, Language: language, Profile: "splunkd", Version: "current"})
		if err != nil {
			t.Fatal(err)
		}
		before := s.Evidence()
		changed := s.Evidence()
		changed.Analysis.Document.Text = "changed"
		if len(changed.Placeholders) > 0 {
			changed.Placeholders[0].ReferenceIDs[0] = "changed"
			changed.Placeholders[0].Locations[0].Start.Offset = -1
			if len(changed.Placeholders[0].OriginalInputIDs) > 0 {
				changed.Placeholders[0].OriginalInputIDs[0] = "changed"
			}
		}
		for i := range changed.Analysis.Inputs {
			for j := range changed.Analysis.Inputs[i].Occurrences {
				o := &changed.Analysis.Inputs[i].Occurrences[j]
				if len(o.UseSiteReferenceIDs) > 0 {
					o.UseSiteReferenceIDs[0] = "changed"
				}
				if len(o.UseSiteLocations) > 0 {
					o.UseSiteLocations[0].Start.Offset = -1
				}
			}
		}
		if len(changed.Coverage.Reasons) > 0 {
			changed.Coverage.Reasons[0].Message = "changed"
		}
		if !reflect.DeepEqual(before, s.Evidence()) {
			t.Fatal("evidence mutation reached session")
		}
		plain, err := Analyze(before.Analysis.Document)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before.Analysis, *plain) {
			t.Fatal("canonical incomplete analysis differs")
		}
		if before.Coverage.State != "partial" {
			t.Fatal("unknown command must preserve discovery gap")
		}
	}
}

func TestResolutionDiscoveryOwnershipAndLimits(t *testing.T) {
	for _, text := range []string{
		`FROM $events | union [FROM $events]`,
		`$base = FROM $events; $next = FROM $base; $consumer = FROM $next | union [FROM $base];`,
	} {
		session, err := PrepareResolution(QueryDocument{Language: "spl2", Text: text})
		if err != nil {
			t.Fatal(err)
		}
		e := session.Evidence()
		if !e.Analysis.Coverage.SyntaxComplete || e.Coverage.State != "complete" || len(e.Placeholders) != 1 {
			t.Fatalf("unexpected discovery: %+v", e)
		}
		p := e.Placeholders[0]
		want := 2
		if text[0] == '$' {
			want = 1
		}
		if len(p.ReferenceIDs) != want || len(p.Locations) != want || len(p.OriginalInputIDs) != 1 || len(session.sites[resolutionGroupKey("$events", "dataset")]) != want {
			t.Fatalf("root ownership lost: %+v", p)
		}
		for _, loc := range p.Locations {
			if text[loc.Start.Offset:loc.End.Offset] != "$events" {
				t.Fatal("local identifier became external site")
			}
		}
	}
	session, err := PrepareResolution(QueryDocument{Language: "spl2", Text: `FROM "${expr}" | fields id`})
	if err != nil {
		t.Fatal(err)
	}
	if session.Evidence().Coverage.State != "partial" || len(session.sites) != 0 {
		t.Fatal("unsupported owner acquired authority or exhaustive coverage")
	}
	for _, doc := range []QueryDocument{{Language: "unsupported"}, {Profile: "unsupported"}, {Version: "unsupported"}} {
		if _, err := PrepareResolution(doc); err == nil {
			t.Fatal("unsupported selector accepted")
		}
	}
	var zero ResolutionSession
	if zero.Evidence().Coverage.State != "partial" {
		t.Fatal("zero session claims exhaustive discovery")
	}
}

func TestResolutionDiscoveryResourceLimit(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		document := QueryDocument{Language: language, Text: strings.Repeat("x ", 5000)}
		session, err := PrepareResolution(document)
		if err != nil {
			t.Fatal(err)
		}
		evidence := session.Evidence()
		if evidence.Coverage.State != "partial" || len(evidence.Placeholders) != 0 || len(session.sites) != 0 {
			t.Fatal("resource failure manufactured exhaustive discovery or authority")
		}
		plain, err := Analyze(document)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(evidence.Analysis, *plain) {
			t.Fatal("resource limited canonical analysis differs")
		}
	}
}
