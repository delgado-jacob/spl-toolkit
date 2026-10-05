package splunkexport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Fake routes follow Splunk search endpoint descriptions: caller-supplied id,
// /services/search/jobs status/control, and /services/search/v2/jobs results.
type jobFixture struct {
	mu              sync.Mutex
	jobs            map[string]bool
	operations      []string
	status          string
	statusSequence  []string
	results         string
	submissionDelay time.Duration
	unexpectedSID   bool
	deleteFails     bool
	rowCount        int
	maxRows         int
	probeOffset     int
	probeResponse   string
	probeStall      bool
	dispatchForm    url.Values
	onStatus        func()
}

func newJobFixture(t *testing.T, f *jobFixture) (*Client, *jobFixture) {
	t.Helper()
	f.jobs = map[string]bool{"unrelated-job": true}
	var c *Client
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == "POST" && r.URL.Path == "/services/search/jobs" {
			r.ParseForm()
			id := r.Form.Get("id")
			f.dispatchForm = r.PostForm
			if !strings.HasPrefix(id, "spl-toolkit-export-") || len(strings.TrimPrefix(id, "spl-toolkit-export-")) != 32 {
				t.Error("missing preallocated owned SID")
			}
			if !c.owns(id) {
				t.Error("SID not registered before submission")
			}
			f.jobs[id] = true
			f.operations = append(f.operations, "submit")
			for k, want := range map[string]string{"exec_mode": "normal", "search_mode": "normal", "enable_lookups": "false", "allow_partial_results": "true", "max_count": func() string {
				if f.maxRows > 0 {
					return strconv.Itoa(f.maxRows + 1)
				}
				if f.rowCount == 501 {
					return "601"
				}
				return "11"
			}()} {
				if r.Form.Get(k) != want {
					t.Errorf("%s=%q", k, r.Form.Get(k))
				}
			}
			for _, key := range []string{"max_time", "auto_cancel"} {
				seconds, err := strconv.ParseInt(r.Form.Get(key), 10, 64)
				if err != nil || seconds < 1 {
					t.Errorf("%s must be positive whole seconds: %q", key, r.Form.Get(key))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			if f.submissionDelay > 0 {
				f.mu.Unlock()
				select {
				case <-r.Context().Done():
				case <-time.After(f.submissionDelay):
				}
				f.mu.Lock()
			}
			if f.unexpectedSID {
				id = "unrelated-job"
			}
			fmt.Fprintf(w, `{"sid":%q}`, id)
			return
		}
		id := strings.Split(strings.TrimSuffix(r.URL.Path, "/control"), "/")
		sid := id[len(id)-1]
		if r.Method == "DELETE" {
			f.operations = append(f.operations, "delete")
			if f.deleteFails {
				w.WriteHeader(500)
				fmt.Fprint(w, "synthetic-secret")
				return
			}
			delete(f.jobs, sid)
			w.WriteHeader(204)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/control") {
			r.ParseForm()
			f.operations = append(f.operations, r.Form.Get("action"))
			fmt.Fprint(w, `{}`)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/results") {
			f.operations = append(f.operations, "results")
			if f.probeStall && r.URL.Query().Get("offset") == strconv.Itoa(f.probeOffset) {
				f.mu.Unlock()
				<-r.Context().Done()
				f.mu.Lock()
			}
			if f.probeResponse != "" && r.URL.Query().Get("offset") == strconv.Itoa(f.probeOffset) {
				fmt.Fprint(w, f.probeResponse)
				return
			}
			if f.results != "" {
				fmt.Fprint(w, f.results)
				return
			}
			count := f.rowCount
			if count == 0 {
				count = 1
			}
			offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
			limit, _ := strconv.Atoi(r.URL.Query().Get("count"))
			if offset+limit > count {
				limit = count - offset
			}
			rows := []map[string]string{}
			for i := 0; i < limit; i++ {
				rows = append(rows, map[string]string{"name": fmt.Sprintf("row-%d", offset+i)})
			}
			json.NewEncoder(w).Encode(map[string]any{"preview": false, "init_offset": offset, "results": rows})
			return
		}
		f.operations = append(f.operations, "status")
		if f.onStatus != nil {
			onStatus := f.onStatus
			f.onStatus = nil
			onStatus()
		}
		if len(f.statusSequence) > 0 {
			fmt.Fprint(w, f.statusSequence[0])
			f.statusSequence = f.statusSequence[1:]
		} else if f.status != "" {
			fmt.Fprint(w, f.status)
		} else {
			count := f.rowCount
			if count == 0 {
				count = 1
			}
			fmt.Fprintf(w, `{"entry":[{"content":{"isDone":"1","isFailed":0,"isFinalized":false,"isZombie":"0","resultCount":%d}}]}`, count)
		}
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	return c, f
}
func (f *jobFixture) assertClean(t *testing.T) {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.jobs["unrelated-job"] {
		t.Error("unrelated job deleted")
	}
	if !f.deleteFails && len(f.jobs) != 1 {
		t.Errorf("owned jobs remain: %d", len(f.jobs))
	}
}
func TestOwnedJobDispatchUsesIntegerRemainingBudget(t *testing.T) {
	for _, limit := range []string{"job", "overall", "parent"} {
		t.Run(limit, func(t *testing.T) {
			c, f := newJobFixture(t, &jobFixture{})
			c.options.JobTimeout = 10 * time.Second
			c.options.OverallTimeout = 10 * time.Second
			const budget = 2750 * time.Millisecond
			ctx := context.Background()
			switch limit {
			case "job":
				c.options.JobTimeout = budget
			case "overall":
				c.options.OverallTimeout = budget
			case "parent":
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, budget)
				t.Cleanup(cancel)
			}
			c.options.Window.Earliest = "2026-01-01T00:00:00.123456789Z"
			c.options.Window.Latest = "2026-01-02T00:00:00.987654321Z"
			r := c.runDiscoveryJob(ctx, discoveryQuery{search: "fixture query"})
			if r.Coverage != "complete" || len(r.Rows) != 1 || r.CleanupFailed {
				t.Fatalf("integer dispatch failed: %+v", r)
			}
			f.assertClean(t)
			f.mu.Lock()
			defer f.mu.Unlock()
			if f.dispatchForm.Get("max_time") != "2" || f.dispatchForm.Get("auto_cancel") != "3" {
				t.Fatalf("dispatch exceeded budget or lost cancellation bound: %v", f.dispatchForm)
			}
			if f.dispatchForm.Get("earliest_time") != "1767225600.123456789" || f.dispatchForm.Get("latest_time") != "1767312000.987654321" {
				t.Fatal("event-time fractions lost")
			}
		})
	}
}

func TestOwnedJobInsufficientBudgetDoesNotDispatch(t *testing.T) {
	for _, limit := range []string{"job", "overall", "parent", "subsecond-job", "cancelled-parent"} {
		t.Run(limit, func(t *testing.T) {
			c, f := newJobFixture(t, &jobFixture{})
			ctx := context.Background()
			switch limit {
			case "job":
				c.options.JobTimeout = 500 * time.Millisecond
			case "overall":
				c.options.OverallTimeout = 500 * time.Millisecond
			case "parent":
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
				t.Cleanup(cancel)
			case "subsecond-job":
				// Exactly one second can still dispatch before a coarse clock ticks.
				c.options.JobTimeout = 900 * time.Millisecond
			case "cancelled-parent":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			r := c.runDiscoveryJob(ctx, discoveryQuery{search: "fixture query"})
			if r.Coverage != "unavailable" || r.Reason != "job_timeout" || len(r.Rows) != 0 || r.CleanupFailed || len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != "job_timeout" {
				t.Fatalf("insufficient dispatch budget: %+v", r)
			}
			f.assertClean(t)
			f.mu.Lock()
			defer f.mu.Unlock()
			if len(f.operations) != 0 {
				t.Fatalf("unsubmitted job contacted server: %v", f.operations)
			}
			c.ownedMu.Lock()
			defer c.ownedMu.Unlock()
			if len(c.owned) != 0 {
				t.Fatalf("unsubmitted owned jobs: %d", len(c.owned))
			}
		})
	}
}

func TestOwnedJobSuccessDeletes(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage != "complete" || len(r.Rows) != 1 || r.CleanupFailed {
		t.Fatalf("result: %+v", r)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.Join(f.operations, ",") != "submit,status,results,delete" {
		t.Fatal(f.operations)
	}
}
func TestOwnedJobFailureRetainsWarning(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":true,"isFailed":true,"resultCount":1}}],"messages":[{"type":"WARN","text":"synthetic-secret warning"}]}`})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" || len(r.Diagnostics) == 0 {
		t.Fatalf("result: %+v", r)
	}
	data, _ := json.Marshal(r.Diagnostics)
	if strings.Contains(string(data), "synthetic-secret") {
		t.Fatal("warning leaked")
	}
	f.assertClean(t)
}
func TestOwnedJobDeadlineFinalizesBeforeDelete(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":false,"resultCount":1}}]}`})
	c.options.JobTimeout = 1100 * time.Millisecond
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage != "partial" || len(r.Rows) != 1 {
		t.Fatalf("salvage: %+v", r)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	ops := strings.Join(f.operations, ",")
	if !strings.Contains(ops, "finalize,status,results,delete") {
		t.Fatal(ops)
	}
}
func TestOwnedJobSubmissionTimeoutCleansKnownID(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{submissionDelay: 50 * time.Millisecond})
	c.options.RequestTimeout = 10 * time.Millisecond
	c.http.Timeout = 10 * time.Millisecond
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" {
		t.Fatal("submission timeout complete")
	}
	f.assertClean(t)
}
func TestOwnedJobCancellationLeavesUnrelatedJob(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{unexpectedSID: true})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" {
		t.Fatal("unexpected SID accepted")
	}
	f.assertClean(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	c2, f2 := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":false,"resultCount":1}}]}`, onStatus: cancel})
	r = c2.runDiscoveryJob(ctx, discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" {
		t.Fatal("cancelled job complete")
	}
	f2.assertClean(t)
	f2.mu.Lock()
	defer f2.mu.Unlock()
	ops := strings.Join(f2.operations, ",")
	if !strings.HasPrefix(ops, "submit,status,") || !strings.HasSuffix(ops, "finalize,status,results,delete") {
		t.Fatalf("cancellation did not salvage dispatched job: %s", ops)
	}
}
func TestOwnedJobGapsAndPagination(t *testing.T) {
	for _, tc := range []struct {
		name, status, results string
		count                 int
	}{{name: "preview", results: `{"preview":true,"init_offset":0,"results":[{"name":"retained"}]}`}, {name: "offset", results: `{"preview":false,"init_offset":9,"results":[{"name":"bad"}]}`}, {name: "malformed-status", status: `{"entry":[{"content":{"isDone":"banana","resultCount":1}}]}`}, {name: "warning", status: `{"entry":[{"content":{"isDone":true,"resultCount":1}}],"messages":[{"type":"WARN","text":"synthetic-secret"}]}`}, {name: "finalized", status: `{"entry":[{"content":{"isDone":true,"isFinalized":true,"resultCount":1}}]}`}, {name: "cap", count: 11}} {
		t.Run(tc.name, func(t *testing.T) {
			c, f := newJobFixture(t, &jobFixture{status: tc.status, results: tc.results, rowCount: tc.count})
			r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
			if r.Coverage == "complete" || len(r.Rows) > 10 {
				t.Fatalf("gap accepted: %+v", r)
			}
			f.assertClean(t)
		})
	}
	t.Run("pages", func(t *testing.T) {
		c, f := newJobFixture(t, &jobFixture{rowCount: 501})
		c.options.MaxRows = 600
		r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
		if r.Coverage != "complete" || len(r.Rows) != 501 {
			t.Fatalf("pages: %+v", r)
		}
		f.assertClean(t)
	})
}
func TestOwnedJobCleanupFailureReturned(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{deleteFails: true})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if !r.CleanupFailed || len(r.Diagnostics) == 0 {
		t.Fatal("cleanup failure lost")
	}
	f.assertClean(t)
}
func TestOwnedJobEmptyResultsChecksPreview(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":true,"resultCount":0}}]}`, results: `{"preview":true,"init_offset":0,"results":[]}`})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" {
		t.Fatal("empty preview established complete coverage")
	}
	f.assertClean(t)
}
func TestOwnedJobDispatchFailure(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":true,"dispatchState":"FAILED","resultCount":1}}]}`})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" {
		t.Fatal("dispatch failure accepted")
	}
	f.assertClean(t)
}
func TestOwnedJobExtraRowsDetected(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{results: `{"preview":false,"init_offset":0,"results":[{"name":"one"},{"name":"unexpected"}]}`})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage == "complete" || len(r.Rows) > 1 {
		t.Fatalf("count mismatch: %+v", r)
	}
	f.assertClean(t)
}
func TestOwnedJobRequestDeadlineSalvages(t *testing.T) {
	var mu sync.Mutex
	var sid string
	var ops []string
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == "DELETE":
			ops = append(ops, "delete")
			w.WriteHeader(404)
		case r.URL.Path == "/services/search/jobs":
			r.ParseForm()
			sid = r.Form.Get("id")
			fmt.Fprintf(w, `{"sid":%q}`, sid)
		case strings.HasSuffix(r.URL.Path, "/control"):
			ops = append(ops, "finalize")
			fmt.Fprint(w, `{}`)
		case strings.HasSuffix(r.URL.Path, "/results"):
			ops = append(ops, "results")
			fmt.Fprint(w, `{"preview":false,"init_offset":0,"results":[{"name":"salvaged"}]}`)
		default:
			ops = append(ops, "status")
			if len(ops) == 1 {
				mu.Unlock()
				<-r.Context().Done()
				mu.Lock()
				return
			}
			fmt.Fprint(w, `{"entry":[{"content":{"isDone":true,"resultCount":1}}]}`)
		}
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	// Leave headroom for TLS submission; only the first status request stalls.
	o.RequestTimeout = 500 * time.Millisecond
	o.CAFile = fixtureCA(t, s)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture"})
	if r.Coverage != "partial" || len(r.Rows) != 1 || r.CleanupFailed {
		t.Fatalf("request salvage: %+v", r)
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(ops, ",") != "status,finalize,status,results,delete" {
		t.Fatal(ops)
	}
}
func TestOwnedJobEpochBounds(t *testing.T) {
	for _, tc := range []struct{ input, want string }{{"2026-01-01T00:00:00.123456789Z", "1767225600.123456789"}, {"1969-12-31T23:59:59.5Z", "-0.5"}} {
		if got := epochBound(tc.input); got != tc.want {
			t.Fatalf("epoch %s: %s", tc.input, got)
		}
	}
}
func TestOwnedJobNumericVariants(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":1.00,"isFailed":"0.0","resultCount":1.0}}]}`})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture"})
	if r.Coverage != "complete" || len(r.Rows) != 1 {
		t.Fatalf("numeric variants: %+v", r)
	}
	f.assertClean(t)
}
func TestOwnedJobSalvageReservesSharedCleanupBudget(t *testing.T) {
	var mu sync.Mutex
	var operations []string
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Method == "DELETE" {
			operations = append(operations, "delete")
			w.WriteHeader(204)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/control") {
			r.ParseForm()
			operations = append(operations, "finalize")
			mu.Unlock()
			<-r.Context().Done()
			mu.Lock()
			return
		}
		operations = append(operations, "status")
		fmt.Fprint(w, `{"entry":[{"content":{"isDone":false,"resultCount":0}}]}`)
	}))
	t.Cleanup(s.Close)
	o := testOptions(t, s.URL)
	o.CAFile = fixtureCA(t, s)
	c, err := NewClient(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(c.Close)
	sid := "spl-toolkit-export-fixture"
	c.owned[sid] = true
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	t.Cleanup(cancel)
	result := JobResult{Coverage: "complete"}
	started := time.Now()
	c.salvage(ctx, sid, &result)
	if ctx.Err() != nil {
		t.Fatal("salvage consumed entire cleanup budget")
	}
	if err = c.deleteOwnedJob(ctx, sid); err != nil {
		t.Fatalf("reserved DELETE failed: %v", err)
	}
	if time.Since(started) > 100*time.Millisecond {
		t.Fatal("cleanup exceeded shared budget")
	}
	mu.Lock()
	defer mu.Unlock()
	if strings.Join(operations, ",") != "finalize,delete" {
		t.Fatal(operations)
	}
}

