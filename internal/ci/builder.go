package ci

const BuilderBlock = "togomak"

type Behavior struct {
	DisableConcurrency bool `hcl:"disable_concurrency,optional" json:"disable_concurrency"`
}

type Builder struct {
	Version  int       `hcl:"version" json:"version"`
	Behavior *Behavior `hcl:"behavior,block" json:"behavior"`
}
