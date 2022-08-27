package context

import (
	"github.com/kendru/darwin/go/depgraph"
	log "github.com/sirupsen/logrus"
)

type Status struct {
	Success  bool
	Message  string
	MatrixId string
}

type Context struct {
	Logger   *log.Entry
	parent   *Context
	Graph    *depgraph.Graph
	IsMatrix bool

	status  Status
	TempDir string
	Data    map[string]interface{}
}

func (c *Context) SetStatus(s Status) {
	c.status = s
	c.Data["status"] = map[string]interface{}{
		"success":  s.Success,
		"message":  s.Message,
		"matrixId": s.MatrixId,
	}
}

func (c *Context) AddChild(k string, v string) *Context {
	if _, ok := c.Data[k]; !ok {
		c.Data[k] = map[string]interface{}{
			v: map[string]interface{}{},
		}
	} else {
		c.Data[k].(map[string]interface{})[v] = map[string]interface{}{}
	}
	return &Context{
		Logger:  c.Logger.WithField(k, v),
		parent:  c,
		TempDir: c.TempDir,
		Data:    c.Data[k].(map[string]interface{})[v].(map[string]interface{}),
	}
}

func (c *Context) RootParent() *Context {
	if c.parent == nil {
		return c
	}
	return c.parent.RootParent()
}

func (c *Context) Parent() *Context {
	return c.parent
}
