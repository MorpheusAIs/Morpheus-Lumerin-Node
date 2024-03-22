package contract

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"net/url"
// 	"sync/atomic"
// 	"time"

// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
// 	hashrateContract "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
// 	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
// 	hr "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
// 	"golang.org/x/exp/slices"
// )

// type ContractWatcherSeller struct {
// 	// config
// 	contractCycleDuration time.Duration

// 	// state
// 	state                resources.ContractState
// 	fullMiners           []string
// 	actualHRGHS          *hr.Hashrate
// 	fulfillmentStartedAt time.Time
// 	deliveryLogs         *DeliveryLog
// 	tsk                  *lib.Task

// 	// deps
// 	*hashrate.Terms
// 	allocator *allocator.Allocator
// 	log       interfaces.ILogger
// }

// func NewContractWatcherSeller(terms *hashrateContract.Terms, cycleDuration time.Duration, hashrateFactory func() *hr.Hashrate, allocator *allocator.Allocator, log interfaces.ILogger) *ContractWatcherSeller {
// 	p := &ContractWatcherSeller{
// 		Terms:                 terms,
// 		state:                 resources.ContractStatePending,
// 		allocator:             allocator,
// 		fullMiners:            []string{},
// 		contractCycleDuration: cycleDuration,
// 		actualHRGHS:           hashrateFactory(),
// 		deliveryLogs:          NewDeliveryLog(),
// 		log:                   log,
// 	}
// 	p.tsk = lib.NewTaskFunc(func(ctx context.Context) error {
// 		p.state = resources.ContractStateRunning
// 		err := p.Run(ctx)
// 		p.state = resources.ContractStatePending
// 		return err
// 	})
// 	return p
// }

// func (p *ContractWatcherSeller) StartFulfilling(ctx context.Context) {
// 	if p.state == resources.ContractStateRunning {
// 		p.log.Warnf("contract already started fulfilling")
// 		return
// 	}
// 	p.log.Infof("contract started fulfilling")
// 	p.fulfillmentStartedAt = time.Now()
// 	p.tsk.Start(ctx)
// }

// func (p *ContractWatcherSeller) StopFulfilling() {
// 	<-p.tsk.Stop()
// 	p.allocator.CancelTasks(p.ID())
// 	p.log.Infof("contract stopped fulfilling")
// }

// func (p *ContractWatcherSeller) Done() <-chan struct{} {
// 	return p.tsk.Done()
// }

// func (p *ContractWatcherSeller) Reset() {
// 	p.tsk = lib.NewTaskFunc(p.Run)
// }

// func (p *ContractWatcherSeller) Err() error {
// 	if errors.Is(p.tsk.Err(), context.Canceled) {
// 		return ErrContractClosed
// 	}
// 	return p.tsk.Err()
// }

// func (p *ContractWatcherSeller) SetData(terms *hashrateContract.Terms) {
// 	p.Terms = terms
// }

// // Run is the main loop of the contract. It is responsible for allocating miners for the contract.
// // Returns nil if the contract ended successfully, ErrClosed if the contract was closed before it ended.
// func (p *ContractWatcherSeller) Run(ctx context.Context) error {
// 	p.actualHRGHS.Reset()
// 	p.actualHRGHS.Start()
// 	partialDeliveryTargetGHS := p.HashrateGHS()
// 	thisCycleJobSubmitted := atomic.Uint64{}
// 	globalUnderdeliveryGHS := 0.0 // global contract underdelivery
// 	jobSubmittedFullMiners := atomic.Uint64{}
// 	jobSubmittedPartialMiners := atomic.Uint64{}
// 	sharesSubmitted := atomic.Uint64{}
// 	partialMinersNum := 0
// 	p.fullMiners = p.fullMiners[:0]

// 	for {
// 		if p.BlockchainState() != hashrate.BlockchainStateRunning {
// 			// this is an extra check, should not happen
// 			err := fmt.Errorf("contract %s blockchain state is not running", p.ID())
// 			return err
// 		}
// 		if p.EndTime().Before(time.Now()) {
// 			p.log.Debug("contract ended!!!!!!")
// 			return nil
// 		}

// 		partialMinersNum = 0
// 		jobSubmittedFullMiners.Store(0)
// 		jobSubmittedPartialMiners.Store(0)

// 		if partialDeliveryTargetGHS > 0 {
// 			fullMiners, newRemainderGHS := p.allocator.AllocateFullMinersForHR(
// 				p.ID(),
// 				partialDeliveryTargetGHS,
// 				p.getAdjustedDest(),
// 				p.Duration(),
// 				func(diff float64, ID string) {
// 					jobSubmittedFullMiners.Add(uint64(diff))
// 					p.actualHRGHS.OnSubmit(diff)
// 					thisCycleJobSubmitted.Add(uint64(diff))
// 					sharesSubmitted.Add(1)
// 				},
// 			)
// 			if len(fullMiners) > 0 {
// 				partialDeliveryTargetGHS = newRemainderGHS
// 				p.log.Infof("fully allocated %d miners, new partialDeliveryTargetGHS = %.0f", len(fullMiners), partialDeliveryTargetGHS)
// 				p.fullMiners = append(p.fullMiners, fullMiners...)
// 			} else {
// 				p.log.Debugf("no full miners were allocated for this contract")
// 			}

