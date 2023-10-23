package conductor

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/sirupsen/logrus"
	"sync"
)

type Conductor interface {
	Eval() Eval
	Logger() logrus.Ext1FieldLogger
	TempDir() string
}

type Eval interface {
	Context() *hcl.EvalContext
	Mutex() *sync.RWMutex
}
