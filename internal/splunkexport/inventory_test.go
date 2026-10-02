package splunkexport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// The fixture implements owned submission, status, bounded results and DELETE.
// Configured sourcetypes are deliberately available only on another endpoint.
type inventoryFixture struct {
	mu          sync.Mutex
	rows        func(url.Values) []map[string]any
	forms       []url.Values
	jobs        map[string]url.Values
	deleted     int
	warning     bool
	deleteFails bool
}

func newInventoryFixture(t *testing.T, rows func(url.Values) []map[string]any) (*Client, *inventoryFixture) {
	t.Helper()
	f := &inventoryFixture{rows: rows, jobs: map[string]url.Values{}}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == http.MethodPost && r.URL.Path == "/services/search/jobs" {
			if err := r.ParseForm(); err != nil {
				t.Error(err)
			}
			form := r.PostForm
			f.forms = append(f.forms, form)
			f.jobs[form.Get("id")] = form
			json.NewEncoder(w).Encode(map[string]any{"sid": form.Get("id")})
			return
		}
		if r.URL.Path == "/services/configs/conf-props" {
			json.NewEncoder(w).Encode(map[string]any{"entry": []any{map[string]any{"name": "configured-only"}}})
			return
		}
		path := strings.Split(r.URL.Path, "/")
		sid := path[len(path)-1]
		if strings.HasSuffix(r.URL.Path, "/results") {
			sid = path[len(path)-2]
		}
		form, found := f.jobs[sid]
		if !found {
			t.Errorf("unexpected endpoint or unowned job: %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method == http.MethodDelete {
			if f.deleteFails {
				http.Error(w, "sensitive detail", 500)
				return
			}
			delete(f.jobs, sid)
			f.deleted++
			w.WriteHeader(204)
			return
		}
		rows := f.rows(form)
		if rows == nil {
			rows = []map[string]any{}
		}
		if strings.HasSuffix(r.URL.Path, "/results") {
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			count, _ := strconv.Atoi(r.URL.Query().Get("count"))
			end := offset + count
			if end > len(rows) {
				end = len(rows)
			}
			json.NewEncoder(w).Encode(map[string]any{"preview": false, "init_offset": offset, "results": rows[offset:end]})
			return
		}
		response := map[string]any{"entry": []any{map[string]any{"content": map[string]any{"isDone": true, "resultCount": len(rows)}}}}
		if f.warning && strings.HasPrefix(form.Get("search"), "| rest ") {
			response["messages"] = []any{map[string]string{"type": "WARN", "text": "sensitive peer failure"}}
		}
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	all := true
	o.IndexSelection = environment.Selector{All: &all}
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	t.Cleanup(func() {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !f.deleteFails && len(f.jobs) != 0 {
			t.Errorf("fixture retained %d jobs", len(f.jobs))
		}
	})
	return c, f
}
func catalogFixture(index string, modes any) []map[string]any {
	return []map[string]any{{"index": index, "datatypes": modes}}
}
func indexRows(form url.Values) bool { return strings.HasPrefix(form.Get("search"), "| rest ") }
func objectNames(objects []environment.Object, kind string) []string {
	out := []string{}
	for _, o := range objects {
		if o.Kind == kind {
			out = append(out, o.Name)
		}
	}
	sort.Strings(out)
	return out
}
func collectionCoverage(collections []environment.Collection, kind string) string {
	for _, c := range collections {
		if c.Kind == kind {
			return c.Coverage
		}
	}
	return ""
}
func TestIndexDiscoveryDeduplicatesAndRetainsModes(t *testing.T) {
	c, f := newInventoryFixture(t, func(form url.Values) []map[string]any {
		return []map[string]any{{"index": "main", "datatypes": []string{"metric", "event", "event"}}, {"index": "main", "datatypes": "event"}, {"index": "archive", "datatypes": "event"}}
	})
	r := c.collectIndexes(context.Background(), "instance")
	if len(r.Objects) != 2 || len(r.Indexes) != 2 || r.Enumeration.Coverage != "complete" {
		t.Fatalf("catalog: %+v", r)
	}
	for _, i := range r.Indexes {
		for _, o := range r.Objects {
			if i.IndexID == o.ID && o.Name == "main" && !reflect.DeepEqual(i.CatalogDatatypes, []string{"event", "metric"}) {
				t.Fatalf("modes: %+v", i)
			}
		}
	}
	want := `| rest splunk_server=* /services/data/indexes datatype=all count=0 timeout=1 | stats values(datatype) as datatypes by title | rename title as index | fields index datatypes`
	if len(f.forms) != 1 || f.forms[0].Get("search") != want || f.deleted != 1 {
		t.Fatalf("protocol: %+v", f.forms)
	}
	for _, o := range r.Objects {
		if o.Namespace != "" || o.App != "" || o.Owner != "" || !strings.HasPrefix(o.ID, "sha256:") {
			t.Fatal(o)
		}
	}
	second := c.collectIndexes(context.Background(), "instance")
	if r.Objects[0].ID != second.Objects[0].ID {
		t.Fatal("identity depends on job")
	}
}
func TestMetadataPerIndexAssociation(t *testing.T) {
	c, f := newInventoryFixture(t, func(form url.Values) []map[string]any {
		if indexRows(form) {
			return []map[string]any{{"index": "main", "datatypes": []string{"event", "metric"}}, {"index": "archive", "datatypes": "event"}}
		}
		q := form.Get("search")
		kind := "source"
		if strings.Contains(q, "type=sourcetypes") {
			kind = "sourcetype"
		}
		if strings.Contains(q, `index="archive"`) && kind == "source" {
			return []map[string]any{{kind: "archive-only"}, {kind: "shared"}}
		}
		if kind == "sourcetype" {
			return []map[string]any{{kind: "observed-type", "host": "sensitive-host", "password": "secret"}}
		}
		if strings.Contains(q, "datatype=metric") {
			return []map[string]any{{kind: "metric-only"}, {kind: "shared"}}
		}
		return []map[string]any{{kind: "shared", "host": "sensitive-host"}, {kind: "shared"}}
	})
	indexes := c.collectIndexes(context.Background(), "instance")
	r := c.collectMetadata(context.Background(), "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if len(r.Captures) != 6 || len(r.Objects) != 4 || len(f.forms) != 7 {
		t.Fatalf("inventory: %+v forms=%d", r, len(f.forms))
	}
	if !reflect.DeepEqual(objectNames(r.Objects, "source"), []string{"archive-only", "metric-only", "shared"}) || !reflect.DeepEqual(objectNames(r.Objects, "sourcetype"), []string{"observed-type"}) {
		t.Fatal(r.Objects)
	}
	seenQueries := map[string]int{}
	for _, form := range f.forms[1:] {
		seenQueries[form.Get("search")]++
	}
	for _, mode := range []struct{ index, datatype string }{{"main", "event"}, {"main", "metric"}, {"archive", "event"}} {
		for _, kind := range []string{"source", "sourcetype"} {
			q := `| metadata type=` + kind + `s datatype=` + mode.datatype + ` index="` + mode.index + `" | fields ` + kind
			if seenQueries[q] != 1 {
				t.Fatalf("required fixed query missing or duplicated: %s", q)
			}
		}
	}
	ids := map[string]bool{}
	for _, o := range r.Objects {
		if ids[o.ID] {
			t.Fatal("duplicate object")
		}
		ids[o.ID] = true
	}
	for _, cap := range r.Captures {
		if cap.Coverage != "complete" || cap.ObjectIDs == nil {
			t.Fatal(cap)
		}
		if cap.Kind == "source" {
			for _, o := range indexes.Objects {
				if cap.IndexID == o.ID {
					want := 1
					if o.Name == "archive" || cap.Datatype == "metric" {
						want = 2
					}
					if len(cap.ObjectIDs) != want {
						t.Fatal(cap)
					}
				}
			}
		}
	}
	data, _ := json.Marshal(r)
	for _, s := range []string{"configured-only", "sensitive-host", "password", "secret", c.origin.Host, "spl-toolkit-export-"} {
		if strings.Contains(string(data), s) {
			t.Errorf("unexpected leaked field %q", s)
		}
	}
	for _, kind := range []string{"source", "sourcetype"} {
		if collectionCoverage(r.Collections, kind) != "complete" {
			t.Fatal(r.Collections)
		}
	}
}
func TestMetadataAllTimeAndExplicitBounds(t *testing.T) {
	for _, tc := range []struct {
		name, earliest, latest string
		historical             bool
	}{
		{"all time", "", "", true}, {"earliest only", "2020-01-01T00:00:00.123456789Z", "", false}, {"latest only", "", "2020-01-02T00:00:00.25Z", true}, {"paired", "2020-01-01T00:00:00.123456789Z", "2020-01-02T00:00:00.25Z", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Historical identity is independently known to be from 2019, outside the bounded pair.
			c, f := newInventoryFixture(t, func(form url.Values) []map[string]any {
				if indexRows(form) {
					return catalogFixture("main", []string{"event", "metric"})
				}
				if strings.Contains(form.Get("search"), "type=sourcetypes") {
					return []map[string]any{}
				}
				rows := []map[string]any{{"source": "current"}}
				if form.Get("earliest_time") == "0" {
					rows = append(rows, map[string]any{"source": "historical-2019"})
				}
				return rows
			})
			c.options.Window = environment.ObservationWindow{Earliest: tc.earliest, Latest: tc.latest}
			c.options.normalizeWindow()
			indexes := c.collectIndexes(context.Background(), "instance")
			r := c.collectMetadata(context.Background(), "instance", indexes)
			assertInventorySnapshot(t, c, indexes, r)
			names := objectNames(r.Objects, "source")
			if (len(names) == 2) != tc.historical {
				t.Fatalf("historical coverage: %v", names)
			}
			for _, form := range f.forms {
				earliest := "0"
				if tc.earliest != "" {
					earliest = "1577836800.123456789"
				}
				if form.Get("earliest_time") != earliest {
					t.Fatal(form)
				}
				latest := ""
				if tc.latest != "" {
					latest = "1577923200.25"
				}
				if form.Get("latest_time") != latest {
					t.Fatal(form)
				}
				if latest == "" {
					if _, present := form["latest_time"]; present {
						t.Fatal("absent latest sent")
					}
				}
			}
			for _, cap := range r.Captures {
				if cap.Kind == "sourcetype" && (cap.Coverage != "complete" || len(cap.ObjectIDs) != 0 || cap.ObjectIDs == nil) {
					t.Fatal(cap)
				}
			}
		})
	}
}
func TestRequestedIndexNotSilentlyDropped(t *testing.T) {
	c, _ := newInventoryFixture(t, func(form url.Values) []map[string]any {
		if indexRows(form) {
			return []map[string]any{{"index": "main", "datatypes": "event"}, {"index": "unselected", "datatypes": "metric"}}
		}
		return []map[string]any{}
	})
	selection, err := selector([]string{"missing-b", "main", "missing-a", "missing-a"})
	if err != nil {
		t.Fatal(err)
	}
	c.options.IndexSelection = selection
	indexes := c.collectIndexes(context.Background(), "instance")
	r := c.collectMetadata(context.Background(), "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if len(indexes.Objects) != 2 || len(indexes.Indexes) != 1 || !reflect.DeepEqual(indexes.UnmatchedIndexes, []string{"missing-a", "missing-b"}) || indexes.Enumeration.Coverage != "complete" {
		t.Fatal(indexes)
	}
	if len(r.Captures) != 2 || collectionCoverage(r.Collections, "source") != "partial" || collectionCoverage(r.Collections, "sourcetype") != "partial" {
		t.Fatal(r)
	}
}
func TestUnknownDatatypeRequiresBothModes(t *testing.T) {
	for _, modes := range []any{nil, "", []string{}, "unknown", []string{"event", "unsupported"}} {
		t.Run(fmt.Sprint(modes), func(t *testing.T) {
			c, _ := newInventoryFixture(t, func(form url.Values) []map[string]any {
				if indexRows(form) {
					return catalogFixture("main", modes)
				}
				return []map[string]any{}
			})
			indexes := c.collectIndexes(context.Background(), "instance")
			r := c.collectMetadata(context.Background(), "instance", indexes)
			assertInventorySnapshot(t, c, indexes, r)
			if len(indexes.Indexes) != 1 || indexes.Indexes[0].CatalogDatatypes == nil || len(indexes.Indexes[0].CatalogDatatypes) != 0 || !reflect.DeepEqual(indexes.Indexes[0].RequiredDatatypes, []string{"event", "metric"}) || len(indexes.Diagnostics) == 0 || len(r.Captures) != 4 {
				t.Fatalf("catalog=%+v metadata=%+v", indexes, r)
			}
			unreported := modes == nil
			if value, ok := modes.(string); ok && value == "" {
				unreported = true
			}
			if value, ok := modes.([]string); ok && len(value) == 0 {
				unreported = true
			}
			expectedCoverage := "partial"
			if unreported {
				expectedCoverage = "complete"
			}
			if indexes.Enumeration.Coverage != expectedCoverage || collectionCoverage(r.Collections, "source") != expectedCoverage {
				t.Fatalf("empty/unreported versus invalid modes: %+v %+v", indexes, r)
			}
		})
	}
}
func TestIndexLiteralRejectsQueryInjection(t *testing.T) {
	for _, name := range []string{`main" | rest /services/configs/conf-props`, `main*`, `a b","`, `a b`, `a` + "`", "main\n", "main\x00", ".main", "-main"} {
		t.Run(strconv.Quote(name), func(t *testing.T) {
			c, f := newInventoryFixture(t, func(form url.Values) []map[string]any { return catalogFixture(name, "event") })
			indexes := c.collectIndexes(context.Background(), "instance")
			r := c.collectMetadata(context.Background(), "instance", indexes)
			assertInventorySnapshot(t, c, indexes, r)
			if len(indexes.Objects) != 1 || len(r.Captures) != 2 || len(f.forms) != 1 {
				t.Fatalf("catalog=%+v metadata=%+v forms=%d", indexes, r, len(f.forms))
			}
			for _, cap := range r.Captures {
				if cap.Coverage != "unavailable" || cap.Reason != "index_literal_unsupported" || cap.ObjectIDs == nil {
					t.Fatal(cap)
				}
			}
		})
	}
}
func TestMetadataCapRetainsRowsAndMarksGap(t *testing.T) {
	c, _ := newInventoryFixture(t, func(form url.Values) []map[string]any {
		if indexRows(form) {
			return catalogFixture("main", "event")
		}
		if strings.Contains(form.Get("search"), "type=sourcetypes") {
			return []map[string]any{}
		}
		return []map[string]any{{"source": "first"}, {"source": "second"}, {"source": "omitted"}}
	})
	c.options.MaxRows = 2
	indexes := c.collectIndexes(context.Background(), "instance")
	r := c.collectMetadata(context.Background(), "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if !reflect.DeepEqual(objectNames(r.Objects, "source"), []string{"first", "second"}) || collectionCoverage(r.Collections, "source") != "partial" || collectionCoverage(r.Collections, "sourcetype") != "complete" {
		t.Fatal(r)
	}
	for _, cap := range r.Captures {
		if cap.Kind == "source" && (cap.Coverage != "partial" || cap.Reason != "job_row_limit" || len(cap.ObjectIDs) != 2) {
			t.Fatal(cap)
		}
	}
}
func TestMetadataInvalidRowsRetainOthers(t *testing.T) {
	c, _ := newInventoryFixture(t, func(form url.Values) []map[string]any {
		if indexRows(form) {
			return catalogFixture("main", "event")
		}
		if strings.Contains(form.Get("search"), "type=sourcetypes") {
			return []map[string]any{}
		}
		return []map[string]any{{"source": " exact ü "}, {"source": []string{"conflict-a", "conflict-b"}}, {"source": nil}, {"source": " exact ü "}, {"host": "host-only"}}
	})
	indexes := c.collectIndexes(context.Background(), "instance")
	r := c.collectMetadata(context.Background(), "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if !reflect.DeepEqual(objectNames(r.Objects, "source"), []string{" exact ü "}) || collectionCoverage(r.Collections, "source") != "partial" {
		t.Fatal(r)
	}
	if r.Objects[0].Name != " exact ü " || r.Objects[0].App != "" || r.Objects[0].Namespace != "" || r.Objects[0].Owner != "" {
		t.Fatal(r.Objects)
	}
}
func TestIndexAcquisitionGapPropagates(t *testing.T) {
	c, f := newInventoryFixture(t, func(form url.Values) []map[string]any {
		if indexRows(form) {
			return catalogFixture("main", "event")
		}
		return []map[string]any{}
	})
	f.warning = true
	indexes := c.collectIndexes(context.Background(), "instance")
	r := c.collectMetadata(context.Background(), "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if indexes.Enumeration.Coverage != "partial" || collectionCoverage(r.Collections, "source") != "partial" || collectionCoverage(r.Collections, "sourcetype") != "partial" {
		t.Fatal(indexes, r)
	}
}
func TestMetadataUnattemptedRequiredWorkAndCleanup(t *testing.T) {
	c, f := newInventoryFixture(t, func(form url.Values) []map[string]any { return catalogFixture("main", []string{"event", "metric"}) })
	indexes := c.collectIndexes(context.Background(), "instance")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := c.collectMetadata(ctx, "instance", indexes)
	assertInventorySnapshot(t, c, indexes, r)
	if len(r.Captures) != 4 || len(f.forms) != 1 || collectionCoverage(r.Collections, "source") != "unavailable" {
		t.Fatal(r)
	}
	for _, cap := range r.Captures {
		if cap.Coverage != "unavailable" || cap.Reason == "" || cap.ObjectIDs == nil || cap.Provenance.ObservedAt == "" {
			t.Fatal(cap)
		}
	}
	f.deleteFails = true
	indexes = c.collectIndexes(context.Background(), "instance")
	if !indexes.CleanupFailed {
		t.Fatal("cleanup failure lost")
	}
}
func TestIndexStableIdentityAndOpaqueProvenance(t *testing.T) {
	raw, _ := json.Marshal([]any{"instance", "source", " exact ü ", "", "", "", nil})
	sum := sha256.Sum256(raw)
	if got := stableObjectID("instance", "source", " exact ü ", "", "", "", nil); got != "sha256:"+hex.EncodeToString(sum[:]) {
		t.Fatal(got)
	}
	arity := 2
	if stableObjectID("instance", "macro", "m", "ns", "a", "o", &arity) == stableObjectID("instance", "macro", "m", "ns", "a", "o", nil) {
		t.Fatal("arity not in identity")
	}
	p := acquisitionProvenance("instance", "rest", "https://sensitive.example/jobs/secret", time.Now().UTC().Format(time.RFC3339Nano))
	if strings.Contains(p.SourceID, "sensitive") || !strings.HasPrefix(p.SourceID, "sha256:") {
		t.Fatal(p)
	}
}

