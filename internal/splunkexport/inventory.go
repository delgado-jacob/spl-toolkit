package splunkexport

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// IndexResult retains the entire confirmed catalog and only the selected
// observation indexes. UnmatchedIndexes includes every unconfirmed requested name.
// CleanupFailed is operational evidence independent of collection coverage.
type IndexResult struct {
	Objects          []environment.Object
	Indexes          []environment.ObservationIndex
	UnmatchedIndexes []string
	Enumeration      environment.IndexEnumeration
	Diagnostics      []Diagnostic
	CleanupFailed    bool
}

// MetadataResult contains the unique union of capture objects, all required
// capture keys (including unavailable work), and source/sourcetype collections.
// Task 7 combines these with IndexResult and the normalized options for observation.
type MetadataResult struct {
	Objects       []environment.Object
	Captures      []environment.ObservationCapture
	Collections   []environment.Collection
	Diagnostics   []Diagnostic
	CleanupFailed bool
}

// stableObjectID is shared with configuration acquisition. Context is empty for
// global observed identities, and non-macro callers pass a nil arity.
func stableObjectID(instanceID, kind, name, namespace, app, owner string, macroArity *int) string {
	if kind != "macro" {
		macroArity = nil
	}
	return inventoryHash([]any{instanceID, kind, name, namespace, app, owner, macroArity})
}
func inventoryHash(identity []any) string {
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// acquisitionProvenance hashes acquisition identity without retaining an origin,
// job SID, or credentials in SourceID. The caller supplies the acquisition time.
func acquisitionProvenance(instanceID, sourceKind, acquisitionIdentity, observedAt string) environment.Provenance {
	return environment.Provenance{SourceKind: sourceKind, SourceID: inventoryHash([]any{instanceID, sourceKind, acquisitionIdentity}), ObservedAt: observedAt}
}
func indexDiscoveryQuery(timeout time.Duration) discoveryQuery {
	seconds := int64(math.Ceil(timeout.Seconds()))
	if seconds <= 0 {
		seconds = 30
	}
	return discoveryQuery{search: "| rest splunk_server=* /services/data/indexes datatype=all count=0 timeout=" + strconv.FormatInt(seconds, 10) + " | stats values(datatype) as datatypes by title | rename title as index | fields index datatypes"}
}

var indexLiteralPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]*$`)

func metadataDiscoveryQuery(kind, datatype, index string) (discoveryQuery, error) {
	if !indexLiteralPattern.MatchString(index) {
		return discoveryQuery{}, failure("index_literal_unsupported")
	}
	if kind != "source" && kind != "sourcetype" || datatype != "event" && datatype != "metric" {
		return discoveryQuery{}, failure("metadata_mode_unsupported")
	}
	return discoveryQuery{search: "| metadata type=" + kind + "s datatype=" + datatype + " index=\"" + index + "\" | fields " + kind}, nil
}
func identityName(raw json.RawMessage) (string, bool) {
	var name string
	if json.Unmarshal(raw, &name) != nil || strings.TrimSpace(name) == "" || !utf8.ValidString(name) {
		return "", false
	}
	return name, true
}
func inventoryDiagnostic(code, kind, indexID, datatype string) Diagnostic {
	return Diagnostic{Code: code, Severity: "warning", Kind: kind, IndexID: indexID, Datatype: datatype, Message: "Inventory acquisition did not establish complete evidence for the declared scope."}
}
func markInventoryGap(coverage, reason *string, code string, usable int) {
	*coverage = "partial"
	if usable == 0 {
		*coverage = "unavailable"
	}
	if *reason == "" {
		*reason = code
	}
}

// catalogModes distinguishes unreported modes from malformed/unknown evidence.
// Either forces conservative event+metric work; malformed evidence also marks a
// catalog gap. A merely unreported mode need not degrade successful observation.
func catalogModes(raw json.RawMessage) (modes []string, unreported, invalid bool) {
	modes = []string{}
	if len(raw) == 0 || string(raw) == "null" {
		return modes, true, false
	}
	var values []string
	var scalar string
	if json.Unmarshal(raw, &scalar) == nil {
		if scalar == "" {
			return modes, true, false
		}
		values = []string{scalar}
	} else if json.Unmarshal(raw, &values) != nil {
		return modes, true, true
	}
	if len(values) == 0 {
		return modes, true, false
	}
	seen := map[string]bool{}
	for _, value := range values {
		if value != "event" && value != "metric" {
			return []string{}, true, true
		}
		seen[value] = true
	}
	for value := range seen {
		modes = append(modes, value)
	}
	sort.Strings(modes)
	return modes, false, false
}
func (c *Client) collectIndexes(ctx context.Context, instanceID string) IndexResult {
	job := c.runDiscoveryJob(ctx, indexDiscoveryQuery(c.options.RequestTimeout))
	provenance := acquisitionProvenance(instanceID, "splunk_rest", "distributed_index_catalog", job.ObservedAt)
	result := IndexResult{Objects: []environment.Object{}, Indexes: []environment.ObservationIndex{}, UnmatchedIndexes: []string{}, Enumeration: environment.IndexEnumeration{Method: "distributed_rest", PeerScope: "configured_search_peers", Coverage: job.Coverage, Reason: job.Reason, Provenance: provenance}, Diagnostics: append([]Diagnostic{}, job.Diagnostics...), CleanupFailed: job.CleanupFailed}
	type catalogEntry struct {
		object     environment.Object
		modes      map[string]bool
		unreported bool
	}
	catalog := map[string]*catalogEntry{}
	invalidRows := false
	for _, row := range job.Rows {
		name, ok := identityName(row["index"])
		if !ok {
			invalidRows = true
			result.Diagnostics = append(result.Diagnostics, inventoryDiagnostic("index_identity_invalid", "index", "", ""))
			continue
		}
		entry := catalog[name]
		if entry == nil {
			entry = &catalogEntry{object: environment.Object{ID: stableObjectID(instanceID, "index", name, "", "", "", nil), Kind: "index", Name: name, Provenance: provenance}, modes: map[string]bool{}}
			catalog[name] = entry
		}
		modes, unreported, invalid := catalogModes(row["datatypes"])
		entry.unreported = entry.unreported || unreported
		for _, mode := range modes {
			entry.modes[mode] = true
		}
		if unreported {
			code := "index_datatype_unreported"
			if invalid {
				code = "index_datatype_invalid"
			}
			result.Diagnostics = append(result.Diagnostics, inventoryDiagnostic(code, "index", entry.object.ID, ""))
		}
		invalidRows = invalidRows || invalid
	}
	if invalidRows {
		markInventoryGap(&result.Enumeration.Coverage, &result.Enumeration.Reason, "index_catalog_invalid", len(catalog))
	}
	// A job with retained valid rows confirms those indexes even after an acquisition gap.
	if len(catalog) > 0 && result.Enumeration.Coverage == "unavailable" {
		result.Enumeration.Coverage = "partial"
	}
	allIndexes := []environment.ObservationIndex{}
	for _, entry := range catalog {
		modes := []string{}
		if !entry.unreported {
			for mode := range entry.modes {
				modes = append(modes, mode)
			}
			sort.Strings(modes)
		}
		required := append([]string{}, modes...)
		if len(required) == 0 {
			required = []string{"event", "metric"}
		}
		result.Objects = append(result.Objects, entry.object)
		allIndexes = append(allIndexes, environment.ObservationIndex{IndexID: entry.object.ID, CatalogDatatypes: modes, RequiredDatatypes: required})
	}
	sort.Slice(result.Objects, func(i, j int) bool { return result.Objects[i].ID < result.Objects[j].ID })
	result.Indexes, result.UnmatchedIndexes = selectIndexes(c.options.IndexSelection, result.Objects, allIndexes)
	return result
}
func selectIndexes(selection environment.Selector, objects []environment.Object, indexes []environment.ObservationIndex) ([]environment.ObservationIndex, []string) {
	names := map[string]string{}
	for _, o := range objects {
		names[o.Name] = o.ID
	}
	selected := map[string]bool{}
	unmatchedSet := map[string]bool{}
	if selection.All != nil && *selection.All {
		for _, o := range objects {
			selected[o.ID] = true
		}
	} else {
		for _, name := range selection.Values {
			if id, ok := names[name]; ok {
				selected[id] = true
			} else {
				unmatchedSet[name] = true
			}
		}
	}
	out := []environment.ObservationIndex{}
	unmatched := []string{}
	for _, index := range indexes {
		if selected[index.IndexID] {
			out = append(out, index)
		}
	}
	for name := range unmatchedSet {
		unmatched = append(unmatched, name)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].IndexID < out[j].IndexID })
	sort.Strings(unmatched)
	return out, unmatched
}
func (c *Client) collectMetadata(ctx context.Context, instanceID string, indexes IndexResult) MetadataResult {
	result := MetadataResult{Objects: []environment.Object{}, Captures: []environment.ObservationCapture{}, Collections: []environment.Collection{}, Diagnostics: []Diagnostic{}}
	names := map[string]string{}
	for _, o := range indexes.Objects {
		names[o.ID] = o.Name
	}
	objects := map[string]environment.Object{}
	for _, index := range indexes.Indexes {
		for _, datatype := range index.RequiredDatatypes {
			for _, kind := range []string{"source", "sourcetype"} {
				capture := environment.ObservationCapture{Kind: kind, IndexID: index.IndexID, Datatype: datatype, ObjectIDs: []string{}, Coverage: "unavailable"}
				query, err := metadataDiscoveryQuery(kind, datatype, names[index.IndexID])
				job := JobResult{Rows: []map[string]json.RawMessage{}, Coverage: "unavailable", ObservedAt: time.Now().UTC().Format(time.RFC3339Nano)}
				if err != nil {
					job.Reason = err.Error()
				} else if ctx.Err() != nil {
					job.Reason = "inventory_not_attempted"
				} else {
					job = c.runDiscoveryJob(ctx, query)
				}
				// Include the window in acquisition identity; source object identity stays global.
				acquisition, _ := json.Marshal([]string{kind, index.IndexID, datatype, c.options.Window.Mode, c.options.Window.Earliest, c.options.Window.Latest})
				capture.Provenance = acquisitionProvenance(instanceID, "splunk_metadata", string(acquisition), job.ObservedAt)
				capture.Coverage, capture.Reason = job.Coverage, job.Reason
				result.CleanupFailed = result.CleanupFailed || job.CleanupFailed
				for _, d := range job.Diagnostics {
					d.Kind = kind
					d.IndexID = index.IndexID
					d.Datatype = datatype
					result.Diagnostics = append(result.Diagnostics, d)
				}
				ids := map[string]bool{}
				invalid := false
				for _, row := range job.Rows {
					name, ok := identityName(row[kind])
					if !ok {
						invalid = true
						continue
					}
					object := environment.Object{ID: stableObjectID(instanceID, kind, name, "", "", "", nil), Kind: kind, Name: name, Provenance: capture.Provenance}
					// Shared global identities are not conflicts when acquisition provenance differs.
					if existing, found := objects[object.ID]; !found || object.Provenance.SourceID < existing.Provenance.SourceID {
						objects[object.ID] = object
					}
					ids[object.ID] = true
				}
				for id := range ids {
					capture.ObjectIDs = append(capture.ObjectIDs, id)
				}
				sort.Strings(capture.ObjectIDs)
				if invalid {
					markInventoryGap(&capture.Coverage, &capture.Reason, "metadata_identity_invalid", len(ids))
					result.Diagnostics = append(result.Diagnostics, inventoryDiagnostic("metadata_identity_invalid", kind, index.IndexID, datatype))
				}
				if capture.Coverage != "complete" && len(job.Diagnostics) == 0 && !invalid {
					result.Diagnostics = append(result.Diagnostics, inventoryDiagnostic(capture.Reason, kind, index.IndexID, datatype))
				}
				result.Captures = append(result.Captures, capture)
			}
		}
	}
	for _, object := range objects {
		result.Objects = append(result.Objects, object)
	}
	sort.Slice(result.Objects, func(i, j int) bool { return result.Objects[i].ID < result.Objects[j].ID })
	sort.Slice(result.Captures, func(i, j int) bool {
		a, b := result.Captures[i], result.Captures[j]
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.IndexID != b.IndexID {
			return a.IndexID < b.IndexID
		}
		return a.Datatype < b.Datatype
	})
	for _, kind := range []string{"source", "sourcetype"} {
		result.Collections = append(result.Collections, metadataCollection(kind, indexes, result.Captures))
	}
	return result
}
func metadataCollection(kind string, indexes IndexResult, captures []environment.ObservationCapture) environment.Collection {
	coverage, reason := indexes.Enumeration.Coverage, indexes.Enumeration.Reason
	count, unavailable := 0, 0
	incompleteReason := ""
	for _, capture := range captures {
		if capture.Kind != kind {
			continue
		}
		count++
		if capture.Coverage == "unavailable" {
			unavailable++
		}
		if capture.Coverage != "complete" && incompleteReason == "" {
			incompleteReason = capture.Reason
		}
	}
	if coverage != "unavailable" {
		if count > 0 && count == unavailable {
			coverage = "unavailable"
			reason = incompleteReason
		} else if incompleteReason != "" || len(indexes.UnmatchedIndexes) > 0 {
			coverage = "partial"
			if reason == "" {
				reason = incompleteReason
				if reason == "" {
					reason = "requested_index_unmatched"
				}
			}
		}
	}
	return environment.Collection{Kind: kind, Coverage: coverage, Reason: reason}
}
