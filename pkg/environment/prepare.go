package environment

import (
	"encoding/json"
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// PreparedEnvironment owns detached indexes for one snapshot and its compatible bindings.
type PreparedEnvironment struct {
	snapshot    *PreparedSnapshot
	schemas     *PreparedSchemaBundle
	report      Report
	objects     map[string]Object
	collections map[string]Collection
	bindings    map[string][]SchemaBinding
}

// Pair checks every binding against this snapshot. A bundle can be paired again
// with another snapshot without carrying linkage decisions between captures.
func Pair(snapshot *PreparedSnapshot, schemas *PreparedSchemaBundle) (*PreparedEnvironment, *Report, error) {
	if snapshot == nil {
		return nil, &Report{SchemaVersion: 1, Status: "invalid", Coverage: []CoverageEntry{}, Diagnostics: []Diagnostic{{Code: "snapshot_missing", Severity: "error", Artifact: "snapshot", Message: "prepared snapshot is required"}}}, nil
	}
	report := snapshot.Report()
	if schemas != nil {
		part := schemas.Report()
		mergeEnvironmentReport(&report, &part)
	}
	env := &PreparedEnvironment{
		snapshot: snapshot, schemas: schemas, objects: make(map[string]Object, len(snapshot.snapshot.Objects)),
		collections: make(map[string]Collection, len(snapshot.snapshot.Collections)), bindings: map[string][]SchemaBinding{},
	}
	for _, object := range snapshot.snapshot.Objects {
		env.objects[object.ID] = object
	}
	for _, collection := range snapshot.snapshot.Collections {
		env.collections[collection.Kind] = collection
	}
	if schemas != nil {
		for _, binding := range schemas.bundle.Bindings {
			path := "/bindings"
			if object, found := env.objects[binding.ObjectID]; found {
				if !sameObjectIdentity(object, binding.Expected) {
					report.Diagnostics = append(report.Diagnostics, Diagnostic{Code: "binding_identity_mismatch", Severity: "error", Artifact: "schema_bundle", Path: path, Message: fmt.Sprintf("object %q identity differs from schema %q binding", binding.ObjectID, binding.SchemaID)})
					report.Status = "invalid"
					continue
				}
				env.bindings[binding.ObjectID] = append(env.bindings[binding.ObjectID], binding)
				continue
			}
			collection, declared := env.collections[binding.Expected.Kind]
			if bindingInScope(snapshot.snapshot.CaptureScope, binding.Expected) && declared && collection.Coverage == "complete" && !hasObservedAbsence(snapshot.snapshot, binding.Expected.Kind) {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Code: "binding_object_absent", Severity: "error", Artifact: "schema_bundle", Path: path, Message: fmt.Sprintf("object %q is absent from complete %s collection", binding.ObjectID, binding.Expected.Kind)})
				report.Status = "invalid"
			} else {
				report.Diagnostics = append(report.Diagnostics, Diagnostic{Code: "binding_unresolved", Severity: "warning", Artifact: "schema_bundle", Path: path, Message: fmt.Sprintf("object %q is not available in this snapshot", binding.ObjectID)})
				if report.Status == "valid" {
					report.Status = "partial"
				}
			}
		}
	}
	if report.Coverage == nil {
		report.Coverage = []CoverageEntry{}
	}
	if report.Diagnostics == nil {
		report.Diagnostics = []Diagnostic{}
	}
	if report.Status == "invalid" {
		return nil, &report, nil
	}
	env.report = copyEnvironmentReport(report)
	return env, &report, nil
}

func hasObservedAbsence(snapshot Snapshot, kind string) bool {
	return snapshot.SchemaVersion == 2 && snapshot.Observation != nil && snapshot.Observation.AbsenceMeaning == "not_observed" && (kind == "source" || kind == "sourcetype")
}

func sameObjectIdentity(object Object, expected ObjectIdentity) bool {
	return object.Kind == expected.Kind && object.Name == expected.Name && object.Namespace == expected.Namespace && object.App == expected.App && object.Owner == expected.Owner
}

func bindingInScope(scope CaptureScope, expected ObjectIdentity) bool {
	if expected.Kind == "index" || expected.Kind == "source" || expected.Kind == "sourcetype" {
		return true
	}
	return inScope(scope.Namespace, expected.Namespace) && inScope(scope.App, expected.App) && inScope(scope.Owner, expected.Owner)
}

// Object returns a detached copy of a captured object by stable ID.
func (p *PreparedEnvironment) Object(id string) (Object, bool) {
	if p == nil {
		return Object{}, false
	}
	object, found := p.objects[id]
	if !found {
		return Object{}, false
	}
	raw, _ := json.Marshal(object)
	var out Object
	_ = json.Unmarshal(raw, &out)
	return out, true
}

// Collection returns coverage for a declared collection kind.
func (p *PreparedEnvironment) Collection(kind string) (Collection, bool) {
	if p == nil {
		return Collection{}, false
	}
	collection, found := p.collections[kind]
	return collection, found
}

// Bindings returns detached, independently valid schema bindings for an object.
func (p *PreparedEnvironment) Bindings(objectID string) []SchemaBinding {
	if p == nil {
		return []SchemaBinding{}
	}
	return append([]SchemaBinding{}, p.bindings[objectID]...)
}

func (p *PreparedEnvironment) boundTarget(objectID, schemaID string) (preparedBundleTarget, bool) {
	if p == nil || p.schemas == nil {
		return preparedBundleTarget{}, false
	}
	for _, binding := range p.bindings[objectID] {
		if binding.SchemaID == schemaID {
			target, found := p.schemas.targets[schemaID]
			return target, found
		}
	}
	return preparedBundleTarget{}, false
}

// FieldCatalog returns the compiled field list for a resolved object binding.
func (p *PreparedEnvironment) FieldCatalog(objectID, schemaID string) (*validation.PreparedFieldCatalog, bool) {
	target, found := p.boundTarget(objectID, schemaID)
	return target.field, found && target.field != nil
}

// SchemaTarget returns the compiled JSON Schema or OCSF target for a resolved object binding.
func (p *PreparedEnvironment) SchemaTarget(objectID, schemaID string) (*validation.PreparedSchemaTarget, bool) {
	target, found := p.boundTarget(objectID, schemaID)
	return target.schema, found && target.schema != nil
}

// Report returns a detached copy of the pairing report.
func (p *PreparedEnvironment) Report() Report {
	if p == nil {
		return Report{}
	}
	return copyEnvironmentReport(p.report)
}
