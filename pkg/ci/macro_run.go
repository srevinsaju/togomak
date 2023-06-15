package ci

import (
	"context"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
)

const (
	SourceTypeGit = "git"
)

func (m *Macro) Prepare(ctx context.Context, skip bool, overridden bool) diag.Diagnostics {
	if m.Source == "" {
		return nil
	}
	// TODO: implement source

	return nil // no-op
}

func (m *Macro) Run(ctx context.Context) diag.Diagnostics {
	return nil
}

func (m *Macro) CanRun(ctx context.Context) (bool, diag.Diagnostics) {
	return true, nil
}
