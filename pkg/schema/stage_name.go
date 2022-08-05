package schema

func (g *StageRPC) Name() string {
	var resp string
	err := g.client.Call("Plugin.Name", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (s *StageRPCServer) Name(args interface{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}
