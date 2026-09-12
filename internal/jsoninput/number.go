package jsoninput

import (
	"math/big"
	"strings"
)

// CanonicalNumber normalizes an already validated JSON number as coefficient *
// 10^exponent. It never expands powers of ten: even exponents beyond int64 use
// storage proportional to the input. Callers retain responsibility for strict
// JSON validation, so an invalid interpretation cannot become ordinary inequality.
func CanonicalNumber(number string) string {
	sign := ""
	if strings.HasPrefix(number, "-") {
		sign, number = "-", number[1:]
	}
	exponent := new(big.Int)
	if i := strings.IndexAny(number, "eE"); i >= 0 {
		exponent.SetString(number[i+1:], 10)
		number = number[:i]
	}
	if i := strings.IndexByte(number, '.'); i >= 0 {
		exponent.Sub(exponent, big.NewInt(int64(len(number)-i-1)))
		number = number[:i] + number[i+1:]
	}
	number = strings.TrimLeft(number, "0")
	if number == "" {
		return "0"
	}
	coefficient := strings.TrimRight(number, "0")
	exponent.Add(exponent, big.NewInt(int64(len(number)-len(coefficient))))
	return sign + coefficient + "e" + exponent.String()
}
