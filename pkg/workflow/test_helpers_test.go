package workflow

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
	"os"
	"testing"
)

func seedRequest(t *testing.T, text string) Request {
	t.Helper()
	raw, err := os.ReadFile("../../examples/resolution/request.json")
	if err != nil {
		t.Fatal(err)
	}
	seed, err := resolution.DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	return Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: "d1", Document: analysis.QueryDocument{Text: text, Language: "spl2", Profile: "splunkd", Version: "current", SourceID: "d1"}}}, Settings: Settings{SchemaVersion: 1, Snapshot: seed.Compatibility.Snapshot, SchemaBundle: seed.Compatibility.SchemaBundle, Entries: []EntrySettings{{ID: "d1", Compatibility: &CheckSettings{QueryScope: seed.Compatibility.QueryScope, InputBindings: []compatibility.InputBinding{}}}}}}
}
