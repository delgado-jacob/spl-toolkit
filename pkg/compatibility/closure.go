package compatibility

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"slices"
	"strconv"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
)

func expansionKind(kind string) bool {
	return kind == "macro" || kind == "saved_search" || kind == "event_type" || kind == "module" || kind == "function"
}

type assessmentQuery struct {
	set              analysis.RequirementSet
	objectID, edgeID string
	provenance       []closure.ProvenanceSegment
}

func (p *Prepared) assessClosure(request AssessmentRequest) (*Report, error) {
	if err := validateDirectBindings(request); err != nil {
		return nil, err
	}
	if request.Document == nil {
		if len(request.DependencyBindings) != 0 {
			return nil, requestErrorAt("request_invalid", "/dependency_bindings", "dependency bindings require the original document")
		}
		resolved, err := p.resolveInputs(request.Requirements.Inputs, request.QueryScope, request.InputBindings)
		if err != nil {
			return nil, err
		}
		report := p.assess(request, resolved, assessmentQuery{set: request.Requirements}, nil)
		p.selectedBodyLimits(report, resolved, nil)
		finalizeReport(report)
		return report, nil
	}
	fresh, err := analysis.Requirements(*request.Document)
	if err != nil {
		return nil, err
	}
	if fresh.Query != request.Requirements.Query {
		return nil, requestErrorAt("requirements_stale", "/requirements/query", "the supplied requirements do not identify the original document and dialect")
	}
	if !reflect.DeepEqual(*fresh, request.Requirements) {
		return nil, requestErrorAt("requirements_inconsistent", "/requirements", "refresh the complete direct requirements from the original document")
	}
	bundle, err := p.env.DefinitionBundle(request.QueryScope)
	if err != nil {
		return nil, err
	}
	// Literal source selections can use the existing exact knowledge-reference
	// binding contract. Placeholder/descriptor selections never become aliases.
	bindings, err := sourceDependencyBindings(request, []assessmentQuery{{set: request.Requirements}}, request.DependencyBindings, p)
	if err != nil {
		return nil, err
	}
	evaluated, err := closure.Evaluate(closure.Request{SchemaVersion: 1, Document: *request.Document, Bundle: bundle, Bindings: bindings})
	if err != nil {
		return nil, closureRequestError(err)
	}
	queries := closureQueries(evaluated)
	// Each pass adds only exact selections already supplied by the caller.
	// Follow newly reached bodies until no further selection can refine traversal.
	for {
		discoveredBindings, err := sourceDependencyBindings(request, queries, bindings, p)
		if err != nil {
			return nil, err
		}
		if reflect.DeepEqual(bindings, discoveredBindings) {
			break
		}
		bindings = discoveredBindings
		evaluated, err = closure.Evaluate(closure.Request{SchemaVersion: 1, Document: *request.Document, Bundle: bundle, Bindings: bindings})
		if err != nil {
			return nil, closureRequestError(err)
		}
		queries = closureQueries(evaluated)
	}
	discovery := append([]assessmentQuery{{set: request.Requirements}}, queries...)
	// Contextual effective body requirements establish discoverable hidden
	// inputs. Uninstantiated macro formal tokens remain direct evidence only.
	inputs := unionInputs(discovery)
	resolved, err := p.resolveInputs(inputs, request.QueryScope, request.InputBindings)
	if err != nil {
		return nil, err
	}
	active := request
	active.Requirements = queries[0].set
	report := p.assess(active, queryInputs(resolved, queries[0].set.Inputs), queries[0], evaluated)
	if request.Requirements.QueryStatus == analysis.Invalid && report.Outcome != "not assessed" {
		report.Outcome = "not assessed"
		report.Reasons = append(report.Reasons, newReason("unsupported_semantics", "The original query is invalid; retained effective and hidden facts cannot establish its compatibility.", "query_semantics"))
	}
	report.Requirements = detach(request.Requirements)
	effective := detach(queries[0].set)
	report.EffectiveRequirements = &effective
	report.Closure = evaluated // Fresh call-owned analyses contain a private rewrite session.
	report.DependencyBindings = detach(request.DependencyBindings)
	report.Provenance.QueryDigest = request.Requirements.Query.QueryDigest
	report.Provenance.SourceID = request.Requirements.Query.SourceID
	report.Provenance.CapabilityRevision = request.Requirements.CapabilityRevision
	attachOutcomeProvenance(report, queries[0], evaluated)
	for _, query := range queries[1:] {
		active.Requirements = query.set
		fragment := p.assess(active, queryInputs(resolved, query.set.Inputs), query, evaluated)
		attachOutcomeProvenance(fragment, query, evaluated)
		mergeAssessment(report, fragment)
	}
	// Expanded-away dependency requirements retain their direct existence fact,
	// but obsolete direct field/source gaps do not enter effective aggregation.
	for _, item := range request.Requirements.Items {
		if !expansionKind(item.Kind) {
			continue
		}
		retained := false
		for _, effectiveItem := range effective.Items {
			if item.Kind == effectiveItem.Kind && item.Identity == effectiveItem.Identity {
				retained = true
				break
			}
		}
		if retained {
			continue
		}
		out := p.selectedObjectOutcome(p.assessObject(item, request.Requirements.Query, request.QueryScope), item, assessmentQuery{set: request.Requirements}, evaluated)
		for _, occurrence := range item.Occurrences {
			out.SourceIntervals = append(out.SourceIntervals, closure.SourceInterval{Kind: "query", SourceID: request.Document.SourceID, Start: occurrence.Location.Start.Offset, End: occurrence.Location.End.Offset})
		}
		report.RequirementOutcomes = append(report.RequirementOutcomes, out)
		aggregateOutcome(report, out.Outcome, out.Applicability, out.Reasons)
	}
	attachInputOccurrences(report, discovery, evaluated)
	relevant := p.reachableExpansion(evaluated)
	reasons := []Reason{}
	for _, gap := range evaluated.Gaps {
		// Query semantics are assessed through the active requirement sets. Raw
		// closure analysis coverage also includes output/correlation uncertainties.
		if gap.Code == "analysis_incomplete" || gap.Code == "collection_incomplete" && resolvedGap(gap, evaluated) {
			continue
		}
		relevant = true
		reason := newReason("dependency_closure_incomplete", "Reachable dependency closure is incomplete: "+gap.Code+".", "dependency_closure")
		reason.ObjectID = gap.Source.ObjectID
		reasons = append(reasons, reason)
	}
	replaceClosureCoverage(report, relevant, reasons)
	p.selectedBodyLimits(report, resolved, evaluated)
	finalizeReport(report)
	return report, nil
}

