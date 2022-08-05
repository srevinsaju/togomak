package schema

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type Context struct {
	Data map[string]string
}

// Stage is the interface that we're exposing as a plugin.
type Stage interface {
	Name() string
	Description() string
	//Version() string
	//Author() string

	//CanRun() bool
	//Run() error
	GatherInfo() error
	SetContext(context Context)
	GetContext() Context
}

// Here is an implementation that talks over RPC
type StageRPC struct{ client *rpc.Client }

// Here is the RPC server that StageRPC talks to, conforming to
// the requirements of net/rpc
type StageRPCServer struct {
	// This is the real implementation
	Impl Stage
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a StageRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return StageRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type StagePlugin struct {
	// Impl Injection
	Impl Stage
}

func (p *StagePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &StageRPCServer{Impl: p.Impl}, nil
}

func (StagePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &StageRPC{client: c}, nil
}
