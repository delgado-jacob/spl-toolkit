package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

// ErrExitWithoutShutdown tells a stdio host to exit with status 1.
var ErrExitWithoutShutdown = errors.New("LSP exit before shutdown")

type work struct {
	snapshot snapshot
	id       json.RawMessage
	position Position
	serial   uint64
}
type completion struct {
	work  work
	value evaluation
	err   error
}
type incoming struct {
	body []byte
	err  error
}
type server struct {
	writer                                               *frameWriter
	config                                               configuration
	defaults                                             Options
	initialized, ready, shutdown, paused, versionSupport bool
	revision, generation, requestSerial                  uint64
	documents                                            map[string]snapshot
	pending                                              map[string]snapshot
	requests                                             map[string]work
	queue                                                []string
}

// Serve runs a framed stdio session. It creates one bounded background worker;
// superseded results are suppressed rather than claiming to interrupt ANTLR.
// The caller owns the streams and should close input on context cancellation
// when it needs to unblock an arbitrary io.Reader's pending Read.
func Serve(ctx context.Context, in io.Reader, out io.Writer, log io.Writer, options Options) error {
	return serve(ctx, in, out, log, options, canonicalAnalysis)
}

func serve(ctx context.Context, in io.Reader, out io.Writer, log io.Writer, options Options, analyze analyzeFunc) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if log == nil {
		log = io.Discard
	}
	s := &server{writer: &frameWriter{out: out}, defaults: options, documents: map[string]snapshot{}, pending: map[string]snapshot{}, requests: map[string]work{}}
	incomingMessages := make(chan incoming)
	go func() {
		r := bufio.NewReader(in)
		for {
			body, err := readFrame(r)
			select {
			case incomingMessages <- incoming{body, err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()
	jobs := make(chan work)
	completed := make(chan completion)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-jobs:
				value, err := analyze(job.snapshot)
				select {
				case completed <- completion{job, value, err}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	busy := false
	for {
		var send chan work
		var next work
		if !busy && !s.shutdown {
			if job, ok := s.next(); ok {
				send = jobs
				next = job
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case send <- next:
			busy = true
			if len(next.id) > 0 {
				s.queue = s.queue[1:]
			} else {
				delete(s.pending, next.snapshot.URI)
			}
		case done := <-completed:
			busy = false
			if err := s.complete(done); err != nil {
				return err
			}
		case input := <-incomingMessages:
			if input.err != nil {
				if input.err == io.EOF && s.shutdown {
					return nil
				}
				fmt.Fprintf(log, "LSP transport: %v\n", input.err)
				if input.err == io.EOF {
					return io.ErrUnexpectedEOF
				}
				return input.err
			}
			m, rpcErr := decodeMessage(input.body)
			if rpcErr != nil {
				if err := s.writer.reply(nil, nil, rpcErr); err != nil {
					return err
				}
				continue
			}
			exit, err := s.handle(m)
			if err != nil {
				fmt.Fprintf(log, "LSP session: %v\n", err)
				return err
			}
			if exit {
				return nil
			}
		}
	}
}

func (s *server) next() (work, bool) {
	for len(s.queue) > 0 {
		if job, ok := s.requests[s.queue[0]]; ok {
			return job, true
		}
		s.queue = s.queue[1:]
	}
	if len(s.pending) == 0 {
		return work{}, false
	}
	uris := make([]string, 0, len(s.pending))
	for uri := range s.pending {
		uris = append(uris, uri)
	}
	sort.Strings(uris)
	return work{snapshot: s.pending[uris[0]]}, true
}

func (s *server) complete(done completion) error {
	job := done.work
	if len(job.id) > 0 {
		key := idKey(job.id)
		if current, ok := s.requests[key]; !ok || current.serial != job.serial {
			return nil
		}
		delete(s.requests, key)
		if done.err != nil {
			return s.writer.reply(job.id, nil, &responseError{-32603, done.err.Error()})
		}
		highlights, err := documentHighlights(done.value.analysis, job.position)
		if err != nil {
			return s.writer.reply(job.id, nil, &responseError{-32602, err.Error()})
		}
		return s.writer.reply(job.id, highlights, nil)
	}
	latest, ok := s.documents[job.snapshot.URI]
	if !ok || s.paused || s.shutdown || latest.Generation != job.snapshot.Generation || latest.Version != job.snapshot.Version || s.revision != job.snapshot.Revision {
		return nil
	}
	if done.err != nil {
		if err := s.showError("Analysis failed: " + done.err.Error()); err != nil {
			return err
		}
		return s.publish(latest, []Diagnostic{})
	}
	return s.publish(latest, done.value.diagnostics)
}

func (s *server) handle(m message) (bool, error) {
	request := len(m.ID) > 0
	fail := func(code int, text string) (bool, error) {
		if !request {
			return false, nil
		}
		return false, s.writer.reply(m.ID, nil, &responseError{code, text})
	}
	if m.Method == "exit" {
		if request {
			return fail(-32600, "exit must be a notification")
		}
		if !s.shutdown {
			return false, ErrExitWithoutShutdown
		}
		return true, nil
	}
	if !s.initialized {
		if m.Method != "initialize" {
			return fail(-32002, "Server not initialized")
		}
		if !request {
			return false, nil
		}
		return false, s.initialize(m)
	}
	if m.Method == "initialize" {
		return fail(-32600, "Already initialized")
	}
	if s.shutdown {
		return fail(-32600, "Server has shut down")
	}
	if m.Method == "shutdown" {
		if !request {
			return false, nil
		}
		s.shutdown = true
		s.pending = map[string]snapshot{}
		if err := s.cancelAll(-32800, "Server shutting down"); err != nil {
			return false, err
		}
		return false, s.writer.reply(m.ID, nil, nil)
	}
	switch m.Method {
	case "initialized":
		if request {
			return fail(-32600, "initialized must be a notification")
		}
		s.ready = true
		return false, nil
	case "$/cancelRequest":
		if request {
			return fail(-32600, "cancelRequest must be a notification")
		}
		f, err := object(m.Params)
		if err != nil {
			return false, nil
		}
		id := f["id"]
		if !validID(id) {
			return false, nil
		}
		if job, ok := s.requests[idKey(id)]; ok {
			s.removeRequest(idKey(id))
			return false, s.writer.reply(job.id, nil, &responseError{-32800, "Request cancelled"})
		}
		return false, nil
	case "textDocument/didOpen", "textDocument/didChange", "textDocument/didClose", "workspace/didChangeConfiguration":
		if request {
			return fail(-32600, "Method must be a notification")
		}
		if !s.ready {
			return false, nil
		}
		var err error
		switch m.Method {
		case "textDocument/didOpen":
			err = s.open(m.Params)
		case "textDocument/didChange":
			err = s.change(m.Params)
		case "textDocument/didClose":
			err = s.close(m.Params)
		default:
			return false, s.configure(m.Params)
		}
		if err != nil {
			return false, s.showError(m.Method + ": " + err.Error())
		}
		return false, nil
	case "textDocument/documentHighlight":
		if !request {
			return false, nil
		}
		if !s.ready {
			return fail(-32002, "Client initialization is incomplete")
		}
		return false, s.highlight(m)
	default:
		return fail(-32601, "Method not found")
	}
}

func (s *server) initialize(m message) error {
	p, err := object(m.Params)
	if err != nil {
		return s.writer.reply(m.ID, nil, &responseError{-32602, "initialize params: " + err.Error()})
	}
	raw, err := json.Marshal(s.defaults)
	if err != nil {
		return s.writer.reply(m.ID, nil, &responseError{-32602, err.Error()})
	}
	if input, ok := p["initializationOptions"]; ok && string(input) != "null" {
		raw = input
	}
	config, err := prepareConfiguration(raw)
	if err != nil {
		return s.writer.reply(m.ID, nil, &responseError{-32602, "configuration: " + err.Error()})
	}
	var caps struct {
		TextDocument struct {
			PublishDiagnostics struct {
				VersionSupport bool `json:"versionSupport"`
			} `json:"publishDiagnostics"`
		} `json:"textDocument"`
	}
	if _, err := object(p["capabilities"]); err != nil {
		return s.writer.reply(m.ID, nil, &responseError{-32602, "initialize capabilities object is required"})
	}
	if raw, ok := p["capabilities"]; !ok || json.Unmarshal(raw, &caps) != nil {
		return s.writer.reply(m.ID, nil, &responseError{-32602, "initialize capabilities object is required"})
	}
	s.config = config
	s.revision++
	s.versionSupport = caps.TextDocument.PublishDiagnostics.VersionSupport
	s.initialized = true
	return s.writer.reply(m.ID, map[string]any{"capabilities": map[string]any{"positionEncoding": "utf-16", "textDocumentSync": map[string]any{"openClose": true, "change": 1}, "documentHighlightProvider": true}}, nil)
}

func (s *server) showError(message string) error {
	return s.writer.notify("window/showMessage", map[string]any{"type": 1, "message": message})
}

func (s *server) configure(raw []byte) error {
	p, err := object(raw)
	var config configuration
	if err == nil {
		var settings map[string]json.RawMessage
		settings, err = object(p["settings"])
		if err == nil {
			config, err = prepareConfiguration(settings["splToolkit"])
		}
	}
	s.revision++
	s.pending = map[string]snapshot{}
	if cancelErr := s.cancelAll(-32801, "Configuration changed"); cancelErr != nil {
		return cancelErr
	}
	s.paused = err != nil
	if err != nil {
		if writeErr := s.showError("Configuration error; diagnostics paused: " + err.Error()); writeErr != nil {
			return writeErr
		}
		for _, uri := range s.uris() {
			if writeErr := s.publish(s.documents[uri], []Diagnostic{}); writeErr != nil {
				return writeErr
			}
		}
		return nil
	}
	s.config = config
	for _, uri := range s.uris() {
		s.schedule(uri)
	}
	return nil
}

func (s *server) uris() []string {
	uris := make([]string, 0, len(s.documents))
	for uri := range s.documents {
		uris = append(uris, uri)
	}
	sort.Strings(uris)
	return uris
}
func (s *server) cancelAll(code int, message string) error {
	ids := make([]string, 0, len(s.requests))
	for id := range s.requests {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := s.writer.reply(s.requests[id].id, nil, &responseError{code, message}); err != nil {
			return err
		}
	}
	s.requests = map[string]work{}
	s.queue = nil
	return nil
}

func (s *server) highlight(m message) error {
	fail := func(code int, text string) error { return s.writer.reply(m.ID, nil, &responseError{code, text}) }
	if s.paused {
		return fail(-32602, "Configuration error; diagnostics and highlights paused")
	}
	p, err := object(m.Params)
	if err != nil {
		return fail(-32602, err.Error())
	}
	d, err := object(p["textDocument"])
	if err != nil {
		return fail(-32602, err.Error())
	}
	var uri string
	if err = requiredString(d["uri"], &uri); err != nil {
		return fail(-32602, err.Error())
	}
	doc, ok := s.documents[uri]
	if !ok {
		return fail(-32602, "Document is not open")
	}
	position, err := decodePosition(p["position"])
	if err != nil {
		return fail(-32602, err.Error())
	}
	if _, err = offsetAt(doc.Document.Text, position); err != nil {
		return fail(-32602, err.Error())
	}
	if _, ok := s.requests[idKey(m.ID)]; ok {
		return fail(-32600, "Request ID is already pending")
	}
	if len(s.requests) >= 64 {
		return fail(-32000, "Too many pending requests")
	}
	s.requestSerial++
	job := work{snapshot: doc, id: m.ID, position: position, serial: s.requestSerial}
	s.requests[idKey(m.ID)] = job
	s.queue = append(s.queue, idKey(m.ID))
	return nil
}

func (s *server) removeRequest(key string) {
	delete(s.requests, key)
	for i, id := range s.queue {
		if id == key {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			break
		}
	}
}
