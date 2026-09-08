package validation

// DecodeSchemaRequest enforces the strict document/target wrapper. Target
// semantics are checked during preparation by ValidateSchema.
func DecodeSchemaRequest(data []byte) (SchemaRequest, error) {
	var request SchemaRequest
	fields, err := object(data, []string{"document", "target"}, "document", "target")
	if err != nil {
		return request, err
	}
	document, err := decodeDocument(fields["document"])
	if err != nil {
		return request, err
	}
	target, err := DecodeSchemaTarget(fields["target"])
	if err != nil {
		return request, err
	}
	return SchemaRequest{Document: document, Target: target}, nil
}
func DecodeSchemaBatchRequest(data []byte) (SchemaBatchRequest, error) {
	var request SchemaBatchRequest
	fields, err := object(data, []string{"documents", "target"}, "documents", "target")
	if err != nil {
		return request, err
	}
	documents, err := DecodeDocuments(fields["documents"])
	if err != nil {
		return request, err
	}
	target, err := DecodeSchemaTarget(fields["target"])
	if err != nil {
		return request, err
	}
	return SchemaBatchRequest{Documents: documents, Target: target}, nil
}
