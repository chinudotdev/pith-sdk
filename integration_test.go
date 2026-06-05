package pithsdk_test

import (
	"context"
	"fmt"
	"strings"
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

func TestSessionMultiTurn(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "First reply."},
		gateway.FauxResponse{Text: "Second reply."},
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

	if _, err := session.Run(context.Background(), "Turn one."); err != nil {
		t.Fatalf("first Run: %v", err)
	}

	result, err := session.Run(context.Background(), "Turn two.")
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if result.Text != "Second reply." {
		t.Fatalf("expected second reply text, got %q", result.Text)
	}

	msgs := session.Messages()
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d: %+v", len(msgs), msgs)
	}
	wantRoles := []string{"user", "assistant", "user", "assistant"}
	for i, role := range wantRoles {
		if msgs[i].Role != role {
			t.Fatalf("message %d: expected role %q, got %q", i, role, msgs[i].Role)
		}
	}
}

func TestSessionReset(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Before reset."},
		gateway.FauxResponse{Text: "After reset."},
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

	if _, err := session.Run(context.Background(), "First."); err != nil {
		t.Fatalf("first Run: %v", err)
	}
	if len(session.Messages()) != 2 {
		t.Fatalf("expected 2 messages before reset, got %d", len(session.Messages()))
	}

	session.Reset()
	if len(session.Messages()) != 0 {
		t.Fatalf("expected 0 messages after reset, got %d", len(session.Messages()))
	}

	result, err := session.Run(context.Background(), "Second.")
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if result.Text != "After reset." {
		t.Fatalf("expected after reset text, got %q", result.Text)
	}
	if len(session.Messages()) != 2 {
		t.Fatalf("expected 2 messages after fresh run, got %d", len(session.Messages()))
	}
}

func TestSessionRunWithStream(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello, world!"},
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

	var deltas []string
	result, err := session.Run(context.Background(), "Hi!", pithsdk.WithStream(func(c pithsdk.TextChunk) {
		if c.Delta != "" {
			deltas = append(deltas, c.Delta)
		}
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Hello, world!" {
		t.Fatalf("expected final text, got %q", result.Text)
	}
	if len(deltas) <= 1 {
		t.Fatalf("expected multiple stream deltas, got %d: %v", len(deltas), deltas)
	}
}

func TestRegisterProviderFaux(t *testing.T) {
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	faux := gateway.NewFauxProvider(
		protocol.ApiAnthropicMessages,
		"anthropic",
		gateway.FauxResponse{Text: "Custom provider reply."},
	)
	if err := client.RegisterProvider(pithsdk.ProviderRegistration{
		Provider: faux,
		APIKey:   "test-key",
		Models: []pithsdk.ModelPreset{
			{ID: "claude-test"},
		},
	}); err != nil {
		t.Fatalf("RegisterProvider: %v", err)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "anthropic/claude-test",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	result, err := session.Run(context.Background(), "Hi!")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Custom provider reply." {
		t.Fatalf("expected custom provider text, got %q", result.Text)
	}
}

func TestRegisterProviderUnknownModel(t *testing.T) {
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	faux := gateway.NewFauxProvider(protocol.ApiAnthropicMessages, "anthropic")
	if err := client.RegisterProvider(pithsdk.ProviderRegistration{
		Provider: faux,
		APIKey:   "test-key",
		Models: []pithsdk.ModelPreset{
			{ID: "claude-test"},
		},
	}); err != nil {
		t.Fatalf("RegisterProvider: %v", err)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "anthropic/missing-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	_, err = client.NewSession(agent)
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if !strings.Contains(err.Error(), `unknown model "anthropic/missing-model"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegisterProviderMissingCredential(t *testing.T) {
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	faux := gateway.NewFauxProvider(protocol.ApiAnthropicMessages, "anthropic")
	err = client.RegisterProvider(pithsdk.ProviderRegistration{
		Provider: faux,
		Models: []pithsdk.ModelPreset{
			{ID: "claude-test"},
		},
	})
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
	if !strings.Contains(err.Error(), "credentials required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionRunMaxTurnsExceeded(t *testing.T) {
	noop := pithsdk.NewTool("noop", "No-op tool.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "ok", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "noop", Arguments: `{}`},
			},
		},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{noop},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	_, err = session.Run(context.Background(), "Keep calling noop.", pithsdk.WithMaxTurns(2))
	if err == nil {
		t.Fatal("expected max turns error")
	}
	if !strings.Contains(err.Error(), "max turns (2) exceeded") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientRunOnce(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "One-shot reply."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	result, err := client.RunOnce(context.Background(), agent, "Quick question.")
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
	if result.Text != "One-shot reply." {
		t.Fatalf("expected one-shot text, got %q", result.Text)
	}
}

func TestSessionAutoID(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello."},
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

	if session.ID() == "" {
		t.Fatal("expected auto-generated session ID")
	}
}

func TestSessionExplicitID(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent, pithsdk.WithSessionID("my-session-42"))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if session.ID() != "my-session-42" {
		t.Fatalf("expected explicit session ID, got %q", session.ID())
	}
}

func TestRunIDAutoGenerated(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello."},
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

	result, err := session.Run(context.Background(), "Hi")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID == "" {
		t.Fatal("expected auto-generated run ID on result")
	}
}

func TestRunIDExplicit(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello."},
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

	result, err := session.Run(context.Background(), "Hi", pithsdk.WithRunID("run-abc-123"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.RunID != "run-abc-123" {
		t.Fatalf("expected explicit run ID, got %q", result.RunID)
	}
}

func TestTracingIDsPropagatedToToolContext(t *testing.T) {
	var gotRunID, gotSessionID, gotToolName, gotCallID string

	weather := pithsdk.NewTool("get_weather", "Return weather.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			gotRunID = ctx.RunID
			gotSessionID = ctx.SessionID
			gotToolName = ctx.ToolName
			gotCallID = ctx.CallID
			return "Sunny", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc-trace-1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{weather},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent, pithsdk.WithSessionID("sess-42"))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	result, err := session.Run(context.Background(), "Go", pithsdk.WithRunID("run-99"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if gotRunID != "run-99" {
		t.Fatalf("expected RunID run-99 in ToolContext, got %q", gotRunID)
	}
	if gotSessionID != "sess-42" {
		t.Fatalf("expected SessionID sess-42 in ToolContext, got %q", gotSessionID)
	}
	if gotToolName != "get_weather" {
		t.Fatalf("expected ToolName get_weather in ToolContext, got %q", gotToolName)
	}
	if gotCallID != "tc-trace-1" {
		t.Fatalf("expected CallID tc-trace-1 in ToolContext, got %q", gotCallID)
	}
	if result.RunID != "run-99" {
		t.Fatalf("expected RunID run-99 on result, got %q", result.RunID)
	}
}
