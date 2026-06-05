// Package mcp discovers tools from MCP (Model Context Protocol) servers and
// converts them to pith-sdk Tools for use in agent sessions.
package mcp

import (
	"context"
	"fmt"
	"os/exec"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	pithsdk "github.com/chinudotdev/pith-sdk"
	"github.com/chinudotdev/pith-sdk/internal"
)

// Config configures an MCP server connection.
type Config struct {
	// Command is the path to the MCP server executable.
	// Required for stdio transport.
	Command string
	// Args are command-line arguments passed to the MCP server.
	Args []string
	// Env are environment variables passed to the MCP server (KEY=VALUE format).
	Env []string
}

// Tools discovers tools from an MCP server and returns them as pith-sdk Tools.
// The returned close function should be called to shut down the MCP server connection.
func Tools(ctx context.Context, cfg Config) (tools []pithsdk.Tool, close func() error, err error) {
	if cfg.Command == "" {
		return nil, nil, fmt.Errorf("mcp: Command is required")
	}

	cmd := exec.Command(cfg.Command, cfg.Args...)
	if len(cfg.Env) > 0 {
		cmd.Env = append(cmd.Environ(), cfg.Env...)
	}

	client := mcpsdk.NewClient(&mcpsdk.Implementation{
		Name:    "pith-sdk-mcp",
		Version: "1.0.0",
	}, nil)

	transport := &mcpsdk.CommandTransport{Command: cmd}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("mcp: connect: %w", err)
	}

	listResult, err := session.ListTools(ctx, nil)
	if err != nil {
		session.Close()
		return nil, nil, fmt.Errorf("mcp: list tools: %w", err)
	}

	for _, mcpTool := range listResult.Tools {
		tool := mcpTool // capture
		schema, err := internal.MarshalSchema(tool.InputSchema)
		if err != nil {
			session.Close()
			return nil, nil, fmt.Errorf("mcp: schema for tool %q: %w", tool.Name, err)
		}

		parsedTool := pithsdk.NewDynamicTool(
			tool.Name,
			tool.Description,
			schema,
			func(runCtx context.Context, args map[string]any) (string, error) {
				result, err := session.CallTool(runCtx, &mcpsdk.CallToolParams{
					Name:      tool.Name,
					Arguments: args,
				})
				if err != nil {
					return "", err
				}
				if result.IsError {
					return extractText(result), fmt.Errorf("mcp tool error: %s", extractText(result))
				}
				return extractText(result), nil
			},
		)
		tools = append(tools, parsedTool)
	}

	return tools, session.Close, nil
}

// extractText returns the concatenated text content from a CallToolResult.
func extractText(result *mcpsdk.CallToolResult) string {
	var text string
	for _, c := range result.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			if text != "" {
				text += "\n"
			}
			text += tc.Text
		}
	}
	return text
}
