package pithsdk_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/loop"
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

func TestShouldStopAfterTurn(t *testing.T) {
	noop := pithsdk.NewTool("noop", "No-op tool.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "ok", nil
		},
	)

	// Return multiple tool call responses to simulate multi-turn
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "noop", Arguments: `{}`},
			},
		},
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc2", Name: "noop", Arguments: `{}`},
			},
		},
		gateway.FauxResponse{Text: "Final."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
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

	var turnNumbers []int
	_, err = session.Run(context.Background(), "Go", pithsdk.WithMaxTurns(10), pithsdk.WithHooks(pithsdk.Hooks{
		ShouldStopAfterTurn: func(ctx pithsdk.TurnContext) bool {
			turnNumbers = append(turnNumbers, ctx.TurnNumber)
			return ctx.TurnNumber >= 1 // stop after first turn
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(turnNumbers) == 0 {
		t.Fatal("expected ShouldStopAfterTurn to be called")
	}
	if turnNumbers[0] != 1 {
		t.Fatalf("expected first turn number 1, got %d", turnNumbers[0])
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
		Name:         "weather-agent",
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

	var gotRunID, gotSessionID, gotAgentName string
	_, err = session.Run(context.Background(), "Go", pithsdk.WithRunID("run-hook"), pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			gotRunID = ctx.RunID
			gotSessionID = ctx.SessionID
			gotAgentName = ctx.AgentName
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
	if gotAgentName != "weather-agent" {
		t.Fatalf("expected AgentName weather-agent, got %q", gotAgentName)
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

func TestRawToolHooksFire(t *testing.T) {
	raw := pithsdk.RawTool(loop.AgentTool{
		Name:        "raw_echo",
		Label:       "raw_echo",
		Description: "Echo via RawTool.",
		Execute: func(callID string, params map[string]any, signal <-chan struct{}, onUpdate func(partial any)) loop.ToolResult {
			text, _ := params["text"].(string)
			return loop.ToolResult{
				Content: []protocol.Content{protocol.TextContent{Type: "text", Text: text}},
			}
		},
	})

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc-raw", Name: "raw_echo", Arguments: `{"text":"hello"}`},
			},
		},
		gateway.FauxResponse{Text: "Done."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        []pithsdk.Tool{raw},
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var calledBefore, calledAfter bool
	result, err := session.Run(context.Background(), "Go", pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			calledBefore = true
			if ctx.ToolName != "raw_echo" {
				t.Fatalf("expected tool name raw_echo, got %q", ctx.ToolName)
			}
			if ctx.CallID != "tc-raw" {
				t.Fatalf("expected call ID tc-raw, got %q", ctx.CallID)
			}
			return nil, nil
		},
		AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
			calledAfter = true
			if ctx.Result != "hello" {
				t.Fatalf("expected tool result hello, got %q", ctx.Result)
			}
			return nil, nil
		},
	}))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !calledBefore {
		t.Fatal("expected BeforeToolCall to fire for RawTool")
	}
	if !calledAfter {
		t.Fatal("expected AfterToolCall to fire for RawTool")
	}

	var foundToolResult bool
	for _, msg := range result.Messages {
		if msg.Role == "toolResult" && msg.Text == "hello" {
			foundToolResult = true
		}
	}
	if !foundToolResult {
		t.Fatalf("expected raw tool result in messages, got %+v", result.Messages)
	}
}

func TestMCPTracingIDs(t *testing.T) {
	// NewDynamicTool uses the same hook path MCP will use after migration.
	dynamic := pithsdk.NewDynamicTool("echo", "Echo text.", map[string]any{
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
		Name:         "mcp-agent",
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

	var gotRunID, gotSessionID, gotAgentName string
	_, err = session.Run(context.Background(), "Go", pithsdk.WithRunID("run-mcp"), pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			gotRunID = ctx.RunID
			gotSessionID = ctx.SessionID
			gotAgentName = ctx.AgentName
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
	if gotAgentName != "mcp-agent" {
		t.Fatalf("expected AgentName mcp-agent, got %q", gotAgentName)
	}
}
