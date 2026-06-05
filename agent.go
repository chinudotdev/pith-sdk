package pithsdk

// Agent is a specialist definition: name, instructions, model, tools, and settings.
type Agent struct {
	name         string
	instructions string
	model        string
	tools        []Tool
	settings     *ModelSettings
}

// AgentConfig configures a new Agent.
type AgentConfig struct {
	Name         string
	Instructions string
	Model        string
	Tools        []Tool
	Settings     *ModelSettings
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
