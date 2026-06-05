package pithsdk

// SessionOption configures session creation via Client.NewSession.
type SessionOption func(*sessionOpts)

type sessionOpts struct {
	sessionID string
}

// WithSessionID sets an explicit session identifier. When omitted, a UUID is generated.
func WithSessionID(id string) SessionOption {
	return func(o *sessionOpts) {
		o.sessionID = id
	}
}

func applySessionOptions(opts []SessionOption) sessionOpts {
	var so sessionOpts
	for _, opt := range opts {
		if opt != nil {
			opt(&so)
		}
	}
	return so
}

// RunOption configures a Session.Run call.
type RunOption func(*RunOptions)

// WithRunID sets an explicit run identifier. When omitted, a UUID is generated per Run.
func WithRunID(id string) RunOption {
	return func(o *RunOptions) {
		o.RunID = id
	}
}

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

// WithHooks registers lifecycle callbacks for a single run.
func WithHooks(h Hooks) RunOption {
	return func(o *RunOptions) {
		o.Hooks = &h
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
