package pithsdk

// BeforeToolContext is passed to the BeforeToolCall hook.
type BeforeToolContext struct {
	// RunID is the unique identifier for the current run.
	RunID string
	// SessionID is the unique identifier for the current session.
	SessionID string
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

// Hooks are lifecycle callbacks for a single Session.Run call.
// All hooks are optional; nil hooks are skipped.
//
// Hook errors: returning an error from BeforeToolCall or AfterToolCall is converted
// to tool-result text; the agent run continues with that text as the tool output.
type Hooks struct {
	// BeforeToolCall is called before each tool execution.
	// Return a result with Block=true to prevent execution.
	// Returning an error becomes tool-result text; the run continues.
	BeforeToolCall func(ctx BeforeToolContext) (*BeforeToolResult, error)

	// AfterToolCall is called after each tool execution.
	// Use OverrideResult to replace the tool's output.
	// Returning an error becomes tool-result text; the run continues.
	AfterToolCall func(ctx AfterToolContext) (*AfterToolResult, error)
}
