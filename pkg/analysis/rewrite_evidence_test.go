package analysis

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func rewriteName(name string) RewriteIdentity { return RewriteIdentity{Name: &name} }
func rewriteTestSession(t *testing.T, language, query string, probes ...RewriteFactProbe) *RewriteSession {
	t.Helper()
	ordinary, err := Analyze(QueryDocument{Language: language, Text: query})
	if err != nil || !ordinary.Coverage.SyntaxComplete {
		t.Fatalf("malformed test corpus: %+v %v", ordinary, err)
	}
	s, err := PrepareRewrite(QueryDocument{Language: language, Text: query}, probes)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := json.Marshal(ordinary)
	b, _ := json.Marshal(s.Evidence().Analysis)
	if string(a) != string(b) {
		t.Fatalf("ordinary analysis changed: %s != %s", a, b)
	}
	return s
}
func rewriteFind(t *testing.T, s *RewriteSession, kind, name string, occurrence int) RewriteSite {
	t.Helper()
	for _, site := range s.Evidence().Sites {
		if site.Kind == kind && site.Identity.Name != nil && *site.Identity.Name == name {
			if occurrence == 0 {
				return site
			}
			occurrence--
		}
	}
	t.Fatalf("missing %s %q in %+v", kind, name, s.Evidence().Sites)
	return RewriteSite{}
}
func TestRewriteEvidenceSPL(t *testing.T) {
	s := rewriteTestSession(t, "spl", `search index=main source=app sourcetype=json status=200 | rename user AS actor | lookup people remote AS actor OUTPUT label AS display | table actor display`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
	input := rewriteFind(t, s, "field", "user", 0)
	if input.Eligibility != "eligible" || input.Role != "rename_input" || input.ReferenceID == "" || input.BindingID == "" || input.SourceEpochID == "" || input.Point.LineageIndex != 1 {
		t.Fatalf("source identity not retained: %+v", input)
	}
	alias := rewriteFind(t, s, "field", "actor", 1)
	if alias.Eligibility == "eligible" || alias.Role != "lookup_local" || alias.BindingID == input.BindingID {
		t.Fatalf("alias boundary lost: %+v", alias)
	}
	if got := input.Facts[0]; len(got.GuaranteedValues) != 1 || string(got.GuaranteedValues[0].Value) != "200" || !got.LiteralComplete || got.ReferenceState != "true" || len(got.ReferenceIDs) == 0 {
		t.Fatalf("missing fact: %+v", got)
	}
	for _, pair := range [][2]string{{"index", "main"}, {"source", "app"}, {"sourcetype", "json"}, {"lookup", "people"}} {
		site := rewriteFind(t, s, pair[0], pair[1], 0)
		if site.Eligibility != "eligible" {
			t.Fatalf("dependency not eligible: %+v", site)
		}
	}
	for _, tc := range []struct{ query, kind, name string }{{`datamodel Traffic All`, "data_model", "Traffic"}, {`datamodel Traffic All`, "dataset", "Traffic.All"}} {
		rewriteFind(t, rewriteTestSession(t, "spl", tc.query), tc.kind, tc.name, 0)
	}
}

func TestRewriteMilestone10OperandsRemainUnproved(t *testing.T) {
	tests := []struct {
		name, query, kind, operand, target string
		occurrence                         int
	}{
		{"tstats aggregate input", `tstats sum(bytes) FROM datamodel=Traffic.All`, "field", "bytes", "octets", 0},
		{"tstats where field", `tstats count FROM datamodel=Traffic.All WHERE status=200`, "field", "status", "state", 0},
		{"tstats group", `tstats count FROM datamodel=Traffic.All BY host`, "field", "host", "server", 0},
		{"fillnull", `search index=main | fillnull value="0" user`, "field", "user", "account", 0},
		{"rex", `search index=main | rex field=payload "(?<user>.+)"`, "field", "payload", "body", 0},
		{"spath", `search index=main | spath input=payload path="event.id" output=event_id`, "field", "payload", "body", 0},
		{"bin", `search index=main | bin span=5m _time AS bucket_time`, "field", "_time", "event_time", 0},
		{"bucket", `search index=main | bucket bins=10 duration`, "field", "duration", "elapsed", 0},
		{"regex", `search index=main | regex message="error"`, "field", "message", "body", 0},
		{"mvexpand", `search index=main | mvexpand values`, "field", "values", "items", 0},
		{"join key", `search index=main | join user [ search index=other ]`, "field", "user", "account", 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := rewriteTestSession(t, "spl", tc.query)
			site := rewriteFind(t, s, tc.kind, tc.operand, tc.occurrence)
			if site.Role != "" || site.Eligibility == "eligible" {
				t.Fatalf("milestone 10 operand escaped its rewrite boundary: %+v", site)
			}
			unproved := false
			for _, limitation := range site.Limitations {
				unproved = unproved || limitation.Code == "unproved_owner"
			}
			if !unproved {
				t.Fatalf("missing unproved-owner refusal: %+v", site)
			}
			if rewriteSupportedForm(t, "spl", site.Kind, site.Role) {
				t.Fatalf("held owner became an advertised rewrite form: %+v", site)
			}
			rendering, err := s.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName(tc.target)}})
			if err != nil {
				t.Fatal(err)
			}
			text := s.Evidence().Analysis.Document.Text
			candidateText := rewriteApply(text, rendering.Edits())
			if len(rendering.Edits()) != 0 || candidateText != text || len(rendering.Requirements()) == 0 {
				t.Fatalf("unproved operand rendered: text=%q candidate_text=%q edits=%+v requirements=%+v", text, candidateText, rendering.Edits(), rendering.Requirements())
			}
		})
	}

	t.Run("macro identity", func(t *testing.T) {
		const text = `search index=main | ` + "`normalize_user(user)`"
		s := rewriteTestSession(t, "spl", text)
		macroReferences := 0
		for _, reference := range s.Evidence().Analysis.References {
			if reference.Kind == "macro" && reference.NormalizedName == "normalize_user" && reference.Resolution == "exact" {
				macroReferences++
			}
		}
		if macroReferences != 1 {
			t.Fatalf("exact macro reference changed: %+v", s.Evidence().Analysis.References)
		}
		for _, site := range s.Evidence().Sites {
			if site.Kind == "macro" {
				t.Fatalf("macro identity became a rewrite site: %+v", site)
			}
		}
		if rewriteSupportedForm(t, "spl", "macro", "") {
			t.Fatal("macro identity became an advertised rewrite form")
		}
		candidateText := rewriteApply(text, nil)
		if candidateText != text {
			t.Fatalf("macro boundary changed text: text=%q candidate_text=%q", text, candidateText)
		}
	})

	t.Run("existing tstats data model", func(t *testing.T) {
		const text = `tstats count FROM datamodel=Traffic.All`
		s := rewriteTestSession(t, "spl", text)
		site := rewriteFind(t, s, "data_model", "Traffic", 0)
		if site.Eligibility != "eligible" || site.Role != "catalog_component" {
			t.Fatalf("existing data-model rewrite role changed: %+v", site)
		}
	})
}
func TestRewriteEvidenceSPL2(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | where status=200 | rename user AS actor | lookup people remote AS actor OUTPUT label AS display | table actor, display`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
	input := rewriteFind(t, s, "field", "user", 0)
	if input.Eligibility != "eligible" || input.Role != "rename_input" || input.Point.LineageIndex != 2 || len(input.Facts[0].GuaranteedValues) != 1 {
		t.Fatalf("bad evidence: %+v", input)
	}
	if rewriteFind(t, s, "field", "actor", 1).Eligibility == "eligible" {
		t.Fatal("explicit alias escaped boundary")
	}
	model := rewriteFind(t, rewriteTestSession(t, "spl2", `tstats aggregates=[count()] datamodel_name='Traffic.All'`), "data_model", "Traffic.All", 0)
	if model.Eligibility != "eligible" || model.Location.End.Offset-model.Location.Start.Offset != len("'Traffic.All'") {
		t.Fatalf("atomic tstats owner: %+v", model)
	}
}

func TestRewriteEvidenceSPL2StructuralNavigationIsIndeterminateAndRefused(t *testing.T) {
	query := `FROM main | eval x=actor.name`
	s := rewriteTestSession(t, "spl2", query)
	var site *rewriteSite
	for _, candidate := range s.sites {
		if reflect.DeepEqual(candidate.public.Identity.Path, []string{"actor", "name"}) {
			if site != nil {
				t.Fatalf("duplicate structural rewrite sites: %+v", s.Evidence().Sites)
			}
			site = candidate
		}
	}
	if site == nil {
		t.Fatalf("missing structural rewrite site: %+v", s.Evidence().Sites)
	}
	var reference *Reference
	for i := range s.result.References {
		if s.result.References[i].ID == site.public.ReferenceID {
			reference = &s.result.References[i]
		}
	}
	if reference == nil || reference.Binding != "indeterminate" || site.binding != "indeterminate" || site.public.Eligibility != "ineligible" || site.public.BindingID != "" {
		t.Fatalf("structural rewrite binding: reference=%+v site=%+v private_binding=%q", reference, site.public, site.binding)
	}
	if !reflect.DeepEqual(site.inputs, reference.OriginReferenceIDs) {
		t.Fatalf("structural rewrite origins: inputs=%v reference=%v", site.inputs, reference.OriginReferenceIDs)
	}
	foundOwnerLimitation := false
	for _, limitation := range site.public.Limitations {
		foundOwnerLimitation = foundOwnerLimitation || limitation.Code == "unproved_owner"
	}
	if !foundOwnerLimitation {
		t.Fatalf("structural rewrite site lost owner refusal: %+v", site.public)
	}
	rendering, err := s.Render([]RewriteReplacement{{SiteID: site.public.ID, Target: rewriteName("account.name")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(rendering.Edits()) != 0 || len(rendering.Requirements()) != 1 || len(rendering.Requirements()[0].Limitations) != 1 || rendering.Requirements()[0].Limitations[0].Code != "binding_not_source" {
		t.Fatalf("structural rewrite was not refused: edits=%+v requirements=%+v", rendering.Edits(), rendering.Requirements())
	}
}
func TestRewriteEvidenceLexicalControls(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		query := `search index=main | eval x="user" | lookup user remote AS user OUTPUT label AS display | table user`
		if language == "spl2" {
			query = `FROM main | eval x="user" | lookup user remote AS user OUTPUT label AS display | table user`
		}
		s := rewriteTestSession(t, language, query)
		count := 0
		for _, site := range s.Evidence().Sites {
			if site.Kind == "field" && site.Identity.Name != nil && *site.Identity.Name == "user" {
				count++
			}
		}
		if count != 2 {
			t.Fatalf("%s labels/literals became fields: %+v", language, s.Evidence().Sites)
		}
	}
}
func TestLocateRewriteBytes(t *testing.T) {
	got, err := LocateRewriteBytes("é\r\n😀x", []RewriteByteRange{{4, 8}, {9, 9}})
	want := []Location{{Position{4, 2, 1}, Position{8, 2, 2}}, {Position{9, 2, 3}, Position{9, 2, 3}}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("coordinates: %+v %v", got, err)
	}
	for _, tc := range []struct {
		text   string
		ranges []RewriteByteRange
	}{{"é", []RewriteByteRange{{1, 2}}}, {"é", []RewriteByteRange{{0, 1}}}, {"x", []RewriteByteRange{{1, 0}}}, {"x", []RewriteByteRange{{0, 2}}}, {"x", []RewriteByteRange{{-1, 0}}}, {string([]byte{0xff}), nil}, {"é", []RewriteByteRange{{0, 2}, {1, 1}}}} {
		if got, err := LocateRewriteBytes(tc.text, tc.ranges); err == nil || got != nil {
			t.Fatalf("accepted invalid %+v: %+v %v", tc, got, err)
		}
	}
	got, err = LocateRewriteBytes("", nil)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("empty: %+v %v", got, err)
	}
	got, err = LocateRewriteBytes("\t\r\nx", []RewriteByteRange{{0, 1}, {1, 2}, {2, 3}, {4, 4}})
	if err != nil || got[0].End.Column != 2 || got[1].End.Line != 2 || got[2].End.Line != 2 || got[3].Start.Column != 2 {
		t.Fatalf("tabs/CRLF: %+v %v", got, err)
	}
}
func TestRewriteRenderIdentity(t *testing.T) {
	for _, lang := range []string{"spl", "spl2"} {
		for _, target := range []string{"octets", "FROM", "actor.name", "é😀", "target | eval pwn=1", "a'b", "a\\b"} {
			t.Run(lang+"/"+target, func(t *testing.T) {
				query := `search index=main | where 'bytes'>0`
				if lang == "spl2" {
					query = `FROM main | where 'bytes'>0`
				}
				s := rewriteTestSession(t, lang, query)
				site := rewriteFind(t, s, "field", "bytes", 0)
				r, err := s.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName(target)}})
				if err != nil {
					t.Fatal(err)
				}
				if len(r.Edits()) != 1 {
					t.Fatalf("expected exact render: %+v %+v", r.Edits(), r.Requirements())
				}
				candidate := rewriteApply(query, r.Edits())
				c := rewriteTestSession(t, lang, candidate)
				got := rewriteFind(t, c, "field", target, 0)
				if got.Identity.Path != nil || len(c.Evidence().Analysis.Stages) != 2 {
					t.Fatalf("identity/injection: %+v", c.Evidence())
				}
				if proof := s.Verify(c, r); !proof.Proven {
					t.Fatalf("correspondence: %+v", proof)
				}
			})
		}
	}
}
func rewriteApply(text string, edits []RewriteTextEdit) string {
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		text = text[:e.Location.Start.Offset] + e.After + text[e.Location.End.Offset:]
	}
	return text
}
func TestRewriteRenderImplicit(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		query := `search index=main | stats sum(bytes) | table 'sum(bytes)'`
		if language == "spl2" {
			query = `FROM main | stats sum(bytes) | table 'sum(bytes)'`
		}
		s := rewriteTestSession(t, language, query)
		site := rewriteFind(t, s, "field", "bytes", 0)
		r, err := s.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName("octets")}})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Requirements()) == 0 {
			t.Fatalf("implicit consumer requirement missing: %+v", r)
		}
		req := r.Requirements()[0]
		if language == "spl2" {
			if len(req.Limitations) == 0 {
				t.Fatal("SPL2 label broadened")
			}
			continue
		}
		if len(req.RequiredChanges) != 1 {
			t.Fatalf("missing consumer: %+v", req)
		}
		all := append([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName("octets")}}, req.RequiredChanges...)
		r, err = s.Render(all)
		if err != nil {
			t.Fatal(err)
		}
		candidate := rewriteApply(query, r.Edits())
		if !strings.Contains(candidate, `sum(octets) | table 'sum(octets)'`) {
			t.Fatal(candidate)
		}
		if proof := s.Verify(rewriteTestSession(t, language, candidate), r); !proof.Proven {
			t.Fatalf("implicit proof: %+v", proof)
		}
	}
}

func TestRewriteEvidenceFactsAndBarriers(t *testing.T) {
	for _, lang := range []string{"spl", "spl2"} {
		start := `search index=main`
		if lang == "spl2" {
			start = `FROM main`
		}
		for _, tc := range []struct {
			body     string
			values   int
			complete bool
		}{{` | where (status=200 AND tag="ok") OR (status=200 AND tag="bad") | table bytes`, 1, true}, {` | where status=200 OR status=201 | table bytes`, 0, true}, {` | where NOT status=200 | table bytes`, 0, true}, {` | where status=200 | eval status=other | table bytes`, 0, false}, {` | where status=200 | fields - status | table bytes`, 0, false}} {
			s := rewriteTestSession(t, lang, start+tc.body, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
			site := rewriteFind(t, s, "field", "bytes", 0)
			fact := site.Facts[0]
			if len(fact.GuaranteedValues) != tc.values || fact.LiteralComplete != tc.complete {
				t.Fatalf("%s %s: %+v", lang, tc.body, fact)
			}
		}
	}
}
func TestRewriteEvidenceSQLPhases(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `SELECT user FROM main WHERE status=200 ORDER BY user`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
	site := rewriteFind(t, s, "field", "user", 0)
	if site.Point.Phase != "evaluate" || len(site.Facts[0].GuaranteedValues) != 1 {
		t.Fatalf("SQL scheduling: %+v", site)
	}
	reads := 0
	for _, site := range s.Evidence().Sites {
		if site.Identity.Name != nil && *site.Identity.Name == "user" {
			reads++
		}
	}
	if reads != 2 {
		t.Fatalf("duplicate SELECT preparation reference: %d", reads)
	}
	r, err := s.Render([]RewriteReplacement{{SiteID: site.ID, Target: rewriteName("account")}, {SiteID: rewriteFind(t, s, "field", "user", 1).ID, Target: rewriteName("account")}})
	if err != nil {
		t.Fatal(err)
	}
	if proof := s.Verify(rewriteTestSession(t, "spl2", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits())), r); !proof.Proven {
		t.Fatalf("SQL proof: %+v", proof)
	}
}
func TestRewriteSQLHiddenHavingPreservesVisiblePredicateEvidence(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `SELECT marker FROM main WHERE marker=1 GROUP BY marker HAVING marker=2 AND hidden=1 | lookup people marker OUTPUT label`, RewriteFactProbe{Kind: "field", Identity: rewriteName("marker")})
	evidence := s.Evidence()
	if evidence.Analysis.Status != Incomplete || len(evidence.Analysis.Diagnostics) != 1 || evidence.Analysis.Diagnostics[0].Code != CodeUnsupportedSemantics {
		t.Fatalf("hidden HAVING diagnostic changed: %+v", evidence.Analysis)
	}
	fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
	if fact.LiteralComplete || fact.ReferenceState != "true" || len(fact.GuaranteedValues) != 0 || !reflect.DeepEqual(fact.ReferenceIDs, []string{"ref-2", "ref-3", "ref-4"}) {
		t.Fatalf("hidden HAVING discarded visible predicate evidence: %+v", fact)
	}
}
func TestRewriteRenderLookupAndDependencies(t *testing.T) {
	for _, lang := range []string{"spl", "spl2"} {
		query := `search index=main | lookup people user OUTPUT label AS display | table user`
		if lang == "spl2" {
			query = `FROM main | lookup people user OUTPUT label AS display | table user`
		}
		s := rewriteTestSession(t, lang, query)
		first := rewriteFind(t, s, "field", "user", 0)
		second := rewriteFind(t, s, "field", "user", 1)
		r, err := s.Render([]RewriteReplacement{{first.ID, rewriteName("account")}, {second.ID, rewriteName("account")}, {rewriteFind(t, s, "lookup", "people", 0).ID, rewriteName("new people")}})
		if err != nil {
			t.Fatal(err)
		}
		candidate := rewriteApply(query, r.Edits())
		if !strings.Contains(candidate, "user AS account OUTPUT") {
			t.Fatalf("remote column changed: %s", candidate)
		}
		if proof := s.Verify(rewriteTestSession(t, lang, candidate), r); !proof.Proven {
			t.Fatalf("lookup proof: %+v candidate=%s", proof, candidate)
		}
	}
	for _, query := range []string{`datamodel Traffic All`, `from datamodel:Traffic.All`, `tstats count FROM datamodel=Traffic.All`} {
		s := rewriteTestSession(t, "spl", query)
		site := rewriteFind(t, s, "data_model", "Traffic", 0)
		r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("Network")}})
		if err != nil {
			t.Fatal(err)
		}
		candidate := rewriteApply(query, r.Edits())
		c := rewriteTestSession(t, "spl", candidate)
		if rewriteFind(t, c, "dataset", "Network.All", 0).Identity.Name == nil {
			t.Fatal("normalization missing")
		}
		if proof := s.Verify(c, r); !proof.Proven {
			t.Fatalf("composite proof: %+v query=%s", proof, query)
		}
	}
}
func TestRewriteEvidenceRefusalsAndCopies(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | where actor.name="bob" | table 'actor.name'`)
	for _, site := range s.Evidence().Sites {
		if site.Kind == "field" && site.Location.Start.Offset == 18 && site.Eligibility == "eligible" {
			t.Fatalf("unresolved navigation root eligible: %+v", site)
		}
	}
	s, _ = PrepareRewrite(QueryDocument{Language: "spl2", Text: `FROM main | table 'user*'`}, nil)
	if rewriteFind(t, s, "field", "user*", 0).Eligibility == "eligible" {
		t.Fatal("wildcard eligible")
	}
	s = rewriteTestSession(t, "spl", `search index=main | where user=1`)
	site := rewriteFind(t, s, "field", "user", 0)
	e := s.Evidence()
	*e.Sites[0].Identity.Name = "corrupted"
	e.Analysis.References[0].NormalizedName = "corrupted"
	if rewriteFind(t, s, "field", "user", 0).Identity.Name == nil {
		t.Fatal("evidence copy mutated session")
	}
	r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("account")}})
	if err != nil {
		t.Fatal(err)
	}
	edits := r.Edits()
	edits[0].After = "bad | eval injected=1"
	c := rewriteTestSession(t, "spl", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()))
	if !s.Verify(c, r).Proven {
		t.Fatal("copied edits corrupted provenance")
	}
	if other := rewriteTestSession(t, "spl", s.Evidence().Analysis.Document.Text); other.Verify(c, r).Proven {
		t.Fatal("cross-session rendering accepted")
	}
	path := RewriteIdentity{Path: []string{"actor", "name"}}
	r, err = s.Render([]RewriteReplacement{{site.ID, path}})
	if err != nil || len(r.Edits()) != 0 || len(r.Requirements()) != 1 || r.Requirements()[0].Limitations[0].Code != "target_not_renderable" {
		t.Fatalf("atom/path confusion: %+v %v", r, err)
	}
}
func TestRewriteScalarUnknownForms(t *testing.T) {
	for _, text := range []string{"1F", "1L", "1.2extra", "[1]", "{\"x\":1}", `"${x}"`} {
		if scalar, ok := rewriteScalar(text, false, "spl2"); ok {
			t.Fatalf("unproved scalar %q accepted %+v", text, scalar)
		}
	}
}

