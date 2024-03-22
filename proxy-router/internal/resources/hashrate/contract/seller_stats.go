package contract

import (
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	hr "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	"go.uber.org/atomic"
)

type stats struct {
	jobFullMiners          *atomic.Uint64
	jobPartialMiners       *atomic.Uint64
	sharesFullMiners       *atomic.Uint64
	sharesPartialMiners    *atomic.Uint64
	globalUnderDeliveryGHS *atomic.Int64
	fullMiners             *lib.Set
	partialMiners          []string
	deliveryTargetGHS      float64
	actualHRGHS            *hr.Hashrate
}

func (s *stats) onFullMinerShare(diff float64, ID string) {
	s.jobFullMiners.Add(uint64(diff))
	s.actualHRGHS.OnSubmit(diff)
	s.sharesFullMiners.Add(1)
}

func (s *stats) onPartialMinerShare(diff float64, ID string) {
	s.jobPartialMiners.Add(uint64(diff))
	s.actualHRGHS.OnSubmit(diff)
	s.sharesPartialMiners.Add(1)
}

func (s *stats) addFullMiners(IDs ...string) {
	s.fullMiners.Add(IDs...)
}

func (s *stats) removeFullMiner(ID string) (ok bool) {
	return s.fullMiners.Remove(ID)
}

func (s *stats) addPartialMiners(IDs ...string) {
	s.partialMiners = append(s.partialMiners, IDs...)
}

func (s *stats) removePartialMiner(ID string) (ok bool) {
	oldLen := len(s.partialMiners)
	s.partialMiners = lib.FilterValue(s.partialMiners, ID)
	return oldLen != len(s.partialMiners)
}

func (c *stats) totalJob() float64 {
	return float64(c.jobFullMiners.Load()) + float64(c.jobPartialMiners.Load())
}
