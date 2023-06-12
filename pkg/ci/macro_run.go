package ci

import (
	"context"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

func (m Macro) Prepare(ctx context.Context, skip bool) {
	return // no-op
}

func (m Macro) Run(ctx context.Context) diag.Diagnostics {
	return nil
}

func (m Macro) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	return true, nil
}
