package workflow

import (
	"reflect"
	"strconv"
	"strings"
)

var disclosureCategories = []string{"query_text", "requirement_names", "source_identity", "environment_metadata", "definitions", "diagnostic_details", "artifact_identity"}

func admitDisclosure(include []string) (map[string]bool, error) {
	if include == nil {
		return nil, requestErrorAt("request_invalid", "/include", "include must be an array")
	}
	out := map[string]bool{}
	for _, c := range include {
		known := false
		for _, allowed := range disclosureCategories {
			known = known || c == allowed
		}
		if !known || out[c] {
			return nil, requestErrorAt("request_invalid", "/include", "include contains duplicate or unsupported categories")
		}
		out[c] = true
	}
	return out, nil
}
func publicEnum(value string, allowed ...string) string {
	for _, a := range allowed {
		if value == a {
			return a
		}
	}
	return "unrecognized"
}
func publicOutcome(value string) string {
	return publicEnum(value, "valid", "invalid", "incomplete", "satisfied", "unsatisfied", "not assessed", "not_applicable", "indeterminate", "verified", "failed", "affected", "unchanged", "missing", "ambiguous", "neutral", "conditional", "required", "optional")
}
func publicState(value string) string {
	return publicEnum(value, "complete", "partial", "not_applicable", "unavailable", "incomplete", "omitted", "unknown", "not_captured")
}
func publicDimension(value string) string {
	return publicEnum(value, "environment_collection", "environment_capability", "field_schema", "target_discovery", "field_attribution", "query_semantics", "dependency_closure", "correlation", "correlation_analysis", "requirements", "syntax", "semantics", "resolution", "collections", "expansion", "definitions", "effective_query", "traversed_definitions")
}
func booleanCoverage(complete bool) string {
	if complete {
		return "complete"
	}
	return "incomplete"
}
func publicRevision(value string) string {
	if savedDigest(value) {
		return value
	}
	return "unrecognized"
}

// This vocabulary is deliberately local. Imported text is never promoted to a
// public code merely because it resembles a toolkit code or passes a regex.
func publicCode(value string) string {
	switch value {
	case "SPL_AMBIGUOUS_FIELD",
		"SPL_ANALYSIS_RESOURCE_LIMIT",
		"SPL_DECLARATION_CYCLE",
		"SPL_DUPLICATE_SYMBOL",
		"SPL_DYNAMIC_REFERENCE",
		"SPL_INVALID_FUNCTION_CALL",
		"SPL_REQUIREMENT_COVERAGE_INCOMPLETE",
		"SPL_REQUIREMENT_DYNAMIC",
		"SPL_REQUIREMENT_INDETERMINATE",
		"SPL_SYNTAX_ERROR",
		"SPL_UNAVAILABLE_FIELD",
		"SPL_UNRESOLVED_MODULE",
		"SPL_UNRESOLVED_SYMBOL",
		"SPL_UNRESOLVED_WILDCARD",
		"SPL_UNSUPPORTED_COMMAND",
		"SPL_UNSUPPORTED_FUNCTION",
		"SPL_UNSUPPORTED_SEMANTICS",
		"ambiguous_object",
		"analysis_error",
		"analysis_incomplete",
		"artifact_missing",
		"assessment_invariant",
		"binding_identity_mismatch",
		"binding_invalid",
		"binding_object_absent",
		"binding_unresolved",
		"dataset_identity_unmapped",
		"environment_capability_not_supplied", "environment_capability_version_unproven", "environment_capability_unavailable", "environment_capability_unknown",
		"environment_collection_partial", "environment_collection_unavailable", "environment_object_ambiguous", "environment_object_missing", "environment_scope_not_covered",
		"field_ownership_ambiguous", "schema_projection_indeterminate",
		"entry_failure", "query_input_mismatch", "meaningful_evidence_changed", "evidence_incomplete", "correspondence_unresolved",
		"capability_available",
		"capability_unavailable",
		"collection_incomplete",
		"collection_not_captured",
		"collection_omitted",
		"collection_partial",
		"collection_unknown",
		"configuration_invalid",
		"correlation_incomplete",
		"correlation_unproved",
		"cycle",
		"definition_analysis_error",
		"dependency_closure_incomplete",
		"dependency_closure_invalid",
		"dynamic_reference",
		"fanout_limit_exceeded",
		"field_attribution_incomplete",
		"macro_context_missing",
		"missing_definition_body",
		"missing_input_binding",
		"object_absent",
		"object_present",
		"object_unresolved",
		"observation_scope_insufficient",
		"operation_failed",
		"render_conflict",
		"render_unproved",
		"request_invalid",
		"requirements_inconsistent",
		"requirements_refresh_required",
		"requirements_stale",
		"resolution_discovery_incomplete",
		"resolution_invariant",
		"resolution_kind_mismatch",
		"resolution_missing",
		"resolution_proof_invalid",
		"resolution_role_incomplete",
		"resolution_unused",
		"schema_bundle_invalid",
		"schema_field_missing",
		"schema_not_supplied",
		"schema_partial",
		"selection_mismatch",
		"snapshot_invalid",
		"snapshot_missing",
		"source_partial",
		"substitution_incomplete",
		"target_discovery_incomplete",
		"target_not_renderable",
		"traversal_failed",
		"unproved_owner",
		"unsupported_semantics":
		return value
	default:
		return "unrecognized_code"
	}
}

