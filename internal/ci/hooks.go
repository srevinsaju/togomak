package ci

func (s *PreStage) ToStage() *Stage {
	return &Stage{Id: "togomak.pre", CoreStage: s.CoreStage}
}

func (s *PostStage) ToStage() *Stage {
	return &Stage{Id: "togomak.post", CoreStage: s.CoreStage}
}
