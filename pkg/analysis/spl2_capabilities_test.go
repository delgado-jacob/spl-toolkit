package analysis

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSPL2CapabilitiesInventoryAndFormTruth(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/spl2/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct {
		Inventory []struct {
			Entry          string
			Splunkd        bool
			Classification string
		}
		Forms []struct{ ID, Disposition string }
		Holds []struct{ ID, Disposition string }
	}
	if err = json.Unmarshal(raw, &provenance); err != nil {
		t.Fatal(err)
	}
	manifest, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	commands := map[string]Capability{}
	disclosures := ""
	for _, c := range manifest.Commands {
		commands[c.Name] = c
		disclosures += strings.Join(c.Limitations, "\n") + "\n"
	}
	for _, f := range manifest.Functions {
		disclosures += strings.Join(f.Limitations, "\n") + "\n"
	}
	if len(commands) != 53 {
		t.Fatalf("inventory has %d commands, want 53", len(commands))
	}
	for _, entry := range provenance.Inventory {
		c, ok := commands[entry.Entry]
		if !ok {
			t.Errorf("missing %s", entry.Entry)
		}
		dedicated := strings.Contains(entry.Classification, "dedicated")
		if c.SyntaxSupported != dedicated || (!entry.Splunkd && c.SemanticSupported) {
			t.Errorf("false inventory support %+v: %+v", entry, c)
		}
	}
	for _, f := range provenance.Forms {
		if !strings.Contains(disclosures, "form "+f.ID+": "+f.Disposition+";") {
			t.Errorf("undisclosed form %s disposition %s", f.ID, f.Disposition)
		}
	}
	for _, h := range provenance.Holds {
		if !strings.Contains(disclosures, h.ID+": held;") {
			t.Errorf("undisclosed hold %s", h.ID)
		}
	}
	for _, tt := range []struct {
		name, text string
		status     Status
		semantic   bool
	}{
		{"select", `FROM main SELECT host`, Valid, true},
		{"append", `FROM main | append [FROM child | table host]`, Incomplete, false},
		{"branch", `FROM main | branch [where host=1 | into output]`, Incomplete, false},
		{"route", `FROM main | route output`, Invalid, false},
		{"stats", `FROM main | stats future=true count()`, Incomplete, true},
	} {
		r := spl2AnalyzeTest(t, tt.text)
		if r.Status != tt.status || commands[tt.name].SemanticSupported != tt.semantic {
			t.Fatalf("capability/canonical truth %s %+v %+v", tt.name, commands[tt.name], r)
		}
	}
}

func TestSPL2CapabilitiesSelectorsAndOwnedValues(t *testing.T) {
	defaultWire, _ := json.Marshal(Capabilities())
	selected, err := CapabilitiesFor(CapabilityOptions{Language: "spl", Profile: "splunkd", Version: "current"})
	wire, _ := json.Marshal(selected)
	if err != nil || string(wire) != string(defaultWire) {
		t.Fatal("default capabilities changed")
	}
	for _, options := range []CapabilityOptions{{Language: "SPL2"}, {Language: " spl2"}, {Language: "sql"}, {Language: "spl2", Profile: "cloud"}, {Language: "spl2", Version: "next"}, {Language: string([]byte{0xff})}, {Profile: string([]byte{0xff})}, {Version: string([]byte{0xff})}} {
		m, err := CapabilitiesFor(options)
		if err == nil || !reflect.DeepEqual(m, CapabilityManifest{}) {
			t.Fatalf("plausible default after selector failure: %+v %v", m, err)
		}
	}
	m, _ := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	before, _ := json.Marshal(m)
	for i := range m.Commands {
		m.Commands[i].Name = "mutated"
		for j := range m.Commands[i].Limitations {
			m.Commands[i].Limitations[j] = "mutated"
		}
	}
	for i := range m.Functions {
		m.Functions[i].Name = "mutated"
		for j := range m.Functions[i].Limitations {
			m.Functions[i].Limitations[j] = "mutated"
		}
	}
	after, _ := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	encoded, _ := json.Marshal(after)
	if string(before) != string(encoded) {
		t.Fatal("mutable capability values escaped")
	}
}

func TestSPL2CapabilitiesDeferredAndProfileOutcomes(t *testing.T) {
	manifest, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range manifest.Commands {
		if c.SyntaxSupported {
			continue
		}
		t.Run(c.Name, func(t *testing.T) {
			r := spl2AnalyzeTest(t, "FROM main | "+c.Name)
			profile := c.Name == "decrypt" || c.Name == "ocsf" || c.Name == "route"
			expected, code := Incomplete, CodeUnsupportedSemantics
			if profile {
				expected, code = Invalid, "SPL_PROFILE_MISMATCH"
			}
			if r.Status != expected || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete || !spl2HasCode(r, code) || len(r.Scopes) != 1 || len(r.References) != 1 {
				t.Fatalf("native/profile capability %s: %+v", c.Name, r)
			}
		})
	}
}
