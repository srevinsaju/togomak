package data

import (
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/path"
)

type ProviderConfig struct {
	Paths *path.Path

	Behavior *behavior.Behavior
}

func NewDefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		Paths: nil,
	}
}

type ProviderOption func(*ProviderConfig)

func WithPaths(paths *path.Path) ProviderOption {
	return func(c *ProviderConfig) {
		c.Paths = paths
	}
}

func WithBehavior(behavior *behavior.Behavior) ProviderOption {
	return func(c *ProviderConfig) {
		c.Behavior = behavior
	}
}

func NewProviderConfig(opts ...ProviderOption) *ProviderConfig {
	c := NewDefaultProviderConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c
}
