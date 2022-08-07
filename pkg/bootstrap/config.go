package bootstrap

import (
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/meta"
	"github.com/srevinsaju/togomak/pkg/schema"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func Config(ctx *context.Context, cfg config.Config) schema.SchemaConfig {
	ctx.Logger.Debugf("Reading config file %s", cfg.CiFile)
	yfile, err := ioutil.ReadFile(cfg.CiFile)
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	data := schema.SchemaConfig{}
	err = yaml.Unmarshal(yfile, &data)
	if err != nil {
		ctx.Logger.Fatal(err)
	}

	ctx.Logger.Tracef("Checking version of %s config", meta.AppName)
	if data.Version != meta.SupportedCiConfigVersion {

		ctx.Logger.Fatal("Unsupported version on togomak config")
	}

	ctx.Logger.Tracef("loaded data: %v", data)
	return data
}
