package resolution

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
)

func TestResolutionFanoutOrder(t *testing.T) {
	resolutions := []Resolution{{Placeholder: "$a", Kind: "dataset", Values: []string{"a", "b"}}, {Placeholder: "$b", Kind: "dataset", Values: []string{"x", "y", "z"}}}
	var got []string
	var retained [][]analysis.ResolutionChoice
	err := visitSelections(resolutions, func(ordinal uint64, selection []analysis.ResolutionChoice) error {
		if ordinal != uint64(len(got)+1) {
			t.Fatal("ordinal")
		}
		got = append(got, selection[0].Value+"/"+selection[1].Value)
		retained = append(retained, selection)
		return nil
	})
	if err != nil || !reflect.DeepEqual(got, []string{"a/x", "a/y", "a/z", "b/x", "b/y", "b/z"}) {
		t.Fatalf("%v %v", got, err)
	}
	retained[0][0].Value = "changed"
	if retained[1][0].Value != "a" {
		t.Fatal("selections alias")
	}
	calls := 0
	if err := visitSelections(nil, func(n uint64, s []analysis.ResolutionChoice) error {
		calls++
		if n != 1 || s == nil || len(s) != 0 {
			t.Fatal("empty product")
		}
		return nil
	}); err != nil || calls != 1 {
		t.Fatalf("%d %v", calls, err)
	}
	sentinel := errors.New("stop")
	calls = 0
	if err := visitSelections(resolutions, func(uint64, []analysis.ResolutionChoice) error { calls++; return sentinel }); err != sentinel || calls != 1 {
		t.Fatal("callback error lost")
	}
}
func TestResolutionFanoutAdmission(t *testing.T) {
	values := make([]string, 100)
	for i := range values {
		values[i] = fmt.Sprint(i)
	}
	r := []Resolution{{Placeholder: "$a", Kind: "dataset", Values: values}}
	count, limit, err := admitFanout(r, nil)
	if err != nil || limit != 100 || count.String() != "100" {
		t.Fatal(count, limit, err)
	}
	r[0].Values = append(values, "100")
	if _, _, err := admitFanout(r, nil); err == nil {
		t.Fatal("default boundary")
	}
	if _, _, err := admitFanout(r, uint64Pointer(101)); err != nil {
		t.Fatal(err)
	}
	huge := make([]Resolution, 65)
	for i := range huge {
		huge[i] = Resolution{Placeholder: fmt.Sprintf("$p%d", i), Kind: "dataset", Values: []string{"a", "b"}}
	}
	if combinationCount(huge).String() != "36893488147419103232" {
		t.Fatal(combinationCount(huge))
	}
	_, _, err = admitFanout(huge, nil)
	detail, ok := RequestErrorDetails(err)
	if !ok || detail.Code != "fanout_limit_exceeded" || detail.TotalCombinations != "36893488147419103232" || *detail.MaxVariants != 100 {
		t.Fatalf("%+v %v", detail, err)
	}
	calls := 0
	err = visitSelections(huge, func(uint64, []analysis.ResolutionChoice) error { calls++; return nil })
	if err == nil || calls != 0 {
		t.Fatal("overflow enumerated")
	}
	request := requestFixture(t)
	request.Resolutions = huge
	if _, err := normalizeRequest(request); err == nil {
		t.Fatal("typed overflow admitted")
	}
}
