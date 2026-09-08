package rewrite

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

const testRule = `{"id":"rename","kind":"field","source":{"name":"src"},"target":{"name":"user"}}`

func testRequest(rule string) string {
	return `{"schema_version":1,"document":{"text":"search src=alice | table src"},"rules":[` + rule + `]}`
}

func TestDecodeRewriteRequest(t *testing.T) {
	cases := []struct {
		name, raw string
		valid     bool
	}{
		{"defaults", testRequest(testRule), true},
		{"empty rules", testRequest(""), true},
		{"self mapping", testRequest(strings.Replace(testRule, `"user"`, `"src"`, 1)), true},
		{"duplicate semantic rules", testRequest(testRule + "," + strings.Replace(testRule, `"rename"`, `"second"`, 1)), true},
		{"meaningful whitespace", testRequest(strings.Replace(testRule, `"src"`, `" src "`, 1)), true},
		{"field path", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"path":["actor","name"]}`, 1)), true},
		{"paired surrogate", testRequest(strings.Replace(testRule, `"src"`, `"\ud83d\ude00"`, 1)), true},
		{"missing version", `{"document":{"text":""},"rules":[]}`, false},
		{"null version", `{"schema_version":null,"document":{"text":""},"rules":[]}`, false},
		{"fractional version", `{"schema_version":1.0,"document":{"text":""},"rules":[]}`, false},
		{"unsupported version", `{"schema_version":2,"document":{"text":""},"rules":[]}`, false},
		{"missing document", `{"schema_version":1,"rules":[]}`, false},
		{"missing rules", `{"schema_version":1,"document":{"text":""}}`, false},
		{"null rules", `{"schema_version":1,"document":{"text":""},"rules":null}`, false},
		{"null document", `{"schema_version":1,"document":null,"rules":[]}`, false},
		{"missing text", `{"schema_version":1,"document":{},"rules":[]}`, false},
		{"null text", `{"schema_version":1,"document":{"text":null},"rules":[]}`, false},
		{"unknown language", `{"schema_version":1,"document":{"text":"","language":"sql"},"rules":[]}`, false},
		{"unknown profile", `{"schema_version":1,"document":{"text":"","profile":"edge"},"rules":[]}`, false},
		{"unknown version", `{"schema_version":1,"document":{"text":"","version":"next"},"rules":[]}`, false},
		{"null option", `{"schema_version":1,"document":{"text":"","source_id":null},"rules":[]}`, false},
		{"extra document key", `{"schema_version":1,"document":{"text":"","options":{}},"rules":[]}`, false},
		{"duplicate document key", `{"schema_version":1,"document":{"text":"","text":"x"},"rules":[]}`, false},
		{"null mode", `{"schema_version":1,"mode":null,"document":{"text":""},"rules":[]}`, false},
		{"empty mode", `{"schema_version":1,"mode":"","document":{"text":""},"rules":[]}`, false},
		{"unknown mode", `{"schema_version":1,"mode":"commit","document":{"text":""},"rules":[]}`, false},
		{"extra key", `{"schema_version":1,"document":{"text":""},"rules":[],"context":{}}`, false},
		{"duplicate key", `{"schema_version":1,"schema_version":1,"document":{"text":""},"rules":[]}`, false},
		{"duplicate IDs", testRequest(testRule + "," + testRule), false},
		{"empty ID", testRequest(strings.Replace(testRule, `"rename"`, `" "`, 1)), false},
		{"null rule", testRequest("null"), false},
		{"missing source", testRequest(`{"id":"x","kind":"field","target":{"name":"x"}}`), false},
		{"null source", testRequest(strings.Replace(testRule, `{"name":"src"}`, `null`, 1)), false},
		{"null when", testRequest(strings.TrimSuffix(testRule, "}") + `,"when":null}`), false},
		{"unknown rule key", testRequest(strings.TrimSuffix(testRule, "}") + `,"priority":1}`), false},
		{"unknown kind", testRequest(strings.Replace(testRule, `"field"`, `"macro"`, 1)), false},
		{"both identities", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"name":"src","path":["src"]}`, 1)), false},
		{"missing identity", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{}`, 1)), false},
		{"null name", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"name":null}`, 1)), false},
		{"blank name", testRequest(strings.Replace(testRule, `"src"`, `" \t "`, 1)), false},
		{"duplicate identity key", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"name":"src","name":"src"}`, 1)), false},
		{"unknown identity key", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"name":"src","quoted":true}`, 1)), false},
		{"empty path", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"path":[]}`, 1)), false},
		{"null path", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"path":null}`, 1)), false},
		{"null segment", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"path":[null]}`, 1)), false},
		{"blank segment", testRequest(strings.Replace(testRule, `{"name":"src"}`, `{"path":["actor"," "]}`, 1)), false},
		{"nonfield path", testRequest(`{"id":"x","kind":"index","source":{"path":["a"]},"target":{"name":"b"}}`), false},
		{"lone surrogate", testRequest(strings.Replace(testRule, `"src"`, `"\ud800"`, 1)), false},
		{"invalid UTF8", testRequest(strings.Replace(testRule, "src", string([]byte{0xff}), 1)), false},
		{"trailing value", testRequest(testRule) + " {}", false},
		{"truncated", `{"schema_version":1`, false},
		{"array root", `[]`, false},
		{"null root", `null`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := DecodeRequest([]byte(tc.raw))
			if !tc.valid {
				if err == nil || !IsInputError(err) {
					t.Fatalf("want input error, got %#v, %v", r, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if r.SchemaVersion != 1 || r.Mode != Preview || r.Document.Language != "spl" || r.Document.Profile != "splunkd" || r.Document.Version != "current" || r.Rules == nil {
				t.Fatalf("bad defaults: %#v", r)
			}
		})
	}
	for _, kind := range []string{"field", "index", "source", "sourcetype", "lookup", "dataset", "data_model"} {
		if _, err := DecodeRequest([]byte(testRequest(strings.Replace(testRule, `"field"`, `"`+kind+`"`, 1)))); err != nil {
			t.Errorf("kind %s: %v", kind, err)
		}
	}
	r, err := DecodeRequest([]byte(`{"schema_version":1,"mode":"apply","document":{"text":"from main\r\n| select café","language":"spl2","source_id":" input "},"rules":[{"id":"x","kind":"field","source":{"name":"actor.name"},"target":{"path":[" actor ","Name"]}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if r.Mode != Apply || r.Document.Text != "from main\r\n| select café" || r.Document.SourceID != " input " || *r.Rules[0].Source.Name != "actor.name" || !reflect.DeepEqual(r.Rules[0].Target.Path, []string{" actor ", "Name"}) {
		t.Fatalf("lost identity/text: %#v", r)
	}
}

func TestDecodeRewriteConditions(t *testing.T) {
	leaf := `{"fact":"literal","kind":"field","identity":{"name":"src"},"operator":"equals","value":`
	cases := []struct {
		name, when string
		valid      bool
	}{
		{"string", leaf + `"1"}`, true}, {"number", leaf + `1}`, true}, {"boolean", leaf + `true}`, true}, {"null", leaf + `null}`, true},
		{"exact integer one", leaf + `9007199254740992}`, true}, {"exact integer two", leaf + `9007199254740993}`, true},
		{"large exponent", leaf + `1e10000}`, true},
		{"contains", strings.Replace(leaf, "equals", "contains", 1) + `"A"}`, true},
		{"reference", `{"fact":"source_reference_present","kind":"field","identity":{"path":["a","b"]}}`, true},
		{"all", `{"all":[` + leaf + `null},{"any":[` + leaf + `true},` + leaf + `false}]}]}`, true},
		{"no union", `{}`, false}, {"mixed union", `{"all":[` + leaf + `1}],"any":[` + leaf + `1}]}`, false},
		{"empty all", `{"all":[]}`, false}, {"null any", `{"any":null}`, false}, {"null child", `{"all":[null]}`, false},
		{"extra combinator key", `{"all":[` + leaf + `1}],"context":{}}`, false},
		{"missing scalar", strings.TrimSuffix(leaf, `,"value":`) + `}`, false},
		{"array scalar", leaf + `[]}`, false}, {"object scalar", leaf + `{}}`, false},
		{"duplicate scalar", leaf + `1,"value":2}`, false},
		{"contains number", strings.Replace(leaf, "equals", "contains", 1) + `1}`, false},
		{"contains null", strings.Replace(leaf, "equals", "contains", 1) + `null}`, false},
		{"regex", strings.Replace(leaf, "equals", "regex", 1) + `".*"}`, false},
		{"dependency path", strings.Replace(leaf, `"kind":"field","identity":{"name":"src"}`, `"kind":"index","identity":{"path":["a"]}`, 1) + `1}`, false},
		{"literal lookup", strings.Replace(leaf, `"kind":"field"`, `"kind":"lookup"`, 1) + `1}`, false},
		{"reference wrong kind", `{"fact":"source_reference_present","kind":"index","identity":{"name":"a"}}`, false},
		{"reference value", `{"fact":"source_reference_present","kind":"field","identity":{"name":"a"},"value":null}`, false},
		{"reference operator", `{"fact":"source_reference_present","kind":"field","identity":{"name":"a"},"operator":"equals"}`, false},
		{"not", `{"not":` + leaf + `1}}`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := testRequest(strings.TrimSuffix(testRule, "}") + `,"when":` + tc.when + `}`)
			r, err := DecodeRequest([]byte(raw))
			if !tc.valid {
				if err == nil || !IsInputError(err) {
					t.Fatalf("want input error: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(tc.name, "exact integer") || tc.name == "null" {
				want := strings.TrimSuffix(strings.TrimPrefix(tc.when, leaf), "}")
				if string(r.Rules[0].When.Value) != want {
					t.Fatalf("scalar changed: %s want %s", r.Rules[0].When.Value, want)
				}
			}
		})
	}
}

func TestDecodeRewriteBatchRequest(t *testing.T) {
	for _, tc := range []struct {
		name, raw string
		valid     bool
	}{
		{"ordered", `{"schema_version":1,"documents":[{"text":"first","source_id":"a"},{"text":"second","language":"spl2","source_id":"b"}],"rules":[]}`, true},
		{"empty", `{"schema_version":1,"documents":[],"rules":[]}`, false},
		{"null", `{"schema_version":1,"documents":null,"rules":[]}`, false},
		{"missing", `{"schema_version":1,"rules":[]}`, false},
		{"both document forms", `{"schema_version":1,"documents":[{"text":""}],"document":{"text":""},"rules":[]}`, false},
		{"later malformed", `{"schema_version":1,"documents":[{"text":"first"},{"text":"later","language":"sql"}],"rules":[]}`, false},
		{"duplicate rules", `{"schema_version":1,"documents":[{"text":""}],"rules":[],"rules":[]}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := DecodeBatchRequest([]byte(tc.raw))
			if !tc.valid {
				if err == nil || !IsInputError(err) {
					t.Fatalf("want input error: %v", err)
				}
				if len(r.Documents) != 0 {
					t.Fatal("published partial batch")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if r.Mode != Preview || len(r.Documents) != 2 || r.Documents[0].SourceID != "a" || r.Documents[1].SourceID != "b" || r.Documents[1].Language != "spl2" {
				t.Fatalf("bad batch: %#v", r)
			}
		})
	}
}

func TestDecodeRewriteRuleSet(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"schema_version":1,"rules":[]}`, true},
		{`{"schema_version":1,"rules":[` + testRule + `]}`, true},
		{`{"schema_version":1,"rules":[],"mode":"apply"}`, false},
		{`{"schema_version":1,"rules":[],"document":{"text":""}}`, false},
		{`{"schema_version":1,"rules":[],"validation_target":{}}`, false},
		{`{"schema_version":1,"rules":null}`, false},
	} {
		r, err := DecodeRuleSet([]byte(tc.raw))
		if tc.valid {
			if err != nil || r.Rules == nil {
				t.Fatalf("%s: %#v, %v", tc.raw, r, err)
			}
		} else if err == nil || !IsInputError(err) {
			t.Fatalf("want input error for %s: %v", tc.raw, err)
		}
	}
}

func TestRewriteModelDirectRequest(t *testing.T) {
	valid, err := DecodeRequest([]byte(testRequest(testRule)))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		change func(*Request)
	}{
		{"version", func(r *Request) { r.SchemaVersion = 0 }},
		{"mode", func(r *Request) { r.Mode = "unknown" }},
		{"document", func(r *Request) { r.Document.Language = "sql" }},
		{"document UTF8", func(r *Request) { r.Document.Text = string([]byte{0xff}) }},
		{"identity UTF8", func(r *Request) { s := string([]byte{0xff}); r.Rules[0].Source.Name = &s }},
		{"identity forms", func(r *Request) { r.Rules[0].Source.Path = []string{} }},
		{"empty ID", func(r *Request) { r.Rules[0].ID = "" }},
		{"duplicate ID", func(r *Request) { r.Rules = append(r.Rules, r.Rules[0]) }},
		{"empty condition", func(r *Request) { r.Rules[0].When = &Condition{} }},
		{"empty all", func(r *Request) { r.Rules[0].When = &Condition{All: []Condition{}} }},
		{"missing value", func(r *Request) {
			id := r.Rules[0].Source
			r.Rules[0].When = &Condition{Fact: "literal", Kind: "field", Identity: &id, Operator: "equals"}
		}},
		{"invalid scalar UTF8", func(r *Request) {
			id := r.Rules[0].Source
			r.Rules[0].When = &Condition{Fact: "literal", Kind: "field", Identity: &id, Operator: "equals", Value: json.RawMessage{'"', 0xff, '"'}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := valid
			r.Rules = append([]Rule{}, valid.Rules...)
			tc.change(&r)
			if _, err := prepareRequest(r); err == nil || !IsInputError(err) {
				t.Fatalf("want input error: %v", err)
			}
		})
	}
	r := valid
	r.Mode = ""
	r.Rules = nil
	prepared, err := prepareRequest(r)
	if err != nil || prepared.Mode != Preview || prepared.Rules == nil {
		t.Fatalf("direct defaults: %#v, %v", prepared, err)
	}
	if _, err := prepareBatchRequest(BatchRequest{SchemaVersion: 1}); err == nil || !IsInputError(err) {
		t.Fatalf("empty direct batch: %v", err)
	}
	if _, err := prepareBatchRequest(BatchRequest{SchemaVersion: 1, Documents: []analysis.QueryDocument{{Text: "first"}, {Text: "later", Profile: "wrong"}}}); err == nil || !IsInputError(err) {
		t.Fatalf("invalid later document: %v", err)
	}
	if !IsInputError(fmt.Errorf("wrapped: %w", inputError("bad request"))) || IsInputError(errors.New("internal")) {
		t.Fatal("input error classification")
	}
}

func TestRewriteModelRequestOwnership(t *testing.T) {
	r, err := DecodeRequest([]byte(testRequest(strings.TrimSuffix(testRule, "}") + `,"when":{"all":[{"fact":"literal","kind":"field","identity":{"path":["actor","name"]},"operator":"equals","value":9007199254740993}]}}`)))
	if err != nil {
		t.Fatal(err)
	}
	r.Rules[0].Target = Identity{Path: []string{"user", "name"}}
	copy, err := prepareRequest(r)
	if err != nil {
		t.Fatal(err)
	}
	*r.Rules[0].Source.Name = "changed"
	r.Rules[0].Target.Path[0] = "changed"
	r.Rules[0].When.All[0].Identity.Path[0] = "changed"
	r.Rules[0].When.All[0].Value[0] = '1'
	if *copy.Rules[0].Source.Name != "src" || copy.Rules[0].Target.Path[0] != "user" || copy.Rules[0].When.All[0].Identity.Path[0] != "actor" || string(copy.Rules[0].When.All[0].Value) != "9007199254740993" {
		t.Fatalf("request retained caller storage: %#v", copy)
	}
}

func TestDecodeRewriteRequestCorpus(t *testing.T) {
	data, err := os.ReadFile("../../testdata/rewrite/requests.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name      string          `json:"name"`
		Operation string          `json:"operation"`
		Request   json.RawMessage `json:"request"`
		Valid     bool            `json:"valid"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var err error
			switch tc.Operation {
			case "request":
				_, err = DecodeRequest(tc.Request)
			case "batch":
				_, err = DecodeBatchRequest(tc.Request)
			case "rules":
				_, err = DecodeRuleSet(tc.Request)
			default:
				t.Fatalf("unknown fixture operation %q", tc.Operation)
			}
			if tc.Valid && err != nil {
				t.Fatal(err)
			}
			if !tc.Valid && (err == nil || !IsInputError(err)) {
				t.Fatalf("want input error: %v", err)
			}
		})
	}
}
