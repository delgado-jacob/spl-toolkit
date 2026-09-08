package validation

import (
	"regexp"
	"strconv"
	"strings"
)

// The supported subset is ASCII literals, edge anchors, positive ASCII classes,
// escaped punctuation and single-atom quantifiers. Everything else is unknown.
// This gate runs before Go regexp, whose language is not ECMAScript.
type schemaPattern struct{ re *regexp.Regexp }

func compileSchemaPattern(s string) (schemaPattern, error) {
	unsupported := func() (schemaPattern, error) { return schemaPattern{}, nil }
	atom := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 128 || c == '\n' || c == '\r' {
			return unsupported()
		}
		switch c {
		case '^':
			if i != 0 {
				return unsupported()
			}
			atom = false
		case '$':
			if i != len(s)-1 {
				return unsupported()
			}
			atom = false
		case '.', '(', ')', '|':
			return unsupported()
		case '\\':
			i++
			if i == len(s) {
				return schemaPattern{}, inputError("unterminated pattern escape")
			}
			if !strings.ContainsRune(`\^$.*+?()[]{}|/-`, rune(s[i])) {
				return unsupported()
			}
			atom = true
		case '[':
			i++
			if i == len(s) {
				return schemaPattern{}, inputError("unterminated pattern class")
			}
			if s[i] == '^' {
				return unsupported()
			}
			start := i
			for i < len(s) && s[i] != ']' {
				if s[i] >= 128 || s[i] == '\n' || s[i] == '\r' || s[i] == '[' {
					return unsupported()
				}
				if s[i] == '\\' {
					i++
					if i >= len(s) {
						return schemaPattern{}, inputError("unterminated class escape")
					}
					if !strings.ContainsRune(`\^$.*+?()[]{}|/-`, rune(s[i])) {
						return unsupported()
					}
				}
				i++
			}
			if i == start {
				return unsupported()
			}
			if i == len(s) {
				return schemaPattern{}, inputError("invalid pattern class")
			}
			atom = true
		case '?', '*', '+':
			if !atom {
				if c == '?' && i > 0 && strings.ContainsRune("?*+}", rune(s[i-1])) {
					return unsupported()
				}
				return schemaPattern{}, inputError("pattern quantifier lacks atom")
			}
			atom = false
		case '{':
			if !atom {
				return unsupported()
			}
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				return unsupported()
			}
			end += i
			parts := strings.Split(s[i+1:end], ",")
			if len(parts) > 2 {
				return unsupported()
			}
			counts := []int{}
			for j, p := range parts {
				if p == "" && j == 1 {
					counts = append(counts, -1)
					continue
				}
				if p == "" {
					return unsupported()
				}
				for _, d := range p {
					if d < '0' || d > '9' {
						return unsupported()
					}
				}
				if len(p) > 1 && p[0] == '0' {
					return unsupported()
				}
				n, e := strconv.ParseUint(p, 10, 64)
				if e != nil || n > 1000 {
					return unsupported()
				}
				counts = append(counts, int(n))
			}
			if len(counts) == 2 && counts[1] >= 0 && counts[0] > counts[1] {
				return schemaPattern{}, inputError("pattern repetition minimum exceeds maximum")
			}
			i = end
			atom = false
		case ']', '}':
			return unsupported()
		default:
			atom = true
		}
	}
	re, e := regexp.Compile(s)
	if e != nil {
		return schemaPattern{}, inputError("invalid supported pattern: %v", e)
	}
	return schemaPattern{re: re}, nil
}
func (p schemaPattern) match(s string) (bool, bool) {
	if p.re == nil {
		return false, false
	}
	for _, r := range s {
		if r >= 128 || r == '\r' || r == '\n' {
			return false, false
		}
	}
	return p.re.MatchString(s), true
}
