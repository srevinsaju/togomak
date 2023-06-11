package orchestra

type ConfigPipeline struct {
	FilePath string
	Stages   []string
	DryRun   bool
}

type Config struct {
	Owd string
	Dir string

	User     string
	Hostname string

	Verbosity int

	Pipeline ConfigPipeline
}
