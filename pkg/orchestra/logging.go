package orchestra

import (
	"github.com/sirupsen/logrus"
	"os"
)

func NewLogger(cfg Config) *logrus.Logger {
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
	if cfg.Ci {
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
	return logger
}