package schema

func (g *StageRPC) Description() string {
	var resp string
	err := g.client.Call("Plugin.Description", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (s *StageRPCServer) Description(args interface{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}
