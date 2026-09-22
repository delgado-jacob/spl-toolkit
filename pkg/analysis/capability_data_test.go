package analysis

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"testing"
)

func TestEmbeddedCapabilityAssetsAreStructurallyValid(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	if len(records) == 0 || len(cases) == 0 {
		t.Fatalf("embedded capability data must not be empty: records=%d cases=%d", len(records), len(cases))
	}
	var authored struct {
		Records []map[string]json.RawMessage `json:"records"`
	}
	if err := json.Unmarshal(embeddedCapabilityLedger, &authored); err != nil {
		t.Fatalf("inspect embedded capability ledger: %v", err)
	}
	for i, record := range authored.Records {
		raw, present := record["grammar_registered"]
		var registered bool
		if !present || json.Unmarshal(raw, &registered) != nil {
			t.Errorf("embedded capability record %d lacks an explicit boolean grammar_registered value", i)
		}
	}
}

func TestCapabilityLedgerCoversEveryKindPerLanguage(t *testing.T) {
	records, _, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	for _, language := range []string{"spl", "spl2"} {
		kinds := make(map[string]bool, len(capabilityRecordKinds))
		for _, record := range records {
			if record.Language == language {
				kinds[record.Kind] = true
			}
		}
		for kind := range capabilityRecordKinds {
			if !kinds[kind] {
				t.Errorf("%s ledger has no %s record", language, kind)
			}
		}
	}
}

func TestCapabilityLedgerCoversLegacyInventory(t *testing.T) {
	records, _, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}

	actualNames := func(language, kind string) []string {
		set := map[string]bool{}
		for _, record := range records {
			if record.Language == language && record.Kind == kind {
				set[record.Name] = true
			}
		}
		names := make([]string, 0, len(set))
		for name := range set {
			names = append(names, name)
		}
		sort.Strings(names)
		return names
	}
	wantSPLCommands := make([]string, 0, len(commands))
	for name := range commands {
		wantSPLCommands = append(wantSPLCommands, name)
	}
	sort.Strings(wantSPLCommands)
	wantSPLFunctions := make([]string, 0, len(functions))
	for name := range functions {
		wantSPLFunctions = append(wantSPLFunctions, name)
	}
	sort.Strings(wantSPLFunctions)
	wantSPL2Commands := make([]string, 0, len(spl2CommandInventory))
	for _, entry := range spl2CommandInventory {
		wantSPL2Commands = append(wantSPL2Commands, entry.name)
	}
	sort.Strings(wantSPL2Commands)
	wantSPL2Functions := make([]string, 0, len(spl2Functions))
	for name := range spl2Functions {
		wantSPL2Functions = append(wantSPL2Functions, name)
	}
	sort.Strings(wantSPL2Functions)

	for _, check := range []struct {
		language string
		kind     string
		want     []string
	}{
		{language: "spl", kind: "command", want: wantSPLCommands},
		{language: "spl", kind: "function", want: wantSPLFunctions},
		{language: "spl2", kind: "command", want: wantSPL2Commands},
		{language: "spl2", kind: "function", want: wantSPL2Functions},
	} {
		if got := actualNames(check.language, check.kind); !slices.Equal(got, check.want) {
			t.Errorf("%s %s names = %v, want %v", check.language, check.kind, got, check.want)
		}
	}

	legacy := []CapabilityManifest{Capabilities()}
	spl2Legacy, err := CapabilitiesFor(CapabilityOptions{Language: "spl2"})
	if err != nil {
		t.Fatalf("load SPL2 legacy capabilities: %v", err)
	}
	legacy = append(legacy, spl2Legacy)
	for _, manifest := range legacy {
		for kind, capabilities := range map[string][]Capability{
			"command":  manifest.Commands,
			"function": manifest.Functions,
		} {
			for _, capability := range capabilities {
				var grammarRegistered, semanticSupported bool
				var limitations []string
				seenLimitations := map[string]bool{}
				for _, record := range records {
					if record.Language != manifest.Language || record.Kind != kind || record.Name != capability.Name {
						continue
					}
					grammarRegistered = grammarRegistered || record.GrammarRegistered
					semanticSupported = semanticSupported || record.Dimensions.Semantics.State == CapabilitySupported
					for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
						for _, limitation := range dimension.claim.Limitations {
							if !seenLimitations[limitation] {
								seenLimitations[limitation] = true
								limitations = append(limitations, limitation)
							}
						}
					}
				}
				if grammarRegistered != capability.SyntaxSupported || semanticSupported != capability.SemanticSupported {
					t.Errorf("%s %s %s support = grammar:%v semantics:%v, want syntax:%v semantics:%v", manifest.Language, kind, capability.Name, grammarRegistered, semanticSupported, capability.SyntaxSupported, capability.SemanticSupported)
				}
				if !slices.Equal(limitations, capability.Limitations) {
					t.Errorf("%s %s %s limitations = %v, want %v", manifest.Language, kind, capability.Name, limitations, capability.Limitations)
				}
			}
		}
	}
}

