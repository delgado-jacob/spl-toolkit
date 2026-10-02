package splunkexport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Hand-authored synthetic fixtures, never live payloads. Endpoint, entry/ACL,
// paging and unpaged tag shapes follow these official references:
// https://help.splunk.com/en/splunk-cloud-platform/leverage-rest-apis/rest-api-reference/10.4.2604/knowledge-endpoints/knowledge-endpoint-descriptions
// https://help.splunk.com/en/splunk-cloud-platform/leverage-rest-apis/rest-api-tutorials/9.2.2406/rest-api-tutorials/managing-knowledge-objects
// Macro default semantics follow the macros.conf specification:
// https://help.splunk.com/en/data-management/splunk-enterprise-admin-manual/9.0/welcome-to-splunk-enterprise-administration/configuration-file-reference/9.0.0-configuration-file-reference/macros.conf
var configurationFamilies = []string{"configs/conf-macros", "saved/searches", "saved/eventtypes", "data/transforms/lookups", "datamodel/model", "search/tags", "data/props/calcfields", "data/props/extractions", "data/transforms/extractions"}

type configurationRequest struct {
	path  string
	query url.Values
}

func newConfigurationFixture(t *testing.T, respond func(http.ResponseWriter, *http.Request, string)) (*Client, *[]configurationRequest) {
	t.Helper()
	requests := []configurationRequest{}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Query().Get("output_mode") != "json" {
			t.Error("configuration must use JSON GET")
		}
		requests = append(requests, configurationRequest{r.URL.EscapedPath(), r.URL.Query()})
		path := strings.SplitN(strings.TrimPrefix(r.URL.EscapedPath(), "/servicesNS/"), "/", 3)
		if len(path) != 3 {
			t.Errorf("unexpected path %s", r.URL.EscapedPath())
			http.NotFound(w, r)
			return
		}
		family := path[2]
		allowed := false
		for _, known := range configurationFamilies {
			if family == known || strings.HasPrefix(family, known+"/") {
				allowed = true
			}
		}
		if !allowed {
			t.Errorf("unapproved endpoint %s", family)
			http.NotFound(w, r)
			return
		}
		respond(w, r, family)
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	all := true
	o.Scope = environment.CaptureScope{Namespace: environment.Selector{All: &all}, App: environment.Selector{All: &all}, Owner: environment.Selector{All: &all}}
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c, &requests
}
func configurationEntry(name string, content map[string]any) map[string]any {
	return map[string]any{"name": name, "acl": map[string]any{"owner": "nobody", "app": "search", "sharing": "app"}, "content": content}
}
func configurationFeed(w http.ResponseWriter, r *http.Request, entries []map[string]any) {
	if entries == nil {
		entries = []map[string]any{}
	}
	if strings.HasSuffix(r.URL.Path, "/search/tags") {
		json.NewEncoder(w).Encode(map[string]any{"entry": entries})
		return
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	count, _ := strconv.Atoi(r.URL.Query().Get("count"))
	if count == 0 {
		count = 100
	}
	end := offset + count
	if end > len(entries) {
		end = len(entries)
	}
	if offset > len(entries) {
		offset = len(entries)
	}
	json.NewEncoder(w).Encode(map[string]any{"entry": entries[offset:end], "paging": map[string]int{"total": len(entries), "offset": offset, "perPage": count}})
}
func configurationObject(t *testing.T, objects []environment.Object, kind, name string, arity *int) environment.Object {
	t.Helper()
	for _, o := range objects {
		if o.Kind == kind && o.Name == name && (arity == nil || o.Arity != nil && *arity == *o.Arity) {
			return o
		}
	}
	t.Fatalf("missing %s %s", kind, name)
	return environment.Object{}
}
func configurationHasDiagnostic(diagnostics []Diagnostic, code, kind string) bool {
	for _, d := range diagnostics {
		if d.Code == code && d.Kind == kind {
			return true
		}
	}
	return false
}
func configurationBundle(t *testing.T, c *Client, result ConfigurationResult) closure.DefinitionBundle {
	t.Helper()
	now := time.Now().UTC()
	snapshot := environment.Snapshot{SchemaVersion: 1, ScopeID: "synthetic-config", CaptureScope: c.options.Scope, Origin: environment.Origin{InstanceID: "synthetic", ProductVersion: "9.4", Producer: "spl-toolkit", ProducerVersion: "test"}, Capture: environment.CaptureInterval{Start: now.Add(-time.Minute).Format(time.RFC3339Nano), End: now.Add(time.Minute).Format(time.RFC3339Nano)}, Capabilities: []environment.Capability{}, Collections: result.Collections, Objects: result.Objects}
	prepared, report, err := environment.PrepareSnapshot(snapshot)
	if err != nil || prepared == nil || report.Status == "invalid" {
		t.Fatalf("snapshot %v %+v", err, report)
	}
	env, report, err := environment.Pair(prepared, nil)
	if err != nil || env == nil {
		t.Fatalf("pair %v %+v", err, report)
	}
	bundle, err := env.DefinitionBundle(c.options.Scope)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}
func TestConfigurationQueryDefinitionsAndClosure(t *testing.T) {
	macroText := "search index=main | eval marker=\"token=literal\"\n"
	c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		var entries []map[string]any
		switch family {
		case "configs/conf-macros":
			entries = []map[string]any{configurationEntry("base", map[string]any{"definition": macroText, "args": "", "iseval": "0", "validation": "", "password": "discard-me"}), configurationEntry("base(1)", map[string]any{"definition": "search index=$index$", "args": "index", "iseval": false, "validation": "isstr($index$)"}), configurationEntry("evaluate(1)", map[string]any{"definition": "upper($value$)", "args": "value", "iseval": "1", "validation": "isstr($value$)"})}
		case "saved/searches":
			entries = []map[string]any{configurationEntry("report", map[string]any{"search": "`base` | stats count", "action.email.to": "discard-me", "token": "discard-me"})}
		case "saved/eventtypes":
			entries = []map[string]any{configurationEntry("event", map[string]any{"search": "index=main marker=\"password=literal\"", "description": "discard-me"})}
		}
		configurationFeed(w, r, entries)
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if len(result.Objects) != 5 || len(result.Collections) != 12 {
		t.Fatalf("mapped result %+v", result)
	}
	zero, one := 0, 1
	base := configurationObject(t, result.Objects, "macro", "base", &zero)
	if base.Document == nil || base.Document.Text != macroText || base.Document.Language != "spl" || base.Document.Profile != "splunkd" || base.Document.Version != "current" || base.Document.SourceID != base.Provenance.SourceID || base.EvalBased == nil || *base.EvalBased || base.Validation == nil || *base.Validation != "" {
		t.Fatalf("macro %+v", base)
	}
	overload := configurationObject(t, result.Objects, "macro", "base", &one)
	if overload.ID == base.ID || !reflect.DeepEqual(overload.Arguments, []string{"index"}) || *overload.Validation != "isstr($index$)" {
		t.Fatal(overload)
	}
	eval := configurationObject(t, result.Objects, "macro", "evaluate", &one)
	if !*eval.EvalBased || eval.Document.Text != "upper($value$)" {
		t.Fatal(eval)
	}
	event := configurationObject(t, result.Objects, "event_type", "event", nil)
	if event.Document.Text != "index=main marker=\"password=literal\"" {
		t.Fatal(event)
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "discard-me") {
		t.Fatal("raw properties retained")
	}
	for _, o := range result.Objects {
		if o.Namespace != "nobody/search" || o.Owner != "nobody" || o.App != "search" || o.Sharing != "app" {
			t.Fatal(o)
		}
	}
	bundle := configurationBundle(t, c, result)
	no, yes := false, true
	emptyValidation, indexValidation, valueValidation := "", "isstr($index$)", "isstr($value$)"
	explicit := closure.DefinitionBundle{SchemaVersion: 1, ScopeID: "synthetic-config", Collections: bundle.Collections, Objects: []closure.Definition{
		{Kind: "macro", Name: "base", Arity: &zero, EvalBased: &no, Validation: &emptyValidation, Document: &analysis.QueryDocument{Text: macroText}},
		{Kind: "macro", Name: "base", Arity: &one, Arguments: []string{"index"}, EvalBased: &no, Validation: &indexValidation, Document: &analysis.QueryDocument{Text: "search index=$index$"}},
		{Kind: "macro", Name: "evaluate", Arity: &one, Arguments: []string{"value"}, EvalBased: &yes, Validation: &valueValidation, Document: &analysis.QueryDocument{Text: "upper($value$)"}},
		{Kind: "saved_search", Name: "report", Document: &analysis.QueryDocument{Text: "`base` | stats count"}},
		{Kind: "event_type", Name: "event", Document: &analysis.QueryDocument{Text: "index=main marker=\"password=literal\""}},
	}}
	// Acquisition identities come from this run; definition content and metadata
	// are supplied independently, as a canonical offline caller would provide them.
	for i := range explicit.Objects {
		o := &explicit.Objects[i]
		mapped := configurationObject(t, result.Objects, o.Kind, o.Name, o.Arity)
		o.ID, o.SourceID = mapped.ID, mapped.Provenance.SourceID
		o.App, o.Owner, o.Sharing = "search", "nobody", "app"
		o.Document.Language, o.Document.Profile, o.Document.Version, o.Document.SourceID = "spl", "splunkd", "current", o.SourceID
	}
	sort.Slice(explicit.Objects, func(i, j int) bool { return explicit.Objects[i].ID < explicit.Objects[j].ID })
	if !reflect.DeepEqual(bundle, explicit) {
		t.Fatalf("projected definition bundle differs from supplied definitions:\ngot %#v\nwant %#v", bundle, explicit)
	}
	for _, query := range []string{"`base` | stats count", "| savedsearch report"} {
		got, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: query, Language: "spl", Profile: "splunkd", Version: "current", SourceID: "synthetic-root"}, Bundle: bundle, Bindings: []closure.Binding{}})
		if err != nil || got == nil || query == "`base` | stats count" && !got.Coverage.Complete {
			t.Fatalf("closure %q %v %+v", query, err, got)
		}
		want, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: query, Language: "spl", Profile: "splunkd", Version: "current", SourceID: "synthetic-root"}, Bundle: explicit, Bindings: []closure.Binding{}})
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("mapped closure differs from supplied definitions: %v", err)
		}
	}
	reportObject := configurationObject(t, result.Objects, "saved_search", "report", nil)
	if reportObject.Document == nil || reportObject.Document.Text != "`base` | stats count" {
		t.Fatal(reportObject)
	}
	savedClosure, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: *reportObject.Document, Bundle: bundle})
	if err != nil || !savedClosure.Coverage.Complete || savedClosure.EffectiveAnalysis == nil {
		t.Fatalf("saved-search definition closure: %v %+v", err, savedClosure)
	}
	for _, kind := range []string{"dataset", "module", "function", "external_command"} {
		if collectionCoverage(result.Collections, kind) != "unavailable" || !configurationHasDiagnostic(result.Diagnostics, "adapter_unsupported", kind) {
			t.Fatalf("unsupported %s %+v", kind, result)
		}
	}
	if len(*requests) != 9 {
		t.Fatalf("extra endpoints: %+v", *requests)
	}
}

