package ops

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/templating"
	"strings"
)

const ConfigMapType = "ConfigMap"

func RunConfigMapStore(cfg config.Config, stageCtx *context.Context, stage schema.StageConfig) error {
	// rootCtx := stageCtx.RootParent()
	if stage.Output.Data == nil {
		return nil
	}

	m := map[string]interface{}{}
	for k, v := range stage.Output.Data.(map[string]interface{}) {
		tpl, err := pongo2.FromString(v.(string))
		if err != nil {

			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}
		parsedV, err := templating.ExecuteWithStage(tpl, stageCtx.Data.AsMap(), stage)
		parsedV = strings.TrimSpace(parsedV)
		if err != nil {
			stageCtx.Logger.Warn(err)
			return fmt.Errorf("cannot render args '%s': %v", v, err)
		}
		m[k] = parsedV
	}
	stageCtx.DataMutex.Lock()
	stageCtx.Data[ConfigMapType] = m
	stageCtx.DataMutex.Unlock()

	return nil
}

func RunOutput(cfg config.Config, stageCtx *context.Context, stage schema.StageConfig) error {

	if stage.Output.Data == nil {

		stageCtx.Logger.Tracef("stage %s has no output defined", stage.Id)
		return nil
	}

	if stage.Output.Kind == ConfigMapType {

		return RunConfigMapStore(cfg, stageCtx, stage)
	}

	return nil

}
