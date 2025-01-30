package config

import (
	"fmt"
	"math/big"
	"runtime"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	"github.com/ethereum/go-ethereum/common"
)

type DerivedConfig struct {
	WalletAddress common.Address
	ChainID       *big.Int
	EthNodeURLs   []string
}

// Validation tags described here: https://pkg.go.dev/github.com/go-playground/validator/v10
type Config struct {
	App struct {
		ResetKeychain bool `env:"APP_RESET_KEYCHAIN" flag:"app-reset-keychain" desc:"reset keychain on start"`
	}
	Blockchain struct {
		ChainID        int    `env:"ETH_NODE_CHAIN_ID"  flag:"eth-node-chain-id"  validate:"number"`
		EthNodeAddress string `env:"ETH_NODE_ADDRESS"   flag:"eth-node-address"   validate:"omitempty,url"`
		EthLegacyTx    bool   `env:"ETH_NODE_LEGACY_TX" flag:"eth-node-legacy-tx" desc:"use it to disable EIP-1559 transactions"`
		// ExplorerApiUrl     string          `env:"EXPLORER_API_URL"   flag:"explorer-api-url"   validate:"required,url"`
		BlockscoutApiUrl   string          `env:"BLOCKSCOUT_API_URL" flag:"blockscout-api-url" validate:"required,url"`
		ExplorerRetryDelay time.Duration   `env:"EXPLORER_RETRY_DELAY" flag:"explorer-retry-delay" validate:"omitempty,duration" desc:"delay between retries"`
		ExplorerMaxRetries uint8           `env:"EXPLORER_MAX_RETRIES" flag:"explorer-max-retries" validate:"omitempty,gte=0" desc:"max retries for explorer requests"`
		UseSubscriptions   bool            `env:"ETH_NODE_USE_SUBSCRIPTIONS"  flag:"eth-node-use-subscriptions"  desc:"set it to true to enable subscriptions for blockchain events, otherwise default polling will be used"`
		PollingInterval    time.Duration   `env:"ETH_NODE_POLLING_INTERVAL" flag:"eth-node-polling-interval" validate:"omitempty,duration" desc:"interval for polling eth node for new events"`
		MaxReconnects      int             `env:"ETH_NODE_MAX_RECONNECTS" flag:"eth-node-max-reconnects" validate:"omitempty,gte=0" desc:"max reconnects to eth node"`
		Multicall3Addr     *common.Address `env:"MULTICALL3_ADDR" flag:"multicall3-addr" validate:"omitempty,eth_addr" desc:"multicall3 custom contract address"`
	}
	Environment string `env:"ENVIRONMENT" flag:"environment"`
	Marketplace struct {
		DiamondContractAddress *common.Address `env:"DIAMOND_CONTRACT_ADDRESS" flag:"diamond-address"   validate:"omitempty,eth_addr"`
		MorTokenAddress        *common.Address `env:"MOR_TOKEN_ADDRESS"        flag:"mor-token-address" validate:"omitempty,eth_addr"`
		WalletPrivateKey       *lib.HexString  `env:"WALLET_PRIVATE_KEY"       flag:"wallet-private-key"     desc:"if set, will use this private key to sign transactions, otherwise it will be retrieved from the system keychain"`
	}
	Log struct {
		Color        bool   `env:"LOG_COLOR"            flag:"log-color"`
		FolderPath   string `env:"LOG_FOLDER_PATH"      flag:"log-folder-path"      validate:"omitempty,dirpath"    desc:"enables file logging and sets the folder path"`
		IsProd       bool   `env:"LOG_IS_PROD"          flag:"log-is-prod"          validate:""                     desc:"affects the format of the log output"`
		JSON         bool   `env:"LOG_JSON"             flag:"log-json"`
		LevelApp     string `env:"LOG_LEVEL_APP"        flag:"log-level-app"        validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelTCP     string `env:"LOG_LEVEL_TCP" flag:"log-level-tcp" validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelEthRPC  string `env:"LOG_LEVEL_ETH_RPC"    flag:"log-level-eth-rpc"        validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelStorage string `env:"LOG_LEVEL_STORAGE"     flag:"log-level-storage"     validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
	}
	Proxy struct {
		Address            string    `env:"PROXY_ADDRESS" flag:"proxy-address" validate:"required,hostname_port"`
		StoragePath        string    `env:"PROXY_STORAGE_PATH"    flag:"proxy-storage-path"    validate:"omitempty,dirpath" desc:"enables file storage and sets the folder path"`
		StoreChatContext   *lib.Bool `env:"PROXY_STORE_CHAT_CONTEXT" flag:"proxy-store-chat-context" desc:"store chat context in the proxy storage"`
		ForwardChatContext *lib.Bool `env:"PROXY_FORWARD_CHAT_CONTEXT" flag:"proxy-forward-chat-context" desc:"prepend whole stored message history to the prompt"`
		ModelsConfigPath   string    `env:"MODELS_CONFIG_PATH" flag:"models-config-path" validate:"omitempty"`
		RatingConfigPath   string    `env:"RATING_CONFIG_PATH" flag:"rating-config-path" validate:"omitempty" desc:"path to the rating config file"`
		CookieFilePath     string    `env:"COOKIE_FILE_PATH" flag:"cookie-file-path" validate:"omitempty" desc:"path to the cookie file"`
		AuthConfigFilePath string    `env:"AUTH_CONFIG_FILE_PATH" flag:"auth-config-file-path" validate:"omitempty"`
	}
	System struct {
		Enable           bool   `env:"SYS_ENABLE"              flag:"sys-enable" desc:"enable system level configuration adjustments"`
		LocalPortRange   string `env:"SYS_LOCAL_PORT_RANGE"    flag:"sys-local-port-range"    desc:""`
		NetdevMaxBacklog string `env:"SYS_NET_DEV_MAX_BACKLOG" flag:"sys-netdev-max-backlog"  desc:""`
		RlimitHard       uint64 `env:"SYS_RLIMIT_HARD"         flag:"sys-rlimit-hard"         desc:""`
		RlimitSoft       uint64 `env:"SYS_RLIMIT_SOFT"         flag:"sys-rlimit-soft"         desc:""`
		Somaxconn        string `env:"SYS_SOMAXCONN"           flag:"sys-somaxconn"           desc:""`
		TcpMaxSynBacklog string `env:"SYS_TCP_MAX_SYN_BACKLOG" flag:"sys-tcp-max-syn-backlog" desc:""`
	}
	Web struct {
		Address   string `env:"WEB_ADDRESS"    flag:"web-address"    validate:"required,hostname_port" desc:"http server address host:port"`
		PublicUrl string `env:"WEB_PUBLIC_URL" flag:"web-public-url" validate:"omitempty,url"          desc:"public url of the proxyrouter, falls back to web-address if empty" `
	}
}

