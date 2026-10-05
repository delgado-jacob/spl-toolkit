package analysis

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// InputIdentity preserves the source spelling category independently of display names.
type InputIdentity struct {
	Form  string `json:"form"`
	Value string `json:"value"`
}

// InputReason locates a limit on a particular dimension of query evidence.
type InputReason struct {
	Code         string   `json:"code"`
	Message      string   `json:"message"`
	Location     Location `json:"location"`
	StageID      string   `json:"stage_id"`
	ScopeID      string   `json:"scope_id"`
	ReferenceIDs []string `json:"reference_ids"`
}

// InputCoverage uses complete only when discovery is exhaustive. A proved empty
// obligation set is not_applicable; recovered emptiness after failure is partial.
type InputCoverage struct {
	State   string        `json:"state"`
	Reasons []InputReason `json:"reasons"`
}

type InputOccurrence struct {
	ID                  string     `json:"id"`
	ReferenceID         string     `json:"reference_id"`
	OriginalReferenceID string     `json:"original_reference_id"`
	StageID             string     `json:"stage_id"`
	ScopeID             string     `json:"scope_id"`
	Alias               string     `json:"alias"`
	Location            Location   `json:"location"`
	UseSiteLocations    []Location `json:"use_site_locations"`
	UseSiteReferenceIDs []string   `json:"use_site_reference_ids"`
}

// QueryInput is a logical external source with distinct situated occurrences.
// Its ID excludes aliases and all environment/schema evidence.
type QueryInput struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Name        string            `json:"name"`
	Identity    InputIdentity     `json:"identity"`
	Evidence    InputCoverage     `json:"evidence"`
	Occurrences []InputOccurrence `json:"occurrences"`
}

type inputFact struct {
	kind, name string
	sourceID   string
	identity   InputIdentity
	occurrence InputOccurrence
}

func opaqueEvidenceID(prefix string, values ...string) string {
	hash := sha256.New()
	for _, value := range values {
		fmt.Fprintf(hash, "%d:", len(value))
		hash.Write([]byte(value))
	}
	return prefix + hex.EncodeToString(hash.Sum(nil))
}
func opaqueInputID(kind string, identity InputIdentity) string {
	return opaqueEvidenceID("input-", kind, identity.Form, identity.Value)
}
func inputOccurrenceID(fact inputFact) string {
	o := fact.occurrence
	values := []string{fact.sourceID, opaqueInputID(fact.kind, fact.identity), strconv.Itoa(o.Location.Start.Offset), strconv.Itoa(o.Location.End.Offset), o.ScopeID}
	for _, location := range o.UseSiteLocations {
		values = append(values, strconv.Itoa(location.Start.Offset), strconv.Itoa(location.End.Offset))
	}
	return opaqueEvidenceID("occurrence-", values...)
}
func cloneInputFacts(in []inputFact) []inputFact {
	out := append([]inputFact{}, in...)
	for i := range out {
		out[i].occurrence.UseSiteLocations = append([]Location{}, in[i].occurrence.UseSiteLocations...)
		out[i].occurrence.UseSiteReferenceIDs = copyIDs(in[i].occurrence.UseSiteReferenceIDs)
	}
	return out
}
func mergeInputFacts(base []inputFact, additions ...[]inputFact) []inputFact {
	out := cloneInputFacts(base)
	seen := map[string]bool{}
	for _, fact := range out {
		seen[inputOccurrenceID(fact)] = true
	}
	for _, facts := range additions {
		for _, fact := range cloneInputFacts(facts) {
			key := inputOccurrenceID(fact)
			if !seen[key] {
				out = append(out, fact)
				seen[key] = true
			}
		}
	}
	return out
}
func cloneInputCoverage(in InputCoverage) InputCoverage {
	out := in
	out.Reasons = append([]InputReason{}, in.Reasons...)
	for i := range out.Reasons {
		out.Reasons[i].ReferenceIDs = copyIDs(in.Reasons[i].ReferenceIDs)
	}
	return out
}
func cloneInputs(in []QueryInput) []QueryInput {
	out := append([]QueryInput{}, in...)
	for i := range out {
		out[i].Evidence = cloneInputCoverage(in[i].Evidence)
		out[i].Occurrences = append([]InputOccurrence{}, in[i].Occurrences...)
		for j := range out[i].Occurrences {
			o := &out[i].Occurrences[j]
			o.UseSiteLocations = append([]Location{}, o.UseSiteLocations...)
			o.UseSiteReferenceIDs = copyIDs(o.UseSiteReferenceIDs)
		}
	}
	return out
}
func (s *semanticStage) recordInput(kind, name, form, value, referenceID, alias string, location Location) {
	st := s.result.Stages[s.stage]
	if !s.env.requirements.trace.reserveSourceEvidence(2, st, location) {
		s.result.Stages[s.stage].SemanticComplete = false
		s.env.uncertain = true
		s.env.requirements.uncertain = true
		return
	}
	if kind == "unresolved_source" {
		value = orderedStringSliceKey([]string{s.result.Document.SourceID, value, strconv.Itoa(location.Start.Offset), strconv.Itoa(location.End.Offset)})
	}
	fact := inputFact{sourceID: s.result.Document.SourceID, kind: kind, name: name, identity: InputIdentity{Form: form, Value: value}, occurrence: InputOccurrence{ReferenceID: referenceID, OriginalReferenceID: referenceID, StageID: st.ID, ScopeID: st.ScopeID, Alias: alias, Location: location, UseSiteLocations: []Location{}, UseSiteReferenceIDs: []string{}}}
	s.env.inputs = mergeInputFacts(s.env.inputs, []inputFact{fact})
	s.env.requirements.inputs = mergeInputFacts(s.env.requirements.inputs, []inputFact{fact})
	if trace := s.env.requirements.trace; trace != nil {
		trace.inputs = mergeInputFacts(trace.inputs, []inputFact{fact})
		if referenceID != "" {
			trace.reference(referenceID).owners = []sourceOwner{{input: fact}}
		}
	}
}

