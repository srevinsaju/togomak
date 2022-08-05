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

func (g *StageRPC) SetContext(context Context) {
	
	err := g.client.Call("Plugin.SetContext", context, nil)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}
}


func (s *StageRPCServer) SetContext(args interface{}, resp interface{}) error {
	s.Impl.SetContext(args.(Context))

	return nil
}
