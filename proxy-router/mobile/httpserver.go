package mobile

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apibus"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/authapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/handlers/httphandlers"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/walletapi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
)

// StartHTTPServer starts the native proxy-router HTTP API + Swagger UI.
// address is "host:port", e.g. "127.0.0.1:8082" or "0.0.0.0:8082".
// Call StopHTTPServer to shut it down.
func (s *SDK) StartHTTPServer(address string) error {
	s.httpSrvMu.Lock()
	defer s.httpSrvMu.Unlock()

	if s.httpSrvCancel != nil {
		return fmt.Errorf("HTTP server already running")
	}

	authCfg := &system.HTTPAuthConfig{
		WhitelistDefault: true,
	}

	blockchainCtrl := blockchainapi.NewBlockchainController(s.blockchain, *authCfg, s.log.Named("BLOCKCHAIN-API"))

	walletCtrl := walletapi.NewWalletController(s.wallet, *authCfg)

	proxyCtrl := proxyapi.NewProxyController(
		s.proxySender,
		&noopAIEngine{},
		s.chatStorage,
		true,  // storeChatContext
		false, // forwardChatContext
		*authCfg,
		nil, // ipfsManager — not available in embedded mode
		s.log.Named("PROXY-API"),
	)

	sysCtrl := system.NewSystemController(
		&config.Config{},
		s.wallet,
		nil,              // ethRPC — not needed for embedded mode health checks
		nil,              // sysConfig
		time.Now(),       // appStartTime
		big.NewInt(s.cfg.ChainID),
		s.log.Named("SYSTEM-API"),
		nil,              // ethConnectionValidator
		*authCfg,
		&noopStorageHealthChecker{},
	)

	authCtrl := authapi.NewAuthController(authCfg, "mobile", s.log.Named("AUTH-API"))

	bus := apibus.NewApiBus(blockchainCtrl, proxyCtrl, walletCtrl, sysCtrl, authCtrl)
	ginEngine := httphandlers.CreateHTTPServer(s.log.Named("HTTP"), *authCfg, bus)

	srv := &http.Server{Addr: address, Handler: ginEngine}

	ctx, cancel := context.WithCancel(context.Background())
	s.httpSrvCancel = cancel

	errCh := make(chan error, 1)
	go func() {
		s.log.Infof("Expert API (swagger) listening on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Errorf("Expert API error: %v", err)
			errCh <- err
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			shutCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
			defer c()
			srv.Shutdown(shutCtx)
			s.log.Info("Expert API stopped")
		case err := <-errCh:
			s.httpSrvMu.Lock()
			s.httpSrvCancel = nil
			s.httpSrvMu.Unlock()
			_ = err
		}
	}()

	s.httpSrvAddr = address
	return nil
}

// StopHTTPServer shuts down the HTTP API server.
func (s *SDK) StopHTTPServer() {
	s.httpSrvMu.Lock()
	defer s.httpSrvMu.Unlock()
	if s.httpSrvCancel != nil {
		s.httpSrvCancel()
		s.httpSrvCancel = nil
		s.httpSrvAddr = ""
	}
}

// HTTPServerAddr returns the address the HTTP server is listening on, or "" if not running.
func (s *SDK) HTTPServerAddr() string {
	s.httpSrvMu.Lock()
	defer s.httpSrvMu.Unlock()
	return s.httpSrvAddr
}

// --- No-op implementations for dependencies not available in embedded mode ---

type noopAIEngine struct{}

func (n *noopAIEngine) GetLocalModels() ([]aiengine.LocalModel, error) {
	return nil, nil
}

func (n *noopAIEngine) GetLocalAgents() ([]aiengine.LocalAgent, error) {
	return nil, nil
}

func (n *noopAIEngine) CallAgentTool(ctx context.Context, sessionID, agentID common.Hash, toolName string, input map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("agent tools not available in embedded mode")
}

func (n *noopAIEngine) GetAgentTools(ctx context.Context, sessionID, agentID common.Hash) ([]aiengine.AgentTool, error) {
	return nil, nil
}

func (n *noopAIEngine) GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeContext, forwardContext bool) (aiengine.AIEngineStream, error) {
	return nil, fmt.Errorf("local AI engine not available in embedded mode")
}

type noopStorageHealthChecker struct{}

func (n *noopStorageHealthChecker) HealthCheck() error        { return nil }
func (n *noopStorageHealthChecker) DBSize() (int64, int64)    { return 0, 0 }
