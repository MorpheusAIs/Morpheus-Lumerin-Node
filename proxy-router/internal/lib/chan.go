package lib

import "sync"

func Merge[T any](cs ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan T) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

type ChanRecvStop[T any] struct {
	DataCh chan T
	StopCh chan struct{}
}

// NewChanRecvStop creates a new ChanRecvStop, an abstraction over a channel that
// supports multiple writers and a single reader. The writes can be stopped and/or
// unblocked by calling the Stop method from a reader
func NewChanRecvStop[T any]() *ChanRecvStop[T] {
	return &ChanRecvStop[T]{
		DataCh: make(chan T),
		StopCh: make(chan struct{}),
	}
}

func (c *ChanRecvStop[T]) Receive() <-chan T {
	return c.DataCh
}

func (c *ChanRecvStop[T]) Stop() {
	close(c.StopCh)
}

func (c *ChanRecvStop[T]) Send(data T) {
	select {
	case <-c.StopCh:
	case c.DataCh <- data:
	}
}
