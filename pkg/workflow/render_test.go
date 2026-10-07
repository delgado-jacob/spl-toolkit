package workflow

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAssessmentRenderingPreservesCanonicalReport(t *testing.T) {
	for _, format := range []string{"", "json", "text", "sarif", "graph", "bom"} {
		t.Run(format, func(t *testing.T) {
			request := seedRequest(t, "from [{id:1}]")
			request.Format = format
			before, _ := json.Marshal(request)
			payload, report, err := AssessOutput(request)
			if err != nil {
				t.Fatal(err)
			}
			if report.CIExitCode != 0 {
				t.Fatalf("CI %d", report.CIExitCode)
			}
			canonical, err := Assess(request)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(report, canonical) {
				t.Fatal("rendering changed canonical report")
			}
			after, _ := json.Marshal(request)
			if string(before) != string(after) {
				t.Fatal("mutated request")
			}
			if _, err = json.Marshal(payload); err != nil {
				t.Fatal(err)
			}
			wire, err := AssessOutputJSON(before)
			if err != nil || !reflect.DeepEqual(payload, wire) {
				t.Fatalf("wire parity: %v", err)
			}
			if format == "text" {
				if _, ok := payload.(string); !ok {
					t.Fatal("text must be a string")
				}
			}
		})
	}
}
func TestAssessmentOutputRejectsFormatBeforeEvaluation(t *testing.T) {
	_, report, err := AssessOutput(Request{Format: "unsupported"})
	detail, ok := RequestErrorDetails(err)
	if report != nil || !ok || detail.Path != "/format" {
		t.Fatalf("report=%v detail=%+v err=%v", report, detail, err)
	}
}

func TestRenderAssessmentJSONDetachesReport(t *testing.T) {
	report, err := Assess(seedRequest(t, "from [{id:1}]"))
	if err != nil {
		t.Fatal(err)
	}
	value, err := RenderAssessment(report, "json")
	if err != nil {
		t.Fatal(err)
	}
	rendered := value.(*Report)
	rendered.Entries[0].ID = "changed"
	rendered.Entries[0].Analysis.Document.Text = "changed"
	if report.Entries[0].ID == "changed" || report.Entries[0].Analysis.Document.Text == "changed" {
		t.Fatal("JSON projection aliases source")
	}
	for _, format := range []string{"json", "text", "sarif", "graph", "bom"} {
		if _, err = RenderAssessment(nil, format); err == nil {
			t.Fatalf("%s accepted missing report", format)
		}
	}
}
