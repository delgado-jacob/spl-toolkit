package workflow

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/closure"
	"github.com/delgado-jacob/spl-toolkit/pkg/compatibility"
	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/impact"
	"github.com/delgado-jacob/spl-toolkit/pkg/resolution"
)

func CompareJSON(raw []byte) (*ComparisonReport, error) {
	if _, _, _, err := readWire(raw, reflect.TypeOf(CompareRequest{})); err != nil {
		return nil, err
	}
	var q CompareRequest
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, requestErrorAt("request_invalid", "", err.Error())
	}
	return Compare(q)
}

// Compare admits and detaches supplied reports without restoring private proof.
func Compare(q CompareRequest) (*ComparisonReport, error) {
	if !validUTF8(reflect.ValueOf(q)) {
		return nil, requestErrorAt("request_invalid", "", "invalid UTF-8")
	}
	raw, err := json.Marshal(q)
	if err != nil {
		return nil, requestErrorAt("request_invalid", "", "cannot serialize comparison")
	}
	if _, _, _, err = readWire(raw, reflect.TypeOf(q)); err != nil {
		return nil, err
	}
	if q.SchemaVersion != 1 {
		return nil, requestErrorAt("request_invalid", "/schema_version", "schema_version must be 1")
	}
	if err = admitSavedReport(q.Before, "/before"); err != nil {
		return nil, err
	}
	if err = admitSavedReport(q.After, "/after"); err != nil {
		return nil, err
	}
	before := map[string]ReportEntry{}
	for _, e := range q.Before.Entries {
		before[e.ID] = e
	}
	if len(before) != len(q.After.Entries) {
		return nil, requestErrorAt("selection_mismatch", "/after/entries", "selected ID sets differ")
	}
	for _, e := range q.After.Entries {
		if _, ok := before[e.ID]; !ok {
			return nil, requestErrorAt("selection_mismatch", "/after/entries", "selected ID sets differ")
		}
	}
	out := &ComparisonReport{SchemaVersion: 1, ExecutionComplete: q.Before.ExecutionComplete && q.After.ExecutionComplete, BeforeProvenance: q.Before.Provenance, AfterProvenance: q.After.Provenance, Counts: ComparisonCounts{Selected: len(before)}, Entries: []ComparisonEntry{}}
	for i, a := range q.After.Entries {
		b := before[a.ID]
		e := ComparisonEntry{ID: a.ID, Before: b, After: a, Classification: impact.Indeterminate, Deltas: []impact.EvidenceDelta{}, Pairs: []EvidencePair{}, Unmatched: []string{}, Ambiguous: []string{}, Reasons: []string{}}
		switch {
		case b.Failure != nil || a.Failure != nil:
			e.Classification = impact.Failed
			e.Reasons = append(e.Reasons, "entry_failure")
			out.Counts.Failed++
			out.ExecutionComplete = false
		case !sameQuery(b.Analysis.Document, a.Analysis.Document):
			e.Reasons = append(e.Reasons, "query_input_mismatch")
			out.Counts.Compared++
			out.Counts.Indeterminate++
		default:
			alignEntry(&e, i)
			definite := evidenceDeltas(&e, i)
			complete := comparisonComplete(b) && comparisonComplete(a)
			aligned := len(e.Unmatched) == 0 && len(e.Ambiguous) == 0
			e.Classification = classifyComparison(false, definite, complete, aligned)
			if definite {
				e.Reasons = append(e.Reasons, "meaningful_evidence_changed")
			}
			if !complete {
				e.Reasons = append(e.Reasons, "evidence_incomplete")
			}
			if !aligned {
				e.Reasons = append(e.Reasons, "correspondence_unresolved")
			}
			out.Counts.Compared++
			switch e.Classification {
			case impact.Affected:
				out.Counts.Affected++
			case impact.Unchanged:
				out.Counts.Unchanged++
			default:
				out.Counts.Indeterminate++
			}
		}
		out.Entries = append(out.Entries, e)
	}
	if !out.ExecutionComplete {
		out.CIExitCode = 2
	} else if out.Counts.Affected > 0 {
		out.CIExitCode = 1
	} else if out.Counts.Indeterminate > 0 {
		out.CIExitCode = 3
	}
	return detachedExport(out)
}
func sameQuery(a, b analysis.QueryDocument) bool {
	return a.Text == b.Text && a.Language == b.Language && a.Profile == b.Profile && a.Version == b.Version
}

