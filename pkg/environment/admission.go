package environment

// DecodeSnapshot preserves raw artifact admission distinctions without preparing
// or compiling evidence. A nonnil report contains the existing invalid diagnostic.
func DecodeSnapshot(raw []byte) (Snapshot, *Report) {
	var value Snapshot
	if err := decodeStrictJSON(raw, &value); err != nil {
		_, report, _ := invalidSnapshot("", err)
		return Snapshot{}, report
	}
	if err := validateSnapshotRawShape(raw); err != nil {
		_, report, _ := invalidSnapshot("", err)
		return Snapshot{}, report
	}
	return value, nil
}

// DecodeSchemaBundle admits raw schema bundle shape without compiling targets.
// The returned value must still be passed to PrepareSchemaBundle for semantics.
func DecodeSchemaBundle(raw []byte) (SchemaBundle, *Report) {
	var value SchemaBundle
	if err := decodeStrictJSON(raw, &value); err != nil {
		_, report, _ := invalidSchemaBundle("", err)
		return SchemaBundle{}, report
	}
	if err := validateSchemaBundleRawShape(raw); err != nil {
		_, report, _ := invalidSchemaBundle("", err)
		return SchemaBundle{}, report
	}
	return value, nil
}