func TestIndexEmptyAndInvalidCatalogCoverage(t *testing.T) {
	for _, tc := range []struct {
		name     string
		rows     []map[string]any
		coverage string
		objects  int
	}{
		{"complete empty", []map[string]any{}, "complete", 0},
		{"invalid only", []map[string]any{{"index": []string{"a", "b"}}}, "unavailable", 0},
		{"retain usable", []map[string]any{{"index": "main", "datatypes": "event"}, {"index": nil}}, "partial", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, _ := newInventoryFixture(t, func(form url.Values) []map[string]any {
				if indexRows(form) {
					return tc.rows
				}
				return []map[string]any{}
			})
			indexes := c.collectIndexes(context.Background(), "instance")
			metadata := c.collectMetadata(context.Background(), "instance", indexes)
			assertInventorySnapshot(t, c, indexes, metadata)
			if indexes.Enumeration.Coverage != tc.coverage || len(indexes.Objects) != tc.objects || indexes.Objects == nil || indexes.Indexes == nil || indexes.UnmatchedIndexes == nil {
				t.Fatal(indexes)
			}
			if collectionCoverage(metadata.Collections, "source") != tc.coverage || metadata.Objects == nil || metadata.Captures == nil {
				t.Fatal(metadata)
			}
		})
	}
}
func TestMetadataModeConstructorRejectsNonInventoryModes(t *testing.T) {
	for _, tc := range []struct{ kind, datatype string }{{"host", "event"}, {"source", "all"}, {"source | rest /services/configs/conf-props", "event"}} {
		if _, err := metadataDiscoveryQuery(tc.kind, tc.datatype, "main"); err == nil {
			t.Fatal("unsupported inventory mode accepted")
		}
	}
}
func TestIndexEndpointTimeoutConfigurationRoundsUp(t *testing.T) {
	for _, tc := range []struct {
		timeout time.Duration
		seconds string
	}{{30 * time.Second, "30"}, {1100 * time.Millisecond, "2"}} {
		c, f := newInventoryFixture(t, func(form url.Values) []map[string]any { return catalogFixture("main", "event") })
		c.options.RequestTimeout = tc.timeout
		indexes := c.collectIndexes(context.Background(), "instance")
		if indexes.Enumeration.Coverage != "complete" || !strings.Contains(f.forms[0].Get("search"), " timeout="+tc.seconds+" | stats ") {
			t.Fatal(indexes, f.forms)
		}
	}
}
func TestIndexStableIdentityNonMacroArityIsNull(t *testing.T) {
	arity := 2
	if stableObjectID("instance", "source", "name", "", "", "", &arity) != stableObjectID("instance", "source", "name", "", "", "", nil) {
		t.Fatal("non-macro identity must use null arity")
	}
}

