package ci

import (
	"context"
	"fmt"
)

const StageContextChildStatuses = "child_statuses"

func (s *Stage) Description() string {
	// TODO: implement
	return ""
}

func (s *Stage) Identifier() string {
	return fmt.Sprintf("%s.%s", StageBlock, s.Id)
}

func (s *Stage) Set(k any, v any) {
	if s.ctxInitialised == false {
		s.ctx = context.Background()
		s.ctxInitialised = true
	}
	s.ctx = context.WithValue(s.ctx, k, v)
}

func (s *Stage) Get(k any) any {
	if s.ctxInitialised {
		return s.ctx.Value(k)
	}
	return nil
}

func (s *Stage) Type() string {
	return StageBlock
}

func (s *Stage) IsDaemon() bool {
	return s.Daemon != nil && s.Daemon.Enabled
}

func (s Stages) ById(id string) (*Stage, error) {
	for _, stage := range s {
		if stage.Id == id {
			return &stage, nil
		}
	}
	return nil, fmt.Errorf("stage with id %s not found", id)
}
