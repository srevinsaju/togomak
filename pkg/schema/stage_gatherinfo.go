package schema

func (g *StageRPC) GatherInfo() error {
	var resp error
	err := g.client.Call("Plugin.GatherInfo", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (s *StageRPCServer) GatherInfo(args interface{}, resp *error) error {
	*resp = s.Impl.GatherInfo()
	return nil
}
