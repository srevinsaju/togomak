package orchestra

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"strings"
)

type FilterOperation string

const (
	FilterOperationDaemonize FilterOperation = "&"
	FilterOperationBlacklist FilterOperation = "^"
	FilterOperationWhitelist FilterOperation = "+"
	FilterOperationRun       FilterOperation = ""
)

type FilterItem struct {
	Id        string
	Type      string
	Operation FilterOperation
}

func (c FilterItem) RunnableId() string {

	return fmt.Sprintf("%s.%s", c.Type, c.Id)
}

func (c FilterItem) Identifier() string {
	return c.RunnableId()
}

type FilterList []FilterItem

func (c FilterList) Get(runnableId string) (FilterList, bool) {
	var stages []FilterItem
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) {
			stages = append(stages, stage)
		}
	}
	return stages, len(stages) > 0
}

func (c FilterItem) Child() FilterItem {
	return FilterItem{
		Id:        c.Id[strings.IndexRune(c.Id, '.')+1:],
		Operation: c.Operation,
	}

}

func (c FilterList) Children(runnableId string) FilterList {
	var stages []FilterItem
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) && stage.Identifier() != runnableId {
			stages = append(stages, stage.Child())
		}
	}
	return stages
}

func (c FilterList) AllOperations(operation FilterOperation) bool {
	for _, stage := range c {
		if stage.Operation != operation {
			return false
		}
	}
	return true
}

func (c FilterList) AnyOperations(operation FilterOperation) bool {
	for _, stage := range c {
		if stage.Operation == operation {
			return true
		}
	}
	return false
}

func (c FilterList) Marshall() []string {
	var stages []string
	for _, stage := range c {
		stages = append(stages, string(stage.Operation)+stage.Type+"."+stage.Id)
	}
	return stages
}

func (c FilterList) HasOperationType(operation FilterOperation) bool {
	for _, stage := range c {
		if stage.Operation == operation {
			return true
		}
	}
	return false
}

func NewConfigPipelineStage(arg string) FilterItem {
	var operation FilterOperation

	ty := ci.StageBlock

	if strings.HasPrefix(arg, string(FilterOperationWhitelist)) {
		operation = FilterOperationWhitelist
	} else if strings.HasPrefix(arg, string(FilterOperationBlacklist)) {
		operation = FilterOperationBlacklist
	} else if strings.HasPrefix(arg, string(FilterOperationDaemonize)) {
		operation = FilterOperationDaemonize
	} else {
		operation = FilterOperationRun
	}

	// TODO: improve this
	if strings.Contains("module.", arg) {
		ty = ci.ModuleBlock
	} else if strings.Contains("stage.", arg) {
		ty = ci.StageBlock
	}

	id := strings.TrimPrefix(arg, string(operation))
	id = strings.TrimPrefix(id, ty+".")
	return FilterItem{
		Id:        id,
		Type:      ty,
		Operation: operation,
	}
}

type ConfigPipeline struct {
	FilePath string
	Filtered FilterList
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
