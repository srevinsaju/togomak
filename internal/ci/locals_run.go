package ci

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/zclconf/go-cty/cty"
)

func (l *Local) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {

	logger := l.Logger()
	logger.Debugf("running %s.%s", LocalBlock, l.Key)
	evalContext := conductor.Eval().Context()

	// region: mutating the data map
	// TODO: move it to a dedicated helper function

	global.LocalBlockEvalContextMutex.Lock()

	conductor.Eval().Mutex().RLock()
	locals := evalContext.Variables[LocalBlock]
	var localMutated map[string]cty.Value
	if locals.IsNull() {
		localMutated = make(map[string]cty.Value)
	} else {
		localMutated = locals.AsValueMap()
	}
	v, d := l.Value.Value(evalContext)
	conductor.Eval().Mutex().RUnlock()

	diags = diags.Extend(d)
	localMutated[l.Key] = v

	conductor.Eval().Mutex().Lock()
	evalContext.Variables[LocalBlock] = cty.ObjectVal(localMutated)
	conductor.Eval().Mutex().Unlock()

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
