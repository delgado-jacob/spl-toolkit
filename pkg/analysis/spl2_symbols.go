package analysis

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser/spl2"
)

type spl2DeclarationKind string

const (
	spl2ViewDeclarationKind       spl2DeclarationKind = "view"
	spl2FunctionDeclarationKind   spl2DeclarationKind = "function"
	spl2ImportDeclarationKind     spl2DeclarationKind = "import"
	spl2ExportDeclarationKind     spl2DeclarationKind = "export"
	spl2AnnotationDeclarationKind spl2DeclarationKind = "annotation"
)

type spl2ProgramDeclaration struct {
	kind        spl2DeclarationKind
	name        string
	ctx         antlr.ParserRuleContext
	annotations []spl2.IAnnotationContext
	stage       int
	ordinal     int
	view        *spl2ViewSymbol
	function    *spl2FunctionSymbol
	imp         *spl2ImportDeclaration
	exports     []*spl2ExportEntry
}

type spl2ViewSymbol struct {
	name          string
	declaration   *spl2ProgramDeclaration
	body          spl2.IPipelineContext
	scopeID       string
	state         spl2BindingState
	summary       *environment
	sourceInputs  []inputFact // All underlying query sources, separate from active row owners.
	cycle         bool
	invalid       bool
	parserTainted bool
}

type spl2FunctionSymbol struct {
	name           string
	declaration    *spl2ProgramDeclaration
	body           spl2.IExpressionContext
	parameters     []string
	parameterSet   map[string]bool
	parameterIndex map[string]int
	callees        []*spl2FunctionSymbol
	scopeID        string
	state          spl2BindingState
	summaryState   spl2BindingState
	summary        *spl2FunctionSummary
	cycle          bool
	invalid        bool
	incomplete     bool
	parserTainted  bool
}

type spl2FunctionSummary struct {
	function       *spl2FunctionSymbol
	parameterRoles [][]string
	evidence       spl2ExpressionEvidence
}

// A summary is populated only while its declaration is bound. Caller stages
// instantiate its parameter roles and evidence template without walking the body.
func newSPL2FunctionSummary(function *spl2FunctionSymbol) *spl2FunctionSummary {
	return &spl2FunctionSummary{function: function, parameterRoles: make([][]string, len(function.parameters))}
}

func (s *spl2FunctionSummary) parameter(name, role string) (spl2ExpressionEvidence, bool) {
	index, ok := s.function.parameterIndex[name]
	if !ok {
		return spl2ExpressionEvidence{}, false
	}
	seen := false
	for _, existing := range s.parameterRoles[index] {
		seen = seen || existing == role
	}
	if !seen {
		s.parameterRoles[index] = append(s.parameterRoles[index], role)
	}
	use := spl2FunctionParameterUse{parameter: index, role: role}
	return spl2ExpressionEvidence{
		ids:     []string{},
		modeled: true,
		template: func(bindings map[spl2FunctionParameterUse]spl2ExpressionEvidence) spl2ExpressionEvidence {
			value, ok := bindings[use]
			if !ok {
				return spl2ExpressionEvidence{ids: []string{}, modeled: false}
			}
			return value
		},
	}, true
}

type spl2ImportDeclaration struct {
	declaration *spl2ProgramDeclaration
	module      string
	moduleCtx   spl2.IQualifiedNameContext
	bindings    []*spl2ImportBinding
	moduleRef   string
}

type spl2ImportBinding struct {
	declaration *spl2ImportDeclaration
	local       string
	remote      string
	container   bool
	invalid     bool
	ctx         antlr.ParserRuleContext
	used        bool
}

type spl2ExportEntry struct {
	local    string
	alias    string
	localCtx antlr.ParserRuleContext
	aliasCtx antlr.ParserRuleContext
}

type spl2DuplicateSymbol struct {
	name            string
	declaration     *spl2ProgramDeclaration
	original        *spl2ProgramDeclaration
	binding         *spl2ImportBinding
	originalBinding *spl2ImportBinding
	location        Location
}

type spl2BindingState uint8

const (
	spl2BindingUnvisited spl2BindingState = iota
	spl2BindingVisiting
	spl2BindingComplete
)

type spl2Program struct {
	parsed                *spl2ParsedDocument
	declarations          []*spl2ProgramDeclaration
	views                 map[string]*spl2ViewSymbol
	functions             map[string]*spl2FunctionSymbol
	imports               map[string]*spl2ImportBinding
	duplicates            []spl2DuplicateSymbol
	result                *Result
	refinement            *sourceRefinement
	trace                 *requirementTrace
	parserDiagnosticCount int
	viewStack             []*spl2ViewSymbol
}

