package wire

import (
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith/protocol"
)

// TypedTool is a tool definition with a handler that receives the active run scope.
type TypedTool struct {
	Name        string
	Description string
	Parameters  any
	Handler     func(holder *RunScopeHolder, callID string, params map[string]any) (string, error)
}

// ToAgentTools converts typed and raw tools into loop.AgentTool values.
func ToAgentTools(typed []TypedTool, raw []loop.AgentTool, holder *RunScopeHolder) []loop.AgentTool {
	out := make([]loop.AgentTool, 0, len(typed)+len(raw))
	for _, td := range typed {
		handler := td.Handler
		out = append(out, loop.AgentTool{
			Name:        td.Name,
			Label:       td.Name,
			Description: td.Description,
			Parameters:  td.Parameters,
			Execute: func(callID string, params map[string]any, signal <-chan struct{}, onUpdate func(partial any)) loop.ToolResult {
				_ = signal
				_ = onUpdate
				text, err := handler(holder, callID, params)
				if err != nil {
					return loop.ToolResult{
						Content: []protocol.Content{protocol.TextContent{Type: "text", Text: err.Error()}},
					}
				}
				return loop.ToolResult{
					Content: []protocol.Content{protocol.TextContent{Type: "text", Text: text}},
				}
			},
		})
	}
	out = append(out, raw...)
	return out
}
