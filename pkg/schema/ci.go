package schema

type StageConfig struct {
	Id      string            `yaml:"id"`
	Backend string            `yaml:"backend"`
	Context map[string]string `yaml:"parameters"`
}

func (p *StageConfig) LoadStage() string {
	return p.Id
}

type ProviderConfig struct { 
	Id      string            `yaml:"id"`
	Backend string            `yaml:"backend"`
	Context map[string]string `yaml:"parameters"`
}

type DataConfig struct {
	Id      string            `yaml:"id"`
	Backend string            `yaml:"backend"`
	Context map[string]string `yaml:"parameters"`
}

type SchemaConfig struct {
	Version        int               `yaml:"version"`
	IncludePlugins []string          `yaml:"include_plugins"`
	Parameters     map[string]string `yaml:"parameters"`

	Providers []ProviderConfig `yaml:"providers"`
	Stages    []StageConfig    `yaml:"stages"`
	Data      []DataConfig     `yaml:"data"`
}
