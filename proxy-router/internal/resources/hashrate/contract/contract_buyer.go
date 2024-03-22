package contract

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources"
	hashrateContract "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/allocator"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	"go.uber.org/atomic"
)

type ContractWatcherBuyer struct {
	// config
	contractCycleDuration    time.Duration
	shareTimeout             time.Duration // time to wait for the share to arrive, otherwise close contract
	hrErrorThreshold         float64       // hashrate relative error threshold for the contract to be considered fulfilling accurately
	hashrateCounterNameBuyer string
	hrValidationFlatness     time.Duration
	role                     resources.ContractRole

	// state
	state                *lib.AtomicValue[resources.ContractState]
	validationStage      *lib.AtomicValue[hashrateContract.ValidationStage]
	fulfillmentStartedAt *atomic.Time
	starvingGHS          *atomic.Uint64
	contractErr          atomic.Error // keeps the last error that happened in the contract that prevents it from fulfilling correctly, like invalid destination

	tsk    *lib.Task
	cancel context.CancelFunc
	err    error
	doneCh chan struct{}

	//deps
	*hashrateContract.Terms
	allocator      *allocator.Allocator
	globalHashrate *hashrate.GlobalHashrate
	log            interfaces.ILogger
}

func NewContractWatcherBuyer(
	terms *hashrateContract.Terms,
	hashrateFactory func() *hashrate.Hashrate,
	allocator *allocator.Allocator,
	globalHashrate *hashrate.GlobalHashrate,
	log interfaces.ILogger,

	cycleDuration time.Duration,
	shareTimeout time.Duration,
	hrErrorThreshold float64,
	hashrateCounterNameBuyer string,
	hrValidationFlatness time.Duration,
	role resources.ContractRole,
) *ContractWatcherBuyer {
	return &ContractWatcherBuyer{
		contractCycleDuration:    cycleDuration,
		shareTimeout:             shareTimeout,
		hrErrorThreshold:         hrErrorThreshold,
		hrValidationFlatness:     hrValidationFlatness,
		hashrateCounterNameBuyer: hashrateCounterNameBuyer,
		role:                     role,

		state:                lib.NewAtomicValue(resources.ContractStatePending),
		validationStage:      lib.NewAtomicValue(hashrateContract.ValidationStageValidating),
		fulfillmentStartedAt: atomic.NewTime(time.Time{}),
		starvingGHS:          atomic.NewUint64(0),

		Terms:          terms,
		allocator:      allocator,
		globalHashrate: globalHashrate,
		log:            log,
	}
}

func (p *ContractWatcherBuyer) StartFulfilling(ctx context.Context) {
	if p.state.Load() == resources.ContractStateRunning {
		p.log.Infof("buyer contract already fulfilling")
		return
	}
	p.log.Infof("buyer contract started fulfilling")
	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.doneCh = make(chan struct{})

	go func() {
		p.state.Store(resources.ContractStateRunning)
		p.err = p.run(ctx)
		close(p.doneCh)
		p.state.Store(resources.ContractStatePending)
	}()
}

func (p *ContractWatcherBuyer) StopFulfilling() {
	p.cancel()
	<-p.doneCh
	p.log.Infof("buyer contract stopped fulfilling")
}

func (p *ContractWatcherBuyer) Done() <-chan struct{} {
	return p.doneCh
}

func (p *ContractWatcherBuyer) Err() error {
	if errors.Is(p.err, context.Canceled) {
		return ErrContractClosed
	}
	return p.err
}

func (p *ContractWatcherBuyer) SetData(terms *hashrateContract.Terms) {
	p.Terms = terms
}

func (p *ContractWatcherBuyer) run(ctx context.Context) error {
	p.state.Store(resources.ContractStateRunning)
	startedAt := time.Now()
	p.fulfillmentStartedAt.Store(startedAt)

	p.globalHashrate.Reset(p.ID())
	p.globalHashrate.Initialize(p.ID())

	ticker := time.NewTicker(p.contractCycleDuration)
	defer ticker.Stop()

	tillEndTime := p.getUntilContractEnd()
	if tillEndTime <= 0 {
		return nil
	}
	endTimer := time.NewTimer(tillEndTime)

	for {
		err := p.checkIncomingHashrate(ctx)
		if err != nil {
			return err
		}

		tillEndTime := p.getUntilContractEnd()
		if tillEndTime <= 0 {
			return nil
		}
		endTimer.Reset(tillEndTime)

		select {
		case <-ctx.Done():
			if !endTimer.Stop() {
				<-endTimer.C
			}
			return ctx.Err()
		case <-endTimer.C:
			return nil
		case <-ticker.C:
			if !endTimer.Stop() {
				<-endTimer.C
			}
		}
	}
}

