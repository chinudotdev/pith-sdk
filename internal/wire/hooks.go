package wire

import (
	"context"
	"fmt"
)

// RunCtx returns the active run context from holder, or context.Background().
func RunCtx(holder *RunScopeHolder) context.Context {
	if scope := holder.Current(); scope != nil && scope.Ctx != nil {
		return scope.Ctx
	}
	return context.Background()
}

// RunWithHooks runs BeforeToolCall, invoke, and AfterToolCall for a tool handler.
func RunWithHooks(holder *RunScopeHolder, agentName, toolName, callID string, params map[string]any, run func() (string, error)) (string, error) {
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
			return "", err
		}
		if block {
			return fmt.Sprintf("tool call blocked: %s", reason), nil
		}
	}

	result, invokeErr := run()

	if hooks != nil && hooks.AfterToolCall != nil {
		override, hookErr := hooks.AfterToolCall(sessionID, runID, agentName, toolName, callID, params, result, invokeErr)
		if hookErr != nil {
			return "", hookErr
		}
		if override != "" {
			return override, nil
		}
	}

	return result, invokeErr
}
