package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

type client struct {
	t    *testing.T
	in   *io.PipeWriter
	out  *bufio.Reader
	done chan error
}

func newClient(t *testing.T, an analyzeFunc) *client {
	t.Helper()
	r, w := io.Pipe()
	or, ow := io.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	c := &client{t, w, bufio.NewReader(or), make(chan error, 1)}
	go func() { err := serve(ctx, r, ow, io.Discard, Options{}, an); _ = ow.Close(); c.done <- err }()
	t.Cleanup(func() { cancel(); _ = w.Close(); _ = or.Close() })
	return c
}
func (c *client) send(raw string) {
	c.t.Helper()
	if err := (&frameWriter{out: c.in}).write(json.RawMessage(raw)); err != nil {
		c.t.Fatal(err)
	}
}
func (c *client) recv() map[string]json.RawMessage {
	c.t.Helper()
	ch := make(chan []byte, 1)
	go func() { b, _ := readFrame(c.out); ch <- b }()
	select {
	case b := <-ch:
		var m map[string]json.RawMessage
		if err := json.Unmarshal(b, &m); err != nil {
			c.t.Fatalf("response %q: %v", b, err)
		}
		return m
	case <-time.After(10 * time.Second):
		c.t.Fatal("server response timeout")
		return nil
	}
}
func (c *client) init() {
	c.send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{"textDocument":{"publishDiagnostics":{"versionSupport":true}}}}}`)
	m := c.recv()
	if m["error"] != nil {
		c.t.Fatalf("initialize: %s", m["error"])
	}
	c.send(`{"jsonrpc":"2.0","method":"initialized","params":{}}`)
}
func (c *client) open(version int, text string) {
	c.t.Helper()
	params := map[string]any{"textDocument": map[string]any{"uri": "file:///query.spl", "languageId": "spl", "version": version, "text": text}}
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didOpen", "params": params})
	c.send(string(b))
}
func (c *client) change(version int, text string) {
	params := map[string]any{"textDocument": map[string]any{"uri": "file:///query.spl", "version": version}, "contentChanges": []any{map[string]any{"text": text}}}
	b, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "method": "textDocument/didChange", "params": params})
	c.send(string(b))
}
func errorCode(m map[string]json.RawMessage) int {
	var e responseError
	_ = json.Unmarshal(m["error"], &e)
	return e.Code
}
func expectID(t *testing.T, m map[string]json.RawMessage, id string) {
	t.Helper()
	if string(m["id"]) != id {
		t.Fatalf("wanted id %s: %+v", id, m)
	}
}
func TestInitializeAndShutdown(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.send(`{"jsonrpc":"2.0","id":0,"method":"x"}`)
	if m := c.recv(); errorCode(m) != -32002 {
		t.Fatalf("pre-init: %+v", m)
	}
	c.init()
	c.send(`{"jsonrpc":"2.0","method":"unknown"}`)
	c.send(`{"jsonrpc":"2.0","id":9007199254740993,"method":"unknown"}`)
	m := c.recv()
	expectID(t, m, "9007199254740993")
	if errorCode(m) != -32601 {
		t.Fatal(m)
	}
	c.send(`[]`)
	m = c.recv()
	expectID(t, m, "null")
	if errorCode(m) != -32600 {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","id":"shutdown","method":"shutdown"}`)
	m = c.recv()
	expectID(t, m, `"shutdown"`)
	if string(m["result"]) != "null" {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","id":9,"method":"textDocument/documentHighlight"}`)
	if m = c.recv(); errorCode(m) != -32600 {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","method":"exit"}`)
	if err := <-c.done; err != nil {
		t.Fatal(err)
	}
}

func TestFullSyncDocumentGenerations(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.init()
	c.open(7, "eval a=host | table a")
	m := c.recv()
	var p struct {
		Version     int          `json:"version"`
		Diagnostics []Diagnostic `json:"diagnostics"`
	}
	_ = json.Unmarshal(m["params"], &p)
	if p.Version != 7 {
		t.Fatal(m)
	}
	c.change(8, "eval a=")
	m = c.recv()
	_ = json.Unmarshal(m["params"], &p)
	if p.Version != 8 || len(p.Diagnostics) == 0 {
		t.Fatalf("no replacement: %s", m["params"])
	}
	c.change(8, "eval a=host")
	if m = c.recv(); string(m["method"]) != `"window/showMessage"` {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"file:///query.spl"}}}`)
	m = c.recv()
	_ = json.Unmarshal(m["params"], &p)
	if len(p.Diagnostics) != 0 {
		t.Fatal(m)
	}
	c.open(1, "eval a=host | table a")
	m = c.recv()
	_ = json.Unmarshal(m["params"], &p)
	if p.Version != 1 {
		t.Fatal(m)
	}
}

func TestLSPExcludesRequirementSpecificMessages(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.init()
	c.open(1, "table host*")
	message := c.recv()
	encoded := string(message["params"])
	if strings.Contains(encoded, `"requirements"`) || strings.Contains(encoded, "SPL_REQUIREMENT_") {
		t.Fatalf("requirement-specific payload entered LSP diagnostics: %s", encoded)
	}
}

// The analyzer barrier makes obsolete publication observable without sleeps.
func TestStaleDiagnosticsSuppressed(t *testing.T) {
	entered := make(chan string, 4)
	release := make(chan struct{}, 4)
	an := func(s snapshot) (evaluation, error) {
		entered <- s.Document.Text
		<-release
		return canonicalAnalysis(s)
	}
	c := newClient(t, an)
	c.init()
	c.open(1, "eval a=")
	<-entered
	c.change(2, "eval a=host | table a")
	c.send(`{"jsonrpc":"2.0","id":99,"method":"barrier"}`)
	expectID(t, c.recv(), "99")
	release <- struct{}{}
	if got := <-entered; got != "eval a=host | table a" {
		t.Fatal(got)
	}
	release <- struct{}{}
	m := c.recv()
	var p struct {
		Version int `json:"version"`
	}
	_ = json.Unmarshal(m["params"], &p)
	if p.Version != 2 {
		t.Fatalf("stale published %s", m["params"])
	}
	c.change(3, "eval b=")
	<-entered
	c.send(`{"jsonrpc":"2.0","method":"textDocument/didClose","params":{"textDocument":{"uri":"file:///query.spl"}}}`)
	c.recv()
	c.open(1, "eval c=host")
	release <- struct{}{}
	<-entered
	release <- struct{}{}
	m = c.recv()
	_ = json.Unmarshal(m["params"], &p)
	if p.Version != 1 {
		t.Fatal(m)
	}
}

func TestHighlightSnapshotCancellation(t *testing.T) {
	entered := make(chan string, 10)
	release := make(chan struct{}, 10)
	an := func(s snapshot) (evaluation, error) {
		entered <- s.Document.Text
		<-release
		return canonicalAnalysis(s)
	}
	c := newClient(t, an)
	c.init()
	c.open(1, "eval a=host | table a")
	<-entered
	c.send(`{"jsonrpc":"2.0","id":"h","method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	c.change(2, "eval z=9")
	c.send(`{"jsonrpc":"2.0","id":99,"method":"barrier"}`)
	expectID(t, c.recv(), "99")
	release <- struct{}{}
	if got := <-entered; got != "eval a=host | table a" {
		t.Fatalf("snapshot replaced: %s", got)
	}
	release <- struct{}{}
	m := c.recv()
	expectID(t, m, `"h"`)
	var highlights []Highlight
	_ = json.Unmarshal(m["result"], &highlights)
	if len(highlights) < 2 {
		t.Fatalf("missing related refs %s", m["result"])
	}
	<-entered
	c.send(`{"jsonrpc":"2.0","id":2,"method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	c.send(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":2}}`)
	m = c.recv()
	expectID(t, m, "2")
	if errorCode(m) != -32800 {
		t.Fatal(m)
	}
	release <- struct{}{}
	c.recv()
}

