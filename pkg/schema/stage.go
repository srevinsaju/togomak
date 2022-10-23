package schema

import (
	"github.com/hashicorp/go-plugin"
	"github.com/srevinsaju/togomak/pkg/context"
	"net/rpc"
)

type Context struct {
	Data context.Data
}

// Stage is the interface that we're exposing as a plugin.
type Stage interface {
	Name() string
	Description() string
	//Version() string
	//Author() string

	//CanRun() bool
	//Run() error
	GatherInfo() StageError
	SetContext(c Context) error
	GetContext() Context
}

type StageError struct {
	Err   string
	IsErr bool
}

func StageErrorFromErr(err error) StageError {
	if err != nil {
		return StageError{
			Err:   err.Error(),
			IsErr: true,
		}
	}
	return StageError{Err: "", IsErr: err != nil}
}

// Here is an/var/mnt/data/repo/github.com/srevinsaju/togomak/.togomak/plugins/togomak-provider-git implementation that talks over RPC
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
