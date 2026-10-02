package splunkexport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// discoveryQuery stays private. Only fixed index and metadata constructors may
// create production queries; this package has no public arbitrary-search API.
type discoveryQuery struct{ search string }

type JobResult struct {
	Rows                         []map[string]json.RawMessage
	Coverage, Reason, ObservedAt string
	Diagnostics                  []Diagnostic
	CleanupFailed                bool
}

func (r *JobResult) gap(code string) {
	r.Coverage = "partial"
	if len(r.Rows) == 0 {
		r.Coverage = "unavailable"
	}
	if r.Reason == "" {
		r.Reason = code
	}
	for _, d := range r.Diagnostics {
		if d.Code == code {
			return
		}
	}
	r.Diagnostics = append(r.Diagnostics, Diagnostic{Code: code, Severity: "warning", Message: "Discovery acquisition did not establish complete coverage."})
}
func (c *Client) owns(sid string) bool {
	c.ownedMu.Lock()
	defer c.ownedMu.Unlock()
	return c.owned[sid]
}
func (c *Client) deleteOwnedJob(ctx context.Context, sid string) error {
	if !c.owns(sid) {
		return failure("job_not_owned")
	}
	err := c.requestJSON(ctx, http.MethodDelete, "/services/search/jobs/"+url.PathEscape(sid), nil, nil, nil)
	var e *requestFailure
	if errors.As(err, &e) && e.status == http.StatusNotFound {
		err = nil
	}
	if err == nil {
		c.ownedMu.Lock()
		delete(c.owned, sid)
		c.ownedMu.Unlock()
	}
	return err
}
func epochBound(value string) string {
	t, _ := time.Parse(time.RFC3339Nano, value)
	base := strconv.FormatInt(t.Unix(), 10)
	if t.Nanosecond() == 0 {
		return base
	}
	// Splunk accepts epoch seconds with fractions. For pre-epoch times, Unix()
	// floors; keep the sign and fractional portion mathematically correct.
	if t.Unix() < 0 {
		seconds := -t.Unix() - 1
		fraction := 1_000_000_000 - t.Nanosecond()
		return "-" + strconv.FormatInt(seconds, 10) + "." + strings.TrimRight(leftNine(fraction), "0")
	}
	return base + "." + strings.TrimRight(leftNine(t.Nanosecond()), "0")
}
func leftNine(n int) string { s := strconv.Itoa(n); return strings.Repeat("0", 9-len(s)) + s }
func (c *Client) runDiscoveryJob(parent context.Context, query discoveryQuery) (result JobResult) {
	result.Rows = []map[string]json.RawMessage{}
	result.Diagnostics = []Diagnostic{}
	result.Coverage = "complete"
	result.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
	overall, cancelOverall := context.WithTimeout(parent, c.options.OverallTimeout)
	defer cancelOverall()
	ctx, cancel := context.WithTimeout(overall, c.options.JobTimeout)
	defer cancel()
	deadline, _ := ctx.Deadline()
	remaining := time.Until(deadline)
	// Splunk requires whole seconds; zero removes the server time limit.
	// Check before allocating an owned SID so unsubmitted jobs need no cleanup.
	maxTime := int64(remaining / time.Second)
	if ctx.Err() != nil || maxTime < 1 {
		result.gap("job_timeout")
		return
	}
	seconds := int64(math.Ceil(remaining.Seconds()))
	entropy := make([]byte, 16)
	if _, err := rand.Read(entropy); err != nil {
		result.gap("job_id_failed")
		return
	}
	sid := "spl-toolkit-export-" + hex.EncodeToString(entropy)
	c.ownedMu.Lock()
	c.owned[sid] = true
	c.ownedMu.Unlock()
	var cleanupContext context.Context
	var cleanupCancel context.CancelFunc
	cleanup := func() context.Context {
		if cleanupContext == nil {
			cleanupContext, cleanupCancel = context.WithTimeout(context.Background(), cleanupTimeout)
		}
		return cleanupContext
	}
	defer func() {
		ctx := cleanup()
		defer cleanupCancel()
		if c.deleteOwnedJob(ctx, sid) != nil {
			result.CleanupFailed = true
			result.Diagnostics = append(result.Diagnostics, Diagnostic{Code: "job_cleanup_failed", Severity: "warning", Message: "Owned discovery job cleanup failed."})
		}
	}()
	form := url.Values{"id": {sid}, "search": {query.search}, "exec_mode": {"normal"}, "search_mode": {"normal"}, "max_time": {strconv.FormatInt(maxTime, 10)}, "auto_cancel": {strconv.FormatInt(seconds, 10)}, "enable_lookups": {"false"}, "allow_partial_results": {"true"}, "max_count": {strconv.Itoa(c.options.MaxRows + 1)}, "earliest_time": {"0"}}
	if c.options.Window.Earliest != "" {
		form.Set("earliest_time", epochBound(c.options.Window.Earliest))
	}
	if c.options.Window.Latest != "" {
		form.Set("latest_time", epochBound(c.options.Window.Latest))
	}
	var submitted struct {
		SID string `json:"sid"`
	}
	if err := c.postFormJSON(ctx, "/services/search/jobs", form, &submitted); err != nil {
		result.gap("job_submission_failed")
		return
	}
	if submitted.SID != sid {
		result.gap("job_ownership_mismatch")
		return
	}
	for {
		state, err := c.jobStatus(ctx, sid, &result)
		if err != nil {
			if ctx.Err() != nil || isRequestTimeout(err) {
				c.salvage(cleanup(), sid, &result)
			} else {
				result.gap("job_status_invalid")
			}
			return
		}
		if state.failed {
			result.gap("job_failed")
		}
		if state.finalized || state.zombie {
			result.gap("job_finalized")
		}
		if state.done || state.failed || state.finalized || state.zombie {
			if err = c.jobRows(ctx, sid, state.count, &result, false); err != nil {
				if ctx.Err() != nil || isRequestTimeout(err) {
					c.salvage(cleanup(), sid, &result)
				} else {
					result.gap("job_results_invalid")
				}
			}
			if result.Reason != "" && len(result.Rows) > 0 {
				result.Coverage = "partial"
			}
			result.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
			return
		}
		timer := time.NewTimer(pollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			c.salvage(cleanup(), sid, &result)
			return
		case <-timer.C:
		}
	}
}