func TestRewriteEvidenceMetricsAndSourceEpoch(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `tstats aggregates=[count()] predicate=(index=main AND port=443) byfields=[host]`, RewriteFactProbe{Kind: "index", Identity: rewriteName("index")})
	index := rewriteFind(t, s, "index", "main", 0)
	if index.Role != "metric_value" {
		t.Fatalf("metric value role: %+v", index)
	}
	host := rewriteFind(t, s, "field", "host", 0)
	if host.Facts[0].ReferenceState != "true" || len(host.Facts[0].GuaranteedValues) != 1 {
		t.Fatalf("metric facts: %+v", host)
	}
	r, err := s.Render([]RewriteReplacement{{index.ID, rewriteName("archive")}})
	if err != nil {
		t.Fatal(err)
	}
	if p := s.Verify(rewriteTestSession(t, "spl2", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits())), r); !p.Proven {
		t.Fatalf("metric correspondence: %+v", p)
	}
	s = rewriteTestSession(t, "spl", `search index=main status=200 | append [ search index=other | table user ]`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
	user := rewriteFind(t, s, "field", "user", 0)
	if len(user.Facts[0].GuaranteedValues) != 0 || user.Facts[0].ReferenceState != "false" {
		t.Fatalf("independent scope inherited facts: %+v", user)
	}
	s = rewriteTestSession(t, "spl", `search index=main status=200 | inputlookup people | table user`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")}, RewriteFactProbe{Kind: "lookup", Identity: rewriteName("people")})
	user = rewriteFind(t, s, "field", "user", 0)
	if len(user.Facts[0].GuaranteedValues) != 0 || user.Facts[1].ReferenceState != "true" || rewriteFind(t, s, "lookup", "people", 0).SourceEpochID != user.SourceEpochID {
		t.Fatalf("source reset facts: %+v", user)
	}
}
func TestRewriteRenderUnknownAndCollision(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | eval account=1 | table user`)
	site := rewriteFind(t, s, "field", "user", 0)
	r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("account")}})
	if err != nil {
		t.Fatal(err)
	}
	if p := s.Verify(rewriteTestSession(t, "spl2", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits())), r); p.Proven {
		t.Fatal("source captured by derived alias")
	}
	s = rewriteTestSession(t, "spl", `search index=main | where user=1 | append [ search index=other ]`)
	site = rewriteFind(t, s, "field", "user", 0)
	r, err = s.Render([]RewriteReplacement{{site.ID, rewriteName("account")}})
	if err != nil {
		t.Fatal(err)
	}
	if p := s.Verify(rewriteTestSession(t, "spl", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits())), r); !p.Proven {
		t.Fatalf("unrelated complete original scope lost evidence: %+v", p)
	}
}
func TestRewriteCapabilities(t *testing.T) {
	for _, lang := range []string{"spl", "spl2"} {
		m, err := CapabilitiesFor(CapabilityOptions{Language: lang})
		if err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(m)
		var value map[string]json.RawMessage
		_ = json.Unmarshal(data, &value)
		if value["rewrite"] == nil {
			t.Fatalf("%s rewrite capability absent", lang)
		}
	}
}

func TestRewriteEvidenceUnknownOwners(t *testing.T) {
	for _, tc := range []struct{ lang, query string }{{"spl", `search index=main | where unknown(user)=1`}, {"spl2", `FROM main | where unknown(user)=1`}} {
		s := rewriteTestSession(t, tc.lang, tc.query)
		site := rewriteFind(t, s, "field", "user", 0)
		if site.Eligibility == "eligible" {
			t.Fatalf("unknown function owner eligible: %+v", site)
		}
	}
	s := rewriteTestSession(t, "spl2", `FROM main | eval '${field}'=1`)
	found := false
	for _, site := range s.Evidence().Sites {
		if site.Role == "dynamic_name" && site.ReferenceID == "" && site.Location.End.Offset > site.Location.Start.Offset && len(site.Limitations) > 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("dynamic owner lost located refusal: %+v", s.Evidence().Sites)
	}
}
func TestRewriteEvidenceExactNumberAndPartialFacts(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | where n=9007199254740993 OR n=9007199254740993.0 | table user`, RewriteFactProbe{Kind: "field", Identity: rewriteName("n")})
	if f := rewriteFind(t, s, "field", "user", 0).Facts[0]; len(f.GuaranteedValues) != 1 {
		t.Fatalf("exact numeric equivalence: %+v", f)
	}
	s = rewriteTestSession(t, "spl2", `FROM main | where n=1 AND unknown(other)=2 | table user`, RewriteFactProbe{Kind: "field", Identity: rewriteName("n")})
	if f := rewriteFind(t, s, "field", "user", 0).Facts[0]; len(f.GuaranteedValues) != 1 || f.LiteralComplete {
		t.Fatalf("partial fact erased guarantee: %+v", f)
	}
}
func TestRewriteRenderCompositeLinkedModel(t *testing.T) {
	s := rewriteTestSession(t, "spl", `datamodel Traffic All`)
	dataset := rewriteFind(t, s, "dataset", "Traffic.All", 0)
	r, err := s.Render([]RewriteReplacement{{dataset.ID, rewriteName("Network.Flows")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Requirements()) != 1 || len(r.Requirements()[0].RequiredChanges) != 1 || len(r.Requirements()[0].Limitations) != 0 {
		t.Fatalf("missing model requirement: %+v", r.Requirements())
	}
	changes := append([]RewriteReplacement{{dataset.ID, rewriteName("Network.Flows")}}, r.Requirements()[0].RequiredChanges...)
	r, err = s.Render(changes)
	if err != nil {
		t.Fatal(err)
	}
	if got := rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()); got != `datamodel Network Flows` {
		t.Fatal(got)
	}
	c := rewriteTestSession(t, "spl", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()))
	if proof := s.Verify(c, r); !proof.Proven {
		t.Fatalf("linked model proof: %+v", proof)
	}
}

