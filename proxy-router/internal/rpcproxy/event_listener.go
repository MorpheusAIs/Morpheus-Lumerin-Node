package rpcproxy

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/repositories/contracts"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/storages"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EventsListener struct {
	sessionRouter *registries.SessionRouter
	store         *storages.SessionStorage
	tsk           *lib.Task
	log           *lib.Logger
	client        *ethclient.Client
}

func NewEventsListener(client *ethclient.Client, store *storages.SessionStorage, sessionRouter *registries.SessionRouter, log *lib.Logger) *EventsListener {
	return &EventsListener{
		store:         store,
		log:           log,
		sessionRouter: sessionRouter,
		client:        client,
	}
}

func (e *EventsListener) Run(ctx context.Context) error {
	defer func() {
		_ = e.log.Close()
	}()

	sub, err := contracts.WatchContractEvents(ctx, e.client, e.sessionRouter.GetContractAddress(), contracts.CreateEventMapper(contracts.BlockchainEventFactory, e.sessionRouter.GetABI()), e.log)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()
	e.log.Infof("started watching events, address %s", e.sessionRouter.GetContractAddress())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-sub.Events():
			err := e.controller(ctx, event)
			if err != nil {
				e.log.Errorf("error loading data: %s", err)
			}
		case err := <-sub.Err():
			return err
		}
	}
}

func (e *EventsListener) controller(ctx context.Context, event interface{}) error {
	switch ev := event.(type) {
	case *sessionrouter.SessionRouterSessionOpened:
		return e.handleSessionOpened(ctx, ev)
	}
	return nil
}

func (e *EventsListener) handleSessionOpened(ctx context.Context, event *sessionrouter.SessionRouterSessionOpened) error {
	sessionId := lib.BytesToString(event.SessionId[:])
	e.log.Debugf("received open session router event, sessionId %s", sessionId)
	// session, err := e.sessionRouter.GetSession(ctx, sessionId)
	// if err != nil {
	// 	return err
	// }

	e.store.AddSession(&storages.Session{
		Id:       sessionId,
		UserAddr: event.UserAddress.Hex(),
	})

	return nil
}
