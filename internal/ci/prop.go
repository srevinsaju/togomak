package ci

type Overrideable interface {
	Override() bool
}

type Distinct interface {
	Overrideable
	Describable
}