func TestOwnedJobFullFinalPageCountConsistency(t *testing.T) {
	for _, tc := range []struct {
		name                      string
		reported, actual, maxRows int
		complete                  bool
	}{{"500-complete", 500, 500, 600, true}, {"500-underreported", 500, 501, 600, false}, {"1000-complete", 1000, 1000, 1100, true}, {"1000-underreported", 1000, 1001, 1100, false}, {"at-row-cap-underreported", 500, 501, 500, false}} {
		t.Run(tc.name, func(t *testing.T) {
			c, f := newJobFixture(t, &jobFixture{maxRows: tc.maxRows, rowCount: tc.actual, status: fmt.Sprintf(`{"entry":[{"content":{"isDone":true,"resultCount":%d}}]}`, tc.reported)})
			c.options.MaxRows = tc.maxRows
			r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture"})
			if (r.Coverage == "complete") != tc.complete || len(r.Rows) != tc.reported {
				t.Fatalf("full final page: coverage=%s reason=%s rows=%d", r.Coverage, r.Reason, len(r.Rows))
			}
			if !tc.complete && (r.Reason == "" || len(r.Diagnostics) == 0) {
				t.Fatal("count mismatch was not diagnosed")
			}
			f.assertClean(t)
			f.mu.Lock()
			defer f.mu.Unlock()
			pages := 0
			for _, op := range f.operations {
				if op == "results" {
					pages++
				}
			}
			if pages != tc.reported/resultPageSize+1 {
				t.Fatalf("full-page end probe omitted: %d requests", pages)
			}
		})
	}
}
func TestOwnedJobEndProbeValidatesPageEvidence(t *testing.T) {
	for _, tc := range []struct{ name, body string }{{"preview", `{"preview":true,"init_offset":500,"results":[]}`}, {"offset", `{"preview":false,"init_offset":0,"results":[]}`}, {"warning", `{"preview":false,"init_offset":500,"results":[],"messages":[{"type":"WARN","text":"synthetic-secret"}]}`}, {"extra-limit", `{"preview":false,"init_offset":500,"results":[{"name":"one"},{"name":"two"}]}`}} {
		t.Run(tc.name, func(t *testing.T) {
			c, f := newJobFixture(t, &jobFixture{maxRows: 600, rowCount: 500, probeOffset: 500, probeResponse: tc.body})
			c.options.MaxRows = 600
			r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture"})
			if r.Coverage != "partial" || len(r.Rows) != 500 {
				t.Fatalf("probe evidence ignored: coverage=%s reason=%s rows=%d", r.Coverage, r.Reason, len(r.Rows))
			}
			data, _ := json.Marshal(r.Diagnostics)
			if strings.Contains(string(data), "synthetic-secret") {
				t.Fatal("probe message leaked")
			}
			f.assertClean(t)
		})
	}
}