// 			minerID, ok := p.allocator.AllocatePartialForHR(
// 				p.ID(),
// 				partialDeliveryTargetGHS,
// 				p.getAdjustedDest(),
// 				p.contractCycleDuration,
// 				func(diff float64, ID string) {
// 					jobSubmittedPartialMiners.Add(uint64(diff))
// 					p.actualHRGHS.OnSubmit(diff)
// 					thisCycleJobSubmitted.Add(uint64(diff))
// 					sharesSubmitted.Add(1)
// 				},
// 			)

// 			if ok {
// 				partialMinersNum = 1
// 				p.log.Debugf("remainderGHS: %.0f, was allocated by partial miners %v", partialDeliveryTargetGHS, minerID)
// 			} else {
// 				partialMinersNum = 0
// 				p.log.Warnf("remainderGHS: %.0f, was not allocated by partial miners", partialDeliveryTargetGHS)
// 			}
// 		}

// 		// in case of too much hashrate
// 		if partialDeliveryTargetGHS < 0 {
// 			p.log.Info("removing least powerful miner from contract")
// 			items := p.getAllocatedMinersSorted()

// 			if len(items) > 0 {
// 				minerToRemove := items[0].ID
// 				miner, ok := p.allocator.GetMiners().Load(minerToRemove)
// 				if ok {
// 					miner.RemoveTasksByID(p.ID())
// 					p.log.Debugf("miner %s tasks removed", miner.ID())
// 					fullMiners := p.getFullMiners()
// 					newFullMiners := make([]string, len(fullMiners)-1)
// 					i := 0
// 					for _, minerID := range fullMiners {
// 						if minerID == minerToRemove {
// 							continue
// 						}
// 						newFullMiners[i] = minerID
// 						i++
// 					}
// 					p.fullMiners = newFullMiners

// 					// sets new target and restarts the cycle
// 					partialDeliveryTargetGHS = miner.HashrateGHS() + partialDeliveryTargetGHS
// 					continue
// 				}
// 			} else {
// 				p.log.Warnf("no miners found to be removed")
// 			}
// 		}

// 		select {
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		case <-time.After(p.getEndsAfter()):
// 			expectedJob := hr.GHSToJobSubmitted(p.HashrateGHS()) * p.Duration().Seconds()
// 			actualJob := p.actualHRGHS.GetTotalWork()
// 			undeliveredJob := expectedJob - actualJob
// 			undeliveredFraction := undeliveredJob / expectedJob

// 			for _, minerID := range p.getFullMiners() {
// 				miner, ok := p.allocator.GetMiners().Load(minerID)
// 				if !ok {
// 					continue
// 				}
// 				miner.RemoveTasksByID(p.ID())
// 				p.log.Debugf("miner %s tasks removed", miner.ID())
// 			}
// 			p.fullMiners = p.fullMiners[:0]

// 			// partial miners tasks are not reset because they are not allocated
// 			// for the full duration of the contract

// 			p.log.Infof("contract ended, undelivered work %d, undelivered fraction %.2f",
// 				int(undeliveredJob), undeliveredFraction)
// 			return nil
// 		case <-time.After(p.contractCycleDuration):
// 		}

// 		thisCycleActualGHS := p.jobToGHS(thisCycleJobSubmitted.Load())
// 		thisCycleUnderDeliveryGHS := p.HashrateGHS() - thisCycleActualGHS
// 		globalUnderdeliveryGHS += thisCycleUnderDeliveryGHS

// 		// plan for the next cycle is to compensate for the under delivery of the contract
// 		// partialDeliveryTargetGHS = partialDeliveryTargetGHS + globalUnderdeliveryGHS
// 		partialDeliveryTargetGHS = p.HashrateGHS() - p.getFullMinersHR() + globalUnderdeliveryGHS

// 		thisCycleJobSubmitted.Store(0)

