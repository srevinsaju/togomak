package bootstrap

import (
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"os"
	"path/filepath"
)

func Chdir(ctx *context.Context, cfg config.Config, data schema.SchemaConfig) {
	if !data.Options.Chdir && cfg.ContextDir == "" {
		// change working directory to the directory of the config file
		cwd := filepath.Dir(cfg.CiFile)
		ctx.Logger.Debugf("Changing directory to %s", cwd)
		err := os.Chdir(cwd)
		if err != nil {
			ctx.Logger.Warn(err)
		}
		ctx.DataMutex.Lock()
		ctx.Data["cwd"] = cwd
		ctx.DataMutex.Unlock()
	} else {
		ctx.Logger.Debugf("Changing directory to %s", cfg.ContextDir)
		err := os.Chdir(cfg.ContextDir)
		if err != nil {
			ctx.Logger.Warn(err)
		}
		ctx.DataMutex.Lock()
		ctx.Data["cwd"] = cfg.ContextDir
		ctx.DataMutex.Unlock()
	}
}
