package lib

import (
	"sync"
)

type IModel interface {
	ID() string
}

type Collection[T IModel] struct {
	items sync.Map
}

func NewCollection[T IModel]() *Collection[T] {
	return &Collection[T]{
		items: sync.Map{},
	}
}

func (p *Collection[T]) Load(ID string) (item T, ok bool) {
	if val, ok := p.items.Load(ID); ok {
		return val.(T), true
	} else {
		if val != nil {
			return val.(T), false
		}
		return *new(T), false
	}
}

func (p *Collection[T]) Range(f func(item T) bool) {
	p.items.Range(func(key, value any) bool {
		item := value.(T)
		return f(item)
	})
}

func (p *Collection[T]) Store(item T) {
	p.items.Store(item.ID(), item)
}

func (p *Collection[T]) LoadOrStore(item T) (actual T, loaded bool) {
	act, load := p.items.LoadOrStore(item.ID(), item)
	return act.(T), load
}

func (p *Collection[T]) Delete(ID string) {
	p.items.Delete(ID)
}

// Len returns the number of items in the collection. It is O(n)
func (p *Collection[T]) Len() int {
	count := 0
	p.items.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}
