package analysis

import (
	"sort"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type spl2SQLCommand interface {
	spl2SQLClauses
	SqlSelectClause() spl2.ISqlSelectClauseContext
	SqlGroupClause() spl2.ISqlGroupClauseContext
	SqlOrderClause() spl2.ISqlOrderClauseContext
}

// Public labels authorize unqualified reads. Qualified visibility additionally
// retains the typed source identity and its supplying occurrence evidence.
type spl2SQLVisibility struct {
	labels    map[string]bool
	qualified map[spl2SQLQualifiedVisibilityKey]bool
}

type spl2SQLQualifiedVisibilityKey struct {
	identity fieldIdentityKey
	owners   string
}

func sqlVisibilityOwnerKey(owners []sourceOwner) string {
	occurrences := make([]string, 0, len(owners))
	for _, owner := range owners {
		occurrences = append(occurrences, inputOccurrenceID(owner.input))
	}
	sort.Strings(occurrences)
	return orderedStringSliceKey(append([]string{sourceOwnerKey(owners)}, uniqueIDs(occurrences)...))
}

func (s *spl2SemanticStage) sqlQualifiedVisibilityKey(identity fieldIdentity) (spl2SQLQualifiedVisibilityKey, bool) {
	binding, known := s.aliases[identity.Qualifier]
	key, exact := identity.privateKey()
	if !known || binding.environment == nil || !exact || identity.Qualifier == "" {
		return spl2SQLQualifiedVisibilityKey{}, false
	}
	relative := pathFieldIdentity("", identity.Segments)
	if len(identity.Segments) == 1 {
		relative = atomicFieldIdentity(identity.Segments[0])
	}
	return spl2SQLQualifiedVisibilityKey{identity: key, owners: sqlVisibilityOwnerKey(binding.environment.requirements.sourceOwners(relative))}, true
}

func (s *spl2SemanticStage) sqlQualifiedVisible(visible spl2SQLVisibility, identity fieldIdentity) bool {
	key, exact := s.sqlQualifiedVisibilityKey(identity)
	return exact && visible.qualified[key]
}

func spl2ScheduledSQL(c spl2SQLCommand) bool {
	return c.SqlFromClause() != nil && len(c.SqlFromClause().AllSqlJoinClause()) > 0 || c.SqlSelectClause() != nil || c.SqlWhereClause() != nil || c.SqlGroupClause() != nil || c.SqlHavingClause() != nil || c.SqlOrderClause() != nil || c.SqlLimitClause() != nil || c.SqlOffsetClause() != nil
}

// Stages retain lexical clause order. Each phase uses its real owning clause;
// preparation and final restriction can therefore share the one SELECT stage.
func executeSPL2SQL(result *Result, parsed *spl2ParsedDocument, refinement *sourceRefinement, c spl2SQLCommand, env *environment, aliases spl2Aliases, scheduler *spl2ScopeScheduler, scopeID string, parent, position int) *environment {
	provedKey := parsed.proveGroupKeyBeforeEmptyHaving(c)
	missingSelect := parsed.groupMissingSelectDiagnostic(c)
	selectedShape := false
	logical := []antlr.ParserRuleContext{}
	for _, ctx := range []antlr.ParserRuleContext{c.SqlFromClause(), c.SqlWhereClause(), c.SqlGroupClause(), c.SqlSelectClause(), c.SqlHavingClause(), c.SqlOrderClause(), c.SqlLimitClause(), c.SqlOffsetClause()} {
		if ctx != nil {
			logical = append(logical, ctx)
		}
	}
	positions := map[antlr.ParserRuleContext]int{}
	for i, ctx := range logical {
		positions[ctx] = position + i
	}
	lexical := append([]antlr.ParserRuleContext{}, logical...)
	sort.SliceStable(lexical, func(i, j int) bool { return lexical[i].GetStart().GetStart() < lexical[j].GetStart().GetStart() })
	stages := map[antlr.ParserRuleContext]int{}
	for _, ctx := range lexical {
		stages[ctx] = registerSPL2Stage(result, parsed.source.contextLocation(ctx), strings.ToLower(ctx.GetStart().GetText()), positions[ctx], scopeID)
	}
	if provedKey != nil {
		index := stages[c.SqlGroupClause()]
		result.Stages[index].SemanticComplete = true
		for i := range result.Diagnostics {
			d := &result.Diagnostics[i]
			original := *d
			original.StageID = ""
			original.ScopeID = ""
			if original == provedKey.diagnostic {
				// This original parser finding remains document recovery evidence.
				d.StageID = ""
				d.ScopeID = ""
			} else if d.StageID == result.Stages[index].ID {
				result.Stages[index].SemanticComplete = false
			}
		}
	}
	scheduler.syncParserDiagnostics()
	s := &spl2SemanticStage{semanticStage: &semanticStage{result: result, env: env, refinement: refinement}, parsed2: parsed, aliases: aliases, program: scheduler.program}
	phase := func(ctx antlr.ParserRuleContext, name string, run func()) {
		if ctx == nil {
			return
		}
		s.stage = stages[ctx]
		s.rewritePhase, s.rewriteOrdinal = name, 0
		s.transitions = []Transition{}
		before := s.env.snapshot()
		scheduler.runChildren(ctx, s.env, aliases, scopeID, parent)
		scheduler.assignParserDiagnostics(s.stage)
		if spl2IntactSyntax(ctx) || (provedKey != nil && ctx == c.SqlGroupClause()) {
			if (provedKey != nil || missingSelect != nil) && ctx == c.SqlGroupClause() {
				local := *parsed
				local.diagnostics = []Diagnostic{}
				for _, d := range parsed.diagnostics {
					if (provedKey == nil || d != provedKey.diagnostic) && (missingSelect == nil || d != *missingSelect) {
						local.diagnostics = append(local.diagnostics, d)
					}
				}
				s.parsed2 = &local
			}
			run()
			s.parsed2 = parsed
		} else {
			if from, ok := ctx.(spl2.ISqlFromClauseContext); ok {
				s.sourceIntentions(from)
			} else if name != "project" {
				s.recoveredSQLInputs(ctx)
			}
			s.unsupported(ctx, "Recovered SQL clause effects are not yet modeled")
		}

		if !result.Stages[s.stage].SemanticComplete && !(name == "project" && selectedShape) {
			s.env.uncertain = true
		}
		if s.env.requirements.stageIncomplete(result.Stages[s.stage].ID) && !(name == "project" && selectedShape) {
			s.env.requirements.uncertain = true
		}
		order := len(result.Lineage)
		st := result.Stages[s.stage]
		result.Lineage = append(result.Lineage, Lineage{StageID: st.ID, ScopeID: st.ScopeID, Before: before, After: s.env.snapshot(), Transitions: s.transitions, Phase: name, ExecutionOrder: &order})
	}
	phase(c.SqlFromClause(), "source", func() {
		s.applySource()
		if parent < 0 || scheduler.children[parent].input != "correlated" {
			clear(aliases)
		}
		from := c.SqlFromClause()
		dataset := from.Dataset()
		resolved := false
		if s.program != nil && dataset != nil {
			resolved = s.program.resolveViewSource(s, dataset.DatasetParameter()) || s.program.resolveImportedDataset(s, dataset)
		}
		if !resolved && !s.exactDatasetSource(dataset) {
			s.dataset(dataset)
		}
		if a := from.SourceAlias(); a != nil {
			o := s.operand(a.Identifier())
			if o.Sound {
				aliases[o.Name] = spl2SourceAlias{environment: s.env.clone()}
			}
		}
		for _, join := range from.AllSqlJoinClause() {
			s.lowerSQLJoin(join)
		}
	})
	phase(c.SqlWhereClause(), "filter", func() {
		s.rewritePredicate(c.SqlWhereClause().SqlPredicate(), "spl2", false, false)
		s.expression(c.SqlWhereClause().SqlPredicate())
	})
	pregroup := s.env
	phase(c.SqlGroupClause(), "group", func() {
		groups := []locatedOperand{}
		qualified := []preparedSelection{}
		qualifiedSources := map[string]*environment{}
		aliases := []string{}
		for _, key := range c.SqlGroupClause().AllSqlGroupKey() {
			field := s.sqlDirectField(key.Expression())
			if provedKey != nil && key == provedKey.key {
				field = s.selector(provedKey.identifier)
			}
			if field.Sound && (field.Identity.Qualifier == "" || !strings.Contains(field.Name, "*")) && (provedKey != nil && key == provedKey.key || s.parsed2.soundOperand(key)) && key.SqlSpanAssignment() == nil {
				if field.Identity.Qualifier == "" {
					groups = append(groups, field)
					continue
				}
				source := s.sqlQualifiedSource(field.Identity.Qualifier, field.Identity.Segments)
				if source != nil {
					identity := atomicFieldIdentity(field.Name)
					id := s.sourceFieldReference(field, source, identity, "group")
					if projected, ok := s.sqlProjectedField(field, []string{id}); ok {
						alias := field.Identity.Qualifier
						row := qualifiedSources[alias]
						if row == nil {
							row = newEnvironmentWithRequirementTrace(s.env.requirements.trace)
							row.open, row.requirements.open = false, false
							qualifiedSources[alias] = row
							aliases = append(aliases, alias)
						}
						row.registerIdentityField(projected)
						if requirement, ok := source.requirements.exactIdentityProjection(identity, []string{id}); ok {
							row.requirements.registerIdentityField(requirement)
						}
						qualified = append(qualified, preparedSelection{Field: projected, InputReferenceIDs: []string{id}, EmitProjectTransition: true})
					}
					continue
				}
			}
			s.expression(key)
			s.unsupported(key, "SQL grouping expression output is unproved")
		}
		s.applyAggregation(nil, groups, false)
		// Retain the proved per-alias group row separately from public suffix
		// collisions. Compose once per alias, rather than cloning traces per key.
		s.sqlGroupedSources = qualifiedSources
		order := append([]fieldIdentityKey{}, s.env.fieldOrder...)
		for _, alias := range aliases {
			combined, _, ok := composeFlowEnvironments(s.env, qualifiedSources[alias])
			if !ok {
				s.unsupported(c.SqlGroupClause(), "SQL qualified grouping source facts are unproved")
				delete(s.sqlGroupedSources, alias)
				continue
			}
			s.installSelectedFlowMerge(s.env, combined)
		}
		seen := map[fieldIdentityKey]bool{}
		for _, key := range order {
			seen[key] = true
		}
		for _, selection := range qualified {
			key, _ := selection.Field.identity.privateKey()
			if !seen[key] {
				order = append(order, key)
				seen[key] = true
			}
			s.appendTransition(Transition{Operation: "project", Output: selection.Field.Name, OutputIdentity: transitionOutputIdentity(selection.Field.identity), InputReferenceIDs: selection.InputReferenceIDs})
		}
		s.env.fieldOrder = order
	})
	aggregate := s.sqlHasAggregate(c.SqlSelectClause())
	selectPhase := "evaluate"
	if aggregate {
		selectPhase = "aggregate"
	}
	selected := []preparedSelection{}
	visible := spl2SQLVisibility{}
	phase(c.SqlSelectClause(), selectPhase, func() {
		selected, visible, selectedShape = s.prepareSQLSelection(c.SqlSelectClause(), pregroup, aggregate, c.SqlGroupClause() != nil)
	})
	phase(c.SqlHavingClause(), "having", func() {
		predicate := c.SqlGroupClause() != nil || aggregate
		if !predicate {
			s.unsupported(c.SqlHavingClause(), "SQL HAVING without grouping is unproved")
		}
		s.sqlRestrictedExpression(c.SqlHavingClause().SqlPredicate(), visible, predicate)
	})
	phase(c.SqlOrderClause(), "order", func() {
		if c.SqlSelectClause() == nil {
			s.expression(c.SqlOrderClause())
		} else {
			s.sqlRestrictedExpression(c.SqlOrderClause(), visible, false)
		}
	})
	phase(c.SqlSelectClause(), "project", func() {
		s.applyPreparedProjection(selected, "table")
		s.sqlGroupedSources = nil
		if selectedShape {
			s.env.open, s.env.uncertain = false, false
			fields := map[fieldIdentityKey]requirementField{}
			for _, selection := range selected {
				identity := selection.Field.identity
				if _, exact := identity.privateKey(); !exact {
					identity = atomicFieldIdentity(selection.Field.Name)
				}
				if field, ok := s.sqlSelectionRequirement(selection); ok {
					key, _ := identity.privateKey()
					fields[key] = field
				}
			}
			s.env.requirements.fields = fields
			s.env.requirements.open, s.env.requirements.uncertain = false, false
		}
		if !selectedShape {
			for _, selection := range selected {
				if selection.Field.ownerCollision {
					field := requirementField{identity: selection.Field.identity.clone(), owners: cloneSourceOwners(selection.Field.owners), source: selection.Field.source, conditional: true, ownerCollision: true, origins: copyIDs(selection.Field.OriginReferenceIDs)}
					key, _ := field.identity.privateKey()
					s.env.requirements.fields[key] = field
				}
			}
		}
	})
	phase(c.SqlLimitClause(), "limit", func() {})
	phase(c.SqlOffsetClause(), "offset", func() {})
	return s.env
}

func spl2SQLDirectField(tree antlr.Tree) spl2.IIdentifierContext {
	access := spl2SQLFieldAccess(tree)
	if access != nil && len(access.AllAccessPart()) == 0 && access.Primary().FieldName() != nil {
		return access.Primary().FieldName().Identifier()
	}
	return nil
}

func (s *spl2SemanticStage) sqlHasAggregate(tree antlr.Tree) bool {
	if tree == nil {
		return false
	}
	if c, ok := tree.(spl2.ICallContext); ok && s.sqlCallAggregate(c) {
		return true
	}
	// Child expressions own their own scheduling and aggregate context.
	switch tree.(type) {
	case spl2.IExistsPredicateContext, spl2.ISearchLiteralContext:
		return false
	}
	for _, child := range tree.GetChildren() {
		if s.sqlHasAggregate(child) {
			return true
		}
	}
	return false
}

func (s *spl2SemanticStage) sqlCallAggregate(call spl2.ICallContext) bool {
	if call == nil {
		return false
	}
	policy, selected := s.typedPolicy()
	function, known := policy.functions[call.Identifier().GetText()]
	return selected && known && function.signature.aggregate
}

// Compound aggregate results remain unproved. Read their real arguments with
// the reviewed function policy, avoiding a false scalar-context contract error.
func (s *spl2SemanticStage) sqlUnprovedAggregateExpression(tree antlr.Tree) spl2ExpressionEvidence {
	if !s.sqlHasAggregate(tree) {
		return s.expression(tree)
	}
	if c, ok := tree.(spl2.ICallContext); ok {
		aggregate := s.sqlCallAggregate(c)
		out := s.callWithExpression(c, aggregate, s.sqlUnprovedAggregateExpression)
		if !aggregate {
			out.modeled, out.nonnull, out.requirementNonnull, out.exactNull, out.truth, out.domain = false, false, false, false, false, ""
		}
		return out
	}
	out := spl2ExpressionEvidence{ids: []string{}}
	for _, child := range tree.GetChildren() {
		out.ids = append(out.ids, s.sqlUnprovedAggregateExpression(child).ids...)
	}
	return out
}

func (s *spl2SemanticStage) prepareSQLSelection(clause spl2.ISqlSelectClauseContext, pregroup *environment, aggregate, grouped bool) ([]preparedSelection, spl2SQLVisibility, bool) {
	type projection struct {
		ctx          spl2.IProjectionContext
		target       locatedOperand
		value        spl2ExpressionEvidence
		field        *trackedField
		aggregate    bool
		countCertain bool
	}
	items := []projection{}
	input := s.env
	if aggregate && !grouped {
		s.applyAggregation(nil, nil, false)
		input = s.env
	}
	groups := map[string]bool{}
	for _, field := range input.fields {
		groups[field.Name] = true
	}
	groupVisibility := spl2SQLVisibility{labels: groups, qualified: map[spl2SQLQualifiedVisibilityKey]bool{}}
	for alias, binding := range s.aliases {
		if binding.environment == nil {
			continue
		}
		for _, field := range input.requirements.fields {
			identity := pathFieldIdentity(alias, field.identity.Segments)
			key, exact := s.sqlQualifiedVisibilityKey(identity)
			if exact && sourceOwnersProved(field.owners) && len(field.owners) > 0 && key.owners == sqlVisibilityOwnerKey(field.owners) {
				groupVisibility.qualified[key] = true
			}
		}
		if retained := s.sqlGroupedSources[alias]; retained != nil {
			for _, field := range retained.fields {
				if key, exact := s.sqlQualifiedVisibilityKey(pathFieldIdentity(alias, field.identity.Segments)); exact {
					groupVisibility.qualified[key] = true
				}
			}
		}
	}
	for _, p := range clause.AllProjection() {
		item := projection{ctx: p}
		field := s.sqlDirectField(p.Expression())
		access := spl2SQLFieldAccess(p.Expression())
		var call spl2.ICallContext
		if access != nil && len(access.AllAccessPart()) == 0 {
			call = access.Primary().Call()
		}
		item.aggregate = s.sqlCallAggregate(call)
		if item.aggregate {
			s.env = pregroup
			item.value = s.call(call, true)
			item.countCertain = s.sqlCountOutputSound(p, call, item.value)
			s.env = input
		} else if s.sqlHasAggregate(p.Expression()) {
			s.env = pregroup
			item.value = s.sqlUnprovedAggregateExpression(p.Expression())
			s.env = input
			s.unsupported(p, "Compound SQL aggregate output is unproved")
		} else if aggregate || grouped {
			if !field.Sound || !groups[field.Name] {
				s.unsupported(p, "Mixed SQL aggregate/non-grouped projection is unproved")
			}
			item.value = s.sqlRestrictedExpression(p.Expression(), groupVisibility, false)
		} else {
			item.value = s.expression(p.Expression())
		}
		if p.ProjectionAlias() != nil {
			item.target = s.operand(p.ProjectionAlias().Identifier())
			// A qualified field explicitly retaining its suffix is the same
			// projection, with no sibling alias or new destination binding.
			if field.Sound && field.Identity.Qualifier != "" && item.target.Name == field.Name {
				if projected, ok := s.sqlProjectedField(field, item.value.ids); ok {
					item.field = &projected
					item.target = locatedOperand{}
				}
			}
		} else if item.aggregate {
			name := call.Identifier().GetText()
			label := ""
			if name == "count" && call.Arguments() == nil {
				label = "count"
			}
			if args := call.Arguments(); args != nil && len(args.AllNamedArgument()) == 0 && len(args.AllExpression()) == 1 {
				f := spl2SQLDirectField(args.Expression(0))
				if f != nil && ((name == "sum" && f.GetText() == "bytes") || (name == "dc" && f.GetText() == "action")) {
					label = name + "(" + f.GetText() + ")"
				}
			}
			if label != "" {
				item.target = locatedOperand{Name: label, Location: s.parsed2.source.contextLocation(call), Resolution: "exact", Sound: true, rewrite: rewriteOwner{role: "implicit_output", location: s.parsed2.source.contextLocation(call), implicit: name}}
			} else {
				s.unsupported(p, "Implicit aggregate output label is unproved; use AS")
			}
		} else if field.Sound {
			if f, ok := s.sqlProjectedField(field, item.value.ids); ok {
				item.field = &f
			}
		} else if spl2SQLDirectField(p.Expression()) != nil {
			// A damaged identifier cannot supply a field or output, and the
			// original parser finding already owns that absent proof.
		} else {
			s.unsupported(p, "Implicit SQL expression output label is unproved; use AS")
		}
		items = append(items, item)
	}
	// All expressions see input, never a sibling SELECT alias. Ambiguous output
	// names remain conditional and cannot authorize later alias visibility.
	counts := map[string]int{}
	ambiguousDestination := false
	for _, item := range items {
		if name, known := s.sqlProjectionCollisionName(item.ctx, item.target); known {
			counts[name]++
		} else {
			ambiguousDestination = true
		}
	}
	collisions := map[string]bool{}
	for _, item := range items {
		if item.field != nil && counts[item.field.Name] > 1 {
			collisions[item.field.Name] = true
			s.unsupported(item.ctx, "SQL projection output field collision is unproved")
		}
		if !item.target.Sound {
			continue
		}
		_, source := pregroup.field(item.target.fieldIdentity())
		if source || counts[item.target.Name] > 1 {
			collisions[item.target.Name] = true
			s.unsupported(item.ctx, "SQL alias collision or sibling alias visibility is unproved")
		}
	}
	selected := []preparedSelection{}
	visible := spl2SQLVisibility{labels: map[string]bool{}, qualified: map[spl2SQLQualifiedVisibilityKey]bool{}}
	finiteShape := !ambiguousDestination && len(collisions) == 0 && s.sqlProjectionEffectSound(clause)
	incompleteInputs := map[string]bool{}
	if s.refinement != nil {
		for _, expansion := range s.refinement.expansions {
			if !expansion.Complete {
				incompleteInputs[expansion.ReferenceID] = true
			}
		}
	}
	for _, item := range items {
		for _, id := range s.origins(item.value.ids) {
			if incompleteInputs[id] {
				finiteShape = false
			}
		}
		if item.field != nil {
			item.field.Conditional = item.field.Conditional || collisions[item.field.Name]
			selected = append(selected, preparedSelection{Field: *item.field, InputReferenceIDs: item.value.ids, EmitProjectTransition: true})
			visible.labels[item.field.Name] = true
			field := s.sqlDirectField(item.ctx.Expression())
			if key, exact := s.sqlQualifiedVisibilityKey(field.Identity); exact {
				visible.qualified[key] = !collisions[item.field.Name]
			}
		}
		if !item.target.Sound {
			if item.field == nil {
				finiteShape = false
			}
			continue
		}
		if item.target.Name == "" || item.target.Resolution != "exact" {
			finiteShape = false
		}
		before := len(s.result.References)
		if item.aggregate && item.countCertain && !collisions[item.target.Name] && !ambiguousDestination && item.target.Resolution == "exact" {
			// This one intact output has its own non-null proof; sibling/owner
			// limitations still govern coverage, visibility and all other outputs.
			s.createAt(item.target, "output", "aggregate", item.value.ids, false)
		} else if item.aggregate {
			s.applyAggregation([]aggregateOutput{{
				Target:                 item.target,
				InputReferenceIDs:      item.value.ids,
				Conditional:            !item.value.nonnull,
				RequirementConditional: !item.value.requirementNonnull,
			}}, nil, true)
		} else {
			s.applyAssignmentWithRequirementConditional(item.target, item.value.ids, !item.value.nonnull || collisions[item.target.Name], !item.value.requirementNonnull || collisions[item.target.Name], item.value.exactNull)
		}
		visible.labels[item.target.Name] = !collisions[item.target.Name]
		if !item.value.exactNull {
			ids := []string{s.result.References[before].ID}
			if f, ok := s.projectedField(item.target.Name, ids); ok {
				selected = append(selected, preparedSelection{Field: f, InputReferenceIDs: ids, EmitProjectTransition: true})
			} else {
				finiteShape = false
			}
		}
	}
	// Repeated labels preserve all supplying owners as candidates; the final
	// output cannot select an owner by whichever projection was visited last.
	ownersByName := map[string][]sourceOwner{}
	originsByName := map[string][]string{}
	for _, selection := range selected {
		name := selection.Field.Name
		if collisions[name] {
			ownersByName[name] = append(ownersByName[name], selection.Field.owners...)
			originsByName[name] = append(originsByName[name], selection.Field.OriginReferenceIDs...)
		}
	}
	for name, owners := range ownersByName {
		ownersByName[name] = mergeSourceOwners(nil, owners)
		originsByName[name] = uniqueIDs(originsByName[name])
	}
	for i := range selected {
		name := selected[i].Field.Name
		if !collisions[name] {
			continue
		}
		selected[i].Field.ownerCollision = true
		selected[i].Field.owners = cloneSourceOwners(ownersByName[name])
		selected[i].Field.OriginReferenceIDs = copyIDs(originsByName[name])
		visible.labels[name] = false
	}
	return selected, visible, finiteShape
}

// Potential SQL destinations participate before field creation, including
// projections whose binding is unproved. Qualified labels use the same static
// suffix policy as selected projections; this check itself proves no owner.
func (s *spl2SemanticStage) sqlProjectionCollisionName(p spl2.IProjectionContext, target locatedOperand) (string, bool) {
	local, _ := s.parsed2.recoveredCountView(p)
	if !local.soundOperand(p) {
		return "", false
	}
	if target.Sound && target.Resolution == "exact" {
		return target.Name, true
	}
	if p.ProjectionAlias() != nil {
		return "", false
	}
	access := spl2SQLFieldAccess(p.Expression())
	if access == nil || access.Primary().FieldName() == nil {
		return "", false
	}
	base := s.operand(access.Primary().FieldName().Identifier())
	if !base.Sound || base.Resolution != "exact" {
		return "", false
	}
	parts := access.AllAccessPart()
	if len(parts) == 0 {
		return base.Name, true
	}
	if len(parts) != 1 || !s.aliases.recognizes(base.Name) || parts[0].DOT() == nil || parts[0].Identifier() == nil {
		return "", false
	}
	suffix := s.operand(parts[0].Identifier())
	return suffix.Name, suffix.Sound && suffix.Resolution == "exact"
}

// Per-output count proof is deliberately narrower than SELECT owner coverage.
func (s *spl2SemanticStage) sqlCountOutputSound(projection spl2.IProjectionContext, call spl2.ICallContext, value spl2ExpressionEvidence) bool {
	return value.modeled && value.nonnull && call.Identifier().GetText() == "count" && call.Arguments() == nil && s.sqlProjectionEffectSound(projection)
}

func (s *spl2SemanticStage) sqlProjectionEffectSound(ctx antlr.ParserRuleContext) bool {
	local, missingSelect := s.parsed2.recoveredCountView(ctx)
	if !local.soundOperand(ctx) {
		return false
	}
	location := s.parsed2.source.contextLocation(ctx)
	for _, d := range s.result.Diagnostics {
		original := d
		original.StageID, original.ScopeID = "", ""
		if missingSelect != nil && original == *missingSelect {
			continue
		}
		if d.Severity == "error" && d.Location.End.Offset >= location.Start.Offset && d.Location.Start.Offset <= location.End.Offset {
			return false
		}
	}
	return true
}

// A temporary visibility view keeps hidden source/group origins without
// claiming availability or mutating the actual phase state. The shared reader
// still decides source/derived/null-test/conditional binding for every operand.
func (s *spl2SemanticStage) sqlRestrictedExpression(tree antlr.Tree, visible spl2SQLVisibility, predicate bool) spl2ExpressionEvidence {
	actual := s.env
	previousVisibility := s.sqlVisibility
	s.sqlVisibility = &visible
	defer func() { s.sqlVisibility = previousVisibility }()
	s.env = actual.clone()
	hidden := []locatedOperand{}
	var inspect func(antlr.Tree)
	inspect = func(node antlr.Tree) {
		switch c := node.(type) {
		case spl2.IExistsPredicateContext, spl2.ISearchLiteralContext:
			return
		case spl2.IAccessContext:
			operand := s.sqlQualifiedAccess(c)
			if operand.Sound && operand.Identity.Qualifier != "" {
				if !s.sqlQualifiedVisible(visible, operand.Identity) {
					hidden = append(hidden, operand)
				}
				return
			}
		case spl2.IFieldNameContext:
			if c.Identifier() != nil {
				o := s.operand(c.Identifier())
				if o.Sound && !visible.labels[o.Name] {
					key, _ := o.fieldIdentity().privateKey()
					f := s.env.fields[key]
					f.Name = o.Name
					f.identity = o.fieldIdentity()
					f.Conditional = true
					s.env.fields[key] = f
					s.env.requirements.markConditional(o.Name)
					s.env.requirements.uncertain = true
					hidden = append(hidden, o)
				}
			}
		}
		for _, child := range node.GetChildren() {
			inspect(child)
		}
	}
	inspect(tree)
	if predicate {
		// Extract at the HAVING phase inside the same restricted source/derived
		// visibility view as its reads. Earlier SQL phases cannot see these facts.
		// A hidden operand must not suppress a conflicting visible-key restriction.
		s.rewritePredicate(tree, "spl2", false, false)
	}
	var value spl2ExpressionEvidence
	if s.sqlHasAggregate(tree) {
		s.unsupported(tree.(antlr.ParserRuleContext), "SQL postaggregate expression visibility is unproved")
		value = s.sqlUnprovedAggregateExpression(tree)
	} else {
		value = s.expression(tree)
	}
	hiddenIDs := map[Location][]string{}
	if trace := s.env.requirements.trace; trace != nil {
		for _, id := range value.ids {
			entry := trace.reference(id)
			hiddenIDs[entry.reference.Location] = append(hiddenIDs[entry.reference.Location], id)
		}
	}
	for _, operand := range hidden {
		owners := hiddenIDs[operand.Location]
		if len(owners) > 1 {
			hiddenIDs[operand.Location] = owners[1:]
		} else {
			hiddenIDs[operand.Location] = nil
		}
		if len(owners) > 0 {
			owners = owners[:1]
		}
		s.result.Stages[s.stage].SemanticComplete = false
		s.diagnosticAtOwnedWithRequirementIncomplete(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "SQL field visibility outside selected/grouped outputs is unproved", operand.Location, owners)
	}
	if predicate {
		// Persist predicate/observed-reference evidence, not the temporary field
		// visibility or expression effects used to classify HAVING operands.
		actual.rewrite = s.env.rewrite
	}
	s.env = actual
	return value
}

// Preserve real expression inputs of damaged clauses once, without preparing
// any projection alias, group output, or final output shape.
func (s *spl2SemanticStage) recoveredSQLInputs(ctx antlr.ParserRuleContext) {
	switch c := ctx.(type) {
	case spl2.ISqlSelectClauseContext:
		for _, projection := range c.AllProjection() {
			if projection.Expression() != nil {
				s.sqlUnprovedAggregateExpression(projection.Expression())
			}
		}
	case spl2.ISqlGroupClauseContext:
		for _, key := range c.AllSqlGroupKey() {
			if span := key.SqlSpanCall(); span != nil {
				if field := span.FieldName(); field != nil {
					s.readIdentifier(field.Identifier(), "read")
				}
			} else if key.SqlSpanAssignment() != nil {
				if field := spl2SQLDirectField(key.Expression()); field != nil {
					s.readIdentifier(field, "read")
				}
			}
		}
	}
}

// Alias ownership comes only from an already established SQL source. A closed
// phase can authorize only a retained field with that same source lineage.
func (s *spl2SemanticStage) sqlQualifiedSource(alias string, segments []string) *environment {
	binding, known := s.aliases[alias]
	if !known || binding.environment == nil || len(segments) == 0 {
		return nil
	}
	source := binding.environment
	identity := pathFieldIdentity("", segments)
	if len(segments) == 1 {
		identity = atomicFieldIdentity(segments[0])
	}
	key, _ := identity.privateKey()
	// A source snapshot cannot restore a field removed from the active row.
	if s.env.removed[key] || s.env.requirements.removed[key] {
		unavailable := newEnvironmentWithRequirementTrace(s.env.requirements.trace)
		unavailable.open, unavailable.requirements.open = false, false
		return unavailable
	}
	// Explicit recreation belongs to the current row and keeps its new
	// derived origins instead of reviving the original source's obligation.
	if current, present := s.env.field(identity); present && !current.source && !current.ownerCollision {
		for _, removed := range s.env.removedOrder {
			if removed == key {
				source = s.env
				break
			}
		}
	}
	groupRetained := false
	if !s.env.open {
		if grouped := s.sqlGroupedSources[alias]; grouped != nil {
			if _, retained := grouped.field(identity); retained {
				source, groupRetained = grouped, true
			}
		}
	}
	if binding.conditional || !s.env.open || s.sqlVisibility != nil {
		source = source.cloneWithRequirementTrace(s.env.requirements.trace)
		if !s.env.open && !groupRetained {
			current, retained := s.env.requirements.field(identity)
			expected := source.requirements.sourceOwners(identity)
			if !retained || sourceOwnerKey(current.owners) != sourceOwnerKey(expected) {
				source.open, source.requirements.open = false, false
				source.fields = map[fieldIdentityKey]trackedField{}
				source.requirements.fields = map[fieldIdentityKey]requirementField{}
				source.uncertain, source.requirements.uncertain = false, false
				return source
			}
		}
		if binding.conditional || s.sqlVisibility != nil && !s.sqlQualifiedVisible(*s.sqlVisibility, pathFieldIdentity(alias, segments)) {
			field, present := source.field(identity)
			if !present {
				field = trackedField{FieldBinding: FieldBinding{Name: identity.PublicName, OriginReferenceIDs: []string{}}, identity: identity, source: true, owners: source.requirements.sourceOwners(identity)}
			}
			field.Conditional = true
			source.fields[key] = field
			requirement, known := source.requirements.fields[key]
			if !known {
				requirement = requirementField{identity: identity, source: true, owners: source.requirements.sourceOwners(identity)}
			}
			// LEFT nullability changes the value's availability, but a known
			// derived value keeps its existing requirement lineage and origin.
			if requirement.source {
				requirement.conditional = true
			}
			source.requirements.fields[key] = requirement
		}
	}
	return source
}

func (s *spl2SemanticStage) lowerSQLJoin(join spl2.ISqlJoinClauseContext) {
	left := s.env
	source := *s
	semantic := *s.semanticStage
	semantic.env = newEnvironmentWithRequirementTrace(left.requirements.trace)
	source.semanticStage = &semantic
	if !source.exactDatasetSource(join.Dataset()) {
		source.dataset(join.Dataset())
	}
	right := source.env
	alias := locatedOperand{}
	if join.SourceAlias() != nil {
		alias = s.operand(join.SourceAlias().Identifier())
	}
	_, repeated := s.aliases[alias.Name]
	sources := map[string]*environment{}
	for name, binding := range s.aliases {
		if binding.environment != nil {
			environment := binding.environment
			if binding.conditional {
				environment = environment.cloneWithRequirementTrace(left.requirements.trace)
				environment.uncertain, environment.requirements.uncertain = true, true
				for key, field := range environment.fields {
					field.Conditional = true
					environment.fields[key] = field
				}
				for key, field := range environment.requirements.fields {
					if field.source {
						field.conditional = true
						environment.requirements.fields[key] = field
					}
				}
			}
			sources[name] = environment
		}
	}
	valid := s.commandEffectSound(join) && alias.Sound && !repeated && len(sources) > 0
	// ANTLR may retain a clean equality prefix before a rejected OR/XOR tail.
	// That prefix cannot prove the Boolean join predicate written by the user.
	if token := s.nextCommandToken(join); token != nil && (token.GetTokenType() == spl2.SPL2ParserOR || token.GetTokenType() == spl2.SPL2ParserXOR) {
		valid = false
	}
	if valid {
		sources[alias.Name] = right
	}
	ids, predicateOK := []string{}, false
	if valid {
		ids, predicateOK = s.sourceJoinPredicate(join.SqlJoinPredicate(), sources, alias.Name)
	}
	if !valid || !predicateOK {
		left.retainHeldJoinCandidates(right)
		// Recognize a written qualifier while withholding its source ownership.
		if alias.Sound {
			s.aliases[alias.Name] = spl2SourceAlias{}
		}
		s.unsupportedOwned(join, "SQL join layout or qualified input binding is unproved", ids)
		return
	}
	// The right key is consumed only by the matched alternative of a LEFT
	// join. Keep its exact source binding for equality proof, but classify
	// its external requirement by that alternative's conditional reachability.
	if join.LEFT() != nil {
		for _, id := range ids {
			entry := left.requirements.trace.reference(id)
			if entry.directExternal && entry.reference.FieldIdentity != nil && entry.reference.FieldIdentity.Qualifier == alias.Name {
				entry.directExternal, entry.pathConditional = false, true
			}
		}
	}
	s.recordSourceJoinCorrelations(join, join.SqlJoinPredicate(), ids, alias.Name, right)
	// Key reads teach each isolated source its field facts. Include those
	// facts in the active left row before simultaneous composition, so an
	// unqualified key cannot inherit only the last visited source's owner.
	preparedLeft := left
	names := make([]string, 0, len(sources))
	for name := range sources {
		if name != alias.Name {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		sourceRow := sources[name].cloneWithRequirementTrace(left.requirements.trace)
		sourceRow.uncertain = s.aliases[name].environment.uncertain
		sourceRow.requirements.uncertain = s.aliases[name].environment.requirements.uncertain
		combined, _, ok := composeFlowEnvironments(preparedLeft, sourceRow)
		if !ok {
			left.retainHeldJoinCandidates(right)
			s.unsupportedOwned(join, "SQL join source facts are unproved", ids)
			return
		}
		preparedLeft = combined
	}
	matched, collisions, composed := composeFlowEnvironments(preparedLeft, right)
	if !composed {
		left.retainHeldJoinCandidates(right)
		s.unsupportedOwned(join, "SQL join source composition is unproved", ids)
		return
	}
	paths := []flowMergePath{{Ordinal: 0, Environment: matched, Reachable: true}}
	if join.LEFT() != nil {
		paths = append(paths, flowMergePath{Ordinal: 1, Environment: preparedLeft.clone(), Reachable: true})
	}
	s.installSelectedFlowMerge(left, mergeFlowEnvironments(left, paths, false))
	s.aliases[alias.Name] = spl2SourceAlias{environment: right, conditional: join.LEFT() != nil}
	for _, name := range collisions {
		s.diagnosticAtOwned(CodeAmbiguousField, "warning", "unsupported_semantics", "Join output contains fields with the same public name", s.parsed2.source.contextLocation(join), false, selectedFlowFieldOrigins(s.env, name))
	}
}

// SQL labels use the static field suffix; a qualifier never becomes label text.
func (s *spl2SemanticStage) sqlDirectField(tree antlr.Tree) locatedOperand {
	if field := spl2SQLDirectField(tree); field != nil {
		return s.selector(field)
	}
	access := spl2SQLFieldAccess(tree)
	if access == nil || access.Primary().FieldName() == nil || len(access.AllAccessPart()) != 1 {
		return locatedOperand{}
	}
	base := s.operand(access.Primary().FieldName().Identifier())
	part := access.AccessPart(0)
	if !base.Sound || !s.aliases.recognizes(base.Name) || part.DOT() == nil || part.Identifier() == nil {
		return locatedOperand{}
	}
	field := s.operand(part.Identifier())
	field.Identity = pathFieldIdentity(base.Name, []string{field.Name})
	field.Location = s.parsed2.source.contextLocation(access)
	return field
}

func (s *spl2SemanticStage) sqlProjectedField(field locatedOperand, ids []string) (trackedField, bool) {
	if field.Identity.Qualifier == "" {
		return s.projectedField(field.Name, ids)
	}
	source := s.sqlQualifiedSource(field.Identity.Qualifier, field.Identity.Segments)
	if source == nil {
		return trackedField{}, false
	}
	identity := atomicFieldIdentity(field.Name)
	result, known := source.field(identity)
	if !known {
		return trackedField{}, false
	}
	result.identity = identity
	result.Name = field.Name
	return result, true
}

func (s *spl2SemanticStage) sqlSelectionRequirement(selection preparedSelection) (requirementField, bool) {
	for _, id := range selection.InputReferenceIDs {
		if trace := s.env.requirements.trace; trace != nil {
			entry := trace.reference(id)
			if entry.reference.Kind == "field" && entry.reference.FieldIdentity != nil && entry.reference.FieldIdentity.Qualifier != "" {
				return requirementField{identity: selection.Field.identity.clone(), owners: cloneSourceOwners(selection.Field.owners), source: selection.Field.source, conditional: entry.conditional, ownerCollision: selection.Field.ownerCollision, origins: copyIDs(selection.Field.OriginReferenceIDs)}, true
			}
		}
	}
	return s.env.requirements.exactIdentityProjection(selection.Field.identity, selection.InputReferenceIDs)
}

// Static expression paths retain typed segments independently of the narrower
// one-suffix SQL default label and GROUP selector contract.
func (s *spl2SemanticStage) sqlQualifiedAccess(access spl2.IAccessContext) locatedOperand {
	if access.Primary().FieldName() == nil {
		return locatedOperand{}
	}
	base := s.operand(access.Primary().FieldName().Identifier())
	if !base.Sound || !s.aliases.recognizes(base.Name) || len(access.AllAccessPart()) == 0 {
		return locatedOperand{}
	}
	segments := []string{}
	for _, part := range access.AllAccessPart() {
		if part.DOT() == nil || part.Identifier() == nil {
			return locatedOperand{}
		}
		field := s.operand(part.Identifier())
		if !field.Sound {
			return locatedOperand{}
		}
		segments = append(segments, field.Name)
	}
	identity := pathFieldIdentity(base.Name, segments)
	return locatedOperand{Name: identity.PublicName, Identity: identity, Location: s.parsed2.source.contextLocation(access), Resolution: "exact", Sound: true}
}
