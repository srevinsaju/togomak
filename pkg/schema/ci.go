package schema

// StageConfig is a block of definition for a stage to run.
// Stage is a job which run internally, concurrently by default
// to achieve a specific task.
type StageConfig struct {

	// Id is a unique key to a stage. Internally the CICD system
	// will use these Id 's as a method of referencing other stages
	Id string `yaml:"id"`

	// DependsOn helps to specify the order of execution by
	// specifying the dependencies. This order will be used to generate
	// an internal graph, which will be topologically sorted.
	DependsOn []string `yaml:"depends-on,omitempty"`

	// Condition is a boolean value that determines if the stage will be
	// run or not. A user can however invoke the tool by explicitly calling
	// the Stage by the Id, and it will run nevertheless.
	// Condition may have pongo expressions which will be evaluated before the
	// stage is called
	Condition string `yaml:"condition,omitempty"`

	// Plugin
	Plugin string `yaml:"plugin,omitempty"`

	// Script multiline shell scripts can be specified here. The container, if
	// specified will be set to have 'sh' as entrypoint and the shell scripts
	// will be passed with -c argument. This does not fail if any of the commands
	// within the shell script fail by default. Use `set -e` to explicitly configure
	// the shell behaviour. Args and Script are mutually exclusive and they should
	// not be specified at the same time.
	Script string `yaml:"script,omitempty"`

	// Args which needs to be passed to Container if they are specified.
	// Args and Script should not be specified simultaneously
	Args []string `yaml:"args,omitempty"`

	// Container the name of the docker container that needs to be called
	// This will use the docker backend, or the podman backend depending on
	// which of the following is available. `docker` backend will have higher
	// precedence over podman.
	Container string `yaml:"container,omitempty"`

	// Name specifies the human friendly name of the stage. This is optional.
	Name string `yaml:"name,omitempty"`

	// Description provides more information about the stage to the user.
	Description string `yaml:"description,omitempty"`
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

// ProviderConfig is a configuration block for creating a
// Provider
type ProviderConfig struct {

	// Id is the unique identification index of the provider
	// This Id will be used when name resolutions on pongo is
	// evaluated
	Id string `yaml:"id"`

	// Git specifies the URL to the git repository which hosts the provider
	Git string `yaml:"git"`

	// Path specifies the path to the plugin
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

// SchemaConfig shows the overall YAML configuration file
type SchemaConfig struct {

	// Version specifies the version of the configuration
	Version int `yaml:"version"`

	// IncludePlugins provides the list of plugins that needs to be
	// executed before or after the stage is completed s
	IncludePlugins []string `yaml:"include_plugins"`

	// Parameters
	Parameters map[string]string `yaml:"parameters"`

	// Extends does an internal deep merge of yaml maps so that
	// on can inherit the properties of another stage without having to write
	// most of the content
	Extends string `yaml:"extends"`

	// Providers are plugins which can be written by users
	// which can do one, or all of the following: gather information,
	// check if all the pre-requisites for running a provider is met
	// and do the job
	Providers []ProviderConfig `yaml:"providers"`

	// Stages are user defined jobs which will need to run. The order
	// of execution depends on the StageConfigs DependsOn parameter
	Stages StageConfigs `yaml:"stages"`

	// Data - have not decided what to do with this yet
	Data []DataConfig `yaml:"data"`
}