func closureQueries(report *closure.Report) []assessmentQuery {
	out := []assessmentQuery{{set: report.EffectiveAnalysis.Requirements, provenance: report.Provenance}}
	for _, context := range report.DefinitionContexts {
		for _, def := range report.DefinitionAnalyses {
			if context.ObjectID != def.ObjectID {
				continue
			}
			result := def.EffectiveAnalysis
			if result == nil {
				result = def.DirectAnalysis
			}
			if result != nil {
				out = append(out, assessmentQuery{set: result.Requirements, objectID: def.ObjectID, edgeID: context.EdgeID, provenance: context.Provenance})
			}
		}
	}
	return out
}

func unionInputs(queries []assessmentQuery) []analysis.QueryInput {
	out := []analysis.QueryInput{}
	index := map[string]int{}
	for _, query := range queries {
		for _, input := range query.set.Inputs {
			if i, found := index[input.ID]; found {
				for _, occurrence := range input.Occurrences {
					if !slices.ContainsFunc(out[i].Occurrences, func(prior analysis.InputOccurrence) bool { return reflect.DeepEqual(prior, occurrence) }) {
						out[i].Occurrences = append(out[i].Occurrences, detach(occurrence))
					}
				}
			} else {
				index[input.ID] = len(out)
				out = append(out, detach(input))
			}
		}
	}
	return out
}

