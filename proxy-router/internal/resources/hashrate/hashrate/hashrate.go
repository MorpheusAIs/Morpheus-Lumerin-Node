package hashrate

import (
	"math"
	"time"
)

const MeanCounterKey = "mean"

type Hashrate struct {
	custom map[string]Counter
}

func NewHashrate(counters map[string]Counter) *Hashrate {
	counters[MeanCounterKey] = NewMean()

	return &Hashrate{
		custom: counters,
	}
}

func (h *Hashrate) Start() {
	for _, item := range h.custom {
		item.Start()
	}
}

func (h *Hashrate) Reset() {
	for _, item := range h.custom {
		item.Reset()
	}
}

func (h *Hashrate) OnSubmit(diff float64) {
	for _, item := range h.custom {
		item.Add(diff)
	}
}

// averageSubmitDiffToGHS converts average value provided by ema to hashrate in GH/S
func (h *Hashrate) averageSubmitDiffToGHS(averagePerSecond float64) float64 {
	return JobSubmittedToGHS(averagePerSecond)
}

func (h *Hashrate) GetHashrateAvgGHSCustom(ID string) (hrGHS float64, ok bool) {
	ema, ok := h.custom[ID]
	if !ok {
		return 0, false
	}
	return h.averageSubmitDiffToGHS(ema.ValuePer(time.Second)), true
}

func (h *Hashrate) GetHashrateAvgGHSAll() map[string]float64 {
	m := make(map[string]float64, len(h.custom))
	for key, item := range h.custom {
		m[key] = h.averageSubmitDiffToGHS(item.ValuePer(time.Second))
	}
	return m
}

func (h *Hashrate) GetTotalWork() float64 {
	return h.custom[MeanCounterKey].Value()
}

func (h *Hashrate) GetTotalDuration() time.Duration {
	return h.custom[MeanCounterKey].(*Mean).GetTotalDuration()
}

func (h *Hashrate) GetLastSubmitTime() time.Time {
	return h.custom[MeanCounterKey].(*Mean).GetLastSubmitTime()
}

func (h *Hashrate) GetTotalShares() int {
	return int(h.custom[MeanCounterKey].(*Mean).GetTotalShares())
}

func JobSubmittedToHS(jobSubmitted float64) float64 {
	return jobSubmitted * math.Pow(2, 32)
}

func HSToJobSubmitted(hrHS float64) float64 {
	return hrHS / math.Pow(2, 32)
}

func HSToGHS(hashrateHS float64) int {
	return int(hashrateHS / math.Pow10(9))
}

func GHSToHS(hrGHS int) float64 {
	return float64(hrGHS) * math.Pow10(9)
}

func GHSToJobSubmitted(hrGHS float64) float64 {
	return HSToJobSubmitted(hrGHS * math.Pow10(9))
}

func JobSubmittedToGHS(jobSubmitted float64) float64 {
	return JobSubmittedToHS(jobSubmitted) / math.Pow10(9)
}

func GHSToJobSubmittedV2(hrGHS float64, duration time.Duration) float64 {
	return HSToJobSubmitted(hrGHS*math.Pow10(9)) * duration.Seconds()
}

func JobSubmittedToGHSV2(jobSubmitted float64, duration time.Duration) float64 {
	return JobSubmittedToHS(jobSubmitted) / math.Pow10(9) / duration.Seconds()
}
