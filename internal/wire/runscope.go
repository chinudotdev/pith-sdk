package wire

import (
	"context"
	"sync"
)

// RunScope holds per-run context passed to tool handlers.
type RunScope struct {
	Ctx       context.Context
	Local     any
	SessionID string
	RunID     string
	Hooks     *HookSet
}

// HookSet holds SDK-level hooks for tool call lifecycle.
type HookSet struct {
	BeforeToolCall func(sessionID, runID, toolName, callID string, args map[string]any) (block bool, reason string, err error)
	AfterToolCall  func(sessionID, runID, toolName, callID string, args map[string]any, result string, resultErr error) (override string, err error)
}

// RunScopeHolder stores the active run scope for tool execution.
type RunScopeHolder struct {
	mu    sync.RWMutex
	scope *RunScope
}

// NewRunScopeHolder creates a holder for per-run tool context.
func NewRunScopeHolder() *RunScopeHolder {
	return &RunScopeHolder{}
}

// Set activates a run scope for the current Session.Run call.
func (h *RunScopeHolder) Set(ctx context.Context, local any, sessionID, runID string, hooks *HookSet) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.scope = &RunScope{Ctx: ctx, Local: local, SessionID: sessionID, RunID: runID, Hooks: hooks}
}

// Clear removes the active run scope after a run completes.
func (h *RunScopeHolder) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.scope = nil
}

// Current returns the active run scope, or nil if none is set.
func (h *RunScopeHolder) Current() *RunScope {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.scope
}