func aggregateOutcome(report *Report, outcome, applicability string, reasons []Reason) {
	if applicability == "inapplicable" || outcome == "satisfied" {
		return
	}
	report.Reasons = append(report.Reasons, reasons...)
	// An invalid or unassessable root retains priority while hidden evidence is
	// recorded. Configuration admission has already completed before assessment.
	if report.Outcome == "not assessed" {
		return
	}
	if outcome == "missing" && applicability == "applicable" {
		report.Outcome = "unsatisfied"
		return
	}
	if report.Outcome != "unsatisfied" {
		report.Outcome = "incomplete"
	}
}
func mergeAssessment(report, fragment *Report) {
	report.RequirementOutcomes = append(report.RequirementOutcomes, fragment.RequirementOutcomes...)
	report.Coverage = append(report.Coverage, fragment.Coverage...)
	report.Reasons = append(report.Reasons, fragment.Reasons...)
	if report.Outcome != "not assessed" {
		if fragment.Outcome == "unsatisfied" {
			report.Outcome = "unsatisfied"
		} else if fragment.Outcome != "satisfied" && report.Outcome != "unsatisfied" {
			report.Outcome = "incomplete"
		}
	}
	for _, input := range fragment.Inputs {
		i := slices.IndexFunc(report.Inputs, func(prior InputOutcome) bool { return prior.InputID == input.InputID })
		if i < 0 {
			report.Inputs = append(report.Inputs, input)
			continue
		}
		existing := &report.Inputs[i]
		existing.RequirementIDs = append(existing.RequirementIDs, input.RequirementIDs...)
		existing.Reasons = append(existing.Reasons, input.Reasons...)
		if input.Outcome == "missing" || existing.Outcome == "satisfied" {
			existing.Outcome = input.Outcome
		} else if input.Outcome == "ambiguous" && existing.Outcome != "missing" {
			existing.Outcome = input.Outcome
		}
		for _, object := range input.Objects {
			if !slices.ContainsFunc(existing.Objects, func(prior ObjectEvidence) bool { return reflect.DeepEqual(prior, object) }) {
				existing.Objects = append(existing.Objects, object)
			}
		}
	}
}

func replaceClosureCoverage(report *Report, relevant bool, reasons []Reason) {
	report.Coverage = slices.DeleteFunc(report.Coverage, func(c Coverage) bool { return c.Dimension == "dependency_closure" })
	state := "not_applicable"
	if relevant {
		state = "complete"
	}
	if len(reasons) > 0 {
		state = "partial"
	}
	report.Coverage = append(report.Coverage, Coverage{Dimension: "dependency_closure", State: state, Reasons: reasons})
	if len(reasons) > 0 {
		aggregateOutcome(report, "indeterminate", "applicable", reasons)
	}
}

func (p *Prepared) selectedBodyLimits(report *Report, resolved map[string]resolvedInput, evaluated *closure.Report) {
	for _, input := range resolved {
		// Only active inputs carry current obligations. Obsolete direct implicit
		// streams remain discovery evidence, never extra assessed requirements.
		if !slices.ContainsFunc(report.Inputs, func(out InputOutcome) bool { return out.InputID == input.input.ID }) {
			continue
		}
		for _, object := range input.objects {
			if object.Kind != "dataset" || object.Document == nil && len(object.Relations) == 0 {
				continue
			}
			reached := false
			if evaluated != nil && input.input.Kind == "explicit_dataset" {
				for _, edge := range evaluated.Traversal {
					if edge.ToObjectID == object.ID && edge.Kind == "dataset" && (edge.Resolution == "resolved" || edge.Resolution == "bound") {
						reached = true
					}
				}
			}
			if reached {
				continue
			}
			reason := newReason("dependency_closure_incomplete", "The selected dataset has a body or relations, but supported closure evidence has not established its source invocation context.", "dependency_closure")
			reason.InputID = input.input.ID
			reason.ObjectID = object.ID
			aggregateOutcome(report, "indeterminate", "applicable", []Reason{reason})
			for i := range report.Coverage {
				if report.Coverage[i].Dimension == "dependency_closure" {
					report.Coverage[i].State = "partial"
					report.Coverage[i].Reasons = append(report.Coverage[i].Reasons, reason)
				}
			}
		}
	}
	if evaluated == nil {
		for _, out := range report.RequirementOutcomes {
			if out.InputID != "" {
				continue
			}
			for _, evidence := range out.Objects {
				object := evidence.Object
				if object == nil || expansionKind(object.Kind) || object.Document == nil && len(object.Relations) == 0 {
					continue
				}
				reason := newReason("dependency_closure_incomplete", "The captured object has hidden dependency evidence that requires the original document and supported traversal.", "dependency_closure")
				reason.ObjectID = object.ID
				reason.RequirementID = out.RequirementID
				aggregateOutcome(report, "indeterminate", "applicable", []Reason{reason})
				for i := range report.Coverage {
					if report.Coverage[i].Dimension == "dependency_closure" {
						report.Coverage[i].State = "partial"
						report.Coverage[i].Reasons = append(report.Coverage[i].Reasons, reason)
					}
				}
			}
		}
	}

}