func TestConfigurationPauseAndRecovery(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.init()
	c.open(1, "eval a=host | table a")
	c.recv()
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{"profile":"bad"}}}}`)
	m := c.recv()
	if string(m["method"]) != `"window/showMessage"` {
		t.Fatal(m)
	}
	m = c.recv()
	if !strings.Contains(string(m["params"]), `"diagnostics":[]`) {
		t.Fatal(m)
	}
	c.change(2, "eval b=host | table b")
	c.send(`{"jsonrpc":"2.0","id":4,"method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	if m = c.recv(); errorCode(m) != -32602 {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{}}}}`)
	m = c.recv()
	if !strings.Contains(string(m["params"]), `"version":2`) {
		t.Fatal(m)
	}
}

// This helper runs the production Serve entry point as a real stdio subprocess
// for the Python protocol harness; it is excluded from ordinary test runs.
func TestStdioHelper(t *testing.T) {
	if os.Getenv("SPL_LSP_STDIO_HELPER") != "1" {
		t.Skip("subprocess helper")
	}
	err := Serve(context.Background(), os.Stdin, os.Stdout, os.Stderr, Options{})
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestCanonicalHighlights(t *testing.T) {
	for _, tc := range []struct {
		text     string
		at       int
		required []int
		absent   []int
	}{
		{"eval a=host | table a", 5, []int{5, 20}, nil},
		{"eval a=1 | table a | eval a=2 | table a", 5, []int{5, 17}, []int{26, 38}},
	} {
		r, err := analysis.Analyze(analysis.QueryDocument{Text: tc.text})
		if err != nil {
			t.Fatal(err)
		}
		got, err := documentHighlights(r, Position{0, tc.at})
		if err != nil {
			t.Fatal(err)
		}
		have := map[int]bool{}
		for _, h := range got {
			have[h.Range.Start.Character] = true
			if h.Range.Start.Character == 5 && h.Kind != 3 {
				t.Fatalf("assignment is not a write: %+v", h)
			}
		}
		for _, p := range tc.required {
			if !have[p] {
				t.Fatalf("%q expected %d, got %+v", tc.text, p, got)
			}
		}
		for _, p := range tc.absent {
			if have[p] {
				t.Fatalf("unrelated %d highlighted", p)
			}
		}
	}
}

func TestCancelRunningRequestAndReuseID(t *testing.T) {
	entered := make(chan string, 10)
	release := make(chan struct{}, 10)
	an := func(s snapshot) (evaluation, error) {
		entered <- s.Document.Text
		<-release
		return canonicalAnalysis(s)
	}
	c := newClient(t, an)
	c.init()
	c.open(1, "eval a=1 | table a")
	<-entered
	release <- struct{}{}
	c.recv()
	c.send(`{"jsonrpc":"2.0","id":"h","method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	<-entered
	c.send(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":"\u0068"}}`)
	m := c.recv()
	expectID(t, m, `"h"`)
	if errorCode(m) != -32800 {
		t.Fatal(m)
	}
	c.change(2, "eval z=2 | table z")
	c.send(`{"jsonrpc":"2.0","id":"h","method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	c.send(`{"jsonrpc":"2.0","id":99,"method":"barrier"}`)
	expectID(t, c.recv(), "99")
	release <- struct{}{}
	if got := <-entered; got != "eval z=2 | table z" {
		t.Fatalf("new request lost: %s", got)
	}
	release <- struct{}{}
	m = c.recv()
	expectID(t, m, `"h"`)
	if m["error"] != nil {
		t.Fatal(m)
	}
	<-entered
	release <- struct{}{}
	c.recv()
}

func TestConfigurationRevisionSuppressesOldWork(t *testing.T) {
	entered := make(chan string, 10)
	release := make(chan struct{}, 10)
	an := func(s snapshot) (evaluation, error) {
		entered <- s.Document.Text
		<-release
		return canonicalAnalysis(s)
	}
	c := newClient(t, an)
	c.init()
	c.open(1, "eval a=")
	<-entered
	c.send(`{"jsonrpc":"2.0","id":4,"method":"textDocument/documentHighlight","params":{"textDocument":{"uri":"file:///query.spl"},"position":{"line":0,"character":5}}}`)
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{"version":"bad"}}}}`)
	m := c.recv()
	expectID(t, m, "4")
	if errorCode(m) != -32801 {
		t.Fatal(m)
	}
	c.recv()
	c.recv()
	c.change(2, "eval a=host | table a")
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{}}}}`)
	c.send(`{"jsonrpc":"2.0","id":99,"method":"barrier"}`)
	expectID(t, c.recv(), "99")
	release <- struct{}{}
	<-entered
	release <- struct{}{}
	m = c.recv()
	if !strings.Contains(string(m["params"]), `"version":2`) {
		t.Fatalf("stale configuration result: %s", m["params"])
	}
}

func TestInvalidInitializeAndLanguage(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.send(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"capabilities":{},"initializationOptions":{"profile":"bad"}}}`)
	if m := c.recv(); errorCode(m) != -32602 {
		t.Fatal(m)
	}
	c.init()
	c.send(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///query.spl","languageId":"sql","version":1,"text":"SELECT x"}}}`)
	m := c.recv()
	if string(m["method"]) != `"window/showMessage"` {
		t.Fatal(m)
	}
	c.open(1, "eval a=host")
	c.recv()
	c.send(`{"jsonrpc":"2.0","method":"textDocument/didChange","params":{"textDocument":{"uri":"file:///query.spl","version":2},"contentChanges":[{"range":{"start":{"line":0,"character":0},"end":{"line":0,"character":1}},"text":"x"}]}}`)
	m = c.recv()
	if string(m["method"]) != `"window/showMessage"` {
		t.Fatal(m)
	}
	c.change(2, "eval b=host")
	m = c.recv()
	if !strings.Contains(string(m["params"]), `"version":2`) {
		t.Fatal(m)
	}
}

