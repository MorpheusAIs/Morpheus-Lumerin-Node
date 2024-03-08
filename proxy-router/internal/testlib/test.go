package testlib

import (
	"sync"
	"testing"
	"time"
)

func Repeat(t *testing.T, n int, f func(*testing.T)) {
	for i := 0; i < n; i++ {
		f(t)
	}
}

func RepeatConcurrent(t *testing.T, n int, f func(*testing.T)) {
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f(t)
		}()
	}
	wg.Wait()
}

func RepeatConcurrentTimeout(t *testing.T, n int, timeout time.Duration, f func(*testing.T)) {
	doneCh := make(chan struct{})
	go func() {
		wg := sync.WaitGroup{}
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				f(t)
			}()
		}
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-time.After(timeout):
		t.Fatalf("test timeout after %s", timeout)
	case <-doneCh:
	}
}
