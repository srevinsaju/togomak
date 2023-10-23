package ci

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/zclconf/go-cty/cty"
	"os"
	"os/exec"
	"path/filepath"
)

// parseEnvironment parses the environment variables for the stage
func (s *stageConductor) parseEnvironment() *stageConductor {
	logger := s.Logger()

	logger.Debug("evaluating environment variables")
	var environment map[string]cty.Value
	environment = make(map[string]cty.Value)
	for _, env := range s.Environment {
		s.Eval().Mutex().RLock()
		v, d := env.Value.Value(s.Eval().Context())
		s.Eval().Mutex().RUnlock()

		s.diags = s.diags.Extend(d)
		if v.IsNull() {
			s.diags = s.diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid environment variable",
				Detail:      fmt.Sprintf("environment variable %s is null", env.Name),
				EvalContext: s.Eval().Context(),
				Subject:     env.Value.Range().Ptr(),
			})
		} else if v.Type() != cty.String {
			s.diags = s.diags.Append(&hcl.Diagnostic{
				Severity:    hcl.DiagError,
				Summary:     "invalid environment variable",
				Detail:      fmt.Sprintf("environment variable %s is not a string", env.Name),
				EvalContext: s.Eval().Context(),
				Subject:     env.Value.Range().Ptr(),
			})
		} else {
			environment[env.Name] = v
		}
	}
	s.attr.environment = environment
	return s
}

func (s *stageConductor) parseExecCommand() *stageConductor {
	var diags hcl.Diagnostics
	logger := s.Logger()

	logger.Trace("evaluating script value")
	s.Eval().Mutex().RLock()
	script, d := s.Script.Value(s.Eval().Context())
	s.Eval().Mutex().RUnlock()

	if d.HasErrors() && s.cfg.Behavior.DryRun {
		script = cty.StringVal(ui.Italic(ui.Yellow("(will be evaluated later)")))
	} else {
		diags = diags.Extend(d)
	}

	logger.Trace("evaluating shell value")
	s.Eval().Mutex().RLock()
	shellRaw, d := s.Shell.Value(s.Eval().Context())
	s.Eval().Mutex().RUnlock()

	shell := ""
	if d.HasErrors() {
		diags = diags.Extend(d)
	} else {
		if shellRaw.IsNull() {
			shell = "bash"
		} else {
			shell = shellRaw.AsString()
		}
	}

	logger.Trace("evaluating args value")
	s.Eval().Mutex().RLock()
	args, d := s.Args.Value(s.Eval().Context())
	s.Eval().Mutex().RUnlock()
	diags = diags.Extend(d)

	cmdHcl, d := s.parseCommand(s.Eval().Context(), shell, script, args)
	diags = diags.Extend(d)
	if diags.HasErrors() {
		s.diags = s.diags.Extend(diags)
		return s
	}

	dir := s.cfg.Paths.Cwd

	s.Eval().Mutex().RLock()
	dirParsed, d := s.Dir.Value(s.Eval().Context())
	s.Eval().Mutex().RUnlock()

	if d.HasErrors() {
		diags = diags.Extend(d)
	} else {
		if !dirParsed.IsNull() && dirParsed.AsString() != "" {
			dir = dirParsed.AsString()
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(s.cfg.Paths.Cwd, dir)
		}
		if s.cfg.Behavior.DryRun {
			fmt.Println(ui.Blue("cd"), dir)
		}
	}

	cmd := exec.CommandContext(s.conductor.Context(), cmdHcl.command, cmdHcl.args...)
	cmd.Stdout = logger.Writer()
	cmd.Stderr = logger.WriterLevel(logrus.WarnLevel)
	cmd.Env = append(os.Environ(), s.attr.environmentStrings...)
	cmd.Dir = dir

	s.attr.cmd = cmd

	logger.Trace("command parsed")
	logger.Tracef("script: %.30s... ", cmd.String())

	s.diags = s.diags.Extend(diags)
	return s
}
