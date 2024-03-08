package allocator

import (
	"fmt"
	"math"
	"net/url"
	"sync"
	"time"

	gi "gitlab.com/TitanInd/proxy/proxy-router-v3/internal/interfaces"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/resources/hashrate/hashrate"
	"golang.org/x/exp/slices"
)

const (
	HashratePredictionAdjustment = 1.0
	AllocationMinDuration        = 5 * time.Second
	AllocationMinJob             = 5000.0
)

type minerSnapshot struct {
	fullMiners    []MinerItem
	partialMiners []MinerItem
	freeMiners    []MinerItem
}

type MinerItem struct {
	ID            string
	HrGHS         float64
	JobRemaining  float64
	TimeRemaining time.Duration
	IsFullMiner   bool
}

type ListenerHandle int

type MinerItemJobScheduled struct {
	ID       string
	Job      float64
	Fraction float64
}

type MinerIDJob = map[string]float64

type Allocator struct {
	// read/write
	lastListenerID  int
	vettedListeners map[int]func(ID string)
	vettedMutex     sync.RWMutex

	// read only
	proxies *lib.Collection[*Scheduler]
	log     gi.ILogger
}

func NewAllocator(proxies *lib.Collection[*Scheduler], log gi.ILogger) *Allocator {
	return &Allocator{
		proxies:         proxies,
		vettedListeners: make(map[int]func(ID string), 0),
		log:             log,
	}
}

func (p *Allocator) GetMiners() *lib.Collection[*Scheduler] {
	return p.proxies
}

func (p *Allocator) AllocateFullMinersForHR(
	ID string,
	hrGHS float64,
	dest *url.URL,
	duration time.Duration,
	onSubmit OnSubmitCb,
	onDisconnect OnDisconnectCb,
	onEnd OnEndCb,
) (minerIDs []string, deltaGHS float64) {
	miners := p.getMinersSnapshot(0)
	p.log.Infow(fmt.Sprintf("available free miners %v", miners.freeMiners), "CtrAddr", lib.AddrShort(ID))

	for _, miner := range miners.freeMiners {
		minerGHS := miner.HrGHS
		if minerGHS <= hrGHS && minerGHS > 0 {
			proxy, ok := p.proxies.Load(miner.ID)
			if ok && !proxy.IsDisconnecting() {
				proxy.AddTask(ID, dest, hashrate.GHSToJobSubmittedV2(minerGHS, duration), onSubmit, onDisconnect, onEnd, time.Now().Add(duration))
				minerIDs = append(minerIDs, miner.ID)
				hrGHS -= minerGHS
				p.log.Infow(fmt.Sprintf("full miner %s allocated for %.0f GHS", miner.ID, minerGHS), "CtrAddr", lib.AddrShort(ID))
			}
		}
	}

	return minerIDs, hrGHS
}

func (p *Allocator) AllocatePartialForJob(
	ID string,
	jobNeeded float64,
	dest *url.URL,
	cycleEndTimeout time.Duration,
	onSubmit func(diff float64, ID string),
	onDisconnect func(ID string, hrGHS float64, remainingJob float64),
	onEnd OnEndCb,
) (minerIDJob MinerIDJob, remainderGHS float64) {
	p.log.Infof("attempting to partially allocate job %.f", jobNeeded)

	miners := p.getMinersSnapshot(cycleEndTimeout)
	p.log.Infof("available partial miners %v", miners.partialMiners)

	minerIDJob = MinerIDJob{}

	for _, miner := range miners.partialMiners {
		if jobNeeded < AllocationMinJob {
			return minerIDJob, 0
		}
		if miner.JobRemaining < AllocationMinJob {
			continue
		}
		if miner.TimeRemaining < AllocationMinDuration {
			continue
		}
		durationToDoJobWithMiner := time.Duration(jobNeeded / hashrate.GHSToHS(int(miner.HrGHS)) * float64(time.Second))
		if durationToDoJobWithMiner < AllocationMinDuration {
			continue
		}

		// try to add the whole chunk and return
		if miner.JobRemaining >= jobNeeded {
			m, ok := p.proxies.Load(miner.ID)
			if ok && !m.IsDisconnecting() {
				m.AddTask(ID, dest, jobNeeded, onSubmit, onDisconnect, onEnd, time.Now().Add(cycleEndTimeout))
				minerIDJob[miner.ID] = jobNeeded
				return minerIDJob, 0
			}
		}

		// try to add at least a minJob and continue
		if miner.JobRemaining >= AllocationMinJob {
			m, ok := p.proxies.Load(miner.ID)
			if ok && !m.IsDisconnecting() {
				m.AddTask(ID, dest, miner.JobRemaining, onSubmit, onDisconnect, onEnd, time.Now().Add(cycleEndTimeout))
				minerIDJob[miner.ID] = miner.JobRemaining
				jobNeeded -= miner.JobRemaining
			}
		}
	}

	// search in free miners
	// missing loop cause we already checked full miners
	p.log.Infof("available free miners %v", miners.freeMiners)
	for _, miner := range miners.freeMiners {
		if jobNeeded < AllocationMinJob {
			jobNeeded = 0
			break
		}

		minerJobRemaining := hashrate.GHSToJobSubmittedV2(miner.HrGHS, cycleEndTimeout)
		if minerJobRemaining <= AllocationMinJob {
			continue
		}

		jobToAllocate := math.Min(minerJobRemaining, jobNeeded)

		m, ok := p.proxies.Load(miner.ID)
		if !ok || m.IsDisconnecting() {
			continue
		}

		m.AddTask(ID, dest, jobToAllocate, onSubmit, onDisconnect, onEnd, time.Now().Add(cycleEndTimeout))
		minerIDJob[miner.ID] = jobToAllocate
		jobNeeded -= jobToAllocate
	}

	return minerIDJob, jobNeeded
}

