package mobile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
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
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	openai "github.com/sashabaranov/go-openai"
	"github.com/tyler-smith/go-bip39"
)

const (
	DefaultDerivationPath        = "m/44'/60'/0'/0/0"
	DefaultCNodePNodeTimeout     = 120 * time.Second
	DefaultCNodePNodeMaxRetries  = 3
	DefaultCNodeAudioMaxRetries  = 1
	DefaultModelCacheTTL         = 5 * time.Minute
)

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
	// TEEPortalURL is the SecretAI (or compatible) quote-parse API. Empty uses the default production URL.
	TEEPortalURL string
	// TEEImageRepo is the GHCR image used to load cosign golden TEE measurements (e.g. ghcr.io/.../morpheus-lumerin-node-tee). Empty uses the Morpheus default repo.
	TEEImageRepo string
}

// SDK is the main entry point for mobile applications.
// It wraps the proxy-router's internal packages behind a clean public API.
type SDK struct {
	cfg            Config
	log            lib.ILogger
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
	modelsByID    map[string]*Model // blockchain ID -> model
	modelsByName  map[string]*Model // lowercase name -> model
	modelsCacheAt time.Time
	modelsHash    string
	httpClient    *http.Client

	// Expert mode HTTP server (native proxy-router swagger API)
	httpSrvMu     sync.Mutex
	httpSrvCancel context.CancelFunc
	httpSrvAddr   string
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
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	baseLog, err := lib.NewLogger(cfg.LogLevel, false, true, false, "")
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

	// Verify chain ID
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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

	// In-memory BadgerDB for session tracking (consumer doesn't need persistence)
	inMemStorage := storages.NewTestStorage()
	sessionStorage := storages.NewSessionStorage(inMemStorage)

	// In-memory wallet storage
	walletKV := NewMemoryKeyValueStorage()
	w := wallet.NewKeychainWallet(walletKV)

	// Blockchain infrastructure
	mc := multicall.NewMulticall3(ethClient)
	rpcLog := log.Named("RPC")
	sessionRouter := registries.NewSessionRouter(diamondAddr, ethClient, mc, rpcLog)
	marketplace := registries.NewMarketplace(diamondAddr, ethClient, mc, rpcLog)
	sessionRepo := sessionrepo.NewSessionRepositoryCached(sessionStorage, sessionRouter, marketplace, log)

	// Proxy sender (consumer-side MOR-RPC client)
	contractLogStorage := lib.NewCollection[*interfaces.LogStorage]()
	proxySender := proxyapi.NewProxySender(
		chainID, w, contractLogStorage, sessionStorage, sessionRepo,
		DefaultCNodePNodeTimeout, DefaultCNodePNodeMaxRetries, DefaultCNodeAudioMaxRetries, log,
	)

	// Explorer (Blockscout) for transaction history
	explorer := blockchainapi.NewBlockscoutApiV2Client(cfg.BlockscoutURL, log.Named("INDEXER"))

	// Rating with default config
	scorer, err := rating.NewRatingFromConfig([]byte(config.RatingConfigDefault), log.Named("RATING"))
	if err != nil {
		ethClient.Close()
		return nil, fmt.Errorf("create rating: %w", err)
	}

	// TEE attestation (register match vs cosign golden manifest) — required for Secure models; same path as proxy-router daemon.
	teeVerifier := attestation.NewVerifier(cfg.TEEPortalURL, cfg.TEEImageRepo, log.Named("TEE"))

	// Blockchain service — the main orchestrator for on-chain operations
	blockchainSvc := blockchainapi.NewBlockchainService(
		ethClient, mc, diamondAddr, morTokenAddr,
		explorer, w, proxySender, sessionRepo, scorer,
		nil,   // authConfig — not needed for direct SDK calls
		log, rpcLog,
		false, // legacyTx
		teeVerifier,
	)

	// Wire the circular dependency: proxy sender needs session service for failover
	proxySender.SetSessionService(blockchainSvc)

	// Chat storage (file-based JSON, portable)
	var cs gcs.ChatStorageInterface
	if cfg.DataDir != "" {
		chatDir := filepath.Join(cfg.DataDir, "chats")
		cs = chatstorage.NewChatStorage(chatDir)
	}

	sdk := &SDK{
		cfg:            cfg,
		log:            log,
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
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}

	log.Info("SDK initialized")
	return sdk, nil
}

// Shutdown releases all resources held by the SDK.
func (s *SDK) Shutdown() {
	s.StopHTTPServer()
	if s.ethClient != nil {
		s.ethClient.Close()
		s.ethClient = nil
	}
	s.log.Info("SDK shut down")
}

