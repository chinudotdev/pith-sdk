// Example 01: Hello — minimal agent run with the pith-sdk.
//
// From the repo root:
//
//	OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
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
		Instructions: "You are helpful. Be concise.",
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

	result, err := session.Run(context.Background(), "What is Go?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "run: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Text)
}
