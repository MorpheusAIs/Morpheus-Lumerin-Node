package hashrate

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGlobalHashrate(t *testing.T) {
	// nowTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	HashrateCounterDefault := "ema-5m"
	HashrateCounterDefaultDuration := 1 * time.Second
	threads := 10
	diff := 100000.0
	loops := 10
	threadSleep := 50 * time.Millisecond
	workerName := "kiki"

	hashrateFactory := func() *Hashrate {
		return NewHashrate(
			map[string]Counter{
				HashrateCounterDefault: NewEma(HashrateCounterDefaultDuration),
				"ema-10m":              NewEma(10 * time.Minute),
				"ema-30m":              NewEma(30 * time.Minute),
			},
		)
	}

	hr := NewGlobalHashrate(hashrateFactory)

	cb := func(thread int) {
		for i := 0; i < loops; i++ {
			hr.OnSubmit(workerName, diff)
			time.Sleep(threadSleep)
		}
	}

	duration := threadSleep * time.Duration(loops) * 2
	end := time.After(duration)
	wg := sync.WaitGroup{}
	for i := 0; i < threads; i++ {
		wg.Add(1)
		k := i
		go func() {
			defer wg.Done()
			cb(k)
		}()
	}

	wg.Wait()
	<-end

	totalWork := float64(threads * loops * int(diff))
	workPerSecond := totalWork / duration.Seconds()
	expected := HSToGHS(JobSubmittedToHS(workPerSecond))
	actual, _ := hr.GetHashRateGHS(workerName, "mean")

	fmt.Printf("exp %d act %.0f\n", expected, actual)
	assert.InEpsilon(t, expected, actual, 0.01, "should be accurate")
}
