package analysis

import "testing"

// Catches discarding independent assignments alongside a damaged assignment.
func TestRecoverySoundAssignments(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search host=web | eval good=host, broken= | stats count by user`})
	if r.Status != Invalid || r.Coverage.SyntaxComplete || r.Coverage.SemanticComplete {
		t.Fatal(r)
	}
	scopedReference(t, r, "scope-0", "host", "filter")
	scopedReference(t, r, "scope-0", "good", "create")
	scopedReference(t, r, "scope-0", "user", "group")
	for _, ref := range r.References {
		if ref.NormalizedName == "broken" {
			t.Fatal("invented broken output", ref)
		}
	}
}
func TestRecoveryUnknownEffectsPrecedence(t *testing.T) {
	for _, tc := range []struct {
		query   string
		status  Status
		binding string
	}{
		{`search a=1 | fields - a | mystery x | where a=2`, Incomplete, "indeterminate"},
		{`search a=1 | fields - a | where a=2 | mystery x`, Invalid, "unavailable"},
	} {
		r, _ := Analyze(QueryDocument{Text: tc.query})
		if r.Status != tc.status || r.Coverage.SemanticComplete {
			t.Fatal(r)
		}
		ref := scopedReference(t, r, "scope-0", "a", "read")
		if ref.Binding != tc.binding {
			t.Fatal(ref)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "x" {
				t.Fatal("inferred opaque argument", ref)
			}
		}
	}
}

// Quoting is grammatical identity, so '*' inside a quoted identifier is exact.
func TestRecoveryQuotedAsteriskResolution(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `| eval 'a*'=1 | where 'a*'=1`})
	if r.Status != Valid || len(r.References) != 2 {
		t.Fatal(r)
	}
	for _, ref := range r.References {
		if ref.NormalizedName != "a*" || ref.Resolution != "exact" {
			t.Fatal(ref)
		}
	}
	if r.References[1].Binding != "derived" {
		t.Fatal(r.References)
	}
}
func TestRecoveryMacroBetweenStages(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: "search host=web | `expand(host)` | stats count by user"})
	if r.Status != Incomplete || !r.Coverage.SyntaxComplete || len(r.Stages) != 3 {
		t.Fatal(r)
	}
	if len(r.Dependencies.Macros) != 1 || r.Dependencies.Macros[0] != "expand" {
		t.Fatal(r.Dependencies)
	}
	ref := scopedReference(t, r, "scope-0", "expand", "read")
	if ref.Kind != "macro" || ref.OriginalName != "expand" || ref.Resolution != "exact" {
		t.Fatal(ref)
	}
	scopedReference(t, r, "scope-0", "user", "group")
	for _, ref := range r.References {
		if ref.OriginalName == "host" && ref.Role == "read" {
			t.Fatal("macro argument invented as field", ref)
		}
	}
}

// Lexer deletion must not turn a malformed assignment into a known output;
// a pipe inside an unclosed expression cannot justify a recovered stage.
func TestRecoveryDamageBoundaries(t *testing.T) {
	for _, tc := range []struct {
		query      string
		keep, omit []string
	}{
		{`search host=web | eval good=host, bad=$oops | table good`, []string{"host", "good"}, []string{"bad", "oops"}},
		{`search host=web | where (bad=1 | stats count by user`, []string{"host"}, []string{"user", "count"}},
		{`search host=web [ search child=1`, []string{"host"}, nil},
	} {
		t.Run(tc.query, func(t *testing.T) {
			r, _ := Analyze(QueryDocument{Text: tc.query})
			if r.Status != Invalid {
				t.Fatal(r)
			}
			for _, name := range tc.keep {
				found := false
				for _, ref := range r.References {
					found = found || ref.NormalizedName == name
				}
				if !found {
					t.Fatalf("lost sound %s: %+v", name, r.References)
				}
			}
			for _, ref := range r.References {
				for _, name := range tc.omit {
					if ref.NormalizedName == name {
						t.Fatal("damaged reference", ref)
					}
				}
			}
		})
	}
}

func TestRecoveryTrailingLexicalDamage(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search host=web | eval good=1, bad=2$ | where bad=2`})
	if r.Status != Invalid {
		t.Fatal(r)
	}
	scopedReference(t, r, "scope-0", "good", "create")
	for _, ref := range r.References {
		if ref.NormalizedName == "bad" && ref.Role == "create" {
			t.Fatal("lexer deletion fabricated output", ref)
		}
	}
	if scopedReference(t, r, "scope-0", "bad", "read").Binding != "indeterminate" {
		t.Fatal(r.References)
	}
}

