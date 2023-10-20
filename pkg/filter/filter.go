package filter

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/blocks"
	"strings"
)

type OperationTemp string

const OperationTempRun OperationTemp = ""

type Item struct {
	Id        string
	Type      string
	Operation OperationTemp
}

func (c Item) RunnableId() string {

	return fmt.Sprintf("%s.%s", c.Type, c.Id)
}

func (c Item) Identifier() string {
	return c.RunnableId()
}

type FilterList []Item

func (c FilterList) Get(runnableId string) (FilterList, bool) {
	var stages []Item
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) {
			stages = append(stages, stage)
		}
	}
	return stages, len(stages) > 0
}

func (c Item) Child() Item {
	return Item{
		Id:        c.Id[strings.IndexRune(c.Id, '.')+1:],
		Operation: c.Operation,
		Type:      c.Type,
	}

}

func (c FilterList) Children(runnableId string) FilterList {
	var stages []Item
	for _, stage := range c {
		if strings.HasPrefix(stage.Identifier(), runnableId) && stage.Identifier() != runnableId {
			stages = append(stages, stage.Child())
		}
	}
	return stages
}

func (c FilterList) AllOperations(operation OperationTemp) bool {
	for _, stage := range c {
		if stage.Operation != operation {
			return false
		}
	}
	return true
}

func (c FilterList) AnyOperations(operation OperationTemp) bool {
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

func (c FilterList) HasOperationType(operation OperationTemp) bool {
	for _, stage := range c {
		if stage.Operation == operation {
			return true
		}
	}
	return false
}

func NewFilterItem(arg string) Item {
	var operation OperationTemp

	ty := blocks.StageBlock

	if strings.HasPrefix(arg, string(OperationWhitelist)) {
		operation = OperationWhitelist
	} else if strings.HasPrefix(arg, string(OperationBlacklist)) {
		operation = OperationBlacklist
	} else if strings.HasPrefix(arg, string(OperationDaemonize)) {
		operation = OperationDaemonize
	} else if strings.HasPrefix(arg, string(OperationRunLifecycle)) {
		operation = OperationRunLifecycle
		ty = blocks.LifecycleBlock
	} else {
		operation = OperationTempRun
	}

	// TODO: improve this
	if strings.Contains("module.", arg) {
		ty = blocks.ModuleBlock
	} else if strings.Contains("stage.", arg) {
		ty = blocks.StageBlock
	} else if strings.Contains("lifecycle.", arg) {
		ty = blocks.LifecycleBlock
	}

	id := strings.TrimPrefix(arg, string(operation))
	id = strings.TrimPrefix(id, ty+".")
	return Item{
		Id:        id,
		Type:      ty,
		Operation: operation,
	}
}
