package main

import (
	"github.com/urfave/cli/v2"
)

func initCli() *cli.App {
	app := &cli.App{
		Name:   "togomak",
		Usage:  "A customizable, powerful CI/CD which works anywhere",
		Action: cliContextRunner,
		Flags: []cli.Flag{
			&cli.PathFlag{
				Name:     "file",
				Required: false,
				Usage:    "The CI file which needs to run",
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
