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
		},
	}

	return app
}
