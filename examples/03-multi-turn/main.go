// Example 03: Multi-turn — conversation across multiple Session.Run calls.
//
// From the repo root:
//
//	OPENAI_API_KEY="sk-..." go run ./examples/03-multi-turn/main.go
//
// Or copy into a new module:
//
//	mkdir my-agent && cd my-agent && go mod init my-agent
//	cp main.go . && go get github.com/chinudotdev/pith-sdk@latest
//	OPENAI_API_KEY="sk-..." go run main.go
package main

import (
	"context"
	"fmt"
	"os"

	pithsdk "github.com/chinudotdev/pith-sdk"
)

func main() {
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "client: %v\n", err)
		os.Exit(1)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Name:         "Assistant",
		Instructions: "You are helpful. Remember details the user shares in the conversation.",
		Model:        "gpt-4o-mini",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent: %v\n", err)
		os.Exit(1)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "session: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	if _, err := session.Run(ctx, "My name is Alex."); err != nil {
		fmt.Fprintf(os.Stderr, "first run: %v\n", err)
		os.Exit(1)
	}

	fmt.Print("Assistant: ")
	result, err := session.Run(ctx, "What is my name?", pithsdk.WithStream(func(c pithsdk.TextChunk) {
		fmt.Print(c.Delta)
	}))
	if err != nil {
		fmt.Fprintf(os.Stderr, "second run: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	fmt.Printf("\nTranscript (%d messages):\n", len(session.Messages()))
	for _, msg := range session.Messages() {
		fmt.Printf("  [%s] %s\n", msg.Role, msg.Text)
	}

	if result.Text == "" {
		fmt.Fprintln(os.Stderr, "expected non-empty response")
		os.Exit(1)
	}
}
