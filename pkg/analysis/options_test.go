package analysis

import (
	"reflect"
	"testing"
)

func TestSelectorsDefaultsAndInvalidProfile(t *testing.T) {
	got, err := normalizeSelectors(CapabilityOptions{})
	if err != nil || got.Language != "spl" || got.Profile != "splunkd" || got.Version != "current" {
		t.Fatalf("unexpected defaults: %#v %v", got, err)
	}
	manifest, err := CapabilitiesFor(CapabilityOptions{})
	if err != nil || !reflect.DeepEqual(manifest, Capabilities()) {
		t.Fatalf("default manifest changed: %#v %v", manifest, err)
	}
	for _, options := range []CapabilityOptions{{Language: "spl2"}, {Language: "SPL"}, {Profile: "edge"}, {Version: "next"}, {Language: "\xff"}, {Profile: "\xff"}, {Version: "\xff"}} {
		if _, err := normalizeSelectors(options); err == nil {
			t.Errorf("accepted %#v", options)
		}
		if _, err := CapabilitiesFor(options); err == nil {
			t.Errorf("manifest advertised unavailable %#v", options)
		}
		if _, err := Analyze(QueryDocument{Text: "search *", Language: options.Language, Profile: options.Profile, Version: options.Version}); err == nil {
			t.Errorf("Analyze accepted %#v", options)
		}
	}
	doc := QueryDocument{Text: "search café=*\r\n", SourceID: " exact\r\n "}
	normalized, err := normalizeDocument(doc)
	if err != nil || normalized.Text != doc.Text || normalized.SourceID != doc.SourceID {
		t.Fatalf("changed caller data %#v %v", normalized, err)
	}
	for _, doc := range []QueryDocument{{Text: "\xff"}, {SourceID: "\xff"}} {
		if _, err := normalizeDocument(doc); err == nil {
			t.Error("invalid UTF-8 accepted")
		}
	}
}
