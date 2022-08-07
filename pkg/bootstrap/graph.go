package bootstrap

import (
	"github.com/kendru/darwin/go/depgraph"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
)

func Graph(ctx *context.Context, data schema.SchemaConfig) {

	graphLog := ctx.Logger.WithField("context", "graph")
	rootStage := schema.NewRootStage()
	// generate the dependency graph with topological sort
	graph := depgraph.New()
	for _, stage := range data.Stages {
		if len(stage.DependsOn) == 0 {
			// no depends on
			graphLog.Debugf("%s stage depends on %s stage", stage.Id, rootStage.Id)
			err := graph.DependOn(stage.Id, rootStage.Id)
			if err != nil {
				ctx.Logger.Warn("Error while creating the dependency tree", err)
			}
		}
		for _, dep := range stage.DependsOn {
			graphLog.Debugf("%s stage depends on %s stage", dep, stage.Id)
			err := graph.DependOn(stage.Id, dep)
			if err != nil {
				ctx.Logger.Warn("Error while creating the dependency tree", err)
			}

		}
	}
	ctx.Graph = graph

}
