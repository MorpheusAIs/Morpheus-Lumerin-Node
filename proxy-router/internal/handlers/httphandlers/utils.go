package httphandlers

import (
	"math/big"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
)

// TimePtrToStringPtr converts nullable time to nullable string
func TimePtrToStringPtr(t *time.Time) *string {
	if t != nil {
		a := formatTime(*t)
		return &a
	}
	return nil
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func DurationPtrToStringPtr(t *time.Duration) *string {
	if t != nil {
		a := formatDuration(*t)
		return &a
	}
	return nil
}

func mapHRToInt(m *allocator.Scheduler) map[string]int {
	hrFloat := m.GetHashrate().GetHashrateAvgGHSAll()
	hrInt := make(map[string]int, len(hrFloat))
	for k, v := range hrFloat {
		hrInt[k] = int(v)
	}
	return hrInt
}

func formatDuration(dur time.Duration) string {
	return dur.Round(time.Second).String()
}

func roundResourceEstimates(estimates map[string]float64) map[string]int {
	res := make(map[string]int, len(estimates))
	for k, v := range estimates {
		res[k] = int(v)
	}
	return res
}

// LMRWithDecimalsToLMR converts LMR with decimals to LMR without decimals
func LMRWithDecimalsToLMR(LMRWithDecimals *big.Int) float64 {
	v, _ := lib.NewRat(LMRWithDecimals, big.NewInt(1e8)).Float64()
	return v
}
