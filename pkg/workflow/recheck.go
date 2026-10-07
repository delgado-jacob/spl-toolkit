package workflow

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
)

// RecheckJSON admits caller configuration and context artifacts using the same
// strict wire and artifact-owner boundaries as ordinary workflow assessment.
func RecheckJSON(raw []byte) (*RecheckReport, error) {
	_, artifacts, bases, err := readWire(raw, reflect.TypeOf(RecheckRequest{}))
	if err != nil {
		return nil, err
	}
	var admitted Settings
	if err := decodeArtifacts(&admitted, artifacts, bases, "/context"); err != nil {
		return nil, err
	}
	var request RecheckRequest
	if err := json.Unmarshal(raw, &request); err != nil {
		return nil, requestErrorAt("request_invalid", "", err.Error())
	}
	request.Context.Snapshot, request.Context.SchemaBundle = admitted.Snapshot, admitted.SchemaBundle
	return Recheck(request)
}

// Recheck performs a fresh assessment with new compatibility and resolution
// sessions. Original hashes link source bytes only; no saved evidence grants proof.
func Recheck(request RecheckRequest) (*RecheckReport, error) {
	if !validUTF8(reflect.ValueOf(request)) {
		return nil, requestErrorAt("request_invalid", "", "request contains invalid UTF-8")
	}
	if request.SchemaVersion != 1 {
		return nil, requestErrorAt("request_invalid", "/schema_version", "schema_version must be integer 1")
	}
	original := request.Context.Original
	if strings.TrimSpace(original.ID) == "" {
		return nil, requestErrorAt("request_invalid", "/context/original/id", "id must be nonblank")
	}
	if _, err := capabilityselector.Normalize(original.Document.Language, original.Document.Profile, original.Document.Version); err != nil {
		return nil, requestErrorAt("request_invalid", "/context/original/document", err.Error())
	}
	proposal := request.Proposal
	if proposal.Settings.ID != original.ID {
		return nil, requestErrorAt("request_invalid", "/proposal/settings/id", "settings id must match original detection id")
	}
	if (proposal.Settings.Compatibility == nil) == (proposal.Settings.Resolution == nil) {
		return nil, requestErrorAt("request_invalid", "/proposal/settings", "exactly one mode payload is required")
	}
	if proposal.Settings.Compatibility != nil && proposal.Document == nil {
		return nil, requestErrorAt("request_invalid", "/proposal/document", "query proposal requires a complete replacement document")
	}
	if proposal.Settings.Resolution != nil && proposal.Document != nil {
		return nil, requestErrorAt("request_invalid", "/proposal/document", "resolution proposal cannot replace the original document")
	}
	document := original.Document
	if proposal.Document != nil {
		document = *proposal.Document
	}
	assessment, err := Assess(Request{SchemaVersion: 1, Documents: []corpus.RequestDocument{{ID: original.ID, Document: document}}, Settings: Settings{SchemaVersion: 1, Snapshot: request.Context.Snapshot, SchemaBundle: request.Context.SchemaBundle, Entries: []EntrySettings{proposal.Settings}}})
	if err != nil {
		return nil, err
	}
	return detachedExport(&RecheckReport{SchemaVersion: 1, DetectionID: original.ID, OriginalSourceHash: corpus.SourceHash(original.Document.Text), ProposedSourceHash: corpus.SourceHash(document.Text), Assessment: *assessment})
}
