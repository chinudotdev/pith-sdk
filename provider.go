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
	// Provider implements gateway.ProviderPort for the custom API.
	Provider gateway.ProviderPort

	// Credentials — set exactly one of the following:
	// APIKey is a static API key string.
	APIKey string
	// APIKeyEnv names an environment variable holding the API key (e.g. "ANTHROPIC_API_KEY").
	APIKeyEnv string
	// Credential is a custom resolver returning the API key for the provider ID.
	Credential func(providerID string) (string, error)

	// Models lists model presets to register in the gateway catalog.
	Models []ModelPreset
}

// ModelPreset describes a model to register in the gateway catalog.
type ModelPreset struct {
	// ID is the model identifier used in provider/model strings.
	ID string
	// Name is an optional display name; defaults to ID.
	Name string
	// BaseURL is an optional API base URL; the provider may use its own default.
	BaseURL string
	// ContextWindow is the model context size; zero uses the SDK default.
	ContextWindow int
	// MaxTokens is the default max output tokens; zero uses the SDK default.
	MaxTokens int
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
