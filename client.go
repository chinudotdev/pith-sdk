package pithsdk

import (
	"fmt"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith-sdk/internal/resolve"
	"github.com/chinudotdev/pith-sdk/internal/wire"
)

// Client holds gateway configuration, credentials, and defaults.
type Client struct {
	gw              *gateway.LLMGateway
	defaultProvider string
	defaultModel    string
	defaultSettings *ModelSettings
}

// ClientConfig configures a new Client.
type ClientConfig struct {
	APIKey          string
	DefaultProvider string
	DefaultModel    string
	DefaultSettings *ModelSettings
}

// NewClient creates a client with a built-in OpenAI-compatible provider.
func NewClient(cfg ClientConfig) (*Client, error) {
	applyClientDefaults(&cfg)
	apiKey, err := resolveAPIKey(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		gw:              setupDefaultGateway(apiKey),
		defaultProvider: cfg.DefaultProvider,
		defaultModel:    cfg.DefaultModel,
		defaultSettings: cfg.DefaultSettings,
	}, nil
}

// NewClientFromGateway wraps a pre-wired gateway (e.g. for testing or custom setups).
func NewClientFromGateway(gw *gateway.LLMGateway) *Client {
	return &Client{
		gw:              gw,
		defaultProvider: "faux",
		defaultModel:    "faux-model",
	}
}

// NewSession creates a session for the given agent definition.
func (c *Client) NewSession(agent *Agent) (*Session, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent is required")
	}

	modelID := agent.model
	if modelID == "" {
		modelID = c.defaultModel
	}

	model, err := resolve.Model(c.gw, c.defaultProvider, modelID)
	if err != nil {
		return nil, err
	}

	settings := mergeSettings(c.defaultSettings, agent.settings)
	ag := wire.NewAgent(c.gw, model, agent.instructions, wire.Settings{
		Temperature: settings.Temperature,
		MaxTokens:   settings.MaxTokens,
	})

	return &Session{ag: ag}, nil
}

func mergeSettings(base, override *ModelSettings) ModelSettings {
	var out ModelSettings
	if base != nil {
		out.Temperature = base.Temperature
		out.MaxTokens = base.MaxTokens
	}
	if override != nil {
		if override.Temperature != nil {
			out.Temperature = override.Temperature
		}
		if override.MaxTokens != nil {
			out.MaxTokens = override.MaxTokens
		}
	}
	return out
}
