package closure

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/antlr4-go/antlr/v4"
	"github.com/delgado-jacob/spl-toolkit/parser"
)

// macroSpan is a half-open UTF-8 byte range in the original SPL text.
type macroSpan struct{ Start, End int }

// macroArgument retains the source spelling of a positional or named argument.
// A positional argument has an empty Name and zero NameSpan.
type macroArgument struct {
	Span      macroSpan
	Name      string
	NameSpan  macroSpan
	ValueSpan macroSpan
}

// macroInvocation describes one unquoted backtick form. Unsupported is nonempty
// when the invocation cannot safely be parsed or placed for expansion.
type macroInvocation struct {
	Span        macroSpan
	Name        string
	NameSpan    macroSpan
	Arguments   []macroArgument
	Nested      []macroInvocation // calls inside argument expressions, in source order
	Position    string            // stage or fragment
	Unsupported string
}

// scanMacroInvocations scans classic SPL without relying on analyzer macro
// diagnostics, which report unresolved definitions rather than expansion shape.
func scanMacroInvocations(source string) []macroInvocation {
	var calls []macroInvocation
	for i := 0; i < len(source); {
		switch source[i] {
		case '\'', '"':
			i = macroQuotedEnd(source, i)
		case '/':
			if end, ok := macroCommentEnd(source, i); ok {
				i = end
			} else {
				i++
			}
		case '`':
			call := scanMacroAt(source, i)
			calls = append(calls, call)
			i = call.Span.End
		default:
			i++
		}
	}
	return calls
}

func macroQuotedEnd(source string, start int) int {
	quote := source[start]
	for i := start + 1; i < len(source); i++ {
		if source[i] == '\\' && i+1 < len(source) {
			i++
			continue
		}
		if source[i] == quote {
			return i + 1
		}
	}
	return len(source)
}

func macroCommentEnd(source string, start int) (int, bool) {
	if start+1 >= len(source) || source[start] != '/' {
		return 0, false
	}
	switch source[start+1] {
	case '/':
		for i := start + 2; i < len(source); i++ {
			if source[i] == '\n' || source[i] == '\r' {
				return i, true
			}
		}
		return len(source), true
	case '*':
		for i := start + 2; i+1 < len(source); i++ {
			if source[i] == '*' && source[i+1] == '/' {
				return i + 2, true
			}
		}
		return len(source), true
	}
	return 0, false
}

func scanMacroAt(source string, start int) macroInvocation {
	var nested []macroInvocation
	parens, brackets := 0, 0
	triviaBeforeTick := false
	for i := start + 1; i < len(source); {
		if source[i] == '\'' || source[i] == '"' {
			i = macroQuotedEnd(source, i)
			triviaBeforeTick = false
			continue
		}
		if source[i] == '/' {
			if end, ok := macroCommentEnd(source, i); ok {
				i = end
				triviaBeforeTick = true
				continue
			}
		}
		r, width := utf8.DecodeRuneInString(source[i:])
		if unicode.IsSpace(r) {
			i += width
			triviaBeforeTick = true
			continue
		}
		switch source[i] {
		case '(':
			parens++
		case ')':
			if parens > 0 {
				parens--
			}
		case '[':
			brackets++
		case ']':
			if brackets > 0 {
				brackets--
			}
		case '|':
			// A pipe outside a subquery cannot occur in a macro argument
			// expression. Stop before it so a later stage can still be scanned.
			if brackets == 0 {
				return macroUnclosedBeforeNested(start, i, nested)
			}
		case '`':
			trial := macroParseForm(source, macroSpan{start, i + 1}, nested)
			if trial.Unsupported == "" {
				return trial
			}
			boundary := triviaBeforeTick
			if parens > 0 || boundary {
				child := scanMacroAt(source, i)
				if child.Unsupported == "" {
					if parens > 0 {
						nested = append(nested, child)
						i = child.Span.End
						triviaBeforeTick = false
						continue
					}
					// The previous form is malformed and this is a separate
					// invocation at a lexical boundary in the same stage.
					return macroUnclosedBeforeNested(start, i, nested)
				}
				if parens > 0 && boundary {
					return macroUnclosedBeforeNested(start, i, nested)
				}
			}
			if len(nested) > 0 {
				return macroUnclosedBeforeNested(start, i, nested)
			}
			return trial
		}
		i += width
		triviaBeforeTick = false
	}
	return macroUnclosedBeforeNested(start, len(source), nested)
}

func macroUnclosedBeforeNested(start, end int, nested []macroInvocation) macroInvocation {
	if len(nested) > 0 && nested[0].Span.Start < end {
		end = nested[0].Span.Start
	}
	return macroInvocation{
		Span:        macroSpan{start, end},
		Unsupported: "unclosed macro invocation",
	}
}

