package pithsdk

import (
	"fmt"
	"os"

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

func resolveAPIKey(cfg ClientConfig) (string, error) {
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key, nil
	}
	return "", fmt.Errorf("API key required: set ClientConfig.APIKey or OPENAI_API_KEY")
}

func setupDefaultGateway(apiKey string) *gateway.LLMGateway {
	gw := gateway.NewLLMGateway()
	gw.Providers.Register(providers.NewOpenAICompatProvider(providers.OpenAICompatConfig{
		BaseURL: defaultBaseURL,
	}))
	gw.Credentials = gateway.CredentialProviderFunc(func(pid protocol.ProviderId) (protocol.Credential, error) {
		if pid == defaultProviderID {
			return protocol.ApiKey{Key: apiKey}, nil
		}
		return nil, &protocol.Error{Code: protocol.ErrAuth, Message: fmt.Sprintf("no credentials for provider %q", pid)}
	})
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