func admitSavedReport(r Report, path string) error {
	fail := func(p, m string) error { return requestErrorAt("request_invalid", path+p, m) }
	if r.Selection.Complete != (len(r.Selection.TraversalFailures) == 0) {
		return fail("/selection", "selection completeness disagrees with traversal failures")
	}
	switch r.Selection.Mode {
	case "inline", "manifest", "directory":
	default:
		return fail("/selection/mode", "unsupported selection mode")
	}
	if r.SchemaVersion != 1 {
		return fail("/schema_version", "report version must be 1")
	}
	if len(r.Entries) == 0 {
		return fail("/entries", "selected entries must be nonempty")
	}
	seen := map[string]bool{}
	for i, e := range r.Entries {
		p := fmt.Sprintf("/entries/%d", i)
		if strings.TrimSpace(e.ID) == "" || seen[e.ID] {
			return fail(p+"/id", "id must be unique and nonblank")
		}
		seen[e.ID] = true
		if e.Mode != "compatibility" && e.Mode != "resolution" {
			return fail(p+"/mode", "unsupported mode")
		}
		if (e.Mode == "compatibility" && e.Resolution != nil) || (e.Mode == "resolution" && e.Compatibility != nil) {
			return fail(p, "mode has contradictory evidence payload")
		}
		if e.Failure != nil {
			switch e.Failure.Phase {
			case "acquisition", "configuration", "internal":
			default:
				return fail(p+"/failure/phase", "unsupported failure phase")
			}
			if strings.TrimSpace(e.Failure.Code) == "" {
				return fail(p+"/failure/code", "failure code required")
			}
			if e.Status != analysis.Incomplete {
				return fail(p+"/status", "failed entry must be incomplete")
			}
			if e.Failure.Phase == "acquisition" && (e.Analysis != nil || e.Compatibility != nil || e.Resolution != nil || e.SourceHash != "") {
				return fail(p, "acquisition failure has analysis evidence")
			}
		} else if e.Analysis == nil || (e.Mode == "compatibility" && (e.Compatibility == nil || e.Resolution != nil)) || (e.Mode == "resolution" && (e.Resolution == nil || e.Compatibility != nil)) {
			return fail(p, "mode requires exactly its evidence payload")
		}
		if e.Analysis != nil {
			if e.SourceHash != corpus.SourceHash(e.Analysis.Document.Text) {
				return fail(p+"/source_hash", "source hash disagrees with exact text")
			}
			if err := admitAnalysis(e.Analysis, path+p+"/analysis"); err != nil {
				return err
			}
		}
		if e.Compatibility != nil {
			if e.Compatibility.Provenance.EnvironmentDigest != r.Provenance.EnvironmentDigest || e.Compatibility.Provenance.SchemaBundleDigest != r.Provenance.SchemaBundleDigest {
				return fail(p+"/compatibility/provenance", "artifact identity disagrees with report provenance")
			}
			if e.Analysis == nil || !reflect.DeepEqual(e.Compatibility.Requirements, e.Analysis.Requirements) {
				return fail(p+"/compatibility/requirements", "requirements disagree with original analysis")
			}
			if err := admitCompatibility(e.Compatibility, e.Analysis, path+p+"/compatibility"); err != nil {
				return err
			}
			if e.Failure == nil && e.Status != compatibilityStatus(e.Analysis, e.Compatibility) {
				return fail(p+"/status", "status disagrees with retained compatibility")
			}
		}
		if e.Resolution != nil {
			if e.Resolution.Provenance.EnvironmentDigest != r.Provenance.EnvironmentDigest || e.Resolution.Provenance.SchemaBundleDigest != r.Provenance.SchemaBundleDigest {
				return fail(p+"/resolution/provenance", "artifact identity disagrees with report provenance")
			}
			if err := admitResolution(e.Resolution, e.Analysis, path+p+"/resolution"); err != nil {
				return err
			}
			if e.Failure == nil && e.Status != resolutionStatus(e.Analysis, e.Resolution) {
				return fail(p+"/status", "status disagrees with resolution")
			}
		}
	}
	expected := Report{Status: analysis.Valid, ExecutionComplete: r.Selection.Complete, Counts: Counts{Selected: len(r.Entries), TraversalFailed: len(r.Selection.TraversalFailures)}, Entries: r.Entries}
	finalize(&expected)
	if r.Counts != expected.Counts || r.Status != expected.Status || r.CIExitCode != expected.CIExitCode || r.ExecutionComplete != expected.ExecutionComplete {
		return fail("/counts", "summary disagrees with retained entries and selection")
	}
	return nil
}

