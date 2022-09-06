package main

import (
	"encoding/gob"
	log "github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/context"
	"os"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	if os.Getenv("TOGOMAK_DEBUG") != "" {
		log.SetLevel(log.TraceLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

}

func main() {
	gob.Register(context.Data{})

	app := initCli()
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