// 		logEntry := DeliveryLogEntry{
// 			Timestamp:                         time.Now(),
// 			ActualGHS:                         int(thisCycleActualGHS),
// 			FullMinersGHS:                     int(p.jobToGHS(jobSubmittedFullMiners.Load())),
// 			FullMiners:                        len(p.getFullMiners()),
// 			PartialMinersGHS:                  int(p.jobToGHS(jobSubmittedPartialMiners.Load())),
// 			PartialMinersNumber:               partialMinersNum,
// 			SharesSubmitted:                   int(sharesSubmitted.Load()),
// 			UnderDeliveryGHS:                  int(thisCycleUnderDeliveryGHS),
// 			GlobalHashrateGHS:                 int(p.actualHRGHS.GetHashrateAvgGHSAll()["mean"]),
// 			GlobalUnderDeliveryGHS:            int(globalUnderdeliveryGHS),
// 			GlobalError:                       1 - p.actualHRGHS.GetHashrateAvgGHSAll()["mean"]/p.HashrateGHS(),
// 			NextCyclePartialDeliveryTargetGHS: int(partialDeliveryTargetGHS),
// 		}
// 		p.deliveryLogs.AddEntry(logEntry)

// 		p.log.Info("contract cycle ended", logEntry)
// 	}
// }

// func (p *ContractWatcherSeller) getFullMiners() []string {
// 	newFullMiners := make([]string, 0, len(p.fullMiners))
// 	for _, minerID := range p.fullMiners {
// 		_, ok := p.allocator.GetMiners().Load(minerID)
// 		if !ok {
// 			continue
// 		}
// 		newFullMiners = append(newFullMiners, minerID)
// 	}
// 	if len(newFullMiners) != len(p.fullMiners) {
// 		p.fullMiners = newFullMiners
// 	}
// 	return p.fullMiners
// }

// func (p *ContractWatcherSeller) getEndsAfter() time.Duration {
// 	endTime := p.EndTime()
// 	if endTime.IsZero() {
// 		return 0
// 	}
// 	return endTime.Sub(time.Now())
// }

// func (p *ContractWatcherSeller) jobToGHS(value uint64) float64 {
// 	return hr.JobSubmittedToGHS(float64(value) / p.contractCycleDuration.Seconds())
// }

// func (p *ContractWatcherSeller) getAllocatedMinersSorted() []*allocator.MinerItem {
// 	var items []*allocator.MinerItem

// 	for _, minerID := range p.fullMiners {
// 		miner, ok := p.allocator.GetMiners().Load(minerID)
// 		if !ok {
// 			continue
// 		}
// 		items = append(items, &allocator.MinerItem{
// 			ID:    miner.ID(),
// 			HrGHS: miner.HashrateGHS(),
// 		})
// 	}

// 	slices.SortStableFunc(items, func(a, b *allocator.MinerItem) bool {
// 		return b.HrGHS > a.HrGHS
// 	})

// 	return items
// }

// func (p *ContractWatcherSeller) getFullMinersHR() float64 {
// 	var total float64
// 	for _, minerID := range p.fullMiners {
// 		miner, ok := p.allocator.GetMiners().Load(minerID)
// 		if !ok {
// 			continue
// 		}
// 		total += miner.HashrateGHS()
// 	}
// 	return total
// }

// // getAdjustedDest returns the destination url with the username set to the contractID
// // this is required for the buyer to distinguish incoming hashrate between different contracts
// func (p *ContractWatcherSeller) getAdjustedDest() *url.URL {
// 	if p.Terms.Dest() == nil {
// 		return nil
// 	}
// 	dest := lib.CopyURL(p.Terms.Dest())
// 	lib.SetUserName(dest, p.Terms.ID())
// 	return dest
// }

// // ShouldBeRunning checks blockchain state and expiration time and returns true if the contract should be running
// func (p *ContractWatcherSeller) ShouldBeRunning() bool {
// 	return p.Terms.BlockchainState() == hashrate.BlockchainStateRunning
// }

// //
// // Public getters
// //

// func (p *ContractWatcherSeller) Role() resources.ContractRole {
// 	return resources.ContractRoleSeller
// }

// func (p *ContractWatcherSeller) Dest() string {
// 	if dest := p.getAdjustedDest(); dest != nil {
// 		return dest.String()
// 	}
// 	return ""
// }

// func (p *ContractWatcherSeller) FulfillmentStartTime() time.Time {
// 	return p.fulfillmentStartedAt
// }

// func (p *ContractWatcherSeller) ResourceType() string {
// 	return ResourceTypeHashrate
// }

// func (p *ContractWatcherSeller) ResourceEstimates() map[string]float64 {
// 	return map[string]float64{
// 		ResourceEstimateHashrateGHS: p.Terms.HashrateGHS(),
// 	}
// }

// func (p *ContractWatcherSeller) ResourceEstimatesActual() map[string]float64 {
// 	return p.actualHRGHS.GetHashrateAvgGHSAll()
// }

// func (p *ContractWatcherSeller) ValidationStage() hashrateContract.ValidationStage {
// 	return hashrateContract.ValidationStageNotApplicable // only for buyer
// }

// func (p *ContractWatcherSeller) GetDeliveryLogs() ([]DeliveryLogEntry, error) {
// 	return p.deliveryLogs.GetEntries()
// }

// func (p *ContractWatcherSeller) State() resources.ContractState {
// 	return p.state
// }
