package graph

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/x"
)

func Resolve(ctx context.Context, pipe *ci.Pipeline, g *depgraph.Graph, v []hcl.Traversal, child string) hcl.Diagnostics {
	var diags hcl.Diagnostics

	_, d := ci.Resolve(ctx, pipe, child)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		return diags
	}

	for _, variable := range v {
		parent, d := ci.ResolveFromTraversal(variable)
		diags = diags.Extend(d)
		if parent == "" {
			continue
		}

		_, d = ci.Resolve(ctx, pipe, parent)
		diags = diags.Extend(d)
		err := g.DependOn(child, parent)

		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid dependency",
				Detail:   err.Error(),
			})
		}

	}
	return diags
}
func TopoSort(ctx context.Context, pipe *ci.Pipeline) (*depgraph.Graph, hcl.Diagnostics) {
	g := depgraph.New()
	var diags hcl.Diagnostics
	logger := ctx.Value("logger").(*logrus.Logger).WithField("component", "graph")

	x.Must(g.DependOn(meta.PreStage, meta.RootStage))
	x.Must(g.DependOn(meta.PostStage, meta.PreStage))

	for _, local := range pipe.Local {
		self := x.RenderBlock(ci.LocalBlock, local.Key)
		err := g.DependOn(self, meta.RootStage)
		if err != nil {
			panic(err)
		}

		v := local.Variables()
		d := Resolve(ctx, pipe, g, v, self)
		diags = diags.Extend(d)
	}

	for _, macro := range pipe.Macros {
		self := x.RenderBlock(ci.MacroBlock, macro.Id)
		err := g.DependOn(self, meta.RootStage)
		// the addition of the root stage is to ensure that the macro block is always executed
		// before any stage
		// this function should succeed always
		if err != nil {
			panic(err)
		}

		v := macro.Variables()
		d := Resolve(ctx, pipe, g, v, self)
		diags = diags.Extend(d)
	}
	for _, data := range pipe.Data {
		self := x.RenderBlock(ci.DataBlock, data.Provider, data.Id)
		err := g.DependOn(self, meta.RootStage)
		// the addition of the root stage is to ensure that the data block is always executed
		// before any stage
		// this function should succeed always
		if err != nil {
			panic(err)
		}

		// all pre-stage blocks depend on the data block
		err = g.DependOn(meta.PreStage, self)
		if err != nil {
			panic(err)
		}

		v := data.Variables()
		d := Resolve(ctx, pipe, g, v, self)
		diags = diags.Extend(d)
	}

	for _, stage := range pipe.Stages {
		self := x.RenderBlock(ci.StageBlock, stage.Id)
		err := g.DependOn(self, meta.PreStage)
		if err != nil {
			panic(err)
		}

		err = g.DependOn(meta.PostStage, self)
		if err != nil {
			panic(err)
		}

		v := stage.Variables()
		d := Resolve(ctx, pipe, g, v, self)
		diags = diags.Extend(d)
	}

	for _, module := range pipe.Modules {
		self := x.RenderBlock(ci.ModuleBlock, module.Id)
		err := g.DependOn(self, meta.PreStage)
		if err != nil {
			panic(err)
		}

		err = g.DependOn(meta.PostStage, self)
		if err != nil {
			panic(err)
		}

		v := module.Variables()
		d := Resolve(ctx, pipe, g, v, self)
		diags = diags.Extend(d)
	}

	for i, layer := range g.TopoSortedLayers() {
		logger.Debugf("layer %d: %s", i, layer)
	}
	return g, diags

}
