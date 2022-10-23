package main

import (
	"fmt"
	"github.com/moby/sys/mountinfo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/runner"
	"github.com/urfave/cli/v2"
	"path"
	"path/filepath"
)

func autoDetectFile(cwd string) string {
	fs := afero.NewOsFs()

	absPath, err := filepath.Abs(cwd)
	if err != nil {
		panic(err)
	}
	mountPoint, err := mountinfo.Mounted(absPath)
	if mountPoint {
		log.Fatalf("Couldn't find togomak.yaml. Searched until %s", absPath)
	}

	p := path.Join(cwd, fmt.Sprintf("%s.yaml", meta.AppName))
	exists, err := afero.Exists(fs, p)
	if err != nil {
		log.Fatal(err)
	}

	if exists {
		return p
	} else {
		return autoDetectFile(path.Join("..", cwd))
	}
}

func cliContextRunner(cliCtx *cli.Context) error {

	if cliCtx.Bool("silent") {
		log.SetLevel(log.ErrorLevel)
	}
	var p string
	contextDir := cliCtx.Path("context")
	if cliCtx.Path("file") != "" {
		p = cliCtx.Path("file")
	} else {
		p = autoDetectFile(contextDir)
	}

	runner.Orchestrator(config.Config{
		RunStages:  cliCtx.Args().Slice(),
		ContextDir: contextDir,
		CiFile:     p,
		DryRun:     cliCtx.Bool("dry-run"),
		JobsNumber: cliCtx.Int("jobs"),
		FailLazy:   cliCtx.Bool("fail-lazy"),
		Summary:    config.GetSummaryType(cliCtx.String("summary")),
	})
	return nil
}
