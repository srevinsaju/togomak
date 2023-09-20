package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"os/exec"
)

const StageBlock = "stage"

// StageContainerVolume allows configuring which volumes can be mounted
type StageContainerVolume struct {
	// Source sets the path on the host which needs to be mounted
	Source hcl.Expression `hcl:"source" json:"source"`

	// Destination sets the path on the container where the value specified in the Source
	// needs to be mounted
	Destination hcl.Expression `hcl:"destination" json:"destination"`
}

// StageContainerVolumes are a list of StageContainerVolume
type StageContainerVolumes []StageContainerVolume

// StageContainerPort allows configuring the container port specifications which will be passed
// through the docker context
type StageContainerPort struct {
	// Hostname is optional, defaults to localhost if unspecified
	Hostname hcl.Expression `hcl:"host,optional" json:"host"`

	// ContainerPort allows you to configure the port on the container that will be exposed
	ContainerPort hcl.Expression `hcl:"container_port" json:"container_port"`

	// Port allows you to configure the port on the host that will be exposed
	Port hcl.Expression `hcl:"port" json:"port"`
}

// StageContainerPorts are a list of StageContainerPort
type StageContainerPorts []StageContainerPort

// StageContainer if defined on Stage uses a compatible docker executor
// to run the stage
type StageContainer struct {
	// Image sets the name of the docker container image
	Image hcl.Expression `hcl:"image" json:"image"`

	// Volumes have a list of host path volume mapping which is bound on docker run
	Volumes StageContainerVolumes `hcl:"volume,block" json:"volumes"`

	// Ports have a list of ports that needs to be exposed from the container
	Ports StageContainerPorts `hcl:"ports,optional" json:"ports"`

	// Entrypoint allows you to specify the entrypoint of the container
	// If the Entrypoint is null, it defaults to the container's default entrypoint
	Entrypoint hcl.Expression `hcl:"entrypoint,optional" json:"entrypoint"`

	// Stdin connect containers stdin to the host stdin
	Stdin bool `hcl:"stdin,optional" json:"stdin"`
}

// Stages are a list of Stage
type Stages []Stage

// StageEnvironment allows you to set environment variables in the stage
type StageEnvironment struct {
	// Name gives the key of the environment variable that needs to be set
	// hcl variables are disallowed in this context
	Name string `hcl:"name" json:"name"`

	// Value gives the value of the environment variable that needs to be set
	// Only string values are accepted in this argument
	Value hcl.Expression `hcl:"value" json:"value"`
}

// StageRetry configures the retry parameters of a stage
type StageRetry struct {
	// Enabled enables the stage retry policy
	Enabled bool `hcl:"enabled" json:"enabled"`
	// Attempts show the number of times the stage needs to be retried before permanently failing it
	Attempts int `hcl:"attempts" json:"attempts"`
	// ExponentialBackoff accepts a boolean value which allows you to choose if exponential backoff needs to be
	// enabled while retrying. If false, the stage will be retried immediately on failure
	ExponentialBackoff bool `hcl:"exponential_backoff" json:"exponential_backoff"`

	// MinBackoff accepts time in seconds before it needs to retry at the minimum
	MinBackoff int `hcl:"min_backoff" json:"min_backoff"`

	// MaxBackoff accepts the maximum delay a stage needs to wait before retrying.
	// This is only applicable if ExponentialBackoff is true
	MaxBackoff int `hcl:"max_backoff" json:"max_backoff"`
}

// StageUse allows you to use a macro to run the stage
type StageUse struct {
	// Macro
	Macro      hcl.Expression `hcl:"macro" json:"macro"`
	Parameters hcl.Expression `hcl:"parameters,optional" json:"parameters"`

	Chdir hcl.Expression `hcl:"chdir,optional" json:"chdir"`
}

// StageDaemon configures the daemon properties and lifecycle of a daemon
type StageDaemon struct {
	// Enabled when set to true, the stage is considered as a daemon (or a service) stage
	Enabled bool `hcl:"enabled" json:"enabled"`

	// Timeout how long the service needs to wait before killing itself
	Timeout int `hcl:"timeout,optional" json:"timeout"`

	// Lifecycle rules tell the termination policy of a daemon stage
	Lifecycle *DaemonLifecycle `hcl:"lifecycle,block" json:"lifecycle"`
}

// StagePostHook is a stage which runs immediately after the stage is run
// It accepts all the properties of CoreStage.
// In addition, it also receives certain properties like this.status
// on runtime which can decide if the stage failed or succeeded.
// StagePostHook run regardless of if the stage failed or succeeded.
type StagePostHook struct {
	Stage CoreStage `hcl:"stage,block" json:"stage"`
}

// StagePreHook is a stage that runs immediately before the stage is run,
// after CoreStage.Use block is evaluated. It accepts all properties
// as that of the CoreStage.
type StagePreHook struct {
	Stage CoreStage `hcl:"stage,block" json:"stage"`
}

