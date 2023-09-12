package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"github.com/zclconf/go-cty/cty"
)

func (l *Local) Run(ctx context.Context, options ...runnable.Option) (diags hcl.Diagnostics) {

	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField(LocalBlock, l.Key)
	logger.Debugf("running %s.%s", LocalBlock, l.Key)
	hclContext := ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	global.LocalBlockEvalContextMutex.Lock()

	global.EvalContextMutex.RLock()
	locals := hclContext.Variables[LocalBlock]
	var localMutated map[string]cty.Value
	if locals.IsNull() {
		localMutated = make(map[string]cty.Value)
	} else {
		localMutated = locals.AsValueMap()
	}
	v, d := l.Value.Value(hclContext)
	global.EvalContextMutex.RUnlock()

	diags = diags.Extend(d)
	localMutated[l.Key] = v

	global.EvalContextMutex.Lock()
	hclContext.Variables[LocalBlock] = cty.ObjectVal(localMutated)
	global.EvalContextMutex.Unlock()

	global.LocalBlockEvalContextMutex.Unlock()

	// endregion

	return diags
}

func (l *Local) CanRun(ctx context.Context, options ...runnable.Option) (bool, hcl.Diagnostics) {
	return true, nil
}

func (l *Local) Prepare(ctx context.Context, skip bool, overridden bool) hcl.Diagnostics {
	return nil
}

func (l *Local) Terminated() bool {
	return true
}
