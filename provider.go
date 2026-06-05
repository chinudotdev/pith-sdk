package pithsdk

import (
	"fmt"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
)

const (
	defaultPresetContextWindow = 128_000
	defaultPresetMaxTokens     = 4096
)

// ProviderRegistration registers a custom LLM provider and its models on the client.
type ProviderRegistration struct {
	Provider gateway.ProviderPort

	// Credentials — one of:
	APIKey     string
	APIKeyEnv  string
	Credential func(providerID string) (string, error)

	Models []ModelPreset
}

// ModelPreset describes a model to register in the gateway catalog.
type ModelPreset struct {
	ID            string
	Name          string
	BaseURL       string // optional; provider may use its own default
	ContextWindow int    // 0 = SDK default
	MaxTokens     int    // 0 = SDK default
}

// RegisterProvider registers a custom provider, its models, and credentials.
func (c *Client) RegisterProvider(reg ProviderRegistration) error {
	if reg.Provider == nil {
		return fmt.Errorf("provider is required")
	}
	if len(reg.Models) == 0 {
		return fmt.Errorf("at least one model preset is required")
	}

	providerID := protocol.ProviderId(reg.Provider.Name())
	if c.registered[providerID] {
		return fmt.Errorf("provider %q is already registered", providerID)
	}

	resolver, err := credentialResolverFor(reg, string(providerID))
	if err != nil {
		return err
	}

	c.gw.Providers.Register(reg.Provider)
	for _, preset := range reg.Models {
		if preset.ID == "" {
			return fmt.Errorf("model preset ID is required")
		}
		c.gw.Catalog.Register(providerID, presetToDescriptor(reg.Provider, providerID, preset))
	}

	c.credentials.set(providerID, resolver)
	c.registered[providerID] = true
	c.syncCredentials()
	return nil
}

func credentialResolverFor(reg ProviderRegistration, providerID string) (credentialResolver, error) {
	switch {
	case reg.APIKey != "":
		return apiKeyResolver(reg.APIKey), nil
	case reg.APIKeyEnv != "":
		return apiKeyEnvResolver(reg.APIKeyEnv), nil
	case reg.Credential != nil:
		return credentialFuncResolver(reg.Credential, providerID), nil
	default:
		return nil, fmt.Errorf("credentials required for provider %q: set APIKey, APIKeyEnv, or Credential", providerID)
	}
}

func presetToDescriptor(provider gateway.ProviderPort, providerID protocol.ProviderId, preset ModelPreset) protocol.ModelDescriptor {
	name := preset.Name
	if name == "" {
		name = preset.ID
	}

	contextWindow := preset.ContextWindow
	if contextWindow == 0 {
		contextWindow = defaultPresetContextWindow
	}

	maxTokens := preset.MaxTokens
	if maxTokens == 0 {
		maxTokens = defaultPresetMaxTokens
	}

	return protocol.ModelDescriptor{
		ID:            preset.ID,
		Name:          name,
		API:           provider.API(),
		Provider:      providerID,
		BaseURL:       preset.BaseURL,
		ContextWindow: contextWindow,
		MaxTokens:     maxTokens,
		Capabilities: protocol.ModelCapabilities{
			Input:     map[protocol.MediaType]bool{protocol.MediaText: true},
			Transport: map[protocol.Transport]bool{protocol.TransportSSE: true},
		},
	}
}
