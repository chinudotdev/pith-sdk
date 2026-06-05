package pithsdk

import (
	"context"
	"fmt"

	pithagent "github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith-sdk/internal/stream"
	"github.com/chinudotdev/pith-sdk/internal/summary"
	"github.com/chinudotdev/pith-sdk/internal/wire"
)

// Session runs an agent and holds its transcript across multiple Run calls.
type Session struct {
	ag    *pithagent.Agent
	scope *wire.RunScopeHolder
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
func (s *Session) Run(ctx context.Context, input string, opts ...RunOption) (*RunResult, error) {
	ro := applyRunOptions(opts)

	originalPrompt := s.ag.State().SystemPrompt
	if ro.Instructions != "" {
		s.ag.SetSystemPrompt(ro.Instructions)
		defer s.ag.SetSystemPrompt(originalPrompt)
	}

	if s.scope != nil {
		s.scope.Set(ctx, ro.Context)
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
	unsubTurns := s.ag.EventBus().Subscribe(func(e pithagent.AgentEvent) {
		if e.LoopEvent == nil || e.LoopEvent.Type != loop.LoopTurnEnd {
			return
		}
		turns++
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
	state := s.ag.State()

	summaries := summary.ToSummaries(state.Messages)
	result := &RunResult{
		Text:     summary.LastAssistantText(state.Messages),
		Messages: toPublicSummaries(summaries),
		Usage:    toPublicUsage(summary.LastUsage(state.Messages)),
	}
	return result, err
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
