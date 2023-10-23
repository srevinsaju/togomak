package ci

import (
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"os"
	"path/filepath"
)

func chdir(cfg ConductorConfig, logger *logrus.Logger) string {
	cwd := cfg.Paths.Cwd
	if cwd == "" {
		cwd = filepath.Dir(cfg.Paths.Pipeline)
		if filepath.Base(cwd) == meta.BuildDirPrefix {
			cwd = filepath.Dir(cwd)
		}
	}
	err := os.Chdir(cwd)
	if err != nil {
		logger.Fatal(err)
	}
	cwd, err = os.Getwd()
	x.Must(err)
	logger.Debug("changing working directory to ", cwd)
	return cwd
}
