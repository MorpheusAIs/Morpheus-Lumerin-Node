package lib

import "github.com/ethereum/go-ethereum/event"

type Subscription struct {
	sub event.Subscription
	ch  chan interface{}
}

func NewSubscription(producer func(quit <-chan struct{}) error, ch chan interface{}) *Subscription {
	return &Subscription{
		sub: event.NewSubscription(producer),
		ch:  ch,
	}
}

func (s *Subscription) Unsubscribe() {
	s.sub.Unsubscribe()
}

func (s *Subscription) Err() <-chan error {
	return s.sub.Err()
}

func (s *Subscription) Events() <-chan interface{} {
	return s.ch
}

func (s *Subscription) Ch() chan interface{} {
	return s.ch
}
