package proxyctl

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/handlers/tcphandlers"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/transport"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
	"golang.org/x/sync/errgroup"
)

type ProxyState int8

const (
	StateStopped        ProxyState = 0
	StateWaitingForPkey            = 1
	StateRunning                   = 2
	StateStopping                  = 3
)

const (
	MORDecimals = 18
	ETHDecimals = 18
)

var (
	MORBalanceThreshold = *lib.Exp10(MORDecimals)     // 1 MOR the balance to show a warning
	ETHBalanceThreshold = *lib.Exp10(ETHDecimals - 1) // 0.1 ETH the balance to show a warning
)

func (s ProxyState) String() string {
	switch s {
	case StateStopped:
		return "stopped"
	case StateWaitingForPkey:
		return "waiting for private key"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	}
	return "unknown"
}

type SchedulerLogFactory = func(remoteAddr string) (lib.ILogger, error)

// Proxy is a struct that represents a proxy-router part of the system
type Proxy struct {
	eventListener        *blockchainapi.EventsListener
	wallet               interfaces.PrKeyProvider
	proxyAddr            string
	chainID              *big.Int
	sessionStorage       *storages.SessionStorage
	sessionRepo          *sessionrepo.SessionRepositoryCached
	log                  *lib.Logger
	connLog              *lib.Logger
	schedulerLogFactory  SchedulerLogFactory
	aiEngine             *aiengine.AiEngine
	validator            *validator.Validate
	modelConfigLoader    *config.ModelConfigLoader
	blockchainService    *blockchainapi.BlockchainService
	sessionExpiryHandler *blockchainapi.SessionExpiryHandler

	state         lib.AtomicValue[ProxyState]
	tsk           *lib.Task
	serverStarted <-chan struct{}
}

// NewProxyCtl creates a new Proxy controller instance
func NewProxyCtl(eventListerer *blockchainapi.EventsListener, wallet interfaces.PrKeyProvider, chainID *big.Int, log *lib.Logger, connLog *lib.Logger, proxyAddr string, scl SchedulerLogFactory, sessionStorage *storages.SessionStorage, modelConfigLoader *config.ModelConfigLoader, valid *validator.Validate, aiEngine *aiengine.AiEngine, blockchainService *blockchainapi.BlockchainService, sessionRepo *sessionrepo.SessionRepositoryCached, sessionExpiryHandler *blockchainapi.SessionExpiryHandler) *Proxy {
	return &Proxy{
		eventListener:        eventListerer,
		chainID:              chainID,
		wallet:               wallet,
		log:                  log,
		connLog:              connLog,
		proxyAddr:            proxyAddr,
		schedulerLogFactory:  scl,
		sessionStorage:       sessionStorage,
		aiEngine:             aiEngine,
		validator:            valid,
		modelConfigLoader:    modelConfigLoader,
		blockchainService:    blockchainService,
		sessionRepo:          sessionRepo,
		sessionExpiryHandler: sessionExpiryHandler,
		serverStarted:        make(chan struct{}),
	}
}

func (p *Proxy) Run(ctx context.Context) error {
	var tsk *lib.Task

	for { // restart proxy loop

		prKey, err := p.wallet.GetPrivateKey()
		if err != nil {
			p.log.Errorf("cannot get private key, waiting for its update: %s", err)
			p.setState(StateWaitingForPkey)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-p.wallet.PrivateKeyUpdated():
				continue // restart the proxy
			}
		}

		tsk = lib.NewTaskFunc(func(ctx context.Context) error {
			return p.run(ctx, prKey)
		})

		tsk.Start(ctx)
		p.setState(StateRunning)

		select {
		case <-p.wallet.PrivateKeyUpdated():
			ch := tsk.Stop()
			p.setState(StateStopping)
			<-ch
			continue // restart the proxy
		case <-tsk.Done():
			err := tsk.Err()
			if err != nil {
				var logFunc func(string, ...interface{})
				if errors.Is(err, context.Canceled) {
					logFunc = p.log.Warnf
				} else {
					logFunc = p.log.Errorf
				}
				logFunc("proxy stopped with error: %s", err)
				return err
			}
		}
	}
}

