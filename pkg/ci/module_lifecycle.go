package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (m *Module) Lifecycle(ctx context.Context) (*DaemonLifecycle, hcl.Diagnostics) {
	if m.Daemon != nil {
		return m.Daemon.Lifecycle.Parse(ctx)
	}
	return nil, nil
}