func TestCapabilityGrammarRegistrationDoesNotAddSyntaxCoverage(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}
	for _, record := range records {
		if record.ID != "spl2.command.spl1.quoted-pipeline" {
			continue
		}
		if !record.GrammarRegistered {
			t.Fatal("SPL2 spl1 grammar registration was not authored")
		}
		if record.Dimensions.Syntax.State != CapabilityUnsupported {
			t.Fatalf("SPL2 spl1 syntax state = %q, want %q", record.Dimensions.Syntax.State, CapabilityUnsupported)
		}
		if len(record.Dimensions.Syntax.EvidenceIDs) == 0 {
			t.Fatal("SPL2 spl1 syntax boundary has no evidence")
		}
		for _, evidenceID := range record.Dimensions.Syntax.EvidenceIDs {
			evidence := evidenceByID[evidenceID]
			if evidence.Classification != CapabilityEvidenceIncomplete || evidence.Observations.Syntax == nil || evidence.Observations.Syntax.Complete {
				t.Fatalf("SPL2 spl1 syntax evidence %q is not an honest incomplete observation: %+v", evidenceID, evidence)
			}
		}
		summary := summarizeCapabilityRecords([]CapabilityRecord{record})
		if summary.Syntax.Covered != 0 || summary.Syntax.Unsupported != 1 {
			t.Fatalf("grammar registration added syntax coverage: %+v", summary.Syntax)
		}
		return
	}
	t.Fatal("SPL2 spl1 capability record is missing")
}

func TestEmbeddedCapabilityReviewedFacts(t *testing.T) {
	_, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}
	evidence := func(id string) CapabilityEvidence {
		t.Helper()
		got, ok := evidenceByID[id]
		if !ok {
			t.Fatalf("capability evidence %q is missing", id)
		}
		return got
	}

	for _, id := range []string{
		"spl.splunkd.baseline.rewrite",
		"spl2.splunkd.baseline.rewrite",
	} {
		got := evidence(id)
		if got.Classification != CapabilityEvidenceNegative || got.Observations.SafeRewriting == nil {
			t.Fatalf("%s is not an authored negative rewrite boundary: %+v", id, got)
		}
		if got.Observations.SafeRewriting.Status != Invalid {
			t.Errorf("%s rewrite status = %q, want %q", id, got.Observations.SafeRewriting.Status, Invalid)
		}
		if len(got.Observations.SafeRewriting.ChangeReasons) != 0 || len(got.Observations.SafeRewriting.RuleEvaluationReasons) != 0 {
			t.Errorf("%s records nonexistent rewrite or rule evaluations: %+v", id, got.Observations.SafeRewriting)
		}
	}
	if got := evidence("spl.splunkd.baseline.rewrite").Observations.SafeRewriting.CoverageReasons; !slices.Equal(got, []string{"SPL_SYNTAX_ERROR", "post_verification_failed"}) {
		t.Errorf("SPL rewrite coverage reasons = %v", got)
	}
	if got := evidence("spl2.splunkd.baseline.rewrite").Observations.SafeRewriting.CoverageReasons; !slices.Equal(got, []string{"SPL_PROFILE_MISMATCH", "SPL_UNSUPPORTED_SEMANTICS", "post_verification_failed"}) {
		t.Errorf("SPL2 rewrite coverage reasons = %v", got)
	}

	for _, id := range []string{
		"spl2.decrypt.profile-mismatch.incomplete",
		"spl2.fillnull.field-list.incomplete",
		"spl2.if.subpipe.incomplete",
		"spl2.ocsf.profile-mismatch.incomplete",
		"spl2.route.profile-mismatch.incomplete",
		"spl2.timewrap.span.incomplete",
	} {
		got := evidence(id)
		if got.Observations.Semantics == nil || got.Observations.Semantics.Status != Invalid {
			t.Errorf("%s semantic status is not invalid: %+v", id, got.Observations.Semantics)
		}
	}

	macro := evidence("spl.macro.exact-invocation.incomplete")
	if got := macro.Observations.Semantics.Stages; len(got) != 1 || got[0].Command != "search" || got[0].SemanticComplete {
		t.Errorf("macro invocation stages = %+v, want incomplete search stage", got)
	}

	for _, check := range []struct {
		id          string
		startOffset int
		startColumn int
		endOffset   int
		endColumn   int
	}{
		{id: "spl.splunkd.baseline.negative", startOffset: 13, startColumn: 14, endOffset: 13, endColumn: 14},
		{id: "spl2.splunkd.baseline.negative", startOffset: 12, startColumn: 13, endOffset: 19, endColumn: 20},
	} {
		got := evidence(check.id)
		for surface, diagnostics := range map[string][]CapabilityDiagnosticExpectation{
			"syntax":    got.Observations.Syntax.Diagnostics,
			"semantics": got.Observations.Semantics.Diagnostics,
		} {
			if len(diagnostics) != 1 {
				t.Errorf("%s %s diagnostics = %+v, want one", check.id, surface, diagnostics)
				continue
			}
			location := diagnostics[0].Location
			if location.Start.Offset != check.startOffset || location.Start.Line != 1 || location.Start.Column != check.startColumn ||
				location.End.Offset != check.endOffset || location.End.Line != 1 || location.End.Column != check.endColumn {
				t.Errorf("%s %s diagnostic location = %+v", check.id, surface, location)
			}
		}
	}

	timewrap := evidence("spl2.timewrap.span.positive")
	if timewrap.Document.Text != "FROM main | timechart count() | timewrap 2day" {
		t.Fatalf("timewrap positive witness = %q", timewrap.Document.Text)
	}
	result, err := Analyze(timewrap.Document)
	if err != nil {
		t.Fatalf("analyze timewrap positive witness: %v", err)
	}
	if !result.Coverage.SyntaxComplete || result.Status == Invalid {
		t.Fatalf("timewrap positive witness status = %q, syntax complete = %v", result.Status, result.Coverage.SyntaxComplete)
	}
}

