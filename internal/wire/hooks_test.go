package wire

import (
	"context"
	"fmt"
	"testing"
)

func TestRunWithHooksBlock(t *testing.T) {
	holder := NewRunScopeHolder()
	holder.Set(context.Background(), nil, "sess-1", "run-1", &HookSet{
		BeforeToolCall: func(sessionID, runID, toolName, callID string, args map[string]any) (bool, string, error) {
			return true, "denied", nil
		},
	})
	defer holder.Clear()

	text, err := RunWithHooks(holder, "my_tool", "call-1", map[string]any{"x": 1}, func() (string, error) {
		t.Fatal("invoke should not run when blocked")
		return "", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "tool call blocked: denied" {
		t.Fatalf("expected blocked message, got %q", text)
	}
}

func TestRunWithHooksOverride(t *testing.T) {
	holder := NewRunScopeHolder()
	holder.Set(context.Background(), nil, "sess-1", "run-1", &HookSet{
		AfterToolCall: func(sessionID, runID, toolName, callID string, args map[string]any, result string, resultErr error) (string, error) {
			return "overridden", nil
		},
	})
	defer holder.Clear()

	text, err := RunWithHooks(holder, "my_tool", "call-1", nil, func() (string, error) {
		return "original", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "overridden" {
		t.Fatalf("expected overridden result, got %q", text)
	}
}

func TestRunWithHooksHookErrorBecomesResult(t *testing.T) {
	holder := NewRunScopeHolder()
	holder.Set(context.Background(), nil, "sess-1", "run-1", &HookSet{
		BeforeToolCall: func(sessionID, runID, toolName, callID string, args map[string]any) (bool, string, error) {
			return false, "", fmt.Errorf("hook failed")
		},
	})
	defer holder.Clear()

	text, err := RunWithHooks(holder, "my_tool", "call-1", nil, func() (string, error) {
		t.Fatal("invoke should not run when before hook returns error")
		return "", nil
	})
	if err == nil {
		t.Fatal("expected hook error")
	}
	if err.Error() != "hook failed" {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "" {
		t.Fatalf("expected empty text on hook error, got %q", text)
	}

	result := toolResultFromText(text, err)
	if len(result.Content) != 1 {
		t.Fatalf("expected one content block, got %d", len(result.Content))
	}
}
