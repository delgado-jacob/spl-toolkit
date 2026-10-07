package resolution

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func requestFixture(t *testing.T) Request {
	t.Helper()
	raw, err := os.ReadFile("../../testdata/resolution/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Document      analysis.QueryDocument `json:"document"`
		Compatibility CompatibilityInputs    `json:"compatibility"`
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	fixtures[0].Compatibility.InputBindings[0].OriginalInputID = "original-events"
	return Request{SchemaVersion: 1, Document: fixtures[0].Document, Resolutions: []Resolution{}, Compatibility: fixtures[0].Compatibility}
}
func TestResolutionRequestTyped(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Request)
	}{
		{"nil resolutions", func(r *Request) { r.Resolutions = nil }},
		{"nil bindings", func(r *Request) { r.Compatibility.InputBindings = nil }},
		{"nil artifact array", func(r *Request) { r.Compatibility.Snapshot.Objects = nil }},
		{"schema version", func(r *Request) { r.SchemaVersion = 2 }},
		{"zero limit", func(r *Request) { r.MaxVariants = uint64Pointer(0) }},
		{"bad marker", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "events", Kind: "dataset", Values: []string{"a"}}}
		}},
		{"kind", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "macro", Values: []string{"a"}}}
		}},
		{"nil values", func(r *Request) { r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset"}} }},
		{"empty values", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{}}}
		}},
		{"blank value", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{""}}}
		}},
		{"duplicate values", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{"a", "a"}}}
		}},
		{"duplicate marker", func(r *Request) {
			r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{"a"}}, {Placeholder: "$events", Kind: "dataset", Values: []string{"b"}}}
		}},
		{"invalid UTF8", func(r *Request) { r.Document.Text = string([]byte{255}) }},
		{"bad binding", func(r *Request) { r.Compatibility.InputBindings[0].ObjectID = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := requestFixture(t)
			tc.mutate(&r)
			if _, err := normalizeRequest(r); err == nil {
				t.Fatal("expected rejection")
			}
		})
	}
	r := requestFixture(t)
	out, err := normalizeRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	out.Compatibility.InputBindings[0].ObjectID = "changed"
	*out.Compatibility.InputBindings[0].ResolvedValue = "detached"
	if *r.Compatibility.InputBindings[0].ResolvedValue == "detached" {
		t.Fatal("binding pointer aliases original")
	}
	if out.Compatibility.SchemaBundle != nil {
		out.Compatibility.SchemaBundle.Schemas[0].Catalog[0] = 32
		if out.Compatibility.SchemaBundle.Schemas[0].Catalog[0] == r.Compatibility.SchemaBundle.Schemas[0].Catalog[0] {
			t.Fatal("raw schema aliases original")
		}
	}
	if r.Compatibility.InputBindings[0].ObjectID == "changed" {
		t.Fatal("binding aliases original")
	}
}
func TestResolutionRequestStrictJSON(t *testing.T) {
	r := requestFixture(t)
	raw, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeRequest(raw); err != nil {
		t.Fatal(err)
	}
	base := string(raw)
	for _, tc := range []struct{ name, body string }{
		{"unknown", strings.Replace(base, `"schema_version":1`, `"unknown":1,"schema_version":1`, 1)},
		{"duplicate", strings.Replace(base, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1)},
		{"missing", strings.Replace(base, `"resolutions":[],`, "", 1)},
		{"null", strings.Replace(base, `"resolutions":[]`, `"resolutions":null`, 1)},
		{"optional null", strings.Replace(base, `"schema_version":1`, `"max_variants":null,"schema_version":1`, 1)},
		{"nested unknown", strings.Replace(base, `"input_bindings":`, `"wat":1,"input_bindings":`, 1)},
		{"nested duplicate", strings.Replace(base, `"input_bindings":`, `"input_bindings":[],"input_bindings":`, 1)},
		{"binding unknown", strings.Replace(base, `"original_input_id":`, `"wat":1,"original_input_id":`, 1)},
		{"trailing", base + ` {}`},
		{"surrogate", strings.Replace(base, `"text":`, `"text":"\ud800","other":`, 1)},
		{"invalid UTF8", base + string([]byte{255})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecodeRequest([]byte(tc.body))
			if err == nil {
				t.Fatal("expected rejection")
			}
			d, ok := RequestErrorDetails(err)
			if !ok || d.Code != "request_invalid" {
				t.Fatalf("detail: %+v %v", d, err)
			}
		})
	}
	for _, limit := range []string{"0", "-1", "1.5", "1e2", "18446744073709551616", `"2"`} {
		t.Run(limit, func(t *testing.T) {
			body := strings.Replace(base, `"schema_version":1`, `"max_variants":`+limit+`,"schema_version":1`, 1)
			if _, err := DecodeRequest([]byte(body)); err == nil {
				t.Fatal("expected integer rejection")
			}
		})
	}
}

