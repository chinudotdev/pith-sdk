package summary_test

import (
	"testing"
	"time"

	"github.com/chinudotdev/pith/protocol"
	"github.com/chinudotdev/pith-sdk/internal/summary"
)

func TestLastAssistantText(t *testing.T) {
	messages := []protocol.Message{
		protocol.UserMessage{
			Role:      "user",
			Content:   []protocol.Content{protocol.TextContent{Type: "text", Text: "Hi"}},
			Timestamp: time.Now(),
		},
		protocol.AssistantMessage{
			Role:      "assistant",
			Content:   []protocol.ContentBlock{protocol.TextContent{Type: "text", Text: "Hello!"}},
			Timestamp: time.Now(),
			Usage:     protocol.Usage{Input: 5, Output: 3, TotalTokens: 8},
		},
	}

	text := summary.LastAssistantText(messages)
	if text != "Hello!" {
		t.Fatalf("expected Hello!, got %q", text)
	}

	usage := summary.LastUsage(messages)
	if usage == nil || usage.Input != 5 || usage.Output != 3 || usage.Total != 8 {
		t.Fatalf("unexpected usage: %+v", usage)
	}

	summaries := summary.ToSummaries(messages)
	if len(summaries) != 2 || summaries[0].Role != "user" || summaries[1].Text != "Hello!" {
		t.Fatalf("unexpected summaries: %+v", summaries)
	}
}
