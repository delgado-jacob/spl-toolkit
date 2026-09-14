package analysis

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
	"sort"
)

// Descriptors retain real grammar ownership for later lowering. Parent is the
// enclosing descriptor index (-1 for the document); input is descriptive only.
// SQL EXISTS owns a direct SQL command, whereas command children own pipelines
// or inherited subpipes. Neither an array nor an embedded string is a scope.
type spl2ChildScope struct {
	kind     string
	input    string
	parent   int
	owner    antlr.ParserRuleContext
	body     antlr.ParserRuleContext
	location Location
}

func spl2ChildScopes(p *spl2ParsedDocument) []spl2ChildScope {
	return spl2ChildScopesIn(p, []antlr.Tree{p.tree})
}
func spl2ChildScopesIn(p *spl2ParsedDocument, trees []antlr.Tree) []spl2ChildScope {
	scopes := []spl2ChildScope{}
	var visit func(antlr.Tree, int)
	visit = func(tree antlr.Tree, parent int) {
		var owner, body antlr.ParserRuleContext
		kind, input := "", ""
		switch ctx := tree.(type) {
		case *spl2.IndependentSearchContext:
			if !spl2OriginalScopeDelimiters(ctx.LBRACKET(), ctx.RBRACKET()) {
				return
			}
			owner, body, kind, input = ctx, ctx.Pipeline(), "search", "independent"
		case *spl2.InheritedSubpipeContext:
			if !spl2OriginalScopeDelimiters(ctx.LBRACKET(), ctx.RBRACKET()) {
				return
			}
			owner, body, kind, input = ctx, ctx, "subpipe", "inherited"
		case *spl2.ExistsPredicateContext:
			if !spl2OriginalScopeDelimiters(ctx.LPAREN(), ctx.RPAREN()) {
				return
			}
			owner, kind, input = ctx, "exists", "correlated"
			if ctx.FromCommand() != nil {
				body = ctx.FromCommand()
			} else {
				body = ctx.SelectCommand()
			}
		}
		// A damaged ancestor cannot promote a nested intact bracket to its
		// parent's scope. Its original child ownership is no longer proved.
		if owner != nil && (body == nil || !spl2IntactSyntax(owner)) {
			return
		}
		if owner != nil && body != nil {
			scopes = append(scopes, spl2ChildScope{kind: kind, input: input, parent: parent, owner: owner, body: body, location: p.source.contextLocation(owner)})
			parent = len(scopes) - 1
		}
		for _, child := range tree.GetChildren() {
			visit(child, parent)
		}
	}
	for _, tree := range trees {
		if tree != nil {
			visit(tree, -1)
		}
	}
	return scopes
}

// Scheduling shares the ordinary field kernel and SQL phase evaluator. A child
// owns its copied input; unknown merge semantics never install its outputs in
// the parent. References are finalized only after every scheduled scope.
type spl2ScopeScheduler struct {
	result     *Result
	parsed     *spl2ParsedDocument
	refinement *sourceRefinement
	children   []spl2ChildScope
	executed   map[int]bool
}

func spl2PipelineContexts(tree antlr.Tree) []antlr.ParserRuleContext {
	contexts := []antlr.ParserRuleContext{}
	if tree == nil {
		return contexts
	}
	for _, node := range tree.GetChildren() {
		switch c := node.(type) {
		case *spl2.StartContext, *spl2.CommandContext:
			for _, child := range c.GetChildren() {
				if ctx, ok := child.(antlr.ParserRuleContext); ok {
					contexts = append(contexts, ctx)
					break
				}
			}
		}
	}
	return contexts
}

