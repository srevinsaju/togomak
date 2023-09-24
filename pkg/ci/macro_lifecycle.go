package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (m *Macro) ExecutionOptions(ctx context.Context) (*DaemonLifecycleConfig, hcl.Diagnostics) {
	return nil, nil
}
