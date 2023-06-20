package orchestra

import (
	"fmt"
	"strings"
)

type ConfigPipelineStageOperation string

const (
	ConfigPipelineStageDaemonizeOperation    ConfigPipelineStageOperation = "&"
	ConfigPipelineStageRunBlacklistOperation ConfigPipelineStageOperation = "^"
	ConfigPipelineStageRunWhitelistOperation ConfigPipelineStageOperation = "+"
	ConfigPipelineStageRunOperation          ConfigPipelineStageOperation = ""
)

type ConfigPipelineStage struct {
	Id        string
	Operation ConfigPipelineStageOperation
}

func (c ConfigPipelineStage) RunnableId() string {
	return fmt.Sprintf("stage.%s", c.Id)
}

func (c ConfigPipelineStage) Identifier() string {
	return c.RunnableId()
}

type ConfigPipelineStageList []ConfigPipelineStage

func (c ConfigPipelineStageList) Get(runnableId string) (ConfigPipelineStageList, bool) {
	var stages []ConfigPipelineStage
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) {
			stages = append(stages, stage)
		}
	}
	return stages, len(stages) > 0
}

func (c ConfigPipelineStage) Child() ConfigPipelineStage {
	return ConfigPipelineStage{
		Id:        c.Id[strings.IndexRune(c.Id, '.')+1:],
		Operation: c.Operation,
	}

}

func (c ConfigPipelineStageList) Children(runnableId string) ConfigPipelineStageList {
	var stages []ConfigPipelineStage
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) && stage.Identifier() != runnableId {
			stages = append(stages, stage.Child())
		}
	}
	return stages
}

func (c ConfigPipelineStageList) AllOperations(operation ConfigPipelineStageOperation) bool {
	for _, stage := range c {
		if stage.Operation != operation {
			return false
		}
	}
	return true
}

func (c ConfigPipelineStageList) AnyOperations(operation ConfigPipelineStageOperation) bool {
	for _, stage := range c {
		if stage.Operation == operation {
			return true
		}
	}
	return false
}

func (c ConfigPipelineStageList) Marshall() []string {
	var stages []string
	for _, stage := range c {
		stages = append(stages, string(stage.Operation)+stage.Id)
	}
	return stages
}

func (c ConfigPipelineStageList) HasOperationType(operation ConfigPipelineStageOperation) bool {
	for _, stage := range c {
		if stage.Operation == operation {
			return true
		}
	}
	return false
}

func NewConfigPipelineStage(arg string) ConfigPipelineStage {
	var operation ConfigPipelineStageOperation
	if strings.HasPrefix(arg, string(ConfigPipelineStageRunWhitelistOperation)) {
		operation = ConfigPipelineStageRunWhitelistOperation
	} else if strings.HasPrefix(arg, string(ConfigPipelineStageRunBlacklistOperation)) {
		operation = ConfigPipelineStageRunBlacklistOperation
	} else if strings.HasPrefix(arg, string(ConfigPipelineStageDaemonizeOperation)) {
		operation = ConfigPipelineStageDaemonizeOperation
	} else {
		operation = ConfigPipelineStageRunOperation
	}
	id := strings.TrimPrefix(arg, string(operation))
	return ConfigPipelineStage{
		Id:        id,
		Operation: operation,
	}
}

type ConfigPipeline struct {
	FilePath string
	Stages   ConfigPipelineStageList
	DryRun   bool
}

type Config struct {
	Owd string
	Dir string

	Unattended   bool
	Ci           bool
	Child        bool
	Parent       string
	ParentParams []string

	User     string
	Hostname string

	Verbosity int

	Pipeline ConfigPipeline
}
