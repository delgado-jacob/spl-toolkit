package analysis

import (
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

func fillnullCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisFillnullStageContext)
	if !ok || stage.AnalysisFillnull() == nil {
		return
	}
	ctx := stage.AnalysisFillnull()
	if option := ctx.AnalysisFillnullValueOption(); option != nil && (option.AnalysisLiteral() == nil || !s.sound(option.AnalysisLiteral())) {
		s.diagnostic(CodeUnsupportedSemantics, "fillnull value requires an exact literal", option)
	}
	for _, identifier := range ctx.AllAnalysisIdentifier() {
		target := fieldCommandOperand(s, identifier)
		if target.Resolution != "exact" {
			s.diagnostic(CodeUnsupportedSemantics, "fillnull target must be an exact field", identifier)
			continue
		}

		field, semanticKnown := s.env.fields[target.Name]
		requirementField, requirementKnown := s.env.requirements.fields[target.Name]
		semanticConditional := !semanticKnown || field.Conditional
		requirementConditional := !requirementKnown || requirementField.conditional
		if semanticConditional && !semanticKnown {
			s.env.install(target.Name, nil, true)
		}
		if requirementConditional && !requirementKnown {
			s.env.requirements.install(target.Name, nil, true)
		}
		input := s.readAt(target, "read")
		s.createAtWithRequirementConditional(target, "create", "create", nonemptyReferenceIDs(input), true, requirementConditional)
	}
}

func rexCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisRexStageContext)
	if !ok || stage.AnalysisRex() == nil {
		return
	}
	ctx := stage.AnalysisRex()
	input := fieldCommandImplicitRaw(s, stage)
	inputExact := true
	fieldOptions := ctx.AllAnalysisRexFieldOption()
	if len(fieldOptions) > 0 {
		identifiers := fieldOptions[0].AllAnalysisIdentifier()
		if len(identifiers) == 2 && s.sound(identifiers[1]) {
			input = fieldCommandOperand(s, identifiers[1])
			inputExact = input.Resolution == "exact"
		} else {
			inputExact = false
		}
	}
	inputID := ""
	if inputExact {
		inputID = s.readAt(input, "read")
	} else {
		inputID = fieldCommandHeldRead(s, input, "read")
	}
	if !inputExact {
		s.diagnostic(CodeUnsupportedSemantics, "rex input must be an exact field", fieldCommandDiagnosticContext(fieldOptions, ctx))
	}
	for _, option := range fieldOptions[1:] {
		s.diagnostic(CodeUnsupportedSemantics, "rex accepts one input field option", option)
	}

	maxMatchOptions := ctx.AllAnalysisRexMaxMatchOption()
	for i, option := range maxMatchOptions {
		if i > 0 || option.AnalysisLiteral() == nil || !s.sound(option.AnalysisLiteral()) {
			s.diagnostic(CodeUnsupportedSemantics, "rex max_match requires one exact literal", option)
		}
	}

	modeledExtraction := true
	modeOptions := ctx.AllAnalysisRexModeOption()
	if len(modeOptions) > 0 {
		modeledExtraction = false
		mode := ""
		identifiers := modeOptions[0].AllAnalysisIdentifier()
		if len(identifiers) == 2 && s.sound(identifiers[1]) {
			mode = strings.ToLower(normalizedName(identifiers[1].GetText()))
		}
		message := "rex mode is unmodeled"
		if mode == "sed" {
			message = "rex sed value rewriting is unmodeled"
		}
		s.diagnostic(CodeUnsupportedSemantics, message, modeOptions[0])
		for _, option := range modeOptions[1:] {
			s.diagnostic(CodeUnsupportedSemantics, "rex accepts one mode option", option)
		}
	}

	offsetOptions := ctx.AllAnalysisRexOffsetFieldOption()
	if len(offsetOptions) > 1 {
		for _, option := range offsetOptions[1:] {
			s.diagnostic(CodeUnsupportedSemantics, "rex accepts one offset_field option", option)
		}
	}
	if len(offsetOptions) > 0 && modeledExtraction && inputExact {
		identifiers := offsetOptions[0].AllAnalysisIdentifier()
		if len(identifiers) == 2 && s.sound(identifiers[1]) {
			target := fieldCommandOperand(s, identifiers[1])
			if target.Resolution == "exact" {
				s.createAt(target, "output", "create", nonemptyReferenceIDs(inputID), true)
			} else {
				s.diagnostic(CodeUnsupportedSemantics, "rex offset_field must be an exact field", identifiers[1])
			}
		} else {
			s.diagnostic(CodeUnsupportedSemantics, "rex offset_field must be an exact field", offsetOptions[0])
		}
	}

	if !modeledExtraction || !inputExact || ctx.STRING() == nil {
		return
	}
	pattern, start := fieldCommandStringContents(ctx.STRING())
	captures, understood := scanRexNamedCaptures(pattern)
	if !understood {
		s.diagnosticAt(CodeUnsupportedSemantics, "warning", "unsupported_semantics", "rex named capture declaration is ambiguous or malformed", fieldCommandTerminalLocation(s, ctx.STRING()), true)
		return
	}
	for _, capture := range captures {
		location := s.parsed.source.location(start+capture.start, start+capture.end)
		target := locatedOperand{Name: capture.name, Location: location, Resolution: "exact", Sound: true}
		s.createAt(target, "output", "create", nonemptyReferenceIDs(inputID), true)
	}
}

func spathCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisSpathStageContext)
	if !ok || stage.AnalysisSpath() == nil {
		return
	}
	ctx := stage.AnalysisSpath()
	input := fieldCommandImplicitRaw(s, stage)
	inputExact := true
	inputOptions := ctx.AllAnalysisSpathInputOption()
	if len(inputOptions) > 0 {
		identifiers := inputOptions[0].AllAnalysisIdentifier()
		if len(identifiers) == 2 && s.sound(identifiers[1]) {
			input = fieldCommandOperand(s, identifiers[1])
			inputExact = input.Resolution == "exact"
		} else {
			inputExact = false
		}
	}
	inputID := ""
	if inputExact {
		inputID = s.readAt(input, "read")
	} else {
		inputID = fieldCommandHeldRead(s, input, "read")
	}
	if !inputExact {
		s.diagnostic(CodeUnsupportedSemantics, "spath input must be an exact field", fieldCommandDiagnosticContext(inputOptions, ctx))
	}
	for _, option := range inputOptions[1:] {
		s.diagnostic(CodeUnsupportedSemantics, "spath accepts one input option", option)
	}

	pathExact := false
	pathOptions := ctx.AllAnalysisSpathPathOption()
	if len(pathOptions) == 0 {
		s.diagnostic(CodeUnsupportedSemantics, "spath auto-extraction has an unknown output set", ctx)
	} else {
		option := pathOptions[0]
		value := option.AnalysisOptionValue()
		pathExact = value != nil && value.AnalysisLiteral() != nil && s.sound(value.AnalysisLiteral())
		if !pathExact {
			s.diagnostic(CodeUnsupportedSemantics, "spath path requires an exact literal", option)
		}
		for _, duplicate := range pathOptions[1:] {
			s.diagnostic(CodeUnsupportedSemantics, "spath accepts one path option", duplicate)
		}
	}

	outputExact := false
	var output locatedOperand
	outputOptions := ctx.AllAnalysisSpathOutputOption()
	if len(outputOptions) == 0 {
		s.diagnostic(CodeUnsupportedSemantics, "spath requires an explicit output field", ctx)
	} else {
		identifier := outputOptions[0].AnalysisIdentifier()
		if identifier != nil && s.sound(identifier) {
			output = fieldCommandOperand(s, identifier)
			outputExact = output.Resolution == "exact"
		}
		if !outputExact {
			s.diagnostic(CodeUnsupportedSemantics, "spath output must be an exact field", outputOptions[0])
		}
		for _, duplicate := range outputOptions[1:] {
			s.diagnostic(CodeUnsupportedSemantics, "spath accepts one output option", duplicate)
		}
	}
	if inputExact && pathExact && outputExact && len(inputOptions) <= 1 && len(pathOptions) == 1 && len(outputOptions) == 1 {
		s.createAt(output, "output", "create", nonemptyReferenceIDs(inputID), true)
	}
}

func binCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisBinStageContext)
	if !ok || stage.AnalysisBin() == nil {
		return
	}
	ctx := stage.AnalysisBin()
	if ctx.AnalysisIdentifier() == nil {
		return
	}
	input := fieldCommandOperand(s, ctx.AnalysisIdentifier())
	modeled := input.Resolution == "exact"
	inputID := ""
	if modeled {
		inputID = s.readAt(input, "read")
	} else {
		inputID = fieldCommandHeldRead(s, input, "read")
	}
	if !modeled {
		s.diagnostic(CodeUnsupportedSemantics, "bin input must be an exact field", ctx.AnalysisIdentifier())
	}
	for _, option := range ctx.AllAnalysisBinOption() {
		name := ""
		if option.AnalysisIdentifier() != nil {
			name = strings.ToLower(normalizedName(option.AnalysisIdentifier().GetText()))
		}
		literal := false
		switch name {
		case "span", "minspan":
			value := option.AnalysisUnitOptionValue()
			literal = value != nil && value.NUMBER() != nil && s.sound(value)
		case "bins":
			value := option.AnalysisOptionValue()
			if value != nil && value.AnalysisLiteral() != nil && value.AnalysisLiteral().NUMBER() != nil && s.sound(value.AnalysisLiteral()) {
				count, err := strconv.Atoi(value.AnalysisLiteral().NUMBER().GetText())
				literal = err == nil && count > 0
			}
		case "start", "end":
			value := option.AnalysisOptionValue()
			literal = value != nil && value.AnalysisLiteral() != nil && value.AnalysisLiteral().NUMBER() != nil && s.sound(value.AnalysisLiteral())
		case "aligntime":
			value := option.AnalysisOptionValue()
			if value != nil && s.sound(value) {
				switch {
				case value.AnalysisLiteral() != nil:
					literal = value.AnalysisLiteral().TIME() != nil
				case value.AnalysisIdentifier() != nil:
					alignment := strings.ToLower(normalizedName(value.AnalysisIdentifier().GetText()))
					literal = alignment == "earliest" || alignment == "latest"
				}
			}
		}
		if !literal {
			modeled = false
			s.diagnostic(CodeUnsupportedSemantics, "bin option requires a supported exact literal", option)
		}
	}
	target := input
	if alias := ctx.AnalysisAlias(); alias != nil {
		if alias.AnalysisIdentifier() == nil {
			modeled = false
		} else {
			target = fieldCommandOperand(s, alias.AnalysisIdentifier())
			if target.Resolution != "exact" {
				modeled = false
				s.diagnostic(CodeUnsupportedSemantics, "bin alias must be an exact field", alias.AnalysisIdentifier())
			}
		}
	}
	if modeled {
		s.applyAssignment(target, nonemptyReferenceIDs(inputID), false, false)
	}
}

func regexCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisRegexStageContext)
	if !ok || stage.AnalysisRegex() == nil {
		return
	}
	ctx := stage.AnalysisRegex()
	input := fieldCommandImplicitRaw(s, stage)
	if ctx.AnalysisIdentifier() != nil {
		input = fieldCommandOperand(s, ctx.AnalysisIdentifier())
	}
	if input.Resolution == "exact" {
		s.readAt(input, "filter")
	} else {
		fieldCommandHeldRead(s, input, "filter")
		s.diagnostic(CodeUnsupportedSemantics, "regex input must be an exact field", ctx.AnalysisIdentifier())
	}
}

func mvexpandCommand(s *semanticStage, node antlr.ParserRuleContext) {
	stage, ok := node.(*parser.AnalysisMvexpandStageContext)
	if !ok || stage.AnalysisMvexpand() == nil {
		return
	}
	ctx := stage.AnalysisMvexpand()
	if ctx.AnalysisIdentifier() == nil {
		return
	}
	input := fieldCommandOperand(s, ctx.AnalysisIdentifier())
	if input.Resolution == "exact" {
		s.readAt(input, "read")
	} else {
		fieldCommandHeldRead(s, input, "read")
		s.diagnostic(CodeUnsupportedSemantics, "mvexpand input must be an exact field", ctx.AnalysisIdentifier())
	}
	for _, option := range ctx.AllAnalysisMvexpandOption() {
		s.diagnostic(CodeUnsupportedSemantics, "mvexpand options are unmodeled", option)
	}
}

type rexCapture struct {
	name       string
	start, end int
}

func rexNamedCaptures(pattern string) ([]string, bool) {
	captures, ok := scanRexNamedCaptures(pattern)
	if !ok {
		return nil, false
	}
	names := make([]string, 0, len(captures))
	for _, capture := range captures {
		names = append(names, capture.name)
	}
	return names, true
}

