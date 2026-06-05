package pithsdk_test

import (
	"testing"

	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestNewAgentAcceptsTools(t *testing.T) {
	tool := pithsdk.NewTool("noop", "No-op tool.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "ok", nil
		},
	)
	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Tools: []pithsdk.Tool{tool},
	})
	if err != nil {
		t.Fatalf("NewAgent with tools: %v", err)
	}
	if agent == nil {
		t.Fatal("expected agent")
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
}
