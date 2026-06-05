package pithsdk

import "github.com/chinudotdev/pith-sdk/internal/summary"

const defaultMaxTurns = 10

// MessageSummary is a simplified view of a transcript message.
type MessageSummary = summary.MessageSummary

// RunResult is the outcome of a single Session.Run call.
type RunResult struct {
	// RunID is the unique identifier for this run, auto-generated unless overridden via WithRunID.
	RunID string
	// Text is the final assistant response text.
	Text string
	// Messages is the session transcript after this run, as simplified summaries.
	Messages []MessageSummary
	// Usage reports token usage from the last assistant response, if available.
	Usage *UsageSummary
}

// UsageSummary reports token usage from the last assistant response.
type UsageSummary struct {
	// Input is the number of input/prompt tokens.
	Input int
	// Output is the number of output/completion tokens.
	Output int
	// Total is the combined token count when reported by the provider.
	Total int
}

// TextChunk is a streaming text delta from the assistant response.
type TextChunk struct {
	// Delta is the incremental text from the latest EventTextDelta.
	Delta string
	// Text is the accumulated assistant text so far.
	Text string
}

// RunOptions configures a single run. Apply via RunOption functions.
type RunOptions struct {
	// RunID overrides the auto-generated run identifier. Empty triggers auto-generation.
	RunID string
	// Context holds run-scoped local dependencies exposed as ToolContext.Local.
	Context any
	// Instructions overrides the agent system prompt for this run only.
	Instructions string
	// Stream receives assistant text deltas during the run; nil blocks until complete.
	Stream func(chunk TextChunk)
	// MaxTurns limits tool-calling loop iterations for this run. Zero uses defaultMaxTurns (10).
	MaxTurns int
	// Hooks are lifecycle callbacks for this run.
	Hooks *Hooks
}