func (p *Proxy) run(ctx context.Context, prKey lib.HexString) error {
	tcpServer := transport.NewTCPServer(p.proxyAddr, p.connLog.Named("TCP"))
	prKey, err := p.wallet.GetPrivateKey()
	if err != nil {
		return err
	}

	walletAddr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		return err
	}
	p.log.Infof("Wallet address: %s", walletAddr.String())

	ethBalance, morBalance, err := p.blockchainService.GetBalance(ctx)
	if err != nil {
		return err
	}

	if ethBalance.Cmp(&ETHBalanceThreshold) < 0 {
		p.log.Warnf(
			"ETH balance is too low: %s (< %s)",
			formatETH(ethBalance),
			formatETH(&ETHBalanceThreshold),
		)
	} else {
		p.log.Infof("ETH balance: %s", formatETH(ethBalance))
	}
	if morBalance.Cmp(&MORBalanceThreshold) < 0 {
		p.log.Warnf(
			"MOR balance is too low: %s (< %s)",
			formatMOR(morBalance),
			formatMOR(&MORBalanceThreshold),
		)
	} else {
		p.log.Infof("MOR balance: %s", formatMOR(morBalance))
	}

	pubKey, err := lib.PubKeyFromPrivate(prKey)
	if err != nil {
		return err
	}

	proxyReceiver := proxyapi.NewProxyReceiver(prKey, pubKey, p.sessionStorage, p.aiEngine, p.chainID, p.modelConfigLoader, p.blockchainService, p.sessionRepo)
	morTcpHandler := proxyapi.NewMORRPCController(proxyReceiver, p.validator, p.sessionRepo, p.sessionStorage, prKey)
	tcpHandler := tcphandlers.NewTCPHandler(
		p.log, p.connLog, p.schedulerLogFactory, morTcpHandler,
	)
	tcpServer.SetConnectionHandler(tcpHandler)

	g, errCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return tcpServer.Run(errCtx)
	})

	g.Go(func() error {
		<-tcpServer.Started()
		return p.afterStart(errCtx, walletAddr)
	})

	g.Go(func() error {
		return p.eventListener.Run(errCtx)
	})

	g.Go(func() error {
		return p.sessionExpiryHandler.Run(errCtx)
	})

	return g.Wait()
}

func (p *Proxy) afterStart(ctx context.Context, walletAddr common.Address) error {
	log := p.log.Named("PROVIDER_CHECK")

	// check if provider exists
	pr, err := p.blockchainService.GetProvider(ctx, walletAddr)
	if err != nil {
		return fmt.Errorf("cannot get provider %s: %w", walletAddr, err)
	} else if pr == nil {
		log.Warnf("provider is not registered under this wallet address: %s", walletAddr)
	} else {
		log.Infof("provider is registered in blockchain, url: %s", pr.Endpoint)
	}

	if pr != nil {
		// check provider connectivity from localhost
		pingDuration, err := p.blockchainService.CheckConnectivity(ctx, pr.Endpoint, pr.Address)
		if err != nil {
			log.Warnf("provider %s is not reachable from localhost by address %s: %s", pr.Address, pr.Endpoint, err)
		} else {
			log.Infof("provider is reachable from localhost, ping: %s", pingDuration)
		}

		// check connectivity from outer network
		ok, err := p.blockchainService.CheckPortOpen(ctx, pr.Endpoint)
		if err != nil {
			log.Warnf("cannot check if port open for %s %s %s", pr.Address, pr.Endpoint, err)
		} else if !ok {
			log.Warnf("provider is not reachable from internet by %s", pr.Endpoint)
		} else {
			log.Infof("provider is reachable from internet")
		}

		// check if provider has active bids
		bids, err := p.blockchainService.GetActiveBidsByProvider(ctx, walletAddr, big.NewInt(0), 100, registries.OrderDESC)
		if err != nil {
			log.Warnf("cannot get active bids by provider: %s", walletAddr, err)
		} else if len(bids) == 0 {
			log.Warnf("provider has no bids available")
		} else {
			log.Infof("provider has %d bids available", len(bids))
		}

		bidsByModelID := make(map[common.Hash][]*structs.Bid)
		for _, bid := range bids {
			bidsByModelID[bid.ModelAgentId] = append(bidsByModelID[bid.ModelAgentId], bid)
		}

		// check if provider has active bids for each model
		modelIDs, _ := p.modelConfigLoader.GetAll()

		for _, modelID := range modelIDs {
			count := len(bidsByModelID[modelID])
			if count == 0 {
				log.Warnf("model %s, no active bids", lib.Short(modelID))
			} else {
				log.Infof("model %s, active bids %d", lib.Short(modelID), count)
			}
		}
	}

	return nil
}

func (p *Proxy) GetState() ProxyState {
	return p.state.Load()
}

func (p *Proxy) setState(s ProxyState) {
	p.state.Store(s)
	p.log.Infof("proxy state: %s", s)
}

func (p *Proxy) ServerStarted() <-chan struct{} {
	return p.serverStarted
}

func formatDecimal(n *big.Int, decimals int) string {
	return lib.NewRat(n, lib.Exp10(decimals)).FloatString(3)
}

func formatMOR(n *big.Int) string {
	return formatDecimal(n, MORDecimals) + " MOR"
}

func formatETH(n *big.Int) string {
	return formatDecimal(n, ETHDecimals) + " ETH"
}
