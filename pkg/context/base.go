package context

import (
	log "github.com/sirupsen/logrus"
	
)

type Context struct {
	Logger    *log.Entry
	parent    *Context
	
	TempDir   string
	Data      map[string]interface{}
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
		Logger: c.Logger.WithField(k, v),
		parent: c,
		TempDir: c.TempDir,
		Data: c.Data[k].(map[string]interface{})[v].(map[string]interface{}),
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
