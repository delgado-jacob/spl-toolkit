package compatibility

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestFormatReportCompleteEvidence(t *testing.T) {
	if FormatReport(nil) != "" {
		t.Fatal("nil report should have no display")
	}
	report := checked(t, assessmentFixture(t, "from $events | fields id"))
	text := FormatReport(report)
	if !strings.HasPrefix(text, "Outcome: satisfied\nCorrelation: not applicable\n") || !strings.HasSuffix(text, "\n") {
		t.Fatalf("summary: %s", text)
	}
	_, raw, ok := strings.Cut(text, "Evidence:\n")
	if !ok {
		t.Fatal("complete evidence missing")
	}
	var got, want any
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(report)
	_ = json.Unmarshal(encoded, &want)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("text dropped canonical evidence")
	}
	if text != FormatReport(report) {
		t.Fatal("text is nondeterministic")
	}
}
