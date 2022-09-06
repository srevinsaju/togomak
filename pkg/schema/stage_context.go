package schema

func (g *StageRPC) GetContext() Context {
	var resp Context
	err := g.client.Call("Plugin.GetContext", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (s *StageRPCServer) GetContext(args interface{}, resp *Context) error {
	*resp = s.Impl.GetContext()
	return nil
}

func (g *StageRPC) SetContext(d Context) error {
	var resp error
	err := g.client.Call("Plugin.SetContext", d, &resp)

	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
	return resp
}

func (s *StageRPCServer) SetContext(context Context, resp *error) error {
	*resp = s.Impl.SetContext(context)
	return nil
}
