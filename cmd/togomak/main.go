package main

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/mattn/go-isatty"
	"github.com/srevinsaju/togomak/v1/internal/behavior"
	"github.com/srevinsaju/togomak/v1/internal/cache"
	"github.com/srevinsaju/togomak/v1/internal/ci"
	"github.com/srevinsaju/togomak/v1/internal/filter"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/logging"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/orchestra"
	"github.com/srevinsaju/togomak/v1/internal/path"
	"github.com/srevinsaju/togomak/v1/internal/rules"
	"github.com/urfave/cli/v2"
	"os"
)

var verboseCount = 0

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	meta.AppVersion = version
	app := cli.NewApp()
	app.Name = meta.AppName
	app.Description = meta.AppDescription
	app.Action = run
	app.Version = fmt.Sprintf("%s (%s, %s)", version, commit, date)

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
			Name:   "fmt",
			Usage:  "format a pipeline file",
			Action: format,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "check",
					Usage:   "check if the file is formatted",
					Aliases: []string{"c"},
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Usage:   "format all the files named togomak.hcl in the current directory, and its children",
					Aliases: []string{"r"},
				},
			},
		},
		{
			Name:  "cache",
			Usage: "manage the cache",
			Subcommands: []*cli.Command{
				{
					Name:   "clean",
					Usage:  "clean the cache",
					Action: cleanCache,
					Flags: []cli.Flag{
						&cli.BoolFlag{
							Name:    "recursive",
							Usage:   "clean the cache recursively",
							Aliases: []string{"r"},
						},
					},
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
		&cli.BoolFlag{
			Name:   "child",
			Usage:  "run the pipeline as a child process (advanced)",
			Hidden: true,
		},
		&cli.StringFlag{
			Name:   "parent",
			Usage:  "the parent process id (advanced)",
			Hidden: true,
		},
		&cli.StringSliceFlag{
			Name:   "parent-param",
			Usage:  "parameter passed to child togomak process (advanced)",
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:    "unattended",
			Aliases: []string{"no-prompt", "no-interactive"},
			Usage:   "do not prompt for responses, or wait for user responses. run in auto-pilot",
			EnvVars: []string{"TOGOMAK_UNATTENDED"},
			Value:   !isatty.IsTerminal(os.Stdin.Fd()),
		},
		&cli.BoolFlag{
			Name:    "ci",
			Usage:   "run in CI mode",
			EnvVars: []string{"CI", "TOGOMAK_CI"},
			Value:   false,
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
			Name:    "disable-concurrency",
			Aliases: []string{"disable-parallel"},
			Usage:   "disable concurrency",
		},
		&cli.BoolFlag{Name: "json", Usage: "enable json logging", EnvVars: []string{"TOGOMAK_JSON_LOG"}},
		&cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"n", "just-print", "recon"},
			Usage:   "Don't actually run any stage; just print the commands",
			EnvVars: []string{"TOGOMAK_DRY_RUN"},
		},
		&cli.StringSliceFlag{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "filter the pipeline by a query",
		},
		&cli.StringSliceFlag{
			Name: "var",
			Usage: "set a variable in the pipeline. " + "The format is <key>=<value>. " +
				"Multiple variables can be set by passing the flag multiple times. " +
				"Variables set this way take precedence over variables set in the pipeline file.",
			Aliases: []string{"variable"},
		},
		&cli.BoolFlag{
			Name:    "logging.remote.google-cloud",
			Usage:   "Enable remote logging to Google Cloud",
			EnvVars: []string{"TOGOMAK_LOGGING_REMOTE_GCLOUD"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "logging.remote.google-cloud.project",
			Usage:   "Google Cloud project ID where logs are ingested",
			EnvVars: []string{"TOGOMAK_LOGGING_REMOTE_GCLOUD_PROJECT", "GOOGLE_CLOUD_PROJECT"},
		},
		&cli.BoolFlag{
			Name:    "logging.local.file",
			Usage:   "Enable local logging to a file",
			EnvVars: []string{"TOGOMAK_LOGGING_LOCAL_FILE"},
			Value:   false,
		},
		&cli.StringFlag{
			Name:    "logging.local.file.path",
			Usage:   "Path to the file where logs are written",
			EnvVars: []string{"TOGOMAK_LOGGING_LOCAL_FILE_PATH"},
			Value:   "togomak.log",
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

func newConfigFromCliContext(ctx *cli.Context) ci.ConductorConfig {
	var diags hcl.Diagnostics
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

	var stages []filter.Item
	for _, stage := range ctx.Args().Slice() {
		stages = append(stages, filter.NewFilterItem(stage))
	}

	diagWriter := hcl.NewDiagnosticTextWriter(os.Stdout, nil, 0, true)
	filterQueries := ctx.StringSlice("query")
	engines, d := ci.NewSlice(filterQueries)
	if d.HasErrors() {
		diagWriter.WriteDiagnostics(d)
		os.Exit(1)
	}
	filtered, d := rules.Unmarshal(ctx.Args().Slice())
	if d.HasErrors() {
		diagWriter.WriteDiagnostics(d)
		os.Exit(1)
	}
	var variables []*ci.Variable
	for _, v := range ctx.StringSlice("var") {
		shell, d := ci.ParseVariableShell(v)
		diags = diags.Extend(d)
		variables = append(variables, shell)
	}
	if diags.HasErrors() {
		diagWriter.WriteDiagnostics(diags)
		os.Exit(1)
	}

	cfg := ci.ConductorConfig{
		Behavior: &behavior.Behavior{
			Unattended:         ctx.Bool("unattended") || ctx.Bool("ci"),
			Ci:                 ctx.Bool("ci"),
			DryRun:             ctx.Bool("dry-run"),
			DisableConcurrency: ctx.Bool("disable-concurrency"),

			Child: behavior.Child{
				Enabled:      ctx.Bool("child"),
				Parent:       ctx.String("parent"),
				ParentParams: ctx.StringSlice("parent-param"),
			},
		},

		Paths: &path.Path{
			Pipeline: pipelineFilePath,
			Cwd:      dir,
			Owd:      owd,
			Module:   "",
		},
		User:      os.Getenv("USER"),
		Hostname:  hostname,
		Interface: ci.Interface{Verbosity: verboseCount, JSONLogging: ctx.Bool("json")},
		Pipeline: ci.ConfigPipeline{
			FilterQuery: engines,
			Filtered:    filtered,
			DryRun:      ctx.Bool("dry-run"),
		},
		Variables: variables,

		Logging: logging.Config{
			Verbosity:     verboseCount,
			Child:         ctx.Bool("child"),
			IsCI:          ctx.Bool("ci"),
			JSON:          ctx.Bool("json"),
			CorrelationID: "",
			Sinks:         logging.ParseSinksFromCLI(ctx),
		},
	}
	return cfg
}

func run(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	logger, err := logging.New(cfg.Logging)
	if err != nil {
		panic(err)
	}

	global.SetLogger(logger)

	t := ci.NewConductor(cfg)
	v := orchestra.Perform(t)
	t.Destroy()
	os.Exit(v)
	return nil
}

func cleanCache(ctx *cli.Context) error {
	recursive := ctx.Bool("recursive")
	owd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dir := ctx.String("dir")
	if dir == "" {
		dir = owd
	}
	cache.CleanCache(dir, recursive)
	return nil
}

func list(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	return orchestra.List(cfg)
}

func format(ctx *cli.Context) error {
	cfg := newConfigFromCliContext(ctx)
	return orchestra.Format(cfg, ctx.Bool("check"), ctx.Bool("recursive"))
}
