package impact

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/delgado-jacob/spl-toolkit/internal/jsoninput"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/rewrite"
)

type InputError struct{ Err error }

func (e *InputError) Error() string { return e.Err.Error() }
func (e *InputError) Unwrap() error { return e.Err }
func IsInputError(err error) bool   { var e *InputError; return errors.As(err, &e) }
func inputError(format string, args ...any) error {
	return &InputError{Err: fmt.Errorf(format, args...)}
}

func classifyInputError(err error) error {
	if err != nil && (corpus.IsInputError(err) || rewrite.IsInputError(err)) {
		return &InputError{Err: err}
	}
	return err
}

type SchemaRequest struct {
	SchemaVersion int
	Input         corpus.Input
	BeforeTarget  corpus.ValidationTarget
	AfterTarget   corpus.ValidationTarget
}

type MappingSide struct {
	Rules            rewrite.RuleSet
	ValidationTarget *rewrite.ValidationTarget
}

type MappingRequest struct {
	SchemaVersion int
	Input         corpus.Input
	Before        MappingSide
	After         MappingSide
}

// DecodeSchemaRequest accepts only inline document selection. Local callers
// prepare targets first and supply a separately acquired corpus.Input.
func DecodeSchemaRequest(data []byte) (SchemaRequest, error) {
	f, err := requestObject(data, "schema_version", "documents", "before_target", "after_target")
	if err != nil {
		return SchemaRequest{}, err
	}
	if !versionOne(f["schema_version"]) {
		return SchemaRequest{}, inputError("schema_version must be integer 1")
	}
	input, err := inlineInput(f["documents"])
	if err != nil {
		return SchemaRequest{}, err
	}
	before, err := corpus.DecodeValidationTarget(f["before_target"])
	if err != nil {
		return SchemaRequest{}, inputError("before_target: %v", err)
	}
	after, err := corpus.DecodeValidationTarget(f["after_target"])
	if err != nil {
		return SchemaRequest{}, inputError("after_target: %v", err)
	}
	return SchemaRequest{SchemaVersion: 1, Input: input, BeforeTarget: *before, AfterTarget: *after}, nil
}

func DecodeMappingRequest(data []byte) (MappingRequest, error) {
	f, err := requestObject(data, "schema_version", "documents", "before_rules", "after_rules", "before_target?", "after_target?")
	if err != nil {
		return MappingRequest{}, err
	}
	if !versionOne(f["schema_version"]) {
		return MappingRequest{}, inputError("schema_version must be integer 1")
	}
	input, err := inlineInput(f["documents"])
	if err != nil {
		return MappingRequest{}, err
	}
	before, err := rewrite.DecodeRuleSet(f["before_rules"])
	if err != nil {
		return MappingRequest{}, inputError("before_rules: %v", err)
	}
	after, err := rewrite.DecodeRuleSet(f["after_rules"])
	if err != nil {
		return MappingRequest{}, inputError("after_rules: %v", err)
	}
	r := MappingRequest{SchemaVersion: 1, Input: input, Before: MappingSide{Rules: before}, After: MappingSide{Rules: after}}
	if raw, ok := f["before_target"]; ok {
		r.Before.ValidationTarget, err = corpus.DecodeValidationTarget(raw)
		if err != nil {
			return MappingRequest{}, inputError("before_target: %v", err)
		}
	}
	if raw, ok := f["after_target"]; ok {
		r.After.ValidationTarget, err = corpus.DecodeValidationTarget(raw)
		if err != nil {
			return MappingRequest{}, inputError("after_target: %v", err)
		}
	}
	return r, nil
}

func versionOne(raw []byte) bool { return string(bytes.TrimSpace(raw)) == "1" }

func inlineInput(raw []byte) (corpus.Input, error) {
	wrapper := append([]byte(`{"schema_version":1,"documents":`), raw...)
	wrapper = append(wrapper, '}')
	r, err := corpus.DecodeRequest(wrapper)
	if err != nil {
		return corpus.Input{}, inputError("documents: %v", err)
	}
	input := corpus.Input{Selection: corpus.Selection{Mode: "inline", Complete: true, IgnoredNames: []string{}, SkippedSymlinks: []string{}, TraversalFailures: []corpus.AcquisitionError{}}, Entries: make([]corpus.Entry, 0, len(r.Documents))}
	for _, item := range r.Documents {
		doc := item.Document
		input.Entries = append(input.Entries, corpus.Entry{ID: item.ID, Origin: corpus.Origin{Kind: "inline"}, Document: &doc})
	}
	return input, nil
}

func requestObject(data []byte, names ...string) (map[string]json.RawMessage, error) {
	if err := jsoninput.ValidateUnicode(data); err != nil {
		return nil, inputError("%v", err)
	}
	required := map[string]bool{}
	allowed := map[string]bool{}
	for _, name := range names {
		if len(name) > 0 && name[len(name)-1] == '?' {
			allowed[name[:len(name)-1]] = true
		} else {
			required[name] = true
			allowed[name] = true
		}
	}
	d := json.NewDecoder(bytes.NewReader(data))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return nil, inputError("expected an object")
	}
	f := map[string]json.RawMessage{}
	for d.More() {
		keyToken, err := d.Token()
		if err != nil {
			return nil, inputError("invalid object key: %v", err)
		}
		key := keyToken.(string)
		if !allowed[key] {
			return nil, inputError("unknown property %q", key)
		}
		if _, exists := f[key]; exists {
			return nil, inputError("duplicate property %q", key)
		}
		var value json.RawMessage
		if err := d.Decode(&value); err != nil {
			return nil, inputError("property %q: %v", key, err)
		}
		f[key] = value
	}
	if _, err := d.Token(); err != nil {
		return nil, inputError("invalid object: %v", err)
	}
	var trailing any
	if err := d.Decode(&trailing); err != io.EOF {
		return nil, inputError("expected exactly one JSON object")
	}
	for name := range required {
		if _, ok := f[name]; !ok {
			return nil, inputError("missing property %q", name)
		}
	}
	return f, nil
}
