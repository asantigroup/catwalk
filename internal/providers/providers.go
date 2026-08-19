// Package providers provides a registry of inference providers
package providers

import (
	_ "embed"
	"encoding/json"
	"log"

	"charm.land/catwalk/pkg/catwalk"
)

//go:embed configs/asanti.json
var asantiConfig []byte

//go:embed configs/opencode-go.json
var openCodeGoConfig []byte

//go:embed configs/opencode-zen.json
var openCodeZenConfig []byte

//go:embed configs/openrouter.json
var openRouterConfig []byte

//go:embed configs/synthetic.json
var syntheticConfig []byte

//go:embed configs/zai.json
var zAIConfig []byte

//go:embed configs/zhipu-coding.json
var zhipuCodingConfig []byte

//go:embed configs/umans.json
var umansConfig []byte

// ProviderFunc is a function that returns a Provider.
type ProviderFunc func() catwalk.Provider

var providerRegistry = []ProviderFunc{
	asantiProvider,
	syntheticProvider,
	zAIProvider,
	zhipuCodingProvider,
	umansProvider,

	// The remaining will be in alphabetical order.
	openCodeGoProvider,
	openCodeZenProvider,
	openRouterProvider,
}

// GetAll returns all registered providers.
func GetAll() []catwalk.Provider {
	providers := make([]catwalk.Provider, 0, len(providerRegistry))
	for _, providerFunc := range providerRegistry {
		providers = append(providers, providerFunc())
	}
	return providers
}

func loadProviderFromConfig(configData []byte) catwalk.Provider {
	var p catwalk.Provider
	if err := json.Unmarshal(configData, &p); err != nil {
		log.Printf("Error loading provider config: %v", err)
		return catwalk.Provider{}
	}
	return p
}

func asantiProvider() catwalk.Provider {
	return loadProviderFromConfig(asantiConfig)
}

func openCodeGoProvider() catwalk.Provider {
	return loadProviderFromConfig(openCodeGoConfig)
}

func openCodeZenProvider() catwalk.Provider {
	return loadProviderFromConfig(openCodeZenConfig)
}

func openRouterProvider() catwalk.Provider {
	return loadProviderFromConfig(openRouterConfig)
}

func syntheticProvider() catwalk.Provider {
	return loadProviderFromConfig(syntheticConfig)
}

func zAIProvider() catwalk.Provider {
	return loadProviderFromConfig(zAIConfig)
}

func zhipuCodingProvider() catwalk.Provider {
	return loadProviderFromConfig(zhipuCodingConfig)
}

func umansProvider() catwalk.Provider {
	return loadProviderFromConfig(umansConfig)
}