func macroParseForm(source string, span macroSpan, nested []macroInvocation) macroInvocation {
	call := macroInvocation{Span: span}
	bodyEnd := span.End - 1
	i := macroSkipSpace(source, span.Start+1, bodyEnd)
	nameStart := i
	if i < bodyEnd && source[i] == '\'' {
		i = macroQuotedEnd(source[:bodyEnd], i)
		if i <= nameStart+1 || i > bodyEnd || source[i-1] != '\'' {
			call.Unsupported = "unrecognized macro name"
			return call
		}
		call.Name = macroUnquoteIdentifier(source[nameStart+1 : i-1])
	} else if i < bodyEnd && macroNameStart(source, i) {
		i = macroNameEnd(source, i, bodyEnd)
		call.Name = source[nameStart:i]
	} else {
		call.Unsupported = "unrecognized macro name"
		return call
	}
	call.NameSpan = macroSpan{nameStart, i}
	i = macroSkipSpace(source, i, bodyEnd)
	if i < bodyEnd {
		if source[i] != '(' || source[bodyEnd-1] != ')' {
			call.Unsupported = "unrecognized macro argument syntax"
			return call
		}
		args, ok := macroArguments(source, i+1, bodyEnd-1, nested)
		if !ok {
			call.Unsupported = "unrecognized macro argument syntax"
			return call
		}
		call.Arguments = args
	}
	call.Nested = nested
	call.Position = macroPosition(source, call.Span)
	if call.Position == "" {
		call.Unsupported = "unrecognized macro insertion position"
	}
	return call
}

func macroUnquoteIdentifier(source string) string {
	var out strings.Builder
	for i := 0; i < len(source); i++ {
		if source[i] == '\\' && i+1 < len(source) {
			i++
		}
		out.WriteByte(source[i])
	}
	return out.String()
}

func macroSkipSpace(source string, start, end int) int {
	for start < end {
		r, width := utf8.DecodeRuneInString(source[start:end])
		if !unicode.IsSpace(r) {
			break
		}
		start += width
	}
	return start
}

func macroTrim(source string, span macroSpan) macroSpan {
	span.Start = macroSkipSpace(source, span.Start, span.End)
	for span.End > span.Start {
		r, width := utf8.DecodeLastRuneInString(source[span.Start:span.End])
		if !unicode.IsSpace(r) {
			break
		}
		span.End -= width
	}
	return span
}

func macroNameStart(source string, i int) bool {
	r, _ := utf8.DecodeRuneInString(source[i:])
	return unicode.IsLetter(r) || r == '_'
}

func macroNameEnd(source string, i, end int) int {
	for i < end {
		r, width := utf8.DecodeRuneInString(source[i:end])
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_' && r != '.' && r != ':' {
			break
		}
		i += width
	}
	return i
}

func macroNestedEnd(nested []macroInvocation, start int) (int, bool) {
	for _, child := range nested {
		if child.Span.Start == start {
			return child.Span.End, true
		}
	}
	return 0, false
}

func macroArguments(source string, start, end int, nested []macroInvocation) ([]macroArgument, bool) {
	if macroOnlyTrivia(source[start:end]) {
		return nil, true
	}
	var args []macroArgument
	var stack []byte
	part := start
	for i := start; i < end; i++ {
		if childEnd, ok := macroNestedEnd(nested, i); ok {
			i = childEnd - 1
			continue
		}
		switch source[i] {
		case '\'', '"':
			quotedEnd := macroQuotedEnd(source[:end], i)
			if quotedEnd == end && (end == i+1 || source[end-1] != source[i]) {
				return nil, false
			}
			i = quotedEnd - 1
		case '/':
			if commentEnd, ok := macroCommentEnd(source[:end], i); ok {
				i = commentEnd - 1
			}
		case '(', '[':
			stack = append(stack, source[i])
		case ')', ']':
			if len(stack) == 0 || (source[i] == ')' && stack[len(stack)-1] != '(') || (source[i] == ']' && stack[len(stack)-1] != '[') {
				return nil, false
			}
			stack = stack[:len(stack)-1]
		case ',':
			if len(stack) == 0 {
				arg, ok := macroParseArgument(source, macroSpan{part, i}, nested)
				if !ok {
					return nil, false
				}
				args = append(args, arg)
				part = i + 1
			}
		case '`':
			return nil, false
		}
	}
	if len(stack) != 0 {
		return nil, false
	}
	arg, ok := macroParseArgument(source, macroSpan{part, end}, nested)
	if !ok {
		return nil, false
	}
	return append(args, arg), true
}