// Source bindings can select an exact literal named reference. Contradictory
// dependency selections at that same original reference are request errors.
func sourceDependencyBindings(request AssessmentRequest, queries []assessmentQuery, bindings []closure.Binding, p *Prepared) ([]closure.Binding, error) {
	out := append([]closure.Binding{}, bindings...)
	for _, query := range queries {
		for _, input := range query.set.Inputs {
			if input.Kind != "explicit_dataset" || input.Identity.Form != "identifier" && input.Identity.Form != "dotted" {
				continue
			}
			for bindingIndex, binding := range request.InputBindings {
				if binding.InputID != input.ID {
					continue
				}
				// Validate selected identity before asking closure to inspect its body.
				if binding.Expected.Kind != "dataset" || binding.Expected.Name != input.Identity.Value || !identityInScope(request.QueryScope, binding.Expected) {
					return nil, requestErrorAt("binding_invalid", "/input_bindings/"+strconv.Itoa(bindingIndex)+"/expected", "binding is not consistent with the exact query identity and scope")
				}
				object, found := p.env.Object(binding.ObjectID)
				if !found {
					continue
				}
				if objectIdentity(object) != binding.Expected {
					return nil, requestErrorAt("binding_invalid", "/input_bindings/"+strconv.Itoa(bindingIndex)+"/expected", "captured object disagrees with expected identity")
				}
				for _, occurrence := range input.Occurrences {
					sources, _ := mappedProvenance(query, occurrence.Location.Start.Offset, occurrence.Location.End.Offset, nil)
					if len(sources) != 1 {
						continue
					}
					source := sources[0]
					document := request.Document
					if source.Kind == "definition" {
						definition, ok := p.env.Object(source.ObjectID)
						if !ok {
							continue
						}
						document = definition.Document
					}
					if document == nil || source.Start < 0 || source.End > len(document.Text) || source.End <= source.Start {
						continue
					}
					// The owner validates range/name admission against the original document.
					candidate := closure.Binding{DocumentDigest: queryDigest(document.Text), Kind: "dataset", Start: source.Start, End: source.End, ObjectID: binding.ObjectID}
					exists := false
					for _, prior := range out {
						if prior.DocumentDigest == candidate.DocumentDigest && prior.Kind == candidate.Kind && prior.Start == candidate.Start && prior.End == candidate.End {
							if prior.ObjectID != candidate.ObjectID {
								return nil, requestErrorAt("binding_invalid", "/dependency_bindings", "source and dependency bindings select different objects at one reference")
							}
							exists = true
						}
					}
					if !exists {
						out = append(out, candidate)
					}
				}
			}
		}
	}
	return out, nil
}