func collectSPL2Program(parsed *spl2ParsedDocument) *spl2Program {
	program := &spl2Program{
		parsed:     parsed,
		views:      map[string]*spl2ViewSymbol{},
		functions:  map[string]*spl2FunctionSymbol{},
		imports:    map[string]*spl2ImportBinding{},
		duplicates: []spl2DuplicateSymbol{},
	}
	if parsed == nil || parsed.tree == nil || parsed.tree.ModuleDeclaration() == nil {
		return program
	}
	valueSymbols := map[string]*spl2ProgramDeclaration{}
	for ordinal, statement := range parsed.tree.ModuleDeclaration().AllAnnotatedStatement() {
		annotations := []spl2.IAnnotationContext{}
		if statement.Annotations() != nil {
			annotations = append(annotations, statement.Annotations().AllAnnotation()...)
		}
		module := statement.ModuleStatement()
		if module == nil {
			declaration := &spl2ProgramDeclaration{kind: spl2AnnotationDeclarationKind, ctx: statement, annotations: annotations, ordinal: ordinal, stage: -1}
			program.declarations = append(program.declarations, declaration)
			continue
		}
		switch {
		case module.ViewDeclaration() != nil:
			ctx := module.ViewDeclaration()
			name := ""
			if ctx.LOCAL() != nil {
				name = ctx.LOCAL().GetText()
			}
			declaration := &spl2ProgramDeclaration{kind: spl2ViewDeclarationKind, name: name, ctx: ctx, annotations: annotations, ordinal: ordinal, stage: -1}
			view := &spl2ViewSymbol{name: name, declaration: declaration, body: ctx.Pipeline()}
			declaration.view = view
			program.declarations = append(program.declarations, declaration)
			if prior := program.views[name]; prior != nil {
				program.duplicates = append(program.duplicates, spl2DuplicateSymbol{name: name, declaration: declaration, original: prior.declaration, location: parsed.source.location(ctx.LOCAL().GetSymbol().GetStart(), ctx.LOCAL().GetSymbol().GetStop()+1)})
				view.invalid = true
			} else if name != "" {
				program.views[name] = view
			}
		case module.FunctionDeclaration() != nil:
			ctx := module.FunctionDeclaration()
			name := spl2ProgramIdentifier(ctx.Identifier())
			declaration := &spl2ProgramDeclaration{kind: spl2FunctionDeclarationKind, name: name, ctx: ctx, annotations: annotations, ordinal: ordinal, stage: -1}
			function := &spl2FunctionSymbol{name: name, declaration: declaration, parameterSet: map[string]bool{}, parameterIndex: map[string]int{}}
			if ctx.ReturnStatement() != nil {
				function.body = ctx.ReturnStatement().Expression()
			}
			if parameters := ctx.FunctionParameters(); parameters != nil {
				for index, parameter := range parameters.AllFunctionParameter() {
					name := ""
					if parameter.LOCAL() != nil {
						name = parameter.LOCAL().GetText()
					}
					function.parameters = append(function.parameters, name)
					if function.parameterSet[name] {
						program.duplicates = append(program.duplicates, spl2DuplicateSymbol{name: name, declaration: declaration, location: parsed.source.contextLocation(parameter)})
						function.invalid = true
					}
					function.parameterSet[name] = true
					function.parameterIndex[name] = index
				}
			}
			declaration.function = function
			program.declarations = append(program.declarations, declaration)
			if prior := valueSymbols[name]; prior != nil {
				program.duplicates = append(program.duplicates, spl2DuplicateSymbol{name: name, declaration: declaration, original: prior, originalBinding: program.imports[name], location: parsed.source.contextLocation(ctx.Identifier())})
				function.invalid = true
			} else if name != "" {
				valueSymbols[name] = declaration
				program.functions[name] = function
			}
		case module.ImportDeclaration() != nil:
			ctx := module.ImportDeclaration()
			declaration := &spl2ProgramDeclaration{kind: spl2ImportDeclarationKind, ctx: ctx, annotations: annotations, ordinal: ordinal, stage: -1}
			imp := &spl2ImportDeclaration{declaration: declaration, module: spl2QualifiedProgramName(ctx.QualifiedName()), moduleCtx: ctx.QualifiedName()}
			declaration.imp = imp
			program.declarations = append(program.declarations, declaration)
			for _, binding := range spl2ImportBindings(imp, ctx.ImportSelection()) {
				imp.bindings = append(imp.bindings, binding)
				if prior := valueSymbols[binding.local]; prior != nil {
					program.duplicates = append(program.duplicates, spl2DuplicateSymbol{name: binding.local, declaration: declaration, original: prior, binding: binding, originalBinding: program.imports[binding.local], location: parsed.source.contextLocation(binding.ctx)})
					binding.invalid = true
					continue
				}
				valueSymbols[binding.local] = declaration
				program.imports[binding.local] = binding
			}
		case module.ExportDeclaration() != nil:
			ctx := module.ExportDeclaration()
			declaration := &spl2ProgramDeclaration{kind: spl2ExportDeclarationKind, ctx: ctx, annotations: annotations, ordinal: ordinal, stage: -1, exports: spl2ExportEntries(ctx.ExportSelection())}
			program.declarations = append(program.declarations, declaration)
		default:
			declaration := &spl2ProgramDeclaration{kind: spl2AnnotationDeclarationKind, ctx: module, annotations: annotations, ordinal: ordinal, stage: -1}
			program.declarations = append(program.declarations, declaration)
		}
	}
	return program
}