func TestConfigurationMacroOmittedDefaultsCompleteFeedAndClosure(t *testing.T) {
	definition := "search index=synthetic | eval marker=\"literal\"\n"
	entries := make([]map[string]any, 100)
	for i := range entries {
		content := map[string]any{"definition": definition}
		switch i % 3 {
		case 1:
			content["args"] = ""
		case 2:
			// macros.conf ignores args for zero-argument stanzas.
			content["args"] = "ignored,ignored"
		}
		entries[i] = configurationEntry(fmt.Sprintf("default_macro_%03d", i), content)
	}
	c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		if family == "configs/conf-macros" {
			configurationFeed(w, r, entries)
			return
		}
		if strings.HasPrefix(family, "configs/conf-macros/") {
			t.Error("documented defaults must not require detail")
			http.NotFound(w, r)
			return
		}
		configurationFeed(w, r, nil)
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if result.Err != nil || len(result.Objects) != 100 || collectionCoverage(result.Collections, "macro") != "complete" {
		t.Fatalf("omitted defaults lost macro evidence: retained=%d coverage=%s err=%v", len(result.Objects), collectionCoverage(result.Collections, "macro"), result.Err)
	}
	for _, object := range result.Objects {
		if object.Arity == nil || *object.Arity != 0 || !reflect.DeepEqual(object.Arguments, []string{}) || object.EvalBased == nil || *object.EvalBased || object.Validation != nil || object.Document == nil || object.Document.Text != definition {
			t.Fatalf("documented defaults or definition changed: %+v", object)
		}
	}
	if len(*requests) != len(configurationFamilies) {
		t.Fatalf("unexpected requests for omitted defaults: %d", len(*requests))
	}
	bundle := configurationBundle(t, c, result)
	report, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "`default_macro_000` | stats count", Language: "spl", Profile: "splunkd", Version: "current", SourceID: "synthetic-root"}, Bundle: bundle})
	if err != nil || report == nil || !report.Coverage.Complete || report.EffectiveAnalysis == nil || len(report.Traversal) != 1 || report.Traversal[0].Resolution != "resolved" {
		t.Fatalf("defaulted macro did not expand offline: %v %+v", err, report)
	}
}

