package pithsdk_test

import (
	"context"
	"strings"
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

func TestNewClientMissingAPIKeyDeferred(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient should succeed without API key: %v", err)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	_, err = session.Run(context.Background(), "Hi")
	if err == nil {
		t.Fatal("expected auth error when OpenAI key is missing at run time")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}
