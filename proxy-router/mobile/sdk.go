package mobile

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"path/filepath"
	"time"

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
	"github.com/ethereum/go-ethereum/rpc"
	openai "github.com/sashabaranov/go-openai"
	"github.com/tyler-smith/go-bip39"
)

const (
	DefaultDerivationPath        = "m/44'/60'/0'/0/0"
	DefaultCNodePNodeTimeout     = 120 * time.Second
	DefaultCNodePNodeMaxRetries  = 3
	DefaultCNodeAudioMaxRetries  = 1
)

// Config holds the minimal configuration needed for the mobile SDK.
type Config struct {
	DataDir       string // persistent storage root (chat history, etc.)
	EthNodeURL    string // Ethereum JSON-RPC endpoint
	ChainID       int64  // e.g. 8453 (Base), 84532 (Base Sepolia)
	DiamondAddr   string // diamond proxy contract (hex with 0x prefix)
	MorTokenAddr  string // MOR token contract (hex with 0x prefix)
	BlockscoutURL string // Blockscout API v2 base URL
	LogLevel      string // "debug", "info", "warn", "error" (default: "info")
}

// SDK is the main entry point for mobile applications.
// It wraps the proxy-router's internal packages behind a clean public API.
type SDK struct {
	cfg            Config
	log            lib.ILogger
	rpcClient      *rpc.Client
	ethClient      *ethclient.Client
	wallet         *wallet.KeychainWallet
	walletStorage  *MemoryKeyValueStorage
	blockchain     *blockchainapi.BlockchainService
	proxySender    *proxyapi.ProxyServiceSender
	sessionRepo    *sessionrepo.SessionRepositoryCached
	chatStorage    gcs.ChatStorageInterface
	storage        *storages.Storage
	sessionStorage *storages.SessionStorage
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

	// Dial Ethereum RPC
	rpcClient, err := rpc.Dial(cfg.EthNodeURL)
	if err != nil {
		return nil, fmt.Errorf("dial eth node: %w", err)
	}

	ethClient := ethclient.NewClient(rpcClient)

	// Verify chain ID
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	chainID, err := ethClient.ChainID(ctx)
	if err != nil {
		rpcClient.Close()
		return nil, fmt.Errorf("get chain ID: %w", err)
	}
	if cfg.ChainID != 0 && chainID.Int64() != cfg.ChainID {
		rpcClient.Close()
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
		rpcClient.Close()
		return nil, fmt.Errorf("create rating: %w", err)
	}

	// Blockchain service — the main orchestrator for on-chain operations
	blockchainSvc := blockchainapi.NewBlockchainService(
		ethClient, mc, diamondAddr, morTokenAddr,
		explorer, w, proxySender, sessionRepo, scorer,
		nil,   // authConfig — not needed for direct SDK calls
		log, rpcLog,
		false, // legacyTx
		nil,   // attestation verifier — can add later for TEE
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
		rpcClient:      rpcClient,
		ethClient:      ethClient,
		wallet:         w,
		walletStorage:  walletKV,
		blockchain:     blockchainSvc,
		proxySender:    proxySender,
		sessionRepo:    sessionRepo,
		chatStorage:    cs,
		storage:        inMemStorage,
		sessionStorage: sessionStorage,
	}

	log.Info("SDK initialized")
	return sdk, nil
}

// Shutdown releases all resources held by the SDK.
func (s *SDK) Shutdown() {
	if s.rpcClient != nil {
		s.rpcClient.Close()
	}
	s.log.Info("SDK shut down")
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

// --- Models ---

// GetAllModels returns all registered models from the blockchain.
func (s *SDK) GetAllModels(ctx context.Context) ([]Model, error) {
	models, err := s.blockchain.GetAllModels(ctx)
	if err != nil {
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

// --- Chat ---

// StreamCallback receives streaming chunks from a chat completion.
// text is the content delta, isLast is true on the final chunk.
type StreamCallback func(text string, isLast bool) error

// SendPrompt sends a chat completion request over an active session.
// Responses stream back through the callback.
func (s *SDK) SendPrompt(ctx context.Context, sessionID string, prompt string, cb StreamCallback) error {
	id := common.HexToHash(sessionID)

	req := &gcs.OpenAICompletionRequestExtra{}
	req.Model = sessionID
	req.Messages = []openai.ChatCompletionMessage{
		{Role: "user", Content: prompt},
	}
	req.Stream = true

	internalCB := func(ctx context.Context, chunk gcs.Chunk, errResp *gcs.AiEngineErrorResponse) error {
		if errResp != nil {
			return fmt.Errorf("provider error: %v", errResp.ProviderModelError)
		}
		text := chunk.String()
		isLast := !chunk.IsStreaming()
		return cb(text, isLast)
	}

	_, err := s.proxySender.SendPromptV2(ctx, id, req, internalCB)
	return err
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
