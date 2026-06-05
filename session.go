package pithsdk

import (
	"context"
	"crypto/rand"
	"fmt"

	pithagent "github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith-sdk/internal/stream"
	"github.com/chinudotdev/pith-sdk/internal/summary"
	"github.com/chinudotdev/pith-sdk/internal/wire"
)

// newUUID generates a v4 UUID using crypto/rand.
func newUUID() string {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		panic("pithsdk: crypto/rand.Read failed: " + err.Error())
	}
	buf[6] = (buf[6] & 0x0f) | 0x40 // version 4
	buf[8] = (buf[8] & 0x3f) | 0x80 // variant 10
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])
}

// Session runs an agent and holds its transcript across multiple Run calls.
type Session struct {
	id        string
	agentName string
	ag        *pithagent.Agent
	scope     *wire.RunScopeHolder
}

// ID returns the session identifier, unique per conversation.
func (s *Session) ID() string {
	return s.id
}

// Messages returns the current session transcript.
func (s *Session) Messages() []MessageSummary {
	state := s.ag.State()
	return toPublicSummaries(summary.ToSummaries(state.Messages))
}

// Reset clears the session transcript and queued messages.
func (s *Session) Reset() {
	s.ag.Reset()
}

// Run sends input to the agent and returns the final text result.
//
// Concurrent Run calls on the same Session are not supported; serialize calls
// or use separate sessions.
//
// When ShouldStopAfterTurn returns true, Run stops gracefully and returns nil
// error along with partial results accumulated so far.
func (s *Session) Run(ctx context.Context, input string, opts ...RunOption) (*RunResult, error) {
	ro := applyRunOptions(opts)

	runID := ro.RunID
	if runID == "" {
		runID = newUUID()
	}

	originalPrompt := s.ag.State().SystemPrompt
	if ro.Instructions != "" {
		s.ag.SetSystemPrompt(ro.Instructions)
		defer s.ag.SetSystemPrompt(originalPrompt)
	}

	var hookSet *wire.HookSet
	if ro.Hooks != nil {
		hookSet = buildHookSet(ro.Hooks, s.id, runID, s.agentName)
	}

	if s.scope != nil {
		s.scope.Set(ctx, ro.Context, s.id, runID, hookSet)
		defer s.scope.Clear()
	}

	if ro.Stream != nil {
		unsub := stream.SubscribeTextDeltas(s.ag.EventBus(), func(delta, accumulated string) {
			ro.Stream(TextChunk{Delta: delta, Text: accumulated})
		})
		defer unsub()
	}

	maxTurns := ro.MaxTurns
	if maxTurns <= 0 {
		maxTurns = defaultMaxTurns
	}

	var turns int
	var maxTurnsExceeded bool
	var hookStopped bool
	unsubTurns := s.ag.EventBus().Subscribe(func(e pithagent.AgentEvent) {
		if e.LoopEvent == nil || e.LoopEvent.Type != loop.LoopTurnEnd {
			return
		}
		turns++

		if ro.Hooks != nil && ro.Hooks.ShouldStopAfterTurn != nil {
			if ro.Hooks.ShouldStopAfterTurn(TurnContext{
				RunID:      runID,
				SessionID:  s.id,
				AgentName:  s.agentName,
				TurnNumber: turns,
			}) {
				hookStopped = true
				s.ag.Abort()
				return
			}
		}

		if turns >= maxTurns {
			maxTurnsExceeded = true
			s.ag.Abort()
		}
	})
	defer unsubTurns()

	err := s.ag.Prompt(ctx, input)
	if maxTurnsExceeded {
		err = fmt.Errorf("max turns (%d) exceeded", maxTurns)
	}
	if hookStopped {
		err = nil
	}
	state := s.ag.State()

	summaries := summary.ToSummaries(state.Messages)
	result := &RunResult{
		RunID:    runID,
		Text:     summary.LastAssistantText(state.Messages),
		Messages: toPublicSummaries(summaries),
		Usage:    toPublicUsage(summary.LastUsage(state.Messages)),
	}
	return result, err
}

func buildHookSet(h *Hooks, sessionID, runID, agentName string) *wire.HookSet {
	hs := &wire.HookSet{}
	if h.BeforeToolCall != nil {
		hs.BeforeToolCall = func(sid, rid, an, toolName, callID string, args map[string]any) (bool, string, error) {
			res, err := h.BeforeToolCall(BeforeToolContext{
				RunID:     rid,
				SessionID: sid,
				AgentName: an,
				ToolName:  toolName,
				CallID:    callID,
				Args:      args,
			})
			if err != nil {
				return false, "", err
			}
			if res != nil {
				return res.Block, res.Reason, nil
			}
			return false, "", nil
		}
	}
	if h.AfterToolCall != nil {
		hs.AfterToolCall = func(sid, rid, an, toolName, callID string, args map[string]any, result string, resultErr error) (string, error) {
			res, err := h.AfterToolCall(AfterToolContext{
				RunID:     rid,
				SessionID: sid,
				AgentName: an,
				ToolName:  toolName,
				CallID:    callID,
				Args:      args,
				Result:    result,
				Error:     resultErr,
			})
			if err != nil {
				return "", err
			}
			if res != nil {
				return res.OverrideResult, nil
			}
			return "", nil
		}
	}
	return hs
}

func toPublicSummaries(in []summary.MessageSummary) []MessageSummary {
	out := make([]MessageSummary, len(in))
	for i, m := range in {
		out[i] = MessageSummary{Role: m.Role, Text: m.Text}
	}
	return out
}

func toPublicUsage(in *summary.UsageSummary) *UsageSummary {
	if in == nil {
		return nil
	}
	return &UsageSummary{
		Input:  in.Input,
		Output: in.Output,
		Total:  in.Total,
	}
}
