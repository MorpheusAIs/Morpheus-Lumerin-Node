package hashrate

import (
	"time"
)

type Counter interface {
	Start()                           // sets the start time
	Add(v float64)                    // adds a measurment performed now to the counter
	Value() float64                   // returns the current value
	ValuePer(t time.Duration) float64 // returns the current value normalized to the given duration
	Reset()                           // resets counter
}

type HashrateFactory = func() *Hashrate
