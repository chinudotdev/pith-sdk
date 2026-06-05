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

func toolResultFromText(text string, err error) loop.ToolResult {
	if err != nil {
		return loop.ToolResult{
			Content: []protocol.Content{protocol.TextContent{Type: "text", Text: err.Error()}},
		}
	}
	return loop.ToolResult{
		Content: []protocol.Content{protocol.TextContent{Type: "text", Text: text}},
	}
}

// ToAgentTools converts typed tools into loop.AgentTool values.
func ToAgentTools(typed []TypedTool, holder *RunScopeHolder) []loop.AgentTool {
	out := make([]loop.AgentTool, 0, len(typed))
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
				return toolResultFromText(text, err)
			},
		})
	}
	return out
}