func TestMilestone10CapabilityClaimsStayBounded(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	recordByID := make(map[string]CapabilityRecord, len(records))
	for _, record := range records {
		recordByID[record.ID] = record
	}
	evidenceByID := make(map[string]CapabilityEvidence, len(cases))
	for _, evidence := range cases {
		evidenceByID[evidence.ID] = evidence
	}

	supportedIDs := []string{
		"spl.command.bin.exact-field",
		"spl.command.bucket.exact-field",
		"spl.command.fillnull.exact-fields",
		"spl.command.mvexpand.exact-field",
		"spl.command.regex.exact-field",
		"spl.command.rex.named-captures",
		"spl.command.spath.explicit-output",
		"spl.command.tstats.exact-model-dataset",
		"spl.function.earliest.one-positional",
		"spl.function.latest.one-positional",
		"spl.function.like.two-positional",
		"spl.function.mvfind.two-positional",
		"spl.function.mvindex.two-positional",
		"spl.function.now.zero-positional",
		"spl.function.null.zero-positional",
		"spl.function.relative_time.two-positional",
		"spl.function.stdev.one-positional",
		"spl.function.strftime.two-positional",
		"spl.function.true.zero-positional",
	}
	for _, id := range supportedIDs {
		record, ok := recordByID[id]
		if !ok {
			t.Errorf("Milestone 10 record %q is missing", id)
			continue
		}
		for name, claim := range map[string]CapabilityClaim{
			"syntax": record.Dimensions.Syntax, "semantics": record.Dimensions.Semantics, "requirements": record.Dimensions.Requirements,
		} {
			if claim.State != CapabilitySupported || len(claim.EvidenceIDs) != 1 {
				t.Errorf("%s %s claim = %+v, want one supported witness", id, name, claim)
				continue
			}
			if name != "requirements" {
				continue
			}
			observation := evidenceByID[claim.EvidenceIDs[0]].Observations.Requirements
			if observation == nil || !observation.Complete || len(observation.Items) == 0 || len(observation.GapCodes) != 0 {
				t.Errorf("%s supported requirements observation = %+v, want exact nonempty items and no gaps", id, observation)
			}
		}
	}

	inline := recordByID["spl.command.tstats.inline-macro"]
	if inline.Dimensions.Syntax.State != CapabilitySupported || inline.Dimensions.Semantics.State != CapabilityUnsupported || inline.Dimensions.Requirements.State != CapabilityUnsupported {
		t.Errorf("inline tstats macro claims = %+v, want supported syntax with unsupported semantics and requirements", inline.Dimensions)
	}
	for _, id := range []string{
		"spl.command.append.append-subsearch",
		"spl.command.appendpipe.appendpipe-subsearch",
		"spl.command.join.field-subsearch",
		"spl.command.macro.exact-invocation",
	} {
		if got := recordByID[id].Dimensions.Requirements.State; got != CapabilityUnsupported {
			t.Errorf("%s requirements state = %q, want %q", id, got, CapabilityUnsupported)
		}
	}
	appendpipe := evidenceByID["spl.appendpipe.appendpipe-subsearch.incomplete"]
	wantAppendpipeItems := []CapabilityRequirementExpectation{
		{Kind: "index", Identity: "main", Role: "read", Necessity: "required", Resolution: "exact"},
		{Kind: "field", Identity: "child", Role: "filter", Necessity: "required", Resolution: "exact"},
	}
	if got := appendpipe.Observations.Requirements; got == nil || !slices.Equal(got.Items, wantAppendpipeItems) || !slices.Equal(got.GapCodes, []string{"SPL_UNSUPPORTED_SEMANTICS"}) {
		t.Errorf("appendpipe retained requirements = %+v, want exact parent and child items plus merge gap", got)
	}

	tstats := evidenceByID["spl.tstats.exact-model-dataset.positive"]
	if got := tstats.Observations.Semantics; got == nil || !got.Complete ||
		!slices.Contains(got.Dependencies, CapabilityDependencyExpectation{Kind: "data_model", Name: "Authentication"}) ||
		!slices.Contains(got.Dependencies, CapabilityDependencyExpectation{Kind: "dataset", Name: "Authentication.Authentication"}) ||
		!slices.Contains(got.Transitions, CapabilityTransitionExpectation{Operation: "aggregate", Output: "total"}) ||
		!slices.Contains(got.Transitions, CapabilityTransitionExpectation{Operation: "aggregate", Output: "count"}) {
		t.Errorf("exact tstats semantic witness lacks reviewed source or aggregate facts: %+v", got)
	}

	spl2Revision, err := capabilityRevisionFor(CapabilityOptions{Language: "spl2", Profile: "splunkd", Version: "current"})
	if err != nil {
		t.Fatal(err)
	}
	const wantSPL2Revision = "sha256:f1391296cfbc616e9bb1b1828e2471e37b60a35c0555654c0734640e072a0437"
	if spl2Revision != wantSPL2Revision {
		t.Errorf("SPL2 capability revision = %q, want preserved %q", spl2Revision, wantSPL2Revision)
	}
}

func TestCapabilityEvidenceIsReferenced(t *testing.T) {
	records, cases, err := loadEmbeddedCapabilityData()
	if err != nil {
		t.Fatalf("load embedded capability data: %v", err)
	}
	references := make(map[string]int, len(cases))
	for _, record := range records {
		for _, dimension := range capabilityDimensionClaims(record.Dimensions) {
			for _, evidenceID := range dimension.claim.EvidenceIDs {
				references[evidenceID]++
			}
		}
	}
	for _, evidence := range cases {
		if references[evidence.ID] == 0 {
			t.Errorf("evidence %q is not referenced", evidence.ID)
		}
	}
}

