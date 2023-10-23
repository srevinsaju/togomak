package ci

import (
	"context"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/sirupsen/logrus"
)

type ConductorOption func(*Conductor)

func ConductorWithLogger(logger logrus.Ext1FieldLogger) ConductorOption {
	return func(c *Conductor) {
		c.RootLogger = logger
	}
}

func ConductorWithConfig(cfg ConductorConfig) ConductorOption {
	return func(c *Conductor) {
		c.Config = cfg
	}
}

func ConductorWithContext(ctx context.Context) ConductorOption {
	return func(c *Conductor) {
		c.ctx = ctx
	}
}

func ConductorWithParser(parser *hclparse.Parser) ConductorOption {
	return func(c *Conductor) {
		c.Parser = parser
	}
}

func ConductorWithDiagWriter(diagWriter hcl.DiagnosticWriter) ConductorOption {
	return func(c *Conductor) {
		c.DiagWriter = diagWriter
	}
}

func ConductorWithEvalContext(evalContext *hcl.EvalContext) ConductorOption {
	return func(c *Conductor) {
		c.eval.context = evalContext
	}
}

func ConductorWithProcess(process Process) ConductorOption {
	return func(c *Conductor) {
		c.Process = process
	}
}

func ConductorWithVariable(variable *Variable) ConductorOption {
	return func(c *Conductor) {
		c.variables = append(c.variables, variable)
	}
}

func ConductorWithVariablesList(variables Variables) ConductorOption {
	return func(c *Conductor) {
		c.variables = variables
	}
}
