package parse

import (
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"github.com/srevinsaju/togomak/v1/pkg/path"
	"path/filepath"
)

// ConfigFilePath returns the path to the configuration file. If the path is not absolute, it is assumed to be
// relative to the working directory
// DEPRECATED: use configFileDir instead
func ConfigFilePath(paths path.Path) string {
	pipelineFilePath := paths.Pipeline
	if pipelineFilePath == "" {
		pipelineFilePath = meta.ConfigFileName
	}

	if filepath.IsAbs(pipelineFilePath) == false {
		pipelineFilePath = filepath.Join(paths.Owd, pipelineFilePath)
	}
	return pipelineFilePath
}

func ConfigFileDir(paths path.Path) string {
	return filepath.Dir(ConfigFilePath(paths))
}
