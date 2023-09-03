package ci

func (s Stage) Override() bool {
	return false
}

func (s Stages) Override() bool {
	return false
}