func queryDigest(text string) string {
	// The producer owns the UTF-8 digest format; caller document admission already
	// normalizes dialect. The same digest is used by dependency binding admission.
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func mappedProvenance(query assessmentQuery, start, end int, evaluated *closure.Report) ([]closure.SourceInterval, []closure.InvocationFrame) {
	sources := []closure.SourceInterval{}
	frames := []closure.InvocationFrame{}
	addSource := func(source closure.SourceInterval) {
		if !slices.Contains(sources, source) {
			sources = append(sources, source)
		}
	}
	addFrame := func(frame closure.InvocationFrame) {
		if !slices.ContainsFunc(frames, func(prior closure.InvocationFrame) bool { return reflect.DeepEqual(prior, frame) }) {
			frames = append(frames, frame)
		}
	}
	if query.edgeID != "" && evaluated != nil {
		for _, frame := range traversalFrames(evaluated, query.edgeID) {
			addFrame(frame)
		}
	}
	for _, segment := range query.provenance {
		lo, hi := max(start, segment.EffectiveStart), min(end, segment.EffectiveEnd)
		if lo >= hi {
			continue
		}
		source := segment.Source
		source.Start += lo - segment.EffectiveStart
		source.End -= segment.EffectiveEnd - hi
		addSource(source)
		for _, placeholder := range segment.Placeholders {
			addSource(placeholder)
		}
		if len(segment.Placeholders) == 0 && segment.Placeholder != nil {
			addSource(*segment.Placeholder)
		}
		for _, frame := range segment.InvocationChain {
			addFrame(frame)
		}
	}
	if len(query.provenance) == 0 && query.objectID == "" {
		addSource(closure.SourceInterval{Kind: "query", SourceID: query.set.Query.SourceID, Start: start, End: end})
	}
	return sources, frames
}
func attachOutcomeProvenance(report *Report, query assessmentQuery, evaluated *closure.Report) {
	for i := range report.RequirementOutcomes {
		out := &report.RequirementOutcomes[i]
		out.DefinitionObjectID = query.objectID
		out.TraversalEdgeID = query.edgeID
		for _, item := range query.set.Items {
			if out.RequirementID == item.ID {
				for _, occurrence := range item.Occurrences {
					sources, frames := mappedProvenance(query, occurrence.Location.Start.Offset, occurrence.Location.End.Offset, evaluated)
					for _, source := range sources {
						if !slices.Contains(out.SourceIntervals, source) {
							out.SourceIntervals = append(out.SourceIntervals, source)
						}
					}
					for _, frame := range frames {
						if !slices.ContainsFunc(out.InvocationProvenance, func(prior closure.InvocationFrame) bool { return reflect.DeepEqual(prior, frame) }) {
							out.InvocationProvenance = append(out.InvocationProvenance, frame)
						}
					}
				}
			}
		}
	}
}
func attachInputOccurrences(report *Report, queries []assessmentQuery, evaluated *closure.Report) {
	for _, query := range queries {
		for _, input := range query.set.Inputs {
			index := slices.IndexFunc(report.Inputs, func(out InputOutcome) bool { return out.InputID == input.ID })
			if index < 0 {
				continue
			}
			for _, occurrence := range input.Occurrences {
				sources, frames := mappedProvenance(query, occurrence.Location.Start.Offset, occurrence.Location.End.Offset, evaluated)
				evidence := InputOccurrenceEvidence{Query: query.set.Query, DefinitionObjectID: query.objectID, TraversalEdgeID: query.edgeID, Occurrence: detach(occurrence), SourceIntervals: sources, InvocationProvenance: frames}
				if !slices.ContainsFunc(report.Inputs[index].Occurrences, func(prior InputOccurrenceEvidence) bool { return reflect.DeepEqual(prior, evidence) }) {
					report.Inputs[index].Occurrences = append(report.Inputs[index].Occurrences, evidence)
				}
			}
		}
	}
	// Local requirement IDs are qualified in the input's cross-query index. The
	// authoritative outcomes retain the structured query/object/edge identity.
	for i := range report.Inputs {
		report.Inputs[i].RequirementIDs = []string{}
		for _, out := range report.RequirementOutcomes {
			if out.InputID == report.Inputs[i].InputID || slices.ContainsFunc(out.Reasons, func(reason Reason) bool { return slices.Contains(reason.CandidateInputIDs, report.Inputs[i].InputID) }) {
				report.Inputs[i].RequirementIDs = append(report.Inputs[i].RequirementIDs, stableKey(out.Query)+":"+out.DefinitionObjectID+":"+out.TraversalEdgeID+":"+out.RequirementID)
			}
		}
	}
}

func queryInputs(resolved map[string]resolvedInput, inputs []analysis.QueryInput) map[string]resolvedInput {
	out := make(map[string]resolvedInput, len(inputs))
	for _, input := range inputs {
		value := resolved[input.ID]
		value.input = input
		out[input.ID] = value
	}
	return out
}
func traversalInvocation(edge closure.TraversalEdge) []closure.SourceInterval {
	if edge.Property != "" {
		return []closure.SourceInterval{}
	}
	return detach(edge.Origins)
}

func (p *Prepared) selectedObjectOutcome(out RequirementOutcome, item analysis.RequirementItem, query assessmentQuery, evaluated *closure.Report) RequirementOutcome {
	if evaluated == nil || item.InputID != "" || item.Resolution != "exact" {
		return out
	}
	objects := []ObjectEvidence{}
	for _, occurrence := range item.Occurrences {
		origins, _ := mappedProvenance(query, occurrence.Location.Start.Offset, occurrence.Location.End.Offset, evaluated)
		found := false
		for _, edge := range evaluated.Traversal {
			if edge.Kind != item.Kind || edge.Name != item.Identity || edge.ToObjectID == "" || edge.Resolution != "bound" && edge.Resolution != "resolved" {
				continue
			}
			if !slices.ContainsFunc(origins, func(source closure.SourceInterval) bool {
				return source == edge.Source || item.Kind == "macro" && source.Kind == edge.Source.Kind && source.SourceID == edge.Source.SourceID && source.ObjectID == edge.Source.ObjectID && edge.Source.Start <= source.Start && edge.Source.End >= source.End
			}) {
				continue
			}
			object, ok := p.env.Object(edge.ToObjectID)
			if !ok {
				continue
			}
			if !slices.ContainsFunc(objects, func(prior ObjectEvidence) bool { return prior.ObjectID == object.ID }) {
				objects = append(objects, ObjectEvidence{ObjectID: object.ID, Expected: objectIdentity(object), Object: &object})
			}
			found = true
		}
		if !found {
			return out
		}
	}
	if len(objects) > 0 {
		out.Outcome = "satisfied"
		out.Reasons = []Reason{}
		out.Objects = objects
	}
	return out
}

// Explicit relation edges carry object context without invented byte intervals.
// Canonical expansion chains supply invocation sites; named body traversal adds
// its actual parent edges in order where those frames are not already present.
func traversalFrames(report *closure.Report, edgeID string) []closure.InvocationFrame {
	index := slices.IndexFunc(report.Traversal, func(edge closure.TraversalEdge) bool { return edge.ID == edgeID })
	if index < 0 {
		return nil
	}
	edge := report.Traversal[index]
	out := []closure.InvocationFrame{}
	for i, objectID := range edge.Path {
		if slices.ContainsFunc(edge.InvocationChain, func(frame closure.InvocationFrame) bool { return frame.ObjectID == objectID }) {
			continue
		}
		for j := index - 1; j >= 0; j-- {
			parent := report.Traversal[j]
			if parent.ToObjectID == objectID && slices.Equal(parent.Path, edge.Path[:i]) {
				out = append(out, closure.InvocationFrame{ObjectID: objectID, InstanceID: parent.ID, Invocation: traversalInvocation(parent)})
				break
			}
		}
	}
	out = append(out, detach(edge.InvocationChain)...)
	out = append(out, closure.InvocationFrame{ObjectID: edge.ToObjectID, InstanceID: edge.ID, Invocation: traversalInvocation(edge)})
	return out
}

func (p *Prepared) reachableExpansion(report *closure.Report) bool {
	for _, edge := range report.Traversal {
		if expansionKind(edge.Kind) {
			return true
		}
		if object, ok := p.env.Object(edge.ToObjectID); ok && (object.Document != nil || len(object.Relations) != 0) {
			return true
		}
	}
	return false
}

// A captured exact target retains its positive evidence even when unrelated
// collection absence cannot be proved. Opaque, unresolved and cyclic paths keep
// their own independent gaps.
func resolvedGap(gap closure.ClosureGap, report *closure.Report) bool {
	matched := false
	for _, edge := range report.Traversal {
		if edge.Kind == gap.Kind && edge.Name == gap.Name && edge.Source == gap.Source {
			matched = true
			if edge.ToObjectID == "" || edge.Resolution != "resolved" && edge.Resolution != "bound" {
				return false
			}
		}
	}
	return matched
}
