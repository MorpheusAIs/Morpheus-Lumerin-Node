package mobile

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/docs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/handlers/httphandlers"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/walletapi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
)

// StartHTTPServer starts the native proxy-router HTTP API + Swagger UI.
// address is "host:port", e.g. "127.0.0.1:8082" or "0.0.0.0:8082".
// publicURL sets the Swagger "Try it out" host for CORS (e.g. "http://192.168.1.42:8082").
// If empty, defaults to "http://<address>".
// adminUser and adminPass enable HTTP Basic Auth on protected routes. If both are
// non-empty, auth is enforced; if both are empty, all routes are open (legacy behavior).
// Call StopHTTPServer to shut it down.
func (s *SDK) StartHTTPServer(address, publicURL, adminUser, adminPass string) error {
	s.httpSrvMu.Lock()
	defer s.httpSrvMu.Unlock()

	if s.httpSrvCancel != nil {
		return fmt.Errorf("HTTP server already running")
	}

	// Set swagger host for CORS / "Try it out" — same as WEB_PUBLIC_URL in the daemon.
	if publicURL != "" {
		if u, err := url.Parse(publicURL); err == nil && u.Host != "" {
			docs.SwaggerInfo.Host = u.Host
		} else {
			docs.SwaggerInfo.Host = publicURL
		}
	} else {
		docs.SwaggerInfo.Host = address
	}

	authCfg := &system.HTTPAuthConfig{
		AuthEntries:      make(map[string]*system.HTTPAuthEntry),
		Whitelists:       make(map[string][]string),
		WhitelistDefault: true,
	}

	if adminUser != "" && adminPass != "" {
		saltBytes := make([]byte, 16)
		if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
			return fmt.Errorf("generate auth salt: %w", err)
		}
		saltHex := hex.EncodeToString(saltBytes)
		h := hmac.New(sha256.New, saltBytes)
		h.Write([]byte(adminPass))
		hashHex := hex.EncodeToString(h.Sum(nil))

		authCfg.AuthEntries[adminUser] = &system.HTTPAuthEntry{
			Username: adminUser,
			Salt:     saltHex,
			Hash:     hashHex,
		}
		authCfg.Whitelists[adminUser] = []string{"*"}
		authCfg.WhitelistDefault = false
		s.log.Infof("Expert API auth enabled for user %q", adminUser)
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

	// Healthcheck-only wrapper for the system controller (skip /config, /files,
	// /config/ethNode which crash on nil ethRPC/sysConfig in embedded mode).
	sysCtrl := system.NewSystemController(
		&config.Config{},
		s.wallet,
		nil,
		nil,
		time.Now(),
		big.NewInt(s.cfg.ChainID),
		s.log.Named("SYSTEM-API"),
		nil,
		*authCfg,
		&noopStorageHealthChecker{},
	)

	// Selective registration: only expose routes whose dependencies are
	// satisfied in embedded mode. Skips: system (except /healthcheck),
	// IPFS, /v1/chats/*, and auth agent management — all would crash on
	// nil ethRPC, sysConfig, ipfsManager, chatStorage, or AuthStorage.
	safeProxy := &embeddedProxyRoutes{ctrl: proxyCtrl, auth: *authCfg}
	healthOnly := &embeddedHealthRoute{ctrl: sysCtrl}
	ginEngine := httphandlers.CreateHTTPServer(s.log.Named("HTTP"), *authCfg,
		blockchainCtrl, walletCtrl, safeProxy, healthOnly,
	)

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

// --- Selective route wrappers for embedded mode ---

// embeddedProxyRoutes registers only the proxy routes that don't require
// ipfsManager, chatStorage, or other nil dependencies.
type embeddedProxyRoutes struct {
	ctrl *proxyapi.ProxyController
	auth system.HTTPAuthConfig
}

func (e *embeddedProxyRoutes) RegisterRoutes(r interfaces.Router) {
	r.POST("/proxy/provider/ping", e.ctrl.Ping)
	r.POST("/proxy/sessions/initiate", e.auth.CheckAuth("initiate_session"), e.ctrl.InitiateSession)
	r.POST("/v1/chat/completions", e.auth.CheckAuth("chat"), e.ctrl.Prompt)
	r.GET("/v1/models", e.auth.CheckAuth("get_local_models"), e.ctrl.Models)
	r.GET("/v1/agents", e.auth.CheckAuth("get_local_agents"), e.ctrl.Agents)
	r.GET("/v1/agents/tools", e.auth.CheckAuth("get_agent_tools"), e.ctrl.GetAgentTools)
	r.POST("/v1/agents/tools", e.auth.CheckAuth("call_agent_tool"), e.ctrl.CallAgentTool)
	r.POST("/v1/audio/transcriptions", e.auth.CheckAuth("audio_transcription"), e.ctrl.AudioTranscription)
	r.POST("/v1/audio/speech", e.auth.CheckAuth("audio_speech"), e.ctrl.AudioSpeech)
	r.POST("/v1/embeddings", e.auth.CheckAuth("embeddings"), e.ctrl.Embeddings)
}

// embeddedHealthRoute registers only /healthcheck from the system controller.
type embeddedHealthRoute struct {
	ctrl *system.SystemController
}

func (e *embeddedHealthRoute) RegisterRoutes(r interfaces.Router) {
	r.GET("/healthcheck", e.ctrl.HealthCheck)
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
