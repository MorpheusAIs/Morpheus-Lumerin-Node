package system

import "context"

type osConfigurator interface {
	GetConfig() (*Config, error)
	ApplyConfig(cfg *Config) error
	GetFileDescriptors(ctx context.Context, pid int) ([]FD, error)
}
