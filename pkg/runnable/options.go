package runnable

type Config struct {
	Status *Status
	Parent *ParentConfig
	Hook   bool
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

func WithParent(parent ParentConfig) Option {
	return func(c *Config) {
		c.Parent = &parent
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		Status: &Status{Status: StatusRunning},
		Parent: nil,
		Hook:   false,
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
