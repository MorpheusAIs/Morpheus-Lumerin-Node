package sessionrepo

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type SessionRepositoryCached struct {
	storage *storages.SessionStorage
	reg     *registries.SessionRouter
	mkt     *registries.Marketplace
	log     lib.ILogger
}

func NewSessionRepositoryCached(storage *storages.SessionStorage, reg *registries.SessionRouter, mkt *registries.Marketplace, log lib.ILogger) *SessionRepositoryCached {
	return &SessionRepositoryCached{
		storage: storage,
		reg:     reg,
		mkt:     mkt,
		log:     log,
	}
}

// GetSession returns a session by its ID from the read-through cache
func (r *SessionRepositoryCached) GetSession(ctx context.Context, id common.Hash) (*SessionModel, error) {
	ses, ok := r.getSessionFromCache(id)
	if ok {
		r.log.Debugf("Session found in cache id: %s, endsAt: %v", ses.id.Hex(), ses.endsAt)
		return ses, nil
	}

	session, err := r.getSessionFromBlockchain(ctx, id)
	if err != nil {
		r.log.Debugf("Error getting session from blockchain: %v", err)
		return nil, err
	}
	r.log.Debugf("Session found in blockchain id: %s, endsAt: %v", session.id.Hex(), session.endsAt)

	err = r.saveSessionToCache(session)
	if err != nil {
		r.log.Debugf("Error saving session to cache: %v", err)
		return nil, err
	}

	r.log.Debugf("Session saved to cache id: %s, endsAt: %v", session.id.Hex(), session.endsAt)
	return session, nil
}

// SaveSession saves a session to the cache. Before saving it to cache you have to call GetSession
func (r *SessionRepositoryCached) SaveSession(ctx context.Context, ses *SessionModel) error {
	return r.saveSessionToCache(ses)
}

// RemoveSession removes a session from the cache
func (r *SessionRepositoryCached) RemoveSession(ctx context.Context, id common.Hash) error {
	return r.storage.RemoveSession(id.Hex())
}

// RefreshSession refreshes the session cache by fetching the session from the blockchain
func (r *SessionRepositoryCached) RefreshSession(ctx context.Context, id common.Hash) error {
	// since the session record in blockchain is immutable if we have it in cache it is noop
	// if we don't have it in cache we need to fetch it and save it to cache
	_, err := r.GetSession(ctx, id)
	return err
}

func (r *SessionRepositoryCached) getSessionFromBlockchain(ctx context.Context, id common.Hash) (*SessionModel, error) {
	session, err := r.reg.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}

	bid, err := r.mkt.GetBidById(ctx, session.BidId)
	if err != nil {
		return nil, err
	}

	return &SessionModel{
		id:               id,
		userAddr:         session.User,
		providerAddr:     bid.Provider,
		endsAt:           session.EndsAt,
		modelID:          bid.ModelId,
		tpsScaled1000Arr: []int{},
		ttftMsArr:        []int{},
		failoverEnabled:  false,
		directPayment:    session.IsDirectPaymentFromUser,
		agentUsername:    "admin",
	}, nil
}

func (r *SessionRepositoryCached) getSessionFromCache(id common.Hash) (*SessionModel, bool) {
	ses, ok := r.storage.GetSession(id.Hex())
	if !ok {
		r.log.Debugf("Session not found in cache: %v", id)
		return nil, false
	}
	r.log.Debugf("getSessionFromCache: Session found in cache id: %s, endsAt: %v", ses.Id, ses.EndsAt)
	return &SessionModel{
		id:               common.HexToHash(ses.Id),
		userAddr:         common.HexToAddress(ses.UserAddr),
		providerAddr:     common.HexToAddress(ses.ProviderAddr),
		modelID:          common.HexToHash(ses.ModelID),
		endsAt:           ses.EndsAt,
		tpsScaled1000Arr: ses.TPSScaled1000Arr,
		ttftMsArr:        ses.TTFTMsArr,
		failoverEnabled:  ses.FailoverEnabled,
		agentUsername:    ses.AgentUsername,
	}, true
}

func (r *SessionRepositoryCached) saveSessionToCache(ses *SessionModel) error {
	return r.storage.AddSession(&storages.Session{
		Id:               ses.id.Hex(),
		UserAddr:         ses.userAddr.Hex(),
		ProviderAddr:     ses.providerAddr.Hex(),
		EndsAt:           ses.endsAt,
		ModelID:          ses.modelID.Hex(),
		TPSScaled1000Arr: ses.tpsScaled1000Arr,
		TTFTMsArr:        ses.ttftMsArr,
		FailoverEnabled:  ses.failoverEnabled,
		DirectPayment:    ses.directPayment,
		AgentUsername:    ses.agentUsername,
	})
}
