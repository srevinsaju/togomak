package schema

import (
	"github.com/srevinsaju/buildsys/pkg/context"

	"github.com/hashicorp/go-plugin"
)

type Provider struct {
	Config ProviderConfig

	Client   *plugin.Client
	Provider Stage
	Context  *context.Context
}
