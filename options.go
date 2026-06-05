package pithsdk

// RunOption configures a Session.Run call.
type RunOption func(*RunOptions)

// WithContext sets run-scoped local dependencies available as ToolContext.Local.
func WithContext(local any) RunOption {
	return func(o *RunOptions) {
		o.Context = local
	}
}

// WithInstructions overrides the agent's system prompt for this run only.
func WithInstructions(instructions string) RunOption {
	return func(o *RunOptions) {
		o.Instructions = instructions
	}
}

// WithStream registers a callback for streaming assistant text deltas during the run.
// When nil (default), Run blocks until the full response is ready.
func WithStream(fn func(TextChunk)) RunOption {
	return func(o *RunOptions) {
		o.Stream = fn
	}
}

// WithMaxTurns limits how many agent loop turns a single Run may take.
// Zero uses the SDK default of 10.
func WithMaxTurns(n int) RunOption {
	return func(o *RunOptions) {
		o.MaxTurns = n
	}
}

func applyRunOptions(opts []RunOption) RunOptions {
	var ro RunOptions
	for _, opt := range opts {
		if opt != nil {
			opt(&ro)
		}
	}
	return ro
}
