package ci

import (
	"github.com/hashicorp/hcl/v2"
	"sync"
)

// Eval is a wrapper around hcl.EvalContext
// It is used to store the HCL evaluation context and the mutex
// It includes a Mutex() method which returns a read-write mutex which can be used to lock
// whilst updating the evaluation context's variable map or the function map
type Eval struct {
	context *hcl.EvalContext
	mu      *sync.RWMutex
}

// Context returns the HCL evaluation context
func (e *Eval) Context() *hcl.EvalContext {
	return e.context
}

// Mutex returns the mutex associated with the evaluation context
func (e *Eval) Mutex() *sync.RWMutex {
	return e.mu
}
