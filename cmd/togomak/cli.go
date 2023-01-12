package main

import (
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/urfave/cli/v2"
)

func initCli() *cli.App {
	app := &cli.App{
		Name:                 meta.AppName,
		Usage:                "A customizable, powerful CI/CD which works anywhere",
		Version:              meta.Version,
		Action:               cliContextRunner,
		EnableBashCompletion: true,

		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List all the available stages",
				Action:  cliListStages,
			},
		},
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "file",
				Required: false,
				Usage:    "The CI file which needs to run",
			},

			&cli.PathFlag{
				Name:     "context",
				Required: false,
				Usage:    "The context directory where the togomak needs to run",
				Aliases:  []string{"C"},
			},

			&cli.BoolFlag{
				Name:     "force",
				Aliases:  []string{"f"},
				Usage:    "Do not check if target is up-to-date, build unconditionally.",
				Required: false,
			},

			&cli.BoolFlag{
				Name:     "debug",
				Required: false,
				Usage:    "Enable debug mode",
				EnvVars:  []string{"TOGOMAK_DEBUG"},
			},

			&cli.StringFlag{
				Name:        "color",
				Required:    false,
				Usage:       "Configure logging colored output",
				EnvVars:     []string{"TOGOMAK_COLOR"},
				DefaultText: "auto",
			},

			&cli.BoolFlag{
				Name:     "ci",
				Required: false,
				Usage:    "Enable CI mode",
				EnvVars:  []string{"TOGOMAK_CI"},
			},

			&cli.BoolFlag{
				Name:     "no-interactive",
				Required: false,
				Usage:    "Disable interactive mode",
				EnvVars:  []string{"TOGOMAK_NO_INTERACTIVE"},
			},

			&cli.BoolFlag{
				Name:    "silent",
				Aliases: []string{"s"},
				Usage:   "Disable All output from togomak",
				EnvVars: []string{"TOGOMAK_SILENT"},
			},

			&cli.BoolFlag{
				Name:     "dry-run",
				Required: false,
				Usage:    "Don't actually run any stage; just print the commands",
				EnvVars:  []string{"TOGOMAK_DRY_RUN"},
				Aliases:  []string{"n", "just-print", "recon"},
			},

			&cli.IntFlag{
				Name:        "jobs",
				Usage:       "Number of parallel jobs to run",
				Aliases:     []string{"j"},
				Value:       -1,
				DefaultText: "Default, -1 implies infinity",
			},

			&cli.BoolFlag{
				Name:     "fail-lazy",
				Required: false,
				Usage:    "Do not fail fast",
				Value:    false,
			},

			&cli.StringFlag{
				Name:    "summary",
				Usage:   "Don't print summary",
				EnvVars: []string{"TOGOMAK_SUMMARY"},
				Value:   "auto",
			},

			&cli.StringSliceFlag{
				Name:     "parameters",
				Required: false,
				Usage:    "Parameters to be passed to the CI file",
				Aliases:  []string{"p", "e"},
			},

			&cli.BoolFlag{
				Name:     "child",
				Required: false,
				Usage:    "Run as a child process",
			},
		},
	}

	return app
}