// parseEthNodeURLs splits a config string into deduplicated RPC URLs (order preserved).
func parseEthNodeURLs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\t'
	})
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u == "" {
			continue
		}
		if _, ok := seen[u]; ok {
			continue
		}
		seen[u] = struct{}{}
		out = append(out, u)
	}
	return out
}

// --- Wallet Operations ---

// CreateWallet generates a new BIP-39 mnemonic and derives the wallet.
// Returns the mnemonic (back it up!) and the Ethereum address.
func (s *SDK) CreateWallet() (mnemonic string, address string, err error) {
	entropy, err := bip39.NewEntropy(128) // 12 words
	if err != nil {
		return "", "", fmt.Errorf("generate entropy: %w", err)
	}
	mnemonic, err = bip39.NewMnemonic(entropy)
	if err != nil {
		return "", "", fmt.Errorf("generate mnemonic: %w", err)
	}
	err = s.wallet.SetMnemonic(mnemonic, DefaultDerivationPath)
	if err != nil {
		return "", "", fmt.Errorf("set mnemonic: %w", err)
	}
	addr, err := s.getAddress()
	if err != nil {
		return "", "", err
	}
	s.log.Infof("wallet created: %s", addr)
	return mnemonic, addr, nil
}

// ImportMnemonic imports a wallet from an existing BIP-39 mnemonic.
func (s *SDK) ImportMnemonic(mnemonic string) (address string, err error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic")
	}
	err = s.wallet.SetMnemonic(mnemonic, DefaultDerivationPath)
	if err != nil {
		return "", fmt.Errorf("set mnemonic: %w", err)
	}
	return s.getAddress()
}

// ImportPrivateKey imports a wallet from a hex-encoded private key.
func (s *SDK) ImportPrivateKey(hexKey string) (address string, err error) {
	pk, err := lib.StringToHexString(hexKey)
	if err != nil {
		return "", fmt.Errorf("parse private key: %w", err)
	}
	err = s.wallet.SetPrivateKey(pk)
	if err != nil {
		return "", fmt.Errorf("set private key: %w", err)
	}
	return s.getAddress()
}

// VerifyMnemonicMatchesCurrent returns true if the mnemonic derives the same address as the loaded wallet (no mutation).
func (s *SDK) VerifyMnemonicMatchesCurrent(mnemonic string) (bool, error) {
	mnemonic = strings.TrimSpace(mnemonic)
	if !bip39.IsMnemonicValid(mnemonic) {
		return false, fmt.Errorf("invalid mnemonic")
	}
	current, err := s.getAddress()
	if err != nil {
		return false, err
	}
	w, err := wallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return false, err
	}
	path, err := accounts.ParseDerivationPath(DefaultDerivationPath)
	if err != nil {
		return false, err
	}
	pk, err := w.DerivePrivateKey(path)
	if err != nil {
		return false, err
	}
	addr, err := lib.PrivKeyToAddr(pk)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(addr.Hex(), current), nil
}

// VerifyPrivateKeyMatchesCurrent returns true if the hex private key matches the loaded wallet (no mutation).
func (s *SDK) VerifyPrivateKeyMatchesCurrent(hexKey string) (bool, error) {
	current, err := s.getAddress()
	if err != nil {
		return false, err
	}
	pk, err := lib.StringToHexString(hexKey)
	if err != nil {
		return false, err
	}
	addr, err := lib.PrivKeyBytesToAddr(pk)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(addr.Hex(), current), nil
}

// ExportPrivateKey returns the private key as a hex string.
func (s *SDK) ExportPrivateKey() (string, error) {
	pk, err := s.wallet.GetPrivateKey()
	if err != nil {
		return "", err
	}
	return pk.Hex(), nil
}

// GetAddress returns the wallet's Ethereum address.
func (s *SDK) GetAddress() (string, error) {
	return s.getAddress()
}

func (s *SDK) getAddress() (string, error) {
	pk, err := s.wallet.GetPrivateKey()
	if err != nil {
		return "", fmt.Errorf("get private key: %w", err)
	}
	addr, err := lib.PrivKeyBytesToAddr(pk)
	if err != nil {
		return "", fmt.Errorf("derive address: %w", err)
	}
	return addr.Hex(), nil
}

// --- Balance ---

// GetBalance returns ETH and MOR balances as decimal strings (in wei).
func (s *SDK) GetBalance(ctx context.Context) (*Balance, error) {
	eth, mor, err := s.blockchain.GetBalance(ctx)
	if err != nil {
		return nil, err
	}
	return &Balance{
		ETH: eth.String(),
		MOR: mor.String(),
	}, nil
}

