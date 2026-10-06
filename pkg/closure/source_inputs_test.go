package closure

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"testing"
)

func TestClosureSourceIntentions(t *testing.T) {
	for _, tc := range []struct {
		text      string
		traverses bool
	}{
		{"from $events | fields id", false},
		{"from '$events' | fields id", true},
		{`from {kind: "index", properties: {name: "main"}} | fields id`, false},
		{"from events | fields id", true},
	} {
		t.Run(tc.text, func(t *testing.T) {
			req := evalRequest(tc.text, []Collection{{Kind: "dataset", Coverage: "complete"}})
			req.Document.Language = "spl2"
			r, err := Evaluate(req)
			if err != nil {
				t.Fatal(err)
			}
			got := false
			for _, e := range r.Traversal {
				if e.Kind == "dataset" {
					got = true
				}
			}
			if got != tc.traverses {
				t.Fatalf("dataset traversal=%v, want %v: %+v inputs=%+v", got, tc.traverses, r.Traversal, r.EffectiveAnalysis.Requirements.Inputs)
			}
			if !tc.traverses && r.EffectiveAnalysis.Requirements.QueryStatus == analysis.Invalid {
				t.Fatal("fixture invalid")
			}
		})
	}
}
