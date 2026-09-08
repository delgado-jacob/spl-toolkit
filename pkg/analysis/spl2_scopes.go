package analysis

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
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
	scopes := []spl2ChildScope{}
	var visit func(antlr.Tree, int)
	visit = func(tree antlr.Tree, parent int) {
		var owner, body antlr.ParserRuleContext
		kind, input := "", ""
		switch ctx := tree.(type) {
		case *spl2.IndependentSearchContext:
			owner, body, kind, input = ctx, ctx.Pipeline(), "search", "independent"
		case *spl2.InheritedSubpipeContext:
			owner, body, kind, input = ctx, ctx, "subpipe", "inherited"
		case *spl2.ExistsPredicateContext:
			owner, kind, input = ctx, "exists", "correlated"
			if ctx.FromCommand() != nil {
				body = ctx.FromCommand()
			} else {
				body = ctx.SelectCommand()
			}
		}
		if owner != nil && body != nil && spl2IntactSyntax(owner) {
			scopes = append(scopes, spl2ChildScope{kind: kind, input: input, parent: parent, owner: owner, body: body, location: p.source.contextLocation(owner)})
			parent = len(scopes) - 1
		}
		for _, child := range tree.GetChildren() {
			visit(child, parent)
		}
	}
	visit(p.tree, -1)
	return scopes
}
