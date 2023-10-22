package conductor

import (
	"github.com/hashicorp/hcl/v2"
	"sync"
)

type Conductor interface {
	Eval() Eval
}

type Eval interface {
	Context() *hcl.EvalContext
	Mutex() *sync.RWMutex
}