func TestRecoveryUnknownFunctionInSoundPrefix(t *testing.T) {
	r, _ := Analyze(QueryDocument{Text: `search host=web | eval good=unknownpure(host), broken= | where good=1`})
	if r.Status != Invalid || r.Coverage.SemanticComplete {
		t.Fatal(r)
	}
	scopedReference(t, r, "scope-0", "good", "create")
	if scopedReference(t, r, "scope-0", "good", "read").Binding != "indeterminate" {
		t.Fatal("recovery invented certain unknown-function output", r.References)
	}
}

func TestRecoveryLeadingLexerDamage(t *testing.T) {
	for _, q := range []string{`search $host=web | table other`, `search * | eval $bad=1 | where bad=2`} {
		r, _ := Analyze(QueryDocument{Text: q})
		if r.Status != Invalid {
			t.Fatal(r)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "host" || (ref.NormalizedName == "bad" && ref.Role == "create") {
				t.Fatal("deleted prefix fabricated reference", ref)
			}
		}
	}
}

// Quoted projection patterns must match another known name, not be accepted as
// exact merely because the query also created a field literally named a*.
func TestRecoveryQuotedSelectorPatterns(t *testing.T) {
	for _, command := range []string{"table", "fields"} {
		r, _ := Analyze(QueryDocument{Text: `| eval 'a*'=1,ab=2 | ` + command + ` 'a*'`})
		if r.Status != Incomplete || len(r.References) != 3 || r.References[0].Resolution != "exact" || r.References[2].Resolution != "wildcard" {
			t.Fatal(r)
		}
		found := false
		for _, f := range r.Lineage[1].After.Fields {
			found = found || f.Name == "ab"
		}
		if !found {
			t.Fatal("quoted selector lost matching ab", r.Lineage)
		}
	}
	r, _ := Analyze(QueryDocument{Text: `| eval ab=1,other=2 | table ab other | fields - 'a*' | where ab=2`})
	if r.Status != Invalid || scopedReference(t, r, "scope-0", "ab", "read").Binding == "" {
		t.Fatal(r)
	}
	if r.References[len(r.References)-1].Binding != "unavailable" {
		t.Fatal("quoted exclusion did not remove matching ab", r.References)
	}
	r, _ = Analyze(QueryDocument{Text: `| eval ab=1 | rename 'a*' AS b | where ab=2`})
	if r.Status != Incomplete || r.References[1].Resolution != "wildcard" || r.References[len(r.References)-1].Binding != "indeterminate" {
		t.Fatal(r)
	}
}
func TestRecoveryUnmodeledSelectorCommands(t *testing.T) {
	for _, command := range []string{`sort 'a*'`, `dedup 'a*'`, `stats count BY 'a*'`, `lookup people id AS 'a*' OUTPUT name AS out`} {
		r, _ := Analyze(QueryDocument{Text: `| eval 'a*'=1,ab=2 | ` + command})
		if r.Status != Incomplete || r.Coverage.SemanticComplete {
			t.Fatal("unestablished wildcard form claimed complete", r)
		}
	}
}

func TestRecoveryUnmodeledPatternOutputs(t *testing.T) {
	for _, command := range []string{`lookup people id AS ab OUTPUT name AS 'b*'`, `rename ab AS 'b*'`} {
		r, _ := Analyze(QueryDocument{Text: `| eval ab=1 | ` + command})
		if r.Status != Incomplete || r.Coverage.SemanticComplete {
			t.Fatal("unmodeled pattern output claimed complete", r)
		}
		for _, ref := range r.References {
			if ref.NormalizedName == "b*" && (ref.Role == "output" || ref.Role == "rename") {
				t.Fatal("invented exact pattern output", ref)
			}
		}
	}
}

// Invalid Go string bytes cannot be silently replaced during JSON encoding.
func TestRecoveryInvalidDocumentEncoding(t *testing.T) {
	for _, tc := range []struct {
		document QueryDocument
		message  string
	}{
		{QueryDocument{Text: string([]byte{0xff})}, "document text is not valid UTF-8"},
		{QueryDocument{Text: "search *", SourceID: string([]byte{0xc3})}, "document source_id is not valid UTF-8"},
	} {
		r, err := Analyze(tc.document)
		if err == nil || r != nil || err.Error() != tc.message {
			t.Errorf("result=%+v error=%v", r, err)
		}
	}
	r, err := Analyze(QueryDocument{Text: `search host="�"`, SourceID: "�"})
	if err != nil || r.Document.SourceID != "�" || r.Status != Valid {
		t.Fatal("valid replacement rune was rejected", r, err)
	}
}
