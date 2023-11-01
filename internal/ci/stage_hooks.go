package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
)

func (s *Stage) BeforeRun(conductor *Conductor, opts ...runnable.Option) hcl.Diagnostics {
	logger := conductor.Logger().WithField("stage", s.Id)
	if s.PreHook == nil {
		logger.Debug("no pre-hook defined")
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PreHook {
		diags = diags.Extend(
			(&Stage{fmt.Sprintf("%s.pre", s.Id), nil, hook.Stage, nil}).Run(conductor, opts...),
		)
	}
	return diags
}

func (s *Stage) AfterRun(conductor *Conductor, opts ...runnable.Option) hcl.Diagnostics {
	logger := conductor.Logger().WithField("stage", s.Id)
	if s.PostHook == nil {
		logger.Debug("no post-hook defined")
		return nil
	}
	var diags hcl.Diagnostics

	for _, hook := range s.PostHook {
		diags = diags.Extend(
			(&Stage{fmt.Sprintf("%s.post", s.Id), nil, hook.Stage, nil}).Run(conductor, opts...),
		)
	}
	return diags
}
