package logging

import (
	"errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"os"
)

type Sink struct {
	Name  string
	Level logrus.Level

	Options map[string]string
}

type Config struct {
	Verbosity     int
	Child         bool
	IsCI          bool
	JSON          bool
	CorrelationID string

	Sinks []Sink
}

func ParseSinksFromCLI(ctx *cli.Context) []Sink {
	var sinks []Sink
	file := ctx.Bool("logging.local.file")
	if file {
		sinks = append(sinks, Sink{
			Name:  "file",
			Level: logrus.DebugLevel,
			Options: map[string]string{
				"path": ctx.String("logging.local.file.path"),
			},
		})
	}
	gcloud := ctx.Bool("logging.remote.google-cloud")
	if gcloud {
		sinks = append(sinks, Sink{
			Name:  "google-cloud",
			Level: logrus.DebugLevel,
			Options: map[string]string{
				"project": ctx.String("logging.remote.google-cloud.project"),
			},
		})
	}
	return sinks
}

func New(cfg Config) (*logrus.Logger, error) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:    false,
		DisableTimestamp: cfg.Child,
	})
	switch cfg.Verbosity {
	case -1:
	case 0:
		logger.SetLevel(logrus.InfoLevel)
		break
	case 1:
		logger.SetLevel(logrus.DebugLevel)
		break
	default:
		logger.SetLevel(logrus.TraceLevel)
		break
	}
	if cfg.IsCI {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	if cfg.Child {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:          true,
			DisableColors:             false,
			EnvironmentOverrideColors: false,
			ForceColors:               true,
			ForceQuote:                false,
		})
	}
	if cfg.JSON {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	for _, sink := range cfg.Sinks {
		switch sink.Name {
		case "file":
			path, ok := sink.Options["path"]
			if !ok {
				path = "togomak.log"
			}
			hook := lfshook.NewHook(path, &logrus.JSONFormatter{})
			logger.AddHook(hook)
		case "google-cloud":
			project, ok := sink.Options["project"]
			if !ok {
				return nil, errors.New("google-cloud sink requires project option")
			}
			hook, err := NewGoogleCloudLoggerHook(cfg, project)
			if err != nil {
				return nil, err
			}
			logger.AddHook(hook)
		default:
			return nil, errors.New("unknown sink: " + sink.Name)
		}
	}

	return logger, nil
}
