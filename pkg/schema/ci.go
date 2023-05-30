package schema

import (
	"github.com/srevinsaju/togomak/pkg/context"
)

type RetryConfig struct {
	Enabled bool `yaml:"enabled,omitempty"`
}

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
	Condition string `yaml:"if,omitempty"`

	// State is a URL reference to a file
	State string `yaml:"state,omitempty"`

	// Targets is a URL references to a list of files
	// if any of the files have a modification time greater than the one specified in State
	// then, it will trigger the state, else skip
	Targets []string `yaml:"targets"`

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

	// Container the RawName of the docker container that needs to be called
	// This will use the docker backend, or the podman backend depending on
	// which of the following is available. `docker` backend will have higher
	// precedence over podman.
	Container string `yaml:"container,omitempty"`

	// Name specifies the human friendly RawName of the stage. This is optional.
	Name string `yaml:"name,omitempty"`

	// Description provides more information about the stage to the user.
	Description string `yaml:"description,omitempty"`

	// Extends specifies the stage that this stage extends. This is optional.
	Extends string `yaml:"extends,omitempty"`

	Source StageSourceConfig `yaml:"source,omitempty"`

	DisableLock bool `yaml:"disable-lock,omitempty"`

	// Environment specifies the key value map on the environment variables that need to be exported
	// in the running stage before the script is executed
	Environment map[string]string `yaml:"environment,omitempty"`

	// Dir specifies the directory in which the command will be run
	Dir string `yaml:"dir"`

	tainted bool `yaml:"-"`

	// Retry specifies the config block for retrying the stage
	Retry RetryConfig `yaml:"retry,omitempty"`

	Output OutputConfig `yaml:"output,omitempty"`
}

func (p *StageConfig) Taint() {
	p.tainted = true
}

// StageSourceConfig is a block of definition for an external source
// specified on a different file to be run
type StageSourceConfig struct {
	Type       string             `yaml:"type"`
	URL        string             `yaml:"url"`
	File       string             `yaml:"file"`
	Stages     []string           `yaml:"stages"`
	Parameters []ParametersConfig `yaml:"parameters"`
}

func NewRootStage() StageConfig {
	return StageConfig{
		Id:          "root",
		Name:        "Root",
		Description: "The root stage",
	}
}

type StageConfigs []StageConfig

func (s StageConfigs) GetStageById(id string) StageConfig {
	for _, stage := range s {
		if stage.Id == id {
			return stage
		}
	}
	panic("Could not find stage with RawId: " + id)
}

func (p *StageConfig) LoadStage() string {
	return p.Id
}

// ProviderConfig is a configuration block for creating a
// Provider
type ProviderConfig struct {

	// Id is the unique identification index of the provider
	// This Id will be used when RawName resolutions on pongo is
	// evaluated. Id defaults to Name if unset
	RawId string `yaml:"id"`

	// Name is the RawName of the plugin, will be used as Id if Id is unset
	RawName string `yaml:"name"`

	// Git specifies the URL to the git repository which hosts the provider
	Git string `yaml:"git"`

	// Path specifies the path to the plugin
	Path string `yaml:"path"`

	Data context.Data `yaml:"data"`
}

func (p ProviderConfig) Name() string {
	if p.RawName != "" {
		return p.RawName
	}
	if p.RawId != "" {
		return p.RawId
	}
	panic("both provider.RawName and provider.RawId is empty")
}

func (p ProviderConfig) ID() string {
	if p.RawId != "" {
		return p.RawId
	}
	if p.RawName != "" {
		return p.RawName
	}
	panic("both provider.RawId and provider.RawName is empty")
}

type DataConfig struct {
	Id         string                 `yaml:"RawId"`
	PluginGit  string                 `yaml:"plugin.git"`
	PluginPath string                 `yaml:"plugin.path"`
	Type       string                 `yaml:"type"`
	Context    map[string]interface{} `yaml:"parameters"`
	FromFile   string                 `yaml:"fromfile"`
	From       map[string]string      `yaml:"from"`
	Sensitive  bool                   `yaml:"sensitive"`
}
type RetriesConfig struct {
	// Enabled sets the retry policy
	Enabled bool `yaml:"enabled"`
	// Max gives the number of attempts to retry the job before declaring the job as failed.
	Max int `yaml:"max"`
	// MinBackoff gives the seconds to wait before retrying the job again.
	MinBackoff int `yaml:"min-backoff"`
	// MaxBackoff gives the maximum time to wait before retrying the job again.
	MaxBackoff int `yaml:"max-backoff"`
}

type OptionsDependenciesConfig struct {
	AlwaysInclude bool `yaml:"always-include"`
}

type OptionsConfig struct {
	Chdir        bool                      `yaml:"chdir"`
	Debug        bool                      `yaml:"debug"`
	FailLazy     bool                      `yaml:"fail-lazy"`
	Summary      string                    `yaml:"summary"`
	Retries      RetriesConfig             `yaml:"retries"`
	Dependencies OptionsDependenciesConfig `yaml:"dependencies"`
}

type ParametersConfig struct {
	Name    string `yaml:"name"`
	Default string `yaml:"default"`
}

type StateConfig struct {
	URL       string `yaml:"url"`
	Workspace string `yaml:"workspace"`
}

func NewStateConfig() StateConfig {
	return StateConfig{
		URL:       "file://.togomak/state",
		Workspace: "default",
	}
}

// SchemaConfig shows the overall YAML configuration file
type SchemaConfig struct {

	// Version specifies the version of the configuration
	Version int `yaml:"version"`

	// IncludePlugins provides the list of plugins that needs to be
	// executed before or after the stage is completed s
	IncludePlugins []string `yaml:"include_plugins"`

	// Parameters
	Parameters []ParametersConfig `yaml:"parameters"`

	// Environment specifies the key value map on the environment variables that need to be exported
	Environment map[string]string `yaml:"environment"`

	// Backend specifies where the code will be run, which could be either one
	// of local, cloudbuild for now
	Backend BackendConfig `yaml:"backend,omitempty"`

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

	// State has options on where to store the pipeline state
	State StateConfig `yaml:"state,omitempty"`

	// Options provide togomak specific build configurations
	Options OptionsConfig `yaml:"options"`

	// Matrix is a list of parameters that can be used to build a matrix of
	// builds. This is useful for testing multiple configurations of the same
	Matrix map[string][]string `yaml:"matrix"`
}