// GetBalanceJSON returns the balance as a JSON string (for FFI).
func (s *SDK) GetBalanceJSON(ctx context.Context) (string, error) {
	b, err := s.GetBalance(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(b)
}

// SendETH sends native ETH (amount in wei, decimal string) to an 0x address. Waits for mining.
func (s *SDK) SendETH(ctx context.Context, toHex string, amountWei string) (txHash string, err error) {
	if !common.IsHexAddress(toHex) {
		return "", fmt.Errorf("invalid recipient address")
	}
	to := common.HexToAddress(toHex)
	amt, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amt.Sign() <= 0 {
		return "", fmt.Errorf("invalid amount: use wei as a decimal string")
	}
	h, err := s.blockchain.SendETH(ctx, to, amt, "")
	if err != nil {
		return "", err
	}
	return h.Hex(), nil
}

// SendMOR sends MOR ERC-20 (amount in token wei, 18 decimals, decimal string) to an 0x address.
func (s *SDK) SendMOR(ctx context.Context, toHex string, amountWei string) (txHash string, err error) {
	if !common.IsHexAddress(toHex) {
		return "", fmt.Errorf("invalid recipient address")
	}
	to := common.HexToAddress(toHex)
	amt, ok := new(big.Int).SetString(amountWei, 10)
	if !ok || amt.Sign() <= 0 {
		return "", fmt.Errorf("invalid amount: use wei as a decimal string")
	}
	h, err := s.blockchain.SendMOR(ctx, to, amt, "")
	if err != nil {
		return "", err
	}
	return h.Hex(), nil
}

// --- Models ---

// activeModelsResponse matches the JSON shape at ActiveModelsURL.
type activeModelsResponse struct {
	Models []struct {
		ID        string   `json:"Id"`
		IpfsCID   string   `json:"IpfsCID"`
		Fee       int64    `json:"Fee"`
		Stake     int64    `json:"Stake"`
		Owner     string   `json:"Owner"`
		Name      string   `json:"Name"`
		Tags      []string `json:"Tags"`
		CreatedAt int64    `json:"CreatedAt"`
		IsDeleted bool     `json:"IsDeleted"`
		ModelType string   `json:"ModelType"`
	} `json:"models"`
	LastUpdated int64 `json:"last_updated"`
}

// refreshModelsCache fetches models from the HTTP endpoint and updates the cache.
func (s *SDK) refreshModelsCache() error {
	if s.cfg.ActiveModelsURL == "" {
		return fmt.Errorf("ActiveModelsURL not configured")
	}

	resp, err := s.httpClient.Get(s.cfg.ActiveModelsURL)
	if err != nil {
		return fmt.Errorf("fetch active models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("active models returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	hash := sha256.Sum256(body)
	hashStr := hex.EncodeToString(hash[:])

	s.modelsMu.RLock()
	same := hashStr == s.modelsHash
	s.modelsMu.RUnlock()
	if same {
		s.modelsMu.Lock()
		s.modelsCacheAt = time.Now()
		s.modelsMu.Unlock()
		s.log.Debug("active models unchanged, extending cache")
		return nil
	}

	var data activeModelsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("parse active models: %w", err)
	}

	models := make([]Model, 0, len(data.Models))
	byID := make(map[string]*Model, len(data.Models))
	byName := make(map[string]*Model, len(data.Models))

	for _, m := range data.Models {
		if m.IsDeleted {
			continue
		}
		model := Model{
			ID:        m.ID,
			Name:      m.Name,
			Tags:      m.Tags,
			Fee:       fmt.Sprintf("%d", m.Fee),
			Stake:     fmt.Sprintf("%d", m.Stake),
			Owner:     m.Owner,
			ModelType: m.ModelType,
			CreatedAt: m.CreatedAt,
		}
		models = append(models, model)
		stored := models[len(models)-1:]
		byID[m.ID] = &stored[0]
		byName[strings.ToLower(m.Name)] = &stored[0]
	}

	s.modelsMu.Lock()
	s.modelsCache = models
	s.modelsByID = byID
	s.modelsByName = byName
	s.modelsCacheAt = time.Now()
	s.modelsHash = hashStr
	s.modelsMu.Unlock()

	s.log.Infof("cached %d active models from HTTP endpoint", len(models))
	return nil
}

// GetAllModels returns active models, preferring the HTTP endpoint with blockchain fallback.
func (s *SDK) GetAllModels(ctx context.Context) ([]Model, error) {
	// Try HTTP cache first
	s.modelsMu.RLock()
	cached := s.modelsCache
	age := time.Since(s.modelsCacheAt)
	s.modelsMu.RUnlock()

	if cached != nil && age < DefaultModelCacheTTL {
		return cached, nil
	}

	// Refresh from HTTP endpoint
	if s.cfg.ActiveModelsURL != "" {
		if err := s.refreshModelsCache(); err != nil {
			s.log.Warnf("HTTP model fetch failed, trying blockchain: %v", err)
		} else {
			s.modelsMu.RLock()
			result := s.modelsCache
			s.modelsMu.RUnlock()
			return result, nil
		}
	}

	// Fallback: blockchain via Multicall
	models, err := s.blockchain.GetAllModels(ctx)
	if err != nil {
		if cached != nil {
			s.log.Warn("blockchain fetch also failed, returning stale cache")
			return cached, nil
		}
		return nil, err
	}
	out := make([]Model, 0, len(models))
	for _, m := range models {
		if m.IsDeleted {
			continue
		}
		out = append(out, Model{
			ID:    m.Id.Hex(),
			Name:  m.Name,
			Tags:  m.Tags,
			Fee:   bigStr(m.Fee),
			Stake: bigStr(m.Stake),
			Owner: m.Owner.Hex(),
		})
	}
	return out, nil
}

// ResolveModelID resolves a model name or blockchain ID to the canonical blockchain ID.
func (s *SDK) ResolveModelID(ctx context.Context, nameOrID string) (string, error) {
	if _, err := s.GetAllModels(ctx); err != nil {
		return "", err
	}
	s.modelsMu.RLock()
	defer s.modelsMu.RUnlock()

	if m, ok := s.modelsByID[nameOrID]; ok {
		return m.ID, nil
	}
	if m, ok := s.modelsByName[strings.ToLower(nameOrID)]; ok {
		return m.ID, nil
	}
	return "", fmt.Errorf("model not found: %s", nameOrID)
}

// GetAllModelsJSON returns all models as a JSON string (for FFI).
func (s *SDK) GetAllModelsJSON(ctx context.Context) (string, error) {
	models, err := s.GetAllModels(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(models)
}

// GetRatedBids returns bids for a model, scored and sorted by quality.
func (s *SDK) GetRatedBids(ctx context.Context, modelID string) ([]ScoredBid, error) {
	id := common.HexToHash(modelID)
	bids, err := s.blockchain.GetRatedBids(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]ScoredBid, 0, len(bids))
	for _, b := range bids {
		out = append(out, ScoredBid{
			ID:             b.ID.Hex(),
			Provider:       b.Bid.Provider.Hex(),
			ModelAgentID:   b.Bid.ModelAgentId.Hex(),
			PricePerSecond: bigIntStr(b.Bid.PricePerSecond),
			Score:          b.Score,
		})
	}
	return out, nil
}

// GetRatedBidsJSON returns rated bids as a JSON string (for FFI).
func (s *SDK) GetRatedBidsJSON(ctx context.Context, modelID string) (string, error) {
	bids, err := s.GetRatedBids(ctx, modelID)
	if err != nil {
		return "", err
	}
	return toJSON(bids)
}

// EstimateOpenSessionStakeJSON returns stake + formula inputs as JSON (for FFI / UI).
// Uses the top-scored bid (same as the first provider tried when opening a session).
func (s *SDK) EstimateOpenSessionStakeJSON(ctx context.Context, modelID string, durationSec int64, directPayment bool) (string, error) {
	id := common.HexToHash(modelID)
	dur := big.NewInt(durationSec)
	est, err := s.blockchain.EstimateOpenSessionStake(ctx, id, dur, directPayment)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(est)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// --- Sessions ---

// OpenSession opens a session with the best-rated provider for a model.
// duration is in seconds. Returns the session ID (tx hash).
func (s *SDK) OpenSession(ctx context.Context, modelID string, durationSec int64, directPayment bool) (string, error) {
	id := common.HexToHash(modelID)
	dur := big.NewInt(durationSec)
	txHash, err := s.blockchain.OpenSessionByModelId(ctx, id, dur, directPayment, true, common.Address{}, "")
	if err != nil {
		return "", err
	}
	return txHash.Hex(), nil
}

// CloseSession closes an active session. Returns the close tx hash.
func (s *SDK) CloseSession(ctx context.Context, sessionID string) (string, error) {
	id := common.HexToHash(sessionID)
	txHash, err := s.blockchain.CloseSession(ctx, id)
	if err != nil {
		return "", err
	}
	return txHash.Hex(), nil
}

// GetSession retrieves session details by ID.
func (s *SDK) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	id := common.HexToHash(sessionID)
	sess, err := s.blockchain.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	return sessionFromInternal(sess), nil
}

// GetSessionJSON returns session details as a JSON string (for FFI).
func (s *SDK) GetSessionJSON(ctx context.Context, sessionID string) (string, error) {
	sess, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return "", err
	}
	return toJSON(sess)
}

// GetUnclosedUserSessions returns on-chain sessions for the wallet where ClosedAt == 0.
// Paginates newest-first until a short page or maxPages.
func (s *SDK) GetUnclosedUserSessions(ctx context.Context) ([]Session, error) {
	addr, err := s.getAddress()
	if err != nil {
		return nil, err
	}
	user := common.HexToAddress(addr)
	var out []Session
	offset := big.NewInt(0)
	const limit uint8 = 50
	const maxPages = 40
	order := registries.OrderDESC

	for page := 0; page < maxPages; page++ {
		unclosed, _, totalFetched, err := s.blockchain.GetUnclosedUserSessions(ctx, user, offset, limit, order)
		if err != nil {
			return nil, err
		}
		for _, ses := range unclosed {
			if ses == nil {
				continue
			}
			out = append(out, *sessionFromInternal(ses))
		}
		if totalFetched < int(limit) {
			break
		}
		offset = new(big.Int).Add(offset, big.NewInt(int64(totalFetched)))
	}
	// JSON null vs []: encoding/json marshals nil slice as null; FFI/Flutter expect [].
	if out == nil {
		out = []Session{}
	}
	return out, nil
}

// GetUnclosedUserSessionsJSON is for FFI / JSON consumers.
func (s *SDK) GetUnclosedUserSessionsJSON(ctx context.Context) (string, error) {
	list, err := s.GetUnclosedUserSessions(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(list)
}

// --- Chat ---

// StreamCallback receives streaming chunks from a chat completion.
// text is the content delta, isLast is true on the final chunk.
type StreamCallback func(text string, isLast bool) error

// SendPrompt sends a chat completion request over an active session.
// If stream is true, the provider may return SSE chunks; otherwise a single JSON completion.
// In both cases, deltas are delivered through cb until the response is complete.
func (s *SDK) SendPrompt(ctx context.Context, sessionID string, prompt string, stream bool, cb StreamCallback) error {
	return s.SendPromptWithMessages(ctx, sessionID, []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: prompt},
	}, stream, cb)
}

