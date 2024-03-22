package lib

import (
	"context"
	"time"
)

var ErrTimeout = context.DeadlineExceeded

// Mutex is a non blocking mutex
type Mutex struct {
	mut chan struct{}
}

func NewMutex() Mutex {
	return Mutex{
		mut: make(chan struct{}, 1),
	}
}

func (m *Mutex) LockCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case m.mut <- struct{}{}:
		return nil
	}
}

func (m *Mutex) Lock() {
	m.mut <- struct{}{}
}

func (m *Mutex) LockTimeout(d time.Duration) error {
	select {
	case <-time.After(d):
		return ErrTimeout
	case m.mut <- struct{}{}:
		return nil
	}
}

func (m *Mutex) Unlock() {
	select {
	case <-m.mut:
	default:
	}
}
