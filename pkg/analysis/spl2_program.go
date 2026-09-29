package analysis

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

const spl2DeferredProgramDiagnosticStage = "spl2-program-pending"

func spl2BoundProgramDiagnostics(parsed *spl2ParsedDocument) ([]Diagnostic, bool) {
	if !spl2BindableProgram(parsed) {
		return append([]Diagnostic{}, parsed.diagnostics...), parsed.syntaxComplete
	}
	diagnostics := make([]Diagnostic, 0, len(parsed.diagnostics))
	syntaxComplete := parsed.syntaxComplete
	removedModuleBoundary := false
	for _, diagnostic := range parsed.diagnostics {
		if diagnostic.Code == "SPL_UNSUPPORTED_MODULE" {
			removedModuleBoundary = true
			continue
		}
		diagnostics = append(diagnostics, diagnostic)
	}
	if !syntaxComplete && removedModuleBoundary {
		syntaxComplete = len(parsed.lexicalErrors) == 0 && len(parsed.predictionErrors) == 0
		for _, diagnostic := range diagnostics {
			if diagnostic.Category == "syntax" || diagnostic.Category == "unsupported_syntax" || diagnostic.Category == "unsupported" || diagnostic.Category == "compatibility" {
				syntaxComplete = false
			}
		}
	}
	return diagnostics, syntaxComplete
}

func spl2BindableProgram(parsed *spl2ParsedDocument) bool {
	if parsed == nil || parsed.tree == nil || parsed.tree.ModuleDeclaration() == nil || len(parsed.lexicalErrors) != 0 || len(parsed.predictionErrors) != 0 || !spl2IntactSyntax(parsed.tree.ModuleDeclaration()) {
		return false
	}
	for _, diagnostic := range parsed.diagnostics {
		if diagnostic.Category == "syntax" || diagnostic.Category == "unsupported_syntax" && diagnostic.Code != "SPL_UNSUPPORTED_MODULE" {
			return false
		}
	}
	for _, statement := range parsed.tree.ModuleDeclaration().AllAnnotatedStatement() {
		module := statement.ModuleStatement()
		if statement.AnnotationStatement() != nil {
			continue
		}
		if module == nil || module.UnsupportedModuleBoundary() != nil {
			return false
		}
	}
	return true
}

func (p *spl2Program) bind(result *Result, refinement *sourceRefinement, trace *requirementTrace, parserDiagnosticCount int) {
	p.result = result
	p.refinement = refinement
	p.trace = trace
	p.parserDiagnosticCount = parserDiagnosticCount
	p.deferParserDiagnosticOwnership()
	p.registerDeclarations()
	p.assignDeclarationParserDiagnosticOwnership()
	p.bindDeclarationMetadata()
	p.validateFunctions()
	for _, declaration := range p.declarations {
		if declaration.view != nil {
			p.bindView(declaration.view)
		}
	}
	p.bindExports()
	p.reportUnusedImports()
	p.finishParserDiagnosticOwnership()
}

func (p *spl2Program) deferParserDiagnosticOwnership() {
	for i := 0; i < p.parserDiagnosticCount; i++ {
		if p.result.Diagnostics[i].StageID == "" {
			p.result.Diagnostics[i].StageID = spl2DeferredProgramDiagnosticStage
		}
	}
}

func (p *spl2Program) assignParserDiagnosticsToOwner(stageIndex int, owner Location) (bool, bool) {
	if stageIndex < 0 || stageIndex >= len(p.result.Stages) {
		return false, false
	}
	assigned, hasError := false, false
	stage := &p.result.Stages[stageIndex]
	for i := 0; i < p.parserDiagnosticCount; i++ {
		diagnostic := &p.result.Diagnostics[i]
		if diagnostic.StageID != spl2DeferredProgramDiagnosticStage || diagnostic.Location.Start.Offset < owner.Start.Offset || diagnostic.Location.End.Offset > owner.End.Offset {
			continue
		}
		diagnostic.StageID = stage.ID
		diagnostic.ScopeID = stage.ScopeID
		stage.SemanticComplete = false
		assigned = true
		hasError = hasError || diagnostic.Severity == "error"
	}
	p.trace.syncParserDiagnostics(p.result.Diagnostics[:p.parserDiagnosticCount])
	return assigned, hasError
}

