package analysis

import "github.com/delgado-jacob/spl-toolkit/internal/capabilityselector"

// normalizeSelectors is the shared contract for all selector-bearing APIs.
// Only available frontends are admitted; unknown selections never fall back.
func normalizeSelectors(options CapabilityOptions) (CapabilityOptions, error) {
	selection, err := capabilityselector.Normalize(options.Language, options.Profile, options.Version)
	options.Language, options.Profile, options.Version = selection.Language, selection.Profile, selection.Version
	return options, err
}