func (p *Allocator) GetMinersFulfillingContract(contractID string, cycleDuration time.Duration) []*MinerItemJobScheduled {
	return []*MinerItemJobScheduled{}
	// Temporary disabling this function to minimize usage of mutexes
	// TODO: rewrite it to use a view of the collection instead of mutexes

	// p.GetMiners().Range(func(item *Scheduler) bool {
	// 	if item.IsVetting() {
	// 		return true
	// 	}

	// 	if item.IsDisconnecting() {
	// 		return true
	// 	}

	// 	tasks := item.GetTasksByID(contractID)
	// 	maxJob := item.getExpectedCycleJob(cycleDuration)

	// 	for _, task := range tasks {
	// 		job := float64(task.RemainingJobToSubmit.Load())
	// 		minerItems = append(minerItems, &MinerItemJobScheduled{
	// 			ID:       item.ID(),
	// 			Job:      job,
	// 			Fraction: job / maxJob,
	// 		})
	// 	}
	// 	return true
	// })

	// return minerItems
}

func (p *Allocator) AddVettedListener(f func(ID string)) ListenerHandle {
	p.vettedMutex.Lock()
	defer p.vettedMutex.Unlock()

	ID := p.lastListenerID
	p.lastListenerID++
	p.vettedListeners[ID] = f

	return ListenerHandle(ID)
}

func (p *Allocator) RemoveVettedListener(s ListenerHandle) {
	p.vettedMutex.Lock()
	defer p.vettedMutex.Unlock()

	delete(p.vettedListeners, int(s))
}

func (p *Allocator) InvokeVettedListeners(minerID string) {
	p.vettedMutex.RLock()
	defer p.vettedMutex.RUnlock()

	for _, f := range p.vettedListeners {
		go f(minerID)
	}
}

func (p *Allocator) getMinersSnapshot(remainingCycleDuration time.Duration) minerSnapshot {
	snap := minerSnapshot{}

	p.proxies.Range(func(item *Scheduler) bool {
		if item.IsVetting() { // atomic
			return true
		}
		if item.IsDisconnecting() { // atomic
			return true
		}
		if item.IsFree() { // has mutex inside
			snap.freeMiners = append(snap.freeMiners, MinerItem{
				ID:            item.ID(),
				HrGHS:         item.HashrateGHS() * HashratePredictionAdjustment,
				JobRemaining:  hashrate.GHSToJobSubmittedV2(item.HashrateGHS(), remainingCycleDuration),
				TimeRemaining: remainingCycleDuration,
				IsFullMiner:   true,
			})
		}
		if remainingCycleDuration == 0 {
			return true
		}
		if item.IsPartialBusy(remainingCycleDuration) {
			jobRemaining := item.GetJobCouldBeScheduledTill(remainingCycleDuration)
			timeRemaining := time.Duration(hashrate.JobSubmittedToGHS(jobRemaining) / item.HashrateGHS() * float64(time.Second))
			snap.partialMiners = append(snap.partialMiners, MinerItem{
				ID:            item.ID(),
				HrGHS:         item.HashrateGHS(),
				JobRemaining:  jobRemaining,
				TimeRemaining: timeRemaining,
				IsFullMiner:   false,
			})
		}
		return true
	})

	slices.SortStableFunc(snap.freeMiners, func(i, j MinerItem) bool {
		return i.HrGHS > j.HrGHS
	})

	slices.SortStableFunc(snap.partialMiners, func(i, j MinerItem) bool {
		return i.JobRemaining < j.JobRemaining
	})

	return snap
}
