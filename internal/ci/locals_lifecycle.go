package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (l *Local) ExecutionOptions(ctx context.Context) (*DaemonLifecycleConfig, hcl.Diagnostics) {
	return nil, nil
}
