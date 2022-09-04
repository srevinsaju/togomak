package context

import (
	"github.com/kendru/darwin/go/depgraph"
	log "github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/pkg/meta"
	"os"
	"strings"
)

type Status struct {
	Success  bool
	Message  string
	MatrixId string
}

type Data map[string]interface{}

func (d Data) GetString(key string) string {
	return d[key].(string)
}
func (d Data) GetBool(key string) bool {
	return d[key].(bool)
}

func (d Data) GetMap(key string) Data {
	if v, ok := d[key]; ok {
		return v.(Data)
	}
	d[key] = Data{}
	return d[key].(Data)
}

func (d Data) AsMap() map[string]interface{} {
	return d
}

func (d Data) GetList(key string) []string {
	return d[key].([]string)
}

type Context struct {
	Logger       *log.Entry
	Key          string
	Value        string
	parent       *Context
	envVarPrefix string
	Graph        *depgraph.Graph
	IsMatrix     bool

	status  Status
	TempDir string
	Data    Data
}

func (c *Context) SetStatus(s Status) {
	c.status = s
	c.Data[KeyStatus] = Data{
		StatusSuccess:  s.Success,
		StatusMessage:  s.Message,
		StatusMatrixId: s.MatrixId,
	}
}

func (c *Context) Getenv(k string) string {
	v := c.envVarPrefix
	if v == "" {
		s := strings.Builder{}
		s.WriteString(meta.EnvPrefix)
		s.WriteString("__")

		childCtx := c
		var kArr []string
		for childCtx != nil {
			if childCtx.Key != "" {
				kArr = append(kArr, childCtx.Key)
			}
			childCtx = childCtx.Parent()
		}
		for i := len(kArr) - 1; i >= 0; i-- {
			s.WriteString(kArr[i])
			s.WriteString("__")
		}
		v = s.String()
		// cache it
		c.envVarPrefix = v
	}
	c.Logger.Tracef("Reading environment variable %s", v+k)
	return os.Getenv(v)
}

func (c *Context) GetenvWithDefault(k string, defaultValue string) string {
	v := c.Getenv(k)
	if v == "" {
		return defaultValue
	}
	return v
}

func (c *Context) AddChild(k string, v string) *Context {
	if v != "" {
		if _, ok := c.Data[k]; !ok {
			c.Data[k] = Data{
				v: Data{},
			}
		} else {
			// key exists
			// we need to check if v exists too
			if _, ok := c.Data.GetMap(k)[v]; !ok {
				c.Data.GetMap(k)[v] = Data{}
			}
		}
	}
	var logger *log.Entry
	var data Data
	if v == "" {
		logger = c.Logger.WithField("ctx", k)
		data = c.Data.GetMap(k)
	} else {
		logger = c.Logger.WithField(k, v)
		data = c.Data.GetMap(k).GetMap(v)
	}

	return &Context{
		Logger:  logger,
		Key:     k,
		Value:   v,
		parent:  c,
		TempDir: c.TempDir,
		Data:    data,
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