func TestOwnedJobEndProbeDeadlineRetainsRows(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{maxRows: 600, rowCount: 500, probeOffset: 500, probeStall: true})
	c.options.MaxRows = 600
	// Leave headroom for submission and the first results page; the probe stalls.
	c.http.Timeout = 500 * time.Millisecond
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture"})
	if r.Coverage != "partial" || len(r.Rows) != 500 || r.CleanupFailed {
		t.Fatalf("probe timeout: coverage=%s reason=%s rows=%d cleanup=%v", r.Coverage, r.Reason, len(r.Rows), r.CleanupFailed)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.Join(f.operations, ",") != "submit,status,results,results,finalize,status,delete" {
		t.Fatalf("probe repeated during salvage or deletion misplaced: %v", f.operations)
	}
}

func TestOwnedJobStartupStatusesBeforeResultCount(t *testing.T) {
	startup := `{"entry":[{"content":{"isDone":false,"isFailed":false,"isFinalized":false,"isZombie":false,"dispatchState":"PARSING","messages":[]}}]}`
	c, f := newJobFixture(t, &jobFixture{statusSequence: []string{
		startup, startup,
		`{"entry":[{"content":{"isDone":true,"dispatchState":"DONE","resultCount":1}}]}`,
	}})
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage != "complete" || len(r.Rows) != 1 || len(r.Diagnostics) != 0 || r.CleanupFailed {
		t.Fatalf("startup polling: %+v", r)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.Join(f.operations, ",") != "submit,status,status,status,results,delete" {
		t.Fatalf("results read before terminal count or startup polling stopped: %v", f.operations)
	}
}

func TestJobStatusAllowsAbsentCountOnlyInExplicitActiveStates(t *testing.T) {
	for _, tc := range []struct {
		name, dispatch string
		valid          bool
	}{
		{"queued", `,"dispatchState":"QUEUED"`, true},
		{"parsing", `,"dispatchState":"PARSING"`, true},
		{"running", `,"dispatchState":"RUNNING"`, true},
		{"finalizing", `,"dispatchState":"FINALIZING"`, true},
		{"pause", `,"dispatchState":"PAUSE"`, true},
		{"absent", "", false},
		{"unknown", `,"dispatchState":"UNKNOWN"`, false},
		{"done", `,"dispatchState":"DONE"`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"entry":[{"content":{"isDone":false,"isFailed":false,"isFinalized":false,"isZombie":false%s}}]}`, tc.dispatch)
			c, _ := newJobFixture(t, &jobFixture{status: body})
			state, err := c.jobStatus(context.Background(), "fixture", &JobResult{Coverage: "complete"})
			if (err == nil) != tc.valid || tc.valid && (state.done || state.failed || state.finalized || state.zombie) {
				t.Fatalf("active status: state=%+v err=%v", state, err)
			}
		})
	}
}

func TestOwnedJobRejectsInvalidTerminalCounts(t *testing.T) {
	for _, terminal := range []struct{ name, fields string }{
		{"done", `"isDone":true,"dispatchState":"PARSING"`},
		{"failed", `"isDone":false,"isFailed":true,"dispatchState":"PARSING"`},
		{"finalized", `"isDone":false,"isFinalized":true,"dispatchState":"PARSING"`},
		{"zombie", `"isDone":false,"isZombie":true,"dispatchState":"PARSING"`},
		{"dispatch-failed", `"isDone":false,"dispatchState":"FAILED"`},
		{"dispatch-internal-error", `"isDone":false,"dispatchState":"INTERNAL_ERROR"`},
		{"dispatch-bad-input", `"isDone":false,"dispatchState":"BAD_INPUT_CANCEL"`},
		{"dispatch-finalized", `"isDone":false,"dispatchState":"FINALIZED"`},
		{"dispatch-zombie", `"isDone":false,"dispatchState":"ZOMBIE"`},
	} {
		for _, count := range []struct{ name, field string }{
			{"absent", ""},
			{"null", `,"resultCount":null`},
			{"malformed", `,"resultCount":"banana"`},
			{"negative", `,"resultCount":-1`},
			{"fractional", `,"resultCount":1.5`},
			{"overflow", `,"resultCount":9223372036854775808`},
		} {
			t.Run(terminal.name+"/"+count.name, func(t *testing.T) {
				body := fmt.Sprintf(`{"entry":[{"content":{%s%s}}]}`, terminal.fields, count.field)
				c, f := newJobFixture(t, &jobFixture{status: body})
				r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
				if r.Coverage != "unavailable" || r.Reason != "job_status_invalid" || len(r.Rows) != 0 || r.CleanupFailed {
					t.Fatalf("invalid terminal count: %+v", r)
				}
				f.assertClean(t)
				f.mu.Lock()
				defer f.mu.Unlock()
				if strings.Join(f.operations, ",") != "submit,status,delete" {
					t.Fatalf("invalid count retrieved results or missed cleanup: %v", f.operations)
				}
			})
		}
	}
}

func TestJobStatusRejectsPresentInvalidActiveCounts(t *testing.T) {
	for _, count := range []string{"null", `"banana"`, "-1", "1.5", "9223372036854775808"} {
		t.Run(count, func(t *testing.T) {
			body := fmt.Sprintf(`{"entry":[{"content":{"isDone":false,"dispatchState":"PARSING","resultCount":%s}}]}`, count)
			c, _ := newJobFixture(t, &jobFixture{status: body})
			if _, err := c.jobStatus(context.Background(), "fixture", &JobResult{Coverage: "complete"}); err == nil {
				t.Fatal("present invalid active count accepted")
			}
		})
	}
}

func TestOwnedJobStartupTimeoutSkipsUnknownCountResults(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":false,"dispatchState":"PARSING"}}]}`})
	c.options.JobTimeout = 1100 * time.Millisecond
	r := c.runDiscoveryJob(context.Background(), discoveryQuery{search: "fixture query"})
	if r.Coverage != "unavailable" || r.Reason != "job_timeout" || len(r.Rows) != 0 || r.CleanupFailed || len(r.Diagnostics) != 1 || r.Diagnostics[0].Code != "job_timeout" {
		t.Fatalf("startup timeout: %+v", r)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.dispatchForm.Get("max_time") != "1" {
		t.Fatalf("startup timeout did not dispatch with integer budget: %v", f.dispatchForm)
	}
	ops := strings.Join(f.operations, ",")
	if !strings.HasPrefix(ops, "submit,status,status,") || !strings.HasSuffix(ops, "finalize,status,delete") || strings.Contains(ops, "results") {
		t.Fatalf("unknown count retrieved results or missed timeout cleanup: %s", ops)
	}
}

func TestJobSalvageUnknownCountPreservesPartialEvidence(t *testing.T) {
	c, f := newJobFixture(t, &jobFixture{status: `{"entry":[{"content":{"isDone":false,"dispatchState":"RUNNING"}}]}`})
	sid := "spl-toolkit-export-fixture"
	c.owned[sid] = true
	f.mu.Lock()
	f.jobs[sid] = true
	f.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	t.Cleanup(cancel)
	r := JobResult{Coverage: "complete", Rows: []map[string]json.RawMessage{{"name": json.RawMessage(`"retained"`)}}}
	r.gap("job_warning")
	originalDiagnostic := r.Diagnostics[0]
	c.salvage(ctx, sid, &r)
	if err := c.deleteOwnedJob(ctx, sid); err != nil {
		t.Fatal(err)
	}
	if r.Coverage != "partial" || r.Reason != "job_warning" || len(r.Rows) != 1 || string(r.Rows[0]["name"]) != `"retained"` || len(r.Diagnostics) != 2 || r.Diagnostics[0] != originalDiagnostic || r.Diagnostics[1].Code != "job_timeout" {
		t.Fatalf("unknown-count salvage lost existing evidence or added invalid-status gap: %+v", r)
	}
	f.assertClean(t)
	f.mu.Lock()
	defer f.mu.Unlock()
	if strings.Join(f.operations, ",") != "finalize,status,delete" {
		t.Fatalf("unknown-count salvage read results: %v", f.operations)
	}
}