func TestDecodeCapabilityAssetsRejectsMalformedInput(t *testing.T) {
	ledger, corpus := validCapabilityAssets(t)

	tests := []struct {
		name   string
		mutate func([]byte, []byte) ([]byte, []byte)
	}{
		{
			name: "unknown ledger field",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return replaceJSON(t, ledger, "schema_version", 1, "unknown", true), corpus
			},
		},
		{
			name: "unknown corpus field",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, replaceJSON(t, corpus, "schema_version", 1, "unknown", true)
			},
		},
		{
			name: "trailing ledger JSON",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return append(ledger, []byte(` {}`)...), corpus
			},
		},
		{
			name: "trailing corpus JSON",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, append(corpus, []byte(` {}`)...)
			},
		},
		{
			name: "wrong ledger schema version",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return replaceJSON(t, ledger, "schema_version", 2), corpus
			},
		},
		{
			name: "wrong corpus schema version",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return ledger, replaceJSON(t, corpus, "schema_version", 2)
			},
		},
		{
			name: "duplicate record IDs",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityLedgerFile
				mustUnmarshal(t, ledger, &file)
				file.Records = append(file.Records, file.Records[0])
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "duplicate evidence IDs",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases = append(file.Cases, file.Cases[0])
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "missing record ID",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityLedgerFile
				mustUnmarshal(t, ledger, &file)
				file.Records[0].ID = ""
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "missing evidence ID",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].ID = ""
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "unknown record language",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Language = "sql" })
			},
		},
		{
			name: "unknown record profile",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Profile = "cloud" })
			},
		},
		{
			name: "unknown record kind",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) { record.Kind = "clause" })
			},
		},
		{
			name: "invalid grammar registration type",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file map[string]any
				mustUnmarshal(t, ledger, &file)
				records := file["records"].([]any)
				records[0].(map[string]any)["grammar_registered"] = "yes"
				return mustJSON(t, file), corpus
			},
		},
		{
			name: "unknown state",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Dimensions.Syntax.State = "unknown"
				})
			},
		},
		{
			name: "unknown classification",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Classification = "unknown"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "unknown record source family",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Provenance.SourceFamily = "blog"
				})
			},
		},
		{
			name: "unknown evidence source family",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Provenance.SourceFamily = "blog"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "missing dimensions object",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return deleteNestedJSONField(t, ledger, "records", 0, "dimensions"), corpus
			},
		},
		{
			name: "missing fixed dimension",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return deleteRecordDimension(t, ledger, "safe_rewriting"), corpus
			},
		},
		{
			name: "evidence selector mismatch",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				file.Cases[0].Document.Language = "spl2"
				return ledger, mustJSON(t, file)
			},
		},
		{
			name: "dangling evidence link",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				return mutateRecord(t, ledger, corpus, func(record *CapabilityRecord) {
					record.Dimensions.Syntax.EvidenceIDs = []string{"missing"}
				})
			},
		},
		{
			name: "unreferenced evidence",
			mutate: func(ledger, corpus []byte) ([]byte, []byte) {
				var file capabilityCorpusFile
				mustUnmarshal(t, corpus, &file)
				extra := cloneCapabilityEvidence(file.Cases[0])
				extra.ID = "unused"
				extra.Document.SourceID = "capability:unused"
				file.Cases = append(file.Cases, extra)
				return ledger, mustJSON(t, file)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			badLedger, badCorpus := tc.mutate(append([]byte(nil), ledger...), append([]byte(nil), corpus...))
			if _, _, err := decodeCapabilityAssets(badLedger, badCorpus); err == nil {
				t.Fatal("malformed capability assets were accepted")
			}
		})
	}
}

