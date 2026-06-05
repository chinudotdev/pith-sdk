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

## Installation

```bash
go get github.com/chinudotdev/pith-sdk@latest
```

Requires Go 1.24+.

## Examples

| Example | Description |
|---------|-------------|
| [01-hello](examples/01-hello/) | Minimal agent run |

Run from the repo root:

```bash
OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
```

## Roadmap

See [plan.md](plan.md) for the full implementation plan. Upcoming:

- **Phase 2:** `NewTool[T]` and tool execution
- **Phase 3:** Multi-turn sessions, streaming, `RunOnce`
- **Phase 4:** `RegisterProvider` and custom providers

## License

MIT — see [LICENSE](LICENSE).