func (s *semanticStage) implicitInput() {
	stage := s.result.Stages[s.stage]
	if stage.Position != 0 || len(s.env.inputs) > 0 {
		return
	}
	switch stage.Command {
	case "makeresults", "gentimes":
		return
	}
	if stage.Command == "from" && s.result.Document.Language == "spl2" {
		return
	}
	s.recordInput("implicit_stream", "", "implicit", "", "", "", stage.Location)
}

func projectInputs(trace *requirementTrace) ([]QueryInput, InputCoverage) {
	inputs := []QueryInput{}
	coverage := InputCoverage{State: "complete", Reasons: []InputReason{}}
	// Discovery can remain complete across field-only semantic gaps. Syntax,
	// dynamic sources, modules, and unknown commands may hide source obligations.
	for _, entry := range trace.diagnostics {
		d := entry.diagnostic
		hidesSources := d.Category == "syntax" || d.Category == "unsupported_syntax" || d.Code == CodeAnalysisResourceLimit || d.Code == CodeUnsupportedCommand || d.Code == CodeUnresolvedModule || d.Code == CodeDeclarationCycle || d.Code == CodeDynamicReference && strings.Contains(strings.ToLower(d.Message), "macro") || d.Code == CodeUnsupportedSemantics && strings.Contains(strings.ToLower(d.Message), "dataset")
		if hidesSources {
			coverage.State = "partial"
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: d.Code, Message: d.Message, Location: d.Location, StageID: d.StageID, ScopeID: d.ScopeID, ReferenceIDs: copyIDs(entry.pendingReferenceIDs)})
		}
	}
	if !trace.syntaxComplete {
		coverage.State = "partial"
		if len(coverage.Reasons) == 0 {
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: "target_discovery_incomplete", Message: "source discovery stopped before syntax was proved", ReferenceIDs: []string{}})
		}
	}
	facts := mergeInputFacts(nil, trace.inputs)
	// Index shorter use chains once instead of comparing every pair of facts.
	superseded := map[string]bool{}
	for _, fact := range facts {
		if fact.occurrence.OriginalReferenceID == "" {
			continue
		}
		for length := range fact.occurrence.UseSiteLocations {
			superseded[inputUsePrefixKey(fact.occurrence.OriginalReferenceID, fact.occurrence.UseSiteLocations[:length])] = true
		}
	}
	groups := map[string]int{}
	for _, fact := range facts {
		// A view declaration's source occurrence is replaced by each terminal use.
		if fact.occurrence.OriginalReferenceID != "" && superseded[inputUsePrefixKey(fact.occurrence.OriginalReferenceID, fact.occurrence.UseSiteLocations)] {
			continue
		}
		id := opaqueInputID(fact.kind, fact.identity)
		index, ok := groups[id]
		if !ok {
			index = len(inputs)
			groups[id] = index
			inputs = append(inputs, QueryInput{ID: id, Kind: fact.kind, Name: fact.name, Identity: fact.identity, Evidence: InputCoverage{State: "complete", Reasons: []InputReason{}}, Occurrences: []InputOccurrence{}})
		}
		fact.occurrence.ID = inputOccurrenceID(fact)
		inputs[index].Occurrences = append(inputs[index].Occurrences, fact.occurrence)
		if fact.kind == "unresolved_source" {
			reason := InputReason{Code: "target_discovery_incomplete", Message: "source identity is unresolved", Location: fact.occurrence.Location, StageID: fact.occurrence.StageID, ScopeID: fact.occurrence.ScopeID, ReferenceIDs: []string{}}
			if fact.occurrence.ReferenceID != "" {
				reason.ReferenceIDs = append(reason.ReferenceIDs, fact.occurrence.ReferenceID)
			}
			inputs[index].Evidence.State = "partial"
			inputs[index].Evidence.Reasons = append(inputs[index].Evidence.Reasons, reason)
			coverage.State = "partial"
			coverage.Reasons = append(coverage.Reasons, reason)
		}
	}
	for i := range inputs {
		sort.SliceStable(inputs[i].Occurrences, func(a, b int) bool {
			x, y := inputs[i].Occurrences[a], inputs[i].Occurrences[b]
			if x.Location.Start.Offset != y.Location.Start.Offset {
				return x.Location.Start.Offset < y.Location.Start.Offset
			}
			if len(x.UseSiteLocations) > 0 && len(y.UseSiteLocations) > 0 {
				a, b := x.UseSiteLocations[len(x.UseSiteLocations)-1], y.UseSiteLocations[len(y.UseSiteLocations)-1]
				if a.Start.Offset != b.Start.Offset {
					return a.Start.Offset < b.Start.Offset
				}
			}
			return x.ID < y.ID
		})
	}
	sort.SliceStable(inputs, func(i, j int) bool {
		a, b := inputs[i].Occurrences[0], inputs[j].Occurrences[0]
		if a.Location.Start.Offset != b.Location.Start.Offset {
			return a.Location.Start.Offset < b.Location.Start.Offset
		}
		return inputs[i].ID < inputs[j].ID
	})
	if len(inputs) == 0 && coverage.State == "complete" {
		coverage.State = "not_applicable"
	}
	return inputs, coverage
}
func inputUsePrefixKey(referenceID string, locations []Location) string {
	values := []string{referenceID}
	for _, location := range locations {
		values = append(values, strconv.Itoa(location.Start.Offset), strconv.Itoa(location.Start.Line), strconv.Itoa(location.Start.Column), strconv.Itoa(location.End.Offset), strconv.Itoa(location.End.Line), strconv.Itoa(location.End.Column))
	}
	return orderedStringSliceKey(values)
}