func TestDecodeCapabilityAssetsRejectsMalformedObservations(t *testing.T) {
	ledger, corpus := validCapabilityAssets(t)
	tests := []struct {
		name    string
		mutate  func(*CapabilityEvidence)
		wantErr bool
	}{
		{name: "valid representative"},
		{
			name: "positive syntax marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Complete = false
			},
			wantErr: true,
		},
		{
			name: "syntax diagnostic without code",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Code = ""
			},
			wantErr: true,
		},
		{
			name: "semantics with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Status = "mystery"
			},
			wantErr: true,
		},
		{
			name: "complete semantic and requirements observations with invalid overall status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Status = Invalid
				evidence.Observations.Requirements.QueryStatus = Invalid
			},
		},
		{
			name: "positive semantics marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Complete = false
			},
			wantErr: true,
		},
		{
			name: "semantics without typed facts",
			mutate: func(evidence *CapabilityEvidence) {
				observation := evidence.Observations.Semantics
				observation.Stages = nil
				observation.References = nil
				observation.Dependencies = nil
				observation.Transitions = nil
				observation.Diagnostics = nil
			},
			wantErr: true,
		},
		{
			name: "semantics stage without command",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Stages[0].Command = ""
			},
			wantErr: true,
		},
		{
			name: "positive semantics with incomplete stage",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Stages[0].SemanticComplete = false
			},
			wantErr: true,
		},
		{
			name: "semantics reference without binding",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.References[0].Binding = ""
			},
			wantErr: true,
		},
		{
			name: "semantics reference without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.References[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "semantics dependency without name",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Dependencies[0].Name = ""
			},
			wantErr: true,
		},
		{
			name: "semantics transition without output",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Transitions[0].Output = ""
			},
			wantErr: true,
		},
		{
			name: "semantics diagnostic without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Semantics.Diagnostics[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "requirements with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.QueryStatus = "mystery"
			},
			wantErr: true,
		},
		{
			name: "positive requirements marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Complete = false
			},
			wantErr: true,
		},
		{
			name: "requirements without typed content",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Items = nil
				evidence.Observations.Requirements.GapCodes = nil
			},
			wantErr: true,
		},
		{
			name: "requirement item without identity",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.Items[0].Identity = ""
			},
			wantErr: true,
		},
		{
			name: "blank requirement gap code",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.GapCodes = []string{" "}
			},
			wantErr: true,
		},
		{
			name: "complete requirements with a gap",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Requirements.GapCodes = []string{"TEST"}
			},
			wantErr: true,
		},
		{
			name: "lint diagnostic without category",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics[0].Category = ""
			},
			wantErr: true,
		},
		{
			name: "lint diagnostic without location",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics[0].Location = Location{}
			},
			wantErr: true,
		},
		{
			name: "positive linting observation without diagnostics",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Linting.Diagnostics = nil
			},
		},
		{
			name: "diagnostic with unknown severity",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Severity = "mystery"
			},
			wantErr: true,
		},
		{
			name: "diagnostic location beyond document",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location.End = Position{Offset: 99, Line: 1, Column: 100}
			},
			wantErr: true,
		},
		{
			name: "diagnostic location with incoherent column",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location.Start.Column = 2
			},
			wantErr: true,
		},
		{
			name: "diagnostic location with reversed range",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.Syntax.Diagnostics[0].Location = Location{
					Start: Position{Offset: 6, Line: 1, Column: 7},
					End:   Position{Offset: 0, Line: 1, Column: 1},
				}
			},
			wantErr: true,
		},
		{
			name: "canonical zero length locations in empty document",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Document.Text = ""
				location := Location{
					Start: Position{Offset: 0, Line: 1, Column: 1},
					End:   Position{Offset: 0, Line: 1, Column: 1},
				}
				for i := range evidence.Observations.Syntax.Diagnostics {
					evidence.Observations.Syntax.Diagnostics[i].Location = location
				}
				for i := range evidence.Observations.Semantics.Diagnostics {
					evidence.Observations.Semantics.Diagnostics[i].Location = location
				}
				for i := range evidence.Observations.Semantics.References {
					evidence.Observations.Semantics.References[i].Location = location
				}
				for i := range evidence.Observations.Linting.Diagnostics {
					evidence.Observations.Linting.Diagnostics[i].Location = location
				}
			},
		},
		{
			name: "safe rewrite with unknown status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Status = "mystery"
			},
			wantErr: true,
		},
		{
			name: "positive safe rewrite marked incomplete",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.RewriteComplete = false
			},
			wantErr: true,
		},
		{
			name: "complete rewrite observation with incomplete overall status",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Status = Incomplete
			},
		},
		{
			name: "safe rewrite without text",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.Text = ""
			},
			wantErr: true,
		},
		{
			name: "safe rewrite with blank reason",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.CoverageReasons = []string{" "}
			},
			wantErr: true,
		},
		{
			name: "committed rewrite differs from candidate",
			mutate: func(evidence *CapabilityEvidence) {
				evidence.Observations.SafeRewriting.CandidateText = "search index=other"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidateCorpus := append([]byte(nil), corpus...)
			if tc.mutate != nil {
				candidateCorpus = mutateCapabilityEvidence(t, candidateCorpus, tc.mutate)
			}
			_, _, err := decodeCapabilityAssets(ledger, candidateCorpus)
			if (err != nil) != tc.wantErr {
				t.Fatalf("decodeCapabilityAssets() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestValidateCapabilityClaims(t *testing.T) {
	record := validCapabilityRecord()
	evidence := map[string]CapabilityEvidence{
		"positive":   validCapabilityEvidence("positive", CapabilityEvidencePositive),
		"incomplete": validCapabilityEvidence("incomplete", CapabilityEvidenceIncomplete),
		"negative":   validCapabilityEvidence("negative", CapabilityEvidenceNegative),
	}

	tests := []struct {
		name      string
		dimension string
		claim     CapabilityClaim
		mutate    func(map[string]CapabilityEvidence)
		wantErr   bool
	}{
		{name: "supported", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"positive"}}},
		{name: "supported without evidence", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported}, wantErr: true},
		{name: "supported by negative evidence", dimension: "syntax", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}}, wantErr: true},
		{name: "partial", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive", "incomplete"}, Limitations: []string{"dynamic forms remain incomplete"}}},
		{name: "partial without positive", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"incomplete"}, Limitations: []string{"limited"}}, wantErr: true},
		{name: "partial without incomplete", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive"}, Limitations: []string{"limited"}}, wantErr: true},
		{name: "partial without limitation", dimension: "semantics", claim: CapabilityClaim{State: CapabilityPartial, EvidenceIDs: []string{"positive", "incomplete"}}, wantErr: true},
		{name: "unsupported by negative", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"negative"}, Limitations: []string{"not modeled"}}},
		{name: "unsupported by incomplete boundary", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"incomplete"}, Limitations: []string{"not modeled"}}},
		{name: "unsupported by positive", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"positive"}, Limitations: []string{"not modeled"}}, wantErr: true},
		{name: "unsupported without limitation", dimension: "requirements", claim: CapabilityClaim{State: CapabilityUnsupported, EvidenceIDs: []string{"negative"}}, wantErr: true},
		{name: "not applicable", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable, Limitations: []string{"rewriting does not apply"}}},
		{name: "not applicable without reason", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable}, wantErr: true},
		{name: "not applicable with evidence", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityNotApplicable, EvidenceIDs: []string{"positive"}, Limitations: []string{"not applicable"}}, wantErr: true},
		{name: "unassessed", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed}},
		{name: "unassessed with evidence", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed, EvidenceIDs: []string{"positive"}}, wantErr: true},
		{name: "unassessed with limitation", dimension: "safe_rewriting", claim: CapabilityClaim{State: CapabilityUnassessed, Limitations: []string{"unknown"}}, wantErr: true},
		{name: "supported linting by negative case with exact observation", dimension: "linting", claim: CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}}},
		{
			name:      "supported linting by negative case without exact observation",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting = nil
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "supported linting by negative case without diagnostics",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics = nil
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without code",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Code = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without category",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Category = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without severity",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Severity = ""
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "negative linting diagnostic without complete location",
			dimension: "linting",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"negative"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["negative"]
				item.Observations.Linting.Diagnostics[0].Location = Location{}
				evidence["negative"] = item
			},
			wantErr: true,
		},
		{
			name:      "evidence observation belongs to another dimension",
			dimension: "syntax",
			claim:     CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{"positive"}},
			mutate: func(evidence map[string]CapabilityEvidence) {
				item := evidence["positive"]
				item.Observations.Syntax = nil
				evidence["positive"] = item
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cases := cloneEvidenceMap(evidence)
			if tc.mutate != nil {
				tc.mutate(cases)
			}
			err := validateCapabilityClaim(tc.dimension, record, tc.claim, cases)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCapabilityClaim() error = %v, wantErr %t", err, tc.wantErr)
			}
		})
	}
}

