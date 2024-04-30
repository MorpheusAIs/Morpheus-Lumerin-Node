package config

import (
	"runtime"
	"strings"
)

type DerivedConfig struct {
	WalletAddress string
}

// Validation tags described here: https://pkg.go.dev/github.com/go-playground/validator/v10
type Config struct {
	Blockchain struct {
		EthNodeAddress string `env:"ETH_NODE_ADDRESS"   flag:"eth-node-address"   validate:"required,url"`
		EthLegacyTx    bool   `env:"ETH_NODE_LEGACY_TX" flag:"eth-node-legacy-tx" desc:"use it to disable EIP-1559 transactions"`
	}
	Environment string `env:"ENVIRONMENT" flag:"environment"`
	Marketplace struct {
		ProviderRegistryAddress string `env:"PROVIDER_REGISTRY_ADDRESS" flag:"provider-registry-address"   validate:"required_if=Disable false,omitempty,eth_addr"`
		ModelRegistryAddress    string `env:"MODEL_REGISTRY_ADDRESS" flag:"model-registry-address"   validate:"required_if=Disable false,omitempty,eth_addr"`
		MarketplaceAddress      string `env:"MARKETPLACE_ADDRESS" flag:"marketplace-address"   validate:"required_if=Disable false,omitempty,eth_addr"`
		Mnemonic                string `env:"CONTRACT_MNEMONIC"     flag:"contract-mnemonic"  validate:"required_without=WalletPrivateKey|required_if=Disable false"`
		WalletPrivateKey        string `env:"WALLET_PRIVATE_KEY"    flag:"wallet-private-key" validate:"required_without=Mnemonic|required_if=Disable false"`
	}
	Log struct {
		Color           bool   `env:"LOG_COLOR"            flag:"log-color"`
		FolderPath      string `env:"LOG_FOLDER_PATH"      flag:"log-folder-path"      validate:"omitempty,dirpath"    desc:"enables file logging and sets the folder path"`
		IsProd          bool   `env:"LOG_IS_PROD"          flag:"log-is-prod"          validate:""                     desc:"affects the format of the log output"`
		JSON            bool   `env:"LOG_JSON"             flag:"log-json"`
		LevelApp        string `env:"LOG_LEVEL_APP"        flag:"log-level-app"        validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelConnection string `env:"LOG_LEVEL_CONNECTION" flag:"log-level-connection" validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelProxy      string `env:"LOG_LEVEL_PROXY"      flag:"log-level-proxy"      validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelScheduler  string `env:"LOG_LEVEL_SCHEDULER"  flag:"log-level-scheduler"  validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
		LevelContract   string `env:"LOG_LEVEL_CONTRACT"   flag:"log-level-contract"   validate:"omitempty,oneof=debug info warn error dpanic panic fatal"`
	}
	Proxy struct {
		Address        string `env:"PROXY_ADDRESS" flag:"proxy-address" validate:"required,hostname_port"`
		MaxCachedDests int    `env:"PROXY_MAX_CACHED_DESTS" flag:"proxy-max-cached-dests" validate:"required,number" desc:"maximum number of cached destinations per proxy"`
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

	// Marketplace

	// normalizes private key
	// TODO: convert and validate to ecies.PrivateKey
	cfg.Marketplace.WalletPrivateKey = strings.TrimPrefix(cfg.Marketplace.WalletPrivateKey, "0x")

	// Log

	if cfg.Log.LevelConnection == "" {
		cfg.Log.LevelConnection = "info"
	}
	if cfg.Log.LevelProxy == "" {
		cfg.Log.LevelProxy = "info"
	}
	if cfg.Log.LevelScheduler == "" {
		cfg.Log.LevelScheduler = "info"
	}
	if cfg.Log.LevelContract == "" {
		cfg.Log.LevelContract = "debug"
	}
	if cfg.Log.LevelApp == "" {
		cfg.Log.LevelApp = "debug"
	}

	// Proxy
	if cfg.Proxy.MaxCachedDests == 0 {
		cfg.Proxy.MaxCachedDests = 5
	}

	// System

	// cfg.System.Enable = true // TODO: Temporary override, remove this line

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
		cfg.Web.Address = "0.0.0.0:8080"
	}
	if cfg.Web.PublicUrl == "" {
		cfg.Web.PublicUrl = "http://localhost:8080"
	}
}

// GetSanitized returns a copy of the config with sensitive data removed
// explicitly adding each field here to avoid accidentally leaking sensitive data
func (cfg *Config) GetSanitized() interface{} {
	publicCfg := Config{}

	publicCfg.Blockchain.EthLegacyTx = cfg.Blockchain.EthLegacyTx
	publicCfg.Environment = cfg.Environment

	publicCfg.Log.Color = cfg.Log.Color
	publicCfg.Log.FolderPath = cfg.Log.FolderPath
	publicCfg.Log.IsProd = cfg.Log.IsProd
	publicCfg.Log.JSON = cfg.Log.JSON
	publicCfg.Log.LevelApp = cfg.Log.LevelApp
	publicCfg.Log.LevelConnection = cfg.Log.LevelConnection
	publicCfg.Log.LevelProxy = cfg.Log.LevelProxy
	publicCfg.Log.LevelScheduler = cfg.Log.LevelScheduler

	publicCfg.Proxy.Address = cfg.Proxy.Address
	publicCfg.Proxy.MaxCachedDests = cfg.Proxy.MaxCachedDests

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
