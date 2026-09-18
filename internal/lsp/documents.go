package lsp

import (
	"encoding/json"
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

// Options selects canonical local configuration. Targets contain inline data;
// the server never retrieves schema URLs or reads document URIs from disk.
type Options struct {
	Profile          string                   `json:"profile,omitempty"`
	Version          string                   `json:"version,omitempty"`
	ValidationTarget *corpus.ValidationTarget `json:"validation_target,omitempty"`
}
type configuration struct {
	options  Options
	prepared *corpus.PreparedScan
}
type snapshot struct {
	URI        string
	Generation uint64
	Version    int32
	Revision   uint64
	Document   analysis.QueryDocument
	Config     configuration
}

func prepareConfiguration(raw []byte) (configuration, error) {
	f, err := object(raw)
	if err != nil {
		return configuration{}, err
	}
	o := Options{}
	for key, value := range f {
		switch key {
		case "profile":
			err = requiredString(value, &o.Profile)
		case "version":
			err = requiredString(value, &o.Version)
		case "validation_target":
			o.ValidationTarget, err = corpus.DecodeValidationTarget(value)
		default:
			err = fmt.Errorf("unknown configuration property %q", key)
		}
		if err != nil {
			return configuration{}, fmt.Errorf("%s: %w", key, err)
		}
	}
	selectors, err := capabilityselector.Normalize("", o.Profile, o.Version)
	if err != nil {
		return configuration{}, err
	}
	o.Profile = selectors.Profile
	o.Version = selectors.Version
	p, err := corpus.Prepare(corpus.ScanOptions{ValidationTarget: o.ValidationTarget})
	return configuration{o, p}, err
}
func requiredString(raw json.RawMessage, out *string) error {
	if len(raw) == 0 || raw[0] != '"' {
		return fmt.Errorf("expected string")
	}
	return json.Unmarshal(raw, out)
}
func requiredVersion(raw json.RawMessage) (int32, error) {
	var n int32
	if len(raw) == 0 || string(raw) == "null" {
		return n, fmt.Errorf("document version is required")
	}
	err := json.Unmarshal(raw, &n)
	return n, err
}

func (s *server) open(raw []byte) error {
	p, err := object(raw)
	if err != nil {
		return err
	}
	d, err := object(p["textDocument"])
	if err != nil {
		return err
	}
	var uri, language, text string
	for _, v := range []struct {
		key string
		out *string
	}{{"uri", &uri}, {"languageId", &language}, {"text", &text}} {
		if err = requiredString(d[v.key], v.out); err != nil {
			return fmt.Errorf("%s: %w", v.key, err)
		}
	}
	if uri == "" {
		return fmt.Errorf("document URI is required")
	}
	if language != "spl" && language != "spl2" {
		return fmt.Errorf("unsupported languageId %q; expected spl or spl2", language)
	}
	version, err := requiredVersion(d["version"])
	if err != nil {
		return err
	}
	if _, ok := s.documents[uri]; ok {
		return fmt.Errorf("document already open")
	}
	s.generation++
	doc := snapshot{URI: uri, Generation: s.generation, Version: version}
	doc.Document = analysis.QueryDocument{Text: text, Language: language, SourceID: uri}
	s.documents[uri] = doc
	s.schedule(uri)
	return nil
}
func (s *server) change(raw []byte) error {
	p, err := object(raw)
	if err != nil {
		return err
	}
	d, err := object(p["textDocument"])
	if err != nil {
		return err
	}
	var uri string
	if err = requiredString(d["uri"], &uri); err != nil {
		return err
	}
	doc, ok := s.documents[uri]
	if !ok {
		return fmt.Errorf("document is not open")
	}
	version, err := requiredVersion(d["version"])
	if err != nil {
		return err
	}
	if version <= doc.Version {
		return fmt.Errorf("document version must increase")
	}
	var changes []json.RawMessage
	if err = json.Unmarshal(p["contentChanges"], &changes); err != nil || len(changes) == 0 {
		return fmt.Errorf("nonempty full-text contentChanges required")
	}
	var text string
	for _, change := range changes {
		f, err := object(change)
		if err != nil {
			return err
		}
		if _, ok := f["range"]; ok {
			return fmt.Errorf("incremental changes are unsupported; send full text")
		}
		if _, ok := f["rangeLength"]; ok {
			return fmt.Errorf("rangeLength is unsupported with full synchronization")
		}
		if err = requiredString(f["text"], &text); err != nil {
			return err
		}
	}
	doc.Version = version
	doc.Document.Text = text
	s.documents[uri] = doc
	s.schedule(uri)
	return nil
}
func (s *server) close(raw []byte) error {
	p, err := object(raw)
	if err != nil {
		return err
	}
	d, err := object(p["textDocument"])
	if err != nil {
		return err
	}
	var uri string
	if err = requiredString(d["uri"], &uri); err != nil {
		return err
	}
	doc, ok := s.documents[uri]
	if !ok {
		return fmt.Errorf("document is not open")
	}
	delete(s.documents, uri)
	delete(s.pending, uri)
	return s.publish(doc, []Diagnostic{})
}
func (s *server) schedule(uri string) {
	if s.paused || s.shutdown {
		return
	}
	doc := s.documents[uri]
	doc.Revision = s.revision
	doc.Config = s.config
	doc.Document.Profile = s.config.options.Profile
	doc.Document.Version = s.config.options.Version
	s.documents[uri] = doc
	s.pending[uri] = doc
}

func decodePosition(raw []byte) (Position, error) {
	p := Position{}
	f, err := object(raw)
	if err != nil {
		return p, err
	}
	for _, v := range []struct {
		key string
		out *int
	}{{"line", &p.Line}, {"character", &p.Character}} {
		n, ok := f[v.key]
		if !ok || string(n) == "null" || json.Unmarshal(n, v.out) != nil || *v.out < 0 || *v.out > 2147483647 {
			return p, fmt.Errorf("invalid position %s", v.key)
		}
	}
	return p, nil
}