func TestCapabilitySummaryUsesStrictDenominator(t *testing.T) {
	states := []CapabilityState{
		CapabilitySupported,
		CapabilityPartial,
		CapabilityUnsupported,
		CapabilityNotApplicable,
		CapabilityUnassessed,
	}
	records := make([]CapabilityRecord, 0, len(states))
	for i, state := range states {
		record := validCapabilityRecord()
		record.ID = fmt.Sprintf("record-%d", i)
		claim := CapabilityClaim{State: state}
		record.Dimensions = CapabilityDimensions{
			Syntax:        claim,
			Semantics:     claim,
			Requirements:  claim,
			Linting:       claim,
			SafeRewriting: claim,
		}
		records = append(records, record)
	}

	summary := summarizeCapabilityRecords(records)
	for name, counts := range map[string]CapabilityStateCounts{
		"syntax":         summary.Syntax,
		"semantics":      summary.Semantics,
		"requirements":   summary.Requirements,
		"linting":        summary.Linting,
		"safe_rewriting": summary.SafeRewriting,
	} {
		if counts.Applicable != counts.Supported+counts.Partial+counts.Unsupported+counts.Unassessed {
			t.Errorf("%s applicable denominator is not strict: %+v", name, counts)
		}
		if counts.Covered != counts.Supported {
			t.Errorf("%s covered count differs from supported: %+v", name, counts)
		}
		if len(records) != counts.Applicable+counts.NotApplicable {
			t.Errorf("%s record count identity failed: records=%d counts=%+v", name, len(records), counts)
		}
		if counts.Supported != 1 || counts.Partial != 1 || counts.Unsupported != 1 || counts.NotApplicable != 1 || counts.Unassessed != 1 {
			t.Errorf("%s state counts changed: %+v", name, counts)
		}
	}
}

func TestDecodeCapabilityAssetsRequiresEveryKindPerLanguage(t *testing.T) {
	ledgerJSON, corpusJSON := validCapabilityAssets(t)
	if _, _, err := decodeCapabilityAssets(ledgerJSON, corpusJSON); err != nil {
		t.Fatalf("valid capability assets failed: %v", err)
	}

	for _, language := range []string{"spl", "spl2"} {
		t.Run(language, func(t *testing.T) {
			var ledger capabilityLedgerFile
			mustUnmarshal(t, ledgerJSON, &ledger)
			for i, record := range ledger.Records {
				if record.Language == language && record.Kind == "command" {
					ledger.Records = append(ledger.Records[:i], ledger.Records[i+1:]...)
					break
				}
			}
			if _, _, err := decodeCapabilityAssets(mustJSON(t, ledger), corpusJSON); err == nil {
				t.Fatalf("ledger missing the command kind for %s was accepted", language)
			}
		})
	}

	t.Run("empty bundle", func(t *testing.T) {
		ledger := mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: []CapabilityRecord{}})
		corpus := mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: []CapabilityEvidence{}})
		if _, _, err := decodeCapabilityAssets(ledger, corpus); err == nil {
			t.Fatal("empty capability bundle was accepted")
		}
	})

	for _, language := range []string{"spl", "spl2"} {
		t.Run(language+" only", func(t *testing.T) {
			var fullLedger capabilityLedgerFile
			mustUnmarshal(t, ledgerJSON, &fullLedger)
			var records []CapabilityRecord
			for _, record := range fullLedger.Records {
				if record.Language == language {
					records = append(records, record)
				}
			}
			var fullCorpus capabilityCorpusFile
			mustUnmarshal(t, corpusJSON, &fullCorpus)
			var cases []CapabilityEvidence
			for _, evidence := range fullCorpus.Cases {
				if evidence.Document.Language == language {
					cases = append(cases, evidence)
				}
			}
			ledger := mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: records})
			corpus := mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: cases})
			if _, _, err := decodeCapabilityAssets(ledger, corpus); err == nil {
				t.Fatalf("%s-only capability bundle was accepted", language)
			}
		})
	}
}

