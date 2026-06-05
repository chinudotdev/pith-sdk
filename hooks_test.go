package pithsdk_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestBeforeToolCallBlock(t *testing.T) {
	weather := pithsdk.NewTool("get_weather", "Return weather.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			return "Sunny", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "Blocked result."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
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

	var calledBefore bool
	result, err := session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			calledBefore = true
			if ctx.ToolName != "get_weather" {
				t.Fatalf("expected tool name get_weather, got %q", ctx.ToolName)
			}
			if ctx.CallID != "tc1" {
				t.Fatalf("expected call ID tc1, got %q", ctx.CallID)
			}
			return &pithsdk.BeforeToolResult{Block: true, Reason: "not allowed"}, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !calledBefore {
		t.Fatal("expected BeforeToolCall to be called")
	}
	if result.Text != "Blocked result." {
		t.Fatalf("expected blocked result text, got %q", result.Text)
	}
}

func TestBeforeToolCallAllow(t *testing.T) {
	weather := pithsdk.NewTool("get_weather", "Return weather.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			return "Sunny in SF", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
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

	var toolExecuted bool
	_, err = session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			return nil, nil // allow
		},
		AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
			toolExecuted = true
			if ctx.Result != "Sunny in SF" {
				t.Fatalf("expected tool result in after hook, got %q", ctx.Result)
			}
			return nil, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !toolExecuted {
		t.Fatal("expected AfterToolCall to be called")
	}
}

func TestAfterToolCallOverride(t *testing.T) {
	weather := pithsdk.NewTool("get_weather", "Return weather.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			return "Original result", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
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

	var foundOverridden bool
	result, err := session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
			return &pithsdk.AfterToolResult{OverrideResult: "Redacted"}, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, msg := range result.Messages {
		if msg.Role == "toolResult" && msg.Text == "Redacted" {
			foundOverridden = true
		}
	}
	if !foundOverridden {
		t.Fatalf("expected overridden tool result in messages, got %+v", result.Messages)
	}
}

func TestHookContextIDs(t *testing.T) {
	weather := pithsdk.NewTool("get_weather", "Return weather.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city"`
		}) (string, error) {
			return "Sunny", nil
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc-1", Name: "get_weather", Arguments: `{"city":"SF"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{weather},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent, pithsdk.WithSessionID("sess-hook"))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var gotRunID, gotSessionID string
	_, err = session.Run(context.Background(), "Go", pithsdk.WithRunID("run-hook"), pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			gotRunID = ctx.RunID
			gotSessionID = ctx.SessionID
			return nil, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotRunID != "run-hook" {
		t.Fatalf("expected RunID run-hook, got %q", gotRunID)
	}
	if gotSessionID != "sess-hook" {
		t.Fatalf("expected SessionID sess-hook, got %q", gotSessionID)
	}
}

func TestAfterToolCallSeesError(t *testing.T) {
	failTool := pithsdk.NewTool("fail", "Always fails.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "", fmt.Errorf("something went wrong")
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "fail", Arguments: `{}`},
			},
		},
		gateway.FauxResponse{Text: "Handled error."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{failTool},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var sawError bool
	_, err = session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
			if ctx.Error != nil {
				sawError = true
			}
			return nil, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !sawError {
		t.Fatal("expected AfterToolCall to see tool error")
	}
}

func TestAfterToolCallOverrideOnError(t *testing.T) {
	failTool := pithsdk.NewTool("fail", "Always fails.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "", fmt.Errorf("something went wrong")
		},
	)

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "fail", Arguments: `{}`},
			},
		},
		gateway.FauxResponse{Text: "Handled."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{failTool},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var sawError bool
	result, err := session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
			if ctx.Error != nil {
				sawError = true
			}
			return &pithsdk.AfterToolResult{OverrideResult: "Recovered"}, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !sawError {
		t.Fatal("expected AfterToolCall to see tool error before override")
	}

	var foundOverridden bool
	for _, msg := range result.Messages {
		if msg.Role == "toolResult" && msg.Text == "Recovered" {
			foundOverridden = true
		}
		if msg.Role == "toolResult" && msg.Text == "something went wrong" {
			t.Fatalf("expected error text to be overridden, got raw error in messages")
		}
	}
	if !foundOverridden {
		t.Fatalf("expected overridden tool result in messages, got %+v", result.Messages)
	}
}

func TestMCPTracingIDs(t *testing.T) {
	dynamic := pithsdk.ToolFromDynamicSchema("echo", "Echo text.", map[string]any{
		"type": "object",
		"properties": map[string]any{
			"text": map[string]any{"type": "string"},
		},
	}, func(ctx context.Context, args map[string]any) (string, error) {
		text, _ := args["text"].(string)
		return text, nil
	})

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc-mcp", Name: "echo", Arguments: `{"text":"hi"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{dynamic},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent, pithsdk.WithSessionID("sess-mcp"))
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var gotRunID, gotSessionID string
	_, err = session.Run(context.Background(), "Go", pithsdk.WithRunID("run-mcp"), pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			gotRunID = ctx.RunID
			gotSessionID = ctx.SessionID
			if ctx.ToolName != "echo" {
				t.Fatalf("expected tool name echo, got %q", ctx.ToolName)
			}
			return nil, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if gotRunID != "run-mcp" {
		t.Fatalf("expected RunID run-mcp, got %q", gotRunID)
	}
	if gotSessionID != "sess-mcp" {
		t.Fatalf("expected SessionID sess-mcp, got %q", gotSessionID)
	}
}
