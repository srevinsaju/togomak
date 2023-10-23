package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/meta"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"path/filepath"
)

func (s *stageConductor) processEnvironment() *stageConductor {
	logger := s.Logger().WithField("stage", s.Id)
	envStrings := make([]string, len(s.attr.environment))
	envCounter := 0
	for k, v := range s.attr.environment {
		envParsed := fmt.Sprintf("%s=%s", k, v.AsString())
		if s.cfg.Behavior.DryRun {
			fmt.Println(ui.Blue("export"), envParsed)
		}

		envStrings[envCounter] = envParsed
		envCounter = envCounter + 1
	}
	togomakEnvExport := fmt.Sprintf("%s=%s", meta.OutputEnvVar, filepath.Join(global.TempDir(), meta.OutputEnvFile))
	logger.Tracef("exporting %s", togomakEnvExport)
	envStrings = append(envStrings, togomakEnvExport)

	if s.Use != nil && s.Use.Parameters != nil {
		for k, v := range s.params {
			envParsed := fmt.Sprintf("%s%s=%s", TogomakParamEnvVarPrefix, k, v.AsString())
			if s.cfg.Behavior.DryRun {
				fmt.Println(ui.Blue("export"), envParsed)
			}

			envStrings = append(envStrings, envParsed)
		}
	}
	s.attr.environmentStrings = envStrings
	return s
}

func (s *stageConductor) runUsingShell() *stageConductor {
	logger := s.Logger()
	s.process = s.attr.cmd
	logger.Tracef("running command: %.30s...", s.attr.cmd.String())
	if !s.cfg.Behavior.DryRun {
		err := s.attr.cmd.Run()

		if err != nil && err.Error() == "signal: terminated" && s.Terminated() {
			logger.Warnf("command terminated with signal: %s", s.attr.cmd.ProcessState.String())
			err = nil
		}
		if err != nil {
			s.diags = s.diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("failed to run command (%s)", s.Identifier()),
				Detail:   err.Error(),
			})
		}
	} else {
		fmt.Println(s.attr.cmd.String())
	}
	return s
}