func TestConfigurationOpaqueDefinitionsAndExtractionFamilyIdentity(t *testing.T) {
	c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		entries := []map[string]any{}
		switch family {
		case "data/transforms/lookups":
			entries = append(entries, configurationEntry("users", map[string]any{"filename": "users.csv", "collection": "private_rows", "external_cmd": "discard-me", "password": "discard-me"}))
		case "datamodel/model":
			entries = append(entries, configurationEntry("security", map[string]any{"description": "{\"objects\":[{\"search\":\"discard-me\"}]}"}))
		case "search/tags":
			entries = append(entries, configurationEntry("tag", map[string]any{"eventtype::private": "discard-me"}))
		case "data/props/calcfields":
			entries = append(entries, configurationEntry("type : EVAL-value", map[string]any{"value": "if(secret,1,0)"}))
		case "data/props/extractions", "data/transforms/extractions":
			entries = append(entries, configurationEntry("same", map[string]any{"value": "(?<password>.*)", "REGEX": "discard-me", "SOURCE_KEY": "_raw"}))
		}
		configurationFeed(w, r, entries)
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if len(result.Objects) != 6 {
		t.Fatal(result)
	}
	for _, o := range result.Objects {
		if o.Document != nil || len(o.Relations) != 0 || o.Arity != nil {
			t.Fatalf("fabricated query/relation %+v", o)
		}
		if !configurationHasDiagnostic(result.Diagnostics, "definition_unsupported", o.Kind) {
			t.Fatalf("missing opaque definition diagnostic %s", o.Kind)
		}
	}
	if !reflect.DeepEqual(objectNames(result.Objects, "field_extraction"), []string{"props:same", "transforms:same"}) {
		t.Fatal(result.Objects)
	}
	for _, kind := range []string{"lookup", "data_model", "tag", "calculated_field", "field_extraction"} {
		if collectionCoverage(result.Collections, kind) != "complete" {
			t.Fatalf("identity coverage must be independent from closure support: %+v", result)
		}
	}
	bundle := configurationBundle(t, c, result)
	got, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: analysis.QueryDocument{Text: "| savedsearch missing", Language: "spl", SourceID: "synthetic-root"}, Bundle: bundle})
	if err != nil || got.Coverage.Complete {
		t.Fatalf("invented closure %v %+v", err, got)
	}
	encoded, _ := json.Marshal(result)
	for _, secret := range []string{"discard-me", "private_rows", "users.csv", "if(secret", "(?<password>"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("opaque content retained: %s", secret)
		}
	}
}

