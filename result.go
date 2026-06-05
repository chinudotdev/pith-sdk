package pithsdk

// RunResult is the outcome of a single Session.Run call.
type RunResult struct {
	Text     string
	Messages []MessageSummary
	Usage    *UsageSummary
}

// MessageSummary is a simplified view of a transcript message.
type MessageSummary struct {
	Role string
	Text string
}

// UsageSummary reports token usage from the last assistant response.
type UsageSummary struct {
	Input  int
	Output int
	Total  int
}

// TextChunk is a streaming text delta from the assistant response.
type TextChunk struct {
	Delta string // incremental text from EventTextDelta
	Text  string // accumulated assistant text so far
}

// RunOptions configures a single run. Apply via RunOption functions.
type RunOptions struct {
	Context      any
	Instructions string
	Stream       func(chunk TextChunk) // nil = blocking
}