func (cfg *Config) SetDefaults() {
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	// Blockchain
	if cfg.Blockchain.MaxReconnects == 0 {
		cfg.Blockchain.MaxReconnects = 30
	}
	if cfg.Blockchain.PollingInterval == 0 {
		cfg.Blockchain.PollingInterval = 10 * time.Second
	}
	if cfg.Blockchain.Multicall3Addr.Cmp(common.Address{}) == 0 {
		cfg.Blockchain.Multicall3Addr = &multicall.MULTICALL3_ADDR
	}
	if cfg.Blockchain.ExplorerRetryDelay == 0 {
		cfg.Blockchain.ExplorerRetryDelay = 5 * time.Second
	}
	if cfg.Blockchain.ExplorerMaxRetries == 0 {
		cfg.Blockchain.ExplorerMaxRetries = 5
	}

	// Log

	if cfg.Log.LevelTCP == "" {
		cfg.Log.LevelTCP = "info"
	}
	if cfg.Log.LevelApp == "" {
		cfg.Log.LevelApp = "debug"
	}
	if cfg.Log.LevelEthRPC == "" {
		cfg.Log.LevelEthRPC = "info"
	}
	if cfg.Log.LevelStorage == "" {
		cfg.Log.LevelStorage = "info"
	}

	// System

	if cfg.System.LocalPortRange == "" {
		cfg.System.LocalPortRange = "1024 65535"
	}
	if cfg.System.TcpMaxSynBacklog == "" {
		cfg.System.TcpMaxSynBacklog = "100000"
	}
	if cfg.System.Somaxconn == "" && runtime.GOOS == "linux" {
		cfg.System.Somaxconn = "100000"
	}
	if cfg.System.Somaxconn == "" && runtime.GOOS == "darwin" {
		// setting high value like 1000000 on darwin
		// for some reason blocks incoming connections
		// TODO: investigate best value for this
		cfg.System.Somaxconn = "2048"
	}
	if cfg.System.NetdevMaxBacklog == "" {
		cfg.System.NetdevMaxBacklog = "100000"
	}
	if cfg.System.RlimitSoft == 0 {
		cfg.System.RlimitSoft = 524288
	}
	if cfg.System.RlimitHard == 0 {
		cfg.System.RlimitHard = 524288
	}

	// Proxy

	if cfg.Proxy.Address == "" {
		cfg.Proxy.Address = "0.0.0.0:3333"
	}
	if cfg.Web.Address == "" {
		cfg.Web.Address = "0.0.0.0:8082"
	}
	if cfg.Web.PublicUrl == "" {
		// handle cases without domain (ex: :8082)
		if string(cfg.Web.Address[0]) == ":" {
			cfg.Web.PublicUrl = fmt.Sprintf("http://localhost%s", cfg.Web.Address)
		} else {
			cfg.Web.PublicUrl = fmt.Sprintf("http://%s", strings.Replace(cfg.Web.Address, "0.0.0.0", "localhost", -1))
		}
	}
	if cfg.Proxy.StoragePath == "" {
		cfg.Proxy.StoragePath = "./data/badger/"
	}
	if cfg.Proxy.StoreChatContext.Bool == nil {
		val := true
		cfg.Proxy.StoreChatContext = &lib.Bool{Bool: &val}
	}
	if cfg.Proxy.ForwardChatContext.Bool == nil {
		val := true
		cfg.Proxy.ForwardChatContext = &lib.Bool{Bool: &val}
	}
	if cfg.Proxy.RatingConfigPath == "" {
		cfg.Proxy.RatingConfigPath = "./rating-config.json"
	}
	if cfg.Proxy.CookieFilePath == "" {
		cfg.Proxy.CookieFilePath = "./.cookie"
	}
	if cfg.Proxy.AuthConfigFilePath == "" {
		cfg.Proxy.AuthConfigFilePath = "./proxy.conf"
	}
}

