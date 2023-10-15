package ci

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
)

func (s *Stage) BeforeRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics {
	if s.PreHook == nil {
		s.Logger().Debug("no pre-hook defined")
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PreHook {
		diags = diags.Extend(
			(&Stage{fmt.Sprintf("%s.pre", s.Id), nil, hook.Stage, nil}).Run(ctx, opts...),
		)
	}
	return diags
}

func (s *Stage) AfterRun(ctx context.Context, opts ...runnable.Option) hcl.Diagnostics {
	if s.PostHook == nil {
		s.Logger().Debug("no post-hook defined")
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PostHook {
		diags = diags.Extend(
			(&Stage{fmt.Sprintf("%s.pre", s.Id), nil, hook.Stage, nil}).Run(ctx, opts...),
		)
	}
	return diags
}
