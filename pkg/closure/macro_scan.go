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
	Position    string // stage or fragment
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
	end := len(source)
	for i := start + 1; i < len(source); {
		if source[i] == '\'' || source[i] == '"' {
			i = macroQuotedEnd(source, i)
			continue
		}
		if source[i] == '`' {
			end = i + 1
			break
		}
		i++
	}
	call := macroInvocation{Span: macroSpan{start, end}}
	if end == len(source) && (end == start+1 || source[end-1] != '`') {
		call.Unsupported = "unclosed backtick invocation"
		return call
	}
	bodyEnd := end - 1
	i := macroSkipSpace(source, start+1, bodyEnd)
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
		args, ok := macroArguments(source, i+1, bodyEnd-1)
		if !ok {
			call.Unsupported = "unrecognized macro argument syntax"
			return call
		}
		call.Arguments = args
	}
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

func macroArguments(source string, start, end int) ([]macroArgument, bool) {
	if macroTrim(source, macroSpan{start, end}).Start == end {
		return nil, true
	}
	var args []macroArgument
	var stack []byte
	part := start
	for i := start; i < end; i++ {
		switch source[i] {
		case '\'', '"':
			quotedEnd := macroQuotedEnd(source[:end], i)
			if quotedEnd == end && (end == i+1 || source[end-1] != source[i]) {
				return nil, false
			}
			i = quotedEnd - 1
		case '(', '[':
			stack = append(stack, source[i])
		case ')', ']':
			if len(stack) == 0 || (source[i] == ')' && stack[len(stack)-1] != '(') || (source[i] == ']' && stack[len(stack)-1] != '[') {
				return nil, false
			}
			stack = stack[:len(stack)-1]
		case ',':
			if len(stack) == 0 {
				arg, ok := macroParseArgument(source, macroSpan{part, i})
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
	arg, ok := macroParseArgument(source, macroSpan{part, end})
	if !ok {
		return nil, false
	}
	return append(args, arg), true
}

func macroParseArgument(source string, span macroSpan) (macroArgument, bool) {
	span = macroTrim(source, span)
	if span.Start == span.End {
		return macroArgument{}, false
	}
	arg := macroArgument{Span: span, ValueSpan: span}
	var stack []byte
	equal := -1
	for i := span.Start; i < span.End; i++ {
		switch source[i] {
		case '\'', '"':
			i = macroQuotedEnd(source[:span.End], i) - 1
		case '(', '[':
			stack = append(stack, source[i])
		case ')', ']':
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case '=':
			if len(stack) == 0 {
				if equal >= 0 {
					return macroArgument{}, false
				}
				equal = i
			}
		}
	}
	if equal >= 0 {
		name := macroTrim(source, macroSpan{span.Start, equal})
		value := macroTrim(source, macroSpan{equal + 1, span.End})
		if name.Start == name.End || value.Start == value.End || !macroNameStart(source, name.Start) || macroNameEnd(source, name.Start, name.End) != name.End {
			return macroArgument{}, false
		}
		arg.NameSpan, arg.Name, arg.ValueSpan = name, source[name.Start:name.End], value
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
	left, right := 0, len(source)
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
		if source[i] == '|' {
			if i < span.Start {
				left = i + 1
			} else if i >= span.End {
				right = i
				break
			}
		}
	}
	if macroOnlyTrivia(source[left:span.Start]) && macroOnlyTrivia(source[span.End:right]) {
		return "stage"
	}
	return "fragment"
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
