package pithsdk

import "fmt"

// Tool is an opaque tool definition. Use NewTool in a future release.
type Tool struct{}

// Agent is a specialist definition: name, instructions, model, and settings.
type Agent struct {
	name         string
	instructions string
	model        string
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

// NewAgent creates an agent definition. Tools are not supported in v0.1 yet.
func NewAgent(cfg AgentConfig) (*Agent, error) {
	if len(cfg.Tools) > 0 {
		return nil, fmt.Errorf("tools are not supported yet")
	}
	return &Agent{
		name:         cfg.Name,
		instructions: cfg.Instructions,
		model:        cfg.Model,
		settings:     cfg.Settings,
	}, nil
}