func TestResolutionRequestNestedShapeAndOffsets(t *testing.T) {
	r := requestFixture(t)
	r.Resolutions = []Resolution{{Placeholder: "$events", Kind: "dataset", Values: []string{"events"}}}
	raw, _ := json.Marshal(r)
	base := string(raw)
	for _, body := range []string{
		strings.Replace(base, `"placeholder":"$events"`, `"placeholder":"$events","unknown":true`, 1),
		strings.Replace(base, `"placeholder":"$events"`, `"placeholder":"$events","placeholder":"$other"`, 1),
		strings.Replace(base, `"values":["events"]`, `"values":null`, 1),
		strings.Replace(base, `"snapshot":{`, `"snapshot":null,"unused":{`, 1),
		strings.Replace(base, `"original_input_id":"original-events"`, `"original_input_id":"original-events","original_input_id":"duplicate"`, 1),
	} {
		if _, err := DecodeRequest([]byte(body)); err == nil {
			t.Fatal("nested shape accepted")
		}
	}
	body := strings.Replace(base, `"placeholder":"$events"`, `"placeholder":"$events","placeholder":"$other"`, 1)
	_, err := DecodeRequest([]byte(body))
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.ByteOffset == nil || *detail.ByteOffset == 0 {
		t.Fatalf("missing duplicate offset: %+v", detail)
	}
	*detail.ByteOffset = 0
	fresh, _ := RequestErrorDetails(err)
	if *fresh.ByteOffset == 0 {
		t.Fatal("error offset aliases")
	}
}

func TestResolutionRequestChoiceKinds(t *testing.T) {
	for _, kind := range []string{"dataset", "index", "source", "sourcetype", "lookup", "data_model"} {
		t.Run(kind, func(t *testing.T) {
			request := requestFixture(t)
			request.Resolutions = []Resolution{{Placeholder: "$target", Kind: kind, Values: []string{"selected"}}}
			if _, err := normalizeRequest(request); err != nil {
				t.Fatalf("typed choice rejected: %v", err)
			}
			raw, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeRequest(raw); err != nil {
				t.Fatalf("wire choice rejected: %v", err)
			}
		})
	}
	request := requestFixture(t)
	request.Resolutions = []Resolution{{Placeholder: "$target", Kind: "field", Values: []string{"selected"}}}
	if _, err := normalizeRequest(request); err == nil {
		t.Fatal("field choice admitted")
	}
	raw, _ := json.Marshal(request)
	if _, err := DecodeRequest(raw); err == nil {
		t.Fatal("wire field choice admitted")
	}
	for _, kind := range []string{"lookup", "data_model"} {
		request := requestFixture(t)
		request.Compatibility.InputBindings[0].Expected.Kind = kind
		if _, err := normalizeRequest(request); err == nil {
			t.Fatalf("%s binding identity admitted", kind)
		}
		raw, _ := json.Marshal(request)
		if _, err := DecodeRequest(raw); err == nil {
			t.Fatalf("wire %s binding identity admitted", kind)
		}
	}
}