func (p *spl2Program) assignDeclarationParserDiagnosticOwnership() {
	for _, declaration := range p.declarations {
		assigned, hasError := false, false
		assign := func(owner Location) {
			owned, ownedError := p.assignParserDiagnosticsToOwner(declaration.stage, owner)
			assigned = assigned || owned
			hasError = hasError || ownedError
		}
		for _, annotation := range declaration.annotations {
			assign(p.parsed.source.contextLocation(annotation))
		}
		if declaration.view == nil {
			assign(p.parsed.source.contextLocation(declaration.ctx))
		}
		if !assigned {
			continue
		}
		if declaration.view != nil {
			declaration.view.invalid = true
			declaration.view.parserTainted = true
		}
		if declaration.function != nil {
			declaration.function.parserTainted = true
			if hasError {
				declaration.function.invalid = true
			} else {
				declaration.function.incomplete = true
			}
		}
	}
}

func (p *spl2Program) finishParserDiagnosticOwnership() {
	for i := 0; i < p.parserDiagnosticCount; i++ {
		if p.result.Diagnostics[i].StageID == spl2DeferredProgramDiagnosticStage {
			p.result.Diagnostics[i].StageID = ""
			p.result.Diagnostics[i].ScopeID = ""
		}
	}
	p.trace.syncParserDiagnostics(p.result.Diagnostics[:p.parserDiagnosticCount])
}

func (p *spl2Program) registerDeclarations() {
	for _, declaration := range p.declarations {
		command := string(declaration.kind)
		declaration.stage = registerSPL2Stage(p.result, p.parsed.source.contextLocation(declaration.ctx), command, declaration.ordinal, "scope-0")
	}
	for _, declaration := range p.declarations {
		switch {
		case declaration.view != nil && declaration.view.body != nil:
			declaration.view.scopeID = p.appendBodyScope("view", declaration.stage, declaration.view.body)
		case declaration.function != nil && declaration.function.body != nil:
			declaration.function.scopeID = p.appendBodyScope("function", declaration.stage, declaration.function.body)
		}
	}
}

func (p *spl2Program) appendBodyScope(kind string, stage int, body antlr.ParserRuleContext) string {
	id := fmt.Sprintf("scope-%d", len(p.result.Scopes))
	p.result.Scopes = append(p.result.Scopes, Scope{
		ID:       id,
		ParentID: "scope-0",
		Kind:     kind,
		StageID:  p.result.Stages[stage].ID,
		Location: p.parsed.source.contextLocation(body),
	})
	return id
}

func (p *spl2Program) declarationStage(declaration *spl2ProgramDeclaration) *spl2SemanticStage {
	env := newEnvironmentWithRequirementTrace(p.trace)
	return &spl2SemanticStage{
		semanticStage: &semanticStage{result: p.result, stage: declaration.stage, env: env, transitions: []Transition{}, refinement: p.refinement},
		parsed2:       p.parsed,
		aliases:       map[string]bool{},
		locals:        map[string]bool{},
		program:       p,
	}
}

func (p *spl2Program) bindDeclarationMetadata() {
	for _, declaration := range p.declarations {
		stage := p.declarationStage(declaration)
		for _, annotation := range declaration.annotations {
			p.bindAnnotation(stage, annotation)
		}
		if declaration.imp != nil {
			p.bindImport(stage, declaration.imp)
		}
	}
	for _, duplicate := range p.duplicates {
		stage := p.declarationStage(duplicate.declaration)
		stage.diagnosticAtOwned(CodeDuplicateSymbol, "error", "contract", fmt.Sprintf("symbol %q is declared more than once", duplicate.name), duplicate.location, true, nil)
		p.markDeclarationInvalid(duplicate.original)
		p.markDeclarationInvalid(duplicate.declaration)
		if duplicate.originalBinding != nil {
			duplicate.originalBinding.invalid = true
		}
		if duplicate.binding != nil {
			duplicate.binding.invalid = true
		}
		p.markDeclarationIncomplete(duplicate.original)
		p.markDeclarationIncomplete(duplicate.declaration)
	}
}

