package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
)

func (d Data) Lifecycle(ctx context.Context) (*DaemonLifecycle, hcl.Diagnostics) {
	return nil, nil
}
