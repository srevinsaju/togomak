package bootstrap

import "github.com/srevinsaju/togomak/pkg/config"

func contains(cfg config.Config, l string) bool {
	for _, s := range cfg.RunStages {
		if s == l {
			return true
		}
	}
	return false

}
