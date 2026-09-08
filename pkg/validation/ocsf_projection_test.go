package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func ocsfPrepared(t *testing.T, raw json.RawMessage, s OCSFSelection) preparedSchemaTarget {
	t.Helper()
	if s.Version == "" {
		s.Version = "1.6.0"
	}
	p, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &s})
	if e != nil {
		t.Fatal(e)
	}
	return p
}
func edgeOCSF(t *testing.T) json.RawMessage {
	t.Helper()
	b, e := os.ReadFile("../../testdata/schemas/ocsf/edge-cases.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func assertOCSF(t *testing.T, p preparedSchemaTarget, name, want string) fieldProjection {
	t.Helper()
	got := p.project(name)
	if got.Outcome != want {
		t.Fatalf("%s: got %+v, want %s", name, got, want)
	}
	return got
}
func hasOCSFReason(v []SchemaEvidence, reason string) bool {
	for _, e := range v {
		if e.Reason == reason {
			return true
		}
	}
	return false
}

func TestOCSFRealProjection(t *testing.T) {
	raw := readOCSFFixture(t, "base")
	p := ocsfPrepared(t, raw, OCSFSelection{Class: "authentication"})
	for n, w := range map[string]string{"time": "required", "actor": "optional", "cloud": "missing", "time_dt": "missing", "observables": "optional", "observables.name": "indeterminate", "unmapped.vendor_field": "permitted_unspecified", "actor.user.name": "optional"} {
		assertOCSF(t, p, n, w)
	}
	if !hasOCSFReason(p.project("observables.name").Evidence, "array_traversal") {
		t.Fatal("array evidence missing")
	}
	f := ocsfPrepared(t, raw, OCSFSelection{Class: "file_activity"})
	g := assertOCSF(t, f, "file.name", "required")
	if len(g.Evidence) < 2 {
		t.Fatal("ancestor evidence missing")
	}
	assertOCSF(t, f, "file.xattributes.vendor_field", "permitted_unspecified")
	p = ocsfPrepared(t, raw, OCSFSelection{Class: "authentication", Profiles: []string{"datetime", "cloud"}})
	assertOCSF(t, p, "cloud", "required")
	assertOCSF(t, p, "time_dt", "optional")
	uid := int64(3002)
	byID := ocsfPrepared(t, raw, OCSFSelection{ClassUID: &uid, Profiles: []string{"cloud", "datetime"}})
	if !reflect.DeepEqual(p.info(), byID.info()) {
		t.Fatal("key/uid normalization differs")
	}
	if p.universe().Complete {
		t.Fatal("generic and array descendants require partial universe")
	}
}

func TestOCSFRealCategory(t *testing.T) {
	raw := readOCSFFixture(t, "base")
	p := ocsfPrepared(t, raw, OCSFSelection{Category: "iam"})
	got := assertOCSF(t, p, "group", "conditional")
	if !reflect.DeepEqual(got.SupportingClasses, []SchemaClass{{Key: "authorize_session", UID: 3003}, {Key: "group_management", UID: 3006}}) {
		t.Fatalf("support: %+v", got)
	}
	if !reflect.DeepEqual(got.MissingClasses, []SchemaClass{{Key: "account_change", UID: 3001}, {Key: "authentication", UID: 3002}, {Key: "entity_management", UID: 3004}, {Key: "user_access", UID: 3005}}) {
		t.Fatalf("missing: %+v", got)
	}
	got = assertOCSF(t, p, "time", "required")
	if len(got.SupportingClasses) != 6 {
		t.Fatalf("time %+v", got)
	}
	uid := int64(3)
	q := ocsfPrepared(t, raw, OCSFSelection{CategoryUID: &uid})
	if !reflect.DeepEqual(p.info(), q.info()) {
		t.Fatal("category normalization")
	}
}

func TestOCSFRealWindows(t *testing.T) {
	raw := readOCSFFixture(t, "windows")
	p := ocsfPrepared(t, raw, OCSFSelection{Class: "win/registry_key_activity", Extensions: []string{"win"}})
	assertOCSF(t, p, "time", "required")
	uid := int64(201001)
	q := ocsfPrepared(t, raw, OCSFSelection{ClassUID: &uid, Extensions: []string{"win"}})
	if !reflect.DeepEqual(p.info(), q.info()) {
		t.Fatal("windows identity")
	}
	p = ocsfPrepared(t, raw, OCSFSelection{Class: "evidence_info", Extensions: []string{"win"}})
	if p.project("query_evidence.reg_key").Admission != analysis.SourceFieldAdmitted {
		t.Fatal("Windows patched base object missing")
	}
	b := ocsfPrepared(t, readOCSFFixture(t, "base"), OCSFSelection{Class: "evidence_info"})
	assertOCSF(t, b, "query_evidence.reg_key", "missing")
	if _, e := prepareOCSF(SchemaTarget{Kind: "ocsf", Catalog: raw, Selection: &OCSFSelection{Version: "1.6.0", Class: "evidence_info"}}); !IsInputError(e) {
		t.Fatalf("extension mismatch: %v", e)
	}
}

func TestOCSFEdgeProjection(t *testing.T) {
	raw := edgeOCSF(t)
	p := ocsfPrepared(t, raw, OCSFSelection{Class: "a"})
	for n, w := range map[string]string{"time": "required", "recommended": "optional", "unconditional_null": "optional", "unconditional_empty": "optional", "profiled": "missing", "empty.x": "missing", "subclass.x": "missing", "generic.x": "permitted_unspecified", "json.x": "permitted_unspecified", "array": "optional", "array.x": "indeterminate", "cycle.next.value": "optional", "single.x": "required", "multi.x": "optional", "unknown.x": "indeterminate", "unknown.y": "optional"} {
		assertOCSF(t, p, n, w)
	}
	if !hasOCSFReason(p.project("unknown.x").Evidence, "ocsf_constraint") {
		t.Fatal("unknown constraint evidence")
	}
	if p.universe().Complete {
		t.Fatal("open/cyclic universe claimed complete")
	}
	got := assertOCSF(t, p, "recommended", "optional")
	if len(got.Evidence) == 0 || got.Evidence[0].Requirement != "recommended" {
		t.Fatal("recommended evidence")
	}
	for _, x := range []struct {
		profiles []string
		want     string
	}{{[]string{"p"}, "indeterminate"}, {[]string{"p", "q"}, "required"}} {
		q := ocsfPrepared(t, raw, OCSFSelection{Class: "a", Profiles: x.profiles})
		assertOCSF(t, q, "profiled", x.want)
	}
	q := ocsfPrepared(t, raw, OCSFSelection{Class: "a", Profiles: []string{"child"}})
	assertOCSF(t, q, "time", "required")
	assertOCSF(t, q, "absent", "indeterminate")
	assertOCSF(t, q, "p_only", "indeterminate")
	q = ocsfPrepared(t, raw, OCSFSelection{Category: "test", Profiles: []string{"p"}})
	if len(q.info().Members) != 2 {
		t.Fatal("profile filtered category members")
	}
	assertOCSF(t, q, "p_only", "conditional")
	if got := p.project(strings.Repeat("cycle.next.", 150) + "value"); got.Outcome != "indeterminate" {
		t.Fatalf("budget %+v", got)
	}
}

func TestOCSFConcurrentOwnership(t *testing.T) {
	p := ocsfPrepared(t, edgeOCSF(t), OCSFSelection{Class: "a", Profiles: []string{"p", "q"}})
	before := p.info()
	mutated := p.info()
	mutated.Members[0].Key = "bad"
	mutated.Selection.Profiles[0] = "bad"
	mutated.Limitations[0] = "bad"
	if !reflect.DeepEqual(before, p.info()) {
		t.Fatal("info aliases target")
	}
	u := p.universe()
	if len(u.Fields) == 0 {
		t.Fatal("no concrete candidates")
	}
	u.Fields[0] = "bad"
	if reflect.DeepEqual(u.Fields, p.universe().Fields) {
		t.Fatal("universe aliases target")
	}
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 30; j++ {
				g := p.project("single.x")
				if g.Outcome != "required" {
					t.Errorf("projection %+v", g)
				}
				g.Evidence[0].Requirement = "bad"
				p.universe()
				p.info()
			}
		}()
	}
	wg.Wait()
}

