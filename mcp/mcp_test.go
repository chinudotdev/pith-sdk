package mcp_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
	pithsdk "github.com/chinudotdev/pith-sdk"
	"github.com/chinudotdev/pith-sdk/mcp"
)

var mockServerBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "pith-sdk-mcp-mock-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	bin := filepath.Join(dir, "mock-mcp-server")
	src := filepath.Join(dir, "main.go")
	if err := os.WriteFile(src, []byte(mockServerSrc), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write mock server src: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "build", "-o", bin, src)
	cmd.Env = append(os.Environ(), "GOOS="+os.Getenv("GOOS"), "GOARCH="+os.Getenv("GOARCH"))
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build mock server: %v\n%s", err, out)
		os.Exit(1)
	}

	mockServerBin = bin
	os.Exit(m.Run())
}

const mockServerSrc = `package main

import (
	"context"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "mock-mcp", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "echo",
		Description: "Echoes back the input text.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Text string ` + "`" + `json:"text"` + "`" + `
	}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: args.Text}},
		}, nil, nil
	})
	server.Run(context.Background(), &mcp.StdioTransport{})
	os.Exit(0)
}
`

func TestMCPToolsDiscovery(t *testing.T) {
	tools, close, err := mcp.Tools(context.Background(), mcp.Config{
		Command: mockServerBin,
	})
	if err != nil {
		t.Fatalf("mcp.Tools: %v", err)
	}
	defer close()

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
}

func TestMCPToolsWithAgent(t *testing.T) {
	mcpTools, close, err := mcp.Tools(context.Background(), mcp.Config{
		Command: mockServerBin,
	})
	if err != nil {
		t.Fatalf("mcp.Tools: %v", err)
	}
	defer close()

	if len(mcpTools) == 0 {
		t.Fatal("expected at least one MCP tool")
	}

	// Compose with local tools
	localTool := pithsdk.NewTool("local_greet", "Greets the user.",
		func(ctx pithsdk.ToolContext, args struct {
			Name string `json:"name"`
		}) (string, error) {
			return fmt.Sprintf("Hello, %s!", args.Name), nil
		},
	)

	allTools := append([]pithsdk.Tool{localTool}, mcpTools...)

	_ = allTools // tools are valid and can be passed to NewAgent
}

func TestMCPConfigValidation(t *testing.T) {
	_, _, err := mcp.Tools(context.Background(), mcp.Config{})
	if err == nil {
		t.Fatal("expected error for empty Command")
	}
	if err.Error() != "mcp: Command is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMCPInvokeViaAgent(t *testing.T) {
	mcpTools, close, err := mcp.Tools(context.Background(), mcp.Config{
		Command: mockServerBin,
	})
	if err != nil {
		t.Fatalf("mcp.Tools: %v", err)
	}
	defer close()

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "echo", Arguments: `{"text":"hello"}`},
			},
		},
		gateway.FauxResponse{Text: "Echo complete."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        mcpTools,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	result, err := session.Run(context.Background(), "Echo hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Echo complete." {
		t.Fatalf("expected final text %q, got %q", "Echo complete.", result.Text)
	}

	var foundToolResult bool
	for _, msg := range result.Messages {
		if msg.Role == "toolResult" && msg.Text == "hello" {
			foundToolResult = true
			break
		}
	}
	if !foundToolResult {
		t.Fatalf("expected toolResult containing hello, got %+v", result.Messages)
	}
}

func TestMCPHooksFire(t *testing.T) {
	mcpTools, close, err := mcp.Tools(context.Background(), mcp.Config{
		Command: mockServerBin,
	})
	if err != nil {
		t.Fatalf("mcp.Tools: %v", err)
	}
	defer close()

	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{
			ToolCalls: []protocol.ToolCall{
				{Type: "toolCall", ID: "tc1", Name: "echo", Arguments: `{"text":"hello"}`},
			},
		},
		gateway.FauxResponse{Text: "Blocked result."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Use tools.",
		Model:        "faux-model",
		Tools:        mcpTools,
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	var calledBefore bool
	result, err := session.Run(context.Background(), "Echo hello", pithsdk.WithHooks(pithsdk.Hooks{
		BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
			calledBefore = true
			if ctx.ToolName != "echo" {
				t.Fatalf("expected tool name echo, got %q", ctx.ToolName)
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
		t.Fatal("expected BeforeToolCall to be called for MCP tool")
	}
	if result.Text != "Blocked result." {
		t.Fatalf("expected blocked result text, got %q", result.Text)
	}
}
