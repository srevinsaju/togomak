package graph

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/ci"
	"github.com/srevinsaju/togomak/v1/pkg/diag"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/x"
)

func TopoSort(ctx context.Context, pipe *ci.Pipeline) (*depgraph.Graph, diag.Diagnostics) {
	g := depgraph.New()
	var diags diag.Diagnostics
	logger := ctx.Value("logger").(*logrus.Logger).WithField("component", "graph")

	for _, local := range pipe.Local {
		err := g.DependOn(x.RenderBlock(ci.LocalBlock, local.Key), meta.RootStage)
		if err != nil {
			panic(err)
		}

		v := local.Variables()
		for _, variable := range v {
			blockType := variable.RootName()
			var child string
			var parent string

			child = x.RenderBlock(ci.LocalBlock, local.Key)
			_, d := ci.Resolve(ctx, pipe, child)
			diags = diags.Extend(d)

			switch blockType {
			case ci.DataBlock:
				// the data block has the provider type as well as the name
				provider := variable[1].(hcl.TraverseAttr).Name
				name := variable[2].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.DataBlock, provider, name)
			case ci.StageBlock:
				// the stage block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.StageBlock, name)
			case ci.LocalBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.LocalBlock, name)
			case ci.ThisBlock:
			case ci.MacroBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.MacroBlock, name)
			case ci.ParamBlock:
				continue
			case ci.BuilderBlock:
				continue
			default:
				continue
			}

			_, d = ci.Resolve(ctx, pipe, parent)
			diags = diags.Extend(d)
			err = g.DependOn(child, parent)

			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "Invalid dependency",
					Detail:   err.Error(),
				})
			}

		}

	}
	for _, macro := range pipe.Macros {
		err := g.DependOn(x.RenderBlock(ci.MacroBlock, macro.Id), meta.RootStage)
		// the addition of the root stage is to ensure that the macro block is always executed
		// before any stage
		// this function should succeed always
		if err != nil {
			panic(err)
		}

		v := macro.Variables()
		for _, variable := range v {
			blockType := variable.RootName()
			var child string
			var parent string
			child = x.RenderBlock(ci.MacroBlock, macro.Id)
			_, d := ci.Resolve(ctx, pipe, child)
			diags = diags.Extend(d)

			switch blockType {
			case ci.DataBlock:
				// the data block has the provider type as well as the name
				provider := variable[1].(hcl.TraverseAttr).Name
				name := variable[2].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.DataBlock, provider, name)
			case ci.StageBlock:
				// the stage block only has the id which is the second element
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.StageBlock, name)
			case ci.LocalBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.LocalBlock, name)
			case ci.MacroBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.MacroBlock, name)
			case ci.ThisBlock:
			case ci.ParamBlock:
				continue

			case ci.BuilderBlock:
				continue
			default:
				continue
			}

			_, d = ci.Resolve(ctx, pipe, parent)
			diags = diags.Extend(d)
			err = g.DependOn(child, parent)

			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "Invalid dependency",
					Detail:   err.Error(),
				})
			}
		}
	}
	for _, data := range pipe.Data {
		err := g.DependOn(x.RenderBlock(ci.DataBlock, data.Provider, data.Id), meta.RootStage)
		// the addition of the root stage is to ensure that the data block is always executed
		// before any stage
		// this function should succeed always
		if err != nil {
			panic(err)
		}

		v := data.Variables()
		for _, variable := range v {
			blockType := variable.RootName()
			var child string
			var parent string
			child = x.RenderBlock(ci.DataBlock, data.Provider, data.Id)
			_, d := ci.Resolve(ctx, pipe, child)
			diags = diags.Extend(d)

			switch blockType {
			case ci.DataBlock:
				// the data block has the provider type as well as the name
				provider := variable[1].(hcl.TraverseAttr).Name
				name := variable[2].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.DataBlock, provider, name)
			case ci.StageBlock:
				// the stage block only has the id which is the second element
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.StageBlock, name)
			case ci.LocalBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.LocalBlock, name)
			case ci.MacroBlock:
				// the local block has the name
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.MacroBlock, name)
			case ci.ThisBlock:
				continue
			case ci.ParamBlock:
				continue
			case ci.BuilderBlock:
				continue
			default:
				continue
			}

			_, d = ci.Resolve(ctx, pipe, parent)
			diags = diags.Extend(d)
			err = g.DependOn(child, parent)

			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "Invalid dependency",
					Detail:   err.Error(),
				})
			}
		}
	}
	for _, stage := range pipe.Stages {
		err := g.DependOn(x.RenderBlock(ci.StageBlock, stage.Id), meta.RootStage)
		if err != nil {
			panic(err)
		}

		v := stage.Variables()
		for _, variable := range v {

			blockType := variable.RootName()
			var child string
			var parent string
			child = x.RenderBlock(ci.StageBlock, stage.Id)
			_, d := ci.Resolve(ctx, pipe, child)
			diags = diags.Extend(d)

			switch blockType {
			case ci.DataBlock:
				// the data block has the provider type as well as the name
				provider := variable[1].(hcl.TraverseAttr).Name
				name := variable[2].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.DataBlock, provider, name)
			case ci.StageBlock:
				// the stage block only has the id which is the second element
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.StageBlock, name)
			case ci.LocalBlock:
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.LocalBlock, name)
			case ci.MacroBlock:
				name := variable[1].(hcl.TraverseAttr).Name
				parent = x.RenderBlock(ci.MacroBlock, name)
			case ci.ThisBlock:
				continue
			case ci.ParamBlock:
				continue
			case ci.BuilderBlock:
				continue
			default:
				continue
			}
			_, d = ci.Resolve(ctx, pipe, parent)
			diags = diags.Extend(d)
			err = g.DependOn(child, parent)

			if err != nil {
				diags = diags.Append(diag.Diagnostic{
					Severity: diag.SeverityError,
					Summary:  "Invalid dependency",
					Detail:   err.Error(),
				})
			}
		}
	}

	for i, layer := range g.TopoSortedLayers() {
		logger.Debugf("layer %d: %s", i, layer)
	}
	return g, diags

}