func admitAnalysis(a *analysis.Result, path string) error {
	fail := func(p, m string) error { return requestErrorAt("request_invalid", path+p, m) }
	if a.SchemaVersion != 1 || a.Requirements.SchemaVersion != 1 || !savedDigest(a.Requirements.CapabilityRevision) {
		return fail("/schema_version", "canonical version 1 and capability identity required")
	}
	if a.Status != analysis.Valid && a.Status != analysis.Invalid && a.Status != analysis.Incomplete {
		return fail("/status", "invalid status")
	}
	d := a.Document
	if d.Language == "" || d.Profile == "" || d.Version == "" {
		return fail("/document", "canonical selectors required")
	}
	q := a.Requirements.Query
	if q.QueryDigest != "sha256:"+corpus.SourceHash(d.Text) || q.SourceID != d.SourceID || q.Language != d.Language || q.Profile != d.Profile || q.Version != d.Version || a.Requirements.QueryStatus != a.Status {
		return fail("/requirements/query", "query identity disagrees with original bytes or selectors")
	}
	if !reflect.DeepEqual(a.Inputs, a.Requirements.Inputs) || !reflect.DeepEqual(a.Correlation, a.Requirements.Correlation) {
		return fail("/requirements", "analysis input evidence disagrees with requirements")
	}
	stages, scopes, refs, inputs, occ := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	add := func(m map[string]bool, id, p string) error {
		if id == "" || m[id] {
			return fail(p, "id must be unique and nonblank")
		}
		m[id] = true
		return nil
	}
	for i, s := range a.Stages {
		if err := add(stages, s.ID, fmt.Sprintf("/stages/%d/id", i)); err != nil {
			return err
		}
	}
	for i, s := range a.Scopes {
		if err := add(scopes, s.ID, fmt.Sprintf("/scopes/%d/id", i)); err != nil {
			return err
		}
	}
	for i, r := range a.References {
		if err := add(refs, r.ID, fmt.Sprintf("/references/%d/id", i)); err != nil {
			return err
		}
	}
	check := func(m map[string]bool, id string) bool { return id == "" || m[id] }
	for _, s := range a.Stages {
		if !scopes[s.ScopeID] {
			return fail("/stages", "unresolved scope")
		}
	}
	for _, s := range a.Scopes {
		if !check(scopes, s.ParentID) || !check(stages, s.StageID) {
			return fail("/scopes", "unresolved stage or parent")
		}
	}
	for _, r := range a.References {
		if !check(stages, r.StageID) || !check(scopes, r.ScopeID) {
			return fail("/references", "unresolved stage or scope")
		}
		for _, id := range r.OriginReferenceIDs {
			if !refs[id] {
				return fail("/references", "unresolved origin reference")
			}
		}
	}
	for i, in := range a.Inputs {
		if err := add(inputs, in.ID, fmt.Sprintf("/inputs/%d/id", i)); err != nil {
			return err
		}
		for _, o := range in.Occurrences {
			if err := add(occ, o.ID, "/inputs"); err != nil {
				return err
			}
			if !check(refs, o.ReferenceID) || !check(refs, o.OriginalReferenceID) || !check(stages, o.StageID) || !check(scopes, o.ScopeID) {
				return fail("/inputs", "unresolved occurrence reference")
			}
			for _, id := range o.UseSiteReferenceIDs {
				if !refs[id] {
					return fail("/inputs", "unresolved use site")
				}
			}
		}
	}
	refFacts := map[string]analysis.Reference{}
	for _, r := range a.References {
		refFacts[r.ID] = r
	}
	ids := map[string]bool{}
	for _, r := range a.Requirements.Items {
		if err := add(ids, r.ID, "/requirements/items"); err != nil {
			return err
		}
		if !check(inputs, r.InputID) {
			return fail("/requirements/items", "unresolved input")
		}
		for _, id := range r.Ownership.CandidateInputIDs {
			if !inputs[id] {
				return fail("/requirements/items", "unresolved ownership")
			}
		}
		for _, o := range r.Occurrences {
			ref := refFacts[o.ReferenceID]
			if ref.Location != o.Location || ref.StageID != o.StageID || ref.ScopeID != o.ScopeID {
				return fail("/requirements/items", "requirement occurrence disagrees with reference source ownership")
			}
			if !refs[o.ReferenceID] || !check(stages, o.StageID) || !check(scopes, o.ScopeID) {
				return fail("/requirements/items", "unresolved occurrence")
			}
			for _, id := range o.InputOccurrenceIDs {
				if !occ[id] {
					return fail("/requirements/items", "unresolved input occurrence")
				}
			}
		}
	}
	for _, l := range a.Lineage {
		if !stages[l.StageID] || !scopes[l.ScopeID] {
			return fail("/lineage", "unresolved stage or scope")
		}
		for _, t := range l.Transitions {
			if !check(refs, t.OutputReferenceID) {
				return fail("/lineage", "unresolved output reference")
			}
			for _, id := range t.InputReferenceIDs {
				if !refs[id] {
					return fail("/lineage", "unresolved input reference")
				}
			}
		}
	}
	for _, n := range a.Correlation.Nodes {
		if !inputs[n.InputID] || !occ[n.OccurrenceID] {
			return fail("/correlation/nodes", "unresolved input or occurrence")
		}
	}
	for _, edge := range a.Correlation.Edges {
		for _, ep := range []analysis.CorrelationEndpoint{edge.Left, edge.Right} {
			if !inputs[ep.InputID] || !occ[ep.OccurrenceID] {
				return fail("/correlation/edges", "unresolved endpoint")
			}
			for _, id := range ep.ReferenceIDs {
				if !refs[id] {
					return fail("/correlation/edges", "unresolved endpoint reference")
				}
			}
		}
	}
	return validateLocations(reflect.ValueOf(*a), d.Text, path)
}
func validateLocations(v reflect.Value, text, path string) error {
	if v.Type() == reflect.TypeOf(analysis.Location{}) {
		l := v.Interface().(analysis.Location)
		for _, p := range []analysis.Position{l.Start, l.End} {
			if p.Offset < 0 || p.Offset > len(text) || p.Line < 1 || p.Column < 1 || (p.Offset < len(text) && !utf8.RuneStart(text[p.Offset])) {
				return requestErrorAt("request_invalid", path, "invalid source location")
			}
		}
		for _, p := range []analysis.Position{l.Start, l.End} {
			line, column := 1, 1
			previous := rune(0)
			for offset, r := range text {
				if offset >= p.Offset {
					break
				}
				switch r {
				case '\r':
					line++
					column = 1
				case '\n':
					if previous != '\r' {
						line++
					}
					column = 1
				default:
					column++
				}
				previous = r
			}
			if p.Line != line || p.Column != column {
				return requestErrorAt("request_invalid", path, "source line or column disagrees with byte offset")
			}
		}
		if l.End.Offset < l.Start.Offset {
			return requestErrorAt("request_invalid", path, "reversed source interval")
		}
		return nil
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			return validateLocations(v.Elem(), text, path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.IsExported() {
				// Canonical input discovery emits unlocated reasons when discovery
				// stops without an attributable source interval. Other location
				// owners, and partially populated reasons, still require a range.
				if v.Type() == reflect.TypeOf(analysis.InputReason{}) && f.Name == "Location" && v.Field(i).IsZero() {
					continue
				}
				if err := validateLocations(v.Field(i), text, path+"/"+strings.Split(f.Tag.Get("json"), ",")[0]); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := validateLocations(v.Index(i), text, fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

func admitCompatibility(c *compatibility.Report, a *analysis.Result, path string) error {
	fail := func(m string) error { return requestErrorAt("request_invalid", path, m) }
	if c.SchemaVersion != 1 || c.Provenance.QueryDigest != a.Requirements.Query.QueryDigest || c.Provenance.CapabilityRevision != a.Requirements.CapabilityRevision || c.Provenance.AnalysisContractVersion != 1 || c.Provenance.RequirementSetVersion != 1 {
		return fail("compatibility identity disagrees with analysis")
	}
	if c.Outcome != "satisfied" && c.Outcome != "unsatisfied" && c.Outcome != "incomplete" && c.Outcome != "not assessed" {
		return fail("invalid compatibility outcome")
	}
	if c.Closure != nil {
		if c.EffectiveRequirements != nil && c.Closure.EffectiveAnalysis != nil && !reflect.DeepEqual(*c.EffectiveRequirements, c.Closure.EffectiveAnalysis.Requirements) {
			return fail("effective requirements disagree with closure analysis")
		}
		if err := admitClosure(c.Closure, a, path+"/closure"); err != nil {
			return err
		}
	}
	sets := []analysis.RequirementSet{c.Requirements}
	if c.EffectiveRequirements != nil {
		sets = append(sets, *c.EffectiveRequirements)
	}
	if c.Closure != nil {
		sets = append(sets, c.Closure.EffectiveAnalysis.Requirements)
		for _, d := range c.Closure.DefinitionAnalyses {
			if d.DirectAnalysis != nil {
				sets = append(sets, d.DirectAnalysis.Requirements)
			}
			if d.EffectiveAnalysis != nil {
				sets = append(sets, d.EffectiveAnalysis.Requirements)
			}
		}
	}
	for _, o := range c.RequirementOutcomes {
		found := o.Query == a.Requirements.Query && o.RequirementID == "capability:language:"+a.Document.Language+":profile:"+a.Document.Profile
		for _, s := range sets {
			if s.Query == o.Query {
				if o.RequirementID == "capability:language:"+s.Query.Language+":profile:"+s.Query.Profile {
					found = true
				}
				for _, r := range s.Items {
					if r.ID == o.RequirementID {
						found = true
					}
				}
			}
		}
		if !found && c.Closure != nil {
			for _, d := range c.Closure.DefinitionAnalyses {
				if d.EffectiveAnalysis != nil && d.EffectiveAnalysis.Requirements.Query == o.Query {
					for _, r := range d.EffectiveAnalysis.Requirements.Items {
						if r.ID == o.RequirementID {
							found = true
						}
					}
				}
			}
		}
		if !found {
			return fail("unresolved requirement outcome")
		}
	}
	return nil
}
func admitClosure(c *closure.Report, a *analysis.Result, path string) error {
	fail := func(m string) error { return requestErrorAt("request_invalid", path, m) }
	if c.SchemaVersion != 1 || c.Query != a.Requirements.Query || c.DirectAnalysis == nil || c.EffectiveAnalysis == nil {
		return fail("closure identity disagrees with original")
	}
	for _, x := range []*analysis.Result{c.DirectAnalysis, c.EffectiveAnalysis} {
		if err := admitAnalysis(x, path); err != nil {
			return err
		}
	}
	if !reflect.DeepEqual(c.DirectRequirements, c.DirectAnalysis.Requirements) || !sameQuery(c.DirectAnalysis.Document, a.Document) {
		return fail("closure original evidence disagrees")
	}
	for _, d := range c.DefinitionAnalyses {
		for _, x := range []*analysis.Result{d.DirectAnalysis, d.EffectiveAnalysis} {
			if x != nil {
				if err := admitAnalysis(x, path+"/definition_analyses"); err != nil {
					return err
				}
			}
		}
	}
	edges, nodes := map[string]bool{}, map[string]bool{}
	for _, e := range c.Traversal {
		if e.ID == "" || edges[e.ID] {
			return fail("duplicate traversal identity")
		}
		edges[e.ID] = true
	}
	for _, n := range c.Graph.Nodes {
		if n.ID == "" || nodes[n.ID] {
			return fail("duplicate closure node")
		}
		nodes[n.ID] = true
	}
	for _, e := range c.Graph.Edges {
		if !edges[e.ID] || !nodes[e.From] || (e.To != "" && !nodes[e.To]) {
			return fail("unresolved closure graph linkage")
		}
	}
	for _, b := range c.BOM {
		for _, o := range b.Occurrences {
			if !edges[o.EdgeID] {
				return fail("unresolved BOM occurrence")
			}
		}
	}
	sources := map[string]string{"query\x00": a.Document.Text}
	for _, d := range c.DefinitionAnalyses {
		if d.DirectAnalysis != nil {
			sources["definition\x00"+d.ObjectID] = d.DirectAnalysis.Document.Text
		}
	}
	if err := admitSourceIntervals(reflect.ValueOf(*c), sources, path); err != nil {
		return err
	}
	return nil
}
func admitResolution(r *resolution.Report, a *analysis.Result, path string) error {
	fail := func(m string) error { return requestErrorAt("request_invalid", path, m) }
	if a == nil || r.SchemaVersion != 1 || !reflect.DeepEqual(r.Original.Analysis.Requirements, a.Requirements) || !sameQuery(r.Original.Analysis.Document, a.Document) || r.Provenance.QueryDigest != a.Requirements.Query.QueryDigest || r.Provenance.CapabilityRevision != a.Requirements.CapabilityRevision {
		return fail("resolution identity disagrees with original")
	}
	if err := admitAnalysis(&r.Original.Analysis, path+"/original/analysis"); err != nil {
		return err
	}
	counts := resolution.Counts{}
	seen := map[string]bool{}
	for _, v := range r.Variants {
		if v.ID == "" || seen[v.ID] {
			return fail("duplicate variant identity")
		}
		seen[v.ID] = true
		switch v.Outcome {
		case "verified":
			counts.Verified++
		case "failed":
			counts.Failed++
		case "incomplete":
			counts.Incomplete++
		default:
			return fail("invalid variant outcome")
		}
		if err := admitResolutionVariant(v, r, a, path+"/variants"); err != nil {
			return err
		}
		if v.CandidateAnalysis != nil {
			if err := admitAnalysis(v.CandidateAnalysis, path+"/variants"); err != nil {
				return err
			}
			if v.CandidateText == nil || *v.CandidateText != v.CandidateAnalysis.Document.Text {
				return fail("candidate bytes disagree with candidate analysis")
			}
		}
		if v.Compatibility != nil {
			c := v.Compatibility
			if v.CandidateAnalysis == nil || !reflect.DeepEqual(c.Requirements, v.CandidateAnalysis.Requirements) || c.SchemaVersion != 1 {
				return fail("candidate compatibility identity mismatch")
			}
			if c.Closure != nil {
				if err := admitClosure(c.Closure, v.CandidateAnalysis, path+"/variants/closure"); err != nil {
					return err
				}
			}
		}
	}
	if r.Counts != counts || r.GeneratedCount != uint64(len(r.Variants)) {
		return fail("resolution counts disagree with variants")
	}
	return nil
}

func savedDigest(s string) bool {
	if !strings.HasPrefix(s, "sha256:") || len(s) != 71 {
		return false
	}
	_, err := hex.DecodeString(s[7:])
	return err == nil
}
func admitSourceIntervals(v reflect.Value, sources map[string]string, path string) error {
	if v.Type() == reflect.TypeOf(closure.SourceInterval{}) {
		s := v.Interface().(closure.SourceInterval)
		text, known := sources[s.Kind+"\x00"+s.ObjectID]
		if s.Start < 0 || s.End < s.Start || (known && (s.End > len(text) || (s.Start < len(text) && !utf8.RuneStart(text[s.Start])) || (s.End < len(text) && !utf8.RuneStart(text[s.End])))) || (!known && (s.Start != 0 || s.End != 0)) {
			return requestErrorAt("request_invalid", path, "invalid captured source interval")
		}
		return nil
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsNil() {
			return admitSourceIntervals(v.Elem(), sources, path)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if f.IsExported() {
				if err := admitSourceIntervals(v.Field(i), sources, path+"/"+strings.Split(f.Tag.Get("json"), ",")[0]); err != nil {
					return err
				}
			}
		}
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if err := admitSourceIntervals(v.Index(i), sources, fmt.Sprintf("%s/%d", path, i)); err != nil {
				return err
			}
		}
	}
	return nil
}

// These checks validate links and consistency in exposed evidence. They do not
// restore a ResolutionSession or certify an imported proof as engine authority.
func admitResolutionVariant(v resolution.Variant, r *resolution.Report, original *analysis.Result, path string) error {
	fail := func(m string) error { return requestErrorAt("request_invalid", path, m) }
	candidate := v.CandidateAnalysis
	if candidate == nil {
		if v.Outcome == "verified" || v.Compatibility != nil || v.ResolvedQuery != nil || v.Proof.Proven || len(v.Proof.References) > 0 || len(v.Proof.Roles) > 0 {
			return fail("candidate evidence required for verification, proof or publication")
		}
		for _, change := range v.Changes {
			if change.CandidateLocation != nil || len(change.CandidateReferenceIDs) > 0 {
				return fail("candidate change evidence requires candidate analysis")
			}
			if err := validateLocations(reflect.ValueOf(change.OriginalLocation), original.Document.Text, path+"/changes/original_location"); err != nil {
				return err
			}
		}
		return nil
	}
	if v.Provenance.QueryDigest != candidate.Requirements.Query.QueryDigest || v.Provenance.CapabilityRevision != candidate.Requirements.CapabilityRevision || v.Provenance.AnalysisContractVersion != 1 || v.Provenance.RequirementSetVersion != 1 || v.Provenance.EnvironmentDigest != r.Provenance.EnvironmentDigest || v.Provenance.SchemaBundleDigest != r.Provenance.SchemaBundleDigest || v.Provenance.ResolutionInputDigest != r.Provenance.ResolutionInputDigest || v.Provenance.AssessmentInputDigest != r.Provenance.AssessmentInputDigest {
		return fail("variant provenance disagrees with retained candidate or artifacts")
	}
	oi, ci := inputIndex(original), inputIndex(candidate)
	or, cr := requirementIndex(original), requirementIndex(candidate)
	orefs, crefs := referenceIndex(original), referenceIndex(candidate)
	for _, p := range v.Proof.References {
		if _, ok := orefs[p.OriginalID]; !ok {
			return fail("unresolved original proof reference")
		}
		if _, ok := crefs[p.CandidateID]; !ok {
			return fail("unresolved candidate proof reference")
		}
	}
	checkOccurrence := func(in analysis.QueryInput, id string) bool {
		for _, o := range in.Occurrences {
			if o.ID == id {
				return true
			}
		}
		return false
	}
	checkReqOccurrence := func(in analysis.RequirementItem, o analysis.RequirementOccurrence) bool {
		for _, x := range in.Occurrences {
			if reflect.DeepEqual(x, o) {
				return true
			}
		}
		return false
	}
	for _, role := range v.Proof.Roles {
		x, xok := oi[role.OriginalInput.ID]
		y, yok := ci[role.CandidateInput.ID]
		if !xok || !yok || !reflect.DeepEqual(x, role.OriginalInput) || !reflect.DeepEqual(y, role.CandidateInput) {
			return fail("proof role disagrees with retained inputs")
		}
		for _, p := range role.Occurrences {
			if p.OriginalInputID != x.ID || p.CandidateInputID != y.ID || !checkOccurrence(x, p.OriginalOccurrenceID) || !checkOccurrence(y, p.CandidateOccurrenceID) {
				return fail("unresolved proof input occurrence")
			}
		}
		for _, p := range role.Requirements {
			before, bok := or[p.OriginalRequirementID]
			after, aok := cr[p.CandidateRequirementID]
			if !bok || !aok || !checkReqOccurrence(before, p.OriginalOccurrence) || !checkReqOccurrence(after, p.CandidateOccurrence) {
				return fail("unresolved proof requirement occurrence")
			}
		}
	}
	for _, change := range v.Changes {
		if err := validateLocations(reflect.ValueOf(change.OriginalLocation), original.Document.Text, path+"/changes/original_location"); err != nil {
			return err
		}
		if original.Document.Text[change.OriginalLocation.Start.Offset:change.OriginalLocation.End.Offset] != change.Before {
			return fail("change original bytes disagree with source interval")
		}
		if change.CandidateLocation != nil {
			if err := validateLocations(reflect.ValueOf(*change.CandidateLocation), candidate.Document.Text, path+"/changes/candidate_location"); err != nil {
				return err
			}
			if candidate.Document.Text[change.CandidateLocation.Start.Offset:change.CandidateLocation.End.Offset] != change.After {
				return fail("change candidate bytes disagree with source interval")
			}
		}
		for _, id := range change.OriginalReferenceIDs {
			if _, ok := orefs[id]; !ok {
				return fail("unresolved change original reference")
			}
		}
		for _, id := range change.CandidateReferenceIDs {
			if _, ok := crefs[id]; !ok {
				return fail("unresolved change candidate reference")
			}
		}
	}
	definitive := candidate.Status == analysis.Invalid
	for _, l := range v.Proof.Limitations {
		if l.Code == "target_not_renderable" || l.Code == "render_conflict" {
			definitive = true
		}
	}
	if v.Compatibility != nil {
		c := v.Compatibility
		adapted := compatibility.Report{SchemaVersion: c.SchemaVersion, Outcome: c.Outcome, Requirements: c.Requirements, EffectiveRequirements: c.EffectiveRequirements, Closure: c.Closure, Provenance: c.Provenance}
		for _, o := range c.RequirementOutcomes {
			adapted.RequirementOutcomes = append(adapted.RequirementOutcomes, o.Evidence)
		}
		if err := admitCompatibility(&adapted, candidate, path+"/compatibility"); err != nil {
			return err
		}
		if c.Provenance.SourceID != candidate.Document.SourceID || c.Provenance.EnvironmentDigest != r.Provenance.EnvironmentDigest || c.Provenance.SchemaBundleDigest != r.Provenance.SchemaBundleDigest {
			return fail("candidate assessment provenance disagrees with source or artifacts")
		}
		assessmentInputs := map[string][]analysis.QueryInput{}
		addAssessmentInputs := func(set analysis.RequirementSet) {
			for _, in := range set.Inputs {
				assessmentInputs[in.ID] = append(assessmentInputs[in.ID], in)
			}
		}
		addAssessmentInputs(candidate.Requirements)
		if c.EffectiveRequirements != nil {
			addAssessmentInputs(*c.EffectiveRequirements)
		}
		if c.Closure != nil {
			addAssessmentInputs(c.Closure.EffectiveAnalysis.Requirements)
			for _, d := range c.Closure.DefinitionAnalyses {
				if d.DirectAnalysis != nil {
					addAssessmentInputs(d.DirectAnalysis.Requirements)
				}
				if d.EffectiveAnalysis != nil {
					addAssessmentInputs(d.EffectiveAnalysis.Requirements)
				}
			}
		}
		assessmentOccurrence := func(inputID, occurrenceID string) bool {
			for _, in := range assessmentInputs[inputID] {
				if checkOccurrence(in, occurrenceID) {
					return true
				}
			}
			return false
		}
		for _, in := range c.Inputs {
			if in.OriginalInputID != "" {
				if _, ok := oi[in.OriginalInputID]; !ok {
					return fail("unresolved assessment original input")
				}
			}
			if len(assessmentInputs[in.CandidateInputID]) == 0 {
				return fail("unresolved assessment candidate input")
			}
			for _, p := range in.Occurrences {
				before, bok := oi[p.OriginalInputID]
				if !bok || !checkOccurrence(before, p.OriginalOccurrenceID) || !assessmentOccurrence(p.CandidateInputID, p.CandidateOccurrenceID) {
					return fail("unresolved assessment input occurrence")
				}
			}
		}
		for _, o := range c.RequirementOutcomes {
			if o.OriginalInputID != "" {
				if _, ok := oi[o.OriginalInputID]; !ok {
					return fail("unresolved assessment original input")
				}
			}
			if o.OriginalRequirementID != "" {
				if _, ok := or[o.OriginalRequirementID]; !ok {
					return fail("unresolved assessment original requirement")
				}
			}
			if o.CandidateRequirementID != o.Evidence.RequirementID || o.CandidateInputID != o.Evidence.InputID {
				return fail("candidate assessment link disagrees with retained outcome")
			}
			// Closure-defined requirements belong to their own query domain, validated
			// by admitCompatibility above rather than the candidate's local ID table.
			if o.Evidence.Query == candidate.Requirements.Query {
				if o.CandidateInputID != "" {
					if _, ok := ci[o.CandidateInputID]; !ok {
						return fail("unresolved assessment candidate input")
					}
				}
				if item, ok := cr[o.CandidateRequirementID]; ok {
					for _, occ := range o.AssessedOccurrences {
						if !checkReqOccurrence(item, occ) {
							return fail("assessment occurrence disagrees with candidate")
						}
					}
				}
			}
		}
		if c.Outcome == "unsatisfied" {
			definitive = true
		}
	}
	if definitive && v.Outcome != "failed" {
		return fail("variant outcome hides retained failure")
	}
	if v.Outcome == "verified" {
		if !v.Proof.Proven || v.Compatibility == nil || v.Compatibility.Outcome != "satisfied" || v.ResolvedQuery == nil || v.CandidateText == nil || *v.ResolvedQuery != *v.CandidateText {
			return fail("verified variant lacks consistent proof, assessment or publication")
		}
		for _, d := range v.Diagnostics {
			if d.Code == "substitution_incomplete" {
				return fail("verified variant retains incomplete substitution")
			}
		}
	} else if v.ResolvedQuery != nil {
		return fail("unverified variant publishes a query")
	}
	return nil
}
func inputIndex(a *analysis.Result) map[string]analysis.QueryInput {
	out := map[string]analysis.QueryInput{}
	for _, x := range a.Inputs {
		out[x.ID] = x
	}
	return out
}
func requirementIndex(a *analysis.Result) map[string]analysis.RequirementItem {
	out := map[string]analysis.RequirementItem{}
	for _, x := range a.Requirements.Items {
		out[x.ID] = x
	}
	return out
}
func referenceIndex(a *analysis.Result) map[string]analysis.Reference {
	out := map[string]analysis.Reference{}
	for _, x := range a.References {
		out[x.ID] = x
	}
	return out
}
