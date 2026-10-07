package resolution

import (
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/environment"
)

// Prepare detaches and validates artifacts once, without retaining query state.
func Prepare(snapshot environment.Snapshot, schemas *environment.SchemaBundle) (*Prepared, error) {
	prepared, err := compatibility.Prepare(snapshot, schemas)
	if err != nil {
		return nil, err
	}
	return &Prepared{compatibility: prepared, identity: prepared.ArtifactIdentity()}, nil
}
