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

type ConfigPipelineStageList []ConfigPipelineStage

func (c ConfigPipelineStageList) Get(runnableId string) (ConfigPipelineStage, bool) {
	for _, stage := range c {
		if runnableId == fmt.Sprintf("stage.%s", stage.Id) {
			return stage, true
		}
	}
	return ConfigPipelineStage{}, false
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

	User     string
	Hostname string

	Verbosity int

	Pipeline ConfigPipeline
}
