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

// RunOptions configures a single run. Apply via RunOption functions.
type RunOptions struct {
	Instructions string
}