// SendPromptWithMessages sends a full chat transcript (OpenAI roles) over an active session.
// Use this when the app persists history locally and must re-supply prior turns after restart.
func (s *SDK) SendPromptWithMessages(ctx context.Context, sessionID string, messages []openai.ChatCompletionMessage, stream bool, cb StreamCallback) error {
	id := common.HexToHash(sessionID)
	if err := s.ensureProviderForSession(ctx, id); err != nil {
		return err
	}

	req := &gcs.OpenAICompletionRequestExtra{}
	req.Model = sessionID
	req.Messages = messages
	req.Stream = stream

	internalCB := func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			return fmt.Errorf("provider error: %v", errResp.ProviderModelError)
		}
		// Match HTTP /v1/chat/completions: [DONE] is ChunkTypeControl (still "streaming" on the chunk).
		if chunk.Type() == gcs.ChunkTypeControl {
			return cb("", true)
		}
		text := chunk.String()
		isLast := !chunk.IsStreaming()
		return cb(text, isLast)
	}

	_, err := s.proxySender.SendPromptV2(ctx, id, req, internalCB)
	return err
}

// ensureProviderForSession repopulates in-memory provider URL + pubkey after SDK restart (mobile).
func (s *SDK) ensureProviderForSession(ctx context.Context, sessionID common.Hash) error {
	sess, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}
	provAddr := sess.ProviderAddr()
	u, err := s.sessionStorage.GetUser(provAddr.Hex())
	if err != nil {
		return fmt.Errorf("provider cache: %w", err)
	}
	if u != nil {
		return nil
	}
	p, err := s.blockchain.GetProvider(ctx, provAddr)
	if err != nil {
		return fmt.Errorf("provider registry: %w", err)
	}
	if p == nil || p.Endpoint == "" {
		return fmt.Errorf("provider endpoint not found for %s", provAddr.Hex())
	}
	return s.proxySender.EnsureProviderRegistered(ctx, provAddr, p.Endpoint)
}

// --- Helpers ---

func toJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func bigStr(b *big.Int) string {
	if b == nil {
		return "0"
	}
	return b.String()
}

func bigIntStr(b *lib.BigInt) string {
	if b == nil {
		return "0"
	}
	return b.String()
}
