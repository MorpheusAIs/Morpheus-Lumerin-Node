package hashrate

import (
	"sync/atomic"
	"time"
)

type Mean struct {
	totalWork       *atomic.Uint64
	firstSubmitTime *atomic.Int64 // stores first submit time in unix seconds
	lastSubmitTime  *atomic.Int64 // stores last submit time in unix seconds
	totalShares     *atomic.Uint32
}

// NewMean creates a new Mean hashrate counter, which adds all submitted work and divides it by the total duration
// it is also used to track the first and last submit time and total work
func NewMean() *Mean {
	return &Mean{
		totalWork:       &atomic.Uint64{},
		firstSubmitTime: &atomic.Int64{},
		lastSubmitTime:  &atomic.Int64{},
		totalShares:     &atomic.Uint32{},
	}
}

func (h *Mean) Start() {
	h.maybeSetFirstSubmitTime(time.Now())
}

func (h *Mean) Reset() {
	h.totalWork.Store(0)
	h.firstSubmitTime.Store(0)
	h.lastSubmitTime.Store(0)
}

func (h *Mean) Add(diff float64) {
	h.totalWork.Add(uint64(diff))
	h.totalShares.Add(1)

	now := time.Now()
	h.maybeSetFirstSubmitTime(now)
	h.setLastSubmitTime(now)
}

func (h *Mean) Value() float64 {
	return float64(h.totalWork.Load())
}

func (h *Mean) ValuePer(t time.Duration) float64 {
	totalDuration := h.GetTotalDuration()
	if totalDuration == 0 {
		return 0
	}
	return float64(h.GetTotalWork()) / float64(totalDuration/t)
}

func (h *Mean) GetLastSubmitTime() time.Time {
	lastSubmitTime := h.lastSubmitTime.Load()
	if lastSubmitTime == 0 {
		return time.Time{}
	}
	return time.Unix(lastSubmitTime, 0)
}

func (h *Mean) GetTotalWork() uint64 {
	return h.totalWork.Load()
}

func (h *Mean) GetTotalShares() uint32 {
	return h.totalShares.Load()
}

func (h *Mean) GetTotalDuration() time.Duration {
	durationSeconds := time.Now().Unix() - h.firstSubmitTime.Load()
	return time.Duration(durationSeconds) * time.Second
}

func (h *Mean) maybeSetFirstSubmitTime(t time.Time) {
	h.firstSubmitTime.CompareAndSwap(0, t.Unix())
}

func (h *Mean) setLastSubmitTime(t time.Time) {
	h.lastSubmitTime.Store(t.Unix())
}