func TestCapabilityDataCanonicalOrder(t *testing.T) {
	recordOrder := []struct {
		name        string
		left, right CapabilityRecord
	}{
		{
			name:  "language precedes every lower key",
			left:  CapabilityRecord{Language: "spl", Profile: "z", Kind: "z", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl2", Profile: "a", Kind: "a", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "profile precedes kind name form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "a", Kind: "z", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "b", Kind: "a", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "kind precedes name form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "z", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "function", Name: "a", Form: "a", ID: "a"},
		},
		{
			name:  "name precedes form and ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "alpha", Form: "z", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "beta", Form: "a", ID: "a"},
		},
		{
			name:  "form precedes ID",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "alpha", ID: "z"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "beta", ID: "a"},
		},
		{
			name:  "ID is the final key",
			left:  CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "basic", ID: "a"},
			right: CapabilityRecord{Language: "spl", Profile: "splunkd", Kind: "command", Name: "search", Form: "basic", ID: "b"},
		},
	}
	for _, tc := range recordOrder {
		t.Run(tc.name, func(t *testing.T) {
			if compareCapabilityRecords(tc.left, tc.right) >= 0 || compareCapabilityRecords(tc.right, tc.left) <= 0 {
				t.Fatalf("record comparison ignored %s", tc.name)
			}
		})
	}

	evidenceOrder := []struct {
		name        string
		left, right CapabilityEvidence
	}{
		{
			name:  "document language precedes profile and ID",
			left:  CapabilityEvidence{ID: "z", Document: QueryDocument{Language: "spl", Profile: "z"}},
			right: CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl2", Profile: "a"}},
		},
		{
			name:  "document profile precedes ID",
			left:  CapabilityEvidence{ID: "z", Document: QueryDocument{Language: "spl", Profile: "a"}},
			right: CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl", Profile: "b"}},
		},
		{
			name:  "evidence ID is the final key",
			left:  CapabilityEvidence{ID: "a", Document: QueryDocument{Language: "spl", Profile: "splunkd"}},
			right: CapabilityEvidence{ID: "b", Document: QueryDocument{Language: "spl", Profile: "splunkd"}},
		},
	}
	for _, tc := range evidenceOrder {
		t.Run(tc.name, func(t *testing.T) {
			if compareCapabilityEvidence(tc.left, tc.right) >= 0 || compareCapabilityEvidence(tc.right, tc.left) <= 0 {
				t.Fatalf("evidence comparison ignored %s", tc.name)
			}
		})
	}

	ledgerJSON, corpusJSON := validCapabilityAssets(t)
	if _, _, err := decodeCapabilityAssets(ledgerJSON, corpusJSON); err != nil {
		t.Fatalf("canonical capability data failed: %v", err)
	}
	var ledger capabilityLedgerFile
	mustUnmarshal(t, ledgerJSON, &ledger)
	ledger.Records[0], ledger.Records[1] = ledger.Records[1], ledger.Records[0]
	if _, _, err := decodeCapabilityAssets(mustJSON(t, ledger), corpusJSON); err == nil {
		t.Fatal("out-of-order capability records were accepted")
	}
	var corpus capabilityCorpusFile
	mustUnmarshal(t, corpusJSON, &corpus)
	corpus.Cases[0], corpus.Cases[1] = corpus.Cases[1], corpus.Cases[0]
	if _, _, err := decodeCapabilityAssets(ledgerJSON, mustJSON(t, corpus)); err == nil {
		t.Fatal("out-of-order capability evidence was accepted")
	}
}

func TestCapabilityDataClonesAreDeep(t *testing.T) {
	record := validCapabilityRecord()
	record.GrammarRegistered = true
	record.Dimensions.Syntax.Limitations = []string{}
	recordClone := cloneCapabilityRecord(record)
	if !recordClone.GrammarRegistered {
		t.Fatal("record clone dropped grammar registration")
	}
	if recordClone.Dimensions.Syntax.Limitations == nil {
		t.Fatal("record clone changed an authored empty slice to nil")
	}
	recordClone.Dimensions.Syntax.EvidenceIDs[0] = "mutated"
	if record.Dimensions.Syntax.EvidenceIDs[0] == "mutated" {
		t.Fatal("record clone aliases evidence IDs")
	}

	evidence := validCapabilityEvidence("positive", CapabilityEvidencePositive)
	evidence.Observations.SafeRewriting.ChangeReasons = []string{}
	evidenceClone := cloneCapabilityEvidence(evidence)
	if evidenceClone.Observations.SafeRewriting.ChangeReasons == nil {
		t.Fatal("evidence clone changed an authored empty slice to nil")
	}
	evidenceClone.Observations.Semantics.Stages[0].Command = "mutated"
	evidenceClone.Observations.Requirements.Items[0].Identity = "mutated"
	evidenceClone.RewriteRequest[0] = '['
	if evidence.Observations.Semantics.Stages[0].Command == "mutated" ||
		evidence.Observations.Requirements.Items[0].Identity == "mutated" ||
		evidence.RewriteRequest[0] == '[' {
		t.Fatal("evidence clone aliases nested authored data")
	}
}

func validCapabilityAssets(t *testing.T) ([]byte, []byte) {
	t.Helper()
	return mustJSON(t, capabilityLedgerFile{SchemaVersion: 1, Records: validCapabilityRecords()}),
		mustJSON(t, capabilityCorpusFile{SchemaVersion: 1, Cases: []CapabilityEvidence{
			validCapabilityEvidenceForLanguage("positive-spl", CapabilityEvidencePositive, "spl"),
			validCapabilityEvidenceForLanguage("positive-spl2", CapabilityEvidencePositive, "spl2"),
		}})
}

func validCapabilityRecords() []CapabilityRecord {
	kinds := []string{
		"annotation",
		"command",
		"dataset",
		"expression",
		"function",
		"lexical_form",
		"macro",
		"module",
		"namespace",
		"pipeline",
		"profile_form",
		"subsearch",
		"variable",
	}
	records := make([]CapabilityRecord, 0, len(kinds)*2)
	for _, language := range []string{"spl", "spl2"} {
		for _, kind := range kinds {
			record := validCapabilityRecord()
			record.ID = language + "-" + kind
			record.Language = language
			record.Kind = kind
			record.Name = kind
			record.Form = "default"
			record.Dimensions = claimsReferencing("positive-" + language)
			records = append(records, record)
		}
	}
	return records
}

func validCapabilityRecord() CapabilityRecord {
	return CapabilityRecord{
		ID:       "record",
		Language: "spl",
		Profile:  "splunkd",
		Kind:     "command",
		Name:     "search",
		Form:     "basic",
		Provenance: CapabilityProvenance{
			SourceFamily: "toolkit",
			Reference:    "pkg/analysis",
			Note:         "exercised by the analysis corpus",
		},
		Dimensions: claimsReferencing("positive"),
	}
}

