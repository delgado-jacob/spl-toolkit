// Package jsoninput validates JSON source encodings before decoding can replace data.
package jsoninput

import (
	"fmt"
	"unicode/utf8"
)

// ValidateUnicode rejects byte sequences and UTF-16 surrogate escapes that a
// JSON decoder would replace instead of preserving. Ordinary JSON syntax is
// intentionally left to the caller's decoder.
func ValidateUnicode(data []byte) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("request body is not valid UTF-8")
	}

	inString := false
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '"':
			inString = !inString
		case '\\':
			if !inString || i+1 >= len(data) {
				continue
			}
			if data[i+1] != 'u' {
				i++
				continue
			}
			value, ok := unicodeEscapeValue(data, i)
			if !ok {
				continue
			}
			switch {
			case value >= 0xd800 && value <= 0xdbff:
				low, paired := unicodeEscapeValue(data, i+6)
				if !paired || low < 0xdc00 || low > 0xdfff {
					return fmt.Errorf("request body contains an unpaired UTF-16 surrogate escape at byte %d", i)
				}
				i += 11
			case value >= 0xdc00 && value <= 0xdfff:
				return fmt.Errorf("request body contains an unpaired UTF-16 surrogate escape at byte %d", i)
			default:
				i += 5
			}
		}
	}
	return nil
}

func unicodeEscapeValue(data []byte, start int) (uint16, bool) {
	if start < 0 || start+6 > len(data) || data[start] != '\\' || data[start+1] != 'u' {
		return 0, false
	}
	var value uint16
	for _, b := range data[start+2 : start+6] {
		value <<= 4
		switch {
		case b >= '0' && b <= '9':
			value |= uint16(b - '0')
		case b >= 'a' && b <= 'f':
			value |= uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			value |= uint16(b-'A') + 10
		default:
			return 0, false
		}
	}
	return value, true
}
