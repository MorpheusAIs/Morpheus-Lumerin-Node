package hashrate

import (
	"sync/atomic"
	"time"

	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/lib"
)

type GlobalHashrate struct {
	data      *lib.Collection[*WorkerHashrateModel]
	hrFactory HashrateFactory
}

func NewGlobalHashrate(hrFactory HashrateFactory) *GlobalHashrate {
	return &GlobalHashrate{
		data:      lib.NewCollection[*WorkerHashrateModel](),
		hrFactory: hrFactory,
	}
}

func (t *GlobalHashrate) Initialize(workerName string) {
	t.data.LoadOrStore(NewWorkerHashrateModel(workerName, t.hrFactory()))
}

func (t *GlobalHashrate) OnSubmit(workerName string, diff float64) {
	actual, _ := t.data.LoadOrStore(NewWorkerHashrateModel(workerName, t.hrFactory()))
	actual.OnSubmit(diff)
}

func (t *GlobalHashrate) OnConnect(workerName string) {
	actual, _ := t.data.LoadOrStore(NewWorkerHashrateModel(workerName, t.hrFactory()))
	actual.OnConnect()
}

func (t *GlobalHashrate) GetLastSubmitTime(workerName string) (tm time.Time, ok bool) {
	record, ok := t.data.Load(workerName)
	if !ok {
		return time.Time{}, false
	}
	time := record.hr.GetLastSubmitTime()
	return time, !time.IsZero()
}

func (t *GlobalHashrate) GetHashRateGHS(workerName string, counterID string) (hrGHS float64, ok bool) {
	record, ok := t.data.Load(workerName)
	if !ok {
		return 0, false
	}
	return record.GetHashRateGHS(counterID)
}

func (t *GlobalHashrate) GetHashRateGHSAll(workerName string) (hrGHS map[string]float64, ok bool) {
	record, ok := t.data.Load(workerName)
	if !ok {
		return nil, false
	}
	return record.GetHashrateAvgGHSAll(), true
}

func (t *GlobalHashrate) GetTotalWork(workerName string) (work float64, ok bool) {
	record, ok := t.data.Load(workerName)
	if !ok {
		return 0, false
	}
	return record.hr.GetTotalWork(), true
}

func (t *GlobalHashrate) GetAll() map[string]time.Time {
	data := make(map[string]time.Time)
	t.data.Range(func(item *WorkerHashrateModel) bool {
		data[item.ID()] = item.hr.GetLastSubmitTime()
		return true
	})
	return data
}

func (t *GlobalHashrate) Range(f func(m *WorkerHashrateModel) bool) {
	t.data.Range(func(item *WorkerHashrateModel) bool {
		return f(item)
	})
}

func (t *GlobalHashrate) Reset(workerName string) {
	t.data.Delete(workerName)
}

func (t *GlobalHashrate) GetWorker(workerName string) *WorkerHashrateModel {
	var worker *WorkerHashrateModel
	t.Range(func(item *WorkerHashrateModel) bool {
		if item.id == workerName {
			worker = item
			return false
		}
		return true
	})
	return worker
}

type WorkerHashrateModel struct {
	id         string
	hr         *Hashrate
	reconnects *atomic.Uint32
}

func NewWorkerHashrateModel(id string, hr *Hashrate) *WorkerHashrateModel {
	return &WorkerHashrateModel{
		id:         id,
		hr:         hr,
		reconnects: &atomic.Uint32{},
	}
}

func (m *WorkerHashrateModel) ID() string {
	return m.id
}

func (m *WorkerHashrateModel) OnSubmit(diff float64) {
	m.hr.OnSubmit(diff)
}

func (m *WorkerHashrateModel) OnConnect() {
	m.reconnects.Add(1)
}

func (m *WorkerHashrateModel) GetHashRateGHS(counterID string) (float64, bool) {
	return m.hr.GetHashrateAvgGHSCustom(counterID)
}

func (m *WorkerHashrateModel) GetHashrateAvgGHSAll() map[string]float64 {
	return m.hr.GetHashrateAvgGHSAll()
}

func (m *WorkerHashrateModel) GetLastSubmitTime() time.Time {
	return m.hr.GetLastSubmitTime()
}

func (m *WorkerHashrateModel) GetHashrateCounter(counterID string) Counter {
	return m.hr.custom[counterID]
}

func (m *WorkerHashrateModel) Reconnects() int {
	return int(m.reconnects.Load())
}

func (m *WorkerHashrateModel) GetTotalShares() int {
	return m.hr.GetTotalShares()
}