func TestOCSFFiniteUniverseAndUnknownConstraint(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	// A closed member has one exact candidate, including through the resolver.
	p := ocsfPrepared(t, edgeOCSF(t), OCSFSelection{Class: "b"})
	u := p.universe()
	if !u.Complete || !reflect.DeepEqual(u.Fields, []string{"time"}) || u.Resolve("lost") != analysis.SourceFieldProhibited {
		t.Fatalf("closed universe %+v", u)
	}
	// Unknown scope-wide semantics must not make unlisted names prohibited by a
	// purportedly complete canonical universe before its resolver can run.
	m["classes"].(map[string]any)["b"].(map[string]any)["constraints"] = map[string]any{"future": map[string]any{"admit": "anything"}}
	raw, _ := json.Marshal(m)
	p = ocsfPrepared(t, raw, OCSFSelection{Class: "b"})
	if p.universe().Complete {
		t.Fatal("unknown constraint falsely exhausted universe")
	}
	assertOCSF(t, p, "lost", "indeterminate")
}

func TestOCSFConstraintAncestorEvidence(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	a := m["classes"].(map[string]any)["a"].(map[string]any)
	a["constraints"] = map[string]any{"at_least_one": []any{"cycle.value"}}
	raw, _ := json.Marshal(m)
	p := ocsfPrepared(t, raw, OCSFSelection{Class: "a"})
	assertOCSF(t, p, "cycle", "required")
	assertOCSF(t, p, "cycle.value", "required")
	assertOCSF(t, p, "cycle.next.value", "optional")
}

