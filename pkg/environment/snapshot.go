package environment

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

var objectKinds = []string{"index", "source", "sourcetype", "dataset", "data_model", "lookup", "macro", "saved_search", "event_type", "tag", "calculated_field", "field_extraction", "module", "function", "external_command"}
var kindSet = func() map[string]bool {
	m := map[string]bool{}
	for _, kind := range objectKinds {
		m[kind] = true
	}
	return m
}()
var queryKinds = map[string]bool{"dataset": true, "data_model": true, "lookup": true, "macro": true, "saved_search": true, "event_type": true, "tag": true, "calculated_field": true, "field_extraction": true, "module": true, "function": true, "external_command": true}

func invalidSnapshot(path string, err error) (*PreparedSnapshot, *Report, error) {
	return nil, &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "snapshot_invalid", Severity: "error", Artifact: "snapshot", Path: path, Message: err.Error()}}}, nil
}
func nonblank(value, name string) error {
	if strings.TrimSpace(value) == "" || !utf8.ValidString(value) {
		return fmt.Errorf("%s must be nonblank valid UTF-8", name)
	}
	return nil
}
func optionalText(value, name string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", name)
	}
	return nil
}
func normalizeSelector(sel Selector, name string) (Selector, error) {
	if sel.All != nil {
		if !*sel.All || sel.Values != nil {
			return Selector{}, fmt.Errorf("%s selector must use exactly one branch", name)
		}
		yes := true
		return Selector{All: &yes}, nil
	}
	if len(sel.Values) == 0 {
		return Selector{}, fmt.Errorf("%s selector requires all or nonempty values", name)
	}
	seen := map[string]bool{}
	out := Selector{Values: append([]string{}, sel.Values...)}
	for _, v := range out.Values {
		if err := nonblank(v, name+" selector value"); err != nil {
			return Selector{}, err
		}
		if seen[v] {
			return Selector{}, fmt.Errorf("duplicate %s selector value %q", name, v)
		}
		seen[v] = true
	}
	sort.Strings(out.Values)
	return out, nil
}
func normalizeTime(value, name string) (string, time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%s must be an RFC3339 UTC timestamp: %w", name, err)
	}
	_, offset := parsed.Zone()
	if offset != 0 {
		return "", time.Time{}, fmt.Errorf("%s must be UTC", name)
	}
	return parsed.UTC().Format(time.RFC3339Nano), parsed.UTC(), nil
}
func normalizeProvenance(p Provenance, start, end time.Time) (Provenance, error) {
	for _, field := range []struct{ name, value string }{{"source_kind", p.SourceKind}, {"source_id", p.SourceID}} {
		if err := nonblank(field.value, field.name); err != nil {
			return Provenance{}, err
		}
	}
	observed, at, err := normalizeTime(p.ObservedAt, "observed_at")
	if err != nil {
		return Provenance{}, err
	}
	if at.Before(start) || at.After(end) {
		return Provenance{}, fmt.Errorf("observation is outside capture interval")
	}
	p.ObservedAt = observed
	return p, nil
}
func inScope(sel Selector, value string) bool {
	if sel.All != nil {
		return true
	}
	for _, candidate := range sel.Values {
		if value == candidate {
			return true
		}
	}
	return false
}
func cloneSnapshot(value Snapshot) Snapshot {
	raw, _ := json.Marshal(value)
	var out Snapshot
	_ = json.Unmarshal(raw, &out)
	return out
}
func validateUTF8(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return utf8.ValidString(value.String())
	case reflect.Pointer, reflect.Interface:
		return value.IsNil() || validateUTF8(value.Elem())
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).IsExported() && !validateUTF8(value.Field(i)) {
				return false
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < value.Len(); i++ {
			if !validateUTF8(value.Index(i)) {
				return false
			}
		}
	}
	return true
}
func normalizeSnapshot(input Snapshot) (Snapshot, []CoverageEntry, []Diagnostic, error) {
	if !validateUTF8(reflect.ValueOf(input)) {
		return Snapshot{}, nil, nil, fmt.Errorf("snapshot contains invalid UTF-8")
	}

	if input.SchemaVersion != 1 {
		return Snapshot{}, nil, nil, fmt.Errorf("schema_version must be integer 1")
	}
	if err := nonblank(input.ScopeID, "scope_id"); err != nil {
		return Snapshot{}, nil, nil, err
	}
	out := cloneSnapshot(input)
	var err error
	out.CaptureScope.Namespace, err = normalizeSelector(input.CaptureScope.Namespace, "namespace")
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	out.CaptureScope.App, err = normalizeSelector(input.CaptureScope.App, "app")
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	out.CaptureScope.Owner, err = normalizeSelector(input.CaptureScope.Owner, "owner")
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	for _, field := range []struct{ name, value string }{{"instance_id", out.Origin.InstanceID}, {"product_version", out.Origin.ProductVersion}, {"producer", out.Origin.Producer}, {"producer_version", out.Origin.ProducerVersion}} {
		if err := nonblank(field.value, field.name); err != nil {
			return Snapshot{}, nil, nil, err
		}
	}
	start, from, err := normalizeTime(out.Capture.Start, "capture.start")
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	end, to, err := normalizeTime(out.Capture.End, "capture.end")
	if err != nil {
		return Snapshot{}, nil, nil, err
	}
	if from.After(to) {
		return Snapshot{}, nil, nil, fmt.Errorf("capture start is after end")
	}
	out.Capture = CaptureInterval{Start: start, End: end}
	out.Capabilities = make([]Capability, 0, len(input.Capabilities))
	caps := map[string]bool{}
	for _, cap := range input.Capabilities {
		if err := nonblank(cap.ID, "capability id"); err != nil {
			return Snapshot{}, nil, nil, err
		}
		if err := nonblank(cap.Version, "capability version"); err != nil {
			return Snapshot{}, nil, nil, err
		}
		if caps[cap.ID] {
			return Snapshot{}, nil, nil, fmt.Errorf("duplicate capability fact %q", cap.ID)
		}
		caps[cap.ID] = true
		if cap.State != "available" && cap.State != "unavailable" && cap.State != "unknown" {
			return Snapshot{}, nil, nil, fmt.Errorf("unsupported capability state %q", cap.State)
		}
		cap.Provenance, err = normalizeProvenance(cap.Provenance, from, to)
		if err != nil {
			return Snapshot{}, nil, nil, err
		}
		out.Capabilities = append(out.Capabilities, cap)
	}
	sort.Slice(out.Capabilities, func(i, j int) bool { return out.Capabilities[i].ID < out.Capabilities[j].ID })
	out.Collections = make([]Collection, 0, len(input.Collections))
	collections := map[string]Collection{}
	for _, c := range input.Collections {
		if !kindSet[c.Kind] {
			return Snapshot{}, nil, nil, fmt.Errorf("unsupported collection kind %q", c.Kind)
		}
		if _, found := collections[c.Kind]; found {
			return Snapshot{}, nil, nil, fmt.Errorf("duplicate collection kind %q", c.Kind)
		}
		if c.Coverage != "complete" && c.Coverage != "partial" && c.Coverage != "unavailable" {
			return Snapshot{}, nil, nil, fmt.Errorf("invalid collection coverage %q", c.Coverage)
		}
		if c.Coverage != "complete" {
			if err := nonblank(c.Reason, "collection reason"); err != nil {
				return Snapshot{}, nil, nil, err
			}
		} else if c.Reason != "" {
			return Snapshot{}, nil, nil, fmt.Errorf("complete collection cannot have a reason")
		}
		if (c.Kind == "index" || c.Kind == "source" || c.Kind == "sourcetype") && c.Coverage == "complete" && (out.CaptureScope.Namespace.All == nil || out.CaptureScope.App.All == nil || out.CaptureScope.Owner.All == nil) {
			return Snapshot{}, nil, nil, fmt.Errorf("complete %s collection cannot use restricted scope", c.Kind)
		}
		collections[c.Kind] = c
		out.Collections = append(out.Collections, c)
	}
	sort.Slice(out.Collections, func(i, j int) bool { return out.Collections[i].Kind < out.Collections[j].Kind })
	out.Objects = make([]Object, 0, len(input.Objects))
	ids := map[string]bool{}
	for _, o := range input.Objects {
		for _, field := range []struct{ name, value string }{{"object id", o.ID}, {"object name", o.Name}} {
			if err := nonblank(field.value, field.name); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		if ids[o.ID] {
			return Snapshot{}, nil, nil, fmt.Errorf("duplicate object id %q", o.ID)
		}
		ids[o.ID] = true
		if !kindSet[o.Kind] {
			return Snapshot{}, nil, nil, fmt.Errorf("unsupported object kind %q", o.Kind)
		}
		c, found := collections[o.Kind]
		if !found {
			return Snapshot{}, nil, nil, fmt.Errorf("object %q has omitted collection", o.ID)
		}
		if c.Coverage == "unavailable" {
			return Snapshot{}, nil, nil, fmt.Errorf("unavailable collection %q contains object", o.Kind)
		}
		if o.Kind == "index" || o.Kind == "source" || o.Kind == "sourcetype" {
			if o.Namespace != "" || o.App != "" || o.Owner != "" {
				return Snapshot{}, nil, nil, fmt.Errorf("%s object has inapplicable context", o.Kind)
			}
		} else if !inScope(out.CaptureScope.Namespace, o.Namespace) || !inScope(out.CaptureScope.App, o.App) || !inScope(out.CaptureScope.Owner, o.Owner) {
			return Snapshot{}, nil, nil, fmt.Errorf("object %q is outside capture scope", o.ID)
		}
		for _, field := range []struct{ name, value string }{{"namespace", o.Namespace}, {"app", o.App}, {"owner", o.Owner}, {"sharing", o.Sharing}} {
			if err := optionalText(field.value, field.name); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		if !queryKinds[o.Kind] && (o.Document != nil || o.Arity != nil || len(o.Arguments) > 0 || o.EvalBased != nil || o.Validation != nil || len(o.Relations) > 0) {
			return Snapshot{}, nil, nil, fmt.Errorf("query metadata on %s object", o.Kind)
		}
		if o.Kind != "macro" && (o.Arity != nil || len(o.Arguments) > 0 || o.EvalBased != nil || o.Validation != nil) {
			return Snapshot{}, nil, nil, fmt.Errorf("macro metadata on non-macro object")
		}
		if o.Kind == "macro" && o.Arity != nil {
			if *o.Arity < 0 || len(o.Arguments) != *o.Arity {
				return Snapshot{}, nil, nil, fmt.Errorf("macro arity and arguments are inconsistent")
			}
		}
		for _, arg := range o.Arguments {
			if err := nonblank(arg, "macro argument"); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		if o.Document != nil {
			if err := optionalText(o.Document.Text, "document text"); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		if o.Validation != nil {
			if err := optionalText(*o.Validation, "macro validation"); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		for _, r := range o.Relations {
			if !queryKinds[r.Kind] {
				return Snapshot{}, nil, nil, fmt.Errorf("unsupported relation kind %q", r.Kind)
			}
			if err := nonblank(r.Name, "relation name"); err != nil {
				return Snapshot{}, nil, nil, err
			}
		}
		o.Provenance, err = normalizeProvenance(o.Provenance, from, to)
		if err != nil {
			return Snapshot{}, nil, nil, err
		}
		if o.Arguments != nil {
			o.Arguments = append([]string{}, o.Arguments...)
		}
		if o.Relations != nil {
			o.Relations = append([]closure.Relation{}, o.Relations...)
		}
		out.Objects = append(out.Objects, o)
	}
	sort.Slice(out.Objects, func(i, j int) bool { return out.Objects[i].ID < out.Objects[j].ID })
	coverage := make([]CoverageEntry, 0, len(objectKinds))
	diagnostics := []Diagnostic{}
	for _, kind := range objectKinds {
		if c, found := collections[kind]; found {
			coverage = append(coverage, CoverageEntry{Artifact: "snapshot", Kind: kind, Coverage: c.Coverage, Reason: c.Reason})
			if c.Coverage != "complete" {
				diagnostics = append(diagnostics, Diagnostic{Code: "collection_" + c.Coverage, Severity: "warning", Artifact: "snapshot", Path: "/collections", Message: fmt.Sprintf("%s collection is %s: %s", kind, c.Coverage, c.Reason)})
			}
		} else {
			coverage = append(coverage, CoverageEntry{Artifact: "snapshot", Kind: kind, Coverage: "unavailable", Reason: "omitted"})
			diagnostics = append(diagnostics, Diagnostic{Code: "collection_omitted", Severity: "warning", Artifact: "snapshot", Path: "/collections", Message: fmt.Sprintf("%s collection is omitted", kind)})
		}
	}
	sort.Slice(coverage, func(i, j int) bool { return coverage[i].Kind < coverage[j].Kind })
	sort.Slice(diagnostics, func(i, j int) bool { return diagnostics[i].Message < diagnostics[j].Message })
	return out, coverage, diagnostics, nil
}

// PrepareSnapshot validates and detaches a typed snapshot.
func PrepareSnapshot(value Snapshot) (*PreparedSnapshot, *Report, error) {
	normalized, coverage, diagnostics, err := normalizeSnapshot(value)
	if err != nil {
		return invalidSnapshot("", err)
	}
	digest, err := snapshotDigest(normalized)
	if err != nil {
		return nil, nil, err
	}
	if value.Digest != "" && value.Digest != digest {
		return invalidSnapshot("/digest", fmt.Errorf("asserted snapshot digest does not match computed digest"))
	}
	status := "valid"
	if len(diagnostics) > 0 {
		status = "partial"
	}
	report := Report{SchemaVersion: 1, Status: status, SnapshotDigest: digest, Coverage: coverage, Diagnostics: diagnostics}
	return &PreparedSnapshot{snapshot: normalized, report: report}, &report, nil
}

// ValidateArtifacts validates supplied artifact bytes. Nil means absent.
func ValidateArtifacts(snapshotJSON, schemaBundleJSON []byte) (*Report, error) {
	if snapshotJSON == nil && schemaBundleJSON == nil {
		return &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "artifact_missing", Severity: "error", Artifact: "request", Message: "at least one artifact is required"}}}, nil
	}
	if snapshotJSON == nil {
		return &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "schema_bundle_unsupported", Severity: "error", Artifact: "schema_bundle", Message: "schema bundle preparation is unavailable"}}}, nil
	}
	var snapshot Snapshot
	if err := decodeStrictJSON(snapshotJSON, &snapshot); err != nil {
		_, report, _ := invalidSnapshot("", err)
		return report, nil
	}
	if err := validateSnapshotRawShape(snapshotJSON); err != nil {
		_, report, _ := invalidSnapshot("", err)
		return report, nil
	}
	_, report, err := PrepareSnapshot(snapshot)
	if err != nil || report.Status == "invalid" || schemaBundleJSON == nil {
		return report, err
	}
	return &Report{SchemaVersion: 1, Status: "invalid", SnapshotDigest: report.SnapshotDigest, Coverage: report.Coverage, Diagnostics: []Diagnostic{{Code: "schema_bundle_unsupported", Severity: "error", Artifact: "schema_bundle", Message: "schema bundle preparation is unavailable"}}}, nil
}

// ValidateJSON validates an inline environment validation request.
func ValidateJSON(raw []byte) (*Report, error) {
	var request struct {
		SchemaVersion int             `json:"schema_version"`
		Snapshot      json.RawMessage `json:"snapshot"`
		SchemaBundle  json.RawMessage `json:"schema_bundle"`
	}
	if err := decodeStrictJSON(raw, &request); err != nil {
		return &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "request_invalid", Severity: "error", Artifact: "request", Message: err.Error()}}}, nil
	}
	if request.SchemaVersion != 1 {
		return &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "request_invalid", Severity: "error", Artifact: "request", Path: "/schema_version", Message: "schema_version must be integer 1"}}}, nil
	}
	return ValidateArtifacts(request.Snapshot, request.SchemaBundle)
}
