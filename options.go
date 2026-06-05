package pithsdk

// RunOption configures a Session.Run call.
type RunOption func(*RunOptions)

// WithInstructions overrides the agent's system prompt for this run only.
func WithInstructions(instructions string) RunOption {
	return func(o *RunOptions) {
		o.Instructions = instructions
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
