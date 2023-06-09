package config

const (
	SummaryAuto = iota
	SummaryOn
	SummaryOff
)

func GetSummaryType(v string) int {
	switch v {
	case "on", "yes":
		return SummaryOn
	case "off", "no":
		return SummaryOff
	default:
		return SummaryAuto
	}

}

type Config struct {
	RunAll        bool
	Color         string
	Force         bool
	NoInteractive bool
	RunStages     []string
	ContextDir    string
	CiFile        string
	DryRun        bool
	JobsNumber    int
	FailLazy      bool
	Parameters    []string
	IsFailFastSet bool
	Summary       int
	Loggers       []string
	Backend       string
	TrackingServer string
}