type jobState struct {
	done, failed, finalized, zombie bool
	count                           int
	countKnown                      bool
}
type serverMessage struct {
	Type string `json:"type"`
}
type statusResponse struct {
	Entry []struct {
		Content  map[string]json.RawMessage `json:"content"`
		Messages []serverMessage            `json:"messages"`
	} `json:"entry"`
	Messages []serverMessage `json:"messages"`
}

func numericVariant(s string) (string, error) {
	if s == "" || s[0] != '-' && (s[0] < '0' || s[0] > '9') || !json.Valid([]byte(s)) {
		return "", failure("job_numeric_invalid")
	}
	return jsoninput.CanonicalNumber(s), nil
}
func normalizedBool(raw json.RawMessage, required bool) (bool, error) {
	if len(raw) == 0 {
		if required {
			return false, failure("job_status_invalid")
		}
		return false, nil
	}
	s := string(raw)
	if strings.HasPrefix(s, "\"") {
		if json.Unmarshal(raw, &s) != nil {
			return false, failure("job_status_invalid")
		}
	}
	switch strings.ToLower(s) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	}
	n, err := numericVariant(s)
	if err == nil {
		switch n {
		case "0":
			return false, nil
		case "1e0":
			return true, nil
		}
	}
	return false, failure("job_status_invalid")
}
func normalizedCount(raw json.RawMessage) (int, error) {
	s := string(raw)
	if strings.HasPrefix(s, "\"") {
		if json.Unmarshal(raw, &s) != nil {
			return 0, failure("job_count_invalid")
		}
	}
	canon, err := numericVariant(s)
	if err != nil {
		return 0, err
	}
	if canon == "0" {
		return 0, nil
	}
	parts := strings.Split(canon, "e")
	n, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || n < 0 || uint64(n) > uint64(math.MaxInt) {
		return 0, failure("job_count_invalid")
	}
	exponent, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || exponent < 0 || exponent > 19 {
		return 0, failure("job_count_invalid")
	}
	for ; exponent > 0; exponent-- {
		if n > int64(math.MaxInt)/10 {
			return 0, failure("job_count_invalid")
		}
		n *= 10
	}
	return int(n), nil
}
func messagesGap(messages []serverMessage, r *JobResult) {
	for _, m := range messages {
		switch strings.ToUpper(m.Type) {
		case "WARN", "WARNING":
			r.gap("job_warning")
		case "ERROR", "FATAL":
			r.gap("job_error")
		}
	}
}
func (c *Client) jobStatus(ctx context.Context, sid string, result *JobResult) (jobState, error) {
	var response statusResponse
	if err := c.getJSON(ctx, "/services/search/jobs/"+url.PathEscape(sid), nil, &response); err != nil {
		return jobState{}, err
	}
	if len(response.Entry) != 1 {
		return jobState{}, failure("job_status_invalid")
	}
	messagesGap(response.Messages, result)
	messagesGap(response.Entry[0].Messages, result)
	fields := response.Entry[0].Content
	// Some versions put messages under content rather than on the entry.
	var nested []serverMessage
	if raw := fields["messages"]; len(raw) > 0 {
		if json.Unmarshal(raw, &nested) != nil {
			return jobState{}, failure("job_status_invalid")
		}
		messagesGap(nested, result)
	}
	state := jobState{}
	var err error
	if state.done, err = normalizedBool(fields["isDone"], true); err != nil {
		return state, err
	}
	if state.failed, err = normalizedBool(fields["isFailed"], false); err != nil {
		return state, err
	}
	if state.finalized, err = normalizedBool(fields["isFinalized"], false); err != nil {
		return state, err
	}
	if state.zombie, err = normalizedBool(fields["isZombie"], false); err != nil {
		return state, err
	}
	var dispatch string
	if raw := fields["dispatchState"]; len(raw) > 0 {
		if json.Unmarshal(raw, &dispatch) != nil {
			return state, failure("job_status_invalid")
		}
		switch strings.ToUpper(dispatch) {
		case "FAILED", "INTERNAL_ERROR", "BAD_INPUT_CANCEL":
			state.failed = true
		case "FINALIZED":
			state.finalized = true
		case "ZOMBIE":
			state.zombie = true
		}
	}
	if raw, present := fields["resultCount"]; present {
		if state.count, err = normalizedCount(raw); err != nil {
			return state, err
		}
		state.countKnown = true
		return state, nil
	}
	// Startup statuses may omit the count; only explicit active states with no
	// terminal flags establish that it is safe to continue polling.
	if !state.done && !state.failed && !state.finalized && !state.zombie {
		switch strings.ToUpper(dispatch) {
		case "QUEUED", "PARSING", "RUNNING", "FINALIZING", "PAUSE":
			return state, nil
		}
	}
	return state, failure("job_status_invalid")
}
func isRequestTimeout(err error) bool {
	var failure *requestFailure
	return errors.As(err, &failure) && failure.code == "request_timeout"
}
func (c *Client) jobRows(ctx context.Context, sid string, count int, result *JobResult, salvage bool) error {
	if count >= c.options.MaxRows+1 {
		result.gap("job_row_limit")
	}
	target := count
	if target > c.options.MaxRows {
		target = c.options.MaxRows
	}
	assembled := 0
	for _, row := range result.Rows {
		b, _ := json.Marshal(row)
		assembled += len(b)
	}
	emptyPage := target == 0 && len(result.Rows) == 0
	// Short final pages already request one extra row. A full final page needs
	// one bounded end probe before its reported count can establish completeness.
	probeEnd := !salvage && count <= c.options.MaxRows && target > 0 && target%resultPageSize == 0
	for offset := len(result.Rows); offset < target || emptyPage || probeEnd && offset == target && result.Coverage == "complete"; {
		emptyPage = false
		expected := target - offset
		if expected == 0 {
			probeEnd = false
		}
		if expected > resultPageSize {
			expected = resultPageSize
		}
		limit := expected
		if limit < resultPageSize {
			limit++
		}
		var page struct {
			Preview  json.RawMessage              `json:"preview"`
			Offset   json.RawMessage              `json:"init_offset"`
			Results  []map[string]json.RawMessage `json:"results"`
			Messages []serverMessage              `json:"messages"`
		}
		if err := c.getJSON(ctx, "/services/search/v2/jobs/"+url.PathEscape(sid)+"/results", url.Values{"offset": {strconv.Itoa(offset)}, "count": {strconv.Itoa(limit)}}, &page); err != nil {
			return err
		}
		preview, err := normalizedBool(page.Preview, true)
		if err != nil {
			return err
		}
		if preview {
			result.gap("job_preview")
		}
		actualOffset, err := normalizedCount(page.Offset)
		if err != nil || actualOffset != offset || page.Results == nil || len(page.Results) > limit {
			return failure("job_page_invalid")
		}
		messagesGap(page.Messages, result)
		rows := page.Results
		if len(rows) > expected {
			rows = rows[:expected]
		}
		for _, row := range rows {
			if row == nil {
				return failure("job_page_invalid")
			}
			data, _ := json.Marshal(row)
			assembled += len(data)
			if assembled > assembledLimit {
				return failure("job_assembled_limit")
			}
			result.Rows = append(result.Rows, row)
		}
		offset += len(rows)
		if len(page.Results) < expected || len(page.Results) > expected && count <= c.options.MaxRows {
			return failure("job_count_mismatch")
		}
		if salvage || target == 0 {
			return nil
		}
	}
	return nil
}
func (c *Client) salvage(cleanup context.Context, sid string, result *JobResult) {
	result.gap("job_timeout")
	// Reserve half the remaining shared cleanup deadline for the mandatory DELETE.
	deadline, _ := cleanup.Deadline()
	ctx, cancel := context.WithTimeout(cleanup, time.Until(deadline)/2)
	defer cancel()
	if c.postFormJSON(ctx, "/services/search/jobs/"+url.PathEscape(sid)+"/control", url.Values{"action": {"finalize"}}, nil) != nil {
		result.gap("job_finalize_failed")
	}
	state, err := c.jobStatus(ctx, sid, result)
	if err == nil {
		if state.countKnown && c.jobRows(ctx, sid, state.count, result, true) != nil {
			result.gap("job_results_invalid")
		}
	} else {
		result.gap("job_status_invalid")
	}
	if len(result.Rows) > 0 {
		result.Coverage = "partial"
	}
	result.ObservedAt = time.Now().UTC().Format(time.RFC3339Nano)
}
