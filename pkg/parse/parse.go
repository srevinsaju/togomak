package parse

import (
	"context"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/meta"
	"path/filepath"
)

// ConfigFilePath returns the path to the configuration file. If the path is not absolute, it is assumed to be
// relative to the working directory
// DEPRECATED: use configFileDir instead
func ConfigFilePath(ctx context.Context) string {
	filePath := ctx.Value(c.TogomakContextPipelineFilePath).(string)
	if filePath == "" {
		filePath = meta.ConfigFileName
	}
	owd := ctx.Value(c.TogomakContextOwd).(string)

	if filepath.IsAbs(filePath) == false {
		filePath = filepath.Join(owd, filePath)
	}
	return filePath
}

func ConfigFileDir(ctx context.Context) string {
	return filepath.Dir(ConfigFilePath(ctx))
}
