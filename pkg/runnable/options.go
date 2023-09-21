package runnable

import "github.com/srevinsaju/togomak/v1/pkg/path"

type Config struct {
	Status *Status
	Parent *ParentConfig
	Hook   bool

	Paths *path.Path

	Behavior *BehaviorConfig
}

type ParentConfig struct {
	Id   string
	Name string
}

type BehaviorConfig struct {
	CI bool
}

type Option func(*Config)

func WithStatus(status StatusType) Option {
	return func(c *Config) {
		c.Status = &Status{
			Status: status,
		}
	}
}

func WithPaths(paths *path.Path) Option {
	return func(c *Config) {
		c.Paths = paths
	}
}

func WithParent(parent ParentConfig) Option {
	return func(c *Config) {
		c.Parent = &parent
	}
}

func WithBehaviorCI() Option {
	return func(c *Config) {
		c.Behavior.CI = true
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		Status: &Status{Status: StatusRunning},
		Parent: nil,
		Hook:   false,
		Paths:  path.NewDefaultPath(),
		Behavior: &BehaviorConfig{
			CI: false,
		},
	}
}

func NewConfig(options ...Option) *Config {
	c := NewDefaultConfig()
	for _, option := range options {
		option(c)
	}
	return c
}

func WithHook() Option {
	return func(c *Config) {
		c.Hook = true
	}
}
