package runnable

import (
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	Status *Status
	Parent *ParentConfig
	Hook   bool

	Paths *path.Path

	Each map[string]cty.Value

	Behavior *behavior.Behavior
}

type ParentConfig struct {
	Id   string
	Name string
}

type Option func(*Config)

func WithStatus(status StatusType) Option {
	return func(c *Config) {
		c.Status = &Status{
			Status: status,
		}
	}
}

func WithStatusOutput(artifact string) Option {
	return func(c *Config) {
		c.Status.Output = artifact
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

func WithEach(k cty.Value, v cty.Value) Option {
	return func(c *Config) {
		c.Each = map[string]cty.Value{
			"key":   k,
			"value": v,
		}
	}
}

func WithBehavior(behavior *behavior.Behavior) Option {
	return func(c *Config) {
		c.Behavior = behavior
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		Status:   &Status{Status: StatusRunning},
		Parent:   nil,
		Hook:     false,
		Paths:    nil,
		Behavior: nil,
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