func scanRexNamedCaptures(pattern string) ([]rexCapture, bool) {
	runes := []rune(pattern)
	captures := []rexCapture{}
	seen := map[string]bool{}
	escaped, inClass, inPOSIXClass, inQuotedLiteral := false, false, false, false
	classCanClose, classMayNegate := false, false
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if inQuotedLiteral {
			if r == '\\' && i+1 < len(runes) && runes[i+1] == 'E' {
				inQuotedLiteral = false
				i++
			}
			continue
		}
		if escaped {
			escaped = false
			if inClass {
				classCanClose = true
				classMayNegate = false
			}
			continue
		}
		if r == '\\' {
			if !inClass && i+1 < len(runes) && runes[i+1] == 'Q' {
				inQuotedLiteral = true
				i++
				continue
			}
			escaped = true
			continue
		}
		if inClass {
			if inPOSIXClass {
				if r == ':' && i+1 < len(runes) && runes[i+1] == ']' {
					inPOSIXClass = false
					classCanClose = true
					classMayNegate = false
					i++
				}
				continue
			}
			if r == '[' && i+1 < len(runes) && runes[i+1] == ':' {
				inPOSIXClass = true
				classCanClose = true
				classMayNegate = false
				i++
				continue
			}
			if r == ']' && classCanClose {
				inClass = false
				classCanClose = false
				classMayNegate = false
				continue
			}
			if classMayNegate && r == '^' {
				classMayNegate = false
				continue
			}
			classCanClose = true
			classMayNegate = false
			continue
		}
		if r == '[' {
			inClass = true
			classCanClose = false
			classMayNegate = true
			continue
		}
		if r == '(' && i+2 < len(runes) && runes[i+1] == '?' && runes[i+2] == '#' {
			commentEnd, commentEscaped := i+3, false
			for ; commentEnd < len(runes); commentEnd++ {
				if commentEscaped {
					commentEscaped = false
					continue
				}
				if runes[commentEnd] == '\\' {
					commentEscaped = true
					continue
				}
				if runes[commentEnd] == ')' {
					break
				}
			}
			if commentEnd == len(runes) {
				return nil, false
			}
			i = commentEnd
			continue
		}
		if r != '(' || i+2 >= len(runes) || runes[i+1] != '?' {
			continue
		}

		nameStart := -1
		switch {
		case runes[i+2] == '<':
			if i+3 < len(runes) && (runes[i+3] == '=' || runes[i+3] == '!') {
				continue
			}
			nameStart = i + 3
		case runes[i+2] == 'P':
			if i+3 >= len(runes) {
				return nil, false
			}
			switch runes[i+3] {
			case '<':
				nameStart = i + 4
			case '=':
				nameEnd := i + 4
				for nameEnd < len(runes) && runes[nameEnd] != ')' {
					nameEnd++
				}
				if nameEnd == len(runes) || !rewriteSPLBare(string(runes[i+4:nameEnd])) {
					return nil, false
				}
				i = nameEnd
				continue
			default:
				return nil, false
			}
		default:
			continue
		}

		nameEnd := nameStart
		for nameEnd < len(runes) && runes[nameEnd] != '>' {
			nameEnd++
		}
		if nameEnd == len(runes) {
			return nil, false
		}
		name := string(runes[nameStart:nameEnd])
		if !rewriteSPLBare(name) {
			return nil, false
		}
		if !seen[name] {
			captures = append(captures, rexCapture{name: name, start: nameStart, end: nameEnd})
			seen[name] = true
		}
		i = nameEnd
	}
	if inPOSIXClass || inQuotedLiteral {
		return nil, false
	}
	return captures, true
}

func fieldCommandHeldRead(s *semanticStage, operand locatedOperand, role string) string {
	if !operand.Sound || operand.Resolution == "exact" {
		return ""
	}
	id := s.operandReference(operand, "field", role)
	if id == "" {
		return ""
	}
	reference := &s.result.References[len(s.result.References)-1]
	reference.Binding = "indeterminate"
	s.rewriteBinding(id, reference.Binding, nil)
	return id
}

func fieldCommandOperand(s *semanticStage, ctx parser.IAnalysisIdentifierContext) locatedOperand {
	name := normalizedName(ctx.GetText())
	operand := s.operand(ctx, name)
	if strings.Contains(name, "*") {
		operand.Resolution = "dynamic"
	}
	return operand
}

func fieldCommandImplicitRaw(s *semanticStage, node interface {
	AnalysisCommandName() parser.IAnalysisCommandNameContext
}) locatedOperand {
	return s.operand(node.AnalysisCommandName(), "_raw")
}

func fieldCommandDiagnosticContext[T antlr.ParserRuleContext](contexts []T, fallback antlr.ParserRuleContext) antlr.ParserRuleContext {
	if len(contexts) > 0 {
		return contexts[0]
	}
	return fallback
}

func nonemptyReferenceIDs(ids ...string) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

func fieldCommandStringContents(node antlr.TerminalNode) (string, int) {
	raw := node.GetText()
	start := node.GetSymbol().GetStart()
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		return raw[1 : len(raw)-1], start + 1
	}
	return raw, start
}

func fieldCommandTerminalLocation(s *semanticStage, node antlr.TerminalNode) Location {
	token := node.GetSymbol()
	return s.parsed.source.location(token.GetStart(), token.GetStop()+1)
}
