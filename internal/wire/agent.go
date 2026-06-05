package wire

import (
	"github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
)

func thinkingOff() *protocol.ThinkingLevel {
	level := protocol.ThinkingOff
	return &level
}

// NewAgent builds a pith agent wired to the given gateway and model.
func NewAgent(gw *gateway.LLMGateway, model protocol.ModelDescriptor, instructions string, settings Settings) *agent.Agent {
	return agent.NewAgent(agent.AgentConfig{
		InitialState: &agent.AgentState{
			Model:         model,
			SystemPrompt:  instructions,
			ThinkingLevel: thinkingOff(),
		},
		StreamFn: NewStreamFn(gw, settings),
	})
}
