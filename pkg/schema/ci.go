package schema

type StageConfig struct {
	Id          string   `yaml:"id"`
	DependsOn   []string `yaml:"depends-on,omitempty"`
	Condition   string   `yaml:"condition,omitempty"`
	Plugin      string   `yaml:"plugin,omitempty"`
	Script      string   `yaml:"script,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Container   string   `yaml:"container,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type StageConfigs []StageConfig

func (s StageConfigs) GetStageById(id string) StageConfig {
	for _, stage := range s {
		if stage.Id == id {
			return stage
		}
	}
	panic("Could not find stage with id: " + id)
}

func (p *StageConfig) LoadStage() string {
	return p.Id
}

type ProviderConfig struct {
	Id string `yaml:"id"`

	Git  string `yaml:"git"`
	Path string `yaml:"path"`
}

type DataConfig struct {
	Id         string                 `yaml:"id"`
	PluginGit  string                 `yaml:"plugin.git"`
	PluginPath string                 `yaml:"plugin.path"`
	Type       string                 `yaml:"type"`
	Context    map[string]interface{} `yaml:"parameters"`
	FromFile   string                 `yaml:"fromfile"`
	From       map[string]string      `yaml:"from"`
	Sensitive  bool                   `yaml:"sensitive"`
}

type SchemaConfig struct {
	Version        int               `yaml:"version"`
	IncludePlugins []string          `yaml:"include_plugins"`
	Parameters     map[string]string `yaml:"parameters"`

	Providers []ProviderConfig `yaml:"providers"`
	Stages    StageConfigs     `yaml:"stages"`
	Data      []DataConfig     `yaml:"data"`
}