// Assemble only inventory's real evidence; the other kinds are empty-complete
// fixture declarations so the v2 validator can distinguish valid from partial.
func assertInventorySnapshot(t *testing.T, c *Client, indexes IndexResult, metadata MetadataResult) {
	t.Helper()
	all := true
	allScope := environment.Selector{All: &all}
	collections := append([]environment.Collection{}, metadata.Collections...)
	collections = append(collections, environment.Collection{Kind: "index", Coverage: indexes.Enumeration.Coverage, Reason: indexes.Enumeration.Reason})
	for _, kind := range []string{"dataset", "data_model", "lookup", "macro", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module", "function", "external_command"} {
		collections = append(collections, environment.Collection{Kind: kind, Coverage: "complete"})
	}
	objects := append([]environment.Object{}, indexes.Objects...)
	objects = append(objects, metadata.Objects...)
	snapshot := environment.Snapshot{
		SchemaVersion: 2, ScopeID: "inventory-fixture",
		CaptureScope: environment.CaptureScope{Namespace: allScope, App: allScope, Owner: allScope},
		Origin:       environment.Origin{InstanceID: "instance", ProductVersion: "fixture", Producer: "fixture", ProducerVersion: "1"},
		Capture:      environment.CaptureInterval{Start: time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano), End: time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)},
		Capabilities: []environment.Capability{}, Collections: collections, Objects: objects,
		Observation: &environment.ObservationScope{
			IndexSelection: c.options.IndexSelection, Enumeration: indexes.Enumeration, Indexes: indexes.Indexes, UnmatchedIndexes: indexes.UnmatchedIndexes,
			Method: "splunk_metadata", Visibility: "exporting_principal", Window: c.options.Window, TimePrecision: "bucket_overlap", AbsenceMeaning: "not_observed", Captures: metadata.Captures,
		},
	}
	prepared, report, err := environment.PrepareSnapshot(snapshot)
	if err != nil || prepared == nil || report == nil || report.Status == "invalid" {
		t.Fatalf("inventory violates snapshot v2: %v %+v", err, report)
	}
	want := "valid"
	for _, collection := range collections {
		if collection.Coverage != "complete" {
			want = "partial"
		}
	}
	if report.Status != want {
		t.Fatalf("snapshot status=%s want=%s", report.Status, want)
	}
	normalized := prepared.Snapshot()
	if len(normalized.Objects) != len(objects) || len(normalized.Observation.Captures) != len(metadata.Captures) {
		t.Fatal("validation lost evidence")
	}
}
