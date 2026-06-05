package pithsdk

import (
	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/gateway/providers"
	"github.com/chinudotdev/pith/protocol"
)

const (
	defaultProviderID = "openai"
	defaultModelID    = "gpt-4o-mini"
	defaultBaseURL    = "https://api.openai.com"
)

func applyClientDefaults(cfg *ClientConfig) {
	if cfg.DefaultProvider == "" {
		cfg.DefaultProvider = defaultProviderID
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = defaultModelID
	}
}

func resolveAPIKey(cfg ClientConfig) string {
	if cfg.APIKey != "" {
		return cfg.APIKey
	}
	return ""
}

func setupDefaultGateway(creds credentialRegistry) *gateway.LLMGateway {
	gw := gateway.NewLLMGateway()
	gw.Providers.Register(providers.NewOpenAICompatProvider(providers.OpenAICompatConfig{
		BaseURL: defaultBaseURL,
	}))
	gw.Credentials = creds.credentialProvider()
	gw.Catalog.Register(defaultProviderID, defaultGPT4oMiniDescriptor())
	return gw
}

func defaultGPT4oMiniDescriptor() protocol.ModelDescriptor {
	return protocol.ModelDescriptor{
		ID:            defaultModelID,
		Name:          "GPT-4o Mini",
		API:           protocol.ApiOpenAICompletions,
		Provider:      defaultProviderID,
		BaseURL:       defaultBaseURL,
		ContextWindow: 128000,
		MaxTokens:     4096,
		Capabilities: protocol.ModelCapabilities{
			Input:     map[protocol.MediaType]bool{protocol.MediaText: true},
			Transport: map[protocol.Transport]bool{protocol.TransportSSE: true},
		},
	}
}
