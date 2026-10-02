package environment

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"time"
)

var observationWindowTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?(Z|[+-]00:00)$`)

type ObservationWindow struct {
	Mode     string `json:"mode"`
	Earliest string `json:"earliest,omitempty"`
	Latest   string `json:"latest,omitempty"`
}

// ObservationIndex describes a confirmed selected index; the index collection may
// also contain catalog entries outside the observation selection.
type ObservationIndex struct {
	IndexID           string   `json:"index_id"`
	CatalogDatatypes  []string `json:"catalog_datatypes"`
	RequiredDatatypes []string `json:"required_datatypes"`
}
type IndexEnumeration struct {
	Method     string     `json:"method"`
	PeerScope  string     `json:"peer_scope"`
	Coverage   string     `json:"coverage"`
	Reason     string     `json:"reason,omitempty"`
	Provenance Provenance `json:"provenance"`
}

// ObservationCapture records acquisition for one declared observed kind and
// index datatype, including unavailable attempts that produced no objects.
type ObservationCapture struct {
	Kind       string     `json:"kind"`
	IndexID    string     `json:"index_id"`
	Datatype   string     `json:"datatype"`
	ObjectIDs  []string   `json:"object_ids"`
	Coverage   string     `json:"coverage"`
	Reason     string     `json:"reason,omitempty"`
	Provenance Provenance `json:"provenance"`
}
type ObservationScope struct {
	IndexSelection   Selector             `json:"index_selection"`
	Enumeration      IndexEnumeration     `json:"enumeration"`
	Indexes          []ObservationIndex   `json:"indexes"`
	UnmatchedIndexes []string             `json:"unmatched_indexes"`
	Method           string               `json:"method"`
	Visibility       string               `json:"visibility"`
	Window           ObservationWindow    `json:"window"`
	TimePrecision    string               `json:"time_precision"`
	AbsenceMeaning   string               `json:"absence_meaning"`
	Captures         []ObservationCapture `json:"captures"`
}

// observationSet retains explicit empty sets and normalizes repeated evidence.
func observationSet(values []string, path string) ([]string, error) {
	if values == nil {
		return nil, at(path, fmt.Errorf("required array must be present and nonnull"))
	}
	seen := map[string]bool{}
	out := []string{}
	for i, value := range values {
		if err := nonblank(value, "set value"); err != nil {
			return nil, at(fmt.Sprintf("%s/%d", path, i), err)
		}
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out, nil
}
func observationCoverage(coverage, reason, path string) error {
	if coverage != "complete" && coverage != "partial" && coverage != "unavailable" {
		return at(path+"/coverage", fmt.Errorf("invalid observation coverage %q", coverage))
	}
	if coverage == "complete" {
		if reason != "" {
			return at(path+"/reason", fmt.Errorf("complete acquisition cannot have a reason"))
		}
	} else if err := nonblank(reason, "acquisition reason"); err != nil {
		return at(path+"/reason", err)
	}
	return nil
}
func observedKind(kind string) bool {
	return kind == "index" || kind == "source" || kind == "sourcetype"
}

func normalizeObservation(input *ObservationScope, objects []Object, collections map[string]Collection, start, end time.Time) (*ObservationScope, error) {
	if input == nil {
		for kind := range collections {
			if observedKind(kind) {
				return nil, at("/observation", fmt.Errorf("Snapshot v2 observed collections require observation"))
			}
		}
		return nil, nil
	}
	out := *input
	var err error
	out.IndexSelection, err = normalizeSelector(input.IndexSelection, "index")
	if err != nil {
		return nil, at("/observation/index_selection", err)
	}
	for _, field := range []struct{ name, value, want string }{
		{"method", out.Method, "splunk_metadata"}, {"visibility", out.Visibility, "exporting_principal"}, {"time_precision", out.TimePrecision, "bucket_overlap"}, {"absence_meaning", out.AbsenceMeaning, "not_observed"},
		{"enumeration/method", out.Enumeration.Method, "distributed_rest"}, {"enumeration/peer_scope", out.Enumeration.PeerScope, "configured_search_peers"},
	} {
		if field.value != field.want {
			return nil, at("/observation/"+field.name, fmt.Errorf("must be %q", field.want))
		}
	}
	if err := observationCoverage(out.Enumeration.Coverage, out.Enumeration.Reason, "/observation/enumeration"); err != nil {
		return nil, err
	}
	out.Enumeration.Provenance, err = normalizeProvenance(out.Enumeration.Provenance, start, end, "/observation/enumeration/provenance")
	if err != nil {
		return nil, err
	}
	switch out.Window.Mode {
	case "all_retained":
		if out.Window.Earliest != "" || out.Window.Latest != "" {
			return nil, at("/observation/window", fmt.Errorf("all_retained forbids bounds"))
		}
	case "bounded":
		if out.Window.Earliest == "" && out.Window.Latest == "" {
			return nil, at("/observation/window", fmt.Errorf("bounded requires at least one bound"))
		}
		var earliest, latest time.Time
		if out.Window.Earliest != "" {
			if !observationWindowTimestampPattern.MatchString(out.Window.Earliest) {
				return nil, at("/observation/window/earliest", fmt.Errorf("earliest must be an RFC3339Nano UTC timestamp"))
			}
			out.Window.Earliest, earliest, err = normalizeTime(out.Window.Earliest, "earliest")
			if err != nil {
				return nil, at("/observation/window/earliest", err)
			}
		}
		if out.Window.Latest != "" {
			if !observationWindowTimestampPattern.MatchString(out.Window.Latest) {
				return nil, at("/observation/window/latest", fmt.Errorf("latest must be an RFC3339Nano UTC timestamp"))
			}
			out.Window.Latest, latest, err = normalizeTime(out.Window.Latest, "latest")
			if err != nil {
				return nil, at("/observation/window/latest", err)
			}
		}
		if out.Window.Earliest != "" && out.Window.Latest != "" && !earliest.Before(latest) {
			return nil, at("/observation/window", fmt.Errorf("bounds must be strictly ordered"))
		}
	default:
		return nil, at("/observation/window/mode", fmt.Errorf("unsupported observation window mode %q", out.Window.Mode))
	}
	out.UnmatchedIndexes, err = observationSet(input.UnmatchedIndexes, "/observation/unmatched_indexes")
	if err != nil {
		return nil, err
	}
	if out.IndexSelection.All != nil && len(out.UnmatchedIndexes) > 0 {
		return nil, at("/observation/unmatched_indexes", fmt.Errorf("all-index selection cannot have unmatched names"))
	}
	byID := map[string]Object{}
	names := map[string]string{}
	for _, object := range objects {
		byID[object.ID] = object
		if observedKind(object.Kind) {
			key := object.Kind + "\x00" + object.Name
			if id, found := names[key]; found {
				return nil, at("/objects", fmt.Errorf("conflicting %s identity %q in %q and %q", object.Kind, object.Name, id, object.ID))
			}
			names[key] = object.ID
		}
	}
	if input.Indexes == nil {
		return nil, at("/observation/indexes", fmt.Errorf("required array must be present and nonnull"))
	}
	out.Indexes = []ObservationIndex{}
	indexes := map[string]ObservationIndex{}
	selected := map[string]bool{}
	for i, index := range input.Indexes {
		path := fmt.Sprintf("/observation/indexes/%d", i)
		object, found := byID[index.IndexID]
		if !found || object.Kind != "index" {
			return nil, at(path+"/index_id", fmt.Errorf("index_id must reference an index object"))
		}
		if _, found := indexes[index.IndexID]; found {
			return nil, at(path+"/index_id", fmt.Errorf("duplicate observation index"))
		}
		if !inScope(out.IndexSelection, object.Name) {
			return nil, at(path+"/index_id", fmt.Errorf("index outside observation selection"))
		}
		index.CatalogDatatypes, err = observationSet(index.CatalogDatatypes, path+"/catalog_datatypes")
		if err != nil {
			return nil, err
		}
		index.RequiredDatatypes, err = observationSet(index.RequiredDatatypes, path+"/required_datatypes")
		if err != nil {
			return nil, err
		}
		for _, values := range [][]string{index.CatalogDatatypes, index.RequiredDatatypes} {
			for _, datatype := range values {
				if datatype != "event" && datatype != "metric" {
					return nil, at(path, fmt.Errorf("unsupported datatype %q", datatype))
				}
			}
		}
		required := index.CatalogDatatypes
		if len(required) == 0 {
			required = []string{"event", "metric"}
		}
		if !equalObservationSet(required, index.RequiredDatatypes) {
			return nil, at(path+"/required_datatypes", fmt.Errorf("required datatypes must match catalog or event and metric when unreported"))
		}
		indexes[index.IndexID] = index
		selected[object.Name] = true
		out.Indexes = append(out.Indexes, index)
	}
	unmatched := map[string]bool{}
	for _, name := range out.UnmatchedIndexes {
		if !inScope(out.IndexSelection, name) || selected[name] || names["index\x00"+name] != "" {
			return nil, at("/observation/unmatched_indexes", fmt.Errorf("unmatched index must be selected and unconfirmed"))
		}
		unmatched[name] = true
	}
	for _, object := range objects {
		if object.Kind == "index" && inScope(out.IndexSelection, object.Name) && !selected[object.Name] {
			return nil, at("/observation/indexes", fmt.Errorf("selected confirmed index %q is missing", object.Name))
		}
	}
	for _, name := range out.IndexSelection.Values {
		if !selected[name] && !unmatched[name] {
			return nil, at("/observation/unmatched_indexes", fmt.Errorf("unconfirmed selected index %q must be unmatched", name))
		}
	}
	if out.Enumeration.Coverage == "unavailable" && len(indexes) > 0 {
		return nil, at("/observation/indexes", fmt.Errorf("unavailable enumeration cannot confirm indexes"))
	}
	if input.Captures == nil {
		return nil, at("/observation/captures", fmt.Errorf("required array must be present and nonnull"))
	}
	out.Captures = []ObservationCapture{}
	type captureKey struct{ kind, index, datatype string }
	captures := map[captureKey]bool{}
	refs := map[string]bool{}
	counts := map[string]int{}
	unavailable := map[string]int{}
	incomplete := map[string]bool{}
	for i, capture := range input.Captures {
		path := fmt.Sprintf("/observation/captures/%d", i)
		if capture.Kind != "source" && capture.Kind != "sourcetype" {
			return nil, at(path+"/kind", fmt.Errorf("capture kind must be source or sourcetype"))
		}
		if _, declared := collections[capture.Kind]; !declared {
			return nil, at(path+"/kind", fmt.Errorf("capture kind must have a declared collection"))
		}
		index, found := indexes[capture.IndexID]
		if !found {
			return nil, at(path+"/index_id", fmt.Errorf("capture must reference a confirmed selected index"))
		}
		if !containsObservation(index.RequiredDatatypes, capture.Datatype) {
			return nil, at(path+"/datatype", fmt.Errorf("capture datatype is not required by index"))
		}
		key := captureKey{capture.Kind, capture.IndexID, capture.Datatype}
		if captures[key] {
			return nil, at(path, fmt.Errorf("duplicate capture key"))
		}
		captures[key] = true
		capture.ObjectIDs, err = observationSet(capture.ObjectIDs, path+"/object_ids")
		if err != nil {
			return nil, err
		}
		if err := observationCoverage(capture.Coverage, capture.Reason, path); err != nil {
			return nil, err
		}
		if capture.Coverage == "unavailable" && len(capture.ObjectIDs) > 0 {
			return nil, at(path+"/object_ids", fmt.Errorf("unavailable capture cannot contain objects"))
		}
		for _, id := range capture.ObjectIDs {
			object, found := byID[id]
			if !found || object.Kind != capture.Kind {
				return nil, at(path+"/object_ids", fmt.Errorf("capture reference must name a %s object", capture.Kind))
			}
			refs[id] = true
		}
		capture.Provenance, err = normalizeProvenance(capture.Provenance, start, end, path+"/provenance")
		if err != nil {
			return nil, err
		}
		counts[capture.Kind]++
		if capture.Coverage == "unavailable" {
			unavailable[capture.Kind]++
		}
		if capture.Coverage != "complete" {
			incomplete[capture.Kind] = true
		}
		out.Captures = append(out.Captures, capture)
	}
	sort.Slice(out.Indexes, func(i, j int) bool { return out.Indexes[i].IndexID < out.Indexes[j].IndexID })
	for _, index := range out.Indexes {
		id := index.IndexID
		for _, datatype := range index.RequiredDatatypes {
			for _, kind := range []string{"source", "sourcetype"} {
				if _, declared := collections[kind]; !declared {
					continue
				}
				if !captures[captureKey{kind, id, datatype}] {
					return nil, at("/observation/captures", fmt.Errorf("missing required %s capture for %s/%s", kind, id, datatype))
				}
			}
		}
	}
	for _, object := range objects {
		if (object.Kind == "source" || object.Kind == "sourcetype") && !refs[object.ID] {
			return nil, at("/observation/captures", fmt.Errorf("unassociated %s object %q", object.Kind, object.ID))
		}
	}
	for _, kind := range []string{"index", "source", "sourcetype"} {
		expected := out.Enumeration.Coverage
		if kind != "index" && expected != "unavailable" {
			if counts[kind] > 0 && counts[kind] == unavailable[kind] {
				expected = "unavailable"
			} else if incomplete[kind] || len(out.UnmatchedIndexes) > 0 {
				expected = "partial"
			}
		}
		if collection, found := collections[kind]; found {
			if collection.Coverage != expected {
				return nil, at("/collections", fmt.Errorf("%s collection coverage %q contradicts acquisition coverage %q", kind, collection.Coverage, expected))
			}
		} else if kind == "index" && len(indexes) > 0 || kind != "index" && counts[kind] > 0 {
			return nil, at("/collections", fmt.Errorf("observation requires %s collection", kind))
		}
	}
	sort.Slice(out.Captures, func(i, j int) bool {
		a, b := out.Captures[i], out.Captures[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.IndexID != b.IndexID {
			return a.IndexID < b.IndexID
		}
		return a.Datatype < b.Datatype
	})
	return &out, nil
}
func equalObservationSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func containsObservation(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func validateObservationRawShape(raw []byte) error {
	var observation map[string]json.RawMessage
	if err := json.Unmarshal(raw, &observation); err != nil || observation == nil {
		return at("/observation", fmt.Errorf("observation must be an object"))
	}
	require := func(object map[string]json.RawMessage, path string, keys []string) error {
		for _, key := range keys {
			if _, found := object[key]; !found {
				return at(path+"/"+key, fmt.Errorf("missing property %q", key))
			}
		}
		return nil
	}
	if err := require(observation, "/observation", []string{"index_selection", "enumeration", "indexes", "unmatched_indexes", "method", "visibility", "window", "time_precision", "absence_meaning", "captures"}); err != nil {
		return err
	}
	var selector map[string]json.RawMessage
	_ = json.Unmarshal(observation["index_selection"], &selector)
	_, all := selector["all"]
	_, values := selector["values"]
	if all == values {
		return at("/observation/index_selection", fmt.Errorf("index selector must use exactly one branch"))
	}
	var window map[string]json.RawMessage
	_ = json.Unmarshal(observation["window"], &window)
	var mode string
	_ = json.Unmarshal(window["mode"], &mode)
	for _, bound := range []string{"earliest", "latest"} {
		if rawBound, found := window[bound]; found {
			var text string
			_ = json.Unmarshal(rawBound, &text)
			if mode == "all_retained" || text == "" {
				return at("/observation/window/"+bound, fmt.Errorf("bound is inapplicable or empty"))
			}
		}
	}
	var indexes []map[string]json.RawMessage
	if err := json.Unmarshal(observation["indexes"], &indexes); err == nil {
		for i, index := range indexes {
			if err := require(index, fmt.Sprintf("/observation/indexes/%d", i), []string{"index_id", "catalog_datatypes", "required_datatypes"}); err != nil {
				return err
			}
		}
	}
	var captures []map[string]json.RawMessage
	if err := json.Unmarshal(observation["captures"], &captures); err == nil {
		for i, capture := range captures {
			if err := require(capture, fmt.Sprintf("/observation/captures/%d", i), []string{"kind", "index_id", "datatype", "object_ids", "coverage", "provenance"}); err != nil {
				return err
			}
		}
	}
	return nil
}