func TestValidationConfigurationUsesCanonicalTarget(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.init()
	c.open(1, "table host,missing")
	c.recv()
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{"validation_target":{"kind":"field_list","catalog":{"fields":["host"]}}}}}}`)
	m := c.recv()
	if string(m["method"]) != `"textDocument/publishDiagnostics"` || !strings.Contains(string(m["params"]), "missing") {
		t.Fatalf("canonical validation not published: %s %s", m["method"], m["params"])
	}
	c.send(`{"jsonrpc":"2.0","method":"workspace/didChangeConfiguration","params":{"settings":{"splToolkit":{}}}}`)
	m = c.recv()
	if !strings.Contains(string(m["params"]), `"diagnostics":[]`) {
		t.Fatal(m)
	}
}

func TestIndependentScopesAndAmbiguousHighlights(t *testing.T) {
	text := "search root=1 | append [ search child=1 | eval root=child ] | where root=2"
	result, err := analysis.Analyze(analysis.QueryDocument{Text: text})
	if err != nil {
		t.Fatal(err)
	}
	highlights, err := documentHighlights(result, Position{0, 7})
	if err != nil {
		t.Fatal(err)
	}
	if len(highlights) != 2 || highlights[0].Range.Start.Character != 7 || highlights[1].Range.Start.Character != 68 {
		t.Fatalf("independent root bindings merged: %+v", highlights)
	}
	duplicate := result.References[0]
	duplicate.ID = "overlap"
	result.References = append(result.References, duplicate)
	highlights, err = documentHighlights(result, Position{0, 7})
	if err != nil || len(highlights) != 0 {
		t.Fatalf("ambiguous group invented: %+v %v", highlights, err)
	}
}

func TestUnicodeQuotedIdentifierHighlights(t *testing.T) {
	result, err := analysis.Analyze(analysis.QueryDocument{Text: "eval '😀a'=host | table '😀a'"})
	if err != nil {
		t.Fatal(err)
	}
	highlights, err := documentHighlights(result, Position{0, 8})
	if err != nil {
		t.Fatal(err)
	}
	if len(highlights) != 3 || highlights[0].Range.Start.Character != 5 || highlights[0].Range.End.Character != 10 || highlights[2].Range.Start.Character != 24 {
		t.Fatalf("Unicode source offsets lost: %+v", highlights)
	}
}

func TestInitializeRejectsNullCapabilities(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.send(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"capabilities":null}}`)
	if m := c.recv(); errorCode(m) != -32602 {
		t.Fatalf("null capabilities accepted: %+v", m)
	}
}

