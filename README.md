# pith-sdk

A minimal, OpenAI Agents-style SDK for Go app developers — built on [pith](https://github.com/chinudotdev/pith).

Define an agent, run it, get text back. No gateway wiring, no `ModelDescriptor`, no EventBus parsing.

## Who this is for

App developers who want `Client` → `Agent` → `Session.Run` in ~15 lines, with optional custom providers (Anthropic, Groq, Ollama, etc.) coming in later releases.

## Who this is not for

Library authors and provider implementers who need direct control over the gateway, loop, or protocol layers. Use [pith](https://github.com/chinudotdev/pith) directly.

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "os"

    pithsdk "github.com/chinudotdev/pith-sdk"
)

func main() {
    client, _ := pithsdk.NewClient(pithsdk.ClientConfig{
        APIKey: os.Getenv("OPENAI_API_KEY"),
    })

    agent, _ := pithsdk.NewAgent(pithsdk.AgentConfig{
        Name:         "Assistant",
        Instructions: "You are helpful. Be concise.",
        Model:        "gpt-4o-mini",
    })

    session, _ := client.NewSession(agent)
    result, _ := session.Run(context.Background(), "What is Go?")
    fmt.Println(result.Text)
}
```

### With tools

```go
weather := pithsdk.NewTool("get_weather", "Return weather for a city.",
    func(ctx pithsdk.ToolContext, args struct {
        City string `json:"city"`
    }) (string, error) {
        return fmt.Sprintf("Sunny in %s", args.City), nil
    },
)

agent, _ := pithsdk.NewAgent(pithsdk.AgentConfig{
    Name:         "Weather bot",
    Instructions: "You are a helpful weather bot.",
    Model:        "gpt-4o-mini",
    Tools:        []pithsdk.Tool{weather},
})
```

Pass run-scoped dependencies to tools with `WithContext`:

```go
session.Run(ctx, "What's the weather?", pithsdk.WithContext(myDeps))
```

## Installation

```bash
go get github.com/chinudotdev/pith-sdk@latest
```

Requires Go 1.24+.

## Examples

| Example | Description |
|---------|-------------|
| [01-hello](examples/01-hello/) | Minimal agent run |
| [02-tools](examples/02-tools/) | Agent with custom tools |

Run from the repo root:

```bash
OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
OPENAI_API_KEY="sk-..." go run ./examples/02-tools/main.go
```

## Roadmap

See [plan.md](plan.md) for the full implementation plan. Upcoming:

- **Phase 3:** Multi-turn sessions, streaming, `RunOnce`
- **Phase 4:** `RegisterProvider` and custom providers

## License

MIT — see [LICENSE](LICENSE).
