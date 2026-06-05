package pithsdk

import (
	"context"
	"fmt"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
	"github.com/chinudotdev/pith-sdk/internal/resolve"
	"github.com/chinudotdev/pith-sdk/internal/wire"
)

// Client holds gateway configuration, credentials, and defaults for running agents.
type Client struct {
	gw              *gateway.LLMGateway
	defaultProvider string
	defaultModel    string
	defaultSettings *ModelSettings
	credentials     credentialRegistry
	registered      map[protocol.ProviderId]bool
}

// ClientConfig configures a new Client.
type ClientConfig struct {
	// APIKey is the OpenAI-compatible API key. When empty, OPENAI_API_KEY is used at run time.
	APIKey string
	// DefaultProvider is the provider ID for bare model strings (default "openai").
	DefaultProvider string
	// DefaultModel is used when an Agent has no Model set (default "gpt-4o-mini").
	DefaultModel string
	// DefaultSettings applies generation defaults to all agents unless overridden.
	DefaultSettings *ModelSettings
}

// NewClient creates a client with a built-in OpenAI-compatible provider.
// An OpenAI API key is optional at construction time; it is required when using
// the default OpenAI provider at run time.
func NewClient(cfg ClientConfig) (*Client, error) {
	applyClientDefaults(&cfg)
	creds := newCredentialRegistry()
	creds.set(defaultProviderID, openAIResolver(resolveAPIKey(cfg)))

	client := &Client{
		gw:              setupDefaultGateway(creds),
		defaultProvider: cfg.DefaultProvider,
		defaultModel:    cfg.DefaultModel,
		defaultSettings: cfg.DefaultSettings,
		credentials:     creds,
		registered: map[protocol.ProviderId]bool{
			defaultProviderID: true,
		},
	}
	return client, nil
}

// NewClientFromGateway wraps a pre-wired gateway (e.g. for testing or custom setups).
func NewClientFromGateway(gw *gateway.LLMGateway) *Client {
	return &Client{
		gw:              gw,
		defaultProvider: "faux",
		defaultModel:    "faux-model",
		credentials:     newCredentialRegistry(),
		registered:      make(map[protocol.ProviderId]bool),
	}
}

func (c *Client) syncCredentials() {
	c.gw.Credentials = c.credentials.credentialProvider()
}

// NewSession creates a session for the given agent definition.
func (c *Client) NewSession(agent *Agent) (*Session, error) {
	if agent == nil {
		return nil, fmt.Errorf("agent is required")
	}

	modelString := agent.model
	if modelString == "" {
		modelString = c.defaultModel
	}

	model, err := resolve.Model(c.gw, c.defaultProvider, modelString)
	if err != nil {
		return nil, err
	}

	settings := mergeSettings(c.defaultSettings, agent.settings)
	scope := wire.NewRunScopeHolder()
	loopTools := toWireTools(agent.tools, scope)
	ag := wire.NewAgent(c.gw, model, agent.instructions, wire.Settings{
		Temperature: settings.Temperature,
		MaxTokens:   settings.MaxTokens,
	}, loopTools)

	return &Session{ag: ag, scope: scope}, nil
}

// RunOnce runs a single prompt without requiring the caller to manage a Session.
func (c *Client) RunOnce(ctx context.Context, agent *Agent, input string, opts ...RunOption) (*RunResult, error) {
	session, err := c.NewSession(agent)
	if err != nil {
		return nil, err
	}
	return session.Run(ctx, input, opts...)
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
