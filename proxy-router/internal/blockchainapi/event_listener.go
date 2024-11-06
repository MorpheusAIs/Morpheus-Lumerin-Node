package blockchainapi

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/ethereum/go-ethereum/common"
)

type EventsListener struct {
	sessionRouter *registries.SessionRouter
	sessionRepo   *sessionrepo.SessionRepositoryCached
	tsk           *lib.Task
	log           *lib.Logger
	wallet        interfaces.Wallet
	logWatcher    contracts.LogWatcher

	//internal state
	addr common.Address
}

func NewEventsListener(sessionRepo *sessionrepo.SessionRepositoryCached, sessionRouter *registries.SessionRouter, wallet interfaces.Wallet, logWatcher contracts.LogWatcher, log *lib.Logger) *EventsListener {
	return &EventsListener{
		log:           log,
		sessionRouter: sessionRouter,
		sessionRepo:   sessionRepo,
		wallet:        wallet,
		logWatcher:    logWatcher,
	}
}

func (e *EventsListener) Run(ctx context.Context) error {
	defer func() {
		_ = e.log.Close()
	}()

	addr, err := e.getWalletAddress()
	if err != nil {
		return err
	}
	e.addr = addr

	//TODO: filter events by user/provider address
	sub, err := e.logWatcher.Watch(
		ctx,
		e.sessionRouter.GetContractAddress(),
		contracts.CreateEventMapper(contracts.SessionRouterEventFactory,
			e.sessionRouter.GetABI(),
		), nil)
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
			err := e.controller(event)
			if err != nil {
				e.log.Errorf("error handling event: %s", err)
			}
		case err := <-sub.Err():
			e.log.Errorf("error in event listener: %s", err)
			return err
		}
	}
}

func (e *EventsListener) controller(event interface{}) error {
	switch ev := event.(type) {
	case *sessionrouter.SessionRouterSessionOpened:
		return e.handleSessionOpened(ev)
	case *sessionrouter.SessionRouterSessionClosed:
		return e.handleSessionClosed(ev)
	}
	return nil
}

func (e *EventsListener) handleSessionOpened(event *sessionrouter.SessionRouterSessionOpened) error {
	if !e.filter(event.ProviderId, event.User) {
		return nil
	}
	e.log.Debugf("received open session event, sessionId %s", lib.BytesToString(event.SessionId[:]))
	return e.sessionRepo.RefreshSession(context.Background(), event.SessionId)
}

func (e *EventsListener) handleSessionClosed(event *sessionrouter.SessionRouterSessionClosed) error {
	if !e.filter(event.ProviderId, event.User) {
		return nil
	}
	e.log.Debugf("received close session event, sessionId %s", lib.BytesToString(event.SessionId[:]))
	return e.sessionRepo.RemoveSession(context.Background(), event.SessionId)
}

// getWalletAddress returns the wallet address from the wallet
func (e *EventsListener) getWalletAddress() (common.Address, error) {
	prkey, err := e.wallet.GetPrivateKey()
	if err != nil {
		return common.Address{}, err
	}
	return lib.PrivKeyStringToAddr(prkey.Hex())
}

// filter returns true if the event is for the user or provider
func (e *EventsListener) filter(provider, user common.Address) bool {
	ret := provider.Hex() == e.addr.Hex() || user.Hex() == e.addr.Hex()
	if !ret {
		e.log.Debugf("received event for another user/provider, skipping")
	}
	return ret
}