func TestNoVersionSupportStillSuppressesStalePublication(t *testing.T) {
	entered := make(chan string, 4)
	release := make(chan struct{}, 4)
	an := func(s snapshot) (evaluation, error) {
		entered <- s.Document.Text
		<-release
		return canonicalAnalysis(s)
	}
	c := newClient(t, an)
	c.send(`{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"capabilities":{}}}`)
	c.recv()
	c.send(`{"jsonrpc":"2.0","method":"initialized"}`)
	c.open(1, "eval a=")
	<-entered
	c.change(2, "eval a=1")
	c.send(`{"jsonrpc":"2.0","id":99,"method":"barrier"}`)
	expectID(t, c.recv(), "99")
	release <- struct{}{}
	<-entered
	release <- struct{}{}
	m := c.recv()
	var p map[string]json.RawMessage
	_ = json.Unmarshal(m["params"], &p)
	if _, ok := p["version"]; ok || string(p["diagnostics"]) != "[]" {
		t.Fatalf("stale or versioned publication: %s", m["params"])
	}
}

func TestSPL2OpenUsesExplicitLanguage(t *testing.T) {
	c := newClient(t, canonicalAnalysis)
	c.init()
	c.send(`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"file:///query.spl2","languageId":"spl2","version":1,"text":"SELECT host FROM main WHERE bytes>0"}}}`)
	m := c.recv()
	if string(m["method"]) != `"textDocument/publishDiagnostics"` || !strings.Contains(string(m["params"]), `"diagnostics":[]`) {
		t.Fatalf("SPL2 open: %s %s", m["method"], m["params"])
	}
}