func (p *spl2Program) bindAnnotation(stage *spl2SemanticStage, annotation spl2.IAnnotationContext) {
	if annotation == nil || annotation.Identifier() == nil || !spl2IntactSyntax(annotation) {
		return
	}
	name := spl2ProgramIdentifier(annotation.Identifier())
	location := p.parsed.source.contextLocation(annotation.Identifier())
	stage.referenceAt(location, name, "annotation", "read", "exact")
	if arguments := annotation.Arguments(); arguments != nil {
		for _, expression := range arguments.AllExpression() {
			stage.expression(expression)
		}
		for _, named := range arguments.AllNamedArgument() {
			stage.expression(named.Expression())
		}
	}
}

func (p *spl2Program) bindImport(stage *spl2SemanticStage, declaration *spl2ImportDeclaration) {
	if declaration.moduleCtx == nil || declaration.module == "" || !spl2IntactSyntax(declaration.moduleCtx) {
		return
	}
	declaration.moduleRef = stage.referenceAt(p.parsed.source.contextLocation(declaration.moduleCtx), declaration.module, "module", "read", "exact")
}

func (p *spl2Program) bindExports() {
	for _, declaration := range p.declarations {
		if len(declaration.exports) == 0 {
			continue
		}
		stage := p.declarationStage(declaration)
		for _, exported := range declaration.exports {
			stage.referenceAt(p.parsed.source.contextLocation(exported.aliasCtx), exported.alias, "symbol", "export", "exact")
			if p.functions[exported.local] != nil || p.views["$"+exported.local] != nil {
				continue
			}
			if binding := p.imports[exported.local]; binding != nil {
				p.useImport(stage, binding, exported.localCtx, "")
				continue
			}
			stage.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("exported symbol %q is unresolved", exported.local), p.parsed.source.contextLocation(exported.localCtx), true, nil)
		}
	}
}

func (p *spl2Program) reportUnusedImports() {
	for _, declaration := range p.declarations {
		if declaration.imp == nil {
			continue
		}
		used := false
		for _, binding := range declaration.imp.bindings {
			used = used || binding.used
		}
		if used {
			continue
		}
		stage := p.declarationStage(declaration)
		pending := []string{}
		if declaration.imp.moduleRef != "" {
			pending = append(pending, declaration.imp.moduleRef)
		}
		stage.diagnosticAtOwnedWithRequirementIncomplete(CodeUnresolvedModule, "warning", "unsupported_semantics", fmt.Sprintf("module %q is not resolved", declaration.imp.module), p.parsed.source.contextLocation(declaration.imp.moduleCtx), pending)
	}
}

func (p *spl2Program) markDeclarationInvalid(declaration *spl2ProgramDeclaration) {
	if declaration == nil {
		return
	}
	if declaration.view != nil {
		declaration.view.invalid = true
	}
	if declaration.function != nil {
		declaration.function.invalid = true
	}
}

func (p *spl2Program) markDeclarationIncomplete(declaration *spl2ProgramDeclaration) {
	if declaration == nil || declaration.stage < 0 || declaration.stage >= len(p.result.Stages) {
		return
	}
	p.result.Stages[declaration.stage].SemanticComplete = false
}

func (p *spl2Program) bindView(view *spl2ViewSymbol) (*environment, bool) {
	if view == nil || view.invalid || view.body == nil {
		if view != nil {
			p.markDeclarationIncomplete(view.declaration)
		}
		return nil, false
	}
	switch view.state {
	case spl2BindingComplete:
		if view.summary == nil {
			return nil, false
		}
		complete := !view.invalid && !view.parserTainted && !view.summary.uncertain && !view.summary.requirements.uncertain
		return view.summary.clone(), complete
	case spl2BindingVisiting:
		p.markViewCycle(view)
		return nil, false
	}
	view.state = spl2BindingVisiting
	p.viewStack = append(p.viewStack, view)
	scheduler := &spl2ScopeScheduler{
		result:                 p.result,
		parsed:                 p.parsed,
		refinement:             p.refinement,
		trace:                  p.trace,
		initialDiagnosticCount: p.parserDiagnosticCount,
		children:               spl2ChildScopesIn(p.parsed, []antlr.Tree{view.body}),
		executed:               map[int]bool{},
		program:                p,
	}
	lineageStart := len(p.result.Lineage)
	env := scheduler.pipeline(spl2Sites(spl2PipelineContexts(view.body), p.parsed.source), newEnvironmentWithRequirementTrace(p.trace), map[string]bool{}, view.scopeID, -1)
	scheduler.syncParserDiagnostics()
	p.assignParserDiagnosticsToOwner(view.declaration.stage, p.parsed.source.contextLocation(view.declaration.ctx))
	if p.claimedParserDiagnosticWithin(p.parsed.source.contextLocation(view.body)) {
		view.parserTainted = true
	}
	if view.parserTainted {
		p.forceParserTaintedViewIncomplete(view, env, lineageStart)
	}
	p.viewStack = p.viewStack[:len(p.viewStack)-1]
	view.state = spl2BindingComplete
	view.summary = env.clone()
	if view.cycle || view.invalid || view.parserTainted || view.summary.uncertain || view.summary.requirements.uncertain {
		p.markDeclarationIncomplete(view.declaration)
		view.summary.uncertain = true
		view.summary.requirements.uncertain = true
		return view.summary.clone(), false
	}
	return view.summary.clone(), true
}

