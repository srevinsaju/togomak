package behavior

type Child struct {
	// Enabled is the flag to indicate whether the program is running in child mode
	Enabled bool

	// Parent is the flag to indicate whether the program is running in parent mode
	Parent string

	// ParentParams is the list of parameters to be passed to the parent
	ParentParams []string
}

type Behavior struct {
	initialized bool

	// Unattended is the flag to indicate whether the program is running in unattended mode
	Unattended bool

	// Ci is the flag to indicate whether the program is running in CI mode
	Ci bool

	// Child is the flag to indicate whether the program is running in child mode
	Child Child

	DryRun bool
}

func NewDefaultBehavior() *Behavior {
	return &Behavior{
		Unattended: false,
		Ci:         false,
		DryRun:     true,
		Child: Child{
			Enabled:      false,
			Parent:       "",
			ParentParams: []string{},
		},
	}
}
