package lib

import "sync"

type Set struct {
	m *sync.Map //map[string]struct{}
}

func NewSet() Set {
	s := &sync.Map{}
	return Set{m: s}
}

func NewSetFromSlice(slice []string) Set {
	s := NewSet()
	for _, v := range slice {
		s.Add(v)
	}
	return s
}

func (s Set) Add(value ...string) {
	for _, v := range value {
		s.m.Store(v, struct{}{})
	}
}

func (s Set) Remove(value string) bool {
	_, loaded := s.m.LoadAndDelete(value)
	return loaded
}

func (s Set) Contains(value string) bool {
	_, c := s.m.Load(value)
	return c
}

func (s Set) Len() int {
	var counter int
	s.m.Range(func(_, _ interface{}) bool {
		counter++
		return true
	})
	return counter
}

func (s Set) ToSlice() []string {
	var keys []string
	s.m.Range(func(k, _ interface{}) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func (s Set) Clear() {
	s.m.Range(func(k, _ interface{}) bool {
		s.m.Delete(k)
		return true
	})
}
