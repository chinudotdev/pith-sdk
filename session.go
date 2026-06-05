package pithsdk

import (
	"context"

	pithagent "github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith-sdk/internal/summary"
)

// Session runs an agent and holds its transcript for the current session.
type Session struct {
	ag *pithagent.Agent
}

// Run sends input to the agent and returns the final text result.
func (s *Session) Run(ctx context.Context, input string, opts ...RunOption) (*RunResult, error) {
	ro := applyRunOptions(opts)

	originalPrompt := s.ag.State().SystemPrompt
	if ro.Instructions != "" {
		s.ag.SetSystemPrompt(ro.Instructions)
		defer s.ag.SetSystemPrompt(originalPrompt)
	}

	err := s.ag.Prompt(ctx, input)
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