func (p *spl2Program) claimedParserDiagnosticWithin(owner Location) bool {
	for i := 0; i < p.parserDiagnosticCount; i++ {
		diagnostic := p.result.Diagnostics[i]
		if diagnostic.StageID != "" && diagnostic.StageID != spl2DeferredProgramDiagnosticStage && diagnostic.Location.Start.Offset >= owner.Start.Offset && diagnostic.Location.End.Offset <= owner.End.Offset {
			return true
		}
	}
	return false
}

func (p *spl2Program) forceParserTaintedViewIncomplete(view *spl2ViewSymbol, env *environment, lineageStart int) {
	p.markDeclarationIncomplete(view.declaration)
	env.uncertain = true
	env.requirements.uncertain = true
	body := p.parsed.source.contextLocation(view.body)
	stages := make(map[string]*Stage, len(p.result.Stages))
	for i := range p.result.Stages {
		stage := &p.result.Stages[i]
		stages[stage.ID] = stage
	}
	tainted := false
	for i := lineageStart; i < len(p.result.Lineage); i++ {
		lineage := &p.result.Lineage[i]
		stage := stages[lineage.StageID]
		if stage == nil || stage.Location.Start.Offset < body.Start.Offset || stage.Location.End.Offset > body.End.Offset {
			continue
		}
		if !tainted && !stage.SemanticComplete {
			tainted = true
		}
		if tainted {
			stage.SemanticComplete = false
			lineage.After.Uncertain = true
		}
	}
}

func (p *spl2Program) markViewCycle(repeated *spl2ViewSymbol) {
	start := 0
	for i, candidate := range p.viewStack {
		if candidate == repeated {
			start = i
			break
		}
	}
	for _, view := range p.viewStack[start:] {
		if view.cycle {
			continue
		}
		view.cycle = true
		view.invalid = true
		stage := p.declarationStage(view.declaration)
		stage.diagnosticAtOwned(CodeDeclarationCycle, "error", "contract", fmt.Sprintf("view %q participates in a declaration cycle", view.name), p.parsed.source.contextLocation(view.declaration.ctx), true, nil)
	}
}

func (p *spl2Program) resolveViewSource(stage *spl2SemanticStage, parameter spl2.IDatasetParameterContext) bool {
	if parameter == nil || parameter.LOCAL() == nil {
		return false
	}
	name := parameter.LOCAL().GetText()
	view := p.views[name]
	if view == nil {
		return false
	}
	location := p.parsed.source.contextLocation(parameter)
	callerTrace := stage.env.requirements.trace
	stage.referenceAt(location, name, "view", "read", "exact")
	var canonicalBefore *requirementTrace
	if callerTrace != nil && p.trace != nil && callerTrace != p.trace {
		canonicalBefore = p.trace.clone()
	}
	summary, complete := p.bindView(view)
	if canonicalBefore != nil {
		callerTrace = rebaseRequirementTrace(canonicalBefore, p.trace, callerTrace)
		stage.env.requirements.trace = callerTrace
	}
	if summary != nil {
		stage.env = summary.cloneWithRequirementTrace(callerTrace)
	}
	if !complete {
		if view.parserTainted && len(p.viewStack) > 0 {
			p.viewStack[len(p.viewStack)-1].parserTainted = true
		}
		p.markStageIncomplete(stage)
	}
	return true
}

