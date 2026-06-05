package pithsdk

// Agent is a specialist definition: name, instructions, model, tools, and settings.
// Agents are immutable configuration; create a Session to run them.
type Agent struct {
	name         string
	instructions string
	model        string
	tools        []Tool
	settings     *ModelSettings
}

// AgentConfig configures a new Agent.
type AgentConfig struct {
	// Name is an optional display name for the agent.
	Name string
	// Instructions is the system prompt sent to the model.
	Instructions string
	// Model is a bare ID (e.g. "gpt-4o-mini") or provider/model (e.g. "anthropic/claude-...").
	Model string
	// Tools are custom tools available during runs.
	Tools []Tool
	// Settings overrides client-level generation defaults for this agent.
	Settings *ModelSettings
}

// NewAgent creates an agent definition.
func NewAgent(cfg AgentConfig) (*Agent, error) {
	return &Agent{
		name:         cfg.Name,
		instructions: cfg.Instructions,
		model:        cfg.Model,
		tools:        cfg.Tools,
		settings:     cfg.Settings,
	}, nil
}
