package stream_test

import (
	"testing"

	"github.com/chinudotdev/pith/agent"
	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith/protocol"
	"github.com/chinudotdev/pith-sdk/internal/stream"
)

func TestSubscribeTextDeltas(t *testing.T) {
	bus := agent.NewEventBus()
	var deltas []string
	var accumulated []string

	unsub := stream.SubscribeTextDeltas(bus, func(delta, acc string) {
		deltas = append(deltas, delta)
		accumulated = append(accumulated, acc)
	})
	defer unsub()

	partial := &protocol.AssistantMessage{
		Role: "assistant",
		Content: []protocol.ContentBlock{
			protocol.TextContent{Type: "text", Text: "Hel"},
		},
	}
	bus.Emit(agent.AgentEvent{
		LoopEvent: &loop.LoopEvent{
			Type: loop.LoopMessageUpdate,
			StreamEvent: &protocol.StreamEvent{
				Type:    protocol.EventTextDelta,
				Delta:   "Hel",
				Partial: partial,
			},
		},
	})

	if len(deltas) != 1 || deltas[0] != "Hel" {
		t.Fatalf("expected delta Hel, got %v", deltas)
	}
	if len(accumulated) != 1 || accumulated[0] != "Hel" {
		t.Fatalf("expected accumulated Hel, got %v", accumulated)
	}

	// Non-text-delta events should be ignored.
	bus.Emit(agent.AgentEvent{
		LoopEvent: &loop.LoopEvent{
			Type: loop.LoopMessageUpdate,
			StreamEvent: &protocol.StreamEvent{
				Type: protocol.EventTextStart,
			},
		},
	})
	if len(deltas) != 1 {
		t.Fatalf("expected no extra deltas, got %v", deltas)
	}
}
