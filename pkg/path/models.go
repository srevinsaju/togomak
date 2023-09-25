package path

type Path struct {
	// Pipeline is the path to the pipeline file
	Pipeline string

	// Owd is the original working directory
	Owd string

	// Cwd is the current working directory
	Cwd string
}
