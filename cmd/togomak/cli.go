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
		},
	}

	return app
}
