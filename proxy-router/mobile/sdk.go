package mobile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/attestation"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/ethclient"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	wallet "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/wallet"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
)

const (
	DefaultDerivationPath       = "m/44'/60'/0'/0/0"
	DefaultCNodePNodeTimeout    = 120 * time.Second
	DefaultCNodePNodeMaxRetries = 3
	DefaultCNodeAudioMaxRetries = 1
	DefaultModelCacheTTL        = 5 * time.Minute

	chainIDVerifyTimeout    = 15 * time.Second
	httpClientTimeout       = 10 * time.Second
	maintenanceStartupGrace = 2 * time.Minute

	paginationLimit    uint8 = 50
	paginationMaxPages int   = 40
)

var ErrSDKClosed = errors.New("SDK has been shut down")

// Config holds the minimal configuration needed for the mobile SDK.
type Config struct {
	DataDir         string // persistent storage root (chat history, etc.)
	EthNodeURL      string // Ethereum JSON-RPC endpoint(s): comma/semicolon/|/newline-separated for failover
	ChainID         int64  // e.g. 8453 (Base), 84532 (Base Sepolia)
	DiamondAddr     string // diamond proxy contract (hex with 0x prefix)
	MorTokenAddr    string // MOR token contract (hex with 0x prefix)
	BlockscoutURL   string // Blockscout API v2 base URL
	ActiveModelsURL string // pre-built active models JSON (production: https://active.mor.org/active_models.json)
	LogLevel        string // "debug", "info", "warn", "error" (default: "info")
	// LogWriter, when non-nil, receives a copy of every SDK log line (tee'd alongside stdout).
	LogWriter io.Writer
	// TEEPortalURL is the SecretAI (or compatible) quote-parse API. Empty uses the default production URL.
	TEEPortalURL string
	// TEEImageRepo is the GHCR image used to load cosign golden TEE measurements. Empty uses the Morpheus default repo.
	TEEImageRepo string
}

// SDK is the main entry point for mobile applications.
// It wraps the proxy-router's internal packages behind a clean public API.
type SDK struct {
	cfg            Config
	log            lib.ILogger
	baseLog        *lib.Logger
	ethClient      *ethclient.Client
	wallet         *wallet.KeychainWallet
	walletStorage  *MemoryKeyValueStorage
	blockchain     *blockchainapi.BlockchainService
	proxySender    *proxyapi.ProxyServiceSender
	sessionRepo    *sessionrepo.SessionRepositoryCached
	chatStorage    gcs.ChatStorageInterface
	storage        *storages.Storage
	sessionStorage *storages.SessionStorage

	// Active models cache (HTTP-based, from Marketplace API)
	modelsMu      sync.RWMutex
	modelsCache   []Model
	modelsByID    map[string]*Model
	modelsByName  map[string]*Model
	modelsCacheAt time.Time
	modelsHash    string
	httpClient    *http.Client

	// Expert mode HTTP server (native proxy-router swagger API)
	httpSrvMu     sync.Mutex
	httpSrvCancel context.CancelFunc
	httpSrvAddr   string

	// Session maintenance (auto-close expired, detect provider closures)
	maintMu     sync.Mutex
	maintCancel context.CancelFunc
	maintWg     sync.WaitGroup

	// Lifecycle guard: prevents use-after-Shutdown panics
	closed atomic.Int32
}