func TestPendingRequestsAreBoundedAndShutdownCompletesEach(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	an := func(s snapshot) (evaluation, error) { entered <- struct{}{}; <-release; return canonicalAnalysis(s) }
	c := newClient(t, an)
	c.init()
	c.open(1, "eval a=1")
	<-entered
	for i := 0; i < 65; i++ {
		raw, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": i + 10, "method": "textDocument/documentHighlight", "params": map[string]any{"textDocument": map[string]any{"uri": "file:///query.spl"}, "position": Position{0, 5}}})
		c.send(string(raw))
	}
	m := c.recv()
	expectID(t, m, "74")
	if errorCode(m) != -32000 {
		t.Fatal(m)
	}
	c.send(`{"jsonrpc":"2.0","id":"s","method":"shutdown"}`)
	seen := map[string]bool{}
	for i := 0; i < 64; i++ {
		m = c.recv()
		id := string(m["id"])
		if seen[id] || errorCode(m) != -32800 {
			t.Fatalf("request response lost or duplicated: %+v", m)
		}
		seen[id] = true
	}
	m = c.recv()
	expectID(t, m, `"s"`)
	if m["error"] != nil {
		t.Fatal(m)
	}
	close(release)
	c.send(`{"jsonrpc":"2.0","method":"exit"}`)
	if err := <-c.done; err != nil {
		t.Fatal(err)
	}
}
