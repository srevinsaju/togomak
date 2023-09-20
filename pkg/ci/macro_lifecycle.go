package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (m *Macro) Lifecycle(ctx context.Context) (*DaemonLifecycleConfig, hcl.Diagnostics) {
	return nil, nil
}
