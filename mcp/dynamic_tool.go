package mcp

import (
	"context"

	pithsdk "github.com/chinudotdev/pith-sdk"
)

// dynamicTool creates a schema-driven pithsdk.Tool for MCP-discovered tools.
func dynamicTool(name, description string, schema map[string]any,
	fn func(context.Context, map[string]any) (string, error)) pithsdk.Tool {
	return pithsdk.ToolFromDynamicSchema(name, description, schema, fn)
}