func (q *spl2ScopeScheduler) pipeline(sites []spl2CommandSite, env *environment, aliases map[string]bool, scopeID string, parent int) *environment {
	position := 0
	for siteIndex, site := range sites {
		ctx := site.context
		if sql, ok := ctx.(spl2SQLCommand); ok && spl2ScheduledSQL(sql) {
			before := len(q.result.Stages)
			env = executeSPL2SQL(q.result, q.parsed, q.refinement, sql, env, aliases, q, scopeID, parent, position)
			for _, stage := range q.result.Stages[before:] {
				if stage.ScopeID == scopeID {
					position++
				}
			}
			continue
		}
		location, command := site.location, site.command
		index := registerSPL2Stage(q.result, location, command, position, scopeID)
		position++
		s := &spl2SemanticStage{semanticStage: &semanticStage{result: q.result, stage: index, env: env, transitions: []Transition{}, refinement: q.refinement}, parsed2: q.parsed, aliases: aliases}
		before := env.snapshot()
		if _, ok := ctx.(*spl2.TimewrapCommandContext); ok && spl2IntactSyntax(ctx) && siteIndex > 0 {
			_, fromStart := sites[0].context.(*spl2.FromCommandContext)
			proved, noTimechart := fromStart, true
			for _, prior := range sites[:siteIndex] {
				if prior.context == nil || !spl2IntactSyntax(prior.context) {
					proved = false
					break
				}
				if _, ok := prior.context.(*spl2.EmbeddedCommandContext); ok {
					proved = false
				}
				if _, ok := prior.context.(*spl2.TimechartCommandContext); ok {
					noTimechart = false
				}
			}
			if proved && noTimechart {
				token := ctx.GetStart()
				s.diagnosticAt(CodeSyntaxError, "error", "contract", "Timewrap requires a preceding timechart command", q.parsed.source.location(token.GetStart(), token.GetStop()+1), true)
			}
		}
		if ctx != nil {
			q.runChildren(ctx, env, aliases, scopeID, parent)
		}
		if ctx != nil && spl2IntactSyntax(ctx) {
			s.command(ctx)
		} else {
			if ctx != nil {
				s.recoveredInputs(ctx)
			}
			message := "Recovered SPL2 command effects are not yet modeled"
			if ctx == nil {
				message = "Standalone command effects are unproved"
			}
			s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", message, location, true)
		}
		if !q.result.Stages[index].SemanticComplete {
			s.env.uncertain = true
		}
		env = s.env
		q.result.Lineage = append(q.result.Lineage, Lineage{StageID: q.result.Stages[index].ID, ScopeID: scopeID, Before: before, After: env.snapshot(), Transitions: s.transitions})
	}
	return env
}

func (q *spl2ScopeScheduler) runChildren(ctx antlr.ParserRuleContext, env *environment, aliases map[string]bool, scopeID string, parent int) {
	for i, child := range q.children {
		if child.parent != parent || q.executed[i] || !spl2Within(child.owner, ctx) {
			continue
		}
		q.executed[i] = true
		ownerStage := ""
		// The phase/stage is already registered, even when SQL lexical registration
		// precedes execution. Prefer the narrowest containing clause.
		for _, stage := range q.result.Stages {
			if stage.ScopeID == scopeID && stage.Location.Start.Offset <= child.location.Start.Offset && stage.Location.End.Offset >= child.location.End.Offset {
				ownerStage = stage.ID
			}
		}
		id := fmt.Sprintf("scope-%d", len(q.result.Scopes))
		q.result.Scopes = append(q.result.Scopes, Scope{ID: id, ParentID: scopeID, Kind: child.kind, StageID: ownerStage, Location: child.location})
		input := newEnvironmentWithRequirementTrace(env.requirements.trace)
		localAliases := map[string]bool{}
		if child.input == "inherited" {
			input = env.clone()
		}
		if child.input != "independent" {
			for name, value := range aliases {
				localAliases[name] = value
			}
		}
		if sql, ok := child.body.(spl2SQLCommand); ok {
			executeSPL2SQL(q.result, q.parsed, q.refinement, sql, input, localAliases, q, id, i, 0)
		} else {
			q.pipeline(spl2Sites(spl2PipelineContexts(child.body), q.parsed.source), input, localAliases, id, i)
		}
	}
}

func spl2Within(child, owner antlr.ParserRuleContext) bool {
	for p := antlr.Tree(child); p != nil; p = p.GetParent() {
		if p == owner {
			return true
		}
	}
	return false
}