func claimsReferencing(id string) CapabilityDimensions {
	claim := CapabilityClaim{State: CapabilitySupported, EvidenceIDs: []string{id}}
	return CapabilityDimensions{
		Syntax:        claim,
		Semantics:     claim,
		Requirements:  claim,
		Linting:       claim,
		SafeRewriting: claim,
	}
}

func validCapabilityEvidence(id string, classification CapabilityEvidenceClassification) CapabilityEvidence {
	return validCapabilityEvidenceForLanguage(id, classification, "spl")
}

func validCapabilityEvidenceForLanguage(id string, classification CapabilityEvidenceClassification, language string) CapabilityEvidence {
	diagnostic := CapabilityDiagnosticExpectation{
		Code:     "TEST",
		Category: "test",
		Severity: "warning",
		Location: Location{
			Start: Position{Offset: 0, Line: 1, Column: 1},
			End:   Position{Offset: 6, Line: 1, Column: 7},
		},
	}
	gapCodes := []string{"TEST"}
	if classification == CapabilityEvidencePositive {
		gapCodes = []string{}
	}
	return CapabilityEvidence{
		ID:             id,
		Classification: classification,
		Document: QueryDocument{
			Text:     "search index=main",
			Language: language,
			Profile:  "splunkd",
			Version:  "current",
			SourceID: "capability:" + id,
		},
		Observations: CapabilityEvidenceObservations{
			Syntax: &CapabilitySyntaxObservation{Complete: classification == CapabilityEvidencePositive, Diagnostics: []CapabilityDiagnosticExpectation{diagnostic}},
			Semantics: &CapabilitySemanticsObservation{
				Status:       classificationStatus(classification),
				Complete:     classification == CapabilityEvidencePositive,
				Stages:       []CapabilityStageExpectation{{Command: "search", SemanticComplete: classification == CapabilityEvidencePositive}},
				References:   []CapabilityReferenceExpectation{{NormalizedName: "main", Kind: "index", Role: "read", Resolution: "exact", Binding: "source", Location: diagnostic.Location}},
				Dependencies: []CapabilityDependencyExpectation{{Kind: "index", Name: "main"}},
				Transitions:  []CapabilityTransitionExpectation{{Operation: "read", Output: "main"}},
				Diagnostics:  []CapabilityDiagnosticExpectation{diagnostic},
			},
			Requirements: &CapabilityRequirementsObservation{
				QueryStatus: classificationStatus(classification),
				Complete:    classification == CapabilityEvidencePositive,
				Items:       []CapabilityRequirementExpectation{{Kind: "index", Identity: "main", Role: "read", Necessity: "required", Resolution: "exact"}},
				GapCodes:    gapCodes,
			},
			Linting:       &CapabilityLintingObservation{Diagnostics: []CapabilityDiagnosticExpectation{diagnostic}},
			SafeRewriting: &CapabilityRewriteObservation{Status: classificationStatus(classification), Committed: true, RewriteComplete: classification == CapabilityEvidencePositive, Text: "search index=main", CandidateText: "search index=main", CoverageReasons: []string{"TEST"}, ChangeReasons: []string{"TEST"}, RuleEvaluationReasons: []string{"TEST"}},
		},
		RewriteRequest: json.RawMessage(`{"schema_version":1}`),
		Provenance: CapabilityProvenance{
			SourceFamily: "toolkit",
			Reference:    "pkg/analysis",
			Note:         "expected observation",
		},
	}
}

func classificationStatus(classification CapabilityEvidenceClassification) Status {
	if classification == CapabilityEvidencePositive {
		return Valid
	}
	if classification == CapabilityEvidenceIncomplete {
		return Incomplete
	}
	return Invalid
}

func cloneEvidenceMap(source map[string]CapabilityEvidence) map[string]CapabilityEvidence {
	cloned := make(map[string]CapabilityEvidence, len(source))
	for id, evidence := range source {
		cloned[id] = cloneCapabilityEvidence(evidence)
	}
	return cloned
}

func mutateRecord(t *testing.T, ledger, corpus []byte, mutate func(*CapabilityRecord)) ([]byte, []byte) {
	t.Helper()
	var file capabilityLedgerFile
	mustUnmarshal(t, ledger, &file)
	mutate(&file.Records[0])
	return mustJSON(t, file), corpus
}

func mutateCapabilityEvidence(t *testing.T, corpus []byte, mutate func(*CapabilityEvidence)) []byte {
	t.Helper()
	var file capabilityCorpusFile
	mustUnmarshal(t, corpus, &file)
	mutate(&file.Cases[0])
	return mustJSON(t, file)
}

func replaceJSON(t *testing.T, raw []byte, key string, value any, additions ...any) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	object[key] = value
	for i := 0; i < len(additions); i += 2 {
		object[additions[i].(string)] = additions[i+1]
	}
	return mustJSON(t, object)
}

func deleteNestedJSONField(t *testing.T, raw []byte, arrayKey string, index int, field string) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	record := object[arrayKey].([]any)[index].(map[string]any)
	delete(record, field)
	return mustJSON(t, object)
}

func deleteRecordDimension(t *testing.T, raw []byte, dimension string) []byte {
	t.Helper()
	var object map[string]any
	mustUnmarshal(t, raw, &object)
	record := object["records"].([]any)[0].(map[string]any)
	dimensions := record["dimensions"].(map[string]any)
	delete(dimensions, dimension)
	return mustJSON(t, object)
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func mustUnmarshal(t *testing.T, raw []byte, value any) {
	t.Helper()
	if err := json.Unmarshal(raw, value); err != nil {
		t.Fatal(err)
	}
}