func macroParseArgument(source string, span macroSpan, nested []macroInvocation) (macroArgument, bool) {
	span = macroTrim(source, span)
	if span.Start == span.End {
		return macroArgument{}, false
	}
	arg := macroArgument{Span: span, ValueSpan: span}
	var stack []byte
	equal, equalCount := -1, 0
	for i := span.Start; i < span.End; i++ {
		if childEnd, ok := macroNestedEnd(nested, i); ok {
			i = childEnd - 1
			continue
		}
		switch source[i] {
		case '\'', '"':
			i = macroQuotedEnd(source[:span.End], i) - 1
		case '/':
			if commentEnd, ok := macroCommentEnd(source[:span.End], i); ok {
				i = commentEnd - 1
			}
		case '(', '[':
			stack = append(stack, source[i])
		case ')', ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case '=':
			if len(stack) == 0 {
				equal, equalCount = i, equalCount+1
			}
		}
	}
	if equalCount == 1 && equal > span.Start && equal+1 < span.End &&
		source[equal-1] != '>' && source[equal-1] != '<' && source[equal-1] != '!' && source[equal+1] != '=' {
		nameStart := macroSkipTrivia(source, span.Start, equal)
		if nameStart < equal && macroNameStart(source, nameStart) {
			nameEnd := macroNameEnd(source, nameStart, equal)
			if macroOnlyTrivia(source[nameEnd:equal]) {
				arg.NameSpan = macroSpan{nameStart, nameEnd}
				arg.Name = source[nameStart:nameEnd]
				arg.ValueSpan = macroTrim(source, macroSpan{equal + 1, span.End})
			}
		}
	}
	if arg.ValueSpan.Start == arg.ValueSpan.End {
		return macroArgument{}, false
	}
	return arg, macroExpression(source[arg.ValueSpan.Start:arg.ValueSpan.End])
}

type macroSyntaxErrors struct {
	*antlr.DefaultErrorListener
	count int
}

func (e *macroSyntaxErrors) SyntaxError(_ antlr.Recognizer, _ interface{}, _, _ int, _ string, _ antlr.RecognitionException) {
	e.count++
}

// The analyzer grammar accepts expressions in macro argument lists. Parse each
// argument independently so malformed forms cannot be inferred as bindings.
func macroExpression(source string) bool {
	errors := &macroSyntaxErrors{DefaultErrorListener: antlr.NewDefaultErrorListener()}
	lexer := parser.NewSPLLexer(antlr.NewInputStream(source))
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errors)
	tokens := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewSPLParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(errors)
	p.AnalysisExpression()
	return errors.count == 0 && p.GetCurrentToken().GetTokenType() == antlr.TokenEOF
}

func macroPosition(source string, span macroSpan) string {
	// A token joined directly to an identifier has no safe replacement edge.
	for _, edge := range []int{span.Start - 1, span.End} {
		if edge < 0 || edge >= len(source) {
			continue
		}
		r, _ := utf8.DecodeRuneInString(source[edge:])
		if edge == span.Start-1 {
			r, _ = utf8.DecodeLastRuneInString(source[:span.Start])
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' || r == '.' || r == ':' {
			return ""
		}
	}
	region := macroStageRegion(source, span)
	left, right := region.Start, region.End
	depth := 0
pipes:
	for i := region.Start; i < region.End; i++ {
		if source[i] == '\'' || source[i] == '"' {
			i = macroQuotedEnd(source, i) - 1
			continue
		}
		if source[i] == '/' {
			if end, ok := macroCommentEnd(source, i); ok {
				i = end - 1
				continue
			}
		}
		switch source[i] {
		case '[':
			depth++
		case ']':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 {
				if i < span.Start {
					left = i + 1
				} else if i >= span.End {
					right = i
					break pipes
				}
			}
		}
	}
	if macroOnlyTrivia(source[left:span.Start]) && macroOnlyTrivia(source[span.End:right]) {
		return "stage"
	}
	return "fragment"
}

// macroStageRegion finds the nearest balanced subquery containing the call.
// Its brackets are stage boundaries; brackets in quoted text are ignored.
func macroStageRegion(source string, span macroSpan) macroSpan {
	region := macroSpan{0, len(source)}
	var opens []int
	for i := 0; i < len(source); i++ {
		if source[i] == '\'' || source[i] == '"' {
			i = macroQuotedEnd(source, i) - 1
			continue
		}
		if source[i] == '/' {
			if end, ok := macroCommentEnd(source, i); ok {
				i = end - 1
				continue
			}
		}
		switch source[i] {
		case '[':
			opens = append(opens, i)
		case ']':
			if len(opens) == 0 {
				continue
			}
			open := opens[len(opens)-1]
			opens = opens[:len(opens)-1]
			if open < span.Start && i >= span.End && open+1 > region.Start {
				region = macroSpan{open + 1, i}
			}
		}
	}
	return region
}

func macroSkipTrivia(source string, start, end int) int {
	for start < end {
		r, width := utf8.DecodeRuneInString(source[start:end])
		if unicode.IsSpace(r) {
			start += width
			continue
		}
		if commentEnd, ok := macroCommentEnd(source[:end], start); ok {
			start = commentEnd
			continue
		}
		break
	}
	return start
}

func macroOnlyTrivia(source string) bool {
	for i := 0; i < len(source); {
		r, width := utf8.DecodeRuneInString(source[i:])
		if unicode.IsSpace(r) {
			i += width
			continue
		}
		if end, ok := macroCommentEnd(source, i); ok {
			i = end
			continue
		}
		return false
	}
	return true
}
