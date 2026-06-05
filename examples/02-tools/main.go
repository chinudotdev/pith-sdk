// Example 02: Tools — agent with a custom tool the model can call.
//
// From the repo root:
//
//	OPENAI_API_KEY="sk-..." go run ./examples/02-tools/main.go
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

	weather := pithsdk.NewTool("get_weather", "Return weather for a city.",
		func(ctx pithsdk.ToolContext, args struct {
			City string `json:"city" desc:"City name"`
		}) (string, error) {
			return fmt.Sprintf("Sunny in %s", args.City), nil
		},
	)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are a helpful weather bot. Use get_weather when asked about weather.",
		Model:        "gpt-4o-mini",
		Tools:        []pithsdk.Tool{weather},
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

	result, err := session.Run(context.Background(), "What's the weather in San Francisco?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "run: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Text)
}
