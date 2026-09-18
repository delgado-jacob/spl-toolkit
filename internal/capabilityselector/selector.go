package capabilityselector

import (
	"fmt"
	"unicode/utf8"
)

type Selection struct {
	Language string
	Profile  string
	Version  string
}

func Normalize(language, profile, version string) (Selection, error) {
	selection := Selection{Language: language, Profile: profile, Version: version}
	for _, option := range []struct {
		name     string
		value    *string
		standard string
	}{
		{"language", &selection.Language, "spl"},
		{"profile", &selection.Profile, "splunkd"},
		{"compatibility version", &selection.Version, "current"},
	} {
		if !utf8.ValidString(*option.value) {
			return selection, fmt.Errorf("%s is not valid UTF-8", option.name)
		}
		if *option.value == "" {
			*option.value = option.standard
		}
		if *option.value != option.standard && !(option.name == "language" && *option.value == "spl2") {
			return selection, fmt.Errorf("unsupported %s %q", option.name, *option.value)
		}
	}
	return selection, nil
}