func TestOCSFLiteralDeclarationAndProfileLimitations(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	b := m["classes"].(map[string]any)["b"].(map[string]any)
	b["attributes"].(map[string]any)["vendor.name"] = map[string]any{"type": "string_t", "requirement": "optional"}
	raw, _ := json.Marshal(m)
	p := ocsfPrepared(t, raw, OCSFSelection{Class: "b"})
	assertOCSF(t, p, "vendor.name", "optional")
	if !containsOCSF(p.universe().Fields, "vendor.name") {
		t.Fatal("literal declaration lost")
	}
	p = ocsfPrepared(t, edgeOCSF(t), OCSFSelection{Class: "a", Profiles: []string{"p"}})
	if !containsOCSF(p.info().Limitations, "ocsf_profile_requirement") {
		t.Fatal("profile limitation missing")
	}
}

func TestOCSFConstraintPreservesDeclaredRequirement(t *testing.T) {
	p := ocsfPrepared(t, edgeOCSF(t), OCSFSelection{Class: "a"})
	g := assertOCSF(t, p, "single.x", "required")
	declared, constrained := false, false
	for _, e := range g.Evidence {
		if e.Pointer == "/objects/single/attributes/x" && e.Keyword == "requirement" && e.Requirement == "optional" {
			declared = true
		}
		if e.Operator == "at_least_one" && e.Requirement == "required" {
			constrained = true
		}
	}
	if !declared || !constrained {
		t.Fatalf("declaration/constraint provenance lost: %+v", g.Evidence)
	}
}

func TestOCSFCategoryProjectionBudget(t *testing.T) {
	var m map[string]any
	json.Unmarshal(edgeOCSF(t), &m)
	classes := m["classes"].(map[string]any)
	original := classes["a"].(map[string]any)
	for i := 0; i < 35; i++ {
		copy := map[string]any{}
		for k, v := range original {
			copy[k] = v
		}
		key := fmt.Sprintf("budget_%02d", i)
		copy["name"] = key
		copy["uid"] = int64(2000 + i)
		classes[key] = copy
	}
	raw, _ := json.Marshal(m)
	p := ocsfPrepared(t, raw, OCSFSelection{Category: "test"})
	got := p.project("cycle." + strings.Repeat("next.", 125) + "value")
	if got.Outcome != "indeterminate" || !hasOCSFReason(got.Evidence, "traversal_budget") {
		t.Fatalf("unbounded category projection: %s", got.Outcome)
	}
}
