package stream

import (
	"github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith/protocol"
	"github.com/chinudotdev/pith-sdk/internal/summary"
)

// SubscribeTextDeltas registers an EventBus listener for assistant text deltas.
// Returns an unsubscribe function that must be called when the run completes.
func SubscribeTextDeltas(bus *agent.EventBus, onDelta func(delta, accumulated string)) func() {
	return bus.Subscribe(func(event agent.AgentEvent) {
		if event.LoopEvent == nil || event.LoopEvent.Type != loop.LoopMessageUpdate {
			return
		}
		se := event.LoopEvent.StreamEvent
		if se == nil || se.Type != protocol.EventTextDelta {
			return
		}
		accumulated := ""
		if se.Partial != nil {
			accumulated = summary.AssistantText(*se.Partial)
		}
		onDelta(se.Delta, accumulated)
	})
}
