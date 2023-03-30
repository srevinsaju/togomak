package bootstrap

import (
	"fmt"
	"github.com/kendru/darwin/go/depgraph"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"regexp"
	"strings"
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
				ctx.Logger.Fatalf("Error while creating the dependency tree for '%s': %s", stage.Id, err)
			}
		}
		for _, dep := range stage.DependsOn {
			graphLog.Debugf("r'%s' stage depends on %s stage", dep, stage.Id)
			found := false

			if strings.HasPrefix(dep, "/") && strings.HasSuffix(dep, "/") {

				for _, s := range data.Stages {
					v, err := regexp.MatchString(fmt.Sprintf("%s$", strings.Trim(dep, "/")), s.Id)
					if err != nil {
						panic(err)
					}
					if v {
						found = true
						// the stage exists
						graphLog.Debugf("%s stage depends on %s stage", s.Id, stage.Id)
						if s.Id == stage.Id {
							// skip if the stage depends on itself
							continue
						}
						err := graph.DependOn(stage.Id, s.Id)
						if err != nil {
							ctx.Logger.Fatalf("Error while creating the dependency tree for '%s -> %s': %s", stage.Id, s.Id, err)
						}
					}
				}
			} else {
				for _, s := range data.Stages {
					if s.Id == dep {
						found = true
						// the stage exists
						graphLog.Debugf("%s stage depends on %s stage", s.Id, stage.Id)
						if s.Id == stage.Id {
							// skip if the stage depends on itself
							continue
						}
						err := graph.DependOn(stage.Id, s.Id)
						if err != nil {
							ctx.Logger.Fatalf("Error while creating the dependency tree for '%s -> %s': %s", stage.Id, s.Id, err)
						}
					}
				}
			}
			if !found {
				ctx.Logger.Fatalf("Error while creating the dependency tree for '%s -> %s': %s", stage.Id, dep, "stage not found")
			}

		}
	}
	ctx.Graph = graph

}
