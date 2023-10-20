package global

import "sync"

var (
	EvalContextMutex = sync.RWMutex{}

	DataBlockEvalContextMutex  = sync.Mutex{}
	MacroBlockEvalContextMutex = sync.Mutex{}
	LocalBlockEvalContextMutex = sync.Mutex{}
)