func TestConfigurationFiniteContextsAndReturnedACLFiltering(t *testing.T) {
	c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		a := configurationEntry("inherited", map[string]any{"search": "index=main"})
		a["acl"] = map[string]any{"owner": "o/a", "app": "a?b", "sharing": "global"}
		b := configurationEntry("other", map[string]any{"search": "index=other"})
		if family != "saved/searches" {
			configurationFeed(w, r, nil)
			return
		}
		configurationFeed(w, r, []map[string]any{a, b})
	})
	c.options.Scope.Owner = environment.Selector{Values: []string{"o/a", "other"}}
	c.options.Scope.App = environment.Selector{Values: []string{"a?b", "else"}}
	c.options.Scope.Namespace = environment.Selector{Values: []string{"o%2Fa/a%3Fb"}}
	result := c.collectConfiguration(context.Background(), "synthetic")
	if len(result.Objects) != 1 {
		t.Fatal(result)
	}
	o := result.Objects[0]
	if o.Name != "inherited" || o.Namespace != "o%2Fa/a%3Fb" || o.App != "a?b" || o.Owner != "o/a" {
		t.Fatal(o)
	}
	contexts := map[string]int{}
	for _, r := range *requests {
		parts := strings.SplitN(strings.TrimPrefix(r.path, "/servicesNS/"), "/", 3)
		contexts[parts[0]+"/"+parts[1]]++
	}
	if len(contexts) != 4 || len(*requests) != 36 {
		t.Fatalf("finite cross-product incomplete %+v", contexts)
	}
	for _, n := range contexts {
		if n != 9 {
			t.Fatal(contexts)
		}
	}
	configurationBundle(t, c, result)
}

