// Minimal stdio MCP server with an echo tool. Built and run by examples/05-mcp.
package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	s := mcp.NewServer(&mcp.Implementation{Name: "echo", Version: "1.0.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "echo",
		Description: "Echoes back the input text.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args struct {
		Text string `json:"text"`
	}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: args.Text}},
		}, nil, nil
	})
	if err := s.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
