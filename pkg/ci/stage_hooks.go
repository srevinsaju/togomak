package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
)

func (s *Stage) BeforeRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics {
	if s.PreHook == nil {
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PreHook {
		diags = diags.Extend(hook.Stage.Run(ctx, opts...))
	}
	return diags
}

func (s *Stage) AfterRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics {
	if s.PostHook == nil {
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PostHook {
		diags = diags.Extend(hook.Stage.Run(ctx, opts...))
	}
	return diags
}