func (p *spl2Program) resolveImportedDataset(stage *spl2SemanticStage, dataset spl2.IDatasetContext) bool {
	if dataset == nil {
		return false
	}
	if identifier := dataset.Identifier(); identifier != nil {
		name := spl2ProgramIdentifier(identifier)
		if binding := p.imports[name]; binding != nil {
			p.useImport(stage, binding, identifier, "")
			return true
		}
	}
	if dotted := dataset.DottedDataset(); dotted != nil {
		base := spl2ProgramIdentifier(dotted.Identifier())
		binding := p.imports[base]
		if binding == nil || !binding.container {
			return false
		}
		parts := []string{}
		for _, identifier := range dotted.DatasetPath().AllIdentifier() {
			parts = append(parts, spl2ProgramIdentifier(identifier))
		}
		p.useImport(stage, binding, dotted, strings.Join(parts, "."))
		return true
	}
	return false
}

func (p *spl2Program) useImport(stage *spl2SemanticStage, binding *spl2ImportBinding, owner antlr.ParserRuleContext, suffix string, functionCall ...bool) string {
	if binding == nil || owner == nil {
		return ""
	}
	binding.used = true
	if binding.invalid {
		p.markStageIncomplete(stage)
		return ""
	}
	member := binding.remote
	if binding.container {
		member = suffix
	}
	name := binding.declaration.module
	if member != "" {
		name += "." + member
	}
	location := p.parsed.source.contextLocation(owner)
	kind := "module_member"
	// A wildcard container names a module namespace, not a function.
	if len(functionCall) != 0 && functionCall[0] && !binding.container {
		kind = "function"
	}
	reference := stage.referenceAt(location, name, kind, "read", "exact")
	pending := []string{}
	if binding.declaration.moduleRef != "" {
		pending = append(pending, binding.declaration.moduleRef)
	}
	pending = append(pending, reference)
	stage.diagnosticAtOwned(CodeUnresolvedModule, "warning", "unsupported_semantics", fmt.Sprintf("imported member %q cannot be resolved without its external module", name), location, true, pending)
	return reference
}

func (p *spl2Program) markStageIncomplete(stage *spl2SemanticStage) {
	if stage == nil || stage.stage < 0 || stage.stage >= len(stage.result.Stages) {
		return
	}
	stage.result.Stages[stage.stage].SemanticComplete = false
	stage.env.uncertain = true
	stage.env.requirements.uncertain = true
}

func (p *spl2Program) bindCommand(stage *spl2SemanticStage, context antlr.ParserRuleContext) bool {
	from, ok := context.(*spl2.FromCommandContext)
	if !ok || from.SqlFromClause() == nil || from.SqlFromClause().Dataset() == nil {
		return false
	}
	dataset := from.SqlFromClause().Dataset()
	localView := dataset.DatasetParameter() != nil && p.views[dataset.DatasetParameter().GetText()] != nil
	imported := p.importedDatasetBinding(dataset) != nil
	if !localView && !imported {
		return false
	}
	stage.applySource()
	clear(stage.aliases)
	if localView {
		p.resolveViewSource(stage, dataset.DatasetParameter())
	} else {
		p.resolveImportedDataset(stage, dataset)
	}
	if alias := from.SqlFromClause().SourceAlias(); alias != nil {
		operand := stage.operand(alias.Identifier())
		if operand.Sound {
			stage.aliases[operand.Name] = true
		}
	}
	return true
}

func (p *spl2Program) importedDatasetBinding(dataset spl2.IDatasetContext) *spl2ImportBinding {
	if dataset == nil {
		return nil
	}
	if dataset.Identifier() != nil {
		return p.imports[spl2ProgramIdentifier(dataset.Identifier())]
	}
	if dotted := dataset.DottedDataset(); dotted != nil {
		binding := p.imports[spl2ProgramIdentifier(dotted.Identifier())]
		if binding != nil && binding.container {
			return binding
		}
	}
	return nil
}

