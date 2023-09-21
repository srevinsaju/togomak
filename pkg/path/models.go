package path

import "github.com/srevinsaju/togomak/v1/pkg/meta"

type Path struct {
	// Pipeline is the path to the pipeline file
	Pipeline string

	// Owd is the original working directory
	Owd string

	// Cwd is the current working directory
	Cwd string
}

func NewDefaultPath() Path {

	return Path{
		Pipeline: meta.ConfigFileName,
		Owd:      ".",
		Cwd:      ".",
	}
}
