// Package avgcounter implements a simple EMA (Exponential Moving Average)
// counter. The New function creates a counter with the only parameter:
// avgInterval. Every Add call adds the value to the counter. The current value
// can be obtained using the Value method.
//
// The counter holds the exponentially (by time) weighted average of all added
// values.
//
// https://github.com/davidmz/avgcounter
package hashrate

import (
	"math"
	"sync"
	"time"
)

var nowTime time.Time

func getNow() time.Time {
	if nowTime.IsZero() {
		return time.Now()
	}
	return nowTime
}

// func sleepNow(time time.Duration) {
// 	nowTime = nowTime.Add(time)
// }

// Ema is an EMA (Exponential Moving Average) counter.
type Ema struct {
	lastValue float64
	lastTime  time.Time
	halfLife  time.Duration
	mutex     sync.RWMutex
}

// NewEma creates a new Counter with the given half-life (time lag at which the exponential weights decay by one half)
func NewEma(halfLife time.Duration) *Ema {
	return &Ema{halfLife: halfLife}
}

func (c *Ema) Start() {
	// noop
}

// Value returns the current value of the counter.
func (c *Ema) Value() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value()
}

// LastValue returns last value of a counter excluding the value decay
func (c *Ema) LastValue() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.valueAfter(0)
}

// ValuePer returns the current value of the counter, normalized to the given
// interval. It is actually a Value() * interval / avgInterval.
func (c *Ema) ValuePer(interval time.Duration) float64 {
	return c.Value() * float64(interval) / float64(c.halfLife)
}

func (c *Ema) LastValuePer(interval time.Duration) float64 {
	return c.valueAfter(0) * float64(interval) / float64(c.halfLife)
}

// Add adds a new value to the counter.
func (c *Ema) Add(v float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastValue = c.value() + v
	c.lastTime = getNow()
}

// Private methods

func (c *Ema) value() float64 {
	return c.valueAfter(getNow().Sub(c.lastTime))
}

// calculates value decay
func (c *Ema) valueAfter(elapsed time.Duration) float64 {
	if c.lastValue == 0 {
		return 0
	}

	w := math.Exp(-float64(elapsed) / float64(c.halfLife))

	return c.lastValue * w
}

func (c *Ema) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastValue = 0
	c.lastTime = time.Time{}
}

var _ Counter = new(Ema)