func spl2ProgramIdentifier(identifier spl2.IIdentifierContext) string {
	if identifier == nil {
		return ""
	}
	name, ok := spl2DecodeKey(identifier.GetText())
	if !ok {
		return ""
	}
	return name
}

func spl2QualifiedProgramName(name spl2.IQualifiedNameContext) string {
	if name == nil {
		return ""
	}
	var normalized strings.Builder
	for _, child := range name.GetChildren() {
		switch part := child.(type) {
		case spl2.IIdentifierContext:
			segment := spl2ProgramIdentifier(part)
			if segment == "" {
				return ""
			}
			normalized.WriteString(segment)
		case antlr.TerminalNode:
			if text := part.GetText(); text == "/" || text == "." {
				normalized.WriteString(text)
			}
		}
	}
	return normalized.String()
}

func spl2ImportBindings(declaration *spl2ImportDeclaration, selection spl2.IImportSelectionContext) []*spl2ImportBinding {
	if selection == nil {
		return nil
	}
	if wildcard := selection.ImportWildcard(); wildcard != nil {
		name := spl2ProgramIdentifier(wildcard.Identifier())
		return []*spl2ImportBinding{{declaration: declaration, local: name, container: true, ctx: wildcard}}
	}
	imports := []spl2.IAliasedImportContext{}
	if list := selection.ImportList(); list != nil {
		imports = append(imports, list.AllAliasedImport()...)
	} else if single := selection.AliasedImport(); single != nil {
		imports = append(imports, single)
	}
	out := make([]*spl2ImportBinding, 0, len(imports))
	for _, imported := range imports {
		identifiers := imported.AllIdentifier()
		if len(identifiers) == 0 {
			continue
		}
		remote := spl2ProgramIdentifier(identifiers[0])
		local := remote
		if len(identifiers) > 1 {
			local = spl2ProgramIdentifier(identifiers[1])
		}
		out = append(out, &spl2ImportBinding{declaration: declaration, local: local, remote: remote, ctx: imported})
	}
	return out
}

func spl2ExportEntries(selection spl2.IExportSelectionContext) []*spl2ExportEntry {
	if selection == nil {
		return nil
	}
	exports := []spl2.IAliasedExportContext{}
	if list := selection.ExportList(); list != nil {
		exports = append(exports, list.AllAliasedExport()...)
	} else if single := selection.AliasedExport(); single != nil {
		exports = append(exports, single)
	}
	out := make([]*spl2ExportEntry, 0, len(exports))
	for _, exported := range exports {
		identifiers := exported.AllIdentifier()
		if len(identifiers) == 0 {
			continue
		}
		local := spl2ProgramIdentifier(identifiers[0])
		alias := local
		aliasCtx := antlr.ParserRuleContext(identifiers[0])
		if len(identifiers) > 1 {
			alias = spl2ProgramIdentifier(identifiers[1])
			aliasCtx = identifiers[1]
		}
		out = append(out, &spl2ExportEntry{local: local, alias: alias, localCtx: identifiers[0], aliasCtx: aliasCtx})
	}
	return out
}
