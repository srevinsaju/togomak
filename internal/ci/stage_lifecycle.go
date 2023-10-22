package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (s *Stage) ExecutionOptions(ctx context.Context) (*DaemonLifecycleConfig, hcl.Diagnostics) {
	if s.Daemon != nil {
		return s.Daemon.Lifecycle.Parse(ctx)
	}
	return nil, nil
}

func (s *Stage) LifecycleConfig() *Lifecycle {
	return s.Lifecycle
}
