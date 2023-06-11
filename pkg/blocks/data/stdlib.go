package data

import "fmt"

var DefaultProviders = Providers{
	&EnvProvider{},
	&PromptProvider{},
	// FileProvider{},
}

type Providers []Provider

func (p Providers) GoString() string {
	message := ""
	for _, provider := range p {
		message += fmt.Sprintf("%#s, ", provider.Name())
	}
	return message
}

func (p Providers) Get(name string) Provider {
	for _, provider := range p {
		if provider.Name() == name {
			return provider
		}
	}
	return nil
}
