# pith-sdk

A minimal, OpenAI Agents-style SDK for Go app developers — built on [pith](https://github.com/chinudotdev/pith).

Define an agent, run it, get text back. No gateway wiring, no `ModelDescriptor`, no EventBus parsing.

## Who this is for

App developers who want `Client` → `Agent` → `Session.Run` in ~15 lines, with optional custom providers (Anthropic, Groq, Ollama, etc.).

## Who this is not for

Library authors and provider implementers who need direct control over the gateway, loop, or protocol layers. Use [pith](https://github.com/chinudotdev/pith) directly.

## Status

Pre-release scaffold. The public API is not yet implemented. See [plan.md](plan.md) for the full roadmap.

## Planned API (v0.1.0)

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

    result, _ := client.NewSession(agent).Run(context.Background(), "What is Go?")
    fmt.Println(result.Text)
}
```

## Installation

Coming in v0.1.0:

```bash
go get github.com/chinudotdev/pith-sdk@v0.1.0
```

## Requirements

Go 1.24+

## License

MIT — see [LICENSE](LICENSE).