// NewSDK creates and initializes the SDK. Call Shutdown() when done.
func NewSDK(cfg Config) (*SDK, error) {
	if cfg.EthNodeURL == "" {
		return nil, fmt.Errorf("EthNodeURL is required")
	}
	if cfg.DiamondAddr == "" {
		return nil, fmt.Errorf("DiamondAddr is required")
	}
	if cfg.MorTokenAddr == "" {
		return nil, fmt.Errorf("MorTokenAddr is required")
	}
	if cfg.BlockscoutURL == "" {
		return nil, fmt.Errorf("BlockscoutURL is required")
	}
	if !common.IsHexAddress(cfg.DiamondAddr) {
		return nil, fmt.Errorf("DiamondAddr is not a valid hex address")
	}
	if !common.IsHexAddress(cfg.MorTokenAddr) {
		return nil, fmt.Errorf("MorTokenAddr is not a valid hex address")
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	var (
		baseLog *lib.Logger
		err     error
	)
	if cfg.LogWriter != nil {
		baseLog, err = lib.NewLoggerMemory(cfg.LogLevel, false, true, false, "", cfg.LogWriter)
	} else {
		baseLog, err = lib.NewLogger(cfg.LogLevel, false, true, false, "")
	}
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	log := baseLog.Named("MOBILE")

	urls := parseEthNodeURLs(cfg.EthNodeURL)
	if len(urls) == 0 {
		return nil, fmt.Errorf("EthNodeURL is required")
	}

	var rpcBackend ethclient.RPCClient
	if len(urls) == 1 {
		rc, derr := ethclient.DialRPC(urls[0])
		if derr != nil {
			return nil, fmt.Errorf("dial eth node: %w", derr)
		}
		rpcBackend = rc
		log.Infof("single RPC endpoint: %s", urls[0])
	} else {
		multi, derr := ethclient.NewRPCClientMultiple(urls, log)
		if derr != nil {
			return nil, fmt.Errorf("dial eth nodes: %w", derr)
		}
		rpcBackend = multi
		log.Infof("using %d RPC endpoints with round-robin + backoff failover", len(multi.GetURLs()))
	}

	ethClient := ethclient.NewClient(rpcBackend)

	ctx, cancel := context.WithTimeout(context.Background(), chainIDVerifyTimeout)
	defer cancel()

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("get chain ID: %w", err)
	}
	if cfg.ChainID != 0 && chainID.Int64() != cfg.ChainID {
		ethClient.Close()
		return nil, fmt.Errorf("chain ID mismatch: config=%d, node=%d", cfg.ChainID, chainID.Int64())
	}
	log.Infof("connected to eth node, chainID=%d", chainID.Int64())

	diamondAddr := common.HexToAddress(cfg.DiamondAddr)
	morTokenAddr := common.HexToAddress(cfg.MorTokenAddr)

	inMemStorage := storages.NewTestStorage()
	sessionStorage := storages.NewSessionStorage(inMemStorage)

	walletKV := NewMemoryKeyValueStorage()
	w := wallet.NewKeychainWallet(walletKV)

	mc := multicall.NewMulticall3(ethClient)
	rpcLog := log.Named("RPC")
	sessionRouter := registries.NewSessionRouter(diamondAddr, ethClient, mc, rpcLog)
	marketplace := registries.NewMarketplace(diamondAddr, ethClient, mc, rpcLog)
	sessionRepo := sessionrepo.NewSessionRepositoryCached(sessionStorage, sessionRouter, marketplace, log)

	contractLogStorage := lib.NewCollection[*interfaces.LogStorage]()
	proxySender := proxyapi.NewProxySender(
		chainID, w, contractLogStorage, sessionStorage, sessionRepo,
		DefaultCNodePNodeTimeout, DefaultCNodePNodeMaxRetries, DefaultCNodeAudioMaxRetries, log,
	)

	explorer := blockchainapi.NewBlockscoutApiV2Client(cfg.BlockscoutURL, log.Named("INDEXER"))

	scorer, err := rating.NewRatingFromConfig([]byte(config.RatingConfigDefault), log.Named("RATING"))
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("create rating: %w", err)
	}

	teeVerifier := attestation.NewVerifier(cfg.TEEPortalURL, cfg.TEEImageRepo, log.Named("TEE"))

	blockchainSvc := blockchainapi.NewBlockchainService(
		ethClient, mc, diamondAddr, morTokenAddr,
		explorer, w, proxySender, sessionRepo, scorer,
		nil,
		log, rpcLog,
		false,
		teeVerifier,
	)

	proxySender.SetSessionService(blockchainSvc)
	proxySender.SetAttestationVerifier(teeVerifier)
	teeVerifier.SetPingFunc(func(ctx context.Context, providerEndpoint string, providerAddr string) (string, error) {
		_, version, err := proxySender.Ping(ctx, providerEndpoint, common.HexToAddress(providerAddr))
		return version, err
	})

	var cs gcs.ChatStorageInterface
	if cfg.DataDir != "" {
		chatDir := filepath.Join(cfg.DataDir, "chats")
		cs = chatstorage.NewChatStorage(chatDir)
	}

	sdk := &SDK{
		cfg:            cfg,
		log:            log,
		baseLog:        baseLog,
		ethClient:      ethClient,
		wallet:         w,
		walletStorage:  walletKV,
		blockchain:     blockchainSvc,
		proxySender:    proxySender,
		sessionRepo:    sessionRepo,
		chatStorage:    cs,
		storage:        inMemStorage,
		sessionStorage: sessionStorage,
		modelsByID:     make(map[string]*Model),
		modelsByName:   make(map[string]*Model),
		httpClient:     &http.Client{Timeout: httpClientTimeout},
	}

	log.Info("SDK initialized")
	sdk.startSessionMaintenance()

	return sdk, nil
}

// SetLogLevel changes the SDK's internal log level at runtime.
func (s *SDK) SetLogLevel(level string) error {
	return s.baseLog.SetLevel(level)
}

// GetLogLevel returns the SDK's current log level string.
func (s *SDK) GetLogLevel() string {
	return s.baseLog.GetLevel()
}

// ProxyRouterVersion returns the proxy-router build version baked in at compile time.
func ProxyRouterVersion() string {
	return config.BuildVersion
}

// ProxyRouterCommit returns the proxy-router git commit baked in at compile time.
func ProxyRouterCommit() string {
	return config.Commit
}

// Shutdown releases all resources held by the SDK.
// After Shutdown returns, all public methods return ErrSDKClosed.
func (s *SDK) Shutdown() {
	s.closed.Store(1)

	s.maintMu.Lock()
	if s.maintCancel != nil {
		s.maintCancel()
		s.maintCancel = nil
	}
	s.maintMu.Unlock()
	s.maintWg.Wait()

	s.StopHTTPServer()

	if s.ethClient != nil {
		s.ethClient.Close()
		s.ethClient = nil
	}

	s.walletStorage.Clear()

	s.log.Info("SDK shut down")
}

func (s *SDK) checkClosed() error {
	if s.closed.Load() != 0 {
		return ErrSDKClosed
	}
	return nil
}
