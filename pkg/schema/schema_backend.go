package schema

// BackendConfigType is the type of the backend
// this is an enum containing local and cloudbuild for now
type BackendConfigType string

const (
	// BackendConfigTypeLocal is the local backend
	BackendConfigTypeLocal      BackendConfigType = "local"
	BackendConfigTypeCloudBuild BackendConfigType = "cloudbuild"
)

// BackendConfig is the configuration for the backend
type BackendConfig struct {
	// Type specifies the backend type
	Type BackendConfigType `yaml:"type,omitempty"`

	// CloudBuild specifies the configuration for the cloudbuild backend
	CloudBuild map[string]interface{} `yaml:"cloudbuild,omitempty"`
	Local      map[string]interface{} `yaml:"local,omitempty"`
}