func (p *spl2Program) bindCall(stage *spl2SemanticStage, call spl2.ICallContext, aggregate bool) (spl2ExpressionEvidence, bool) {
	name := spl2ProgramIdentifier(call.Identifier())
	function := p.functions[name]
	imported := p.imports[name]
	if function == nil && imported == nil {
		return spl2ExpressionEvidence{}, false
	}
	arguments := []spl2.IExpressionContext{}
	namedArguments := []spl2.INamedArgumentContext{}
	if callArguments := call.Arguments(); callArguments != nil {
		arguments = append(arguments, callArguments.AllExpression()...)
		namedArguments = append(namedArguments, callArguments.AllNamedArgument()...)
	}
	if function != nil {
		return p.callLocalFunction(stage, call, function, arguments, namedArguments, aggregate), true
	}
	out := spl2ExpressionEvidence{ids: []string{}}
	for _, argument := range arguments {
		value := stage.expression(argument)
		out.ids = uniqueIDs(out.ids, value.ids)
	}
	for _, argument := range namedArguments {
		value := stage.expression(argument.Expression())
		out.ids = uniqueIDs(out.ids, value.ids)
	}
	if imported != nil {
		p.useImport(stage, imported, call.Identifier(), "", true)
		return out, true
	}
	return out, false
}

func (p *spl2Program) validateFunctions() {
	for _, declaration := range p.declarations {
		if declaration.function == nil || declaration.function.body == nil {
			continue
		}
		p.validateFunctionExpression(declaration.function, declaration.function.body, declaration.function.parameterSet)
	}
	stack := []*spl2FunctionSymbol{}
	var visit func(*spl2FunctionSymbol)
	visit = func(function *spl2FunctionSymbol) {
		if function == nil || function.state == spl2BindingComplete {
			return
		}
		if function.state == spl2BindingVisiting {
			p.markFunctionCycle(stack, function)
			return
		}
		function.state = spl2BindingVisiting
		stack = append(stack, function)
		for _, callee := range function.callees {
			visit(callee)
		}
		stack = stack[:len(stack)-1]
		function.state = spl2BindingComplete
	}
	for _, declaration := range p.declarations {
		visit(declaration.function)
	}
	p.propagateFunctionFailures()
	for _, declaration := range p.declarations {
		p.bindFunctionSummary(declaration.function)
	}
	p.propagateFunctionFailures()
}

func (p *spl2Program) bindFunctionSummary(function *spl2FunctionSymbol) bool {
	if function == nil {
		return false
	}
	if function.summaryState == spl2BindingComplete {
		return function.summary != nil && !function.invalid && !function.incomplete
	}
	if function.summaryState == spl2BindingVisiting {
		function.invalid = true
		p.markDeclarationIncomplete(function.declaration)
		return false
	}
	function.summaryState = spl2BindingVisiting
	for _, callee := range function.callees {
		if !p.bindFunctionSummary(callee) {
			function.invalid = function.invalid || callee.invalid
			function.incomplete = function.incomplete || callee.incomplete
		}
	}
	if function.invalid || function.incomplete || function.body == nil {
		function.summaryState = spl2BindingComplete
		p.markDeclarationIncomplete(function.declaration)
		return false
	}
	summary := newSPL2FunctionSummary(function)
	stage := p.declarationStage(function.declaration)
	stage.functionSummary = summary
	stage.suppressLocalCallReference = true
	diagnosticStart := len(p.result.Diagnostics)
	summary.evidence = stage.expression(function.body)
	for _, diagnostic := range p.result.Diagnostics[diagnosticStart:] {
		if diagnostic.Severity == "error" {
			function.invalid = true
		} else {
			function.incomplete = true
		}
	}
	function.summaryState = spl2BindingComplete
	if function.invalid || function.incomplete {
		p.markDeclarationIncomplete(function.declaration)
		return false
	}
	function.summary = summary
	return true
}

func (p *spl2Program) propagateFunctionFailures() {
	changed := true
	for changed {
		changed = false
		for _, declaration := range p.declarations {
			function := declaration.function
			if function == nil {
				continue
			}
			invalid, incomplete, parserTainted := function.invalid, function.incomplete, function.parserTainted
			for _, callee := range function.callees {
				invalid = invalid || callee.invalid
				incomplete = incomplete || callee.incomplete
				parserTainted = parserTainted || callee.parserTainted
			}
			if invalid != function.invalid || incomplete != function.incomplete || parserTainted != function.parserTainted {
				function.invalid = invalid
				function.incomplete = incomplete
				function.parserTainted = parserTainted
				p.markDeclarationIncomplete(declaration)
				changed = true
			}
		}
	}
}

