package corpus

import (
	"encoding/json"

	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

// PreparedScan owns one canonical validation target for repeated corpus scans.
type PreparedScan struct {
	mode   string
	digest string
	field  *validation.PreparedFieldCatalog
	schema *validation.PreparedSchemaTarget
}

// Prepare validates a shared target before callers acquire any documents.
func Prepare(options ScanOptions) (*PreparedScan, error) {
	p := &PreparedScan{mode: "analysis"}
	if options.ValidationTarget == nil {
		return p, nil
	}
	raw, err := json.Marshal(options.ValidationTarget)
	if err != nil {
		return nil, inputError("validation target: %v", err)
	}
	target, err := DecodeValidationTarget(raw)
	if err != nil {
		return nil, err
	}
	p.digest, err = TargetDigest(target)
	if err != nil {
		return nil, err
	}
	p.mode = target.Kind
	if target.Kind == "field_list" {
		p.field, err = validation.PrepareFieldCatalog(*target.Catalog)
	} else {
		p.schema, err = validation.PrepareSchemaTarget(*target.SchemaTarget)
	}
	if err != nil {
		if validation.IsInputError(err) {
			return nil, inputError("validation target: %v", err)
		}
		return nil, err
	}
	return p, nil
}
