package ci

import (
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/dg"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/srevinsaju/togomak/v1/internal/runnable"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"github.com/zclconf/go-cty/cty"
	"path/filepath"
	"sync"
)

func (m *Module) Prepare(conductor *Conductor, skip bool, overridden bool) hcl.Diagnostics {
	logger := conductor.Logger().WithField("module", m.Id)
	// show some user-friendly output on the details of the stage about to be run

	var id string
	if !skip {
		id = ""
	} else {
		id = fmt.Sprintf("%s", ui.Grey("skipped"))
	}
	if overridden {
		id = fmt.Sprintf("%s", ui.Blue("overridden"))
	}
	logger.Infof("%s", id)
	return nil
}

func (m *Module) Run(conductor *Conductor, options ...runnable.Option) (diags hcl.Diagnostics) {
	cfg := runnable.NewConfig(options...)
	evalCtx := conductor.Eval().Context()
	evalCtx = evalCtx.NewChild()

	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"id":     cty.StringVal(m.Id),
			"status": cty.StringVal(string(cfg.Status.Status)),
		}),
	}

	conductor.Eval().Mutex().RLock()
	source, d := m.Source.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()
	if d.HasErrors() {
		return diags.Extend(d)
	}
	if source.Type() != cty.String {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "source must be a string",
			Detail:      fmt.Sprintf("source must be a string, got %s", source.Type().FriendlyName()),
			Subject:     m.Source.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}
	src := source.AsString()

	if m.ForEach == nil {
		d = m.run(conductor, src, evalCtx, options...)
		diags = diags.Extend(d)
		return diags
	}

	conductor.Eval().Mutex().RLock()
	forEachItems, d := m.ForEach.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()

	diags = diags.Extend(d)
	if d.HasErrors() {
		return diags
	}

	if forEachItems.IsNull() {
		d = m.run(conductor, src, evalCtx, options...)
		diags = diags.Extend(d)
		return diags
	}

	if !forEachItems.CanIterateElements() {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "invalid type for for_each",
			Detail:   fmt.Sprintf("for_each must be a set or map of objects"),
		})
		return diags
	}

	var wg sync.WaitGroup

	var safeDg dg.SafeDiagnostics

	var counter int
	forEachItems.ForEachElement(func(k cty.Value, v cty.Value) bool {
		var key string
		var keyCty cty.Value
		if k.Type() == cty.String {
			key = fmt.Sprintf("\"%s\"", k.AsString())
			keyCty = k
		} else {
			key = fmt.Sprintf("%d", counter)
			keyCty = cty.NumberIntVal(int64(counter))
		}
		counter++
		id := fmt.Sprintf("%s[%s]", m.Id, key)
		wg.Add(1)
		module := &Module{
			Id:        id,
			DependsOn: nil,
			Condition: m.Condition,
			ForEach:   nil,
			Source:    m.Source,
			pipeline:  m.pipeline,
			Lifecycle: m.Lifecycle,
			Retry:     m.Retry,
			Daemon:    m.Daemon,
			Body:      m.Body,
		}
		go func(keyCty cty.Value, options ...runnable.Option) {
			options = append(options, runnable.WithEach(keyCty, v))
			d := module.Run(conductor, options...)
			safeDg.Extend(d)
			wg.Done()
		}(keyCty, options...)
		return false
	})
	wg.Wait()
	diags = diags.Extend(safeDg.Diagnostics())
	return diags
}