func (p *spl2Program) validateFunctionExpression(function *spl2FunctionSymbol, node antlr.Tree, locals map[string]bool) {
	if node == nil {
		return
	}
	stage := p.declarationStage(function.declaration)
	switch context := node.(type) {
	case spl2.ICallContext:
		name := spl2ProgramIdentifier(context.Identifier())
		arguments := []spl2.IExpressionContext{}
		named := 0
		if context.Arguments() != nil {
			arguments = append(arguments, context.Arguments().AllExpression()...)
			named = len(context.Arguments().AllNamedArgument())
			for _, argument := range arguments {
				p.validateFunctionExpression(function, argument, locals)
			}
			for _, argument := range context.Arguments().AllNamedArgument() {
				p.validateFunctionExpression(function, argument.Expression(), locals)
			}
		}
		if local := p.functions[name]; local != nil {
			function.callees = appendUniqueSPL2Function(function.callees, local)
			stage.referenceAt(p.parsed.source.contextLocation(context.Identifier()), name, "function", "call", "exact")
			if named != 0 || len(arguments) != len(local.parameters) {
				stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", fmt.Sprintf("function %q requires %d positional arguments", name, len(local.parameters)), p.parsed.source.contextLocation(context), true, nil)
				function.invalid = true
			}
			return
		}
		if imported := p.imports[name]; imported != nil {
			p.useImport(stage, imported, context.Identifier(), "")
			function.incomplete = true
			return
		}
		policy, selected := spl2TypedPolicyFor(p.result.Document.Profile, p.result.Document.Version)
		spec, known := policy.functions[name]
		if !selected || !known {
			stage.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("function %q is unresolved", name), p.parsed.source.contextLocation(context.Identifier()), true, nil)
			function.invalid = true
			return
		}
		arity := len(arguments)
		if named != 0 || spec.signature.aggregate || spec.unmodeledScalar || arity < spec.signature.min || spec.signature.max >= 0 && arity > spec.signature.max || spec.pairedArguments && arity%2 != 0 {
			stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", fmt.Sprintf("function %q is not valid in a pure scalar declaration", name), p.parsed.source.contextLocation(context), true, nil)
			function.invalid = true
		}
		return
	case spl2.IAccessContext:
		primary := context.Primary()
		if primary == nil {
			return
		}
		if primary.LOCAL() != nil {
			name := primary.LOCAL().GetText()
			if !locals[name] {
				stage.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("local symbol %q is unresolved", name), p.parsed.source.contextLocation(primary), true, nil)
				function.invalid = true
			}
			if len(context.AllAccessPart()) > 0 {
				stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", "local function parameters must be used as scalar values", p.parsed.source.contextLocation(context), true, nil)
				function.invalid = true
			}
			for _, part := range context.AllAccessPart() {
				if part.Expression() != nil {
					p.validateFunctionExpression(function, part.Expression(), locals)
				}
			}
			return
		}
		if primary.Call() != nil {
			p.validateFunctionExpression(function, primary.Call(), locals)
			return
		}
		if field := primary.FieldName(); field != nil && field.Identifier() != nil {
			name := spl2ProgramIdentifier(field.Identifier())
			if imported := p.imports[name]; imported != nil {
				if imported.container && len(context.AllAccessPart()) > 0 {
					parts := []string{}
					static := true
					for _, part := range context.AllAccessPart() {
						if part.DOT() == nil || part.Identifier() == nil {
							static = false
							break
						}
						parts = append(parts, spl2ProgramIdentifier(part.Identifier()))
					}
					if static {
						p.useImport(stage, imported, context, strings.Join(parts, "."))
						function.incomplete = true
						return
					}
				}
				p.useImport(stage, imported, field.Identifier(), "")
				function.incomplete = true
				return
			}
			stage.diagnosticAtOwned(CodeUnresolvedSymbol, "error", "contract", fmt.Sprintf("free field %q is not allowed in a local scalar function", name), p.parsed.source.contextLocation(field), true, nil)
			function.invalid = true
			return
		}
	case spl2.ILambdaExpressionContext:
		nested := make(map[string]bool, len(locals)+len(context.AllLambdaParameter()))
		for name, bound := range locals {
			nested[name] = bound
		}
		for _, parameter := range context.AllLambdaParameter() {
			if parameter.LOCAL() != nil {
				nested[parameter.LOCAL().GetText()] = true
			}
		}
		if context.Expression() != nil {
			p.validateFunctionExpression(function, context.Expression(), nested)
		}
		if context.LambdaBlock() != nil {
			p.validateFunctionExpression(function, context.LambdaBlock(), nested)
		}
		return
	case spl2.IFieldNameContext, spl2.IIdentifierContext, spl2.IObjectKeyContext, spl2.ILambdaParameterContext:
		return
	case antlr.TerminalNode:
		return
	}
	for _, child := range node.GetChildren() {
		p.validateFunctionExpression(function, child, locals)
	}
}

