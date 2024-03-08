package lib

import (
	"sync/atomic"
)

type AtomicValue[T any] struct {
	v atomic.Value
}

func NewAtomicValue[T any](v T) *AtomicValue[T] {
	val := atomic.Value{}
	val.Store(v)
	return &AtomicValue[T]{v: val}
}

func (a *AtomicValue[T]) Load() T {
	return a.v.Load().(T)
}

func (a *AtomicValue[T]) Store(v T) {
	a.v.Store(v)
}

func (a *AtomicValue[T]) CompareAndSwap(old, new T) bool {
	return a.v.CompareAndSwap(old, new)
}

func (a *AtomicValue[T]) Swap(v T) T {
	return a.v.Swap(v).(T)
}