func (p *ContractWatcherBuyer) proceedToNextStage() {
	if p.isContractExpired() {
		p.validationStage.Store(hashrateContract.ValidationStageFinished)
		p.log.Infof("new validation stage %s", p.validationStage.Load().String())
		return
	}
}

func (p *ContractWatcherBuyer) checkIncomingHashrate(ctx context.Context) error {
	p.proceedToNextStage()

	isHashrateOK := p.isReceivingAcceptableHashrate()

	switch p.validationStage.Load() {
	case hashrateContract.ValidationStageValidating:
		lastShareTime, ok := p.globalHashrate.GetLastSubmitTime(p.getWorkerName())
		if !ok {
			lastShareTime = p.fulfillmentStartedAt.Load()
		}
		if time.Since(lastShareTime) > p.shareTimeout {
			return fmt.Errorf("no share submitted within shareTimeout (%s), lastShare at (%s)", p.shareTimeout, lastShareTime.Format(time.RFC3339))
		}

		if !isHashrateOK {
			return fmt.Errorf("contract is not delivering accurate hashrate")
		}
		return nil
	case hashrateContract.ValidationStageFinished:
		return fmt.Errorf("contract is finished")
	default:
		return fmt.Errorf("unknown validation state")
	}
}

func (p *ContractWatcherBuyer) isReceivingAcceptableHashrate() bool {
	actualHashrate, ok := p.globalHashrate.GetHashRateGHS(p.getWorkerName(), p.hashrateCounterNameBuyer)
	if !ok {
		p.log.Warnf("no hashrate submitted yet")
	}
	targetHashrateGHS := p.HashrateGHS()

	starvingGHS := math.Max(targetHashrateGHS-actualHashrate, 0.0)
	p.starvingGHS.Store(uint64(starvingGHS))
	fulfilmentElapsed := time.Since(p.fulfillmentStartedAt.Load())

	hrError := lib.RelativeError(targetHashrateGHS, actualHashrate)
	maxHrError := GetMaxGlobalError(fulfilmentElapsed, p.hrErrorThreshold, p.hrValidationFlatness, 5*time.Minute)

	totalShares := 0
	worker := p.globalHashrate.GetWorker(p.getWorkerName())
	if worker != nil {
		totalShares = worker.GetTotalShares()
	}

	hrMsg := fmt.Sprintf(
		"elapsed %s target GHS %.0f, actual GHS %.0f, error %.0f%%, threshold(%.0f%%) totalShares(%d)",
		fulfilmentElapsed.Round(time.Second), targetHashrateGHS, actualHashrate, hrError*100, maxHrError*100, totalShares,
	)

	if hrError <= maxHrError {
		p.log.Infof("contract is delivering accurately: %s", hrMsg)
		return true
	}

	if actualHashrate > targetHashrateGHS {
		p.log.Infof("contract is overdelivering: %s", hrMsg)
		// contract overdelivery is ok for buyer
		return true
	}

	p.log.Warnf("contract is underdelivering: %s, %f", hrMsg, hrError-maxHrError)
	return false
}

func (p *ContractWatcherBuyer) getUntilContractEnd() time.Duration {
	return time.Until(p.EndTime())
}

func (p *ContractWatcherBuyer) isContractExpired() bool {
	return time.Now().After(p.EndTime())
}

func (p *ContractWatcherBuyer) getWorkerName() string {
	return p.ID()
}

// public getters

func (p *ContractWatcherBuyer) Role() resources.ContractRole {
	return p.role
}

func (p *ContractWatcherBuyer) FulfillmentStartTime() time.Time {
	return p.fulfillmentStartedAt.Load()
}

func (p *ContractWatcherBuyer) State() resources.ContractState {
	return p.state.Load()
}

func (p *ContractWatcherBuyer) ValidationStage() hashrateContract.ValidationStage {
	return p.validationStage.Load()
}

func (p *ContractWatcherBuyer) ResourceEstimates() map[string]float64 {
	return map[string]float64{
		ResourceEstimateHashrateGHS: p.HashrateGHS(),
	}
}

func (p *ContractWatcherBuyer) ResourceEstimatesActual() map[string]float64 {
	res, _ := p.globalHashrate.GetHashRateGHSAll(p.getWorkerName())
	return res
}

func (p *ContractWatcherBuyer) ResourceType() string {
	return ResourceTypeHashrate
}

func (p *ContractWatcherBuyer) Dest() string {
	// the destination is localhost for the buyer
	return ""
}

func (p *ContractWatcherBuyer) PoolDest() string {
	url := p.Terms.DestinationURL
	if url == nil {
		return ""
	}
	return url.String()
}

func (p *ContractWatcherBuyer) StarvingGHS() int {
	return int(p.starvingGHS.Load())
}

func (p *ContractWatcherBuyer) Error() error {
	return p.contractErr.Load()
}
