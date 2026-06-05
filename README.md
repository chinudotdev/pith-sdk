# pith-sdk

A minimal, OpenAI Agents-style SDK for Go app developers — built on [pith](https://github.com/chinudotdev/pith).

Define an agent, run it, get text back. No gateway wiring, no `ModelDescriptor`, no EventBus parsing.

## Who this is for

App developers who want `Client` → `Agent` → `Session.Run` in ~15 lines, with optional custom providers (Anthropic, Groq, Ollama, etc.).

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

### Multi-turn sessions

Reuse a `Session` to keep conversation history across runs:

```go
session, _ := client.NewSession(agent)
session.Run(ctx, "My name is Alex.")
result, _ := session.Run(ctx, "What is my name?")
fmt.Println(result.Text)
fmt.Println(len(session.Messages())) // full transcript
session.Reset()                      // start fresh
```

Stream assistant text as it arrives:

```go
session.Run(ctx, "Tell me a joke.", pithsdk.WithStream(func(c pithsdk.TextChunk) {
    fmt.Print(c.Delta)
}))
```

For one-shot scripts without managing a session:

```go
result, _ := client.RunOnce(ctx, agent, "What is Go?")
```

### Custom providers

Register any `gateway.ProviderPort` (e.g. Anthropic) once on the client:

```go
client, _ := pithsdk.NewClient(pithsdk.ClientConfig{})

client.RegisterProvider(pithsdk.ProviderRegistration{
    Provider:  myAnthropicProvider,
    APIKeyEnv: "ANTHROPIC_API_KEY",
    Models: []pithsdk.ModelPreset{
        {ID: "claude-sonnet-4-20250514", ContextWindow: 200_000, MaxTokens: 8192},
    },
})

agent, _ := pithsdk.NewAgent(pithsdk.AgentConfig{
    Model: "anthropic/claude-sonnet-4-20250514",
    // ... same Agent API as OpenAI
})
```

Use `provider/model` strings when not using the default OpenAI provider. See [examples/04-anthropic-provider](examples/04-anthropic-provider/) for a full custom provider implementation.

Limit tool-calling loop iterations per run (default 10):

```go
session.Run(ctx, input, pithsdk.WithMaxTurns(5))
```

## API overview

| Type | Role |
|------|------|
| `Client` | Gateway, credentials, defaults; creates sessions |
| `Agent` | Specialist definition: instructions, model, tools |
| `Session` | Runs the agent; holds multi-turn transcript |
| `RunOnce` | One-shot run without managing a session |
| `NewTool` | Typed tool with struct-based JSON Schema |
| `RegisterProvider` | Custom `ProviderPort` + model catalog |

## Installation

```bash
go get github.com/chinudotdev/pith-sdk@v0.1.0
```

Requires Go 1.24+.

## Examples

| Example | Description |
|---------|-------------|
| [01-hello](examples/01-hello/) | Minimal agent run |
| [02-tools](examples/02-tools/) | Agent with custom tools |
| [03-multi-turn](examples/03-multi-turn/) | Multi-turn conversation with streaming |
| [04-anthropic-provider](examples/04-anthropic-provider/) | Custom Anthropic provider via `RegisterProvider` |

Run from the repo root:

```bash
OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
OPENAI_API_KEY="sk-..." go run ./examples/02-tools/main.go
OPENAI_API_KEY="sk-..." go run ./examples/03-multi-turn/main.go
ANTHROPIC_API_KEY="sk-ant-..." go run ./examples/04-anthropic-provider/
```

## Releases

**v0.1.0** is the first public release. See [CHANGELOG.md](CHANGELOG.md) for details.

Future work is tracked in [plan.md](plan.md).

## License

MIT — see [LICENSE](LICENSE).
