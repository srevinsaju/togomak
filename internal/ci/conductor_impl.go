package ci

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/conductor"
)

// Eval is a wrapper around hcl.EvalContext
// Returns the HCL evaluation context
func (c *Conductor) Eval() conductor.Eval {
	return c.eval
}

// StdinLock locks the stdin mutex
// Lock this mutex before reading from stdin
func (c *Conductor) StdinLock() {
	c.stdinMu.Lock()
}

// StdinUnlock unlocks the stdin mutex
// Unlock this mutex after reading from stdin
func (c *Conductor) StdinUnlock() {
	c.stdinMu.Unlock()
}

// Child creates a child conductor.
// A child conductor inherits the parent's configuration, logger, context and variables
// The child conductor can be used to run a sub-pipeline.
func (c *Conductor) Child(opts ...ConductorOption) *Conductor {
	inheritOpts := []ConductorOption{
		ConductorWithConfig(c.Config),
	}
	opts = append(inheritOpts, opts...)
	child := NewConductor(c.Config, opts...)
	child.parent = c
	return child
}

// Parent returns the parent conductor
func (c *Conductor) Parent() *Conductor {
	return c.parent
}

// RootParent returns the root parent conductor
func (c *Conductor) RootParent() *Conductor {
	if c.parent == nil {
		return c
	}
	return c.parent.RootParent()
}

// Logger returns the logger
func (c *Conductor) Logger() logrus.Ext1FieldLogger {
	return c.RootLogger
}

// Context returns the context.Context
func (c *Conductor) Context() context.Context {
	return c.ctx
}

// Variables returns the variables
func (c *Conductor) Variables() Variables {
	return c.variables
}
