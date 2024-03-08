package system

import (
	"context"
	"errors"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type WindowsConfigurator struct {
}

func NewOSConfigurator() *WindowsConfigurator {
	return &WindowsConfigurator{}
}

func (c *WindowsConfigurator) GetConfig() (*Config, error) {
	return &Config{}, nil
}

func (c *WindowsConfigurator) ApplyConfig(cfg *Config) error {
	return nil
}

func (*WindowsConfigurator) GetFileDescriptors(ctx context.Context, pid int) ([]FD, error) {
	return nil, ErrNotImplemented
}
