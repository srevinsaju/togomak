package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/zclconf/go-cty/cty"
	"sync"
)

func (l *Local) Run(ctx context.Context) hcl.Diagnostics {

	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(LocalBlock, l.Key)
	logger.Debugf("running %s.%s", LocalBlock, l.Key)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	muLocals := ctx.Value(c.TogomakContextMutexLocals).(*sync.Mutex)
	var diags hcl.Diagnostics

	// region: mutating the data map
	// TODO: move it to a dedicated helper function
	muLocals.Lock()
	locals := hclContext.Variables[LocalBlock]
	var localMutated map[string]cty.Value
	if locals.IsNull() {
		localMutated = make(map[string]cty.Value)
	} else {
		localMutated = locals.AsValueMap()
	}
	v, d := l.Value.Value(hclContext)
	diags = diags.Extend(d)
	localMutated[l.Key] = v
	hclContext.Variables[LocalBlock] = cty.ObjectVal(localMutated)
	muLocals.Unlock()
	// endregion

	return diags
}

func (l *Local) CanRun(ctx context.Context) (bool, hcl.Diagnostics) {
	return true, nil
}

func (l *Local) Prepare(ctx context.Context, skip bool, overridden bool) hcl.Diagnostics {
	return nil
}

func (l *Local) Terminated() bool {
	return true
}
