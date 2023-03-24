package schema

type OutputConfigMetadata struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Name        string            `yaml:"name,omitempty"`
}

type OutputConfig struct {
	Data     any                  `yaml:"data,omitempty"`
	Kind     string               `yaml:"kind"`
	Metadata OutputConfigMetadata `yaml:"metadata,omitempty"`
}
