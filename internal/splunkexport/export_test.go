package splunkexport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

type exportFixture struct {
	mu                                      sync.Mutex
	guid, version                           string
	originFails, deleteFails, metadataFails bool
	paths                                   []string
	jobs                                    map[string]url.Values
}

func newExportFixture(t *testing.T) (Options, *exportFixture) {
	t.Helper()
	f := &exportFixture{guid: "private-guid", version: "9.4.2", jobs: map[string]url.Values{}}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		f.paths = append(f.paths, r.URL.Path)
		if r.URL.Path == "/services/server/info" {
			if f.originFails {
				http.Error(w, "private failure", 403)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"entry": []any{map[string]any{"content": map[string]any{"version": f.version, "guid": f.guid, "serverName": "private-host"}}}})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/servicesNS/") {
			family := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/servicesNS/"), "/", 3)[2]
			if family == "saved/eventtypes" {
				http.Error(w, "private failure", 403)
				return
			}
			var entries []map[string]any
			if family == "saved/searches" {
				entries = []map[string]any{configurationEntry("daily", map[string]any{"search": "index=main | stats count"})}
			}
			configurationFeed(w, r, entries)
			return
		}
		if r.Method == "POST" && r.URL.Path == "/services/search/jobs" {
			r.ParseForm()
			f.jobs[r.PostForm.Get("id")] = r.PostForm
			json.NewEncoder(w).Encode(map[string]any{"sid": r.PostForm.Get("id")})
			return
		}
		parts := strings.Split(r.URL.Path, "/")
		sid := parts[len(parts)-1]
		if sid == "results" {
			sid = parts[len(parts)-2]
		}
		form, ok := f.jobs[sid]
		if !ok {
			t.Errorf("unexpected endpoint %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Method == "DELETE" {
			if f.deleteFails {
				http.Error(w, "private failure", 500)
			} else {
				fmt.Fprint(w, `{}`)
			}
			return
		}
		rows := []map[string]any{{"index": "main", "datatypes": "event"}}
		if !indexRows(form) {
			kind := "source"
			if strings.Contains(form.Get("search"), "type=sourcetypes") {
				kind = "sourcetype"
			}
			rows = []map[string]any{{kind: "observed"}}
			if f.metadataFails {
				http.Error(w, "private failure", 403)
				return
			}
		}
		if strings.HasSuffix(r.URL.Path, "/results") {
			json.NewEncoder(w).Encode(map[string]any{"preview": false, "init_offset": 0, "results": rows})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"entry": []any{map[string]any{"content": map[string]any{"isDone": true, "resultCount": len(rows)}}}})
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	all := true
	o.Scope = environment.CaptureScope{Namespace: environment.Selector{All: &all}, App: environment.Selector{All: &all}, Owner: environment.Selector{All: &all}}
	o.IndexSelection = environment.Selector{All: &all}
	return o, f
}
func assertExportValid(t *testing.T, r *Result) {
	t.Helper()
	raw, err := json.Marshal(r.Snapshot)
	if err != nil {
		t.Fatal(err)
	}
	validation, err := environment.ValidateArtifacts(raw, nil)
	if err != nil || validation == nil || validation.Status == "invalid" {
		t.Fatalf("invalid exported artifact: %v %+v", err, validation)
	}
	if validation.SnapshotDigest != r.Snapshot.Digest || r.Report.SnapshotDigest != r.Snapshot.Digest || r.Report.ScopeID != r.Snapshot.ScopeID {
		t.Fatal("digest/scope disagreement")
	}
}
func TestExportRetainsSuccessfulCollections(t *testing.T) {
	o, f := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	if len(r.Snapshot.Collections) != 15 || collectionCoverage(r.Snapshot.Collections, "saved_search") != "complete" || collectionCoverage(r.Snapshot.Collections, "event_type") != "unavailable" || len(objectNames(r.Snapshot.Objects, "saved_search")) != 1 || r.ExitCode != 3 {
		t.Fatalf("lost successful evidence: %+v", r)
	}
	raw, _ := json.Marshal(r)
	for _, private := range []string{f.guid, "private-host", o.ManagementURL, "synthetic-secret"} {
		if strings.Contains(string(raw), private) {
			t.Fatalf("private origin field exported: %s", private)
		}
	}
	if !strings.HasPrefix(r.Snapshot.ScopeID, "capture:") || len(r.Snapshot.ScopeID) != 40 {
		t.Fatal(r.Snapshot.ScopeID)
	}
}
func TestExportNoConfiguredSourcetypes(t *testing.T) {
	o, f := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	if got := objectNames(r.Snapshot.Objects, "sourcetype"); len(got) != 1 || got[0] != "observed" {
		t.Fatal(got)
	}
	for _, path := range f.paths {
		if strings.Contains(path, "conf-props") {
			t.Fatal("queried configured sourcetypes")
		}
	}
}
func TestExportUnavailableOriginEmitsNothing(t *testing.T) {
	o, f := newExportFixture(t)
	f.originFails = true
	r, err := Export(context.Background(), o)
	if err == nil || r != nil {
		t.Fatalf("origin accepted: %+v %v", r, err)
	}
	if len(f.paths) != 1 {
		t.Fatal(f.paths)
	}
}
func TestExportCleanupFailureKeepsSnapshot(t *testing.T) {
	o, f := newExportFixture(t)
	f.deleteFails = true
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	if r.ExitCode != 3 || r.Report.Status != "partial" || len(objectNames(r.Snapshot.Objects, "index")) != 1 {
		t.Fatal(r)
	}
	if !configurationHasDiagnostic(r.Report.Diagnostics, "job_cleanup_failed", "") {
		t.Fatal("cleanup failure missing")
	}
}
func TestExportGUIDFallbackAndMissingIdentity(t *testing.T) {
	o, f := newExportFixture(t)
	f.guid = ""
	o.InstanceID = "opaque-user-id"
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	if r.Snapshot.Origin.InstanceID != o.InstanceID {
		t.Fatal(r.Snapshot.Origin)
	}
	o.InstanceID = " "
	if r, err = Export(context.Background(), o); err == nil || r != nil {
		t.Fatal("missing identity accepted")
	}
}
func TestExportRetainsUnavailableMetadataCaptures(t *testing.T) {
	o, f := newExportFixture(t)
	f.metadataFails = true
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	if len(r.Snapshot.Observation.Captures) != 2 {
		t.Fatal(r.Snapshot.Observation)
	}
	for _, capture := range r.Snapshot.Observation.Captures {
		if capture.Coverage != "unavailable" || capture.Reason == "" {
			t.Fatal(capture)
		}
	}
}

func TestExportRequiresProductVersion(t *testing.T) {
	o, f := newExportFixture(t)
	f.version = " "
	r, err := Export(context.Background(), o)
	if err == nil || r != nil {
		t.Fatal("missing product version accepted")
	}
	if len(f.paths) != 1 {
		t.Fatal("acquired evidence without usable origin")
	}
}
func TestExportAcquisitionOrder(t *testing.T) {
	o, f := newExportFixture(t)
	r, err := Export(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	assertExportValid(t, r)
	configurationStarted, configurationFinished := false, false
	for _, path := range f.paths {
		if strings.HasPrefix(path, "/servicesNS/") {
			if configurationFinished {
				t.Fatal("configuration after metadata")
			}
			configurationStarted = true
		}
		if path == "/services/search/jobs" && configurationStarted {
			configurationFinished = true
		}
	}
	if !configurationStarted || !configurationFinished || f.paths[0] != "/services/server/info" {
		t.Fatal("missing acquisition phase")
	}
}
