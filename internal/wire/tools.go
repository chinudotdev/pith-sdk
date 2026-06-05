package wire

import (
	"fmt"
	"strings"

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

// WrapRawTool wraps a loop.AgentTool so Before/After hooks fire around inner.Execute.
func WrapRawTool(holder *RunScopeHolder, agentName string, raw loop.AgentTool) loop.AgentTool {
	inner := raw
	toolName := raw.Name
	return loop.AgentTool{
		Name:        inner.Name,
		Label:       inner.Label,
		Description: inner.Description,
		Parameters:  inner.Parameters,
		Execute: func(callID string, params map[string]any, signal <-chan struct{}, onUpdate func(partial any)) loop.ToolResult {
			_ = signal
			_ = onUpdate

			var sessionID, runID string
			var hooks *HookSet
			if scope := holder.Current(); scope != nil {
				sessionID = scope.SessionID
				runID = scope.RunID
				hooks = scope.Hooks
			}

			if hooks != nil && hooks.BeforeToolCall != nil {
				block, reason, err := hooks.BeforeToolCall(sessionID, runID, agentName, toolName, callID, params)
				if err != nil {
					return loop.ToolResult{
						Content: []protocol.Content{protocol.TextContent{Type: "text", Text: err.Error()}},
					}
				}
				if block {
					return loop.ToolResult{
						Content: []protocol.Content{protocol.TextContent{Type: "text", Text: fmt.Sprintf("tool call blocked: %s", reason)}},
					}
				}
			}

			result := inner.Execute(callID, params, signal, onUpdate)
			text := extractToolResultText(result)

			if hooks != nil && hooks.AfterToolCall != nil {
				override, hookErr := hooks.AfterToolCall(sessionID, runID, agentName, toolName, callID, params, text, nil)
				if hookErr != nil {
					return loop.ToolResult{
						Content: []protocol.Content{protocol.TextContent{Type: "text", Text: hookErr.Error()}},
					}
				}
				if override != "" {
					return loop.ToolResult{
						Content: []protocol.Content{protocol.TextContent{Type: "text", Text: override}},
					}
				}
			}

			return result
		},
	}
}

func extractToolResultText(result loop.ToolResult) string {
	var texts []string
	for _, c := range result.Content {
		if tc, ok := c.(protocol.TextContent); ok {
			texts = append(texts, tc.Text)
		}
	}
	return strings.Join(texts, "")
}

// ToAgentTools converts typed and raw tools into loop.AgentTool values.
func ToAgentTools(typed []TypedTool, raw []loop.AgentTool, holder *RunScopeHolder, agentName string) []loop.AgentTool {
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
	for _, r := range raw {
		out = append(out, WrapRawTool(holder, agentName, r))
	}
	return out
}
