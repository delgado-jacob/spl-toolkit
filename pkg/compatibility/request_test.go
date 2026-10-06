package compatibility

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

func requestFixture(t *testing.T, text string) Request {
	t.Helper()
	requirements, err := analysis.Requirements(analysis.QueryDocument{Text: text, Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	yes := true
	all := environment.Selector{All: &yes}
	provenance := environment.Provenance{SourceKind: "fixture", SourceID: "fixture", ObservedAt: "2026-10-01T12:02:00Z"}
	request := Request{SchemaVersion: 1, Requirements: *requirements, QueryScope: environment.CaptureScope{Namespace: all, App: all, Owner: all}, InputBindings: []InputBinding{}, Snapshot: environment.Snapshot{SchemaVersion: 1, ScopeID: "capture", CaptureScope: environment.CaptureScope{Namespace: all, App: all, Owner: all}, Origin: environment.Origin{InstanceID: "offline", ProductVersion: "9", Producer: "fixture", ProducerVersion: "1"}, Capture: environment.CaptureInterval{Start: "2026-10-01T12:00:00Z", End: "2026-10-01T12:05:00Z"}, Capabilities: []environment.Capability{}, Collections: []environment.Collection{{Kind: "dataset", Coverage: "complete"}, {Kind: "index", Coverage: "complete"}}, Objects: []environment.Object{{ID: "events", Kind: "dataset", Name: "events", Namespace: "search", App: "app", Owner: "nobody", Provenance: provenance}, {ID: "users", Kind: "dataset", Name: "users", Namespace: "search", App: "app", Owner: "nobody", Provenance: provenance}}}}
	for _, input := range requirements.Inputs {
		if input.Kind == "named_placeholder" {
			request.InputBindings = append(request.InputBindings, InputBinding{InputID: input.ID, ObjectID: "events", Expected: environment.ObjectIdentity{Kind: "dataset", Name: "events", Namespace: "search", App: "app", Owner: "nobody"}})
		}
	}
	return request
}
func requestRaw(t *testing.T, request Request) []byte {
	t.Helper()
	raw, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
func requireRequestError(t *testing.T, err error, code, path string) {
	t.Helper()
	if !validation.IsInputError(err) {
		t.Fatalf("expected input error, got %v", err)
	}
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.Code != code || detail.Path != path {
		t.Fatalf("detail %#v, expected %s %s", detail, code, path)
	}
	var wrapped *validation.InputError
	if !errors.As(err, &wrapped) {
		t.Fatal("missing wrapper")
	}
	var encoded map[string]any
	if json.Unmarshal([]byte(err.Error()), &encoded) != nil {
		t.Fatal("error is not structured JSON")
	}
}
func TestRequestStrictAdmission(t *testing.T) {
	fixture := requestFixture(t, "from $events | fields id")
	valid := string(requestRaw(t, fixture))
	tests := []struct{ name, raw, code, path string }{
		{"duplicate", strings.Replace(valid, `"schema_version":1`, `"schema_version":1,"schema_version":1`, 1), "request_invalid", "/schema_version"},
		{"case", strings.Replace(valid, `"input_bindings":[{"input_id":`, `"input_bindings":[{"Input_ID":`, 1), "request_invalid", "/input_bindings/0/Input_ID"},
		{"unknown", strings.Replace(valid, `"input_bindings":[{"input_id":`, `"input_bindings":[{"surprise":0,"input_id":`, 1), "request_invalid", "/input_bindings/0/surprise"},
		{"trailing", valid + ` {}`, "request_invalid", ""},
		{"fraction version", strings.Replace(valid, `"schema_version":1`, `"schema_version":1.0`, 1), "request_invalid", "/schema_version"},

		{"both selector", strings.Replace(valid, `"namespace":{"all":true}`, `"namespace":{"all":true,"values":[]}`, 1), "request_invalid", "/query_scope/namespace"},
		{"nested raw both selector", strings.Replace(valid, `"capture_scope":{"namespace":{"all":true}`, `"capture_scope":{"namespace":{"all":true,"values":[]}`, 1), "snapshot_invalid", "/snapshot/capture_scope/namespace"},
		{"schema null", strings.TrimSuffix(valid, "}") + `,"schema_bundle":null}`, "schema_bundle_invalid", "/schema_bundle"},
		{"raw surrogate", strings.Replace(valid, `"object_id":"events"`, `"object_id":"\ud800"`, 1), "request_invalid", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeRequest([]byte(tt.raw))
			requireRequestError(t, err, tt.code, tt.path)
		})
	}
	raw := append([]byte{}, []byte(valid)...)
	raw = append(raw[:len(raw)-1], 0xff, '}')
	_, err := DecodeRequest(raw)
	requireRequestError(t, err, "request_invalid", "")
	var tree map[string]any
	_ = json.Unmarshal([]byte(valid), &tree)
	tree["input_bindings"] = nil
	raw, _ = json.Marshal(tree)
	_, err = DecodeRequest(raw)
	requireRequestError(t, err, "request_invalid", "/input_bindings")
	delete(tree, "input_bindings")
	raw, _ = json.Marshal(tree)
	_, err = DecodeRequest(raw)
	requireRequestError(t, err, "request_invalid", "/input_bindings")
	tree["input_bindings"] = []any{}
	requirements := tree["requirements"].(map[string]any)
	delete(requirements, "inputs")
	raw, _ = json.Marshal(tree)
	_, err = DecodeRequest(raw)
	requireRequestError(t, err, "requirements_refresh_required", "/requirements/inputs")
}
func TestRequestSemanticParity(t *testing.T) {
	tests := []struct {
		name, code, path string
		mutate           func(*Request)
	}{
		{"version", "request_invalid", "/schema_version", func(r *Request) { r.SchemaVersion = 2 }},
		{"stale", "requirements_stale", "/requirements/capability_revision", func(r *Request) { r.Requirements.CapabilityRevision = "old" }},
		{"old evidence", "requirements_refresh_required", "/requirements/inputs", func(r *Request) { r.Requirements.Inputs = nil }},
		{"missing coverage", "requirements_refresh_required", "/requirements/field_attribution_coverage/reasons", func(r *Request) { r.Requirements.FieldAttributionCoverage.Reasons = nil }},
		{"duplicate input", "requirements_inconsistent", "/requirements/inputs/1/id", func(r *Request) { r.Requirements.Inputs = append(r.Requirements.Inputs, r.Requirements.Inputs[0]) }},
		{"duplicate requirement", "requirements_inconsistent", "/requirements/items/2/id", func(r *Request) { r.Requirements.Items = append(r.Requirements.Items, r.Requirements.Items[0]) }},
		{"duplicate occurrence", "requirements_inconsistent", "/requirements/inputs/0/occurrences/1/id", func(r *Request) {
			r.Requirements.Inputs[0].Occurrences = append(r.Requirements.Inputs[0].Occurrences, r.Requirements.Inputs[0].Occurrences[0])
		}},
		{"invalid field identity", "requirements_inconsistent", "/requirements/items/1/field_identity", func(r *Request) { r.Requirements.Items[1].FieldIdentity.Kind = "display" }},
		{"invalid location", "requirements_inconsistent", "/requirements/inputs/0/occurrences/0", func(r *Request) { r.Requirements.Inputs[0].Occurrences[0].Location.End.Offset = -1 }},
		{"bad digest", "requirements_inconsistent", "/requirements/query", func(r *Request) { r.Requirements.Query.QueryDigest = "sha256:bad" }},
		{"typed selector union", "request_invalid", "/query_scope/app", func(r *Request) { r.QueryScope.App.Values = []string{"app"} }},
		{"duplicate selector", "request_invalid", "/query_scope/app", func(r *Request) { r.QueryScope.App = environment.Selector{Values: []string{"app", "app"}} }},
		{"missing captured array", "snapshot_invalid", "/snapshot/objects", func(r *Request) { r.Snapshot.Objects = nil }},
		{"duplicate binding", "binding_invalid", "/input_bindings/1/input_id", func(r *Request) { r.InputBindings = append(r.InputBindings, r.InputBindings[0]) }},
		{"unknown input", "binding_invalid", "/input_bindings/0/input_id", func(r *Request) { r.InputBindings[0].InputID = "unknown" }},
		{"missing binding", "binding_missing", "/input_bindings", func(r *Request) { r.InputBindings = []InputBinding{} }},
		{"identity mismatch", "binding_invalid", "/input_bindings/0/expected", func(r *Request) { r.InputBindings[0].Expected.Name = "users" }},
		{"blank object", "binding_invalid", "/input_bindings/0/object_id", func(r *Request) { r.InputBindings[0].ObjectID = " " }},
		{"wrong kind", "binding_invalid", "/input_bindings/0/expected/kind", func(r *Request) { r.InputBindings[0].Expected.Kind = "lookup" }},
		{"schema without bundle", "binding_invalid", "/input_bindings/0/schema_id", func(r *Request) { r.InputBindings[0].SchemaID = "schema" }},
		{"item cross link", "requirements_inconsistent", "/requirements/items/0/input_id", func(r *Request) { r.Requirements.Items[0].InputID = "unknown" }},
		{"occurrence link", "requirements_inconsistent", "/requirements/items/0/occurrences/0/input_occurrence_ids", func(r *Request) { r.Requirements.Items[0].Occurrences[0].InputOccurrenceIDs = []string{"unknown"} }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := requestFixture(t, "from $events | fields id")
			tt.mutate(&r)
			_, err := Check(r)
			requireRequestError(t, err, tt.code, tt.path)
			_, err = CheckJSON(requestRaw(t, r))
			requireRequestError(t, err, tt.code, tt.path)
		})
	}
}
func TestRequestDecodeDetachedAndOffset(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	raw := requestRaw(t, r)
	decoded, err := DecodeRequest(raw)
	if err != nil {
		t.Fatal(err)
	}
	for i := range raw {
		raw[i] = ' '
	}
	if decoded.InputBindings[0].ObjectID != "events" {
		t.Fatal("aliased raw input")
	}
	valid := string(requestRaw(t, r))
	bad := strings.Replace(valid, `"capture_scope":{"namespace":{"all":true}`, `"capture_scope":{"namespace":{"all":true,"all":true}`, 1)
	_, err = DecodeRequest([]byte(bad))
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.ByteOffset == nil || *detail.ByteOffset < strings.Index(bad, `"snapshot"`) {
		t.Fatalf("bad artifact offset: %#v", detail)
	}
	first := *detail.ByteOffset
	*detail.ByteOffset = 0
	again, _ := RequestErrorDetails(err)
	if *again.ByteOffset != first {
		t.Fatal("aliased details")
	}
}

func TestRequestFieldIdentityArraysAndDescriptor(t *testing.T) {
	r := requestFixture(t, "from $events | fields id")
	r.Requirements.Items[1].FieldIdentity.Segments = nil
	_, err := Check(r)
	requireRequestError(t, err, "requirements_refresh_required", "/requirements/items/1/field_identity/segments")
	_, err = CheckJSON(requestRaw(t, r))
	requireRequestError(t, err, "requirements_refresh_required", "/requirements/items/1/field_identity/segments")
	r = requestFixture(t, `from {kind:"index",properties:{name:"main"}} | fields id`)
	r.Requirements.Inputs[0].Identity.Value = `{"kind":"index","kind":"index","properties":{"name":"main"}}`
	r.Requirements.Inputs[0].Name = r.Requirements.Inputs[0].Identity.Value
	_, err = Check(r)
	requireRequestError(t, err, "requirements_inconsistent", "/requirements/inputs/0/identity")
}
