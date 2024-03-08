package httphandlers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
	"golang.org/x/exp/slices"
)

func (c *HTTPHandler) GetMiners(ctx *gin.Context) {
	Miners := []Miner{}

	var (
		TotalHashrateGHS float64
		UsedHashrateGHS  float64

		TotalMiners       int
		BusyMiners        int
		PartialBusyMiners int
		FreeMiners        int
		VettingMiners     int
	)

	c.allocator.GetMiners().Range(func(m *allocator.Scheduler) bool {
		hrGHS, ok := m.GetHashrate().GetHashrateAvgGHSCustom(c.hashrateCounterDefault)
		if !ok {
			c.log.DPanicf("hashrate counter not found, %s", c.hashrateCounterDefault)
		} else {
			TotalHashrateGHS += hrGHS
		}

		hrGHS, ok = m.GetUsedHashrate().GetHashrateAvgGHSCustom(c.hashrateCounterDefault)
		if !ok {
			c.log.DPanicf("hashrate counter not found, %s", c.hashrateCounterDefault)
		} else {
			UsedHashrateGHS += hrGHS
		}

		TotalMiners += 1

		switch m.GetStatus(c.cycleDuration) {
		case allocator.MinerStatusFree:
			FreeMiners += 1
		case allocator.MinerStatusVetting:
			VettingMiners += 1
		case allocator.MinerStatusBusy:
			BusyMiners += 1
		case allocator.MinerStatusPartialBusy:
			PartialBusyMiners += 1
		}

		miner := c.MapMiner(m)
		Miners = append(Miners, *miner)

		return true
	})

	slices.SortStableFunc(Miners, func(a Miner, b Miner) bool {
		return a.ID < b.ID
	})

	res := &MinersResponse{
		TotalMiners:       TotalMiners,
		VettingMiners:     VettingMiners,
		FreeMiners:        FreeMiners,
		PartialBusyMiners: PartialBusyMiners,
		BusyMiners:        BusyMiners,

		TotalHashrateGHS:     int(TotalHashrateGHS),
		AvailableHashrateGHS: int(TotalHashrateGHS - UsedHashrateGHS),
		UsedHashrateGHS:      int(UsedHashrateGHS),

		Miners: Miners,
	}

	ctx.JSON(200, res)
}

func (c *HTTPHandler) MapMiner(m *allocator.Scheduler) *Miner {
	return &Miner{
		Resource: Resource{
			Self: c.publicUrl.JoinPath(fmt.Sprintf("/miners/%s", m.ID())).String(),
		},
		ID:                    m.ID(),                                  // readonly
		WorkerName:            m.GetWorkerName(),                       // readonly
		Status:                m.GetStatus(c.cycleDuration).String(),   // atomic
		CurrentDifficulty:     int(m.GetCurrentDifficulty()),           // atomic
		HashrateAvgGHS:        mapHRToInt(m),                           // atomic or single lock
		CurrentDestination:    m.GetCurrentDest().String(),             // atomic
		ConnectedAt:           m.GetConnectedAt().Format(time.RFC3339), // readonly
		Stats:                 m.GetStats(),                            // multiple atomics
		Uptime:                formatDuration(m.GetUptime()),           // readonly
		ActivePoolConnections: m.GetDestConns(),                        // sync map range + multiple atomics
		// Destinations:          m.GetDestinations(c.cycleDuration),      // readonly temporarily
	}
}
