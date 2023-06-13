package main

import (
	"github.com/srevinsaju/togomak/v1/pkg/cache"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/orchestra"
	"github.com/urfave/cli/v2"
	"os"
)

var verboseCount = 0

func main() {
	app := cli.NewApp()
	app.Name = meta.AppName
	app.Description = meta.AppDescription
	app.Action = run

	app.Commands = []*cli.Command{
		{
			Name:    "init",
			Usage:   "initialize a new pipeline",
			Aliases: []string{"i"},
			Action:  initPipeline,
		},
		{
			Name:   "run",
			Usage:  "run a pipeline",
			Action: run,
		},
		{
			Name:    "list",
			Usage:   "list all the pipelines",
			Aliases: []string{"ls", "l"},
			Action:  list,
		},
		{
			Name:  "cache",
			Usage: "manage the cache",
			Subcommands: []*cli.Command{
				{
					Name:   "clean",
					Usage:  "clean the cache",
					Action: cleanCache,
				},
			},
		},
	}

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "file",
			Aliases: []string{"f"},
			Usage:   "path to the pipeline file",
		},
		&cli.StringFlag{
			Name:    "dir",
			Aliases: []string{"C", "directory"},
			Usage:   "path to the directory where the pipeline file is located",
		},
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
			Usage:   "enable verbose logging",
			Count:   &verboseCount,
		},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n", "just-print", "recon"},
			Usage:   "Don't actually run any stage; just print the commands",
			EnvVars: []string{"TOGOMAK_DRY_RUN"},
		},
	}
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "display the version of the application",
	}
	app.Version = meta.AppVersion
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

func initPipeline(ctx *cli.Context) error {
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	if dir == "" {
		dir = owd
	}
	orchestra.InitPipeline(dir)
	return nil
}

func newConfigFromCliContext(ctx *cli.Context) orchestra.Config {
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	pipelineFilePath := ctx.String("file")
	if pipelineFilePath == "" {
		dir := dir
		if dir == "" {
			dir = owd
		}
		pipelineFilePath = autoDetectFilePath(dir)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	cfg := orchestra.Config{
		Owd:       owd,
		Dir:       dir,
		User:      os.Getenv("USER"),
		Hostname:  hostname,
		Verbosity: verboseCount,
		Pipeline: orchestra.ConfigPipeline{
			FilePath: pipelineFilePath,
			DryRun:   ctx.Bool("dry-run"),
		},
	}
	return cfg
}

func run(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)

	orchestra.Orchestra(cfg)
	return nil
}

func cleanCache(ctx *cli.Context) error {
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	if dir == "" {
		dir = owd
	}
	cache.CleanCache(dir)
	return nil
}

func list(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	return orchestra.List(cfg)
}