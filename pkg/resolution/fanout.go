package resolution

import (
	"math/big"

	"github.com/delgado-jacob/spl-toolkit/pkg/analysis"
	"github.com/delgado-jacob/spl-toolkit/pkg/validation"
)

const defaultMaxVariants uint64 = 100

func combinationCount(resolutions []Resolution) *big.Int {
	count := big.NewInt(1)
	for _, r := range resolutions {
		count.Mul(count, new(big.Int).SetUint64(uint64(len(r.Values))))
	}
	return count
}
func admitFanout(resolutions []Resolution, maximum *uint64) (*big.Int, uint64, error) {
	limit := defaultMaxVariants
	if maximum != nil {
		limit = *maximum
	}
	if limit == 0 {
		return nil, limit, requestErrorAt("request_invalid", "/max_variants", "max_variants must be positive")
	}
	count := combinationCount(resolutions)
	if count.Cmp(new(big.Int).SetUint64(limit)) > 0 {
		return count, limit, &validation.InputError{Err: &requestError{detail: RequestErrorDetail{Code: "fanout_limit_exceeded", Path: "/max_variants", Message: "combination count exceeds max_variants", TotalCombinations: count.String(), MaxVariants: &limit}}}
	}
	return count, limit, nil
}

// visitSelections enumerates a previously admitted product without parser or evidence work.
func visitSelections(resolutions []Resolution, visit func(uint64, []analysis.ResolutionChoice) error) error {
	count := combinationCount(resolutions)
	if !count.IsUint64() {
		_, _, err := admitFanout(resolutions, uint64Pointer(^uint64(0)))
		return err
	}
	indices := make([]int, len(resolutions))
	total := count.Uint64()
	for ordinal := uint64(1); ordinal <= total; ordinal++ {
		selection := make([]analysis.ResolutionChoice, len(resolutions))
		for i, r := range resolutions {
			selection[i] = analysis.ResolutionChoice{Placeholder: r.Placeholder, Kind: r.Kind, Value: r.Values[indices[i]]}
		}
		if err := visit(ordinal, selection); err != nil {
			return err
		}
		// Return before incrementing the final ordinal, even at uint64 maximum.
		if ordinal == total {
			return nil
		}
		for i := len(indices) - 1; i >= 0; i-- {
			indices[i]++
			if indices[i] < len(resolutions[i].Values) {
				break
			}
			indices[i] = 0
		}
	}
	return nil
}
func uint64Pointer(v uint64) *uint64 { return &v }
