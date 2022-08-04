package context

import (
	log "github.com/sirupsen/logrus"
)

type Context struct {
	Logger *log.Entry
	parent *Context
	TempDir string
}

func (c *Context) AddChild(k string, v string) *Context {
	return &Context{
		Logger: c.Logger.WithField(k, v),
		parent: c,
	}
}

func (c *Context) Parent() *Context {
	return c.parent
}
