package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/srevinsaju/togomak/v1/internal/rules"
	"github.com/zclconf/go-cty/cty"
	"testing"
)

func TestCanRun(t *testing.T) {
	ctx := context.Background()
	conductor := NewConductor(ConductorConfig{
		User:     "",
		Hostname: "",
		Paths: &path.Path{
			Pipeline: "",
			Owd:      "",
			Cwd:      "",
			Module:   "",
		},
		Interface: Interface{
			JSONLogging: true,
			Verbosity:   LifecycleInvalid,
		},
		Pipeline: ConfigPipeline{
			Filtered:    nil,
			FilterQuery: nil,
			DryRun:      false,
		},
		Behavior: behavior.NewDefaultBehavior(),
	})
	conductor.Update(ConductorWithContext(ctx))

	stage3 := Stage{
		Id: "stage2",
		CoreStage: CoreStage{
			DependsOn:   hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Condition:   hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Use:         nil,
			Daemon:      nil,
			Retry:       nil,
			Name:        "",
			Dir:         hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Script:      hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Shell:       hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Args:        hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Container:   nil,
			Environment: nil,
			PreHook:     nil,
			PostHook:    nil,
			ContainerId: "",
		},
	}

	stage1 := Stage{
		Id: "test",
		CoreStage: CoreStage{
			DependsOn: hcl.StaticExpr(cty.ListVal([]cty.Value{
				cty.StringVal(stage3.Identifier()),
			}), hcl.Range{}),
			Condition:   hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Use:         nil,
			Daemon:      nil,
			Retry:       nil,
			Name:        "",
			Dir:         hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Script:      hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Shell:       hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Args:        hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Container:   nil,
			Environment: nil,
			PreHook:     nil,
			PostHook:    nil,
			ContainerId: "",
		},
	}

	stage2 := Stage{
		Id: "test2",
		Lifecycle: &Lifecycle{
			Phase: hcl.StaticExpr(cty.ListVal([]cty.Value{
				cty.StringVal("build"),
			}), hcl.Range{}),
			Timeout: hcl.StaticExpr(cty.NumberIntVal(0), hcl.Range{}),
		},
		CoreStage: CoreStage{
			DependsOn:   hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Condition:   hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Use:         nil,
			Daemon:      nil,
			Retry:       nil,
			Name:        "",
			Dir:         hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Script:      hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Shell:       hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Args:        hcl.StaticExpr(cty.NilVal, hcl.Range{}),
			Container:   nil,
			Environment: nil,
			PreHook:     nil,
			PostHook:    nil,
			ContainerId: "",
		},
	}

	pipe := &Pipeline{
		Stages: []Stage{
			stage1,
			stage2,
			stage3,
		},
	}

	depGraph, d := GraphTopoSort(conductor, pipe)
	depGraph.DependOn(stage1.Identifier(), stage3.Identifier())
	if d.HasErrors() {
		t.Errorf("error while sorting: %s", d.Error())
		return
	}

	filtered, d := rules.Unmarshal([]string{stage1.Identifier()})
	if d.HasErrors() {
		t.Errorf("error while parsing rules: %s", d.Error())
		return
	}

	ok, overridden, err := BlockCanRun(&stage1, conductor, filtered, nil, stage1.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage1.Identifier())
		return
	}
	if !overridden {
		t.Errorf("%s should be overridden", stage1.Identifier())
		return
	}

	ok, overridden, err = BlockCanRun(&stage3, conductor, filtered, nil, stage3.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage3.Identifier())
		return
	}

	ok, overridden, err = BlockCanRun(&stage2, conductor, filtered, nil, stage2.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if ok {
		t.Errorf("%s should not be runnable", stage2.Identifier())
		return
	}

	ok, overridden, err = BlockCanRun(&stage2, conductor, nil, nil, stage2.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if ok {
		t.Errorf("%s should not be runnable", stage2.Identifier())
		return
	}

	ok, overridden, err = BlockCanRun(&stage1, conductor, nil, nil, stage1.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage1.Identifier())
		return
	}

	filtered, d = rules.Unmarshal([]string{"all"})
	if d.HasErrors() {
		t.Errorf("error while parsing rules: %s", d.Error())
		return
	}

	ok, overridden, err = BlockCanRun(&stage1, conductor, filtered, nil, stage1.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage1.Identifier())
		return
	}
	ok, overridden, err = BlockCanRun(&stage2, conductor, filtered, nil, stage2.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage2.Identifier())
		return
	}

	filtered, d = rules.Unmarshal([]string{"build"})
	if d.HasErrors() {
		t.Errorf("error while parsing rules: %s", d.Error())
		return
	}
	ok, overridden, err = BlockCanRun(&stage1, conductor, filtered, nil, stage1.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if ok {
		t.Errorf("%s should not be runnable", stage1.Identifier())
	}
	ok, overridden, err = BlockCanRun(&stage2, conductor, filtered, nil, stage2.Identifier(), depGraph)
	if err != nil {
		t.Errorf("error while running BlockCanRun: %s", err.Error())
		return
	}
	if !ok {
		t.Errorf("%s should be runnable", stage2.Identifier())
		return
	}

	//ok, overridden, err = BlockCanRun(&stage1, conductor, nil, nil, stage1.Identifier(), depGraph)
	//if err != nil {
	//	t.Errorf("error while running BlockCanRun: %s", err.Error())
	//	return
	//}
	//if !ok {
	//	t.Errorf("stage1 should be runnable")
	//	return
	//}
	//if !overridden {
	//	t.Errorf("stage1 should be overridden")
	//	return
	//}
}
