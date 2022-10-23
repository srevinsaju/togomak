package config

const (
	SummaryAuto = iota
	SummaryOn
	SummaryOff
)

func GetSummaryType(v string) int {
	switch v {
	case "on":
		return SummaryOn
	case "off":
		return SummaryOff
	default:
		return SummaryAuto
	}

}

type Config struct {
	RunAll        bool
	Force         bool
	RunStages     []string
	ContextDir    string
	CiFile        string
	DryRun        bool
	JobsNumber    int
	FailLazy      bool
	IsFailFastSet bool
	Summary       int
}
