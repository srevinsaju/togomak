package main

type RepositoryConfig struct {
	URL   string `mapstructure:"url"`
	Depth int    `mapstructure:"depth"`
}
type GitConfig struct {
	Repository      RepositoryConfig `mapstructure:"repo"`
	SkipTLSInsecure bool             `mapstructure:"skip_tls_insecure"`
	ReferenceName   string           `mapstructure:"reference_name"`
}
