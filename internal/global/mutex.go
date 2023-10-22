package global

import "sync"

var (
	DataBlockEvalContextMutex     = sync.Mutex{}
	VariableBlockEvalContextMutex = sync.Mutex{}
	MacroBlockEvalContextMutex    = sync.Mutex{}
	LocalBlockEvalContextMutex    = sync.Mutex{}
)
