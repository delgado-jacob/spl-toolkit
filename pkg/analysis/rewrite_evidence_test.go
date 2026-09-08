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
		}{{` | where (status=200 AND tag="ok") OR (status=200 AND tag="bad") | table bytes`, 1, true}, {` | where status=200 OR status=201 | table bytes`, 0, true}, {` | where NOT status=200 | table bytes`, 0, false}, {` | where status=200 | eval status=other | table bytes`, 0, false}, {` | where status=200 | fields - status | table bytes`, 0, false}} {
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
	s := rewriteTestSession(t, "spl2", `tstats aggregates=[count()] predicate=(index=main AND port=443) byfields=[host]`, RewriteFactProbe{Kind: "index", Identity: rewriteName("main")})
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
			{`index="main*"`, "index", "main*"}, {`index=main*`, "index", "main*"},
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
			{`search index="main"`, "index", "main", `"main"`},
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
		language, expression string
		eligible             bool
	}{
		{"spl", `unknown(isnull(user))=1`, false},
		{"spl2", `unknown(isnull(user))=1`, false},
		{"spl2", `abs(value:isnull(user))=1`, false},
		{"spl", `isnull(user)`, true},
		{"spl2", `isnull(user)`, true},
	} {
		t.Run(tc.language+"/"+tc.expression, func(t *testing.T) {
			s := rewriteTestSession(t, tc.language, "search index=main | where "+tc.expression)
			site := rewriteFind(t, s, "field", "user", 0)
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
