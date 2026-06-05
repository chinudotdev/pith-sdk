package pithsdk_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestSessionRunFauxGateway(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello! I am a helpful assistant."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	result, err := session.Run(context.Background(), "Hi there!")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Hello! I am a helpful assistant." {
		t.Fatalf("expected assistant text, got %q", result.Text)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected messages in result")
	}
}

func TestSessionRunInstructionsOverride(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Override works."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Original instructions.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	_, err = session.Run(context.Background(), "Test", pithsdk.WithInstructions("Temporary instructions."))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Ensure original prompt is restored — second run should still succeed.
	if _, err = session.Run(context.Background(), "Again"); err != nil {
		t.Fatalf("second Run: %v", err)
	}
}

func TestSessionRunWithTool(t *testing.T) {
	weather := pithsdk.NewTool("get_weather", "Return weather for a city.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			return fmt.Sprintf("Sunny in %s", args.City), nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "It's sunny in SF."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are a helpful weather bot.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{weather},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	result, err := session.Run(context.Background(), "What's the weather in SF?")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "It's sunny in SF." {
		t.Fatalf("expected final text, got %q", result.Text)
	}

	var foundToolResult bool
	for _, msg := range result.Messages {
		if msg.Role == "toolResult" && msg.Text == "Sunny in SF" {
			foundToolResult = true
			break
		}
	}
	if !foundToolResult {
		t.Fatalf("expected toolResult in messages, got %+v", result.Messages)
	}
}

func TestToolContextLocal(t *testing.T) {
	type deps struct {
		Label string
	}

	var gotLocal any
	tool := pithsdk.NewTool("echo_label", "Echo the run-scoped label.",
		func(ctx pithsdk.ToolContext, args struct {
			Unused string `json:"unused,omitempty"`
		}) (string, error) {
			gotLocal = ctx.Local
			return "ok", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "echo_label", Arguments: `{}`},
			},
		},
		gateway.FauxResponse{Text: "done"},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools when asked.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{tool},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	local := deps{Label: "test-db"}
	if _, err := session.Run(context.Background(), "go", pithsdk.WithContext(local)); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, ok := gotLocal.(deps)
	if !ok {
		t.Fatalf("expected deps in ToolContext.Local, got %T", gotLocal)
	}
	if got.Label != "test-db" {
		t.Fatalf("expected label test-db, got %q", got.Label)
	}
}
