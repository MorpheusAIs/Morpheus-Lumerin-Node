package system

import (
	"context"
	"time"
)

type osConfigurator interface {
	GetConfig() (*Config, error)
	ApplyConfig(cfg *Config) error
	GetFileDescriptors(ctx context.Context, pid int) ([]FD, error)
}

type EthConnectionValidator interface {
	ValidateEthResourse(ctx context.Context, url string, timeout time.Duration) error
}
