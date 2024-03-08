package proxy

import "sync/atomic"

type DestStats struct {
	WeAcceptedTheyAccepted atomic.Uint64 // our validator accepted and dest accepted
	WeAcceptedTheyRejected atomic.Uint64 // our validator accepted and dest rejected
	WeRejectedTheyAccepted atomic.Uint64 // our validator rejected, but dest accepted
}

func (s *DestStats) IncWeAcceptedTheyAccepted() {
	s.WeAcceptedTheyAccepted.Add(1)
}

func (s *DestStats) IncWeAcceptedTheyRejected() {
	s.WeAcceptedTheyRejected.Add(1)
}

func (s *DestStats) IncWeRejectedTheyAccepted() {
	s.WeRejectedTheyAccepted.Add(1)
}

func (s *DestStats) GetStatsMap() map[string]int {
	return map[string]int{
		"we_accepted_they_accepted": int(s.WeAcceptedTheyAccepted.Load()),
		"we_accepted_they_rejected": int(s.WeAcceptedTheyRejected.Load()),
		"we_rejected_they_accepted": int(s.WeRejectedTheyAccepted.Load()),
	}
}

type SourceStats struct {
	WeAcceptedShares       atomic.Uint64 // shares that passed our validator (incl AcceptedUsRejectedThem)
	WeRejectedShares       atomic.Uint64 // shares that failed during validation (incl RejectedUsAcceptedThem)
	WeAcceptedTheyRejected atomic.Uint64 // shares that passed our validator, but rejected by the destination
	WeRejectedTheyAccepted atomic.Uint64 // shares that failed our validator, but accepted by the destination
}

func (s *SourceStats) IncWeAcceptedShares() {
	s.WeAcceptedShares.Add(1)
}

func (s *SourceStats) IncWeRejectedShares() {
	s.WeRejectedShares.Add(1)
}

func (s *SourceStats) IncWeAcceptedTheyRejected() {
	s.WeAcceptedTheyRejected.Add(1)
}

func (s *SourceStats) IncWeRejectedTheyAccepted() {
	s.WeRejectedTheyAccepted.Add(1)
}

func (s *SourceStats) GetStatsMap() map[string]int {
	return map[string]int{
		"we_accepted_shares":        int(s.WeAcceptedShares.Load()),
		"we_rejected_shares":        int(s.WeRejectedShares.Load()),
		"we_accepted_they_rejected": int(s.WeAcceptedTheyRejected.Load()),
		"we_rejected_they_accepted": int(s.WeRejectedTheyAccepted.Load()),
	}
}