func TestConfigurationACLNullFallbackConflictAndMissingContext(t *testing.T) {
	for _, exact := range []bool{false, true} {
		t.Run(fmt.Sprint(exact), func(t *testing.T) {
			c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family != "saved/searches" {
					configurationFeed(w, r, nil)
					return
				}
				valid := configurationEntry("valid", map[string]any{"search": "index=main", "eai:acl": nil})
				fallback := configurationEntry("fallback", map[string]any{"search": "index=main", "eai:acl": map[string]any{"owner": "nobody", "app": "search", "sharing": "app"}})
				delete(fallback, "acl")
				conflict := configurationEntry("conflict", map[string]any{"search": "index=main", "eai:acl": map[string]any{"owner": "alice", "app": "search", "sharing": "app"}})
				unknown := configurationEntry("unknown", map[string]any{"search": "index=main"})
				delete(unknown, "acl")
				configurationFeed(w, r, []map[string]any{valid, fallback, conflict, unknown})
			})
			if exact {
				c.options.Scope.App = environment.Selector{Values: []string{"search"}}
			}
			result := c.collectConfiguration(context.Background(), "synthetic")
			expected := []string{"fallback", "unknown", "valid"}
			if exact {
				expected = []string{"fallback", "valid"}
			}
			if !reflect.DeepEqual(objectNames(result.Objects, "saved_search"), expected) || collectionCoverage(result.Collections, "saved_search") != "partial" || !configurationHasDiagnostic(result.Diagnostics, "context_conflict", "saved_search") || !configurationHasDiagnostic(result.Diagnostics, "context_unavailable", "saved_search") {
				t.Fatal(result)
			}
			if !exact {
				unknown := configurationObject(t, result.Objects, "saved_search", "unknown", nil)
				if unknown.Namespace != "" || unknown.Owner != "" || unknown.App != "" {
					t.Fatal("request context invented")
				}
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationTagsAreUnpagedUnknownContext(t *testing.T) {
	for _, dimension := range []string{"all", "app", "owner", "namespace"} {
		t.Run(dimension, func(t *testing.T) {
			c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family != "search/tags" {
					configurationFeed(w, r, nil)
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"entry": []any{map[string]any{"name": "synthetic-tag", "links": map[string]any{"alternate": "https://elsewhere.invalid/arbitrary"}}}})
			})
			if dimension != "all" {
				selector := environment.Selector{Values: []string{"search"}}
				switch dimension {
				case "app":
					c.options.Scope.App = selector
				case "owner":
					c.options.Scope.Owner = selector
				case "namespace":
					c.options.Scope.Namespace = selector
				}
			}
			result := c.collectConfiguration(context.Background(), "synthetic")
			want := 0
			if dimension == "all" {
				want = 1
				o := configurationObject(t, result.Objects, "tag", "synthetic-tag", nil)
				if o.Namespace != "" || o.App != "" || o.Owner != "" {
					t.Fatal(o)
				}
			}
			if len(result.Objects) != want || !configurationHasDiagnostic(result.Diagnostics, "context_unavailable", "tag") {
				t.Fatal(result)
			}
			for _, r := range *requests {
				if strings.HasSuffix(r.path, "/search/tags") && (r.query.Has("count") || r.query.Has("offset")) {
					t.Fatal("invented tag pagination")
				}
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationMacroInvalidArityAndMissingDefinition(t *testing.T) {
	c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		if strings.HasPrefix(family, "configs/conf-macros/") {
			http.Error(w, "discard-me", 503)
			return
		}
		if family != "configs/conf-macros" {
			configurationFeed(w, r, nil)
			return
		}
		configurationFeed(w, r, []map[string]any{configurationEntry("valid", map[string]any{"definition": "index=main", "args": "", "iseval": 0}), configurationEntry("missing", map[string]any{"args": "", "iseval": 0}), configurationEntry("bad(2)", map[string]any{"definition": "index=main", "args": "one"}), configurationEntry("implicit", map[string]any{"definition": "index=main", "args": "one"}), configurationEntry("bad(x)", map[string]any{"definition": "index=main", "args": ""}), configurationEntry("repeated(2)", map[string]any{"definition": "index=main", "args": "a,a"}), configurationEntry("null_args", map[string]any{"definition": "index=main", "args": nil}), configurationEntry("array_args", map[string]any{"definition": "index=main", "args": []string{}}), configurationEntry("boolean_args", map[string]any{"definition": "index=main", "args": false}), configurationEntry("number_args", map[string]any{"definition": "index=main", "args": 0})})
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if !reflect.DeepEqual(objectNames(result.Objects, "macro"), []string{"implicit", "missing", "valid"}) || collectionCoverage(result.Collections, "macro") != "partial" || !configurationHasDiagnostic(result.Diagnostics, "macro_arity_invalid", "macro") || !configurationHasDiagnostic(result.Diagnostics, "definition_unavailable", "macro") || !configurationHasDiagnostic(result.Diagnostics, "configuration_detail_unavailable", "macro") {
		t.Fatal(result)
	}
	missing := configurationObject(t, result.Objects, "macro", "missing", nil)
	if missing.Document != nil || missing.Arity == nil || *missing.Arity != 0 {
		t.Fatal(missing)
	}
	detail := 0
	for _, r := range *requests {
		if strings.Contains(r.path, "conf-macros/") {
			detail++
		}
	}
	if detail != 1 {
		t.Fatalf("known definition detail attempts %d", detail)
	}
	configurationBundle(t, c, result)
}

func TestConfigurationPaginationRetainsEarlierEvidence(t *testing.T) {
	for _, mode := range []string{"complete", "offset", "total", "empty", "failure"} {
		t.Run(mode, func(t *testing.T) {
			entries := []map[string]any{}
			for i := 0; i < 101; i++ {
				entries = append(entries, configurationEntry(fmt.Sprintf("report-%03d", i), map[string]any{"search": "index=main"}))
			}
			c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family != "saved/searches" {
					configurationFeed(w, r, nil)
					return
				}
				offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
				if offset == 100 && mode != "complete" {
					if mode == "failure" {
						http.Error(w, "discard-me", 500)
						return
					}
					total, pageOffset := 101, 100
					pageEntries := entries[100:]
					if mode == "offset" {
						pageOffset = 0
					}
					if mode == "total" {
						total = 102
					}
					if mode == "empty" {
						pageEntries = []map[string]any{}
					}
					json.NewEncoder(w).Encode(map[string]any{"entry": pageEntries, "paging": map[string]int{"total": total, "offset": pageOffset, "perPage": 100}, "links": map[string]string{"next": "https://elsewhere.invalid"}})
					return
				}
				configurationFeed(w, r, entries)
			})
			c.options.MaxRows = 200
			result := c.collectConfiguration(context.Background(), "synthetic")
			want, coverage := 100, "partial"
			if mode == "complete" {
				want = 101
				coverage = "complete"
			}
			if len(result.Objects) != want || collectionCoverage(result.Collections, "saved_search") != coverage {
				t.Fatalf("mode %s: %d %+v", mode, len(result.Objects), result.Collections)
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationDetailUsesFixedFamilyAndValidatesIdentity(t *testing.T) {
	for _, mode := range []string{"complete", "missing", "denied", "failed", "invalid-json", "wrong-name", "conflict"} {
		t.Run(mode, func(t *testing.T) {
			c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family == "saved/searches/detail%2Freport" {
					if mode == "denied" || mode == "failed" {
						status := http.StatusForbidden
						if mode == "failed" {
							status = http.StatusServiceUnavailable
						}
						http.Error(w, "discard-me", status)
						return
					}
					if mode == "invalid-json" {
						fmt.Fprint(w, "{")
						return
					}
					name := "detail/report"
					if mode == "wrong-name" {
						name = "wrong"
					}
					content := map[string]any{"search": " index=main\n"}
					if mode == "missing" {
						content = map[string]any{}
					}
					e := configurationEntry(name, content)
					if mode == "conflict" {
						e["acl"] = map[string]any{"owner": "alice", "app": "search", "sharing": "app"}
					}
					json.NewEncoder(w).Encode(map[string]any{"entry": []any{e}})
					return
				}
				if family != "saved/searches" {
					configurationFeed(w, r, nil)
					return
				}
				e := configurationEntry("detail/report", map[string]any{"action.token": "discard-me"})
				e["id"] = "https://elsewhere.invalid/steal"
				configurationFeed(w, r, []map[string]any{e})
			})
			result := c.collectConfiguration(context.Background(), "synthetic")
			o := configurationObject(t, result.Objects, "saved_search", "detail/report", nil)
			if mode == "complete" {
				if o.Document == nil || o.Document.Text != " index=main\n" || collectionCoverage(result.Collections, "saved_search") != "complete" {
					t.Fatal(result)
				}
			} else if o.Document != nil || !configurationHasDiagnostic(result.Diagnostics, "definition_unavailable", "saved_search") {
				t.Fatal(result)
			}
			expectedCoverage := "partial"
			failedDetail := mode != "complete" && mode != "missing"
			if !failedDetail {
				expectedCoverage = "complete"
			}
			if len(result.Objects) != 1 || o.Owner != "nobody" || o.App != "search" || collectionCoverage(result.Collections, "saved_search") != expectedCoverage || configurationHasDiagnostic(result.Diagnostics, "configuration_detail_unavailable", "saved_search") != failedDetail {
				t.Fatalf("detail acquisition must retain list identity with honest coverage: %+v", result)
			}
			if len(*requests) != 10 {
				t.Fatalf("detail protocol %+v", *requests)
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationDuplicatesUseOnlySafeMappedContent(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		t.Run(fmt.Sprint(conflict), func(t *testing.T) {
			c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family != "saved/searches" {
					configurationFeed(w, r, nil)
					return
				}
				text := "index=main"
				secret := "discard-first"
				if strings.Contains(r.URL.Path, "/second/") {
					secret = "discard-second"
					if conflict {
						text = "index=other"
					}
				}
				configurationFeed(w, r, []map[string]any{configurationEntry("report", map[string]any{"search": text, "password": secret, "action.token": secret})})
			})
			c.options.Scope.App = environment.Selector{Values: []string{"search", "second"}}
			result := c.collectConfiguration(context.Background(), "synthetic")
			if len(result.Objects) != 1 || result.Objects[0].Document.Text != "index=main" {
				t.Fatal(result)
			}
			coverage := "complete"
			if conflict {
				coverage = "partial"
			}
			if collectionCoverage(result.Collections, "saved_search") != coverage || configurationHasDiagnostic(result.Diagnostics, "object_conflict", "saved_search") != conflict {
				t.Fatal(result)
			}
			next := c.collectConfiguration(context.Background(), "synthetic")
			if next.Objects[0].Provenance.SourceID != result.Objects[0].Provenance.SourceID {
				t.Fatal("nondeterministic retained provenance")
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationMalformedIdentityRetainsOthers(t *testing.T) {
	c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		if family != "saved/searches" {
			configurationFeed(w, r, nil)
			return
		}
		bad := configurationEntry("", map[string]any{"search": "index=main"})
		configurationFeed(w, r, []map[string]any{configurationEntry("first", map[string]any{"search": "index=main"}), bad, configurationEntry("second", map[string]any{"search": "index=main"}), configurationEntry("omitted", map[string]any{"search": "index=main"})})
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	names := objectNames(result.Objects, "saved_search")
	sort.Strings(names)
	if !reflect.DeepEqual(names, []string{"first", "omitted", "second"}) || collectionCoverage(result.Collections, "saved_search") != "partial" || !configurationHasDiagnostic(result.Diagnostics, "configuration_identity_invalid", "saved_search") {
		t.Fatal(result)
	}
	configurationBundle(t, c, result)
}

func TestConfigurationArtifactBudgetIsFatal(t *testing.T) {
	text := strings.Repeat("x", 7500*1024)
	c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		if family != "saved/searches" {
			configurationFeed(w, r, nil)
			return
		}
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{configurationEntry(fmt.Sprintf("report-%d", offset), map[string]any{"search": text})}, "paging": map[string]int{"total": 9, "offset": offset, "perPage": 1}})
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if result.Err == nil || result.Err.Error() != "artifact_too_large" {
		t.Fatalf("budget must stop export: %v", result.Err)
	}
}

func TestConfigurationMacroMetadataCannotInventExpansion(t *testing.T) {
	for _, mode := range []string{"true", "false", "unknown-eval", "missing-eval", "null-eval", "array-eval", "object-eval", "number-eval", "bad-validation"} {
		t.Run(mode, func(t *testing.T) {
			c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				content := map[string]any{"definition": "index=main", "args": "", "iseval": mode}
				if mode == "missing-eval" {
					delete(content, "iseval")
				}
				switch mode {
				case "null-eval":
					content["iseval"] = nil
				case "array-eval":
					content["iseval"] = []string{"false"}
				case "object-eval":
					content["iseval"] = map[string]any{"value": false}
				case "number-eval":
					content["iseval"] = 2
				}
				if mode == "bad-validation" {
					content["iseval"] = 0
					content["validation"] = []string{"not-a-string"}
				}
				e := configurationEntry("macro", content)
				if family == "configs/conf-macros/macro" {
					json.NewEncoder(w).Encode(map[string]any{"entry": []any{e}})
					return
				}
				if family != "configs/conf-macros" {
					configurationFeed(w, r, nil)
					return
				}
				configurationFeed(w, r, []map[string]any{e})
			})
			result := c.collectConfiguration(context.Background(), "synthetic")
			o := configurationObject(t, result.Objects, "macro", "macro", nil)
			valid := mode == "true" || mode == "false" || mode == "missing-eval"
			if valid {
				if o.EvalBased == nil || *o.EvalBased != (mode == "true") || o.Document == nil {
					t.Fatal(o)
				}
			} else if o.Document != nil || !configurationHasDiagnostic(result.Diagnostics, "definition_unavailable", "macro") {
				t.Fatalf("unsafe macro metadata became expandable: %+v", o)
			}
			if len(*requests) != len(configurationFamilies) {
				t.Fatalf("optional eval metadata requested detail: %d", len(*requests))
			}
			if mode == "null-eval" && (o.EvalBased != nil || !configurationHasDiagnostic(result.Diagnostics, "macro_metadata_unavailable", "macro")) {
				t.Fatalf("explicit null became a default: %+v", o)
			}
			if !valid && mode != "null-eval" && !configurationHasDiagnostic(result.Diagnostics, "macro_metadata_invalid", "macro") {
				t.Fatalf("malformed metadata did not retain its gap: %+v", result.Diagnostics)
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationMacroInvalidArgumentsCannotBeRepairedByDetail(t *testing.T) {
	for _, name := range []string{"macro", "macro(1)"} {
		for _, args := range []any{nil, []string{"index"}, false, 1} {
			t.Run(fmt.Sprintf("%s/%T", name, args), func(t *testing.T) {
				c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
					if family == "configs/conf-macros" {
						// The missing body would otherwise require detail.
						configurationFeed(w, r, []map[string]any{configurationEntry(name, map[string]any{"args": args})})
						return
					}
					if strings.HasPrefix(family, "configs/conf-macros/") {
						json.NewEncoder(w).Encode(map[string]any{"entry": []any{configurationEntry(name, map[string]any{"args": "index", "definition": "search index=synthetic", "iseval": false})}})
						return
					}
					configurationFeed(w, r, nil)
				})
				result := c.collectConfiguration(context.Background(), "synthetic")
				if len(result.Objects) != 0 || !configurationHasDiagnostic(result.Diagnostics, "macro_arity_invalid", "macro") || collectionCoverage(result.Collections, "macro") != "unavailable" || len(*requests) != len(configurationFamilies) {
					t.Fatalf("invalid args repaired by detail: objects=%d requests=%d diagnostics=%+v", len(result.Objects), len(*requests), result.Diagnostics)
				}
				configurationBundle(t, c, result)
			})
		}
	}
}

func TestConfigurationMacroNecessaryDetailPreservesMetadataEvidence(t *testing.T) {
	for _, mode := range []string{"missing-positive-args", "missing-definition", "detail-eval", "null-eval", "malformed-eval"} {
		t.Run(mode, func(t *testing.T) {
			name := "macro"
			list := map[string]any{"args": ""}
			detail := map[string]any{"args": "", "definition": "search index=synthetic\n", "validation": ""}
			switch mode {
			case "missing-positive-args":
				name = "macro(1)"
				list = map[string]any{"definition": detail["definition"]}
				detail["args"] = "index"
			case "detail-eval":
				detail["iseval"] = true
			case "null-eval":
				list["iseval"] = nil
				detail["iseval"] = false
			case "malformed-eval":
				list["iseval"] = []bool{false}
				detail["iseval"] = false
			}
			c, requests := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
				if family == "configs/conf-macros/"+url.PathEscape(name) {
					json.NewEncoder(w).Encode(map[string]any{"entry": []any{configurationEntry(name, detail)}})
					return
				}
				if family == "configs/conf-macros" {
					configurationFeed(w, r, []map[string]any{configurationEntry(name, list)})
					return
				}
				configurationFeed(w, r, nil)
			})
			result := c.collectConfiguration(context.Background(), "synthetic")
			o := configurationObject(t, result.Objects, "macro", "macro", nil)
			if len(*requests) != len(configurationFamilies)+1 {
				t.Fatalf("necessary detail requests: %d", len(*requests))
			}
			if mode == "null-eval" || mode == "malformed-eval" {
				if o.EvalBased != nil || o.Document != nil {
					t.Fatalf("detail overwrote explicit metadata evidence: %+v", o)
				}
			} else if o.EvalBased == nil || *o.EvalBased != (mode == "detail-eval") || o.Document == nil || o.Document.Text != detail["definition"] || o.Validation == nil || *o.Validation != "" {
				t.Fatalf("necessary detail lost supported evidence: %+v", o)
			}
			if mode == "missing-positive-args" && (o.Arity == nil || *o.Arity != 1 || !reflect.DeepEqual(o.Arguments, []string{"index"})) {
				t.Fatalf("positive arity detail was not mapped: %+v", o)
			}
			configurationBundle(t, c, result)
		})
	}
}

func TestConfigurationDetailDoesNotOverwriteConflictingListEvidence(t *testing.T) {
	c, _ := newConfigurationFixture(t, func(w http.ResponseWriter, r *http.Request, family string) {
		if family == "configs/conf-macros/"+url.PathEscape("macro(1)") {
			json.NewEncoder(w).Encode(map[string]any{"entry": []any{configurationEntry("macro(1)", map[string]any{"definition": "index=other", "args": "index", "iseval": 0})}})
			return
		}
		if family != "configs/conf-macros" {
			configurationFeed(w, r, nil)
			return
		}
		configurationFeed(w, r, []map[string]any{configurationEntry("macro(1)", map[string]any{"definition": "index=main"}), configurationEntry("valid", map[string]any{"definition": "index=main", "args": "", "iseval": 0})})
	})
	result := c.collectConfiguration(context.Background(), "synthetic")
	if !reflect.DeepEqual(objectNames(result.Objects, "macro"), []string{"valid"}) || !configurationHasDiagnostic(result.Diagnostics, "definition_conflict", "macro") || collectionCoverage(result.Collections, "macro") != "partial" {
		t.Fatalf("conflicting detail trusted: %+v", result)
	}
	configurationBundle(t, c, result)
}
