package bootstrap

import (
	"github.com/schwarmco/go-cartesian-product"
	"github.com/srevinsaju/togomak/pkg/config"
	"github.com/srevinsaju/togomak/pkg/context"
	"github.com/srevinsaju/togomak/pkg/schema"
	"github.com/srevinsaju/togomak/pkg/ui"
)

func MatrixRun(ctx *context.Context, data schema.SchemaConfig, cfg config.Config) {

	matrixLogger := ctx.Logger
	var keys []string
	var s [][]interface{}
	for k, v := range data.Matrix {
		keys = append(keys, k)
		var ss []interface{}
		for _, vv := range v {
			ss = append(ss, vv)
		}
		s = append(s, ss)
	}

	matrixText := ui.Grey("matrix")
	for product := range cartesian.Iter(s...) {

		matrixLogger.Infof("[%s] %s %s build", ui.Plus, ui.SubStage, ui.Matrix)

		ctx.Data["matrix"] = map[string]string{}
		for i := range keys {
			matrixLogger.Infof("%s %s %s.%s=%s", ui.SubStage, ui.SubSubStage, matrixText, ui.Grey(keys[i]), product[i])
			ctx.Data["matrix"].(map[string]string)[keys[i]] = product[i].(string)
		}

		SimpleRun(ctx, cfg, data)
	}

}
