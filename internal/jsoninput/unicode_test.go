package jsoninput

import "testing"

func TestValidateUnicodeRejectsLossyJSONInput(t *testing.T) {
	for _, test := range []struct {
		name string
		body []byte
	}{
		{name: "invalid raw UTF-8", body: append([]byte(`{"text":"`), 0xff)},
		{name: "lone high surrogate", body: []byte(`{"text":"\ud800"}`)},
		{name: "high surrogate without low", body: []byte(`{"text":"\ud800x"}`)},
		{name: "high followed by ordinary escape", body: []byte(`{"text":"\ud800\u0061"}`)},
		{name: "lone low surrogate", body: []byte(`{"source_id":"\udc00"}`)},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateUnicode(test.body); err == nil {
				t.Fatal("expected source-preservation error")
			}
		})
	}
}

func TestValidateUnicodeAcceptsLosslessJSONInput(t *testing.T) {
	for _, body := range [][]byte{
		[]byte(`{"text":"search a=1"}`),
		[]byte(`{"text":"\ud83d\ude00"}`),
		[]byte(`{"text":"\\ud800"}`),
		[]byte(`{"text":"\u0061"}`),
		[]byte(`{"text":`),
	} {
		if err := ValidateUnicode(body); err != nil {
			t.Errorf("ValidateUnicode(%q): %v", body, err)
		}
	}
}