func (m *Module) run(conductor *Conductor, source string, evalCtx *hcl.EvalContext, options ...runnable.Option) hcl.Diagnostics {
	var diags hcl.Diagnostics
	logger := conductor.Logger().WithField("module", m.Id)
	cfg := runnable.NewConfig(options...)

	paths := cfg.Paths
	get := &getter.Client{
		Ctx: conductor.Context(),
		Src: source,
		Dst: filepath.Join(conductor.TempDir(), "modules", m.Id),
		Pwd: paths.Module,
		Dir: true,
	}
	err := get.Get()
	if err != nil {
		return diags.Append(&hcl.Diagnostic{
			Severity:    hcl.DiagError,
			Summary:     "failed to download source",
			Detail:      err.Error(),
			Subject:     m.Source.Range().Ptr(),
			EvalContext: evalCtx,
		})
	}

	b := &behavior.Behavior{
		Unattended: conductor.Config.Behavior.Unattended,
		Ci:         conductor.Config.Behavior.Ci,
		Child: behavior.Child{
			Enabled:      true,
			Parent:       "",
			ParentParams: nil,
		},
		DryRun: false,
	}
	childCfg := ConductorConfig{
		User:     conductor.Config.User,
		Hostname: conductor.Config.Hostname,
		Paths: &path.Path{
			Pipeline: filepath.Join(get.Dst, meta.ConfigFileName),
			Owd:      conductor.Config.Paths.Owd,
			Cwd:      conductor.Config.Paths.Cwd,
			Module:   get.Dst,
		},
		Interface: conductor.Config.Interface,
		Pipeline:  conductor.Config.Pipeline,
		Behavior:  b,
	}
	childConductor := conductor.Child(ConductorWithConfig(childCfg))
	var conductorOptions []ConductorOption

	// send the host conductor's parser to the module conductor
	// this will make hcl.Diagnostics more descriptive
	conductorOptions = append(conductorOptions, ConductorWithParser(conductor.Parser))

	// populate input variables for the child conductor, which would be passed to the module
	attrs, d := m.Body.JustAttributes()
	diags = diags.Extend(d)
	for _, attr := range attrs {
		//we need to evaluate the values first within the parent's evaluation context
		//before sending it to the child goroutine and child conductor
		//because the child evaluation context is independent of the parent's, and it is
		//possible that the particular value may not exist in child.
		var expr hcl.Expression
		v, d := attr.Expr.Value(evalCtx)
		if d.HasErrors() {
			expr = attr.Expr
		} else {
			expr = hcl.StaticExpr(v, attr.Expr.Range())
		}
		variable := &Variable{
			Id:    attr.Name,
			Value: expr,
		}
		conductorOptions = append(conductorOptions, ConductorWithVariable(variable))
	}

	// update the child conductor's logger with the parent's logger
	conductorOptions = append(conductorOptions, ConductorWithLogger(logger))

	childConductor.Update(conductorOptions...)

	// parse the config file
	pipe, hclDiags := Read(childConductor)
	if hclDiags.HasErrors() {
		logger.Fatal(childConductor.DiagWriter.WriteDiagnostics(hclDiags))
	}

	// update the child conductor's eval context with the each.key and each.value
	// if the cfg includes the option
	evalCtx = childConductor.Eval().Context().NewChild()
	evalCtx.Variables = map[string]cty.Value{}
	if cfg.Each != nil {
		evalCtx.Variables[EachBlock] = cty.ObjectVal(cfg.Each)
	}
	childConductor.Update(ConductorWithEvalContext(evalCtx))
	//  safe diagnostics
	_, sd := pipe.Run(childConductor)

	diags = diags.Extend(sd.Diagnostics())
	return diags
}

func (m *Module) CanRun(conductor *Conductor, options ...runnable.Option) (ok bool, diags hcl.Diagnostics) {
	logger := conductor.Logger().WithField("module", m.Id)
	logger.Debugf("checking if %s can run", x.RenderBlock(blocks.ModuleBlock, m.Id))
	evalCtx := conductor.Eval().Context()
	evalCtx = evalCtx.NewChild()

	cfg := runnable.NewConfig(options...)

	evalCtx.Variables = map[string]cty.Value{
		"this": cty.ObjectVal(map[string]cty.Value{
			"id":     cty.StringVal(m.Id),
			"status": cty.StringVal(string(cfg.Status.Status)),
		}),
	}

	// determine the value of the condition 'if', return truthiness of the value
	conductor.Eval().Mutex().RLock()
	v, d := m.Condition.Value(evalCtx)
	conductor.Eval().Mutex().RUnlock()
	if d.HasErrors() {
		return false, diags.Extend(d)
	}

	if v.Equals(cty.False).True() {
		// this stage has been explicitly evaluated to false
		// we will not run this
		return false, diags
	}

	return true, diags
}

func (m *Module) Terminated() bool {
	return true
}