// Stage is an atomic runnable unit which can perform any operation like running UNIX
// scripts, docker containers, etc. A Stage receives all properties as that of CoreStage
// along with an Id which is used by Stages to uniquely identify a stage.
type Stage struct {
	Id        string `hcl:"id,label" json:"id"`
	CoreStage `hcl:",remain"`

	// Lifecycle rules tell the termination policy of a daemon stage
	Lifecycle *Lifecycle `hcl:"lifecycle,block" json:"lifecycle"`
}

// CoreStage is an abstract struct which is implemented by Stage, StagePreHook, StagePostHook,
// PreStage and PostStage. It is also used within Macro
type CoreStage struct {
	ctx            context.Context
	ctxInitialised bool
	terminated     bool

	// DependsOn accepts a list of dependencies that this stage needs to explicitly
	// wait for before executing it. DependsOn accepts only togomak references, as in
	// stage.<name_of_stage>, etc. The address of the block needs to be specified and
	// this will be used as additional input while mapping out the dependency tree
	DependsOn hcl.Expression `hcl:"depends_on,optional" json:"depends_on"`

	// Condition accepts a boolean value which is evaluated after the dependency tree is generated
	// to check if the stage can be executed, or skipped.
	// If stages are whitelisted as command line arguments on the togomak CLI, then that takes precedence
	Condition hcl.Expression `hcl:"if,optional" json:"if"`

	// TODO: implemennt for each
	// ForEach - not implemented yet
	// ForEach   hcl.Expression `hcl:"for_each,optional" json:"for_each"`

	// Use is a block which allows accepting a macro to be substituted instead of a stage
	// It accepts a single block of StageUse struct
	Use *StageUse `hcl:"use,block" json:"use"`

	// Daemon block allows you to configure if a stage needs to run as a daemon (or a service)
	// These blocks allow the command to continue running across multiple stages if they are
	// configured through StageDaemon configuration block. A stage can have only a single
	// StageDaemon configuration
	Daemon *StageDaemon `hcl:"daemon,block" json:"daemon"`

	// Retry block allows you to retry the stage in case it has failed
	// It is worth noting that the PreHook and the PostHook hooks are also retried
	// Additional documentation on the Retry block is available on the StageRetry block
	Retry *StageRetry `hcl:"retry,block" json:"retry"`

	// Name allows you to set a friendly name for the stage
	Name string `hcl:"name,optional" json:"name"`

	// Dir shows the working directory where the stage would be executed.
	// The directory must exist, otherwise the stage fails.
	Dir hcl.Expression `hcl:"dir,optional" json:"dir"`

	// Script accepts a multiline script which is executed using a Bourne shell, or a bash instance, or a shell
	// defined through the Shell argument. In normal cases, togomak checks for the existence of the bash
	// shell and also includes extra `set -eu` options on the script. If `bash` executable could not be
	// detected, it falls back to `sh` with `-e` set.
	// Script and Args are mutually exclusive and only one of them should be specified at a time, otherwise it
	// gives out an error
	Script hcl.Expression `hcl:"script,optional" json:"script"`

	// Shell accepts a string with the name of the shell, or the absolute path to the shell
	Shell hcl.Expression `hcl:"shell,optional" json:"shell"`

	// Args accepts a list of arguments. The functionality may differ based on whether Container is specified or not.
	// If Args is directly used, it may be executed directly without invoking a shell, as defined in Shell.
	// If Args are executed along a container, it could pass the arguments to the Docker container's StageContainer.Entrypoint
	Args hcl.Expression `hcl:"args,optional" json:"args"`

	// Container allows you to use a Docker container image as a backend
	Container *StageContainer `hcl:"container,block" json:"container"`

	// Environment accepts multiple environment key-value pairs which will be exported
	// in addition to the existing env vars from the host
	Environment []*StageEnvironment `hcl:"env,block" json:"environment"`

	// PreHook
	PreHook  []*StagePreHook  `hcl:"pre_hook,block" json:"pre_hook"`
	PostHook []*StagePostHook `hcl:"post_hook,block" json:"post_hook"`

	process                 *exec.Cmd
	macroWhitelistedStages  []string
	dependsOnVariablesMacro []hcl.Traversal
	ContainerId             string
}

type Lifecycle struct {
	// Phase type of the phase needs to be specified
	Phase []string `hcl:"phase,optional" json:"stage"`

	// Timeout how long the service needs to wait before killing itself
	Timeout hcl.Expression `hcl:"timeout,optional" json:"timeout"`
}

// PreStage is a special stage identified by `togomak.pre` which is always run before
// all stages are run. You can only specify a single PreStage instance
type PreStage struct {
	CoreStage `hcl:",remain"`
}

// PostStage is a special stage identified by `togomak.post` which is run after all
// the stages complete.
type PostStage struct {
	CoreStage `hcl:",remain"`
}
