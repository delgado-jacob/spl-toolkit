package analysis

import (
	"encoding/json"
	"fmt"
	"testing"
)

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
	record.Dimensions.Syntax.Limitations = []string{}
	recordClone := cloneCapabilityRecord(record)
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
	evidenceClone.Observations.Requirements.GapCodes[0] = "mutated"
	evidenceClone.RewriteRequest[0] = '['
	if evidence.Observations.Semantics.Stages[0].Command == "mutated" ||
		evidence.Observations.Requirements.GapCodes[0] == "mutated" ||
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
				References:   []CapabilityReferenceExpectation{{NormalizedName: "main", Kind: "index", Role: "read", Resolution: "exact", Binding: "source"}},
				Dependencies: []CapabilityDependencyExpectation{{Kind: "index", Name: "main"}},
				Transitions:  []CapabilityTransitionExpectation{{Operation: "read", Output: "main"}},
				Diagnostics:  []CapabilityDiagnosticExpectation{diagnostic},
			},
			Requirements: &CapabilityRequirementsObservation{
				QueryStatus: classificationStatus(classification),
				Complete:    classification == CapabilityEvidencePositive,
				Items:       []CapabilityRequirementExpectation{{Kind: "index", Identity: "main", Role: "read", Necessity: "required", Resolution: "exact"}},
				GapCodes:    []string{"TEST"},
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
