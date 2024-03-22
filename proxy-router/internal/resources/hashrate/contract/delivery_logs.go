package contract

import (
	"sync"
	"time"
)

type DeliveryLogEntry struct {
	Timestamp                         time.Time
	ActualGHS                         int
	FullMinersGHS                     int
	FullMiners                        []string
	FullMinersShares                  int
	PartialMinersGHS                  int
	PartialMiners                     []string
	PartialMinersShares               int
	UnderDeliveryGHS                  int
	GlobalHashrateGHS                 int
	GlobalUnderDeliveryGHS            int
	GlobalError                       float64
	NextCyclePartialDeliveryTargetGHS int
}

type DeliveryLog struct {
	Entries []DeliveryLogEntry
	mutex   sync.RWMutex
}

func NewDeliveryLog() *DeliveryLog {
	return &DeliveryLog{
		Entries: make([]DeliveryLogEntry, 0, 100),
	}
}

func (l *DeliveryLog) AddEntry(entry DeliveryLogEntry) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.Entries = append(l.Entries, entry)
}

func (l *DeliveryLog) GetEntries() ([]DeliveryLogEntry, error) {
	// copy values for safety
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	data := make([]DeliveryLogEntry, len(l.Entries))
	copy(data, l.Entries)

	return data, nil
}
