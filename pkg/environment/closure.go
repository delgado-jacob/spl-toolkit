package environment

import (
	"fmt"

	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

// DefinitionBundle projects captured knowledge objects into the caller's
// resolution scope. Coverage is complete only when that scope was captured.
func (p *PreparedEnvironment) DefinitionBundle(queryScope CaptureScope) (closure.DefinitionBundle, error) {
	if p == nil || p.snapshot == nil {
		return closure.DefinitionBundle{}, fmt.Errorf("prepared environment is required")
	}
	var err error
	queryScope.Namespace, err = normalizeSelector(queryScope.Namespace, "namespace")
	if err != nil {
		return closure.DefinitionBundle{}, fmt.Errorf("query namespace: %w", err)
	}
	queryScope.App, err = normalizeSelector(queryScope.App, "app")
	if err != nil {
		return closure.DefinitionBundle{}, fmt.Errorf("query app: %w", err)
	}
	queryScope.Owner, err = normalizeSelector(queryScope.Owner, "owner")
	if err != nil {
		return closure.DefinitionBundle{}, fmt.Errorf("query owner: %w", err)
	}
	snapshot := p.snapshot.Snapshot()
	covered := selectorSubset(queryScope.Namespace, snapshot.CaptureScope.Namespace) &&
		selectorSubset(queryScope.App, snapshot.CaptureScope.App) &&
		selectorSubset(queryScope.Owner, snapshot.CaptureScope.Owner)
	bundle := closure.DefinitionBundle{SchemaVersion: 1, ScopeID: snapshot.ScopeID, Collections: []closure.Collection{}, Objects: []closure.Definition{}}
	for _, collection := range snapshot.Collections {
		if !queryKinds[collection.Kind] {
			continue
		}
		coverage := collection.Coverage
		if !covered && coverage == "complete" {
			coverage = "partial"
		}
		bundle.Collections = append(bundle.Collections, closure.Collection{Kind: collection.Kind, Coverage: coverage})
	}
	for _, object := range snapshot.Objects {
		if !queryKinds[object.Kind] || !inScope(queryScope.Namespace, object.Namespace) ||
			!inScope(queryScope.App, object.App) || !inScope(queryScope.Owner, object.Owner) {
			continue
		}
		bundle.Objects = append(bundle.Objects, closure.Definition{
			ID: object.ID, Kind: object.Kind, Name: object.Name,
			App: object.App, Owner: object.Owner, Sharing: object.Sharing,
			SourceID: object.Provenance.SourceID, Document: object.Document,
			Arity: object.Arity, Arguments: object.Arguments, EvalBased: object.EvalBased,
			Validation: object.Validation, Relations: object.Relations,
		})
	}
	return bundle, nil
}

func selectorSubset(query, capture Selector) bool {
	if capture.All != nil {
		return true
	}
	if query.All != nil {
		return false
	}
	for _, value := range query.Values {
		if !inScope(capture, value) {
			return false
		}
	}
	return true
}
