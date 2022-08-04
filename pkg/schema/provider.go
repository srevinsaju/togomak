package schema

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Provider is the interface that we're exposing as a plugin.
type Provider interface {
	Name() string
}

// Here is an implementation that talks over RPC
type ProviderRPC struct{ client *rpc.Client }

func (g *ProviderRPC) Name() string {
	var resp string
	err := g.client.Call("Plugin.Name", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

// Here is the RPC server that ProviderRPC talks to, conforming to
// the requirements of net/rpc
type ProviderRPCServer struct {
	// This is the real implementation
	Impl Provider
}

func (s *ProviderRPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a ProviderRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return ProviderRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type ProviderPlugin struct {
	// Impl Injection
	Impl Provider
}

func (p *ProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ProviderRPCServer{Impl: p.Impl}, nil
}

func (ProviderPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ProviderRPC{client: c}, nil
}