// situatedViewInputs reuses semantic source summaries without replaying bodies.
func situatedViewInputs(summary []inputFact, referenceID string, stage Stage, location Location, alias string) []inputFact {
	facts := cloneInputFacts(summary)
	for i := range facts {
		o := &facts[i].occurrence
		o.ReferenceID = referenceID
		o.Alias = alias
		o.StageID = stage.ID
		o.ScopeID = stage.ScopeID
		o.UseSiteLocations = append(o.UseSiteLocations, location)
		o.UseSiteReferenceIDs = append(o.UseSiteReferenceIDs, referenceID)
	}
	return facts
}

func remapInputReferences(facts []inputFact, mapping map[string]string) {
	for i := range facts {
		o := &facts[i].occurrence
		if o.ReferenceID != "" {
			o.ReferenceID = mapping[o.ReferenceID]
		}
		if o.OriginalReferenceID != "" {
			o.OriginalReferenceID = mapping[o.OriginalReferenceID]
		}
		remapTraceIDs(o.UseSiteReferenceIDs, mapping)
	}
}
func remapInputStages(facts []inputFact, mapping map[string]string) {
	for i := range facts {
		if id := facts[i].occurrence.StageID; id != "" {
			facts[i].occurrence.StageID = mapping[id]
		}
	}
}

func inputAttributionCoverage(trace *requirementTrace) InputCoverage {
	coverage := InputCoverage{State: "not_applicable", Reasons: []InputReason{}}
	for _, entry := range trace.references {
		if entry.reference.Kind == "field" && (entry.directExternal || entry.conditional || entry.pathConditional) {
			if coverage.State == "not_applicable" {
				coverage.State = "complete"
			}
			if _, _, proved := provedSourceOwner(entry.owners); proved && entry.reference.Resolution == "exact" {
				continue
			}
			coverage.State = "partial"
			r := entry.reference
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: "field_attribution_incomplete", Message: "source field ownership has not been proved", Location: r.Location, StageID: r.StageID, ScopeID: r.ScopeID, ReferenceIDs: []string{r.ID}})
		}
	}
	for _, entry := range trace.diagnostics {
		diagnostic := entry.diagnostic
		if diagnostic.Code == CodeAnalysisResourceLimit && diagnostic.Message == sourceEvidenceResourceLimitMessage {
			coverage.State = "partial"
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: diagnostic.Code, Message: diagnostic.Message, Location: diagnostic.Location, StageID: diagnostic.StageID, ScopeID: diagnostic.ScopeID, ReferenceIDs: copyIDs(entry.pendingReferenceIDs)})
		}
	}
	if coverage.State == "not_applicable" && (!trace.syntaxComplete || !trace.semanticComplete) {
		coverage.State = "partial"
		for _, entry := range trace.diagnostics {
			if !entry.incomplete {
				continue
			}
			diagnostic := entry.diagnostic
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: "field_attribution_incomplete", Message: diagnostic.Message, Location: diagnostic.Location, StageID: diagnostic.StageID, ScopeID: diagnostic.ScopeID, ReferenceIDs: copyIDs(entry.pendingReferenceIDs)})
		}
		if len(coverage.Reasons) == 0 {
			coverage.Reasons = append(coverage.Reasons, InputReason{Code: "field_attribution_incomplete", Message: "field discovery is incomplete", ReferenceIDs: []string{}})
		}
	}
	return coverage
}
