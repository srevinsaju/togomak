package ci

import (
	"fmt"
)

func (s Stage) Description() string {
	// TODO: implement
	return ""
}

func (s Stage) Identifier() string {
	return fmt.Sprintf("%s.%s", StageBlock, s.Id)
}

func (s Stages) ById(id string) (*Stage, error) {
	for _, stage := range s {
		if stage.Id == id {
			return &stage, nil
		}
	}
	return nil, fmt.Errorf("stage with id %s not found", id)
}

func (s Stage) Type() string {
	return StageBlock
}
