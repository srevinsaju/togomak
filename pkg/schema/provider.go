package schema

import (
	"github.com/srevinsaju/togomak/pkg/context"

	"github.com/hashicorp/go-plugin"
)

type Provider struct {
	Config ProviderConfig

	Client   *plugin.Client
	Provider Stage
	Context  *context.Context
}