func appendUniqueSPL2Function(functions []*spl2FunctionSymbol, candidate *spl2FunctionSymbol) []*spl2FunctionSymbol {
	for _, function := range functions {
		if function == candidate {
			return functions
		}
	}
	return append(functions, candidate)
}

func (p *spl2Program) markFunctionCycle(stack []*spl2FunctionSymbol, repeated *spl2FunctionSymbol) {
	start := 0
	for i, function := range stack {
		if function == repeated {
			start = i
			break
		}
	}
	for _, function := range stack[start:] {
		if function.cycle {
			continue
		}
		function.cycle = true
		function.invalid = true
		stage := p.declarationStage(function.declaration)
		stage.diagnosticAtOwned(CodeDeclarationCycle, "error", "contract", fmt.Sprintf("function %q participates in a declaration cycle", function.name), p.parsed.source.contextLocation(function.declaration.ctx), true, nil)
	}
}

func (p *spl2Program) callLocalFunction(stage *spl2SemanticStage, call spl2.ICallContext, function *spl2FunctionSymbol, arguments []spl2.IExpressionContext, namedArguments []spl2.INamedArgumentContext, aggregate bool) spl2ExpressionEvidence {
	out := spl2ExpressionEvidence{ids: []string{}}
	if function.parserTainted && len(p.viewStack) > 0 {
		p.viewStack[len(p.viewStack)-1].parserTainted = true
	}
	if !stage.suppressLocalCallReference {
		stage.referenceAt(p.parsed.source.contextLocation(call.Identifier()), function.name, "function", "call", "exact")
	}
	evaluateArguments := func() {
		for _, argument := range arguments {
			value := stage.expression(argument)
			out.ids = uniqueIDs(out.ids, value.ids)
		}
		for _, argument := range namedArguments {
			value := stage.expression(argument.Expression())
			out.ids = uniqueIDs(out.ids, value.ids)
		}
	}
	if len(namedArguments) != 0 || aggregate || len(arguments) != len(function.parameters) {
		evaluateArguments()
		stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", fmt.Sprintf("function %q requires %d positional scalar arguments", function.name, len(function.parameters)), p.parsed.source.contextLocation(call), true, nil)
		return out
	}
	if function.invalid || function.cycle || function.body == nil {
		evaluateArguments()
		stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", fmt.Sprintf("function %q has no valid scalar body", function.name), p.parsed.source.contextLocation(call), true, nil)
		return out
	}
	if function.incomplete {
		evaluateArguments()
		p.markStageIncomplete(stage)
		return out
	}
	if function.summary == nil {
		evaluateArguments()
		stage.diagnosticAtOwned(CodeInvalidFunctionCall, "error", "contract", fmt.Sprintf("function %q has no valid scalar body", function.name), p.parsed.source.contextLocation(call), true, nil)
		return out
	}
	bindings := make(map[spl2FunctionParameterUse]spl2ExpressionEvidence)
	for index, argument := range arguments {
		roles := function.summary.parameterRoles[index]
		if len(roles) == 0 {
			stage.expression(argument)
			continue
		}
		evidenceByRole := map[string]spl2ExpressionEvidence{}
		for _, role := range roles {
			effectiveRole := spl2FunctionArgumentRole(argument, role)
			evidence, ok := evidenceByRole[effectiveRole]
			if !ok {
				evidence = stage.expressionWithRole(argument, effectiveRole)
				evidenceByRole[effectiveRole] = evidence
			}
			bindings[spl2FunctionParameterUse{parameter: index, role: role}] = evidence
		}
	}
	out = function.summary.evidence.instantiate(bindings)
	out.ids = uniqueIDs(out.ids)
	return out
}

func spl2FunctionArgumentRole(argument spl2.IExpressionContext, role string) string {
	if role == "null_test" {
		access := spl2SQLFieldAccess(argument)
		direct := access != nil && len(access.AllAccessPart()) == 0 && (access.Primary().FieldName() != nil || access.Primary().LOCAL() != nil)
		if !direct {
			return "read"
		}
	}
	return role
}
