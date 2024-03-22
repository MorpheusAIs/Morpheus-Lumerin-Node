package allocator

import (
	"net/url"
	"sync"
	"time"

	"github.com/gammazero/deque"
	"go.uber.org/atomic"
)

type OnSubmitCb func(diff float64, ID string)
type OnDisconnectCb func(ID string, HrGHS float64, remainingJob float64)
type OnEndCb func(ID string, HrGHS float64, remainingJob float64, err error)

type MinerTask struct {
	ID           string
	Dest         *url.URL
	Job          float64
	Deadline     time.Time
	OnSubmit     OnSubmitCb
	OnDisconnect OnDisconnectCb
	OnEnd        OnEndCb

	RemainingJobToSubmit *atomic.Int64
	cancelCh             chan struct{}
}

func NewTask(ID string, dest *url.URL, job float64, deadline time.Time, onSubmit OnSubmitCb, onDisconnect OnDisconnectCb, onEnd OnEndCb) *MinerTask {
	return &MinerTask{
		ID:           ID,
		Dest:         dest,
		Job:          job,
		Deadline:     deadline,
		OnSubmit:     onSubmit,
		OnDisconnect: onDisconnect,
		OnEnd:        onEnd,

		RemainingJobToSubmit: atomic.NewInt64(int64(job)),
		cancelCh:             make(chan struct{}),
	}
}

func (t *MinerTask) RemainingJob() float64 {
	return float64(t.RemainingJobToSubmit.Load())
}

func (t *MinerTask) Cancel() (firstCancel bool) {
	select {
	case <-t.cancelCh:
		return false
	default:
		close(t.cancelCh)
		return true
	}
}

type TaskList struct {
	tasks     *deque.Deque[*MinerTask]
	mutex     sync.RWMutex
	size      atomic.Int32
	taskTaken bool
}

func NewTaskList() *TaskList {
	return &TaskList{
		tasks:     deque.New[*MinerTask](),
		mutex:     sync.RWMutex{},
		taskTaken: false,
	}
}

func (p *TaskList) Add(ID string, dest *url.URL, job float64, deadline time.Time, onSubmit OnSubmitCb, onDisconnect OnDisconnectCb, onEnd OnEndCb) int {
	task := NewTask(ID, dest, job, deadline, onSubmit, onDisconnect, onEnd)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.tasks.PushBack(task)
	p.size.Inc()

	return p.tasks.Len()
}

// returns the first element of the task queue
func (p *TaskList) LockNextTask() (t *MinerTask, ok bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.taskTaken {
		panic("task already taken")
	}

	if p.tasks.Len() == 0 {
		return nil, false
	}

	p.taskTaken = true
	return p.tasks.Front(), true
}

// removes lock and removes from the task queue
func (p *TaskList) UnlockAndRemove() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.taskTaken {
		panic("task not taken")
	}
	p.taskTaken = false

	if p.tasks.Len() == 0 {
		panic("no tasks in queue, when there should be at least one")
	}
	p.tasks.PopFront()
	p.size.Dec()
}

func (p *TaskList) Unlock() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.taskTaken {
		panic("task not taken")
	}
	p.taskTaken = false
}

func (p *TaskList) Size() int {
	return int(p.size.Load())
}

func (p *TaskList) CancelAll() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.taskTaken {
		p.tasks.Front().Cancel()
	}

	p.tasks.Clear()
	p.size.Store(0)
}

func (p *TaskList) Cancel(contractID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for i := 0; i < p.tasks.Len(); i++ {
		task := p.tasks.At(i)
		if task.ID == contractID {
			if i == 0 && p.taskTaken {
				p.tasks.Front().Cancel()
			} else {
				p.tasks.Remove(i)
				p.size.Dec()
			}
		}
	}
}

func (p *TaskList) Range(f func(task *MinerTask) bool) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for i := 0; i < p.tasks.Len(); i++ {
		if !f(p.tasks.At(i)) {
			return
		}
	}
}
