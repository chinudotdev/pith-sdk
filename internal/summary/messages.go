package summary

import (
	"strings"

	"github.com/chinudotdev/pith/protocol"
)

// MessageSummary is a simplified transcript entry.
type MessageSummary struct {
	Role string
	Text string
}

// UsageSummary reports token usage.
type UsageSummary struct {
	Input  int
	Output int
	Total  int
}

// ToSummaries converts protocol messages to simplified summaries.
func ToSummaries(messages []protocol.Message) []MessageSummary {
	out := make([]MessageSummary, 0, len(messages))
	for _, msg := range messages {
		role := protocol.MessageRole(msg)
		text := messageText(msg)
		out = append(out, MessageSummary{Role: role, Text: text})
	}
	return out
}

// LastAssistantText returns the text from the last assistant message.
func LastAssistantText(messages []protocol.Message) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if am, ok := messages[i].(protocol.AssistantMessage); ok {
			return assistantText(am)
		}
	}
	return ""
}

// LastUsage returns usage from the last assistant message, if any.
func LastUsage(messages []protocol.Message) *UsageSummary {
	for i := len(messages) - 1; i >= 0; i-- {
		if am, ok := messages[i].(protocol.AssistantMessage); ok {
			return &UsageSummary{
				Input:  am.Usage.Input,
				Output: am.Usage.Output,
				Total:  am.Usage.TotalTokens,
			}
		}
	}
	return nil
}

func messageText(msg protocol.Message) string {
	switch m := msg.(type) {
	case protocol.UserMessage:
		return contentsText(m.Content)
	case protocol.AssistantMessage:
		return assistantText(m)
	case protocol.ToolResultMessage:
		return contentsText(m.Content)
	default:
		return ""
	}
}

// AssistantText returns the text content from an assistant message.
func AssistantText(am protocol.AssistantMessage) string {
	return assistantText(am)
}

func assistantText(am protocol.AssistantMessage) string {
	var parts []string
	for _, block := range am.Content {
		if tc, ok := block.(protocol.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "")
}

func contentsText(content []protocol.Content) string {
	var parts []string
	for _, c := range content {
		if tc, ok := c.(protocol.TextContent); ok {
			parts = append(parts, tc.Text)
		}
	}
	return strings.Join(parts, "")
}