// SQL is registered by lexical clauses and executed by phase. Children may run
// between those phases, so remap lexical stage IDs once before reference IDs.
// No-child reports retain their accepted wire values exactly.
func spl2FinalizeStages(r *Result) map[string]string {
	sort.SliceStable(r.Stages, func(i, j int) bool { return r.Stages[i].Location.Start.Offset < r.Stages[j].Location.Start.Offset })
	mapping := map[string]string{}
	for i := range r.Stages {
		old := r.Stages[i].ID
		r.Stages[i].ID = fmt.Sprintf("stage-%d", i)
		mapping[old] = r.Stages[i].ID
	}
	if r.rewrite != nil {
		for _, site := range r.rewrite.sites {
			site.public.Point.StageID = mapping[site.public.Point.StageID]
		}
	}
	for i := range r.References {
		r.References[i].StageID = mapping[r.References[i].StageID]
	}
	for i := range r.Diagnostics {
		if id := r.Diagnostics[i].StageID; id != "" {
			r.Diagnostics[i].StageID = mapping[id]
		}
	}
	for i := range r.Scopes {
		if id := r.Scopes[i].StageID; id != "" {
			r.Scopes[i].StageID = mapping[id]
		}
	}
	for i := range r.Lineage {
		r.Lineage[i].StageID = mapping[r.Lineage[i].StageID]
		order := i
		r.Lineage[i].ExecutionOrder = &order
	}
	return mapping
}

// Dataset object keys define a closed local event shape. A key missing or null
// in any row is conditional, and its actual key token supplies the definition.
func (s *spl2SemanticStage) dataset(dataset spl2.IDatasetContext) {
	if dataset.Identifier() != nil {
		s.dependency(dataset.Identifier(), "dataset")
		return
	}
	rows := dataset.Array()
	if rows == nil || !spl2IntactSyntax(rows) {
		s.unsupported(dataset, "Dataset literal shape is unproved")
		return
	}
	if len(rows.AllExpression()) == 0 {
		s.result.Coverage.SyntaxComplete = false
		s.unsupported(rows, "H09 empty dataset literal shape and effects remain unproved")
		return
	}
	s.env.open = false
	type keyValue struct {
		target           locatedOperand
		ids              []string
		rows             int
		nonnull, allnull bool
	}
	values := map[string]*keyValue{}
	order := []string{}
	for _, row := range rows.AllExpression() {
		access := spl2SingleAccess(row)
		if access == nil || len(access.AllAccessPart()) != 0 || access.Primary().Object() == nil {
			s.unsupported(row, "Dataset row shape is unproved")
			s.env.uncertain = true
			continue
		}
		for _, entry := range access.Primary().Object().AllObjectEntry() {
			key := s.operand(entry.ObjectKey())
			if !key.Sound {
				continue
			}
			value := s.expression(entry.Expression())
			v := values[key.Name]
			if v == nil {
				v = &keyValue{target: key, nonnull: true, allnull: true}
				values[key.Name] = v
				order = append(order, key.Name)
			}
			v.rows++
			v.ids = uniqueIDs(v.ids, value.ids)
			v.nonnull = v.nonnull && value.nonnull
			v.allnull = v.allnull && value.exactNull
		}
	}
	for _, name := range order {
		v := values[name]
		s.applyAssignment(v.target, v.ids, !v.nonnull || v.rows != len(rows.AllExpression()), v.allnull)
	}
}

func spl2OriginalScopeDelimiters(open, close antlr.TerminalNode) bool {
	if open == nil || close == nil {
		return false
	}
	a, b := open.GetSymbol(), close.GetSymbol()
	return a.GetTokenIndex() >= 0 && b.GetTokenIndex() > a.GetTokenIndex() &&
		((a.GetTokenType() == spl2.SPL2ParserLBRACKET && b.GetTokenType() == spl2.SPL2ParserRBRACKET) ||
			(a.GetTokenType() == spl2.SPL2ParserLPAREN && b.GetTokenType() == spl2.SPL2ParserRPAREN))
}
