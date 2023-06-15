package ci

import (
	"fmt"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

func (s *Stage) Description() string {
	// TODO: implement
	return ""
}

func (s *Stage) Identifier() string {
	return fmt.Sprintf("%s.%s", StageBlock, s.Id)
}

func (s *Stage) Type() string {
	return StageBlock
}

func (s *Stage) IsDaemon() bool {
	return s.Daemon != nil && s.Daemon.Enabled
}

func (s *Stage) Terminate() diag.Diagnostics {
	return nil
}

func (s *Stage) Kill() diag.Diagnostics {
	return nil
}

func (s Stages) ById(id string) (*Stage, error) {
	for _, stage := range s {
		if stage.Id == id {
			return &stage, nil
		}
	}
	return nil, fmt.Errorf("stage with id %s not found", id)
}
