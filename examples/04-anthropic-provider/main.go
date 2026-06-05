// Example 04: Custom Anthropic provider via pith-sdk RegisterProvider.
//
//	ANTHROPIC_API_KEY="sk-ant-..." go run ./examples/04-anthropic-provider/
package main

import (
	"context"
	"fmt"
	"os"

	pithsdk "github.com/chinudotdev/pith-sdk"
)

func main() {
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewClient: %v\n", err)
		os.Exit(1)
	}

	if err := client.RegisterProvider(pithsdk.ProviderRegistration{
		Provider:  NewAnthropicProvider(AnthropicConfig{}),
		APIKeyEnv: "ANTHROPIC_API_KEY",
		Models: []pithsdk.ModelPreset{
			{
				ID:            "claude-sonnet-4-20250514",
				Name:          "Claude Sonnet 4",
				BaseURL:       "https://api.anthropic.com",
				ContextWindow: 200_000,
				MaxTokens:     8192,
			},
		},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "RegisterProvider: %v\n", err)
		os.Exit(1)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are a helpful assistant. Be concise.",
		Model:        "anthropic/claude-sonnet-4-20250514",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewAgent: %v\n", err)
		os.Exit(1)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewSession: %v\n", err)
		os.Exit(1)
	}

	result, err := session.Run(context.Background(), "Hello, who are you? Answer in one sentence.")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Run: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Text)
}
