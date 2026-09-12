package jsoninput

import "testing"

func TestCanonicalNumberExactInputSized(t *testing.T) {
	for _, tc := range []struct{ input, want string }{
		{"9007199254740993.00", "9007199254740993e0"},
		{"10e1000000", "1e1000001"},
		{"10e-1000002", "1e-1000001"},
		{"10e9223372036854775808", "1e9223372036854775809"},
		{"10e-9223372036854775810", "1e-9223372036854775809"},
		{"-1.00e+9999999999999999999999999999", "-1e9999999999999999999999999999"},
		{"-10e-9999999999999999999999999999", "-1e-9999999999999999999999999998"},
		{"-0e9999999999999999999999999999", "0"},
		{"0e-9999999999999999999999999999", "0"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			if got := CanonicalNumber(tc.input); got != tc.want {
				t.Fatalf("got %s; want %s", got, tc.want)
			}
		})
	}
}