// GetSanitized returns a copy of the config with sensitive data removed
// explicitly adding each field here to avoid accidentally leaking sensitive data
func (cfg *Config) GetSanitized() interface{} {
	publicCfg := Config{}

	publicCfg.Blockchain.EthLegacyTx = cfg.Blockchain.EthLegacyTx
	publicCfg.Blockchain.ChainID = cfg.Blockchain.ChainID
	publicCfg.Blockchain.MaxReconnects = cfg.Blockchain.MaxReconnects
	publicCfg.Blockchain.PollingInterval = cfg.Blockchain.PollingInterval
	publicCfg.Blockchain.UseSubscriptions = cfg.Blockchain.UseSubscriptions
	// publicCfg.Blockchain.ExplorerApiUrl = cfg.Blockchain.ExplorerApiUrl

	publicCfg.Environment = cfg.Environment

	publicCfg.Marketplace.DiamondContractAddress = cfg.Marketplace.DiamondContractAddress
	publicCfg.Marketplace.MorTokenAddress = cfg.Marketplace.MorTokenAddress

	publicCfg.Log.Color = cfg.Log.Color
	publicCfg.Log.FolderPath = cfg.Log.FolderPath
	publicCfg.Log.IsProd = cfg.Log.IsProd
	publicCfg.Log.JSON = cfg.Log.JSON
	publicCfg.Log.LevelApp = cfg.Log.LevelApp
	publicCfg.Log.LevelTCP = cfg.Log.LevelTCP
	publicCfg.Log.LevelEthRPC = cfg.Log.LevelEthRPC

	publicCfg.Proxy.Address = cfg.Proxy.Address
	publicCfg.Proxy.ModelsConfigPath = cfg.Proxy.ModelsConfigPath
	publicCfg.Proxy.StoragePath = cfg.Proxy.StoragePath
	publicCfg.Proxy.StoreChatContext = cfg.Proxy.StoreChatContext
	publicCfg.Proxy.ForwardChatContext = cfg.Proxy.ForwardChatContext
	publicCfg.Proxy.RatingConfigPath = cfg.Proxy.RatingConfigPath

	publicCfg.System.Enable = cfg.System.Enable
	publicCfg.System.LocalPortRange = cfg.System.LocalPortRange
	publicCfg.System.NetdevMaxBacklog = cfg.System.NetdevMaxBacklog
	publicCfg.System.RlimitHard = cfg.System.RlimitHard
	publicCfg.System.RlimitSoft = cfg.System.RlimitSoft
	publicCfg.System.Somaxconn = cfg.System.Somaxconn
	publicCfg.System.TcpMaxSynBacklog = cfg.System.TcpMaxSynBacklog

	publicCfg.Web.Address = cfg.Web.Address
	publicCfg.Web.PublicUrl = cfg.Web.PublicUrl

	return publicCfg
}
