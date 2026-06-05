package pithsdk

// BeforeToolContext is passed to the BeforeToolCall hook.
type BeforeToolContext struct {
	// RunID is the unique identifier for the current run.
	RunID string
	// SessionID is the unique identifier for the current session.
	SessionID string
	// AgentName is the name of the agent running this tool.
	AgentName string
	// ToolName is the name of the tool being called.
	ToolName string
	// CallID is the provider-assigned tool call identifier.
	CallID string
	// Args is the parsed tool arguments.
	Args map[string]any
}

// BeforeToolResult controls tool execution from the BeforeToolCall hook.
type BeforeToolResult struct {
	// Block prevents the tool from executing when true.
	Block bool
	// Reason is the message returned when blocking a tool call.
	Reason string
}

// AfterToolContext is passed to the AfterToolCall hook.
type AfterToolContext struct {
	// RunID is the unique identifier for the current run.
	RunID string
	// SessionID is the unique identifier for the current session.
	SessionID string
	// AgentName is the name of the agent running this tool.
	AgentName string
	// ToolName is the name of the tool that was called.
	ToolName string
	// CallID is the provider-assigned tool call identifier.
	CallID string
	// Args is the parsed tool arguments.
	Args map[string]any
	// Result is the text output from the tool.
	Result string
	// Error is non-nil when the tool returned an error.
	Error error
}

// AfterToolResult controls tool result override from the AfterToolCall hook.
type AfterToolResult struct {
	// OverrideResult replaces the tool's output text when non-empty.
	OverrideResult string
}

// TurnContext is passed to the ShouldStopAfterTurn hook.
type TurnContext struct {
	// RunID is the unique identifier for the current run.
	RunID string
	// SessionID is the unique identifier for the current session.
	SessionID string
	// AgentName is the name of the running agent.
	AgentName string
	// TurnNumber is the 1-based turn count within the current run.
	TurnNumber int
}

// Hooks are lifecycle callbacks for a single Session.Run call.
// All hooks are optional; nil hooks are skipped.
type Hooks struct {
	// BeforeToolCall is called before each tool execution.
	// Return a result with Block=true to prevent execution.
	BeforeToolCall func(ctx BeforeToolContext) (*BeforeToolResult, error)

	// AfterToolCall is called after each tool execution.
	// Use OverrideResult to replace the tool's output.
	AfterToolCall func(ctx AfterToolContext) (*AfterToolResult, error)

	// ShouldStopAfterTurn is called after each agent turn.
	// Return true to stop the run after this turn.
	ShouldStopAfterTurn func(ctx TurnContext) bool
}
