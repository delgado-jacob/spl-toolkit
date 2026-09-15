package corpus

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestCorpusEvaluationCarriesRequirements(t *testing.T) {
	document := analysis.QueryDocument{Text: "search host=web | table host*", SourceID: "query"}
	report, err := Scan(Request{SchemaVersion: 1, Documents: []RequestDocument{{ID: "query", Document: document}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Entries) != 1 || report.Entries[0].Evaluation == nil || report.Entries[0].Evaluation.Analysis == nil {
		t.Fatalf("missing corpus analysis: %+v", report)
	}
	want, err := analysis.Requirements(document)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(report.Entries[0].Evaluation.Analysis.Requirements, *want) {
		t.Fatalf("corpus requirements differ: got %+v want %+v", report.Entries[0].Evaluation.Analysis.Requirements, *want)
	}
	for _, field := range reflect.VisibleFields(reflect.TypeOf(Report{})) {
		if strings.Contains(strings.ToLower(field.Name), "requirement") {
			t.Fatalf("requirement-specific corpus aggregate introduced: %s", field.Name)
		}
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &root); err != nil {
		t.Fatal(err)
	}
	if _, found := root["requirements"]; found {
		t.Fatalf("requirement aggregate introduced: %s", encoded)
	}
}