func TestRewriteRenderRemovalMustPreserveFieldState(t *testing.T) {
	s := rewriteTestSession(t, "spl", `search index=main | where user=1 | fields - user`)
	removal := rewriteFind(t, s, "field", "user", 1)
	r, err := s.Render([]RewriteReplacement{{removal.ID, rewriteName("account")}})
	if err != nil {
		t.Fatal(err)
	}
	c := rewriteTestSession(t, "spl", rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()))
	if proof := s.Verify(c, r); proof.Proven {
		t.Fatal("changing only the removal silently retained the earlier source field")
	}
}
func TestRewriteRenderConditionalImplicitConsumer(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | stats sum(bytes) | lookup people id | table 'sum(bytes)'`)
	site := rewriteFind(t, s, "field", "bytes", 0)
	r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("octets")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Requirements()) == 0 || len(r.Requirements()[0].Limitations) == 0 {
		t.Fatal("uncertain implicit group escaped refusal")
	}
}
func TestRewriteRenderSearchFieldCannotBecomeDependency(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		s := rewriteTestSession(t, language, `search user=1`)
		site := rewriteFind(t, s, "field", "user", 0)
		r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("index")}})
		if err != nil {
			t.Fatal(err)
		}
		if len(r.Edits()) != 0 || len(r.Requirements()) == 0 {
			t.Fatalf("search field changed grammar category: %+v", r.Edits())
		}
	}
}

func TestRewriteEvidenceDerivedPredicateIsNotSourceFact(t *testing.T) {
	s := rewriteTestSession(t, "spl2", `FROM main | eval user=1 | where user=1 | table bytes`, RewriteFactProbe{Kind: "field", Identity: rewriteName("user")})
	fact := rewriteFind(t, s, "field", "bytes", 0).Facts[0]
	if fact.ReferenceState != "false" || len(fact.GuaranteedValues) != 0 {
		t.Fatalf("derived predicate became a source fact: %+v", fact)
	}
}

func TestRewriteEvidenceAggregationInvalidatesRemovedFacts(t *testing.T) {
	s := rewriteTestSession(t, "spl", `search index=main status=200 | stats count | lookup people count OUTPUT label`, RewriteFactProbe{Kind: "field", Identity: rewriteName("status")})
	fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
	if fact.ReferenceState != "false" || len(fact.GuaranteedValues) != 0 {
		t.Fatalf("aggregate retained removed source facts: %+v", fact)
	}
}

func TestRewriteEvidenceHeldSearchValues(t *testing.T) {
	for _, query := range []string{`search index="${name}"`, `search index=+42`, `search index="main*"`} {
		s, err := PrepareRewrite(QueryDocument{Language: "spl2", Text: query}, nil)
		if err != nil {
			t.Fatal(err)
		}
		for _, site := range s.Evidence().Sites {
			if site.Kind == "index" && site.Eligibility == "eligible" {
				t.Fatalf("held search value eligible in %q: %+v", query, site)
			}
		}
	}
}

func TestRewriteSearchPatternFacts(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, tc := range []struct{ predicate, kind, name string }{
			{`tag="x*"`, "field", "tag"}, {`tag=x*`, "field", "tag"},
			{`index="main*"`, "index", "index"}, {`index=main*`, "index", "index"},
		} {
			t.Run(language+"/"+tc.predicate, func(t *testing.T) {
				s := rewriteTestSession(t, language, "search "+tc.predicate+" | lookup people user OUTPUT label", RewriteFactProbe{Kind: tc.kind, Identity: rewriteName(tc.name)})
				fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
				if len(fact.GuaranteedValues) != 0 || fact.LiteralComplete {
					t.Errorf("search pattern became an exact literal: %+v", fact)
				}
				if tc.kind != "field" && (fact.ReferenceState == "true" || len(fact.ReferenceIDs) != 0) {
					t.Errorf("pattern dependency became an exact reference: %+v", fact)
				}
				if tc.kind == "field" && fact.ReferenceState != "true" {
					t.Errorf("pattern value erased its exact field read: %+v", fact)
				}
			})
		}
		for _, tc := range []struct{ predicate, kind, name, value string }{
			{`search tag="exact"`, "field", "tag", `"exact"`},
			{`search index="main"`, "index", "index", `"main"`},
			{`search index=main | where tag="x*"`, "field", "tag", `"x*"`},
			{`search index=main | where index="main*"`, "field", "index", `"main*"`},
		} {
			t.Run(language+"/exact/"+tc.predicate, func(t *testing.T) {
				s := rewriteTestSession(t, language, tc.predicate+" | lookup people user OUTPUT label", RewriteFactProbe{Kind: tc.kind, Identity: rewriteName(tc.name)})
				fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
				if !fact.LiteralComplete || len(fact.GuaranteedValues) != 1 || string(fact.GuaranteedValues[0].Value) != tc.value || fact.ReferenceState != "true" {
					t.Fatalf("exact typed literal lost: %+v", fact)
				}
			})
		}
	}
}

func TestRewriteRenderRejectsOriginalInvalidUTF8(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		s := rewriteTestSession(t, language, `search index=main | where user=1`)
		site := rewriteFind(t, s, "field", "user", 0)
		for _, invalid := range []string{string([]byte{0xff}), string([]byte{'a', 0xc3}), string([]byte{0xed, 0xa0, 0x80})} {
			for _, target := range []RewriteIdentity{rewriteName(invalid), {Path: []string{"valid", invalid}}} {
				if r, err := s.Render([]RewriteReplacement{{site.ID, target}}); err == nil || r != nil {
					t.Errorf("%s accepted malformed target bytes %x: %+v %v", language, []byte(invalid), target, err)
				}
			}
		}
		target := "é😀"
		changes := []RewriteReplacement{{site.ID, rewriteName(target)}}
		r, err := s.Render(changes)
		if err != nil {
			t.Fatal(err)
		}
		*changes[0].Target.Name = "mutated"
		if len(r.Effects()) != 1 || r.Effects()[0].After.Name == nil || *r.Effects()[0].After.Name != target {
			t.Fatalf("valid UTF-8 target not privately preserved: %+v", r.Effects())
		}
		c := rewriteTestSession(t, language, rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()))
		if proof := s.Verify(c, r); !proof.Proven {
			t.Fatalf("valid identity did not survive copy: %+v", proof)
		}
	}
}

func TestRewriteNullInspectionOwnerProof(t *testing.T) {
	for _, tc := range []struct {
		language, expression, role string
		eligible                   bool
	}{
		{"spl", `unknown(isnull(user))=1`, "", false},
		{"spl2", `unknown(isnull(user))=1`, "", false},
		{"spl2", `abs(value:isnull(user))=1`, "", false},
		{"spl", `abs(value:isnull(user))=1`, "", false},
		{"spl", `searchmatch(isnull(user))=1`, "", false},
		{"spl", `isnull(user)`, "null_test", true},
		{"spl2", `isnull(user)`, "null_test", true},
		{"spl", `isnotnull(user)`, "null_test", true},
		{"spl2", `isnotnull(user)`, "null_test", true},
		{"spl", `ISNULL('user')`, "null_test", true},
		{"spl", `abs(isnull(user))=1`, "null_test", true},
		{"spl2", `abs(isnull(user))=1`, "null_test", true},
		{"spl", `isnull(isnotnull(user))`, "null_test", true},
		{"spl", `abs(user)=1`, "expression_atom", true},
		{"spl", `isnull(abs(user))`, "expression_atom", true},
		{"spl2", `isnull(abs(user))`, "expression_atom", true},
		{"spl", `isnull(user+1)`, "expression_atom", true},
		{"spl2", `isnull(user+1)`, "expression_atom", true},
		{"spl", `isnull((user))`, "expression_atom", true},
	} {
		t.Run(tc.language+"/"+tc.expression, func(t *testing.T) {
			s := rewriteTestSession(t, tc.language, "search index=main | where "+tc.expression)
			site := rewriteFind(t, s, "field", "user", 0)
			if site.Role != tc.role {
				t.Errorf("null inspection role = %q, want %q", site.Role, tc.role)
			}
			if (site.Eligibility == "eligible") != tc.eligible {
				t.Errorf("null inspection changed owner proof: %+v", site)
			}
			r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("account")}})
			if err != nil {
				t.Fatal(err)
			}
			if !tc.eligible {
				if len(r.Edits()) != 0 || len(r.Requirements()) == 0 || len(r.Requirements()[0].Limitations) == 0 {
					t.Fatalf("unproved null owner rendered: %+v %+v", r.Edits(), r.Requirements())
				}
				return
			}
			c := rewriteTestSession(t, tc.language, rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits()))
			if len(r.Edits()) != 1 || !s.Verify(c, r).Proven {
				t.Fatalf("proved null owner refused: %+v %+v", r.Requirements(), s.Verify(c, r))
			}
		})
	}
}

func TestRewriteSPLNullInspectionBoundaries(t *testing.T) {
	for _, tc := range []struct {
		query, name, role, binding string
		syntax                     bool
	}{
		{`search user=x | where isnull(user)`, "user", "null_test", "source", true},
		{`search user=x | eval checked=isnotnull(user)`, "user", "null_test", "source", true},
		{`search user=x | eval user=1 | where isnull(user)`, "user", "null_test", "derived", true},
		{`search user=x | table other | where isnull(user)`, "user", "null_test", "unavailable", true},
		{`search index=main | mystery | where isnull(user)`, "user", "null_test", "indeterminate", true},
		{`search user=x | where isnull(user, 1)`, "user", "expression_atom", "source", true},
		{`search user=x | where isnotnull(user, 1)`, "user", "expression_atom", "source", true},
		{`search user=x | stats isnull(user)`, "user", "expression_atom", "source", true},
		{`search user=x | where isnull(user . other)`, "user", "expression_atom", "source", true},
		{`search user=x | where isnull(user`, "user", "expression_atom", "source", false},
		{`search user=x | where isnull(user @)`, "user", "expression_atom", "source", false},
		{`search user=x | where isnull(user[0])`, "user", "expression_atom", "source", false},
	} {
		t.Run(tc.query, func(t *testing.T) {
			document := QueryDocument{Language: "spl", Text: tc.query}
			ordinary, err := Analyze(document)
			if err != nil {
				t.Fatal(err)
			}
			s, err := PrepareRewrite(document, nil)
			if err != nil {
				t.Fatal(err)
			}
			evidence := s.Evidence()
			a, _ := json.Marshal(ordinary)
			b, _ := json.Marshal(evidence.Analysis)
			if string(a) != string(b) || evidence.Analysis.Coverage.SyntaxComplete != tc.syntax {
				t.Fatalf("ordinary analysis or syntax changed: %+v", evidence.Analysis)
			}
			offset := strings.LastIndex(tc.query, tc.name)
			for _, site := range evidence.Sites {
				if site.Kind != "field" || site.Location.Start.Offset != offset {
					continue
				}
				if site.Role != tc.role || site.Location.End.Offset != offset+len(tc.name) || site.OwnerLocation != site.Location {
					t.Errorf("wrong role or operand location: %+v", site)
				}
				var reference Reference
				for _, ref := range evidence.Analysis.References {
					if ref.ID == site.ReferenceID {
						reference = ref
					}
				}
				wantReferenceRole := "read"
				if tc.role == "null_test" {
					wantReferenceRole = "null_test"
				}
				if reference.Binding != tc.binding || reference.Role != wantReferenceRole || reference.Location != site.Location {
					t.Errorf("ordinary null-inspection reference changed: %+v", reference)
				}
				if tc.binding != "source" || !tc.syntax {
					r, err := s.Render([]RewriteReplacement{{site.ID, rewriteName("account")}})
					if err != nil || len(r.Edits()) != 0 || len(r.Requirements()) == 0 {
						t.Fatalf("unproved or malformed null inspection rendered: %+v %v", r, err)
					}
				}
				return
			}
			// Parser recovery may discard the damaged operand entirely.
			if tc.syntax {
				t.Fatal("missing null-inspection operand")
			}
		})
	}
}

func TestRewriteModelOnlyQuotedDatasetEffect(t *testing.T) {
	for _, component := range []string{`'All'`, `"All"`} {
		t.Run(component, func(t *testing.T) {
			s := rewriteTestSession(t, "spl", "datamodel Traffic "+component)
			model := rewriteFind(t, s, "data_model", "Traffic", 0)
			dataset := rewriteFind(t, s, "dataset", "Traffic.All", 0)
			r, err := s.Render([]RewriteReplacement{{model.ID, rewriteName("Network")}})
			if err != nil {
				t.Fatal(err)
			}
			candidate := rewriteApply(s.Evidence().Analysis.Document.Text, r.Edits())
			if candidate != "datamodel Network "+component {
				t.Fatalf("component spelling changed: %s", candidate)
			}
			found := false
			for _, effect := range r.Effects() {
				if effect.ReferenceID == dataset.ReferenceID {
					found = true
					if !reflect.DeepEqual(effect.After, rewriteName("Network.All")) {
						t.Errorf("quoted token became logical identity: %+v", effect)
					}
				}
			}
			if !found {
				t.Error("missing enclosing dataset effect")
			}
			if proof := s.Verify(rewriteTestSession(t, "spl", candidate), r); !proof.Proven {
				t.Errorf("model-only quoted component correspondence: %+v", proof)
			}
		})
	}
}

func TestRewriteAggregationFirstReadEpoch(t *testing.T) {
	s := rewriteTestSession(t, "spl", `stats count BY host | table host`, RewriteFactProbe{Kind: "field", Identity: rewriteName("host")})
	group, consumer := rewriteFind(t, s, "field", "host", 0), rewriteFind(t, s, "field", "host", 1)
	if group.SourceEpochID == "" || group.SourceEpochID != consumer.SourceEpochID || group.BindingID == "" || group.BindingID != consumer.BindingID {
		t.Errorf("aggregation reset a retained source binding: group=%+v consumer=%+v", group, consumer)
	}
	fact := consumer.Facts[0]
	if fact.ReferenceState != "true" || !slices.Contains(fact.ReferenceIDs, group.ReferenceID) {
		t.Errorf("aggregation discarded grouping provenance: %+v", fact)
	}
}

func TestRewriteLexicalMappingControls(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		count := "count"
		if language == "spl2" {
			count = "count()"
		}
		for _, tc := range []struct{ name, query, want string }{
			{"comment", `search index=main | where user=1 /* user */ | table user`, `search index=main | where account=1 /* user */ | table account`},
			{"option", `search index=main | where user=1 AND user=1 | stats delim="user" ` + count, `search index=main | where account=1 AND account=1 | stats delim="user" ` + count},
			{"remote", `search index=main | where user=1 | lookup people user AS local OUTPUT label AS display | table user`, `search index=main | where account=1 | lookup people user AS local OUTPUT label AS display | table account`},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				s := rewriteTestSession(t, language, tc.query)
				changes := []RewriteReplacement{}
				for _, site := range s.Evidence().Sites {
					if site.Kind == "field" && site.Identity.Name != nil && *site.Identity.Name == "user" {
						if site.Eligibility != "eligible" {
							t.Fatalf("source mapping control is not eligible: %+v", site)
						}
						changes = append(changes, RewriteReplacement{site.ID, rewriteName("account")})
					}
				}
				if len(changes) != 2 {
					t.Fatalf("lexical text acquired field identity: %+v", s.Evidence().Sites)
				}
				r, err := s.Render(changes)
				if err != nil {
					t.Fatal(err)
				}
				if got := rewriteApply(tc.query, r.Edits()); got != tc.want {
					t.Fatalf("lexical control changed: %s; %+v", got, r.Requirements())
				}
				if proof := s.Verify(rewriteTestSession(t, language, tc.want), r); !proof.Proven {
					t.Fatalf("lexical-control correspondence: %+v", proof)
				}
			})
		}
	}
}

func TestRewriteSPL2PatternSlotControls(t *testing.T) {
	for _, tc := range []struct {
		query, kind, name string
		exact             bool
	}{
		{`search index="main\u002a" | lookup people user OUTPUT label`, "index", "main*", false},
		{`tstats aggregates=[count()] predicate=(index="main*") byfields=[host] | lookup people host OUTPUT label`, "index", "main*", false},
		{`tstats aggregates=[count()] predicate=(port="x*") byfields=[host] | lookup people host OUTPUT label`, "field", "port", true},
	} {
		t.Run(tc.query, func(t *testing.T) {
			s := rewriteTestSession(t, "spl2", tc.query, RewriteFactProbe{Kind: tc.kind, Identity: rewriteName(tc.name)})
			fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
			if tc.exact {
				// Probe at the predicate before aggregation projects port away.
				fact = rewriteFind(t, s, "field", "port", 0).Facts[0]
				if !fact.LiteralComplete || len(fact.GuaranteedValues) != 1 || string(fact.GuaranteedValues[0].Value) != `"x*"` {
					t.Fatalf("metric expression string lost exact semantics: %+v", fact)
				}
			} else if fact.LiteralComplete || len(fact.GuaranteedValues) != 0 || fact.ReferenceState == "true" || len(fact.ReferenceIDs) != 0 {
				t.Fatalf("pattern slot acquired exact proof: %+v", fact)
			}
		})
	}
}

func TestRewriteProducerCompleteAbsence(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, tc := range []struct{ predicate, kind, name string }{
			{`(sourcetype=a OR sourcetype=b) src=x`, "sourcetype", "sourcetype"},
			{`NOT sourcetype=a src=x`, "sourcetype", "sourcetype"},
			{`(EventCode=1 OR EventCode=2) src=x`, "field", "EventCode"},
			{`NOT EventCode=1 src=x`, "field", "EventCode"},
			{`src=x`, "field", "EventCode"},
		} {
			t.Run(language+"/"+tc.predicate, func(t *testing.T) {
				s := rewriteTestSession(t, language, "search "+tc.predicate+" | table src", RewriteFactProbe{Kind: tc.kind, Identity: rewriteName(tc.name)})
				for occurrence := 0; occurrence < 2; occurrence++ {
					fact := rewriteFind(t, s, "field", "src", occurrence).Facts[0]
					if !fact.LiteralComplete || len(fact.GuaranteedValues) != 0 {
						t.Errorf("supported absence is not conclusive at occurrence %d: %+v", occurrence, fact)
					}
				}
			})
		}
		t.Run(language+"/no-backfill", func(t *testing.T) {
			s := rewriteTestSession(t, language, `search src=x | where EventCode=1 | table src`, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
			early, late := rewriteFind(t, s, "field", "src", 0).Facts[0], rewriteFind(t, s, "field", "src", 1).Facts[0]
			if !early.LiteralComplete || len(early.GuaranteedValues) != 0 || early.ReferenceState != "false" || len(early.ReferenceIDs) != 0 || len(early.Locations) != 0 {
				t.Errorf("earlier complete absence lost or backfilled: %+v", early)
			}
			if !late.LiteralComplete || len(late.GuaranteedValues) != 1 || late.GuaranteedValues[0].Kind != "number" || string(late.GuaranteedValues[0].Value) != "1" || late.ReferenceState != "true" {
				t.Errorf("later positive guarantee lost: %+v", late)
			}
		})
	}
}

func TestRewriteProducerUnknownCompleteness(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		for _, query := range []string{
			`search EventCode="x*" src=x | table src`,
			`search tag="x*" src=x | table src`,
			`search NOT EventCode="x*" src=x | table src`,
			`search src=x | where EventCode=unknown(other) | table src`,
			`search src=x | where NOT unknown(EventCode)=1 | table src`,
			`search src=x | where -(EventCode=1) | table src`,
			`search src=x | eval EventCode=other | table src`,
			`search src=x | fields - EventCode | table src`,
		} {
			t.Run(language+"/"+query, func(t *testing.T) {
				s := rewriteTestSession(t, language, query, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
				fact := rewriteFind(t, s, "field", "src", 1).Facts[0]
				if fact.LiteralComplete || len(fact.GuaranteedValues) != 0 {
					t.Fatalf("unproved flow became conclusive: %+v", fact)
				}
			})
		}
		t.Run(language+"/partial-positive", func(t *testing.T) {
			s := rewriteTestSession(t, language, `search src=x | where EventCode=1 AND unknown(other)=2 | table src`, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
			fact := rewriteFind(t, s, "field", "src", 1).Facts[0]
			if fact.LiteralComplete || len(fact.GuaranteedValues) != 1 || string(fact.GuaranteedValues[0].Value) != "1" {
				t.Fatalf("incomplete predicate discarded an independent guarantee: %+v", fact)
			}
		})
	}
	for _, query := range []string{`search src=x | where EventCode=1F | table src`, `search src=x | where EventCode=@"one" | table src`, `search src=x | where EventCode="${other}" | table src`, `search src=x | where [EventCode=1] | table src`} {
		s := rewriteTestSession(t, "spl2", query, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
		if fact := rewriteFind(t, s, "field", "src", 1).Facts[0]; fact.LiteralComplete || len(fact.GuaranteedValues) != 0 {
			t.Fatalf("unproved scalar became conclusive in %q: %+v", query, fact)
		}
	}
	s := rewriteTestSession(t, "spl2", `FROM main | table src`, RewriteFactProbe{Kind: "field", Identity: RewriteIdentity{Path: []string{"event", "code"}}})
	if fact := rewriteFind(t, s, "field", "src", 0).Facts[0]; fact.LiteralComplete || fact.ReferenceState != "unknown" {
		t.Fatalf("unproved path binding became conclusive: %+v", fact)
	}
}

func TestRewriteProducerScopeAndSourceBoundaries(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		s, err := PrepareRewrite(QueryDocument{Language: language, Text: `search src=x | where EventCode=1 | mystery | table src`}, []RewriteFactProbe{{Kind: "field", Identity: rewriteName("EventCode")}})
		if err != nil {
			t.Fatal(err)
		}
		if fact := rewriteFind(t, s, "field", "src", 1).Facts[0]; fact.LiteralComplete || len(fact.GuaranteedValues) != 0 {
			t.Fatalf("unknown command effects acquired complete facts: %+v", fact)
		}
		for _, query := range []string{
			`search EventCode=1 src=x | inputlookup people | table child`,
			`search EventCode=1 src=x | append [search child=x | table child]`,
		} {
			if language == "spl2" {
				query = strings.Replace(query, "inputlookup people", "FROM people", 1)
			}
			s := rewriteTestSession(t, language, query, RewriteFactProbe{Kind: "field", Identity: rewriteName("EventCode")})
			fact := rewriteFind(t, s, "field", "child", 0).Facts[0]
			if len(fact.GuaranteedValues) != 0 || fact.ReferenceState != "false" || len(fact.ReferenceIDs) != 0 {
				t.Fatalf("independent source inherited a guarantee: %+v", fact)
			}
		}
	}
}

func TestRewriteProducerMetricBooleanRole(t *testing.T) {
	for _, value := range []string{"true", "false", `"true"`, `"false"`} {
		kind := "boolean"
		if strings.HasPrefix(value, `"`) {
			kind = "string"
		}
		for _, query := range []string{`tstats aggregates=[count()] predicate=(flag=` + value + `) byfields=[host]`, `FROM main | where flag=` + value} {
			t.Run(query, func(t *testing.T) {
				s := rewriteTestSession(t, "spl2", query, RewriteFactProbe{Kind: "field", Identity: rewriteName("flag")})
				fact := rewriteFind(t, s, "field", "flag", 0).Facts[0]
				if !fact.LiteralComplete || len(fact.GuaranteedValues) != 1 || fact.GuaranteedValues[0].Kind != kind || string(fact.GuaranteedValues[0].Value) != value {
					t.Fatalf("expression scalar changed grammar category: %+v", fact)
				}
			})
		}
	}
	for _, language := range []string{"spl", "spl2"} {
		s := rewriteTestSession(t, language, `search flag=true | table src`, RewriteFactProbe{Kind: "field", Identity: rewriteName("flag")})
		fact := rewriteFind(t, s, "field", "src", 0).Facts[0]
		if len(fact.GuaranteedValues) != 1 || fact.GuaranteedValues[0].Kind != "string" || string(fact.GuaranteedValues[0].Value) != `"true"` {
			t.Fatalf("search word became an expression Boolean: %+v", fact)
		}
	}
}

func TestRewriteProducerUnobservedPrefix(t *testing.T) {
	for _, language := range []string{"spl", "spl2"} {
		count := "count"
		if language == "spl2" {
			count = "count()"
		}
		s, err := PrepareRewrite(QueryDocument{Language: language, Text: "mystery | stats " + count + " | lookup people count OUTPUT label"}, []RewriteFactProbe{{Kind: "sourcetype", Identity: rewriteName("sourcetype")}})
		if err != nil {
			t.Fatal(err)
		}
		fact := rewriteFind(t, s, "lookup", "people", 0).Facts[0]
		if fact.LiteralComplete || len(fact.GuaranteedValues) != 0 {
			t.Errorf("%s unknown prefix lost before first evidence read: %+v", language, fact)
		}
	}
}
