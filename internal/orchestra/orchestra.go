package orchestra

import (
	"context"
	"github.com/srevinsaju/togomak/v1/internal/blocks"
	"github.com/srevinsaju/togomak/v1/internal/ci"
	"strings"

	"github.com/zclconf/go-cty/cty"
	"os"
)

func ExpandGlobalParams(conductor *ci.Conductor) {
	paramsGo := make(map[string]cty.Value)
	if conductor.Config.Behavior.Child.Enabled {
		m := make(map[string]string)
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				if strings.HasPrefix(e[:i], ci.TogomakParamEnvVarPrefix) {
					m[e[:i]] = e[i+1:]
				}
			}
		}
		for k, v := range m {
			if ci.TogomakParamEnvVarRegex.MatchString(k) {
				paramsGo[ci.TogomakParamEnvVarRegex.FindStringSubmatch(k)[1]] = cty.StringVal(v)
			}
		}
	}
	conductor.Eval().Mutex().Lock()
	conductor.Eval().Context().Variables[blocks.ParamBlock] = cty.ObjectVal(paramsGo)
	conductor.Eval().Mutex().Unlock()
}

func Perform(conductor *ci.Conductor) int {
	ctx, cancel := context.WithCancel(conductor.Context())
	defer cancel()
	conductor.Update(ci.ConductorWithContext(ctx))

	logger := conductor.Logger().WithField("orchestra", "perform")
	logger.Debugf("starting watchdogs and signal handlers")
	ExpandGlobalParams(conductor)

	// parse the config file
	pipe, hclDiags := ci.Read(conductor)
	if hclDiags.HasErrors() {
		logger.Fatal(conductor.DiagWriter.WriteDiagnostics(hclDiags))
	}

	h, d := pipe.Run(conductor)
	if d.HasErrors() {
		return h.Fatal()
	}
	return h.Ok()
}
