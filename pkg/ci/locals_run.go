package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/pkg/global"
	"github.com/srevinsaju/togomak/v1/pkg/runnable"
	"github.com/zclconf/go-cty/cty"
)

func (l *Local) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {

	logger := l.Logger()
	logger.Debugf("running %s.%s", LocalBlock, l.Key)
	hclContext := global.HclEvalContext()

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

func (l *Local) CanRun(conductor *Conductor, options ...runnable.Option) (bool, hcl.Diagnostics) {
	return true, nil
}

func (l *Local) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	return nil
}

func (l *Local) Terminated() bool {
	return true
}
