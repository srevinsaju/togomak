package meta

var (
	AppVersion = "v1.x"
)

const (
	AppName = "togomak"

	AppDescription = "A simple, declarative, and reproducible CI/CD pipeline generator powered by HCL"

	ConfigFileName = "togomak.hcl"
	BuildDirPrefix = ".togomak"

	EnvVarPrefix = "TOGOMAK__"

	OutputEnvFile = ".togomak.env"
	OutputEnvVar  = "TOGOMAK_OUTPUTS"

	RootStage = "togomak.root"
	PreStage  = "togomak.pre"
	PostStage = "togomak.post"
)
