# pith-sdk

A minimal, OpenAI Agents-style SDK for Go app developers — built on [pith](https://github.com/chinudotdev/pith).

Define an agent, run it, get text back. No gateway wiring, no `ModelDescriptor`, no EventBus parsing.

## Who this is for

App developers who want `Client` → `Agent` → `Session.Run` in ~15 lines, with optional custom providers (Anthropic, Groq, Ollama, etc.).

## Who this is not for

Library authors and provider implementers who need direct control over the gateway, loop, or protocol layers. Use [pith](https://github.com/chinudotdev/pith) directly.

## Defended requirements

pith-sdk optimizes for a small, stable surface:

- **~15-line run path** — `Client` → `Agent` → `Session.Run` → text back
- **Hidden `pith` wiring** — no gateway, loop, or protocol imports in app code
- **Typed tools** — `NewTool[T]` with struct-based JSON Schema
- **Sessions** — multi-turn transcript via `Session.Messages()` / `Reset()`
- **Providers** — built-in OpenAI-compat plus `RegisterProvider` for custom backends
- **Hooks** — HITL approval, logging, and output redaction via `BeforeToolCall` / `AfterToolCall`
- **Tracing IDs** — `SessionID`, `RunID`, `CallID` for external observability (OpenTelemetry, Datadog, etc.)
- **MCP as tool source** — `mcp.Tools()` composes with local tools

**Escape hatches deferred for v1 review** (removed in v0.3.0; use alternatives below):

| Removed API | Replacement |
|-------------|-------------|
| `RunOnce` | `NewSession` + `Run` (two lines) |
| `RawTool` | `NewClientFromGateway` + `pith/loop` directly |
| `NewDynamicTool` | `mcp.Tools()` for schema-driven tools |
| `ShouldStopAfterTurn` | `WithMaxTurns` or app-side abort |
| `Agent.Name` | Pass labels via `WithContext` |

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

## Tracing IDs

Every session and run gets a unique ID for observability. Override or read them:

```go
session, _ := client.NewSession(agent, pithsdk.WithSessionID("my-session"))
fmt.Println(session.ID()) // "my-session"

result, _ := session.Run(ctx, "Hi", pithsdk.WithRunID("run-42"))
fmt.Println(result.RunID) // "run-42"
```

IDs are also available inside tool handlers:

```go
func(ctx pithsdk.ToolContext, args struct{ City string `json:"city"` }) (string, error) {
    fmt.Println(ctx.SessionID, ctx.RunID, ctx.ToolName, ctx.CallID)
    return "Sunny", nil
}
```

Auto-generated UUIDs are used when IDs are not provided.

## Hooks

Add lifecycle callbacks to tool execution:

```go
session.Run(ctx, "Go", pithsdk.WithHooks(pithsdk.Hooks{
    BeforeToolCall: func(ctx pithsdk.BeforeToolContext) (*pithsdk.BeforeToolResult, error) {
        if ctx.ToolName == "dangerous_tool" {
            return &pithsdk.BeforeToolResult{Block: true, Reason: "not allowed"}, nil
        }
        return nil, nil
    },
    AfterToolCall: func(ctx pithsdk.AfterToolContext) (*pithsdk.AfterToolResult, error) {
        log.Printf("tool %s returned: %s", ctx.ToolName, ctx.Result)
        return nil, nil
    },
}))
```

Returning an error from a hook becomes tool-result text; the run continues.

Hooks apply to all SDK-created tools, including typed tools (`NewTool`) and MCP-discovered tools from `mcp.Tools()`.

## MCP tools

Discover tools from MCP (Model Context Protocol) servers and compose them with local tools:

```go
import "github.com/chinudotdev/pith-sdk/mcp"

mcpTools, close, _ := mcp.Tools(ctx, mcp.Config{
    Command: "path/to/mcp-server",
    Args:    []string{"--flag"},
})
defer close()

allTools := append(localTools, mcpTools...)
```

## API overview

| Type | Role |
|------|------|
| `Client` | Gateway, credentials, defaults; creates sessions |
| `Agent` | Specialist definition: instructions, model, tools |
| `Session` | Runs the agent; holds multi-turn transcript |
| `NewTool` | Typed tool with struct-based JSON Schema |
| `mcp.Tools` | Discover tools from MCP servers |
| `RegisterProvider` | Custom `ProviderPort` + model catalog |
| `NewClientFromGateway` | Escape hatch for tests and custom setups |

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
| [03-multi-turn](examples/03-multi-turn/) | Multi-turn conversation with streaming |
| [04-anthropic-provider](examples/04-anthropic-provider/) | Custom Anthropic provider via `RegisterProvider` |
| [05-mcp](examples/05-mcp/) | Compose local and MCP tools |

Run from the repo root:

```bash
OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
OPENAI_API_KEY="sk-..." go run ./examples/02-tools/main.go
OPENAI_API_KEY="sk-..." go run ./examples/03-multi-turn/main.go
ANTHROPIC_API_KEY="sk-ant-..." go run ./examples/04-anthropic-provider/
OPENAI_API_KEY="sk-..." go run ./examples/05-mcp/main.go
```

## Releases

**v0.1.0** — Core run path, tools, sessions, providers. See [CHANGELOG.md](CHANGELOG.md) for details.

**v0.2.0** — Hooks, tracing IDs, `pithsdk/mcp`.

**v0.3.0** — API trim before v1 lock. See [CHANGELOG.md](CHANGELOG.md) for migration notes.

Future work is tracked in [plan.md](plan.md).

## License

MIT — see [LICENSE](LICENSE).
