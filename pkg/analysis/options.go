package analysis

import (
	"fmt"
	"unicode/utf8"
)

// normalizeSelectors is the shared contract for all selector-bearing APIs.
// Only available frontends are admitted; unknown selections never fall back.
func normalizeSelectors(options CapabilityOptions) (CapabilityOptions, error) {
	for _, option := range []struct {
		name     string
		value    *string
		standard string
	}{
		{"language", &options.Language, "spl"},
		{"profile", &options.Profile, "splunkd"},
		{"compatibility version", &options.Version, "current"},
	} {
		if !utf8.ValidString(*option.value) {
			return options, fmt.Errorf("%s is not valid UTF-8", option.name)
		}
		if *option.value == "" {
			*option.value = option.standard
		}
		if *option.value != option.standard && !(option.name == "language" && *option.value == "spl2") {
			return options, fmt.Errorf("unsupported %s %q", option.name, *option.value)
		}
	}
	return options, nil
}