func jsonFieldName(f reflect.StructField) string {
	name := strings.Split(f.Tag.Get("json"), ",")[0]
	if name == "-" {
		return ""
	}
	return name
}
func isAbsent(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map:
		return v.IsNil()
	}
	return false
}

// Select only named fields of known exported owners. In particular, raw JSON
// and entire reports never become catch-all Details payloads.
func disclosureField(owner, field string, definition bool) string {
	if strings.Contains(field, "Digest") || field == "SourceHash" {
		return "artifact_identity"
	}
	if field == "Code" || field == "Message" || field == "Detail" || field == "DiagnosticCodes" || field == "Reason" {
		return "diagnostic_details"
	}
	if field == "SourceID" || strings.HasSuffix(field, "Path") || strings.HasSuffix(field, "URI") || strings.HasSuffix(field, "URIs") || field == "ID" || strings.HasSuffix(field, "ObjectID") || field == "SchemaID" || field == "InstanceID" || field == "ObjectIDs" || field == "IndexID" || field == "IgnoredNames" || field == "SkippedSymlinks" {
		return "source_identity"
	}
	switch owner {
	case "QueryDocument":
		if field == "Text" {
			if definition {
				return "definitions"
			}
			return "query_text"
		}
	case "Variant":
		if field == "CandidateText" || field == "ResolvedQuery" {
			return "query_text"
		}
	case "RequirementItem":
		if field == "Identity" || field == "FieldIdentity" {
			return "requirement_names"
		}
	case "QueryInput":
		if field == "Name" || field == "Identity" {
			return "requirement_names"
		}
	case "Reference":
		if field == "OriginalName" || field == "NormalizedName" || field == "FieldIdentity" {
			return "requirement_names"
		}
	case "RequirementOccurrence":
		if field == "OriginalName" {
			return "requirement_names"
		}
	case "Resolution", "ResolutionChoice", "ResolutionPlaceholder":
		if field == "Placeholder" || field == "Values" || field == "Value" {
			return "requirement_names"
		}
	case "Object":
		switch field {
		case "Kind", "Name", "Namespace", "App", "Owner", "Sharing", "Arity", "Arguments", "EvalBased", "Relations":
			return "environment_metadata"
		case "Validation":
			return "definitions"
		}
	case "ObjectIdentity":
		switch field {
		case "Kind", "Name", "Namespace", "App", "Owner":
			return "environment_metadata"
		}
	case "Report", "ResolutionReport":
		if field == "QueryScope" {
			return "environment_metadata"
		}
	case "SchemaTarget", "SchemaEntry":
		if field == "Schema" || field == "Catalog" || field == "Target" || field == "Metadata" {
			return "environment_metadata"
		}
	case "ObservationScope":
		switch field {
		case "Method", "Visibility", "Window", "TimePrecision", "AbsenceMeaning", "IndexSelection", "UnmatchedIndexes":
			return "environment_metadata"
		}
	case "ObservationCapture":
		switch field {
		case "Kind", "Datatype", "Coverage":
			return "environment_metadata"
		}
	case "ObservationIndex":
		if field == "CatalogDatatypes" || field == "RequiredDatatypes" {
			return "environment_metadata"
		}
	case "IndexEnumeration":
		if field == "Method" || field == "PeerScope" || field == "Coverage" {
			return "environment_metadata"
		}
	case "SchemaEvidence":
		switch field {
		case "Keyword", "Operator", "Branch", "ClassKey", "ClassUID", "DeclarationBasis", "Requirement":
			return "environment_metadata"
		case "Pointer":
			return "source_identity"
		}
	case "SchemaClass", "SchemaExtension":
		switch field {
		case "Key", "Name", "UID", "Version", "Category", "Profiles", "Extensions":
			return "environment_metadata"
		}
	case "FieldProjection":
		if field == "Target" || field == "Catalog" {
			return "environment_metadata"
		}
	case "Provenance":
		if field == "ObservedAt" || field == "SourceKind" {
			return "environment_metadata"
		}
	}
	return ""
}

// Shared JSON admission retains useful structured offsets, but its messages
// and unknown property names may contain caller secrets. Limit the path to its
// longest fixed-schema prefix and replace every message with a fixed summary.
func safeEvidenceError(err error) error {
	detail, ok := RequestErrorDetails(err)
	if !ok {
		return requestErrorAt("request_invalid", "", "invalid evidence request")
	}
	path := safeEvidencePath(detail.Path)
	if detail.ByteOffset != nil {
		return requestErrorOffset("request_invalid", path, "invalid evidence request", *detail.ByteOffset)
	}
	return requestErrorAt("request_invalid", path, "invalid evidence request")
}
func safeEvidencePath(path string) string {
	typ := reflect.TypeOf(EvidenceRequest{})
	out := ""
	for _, part := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		if typ == reflect.TypeOf([]byte{}) || typ.PkgPath() == "encoding/json" {
			break
		}
		switch typ.Kind() {
		case reflect.Struct:
			known := false
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				if f.IsExported() && jsonFieldName(f) == part {
					typ = f.Type
					known = true
					break
				}
			}
			if !known {
				return out
			}
		case reflect.Slice, reflect.Array:
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || strconv.Itoa(n) != part {
				return out
			}
			typ = typ.Elem()
		default:
			return out
		}
		out += "/" + part
	}
	return out
}
