# pith-sdk — Implementation Plan

This document is the blueprint for **`github.com/chinudotdev/pith-sdk`**, a high-level Go SDK built on top of the stable [Pith](https://github.com/chinudotdev/pith) primitives.

Copy this file into the new repo root when you create it.

---

## 1. Goals

### What pith-sdk is

A **minimal, OpenAI Agents-style SDK** for Go app developers:

- Define an agent (name, instructions, model, tools)
- Run it (`Session.Run` → get text back)
- Register custom providers (e.g. Anthropic via `ProviderPort`)
- No gateway wiring, no `ModelDescriptor`, no EventBus parsing

### What pith-sdk is not

- Not a full agent framework (no graphs, Temporal, MCP, handoffs on v1)
- Not a replacement for `pith` primitives — it wraps them
- Not OpenAI-locked — provider-neutral by design

### Success criteria for v1

A developer can go from zero to a tool-calling agent in **~15 lines**, using either OpenAI-compat or a custom Anthropic provider, without importing `protocol`, `gateway`, or `loop` directly.

---

## 2. Repository split

| Repo | Module root | Audience | Change policy |
|------|-------------|----------|---------------|
| **`chinudotdev/pith`** | `github.com/chinudotdev/pith/{protocol,loop,gateway,agent}` | Library authors, provider implementers | Slow — bugs, security, intentional breaking changes only |
| **`chinudotdev/pith-sdk`** | `github.com/chinudotdev/pith-sdk` | App developers | Fast — features, DX, examples, presets |

### Dependency direction

```
App code
  → pith-sdk
    → pith/agent, pith/gateway, pith/loop, pith/protocol
```

**Never** the reverse. Primitives never import pith-sdk.

### Version coupling

```go
// pith-sdk/go.mod (published)
require (
    github.com/chinudotdev/pith/agent    v0.1.x
    github.com/chinudotdev/pith/gateway  v0.1.x
    github.com/chinudotdev/pith/loop     v0.1.x
    github.com/chinudotdev/pith/protocol v0.1.x
)
```

- SDK patch/minor releases: no primitive changes required
- Primitive breaking change (future v1): SDK pins old primitive version until migrated, then SDK major bump
- Local dev: use `replace` in go.mod or a `go.work` spanning both repos (strip before release)

---

## 3. Positioning

### Comparable projects

| Project | Relationship |
|---------|--------------|
| [openai-agents-go](https://github.com/nlpodyssey/openai-agents-go) | Closest API shape — Go port of OpenAI Agents SDK |
| [Gollem](https://github.com/fugue-labs/gollem) | Typed tools + multi-provider — more features, monolithic |
| [Genkit Go](https://github.com/genkit-ai/genkit/tree/main/go) | Good DX — heavier platform (CLI, Firebase) |
| [langchaingo](https://github.com/tmc/langchaingo) | Chains/executors — older pattern |
| [agent-sdk-go](https://github.com/agenticenv/agent-sdk-go) | Production SDK — Temporal-required |

### pith-sdk differentiator

1. **Smallest API surface** — Agent, Session, Run, Tool only on v1
2. **Separate stable primitive core** — drop to `pith/*` when needed
3. **First-class custom providers** — `RegisterProvider` + `gateway.ProviderPort` (see pith example 11)
4. **Provider-neutral defaults** — OpenAI-compat, Groq, Ollama via BaseURL; not OpenAI Responses-first

---

## 4. Architecture

```
┌─────────────────────────────────────────┐
│  App code                               │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  pith-sdk                               │
│  Client · Agent · Session · Tool · Run  │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  pith/agent + pith/gateway              │
│  (wired internally — not exposed)       │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  pith/loop + pith/protocol            │
└─────────────────────────────────────────┘
```

### Internal wiring (SDK owns this)

On `Client` creation / `RegisterProvider`:

1. Create `gateway.LLMGateway`
2. Register built-in OpenAI-compat provider (default)
3. Register custom providers via `RegisterProvider`
4. Wire credentials per provider
5. Register models from `ModelPreset` → internal `ModelDescriptor`
6. On `Session.Run`: call `agent.Prompt` with pre-built `StreamFn`

App developers never write `StreamFn`, `CredentialProviderFunc`, or `Catalog.Register` by hand.

---

## 5. Public API (v1)

### Design rules (API stability)

1. **`Agent` stays small** — only specialist identity (name, instructions, model, tools, settings)
2. **Run behavior on `Session` / `RunOptions`** — not on Agent
3. **Infrastructure on `Client`** — keys, providers, defaults
4. **String model IDs only** — never expose `ModelDescriptor` in public SDK docs
5. **One multi-turn path** — `Session` (not ambiguous history params on day one)
6. **Escape hatch** — `NewClientFromGateway(gw *gateway.LLMGateway)` for power users

---

### 5.1 Client

```go
package pithsdk

type Client struct { /* ... */ }

type ClientConfig struct {
    APIKey          string         // explicit key for default provider
    DefaultProvider string         // "openai" (default)
    DefaultModel    string         // fallback if Agent.Model is empty, e.g. "gpt-4o-mini"
    DefaultSettings *ModelSettings
}

func NewClient(cfg ClientConfig) (*Client, error)

// Advanced: pre-wired gateway (e.g. existing example 11 setup)
func NewClientFromGateway(gw *gateway.LLMGateway) *Client

func (c *Client) RegisterProvider(reg ProviderRegistration) error
func (c *Client) NewSession(agent *Agent) *Session
```

---

### 5.2 Agent (definition only)

```go
type Agent struct { /* opaque */ }

type AgentConfig struct {
    Name         string
    Instructions string
    Model        string         // "gpt-4o-mini" or "anthropic/claude-sonnet-4-20250514"
    Tools        []Tool
    Settings     *ModelSettings // optional
}

func NewAgent(cfg AgentConfig) (*Agent, error)
```

| Field | v1 | Notes |
|-------|-----|-------|
| `Name` | Yes | Traces, logs, future handoffs |
| `Instructions` | Yes | Static system prompt |
| `Model` | Yes | Provider/model shorthand — single selector, no separate Provider field |
| `Tools` | Yes | From `NewTool[T]` |
| `Settings` | Yes | Temperature, max tokens only |

**Do not put on Agent:** APIKey, StreamFn, Provider, handoffs, guardrails, MCP, output schema.

---

### 5.3 ModelSettings

```go
type ModelSettings struct {
    Temperature *float64
    MaxTokens   *int
}
```

Defer: thinking/reasoning, transport, parallel tool calls, full capabilities maps.

---

### 5.4 Model string convention

| String | Resolves to |
|--------|-------------|
| `"gpt-4o-mini"` | `DefaultProvider` (openai compat) + model ID |
| `"groq/llama-3.3-70b"` | Provider `groq` + model ID |
| `"anthropic/claude-sonnet-4-20250514"` | Registered custom provider + model ID |
| `"ollama/llama3"` | Provider `ollama` + model ID |

Format: `"<provider>/<model-id>"` when not using default provider.

Clear errors when provider or model is missing:

```
unknown model "anthropic/claude-foo": provider "anthropic" registered but model not found
```

---

### 5.5 Tool

```go
type Tool struct { /* opaque */ }

type ToolContext struct {
    Run    context.Context // cancellation
    Local  any             // from RunOptions.Context — NOT sent to model
    CallID string
}

func NewTool[T any](
    name string,
    description string,
    fn func(ToolContext, T) (string, error),
) Tool

// Advanced escape hatch
func RawTool(t loop.AgentTool) Tool
```

- v1: struct tags (`json`, optional `desc`) → JSON Schema generation
- Defer on Tool: needsApproval, guardrails, timeoutMs, deferLoading

---

### 5.6 Session + Run

```go
type Session struct { /* wraps pith agent + transcript */ }

func (s *Session) Run(ctx context.Context, input string, opts ...RunOption) (*RunResult, error)
func (s *Session) Messages() []MessageSummary
func (s *Session) Reset()
```

```go
type RunOptions struct {
    Context      any                        // local app context → ToolContext.Local
    Stream       func(chunk TextChunk)      // nil = blocking
    MaxTurns     int                        // default 10
    Instructions string                     // one-off override
}

type RunResult struct {
    Text     string              // OpenAI: finalOutput
    Messages []MessageSummary
    Usage    *UsageSummary
}
```

```go
// Optional sugar for scripts (stateless single shot)
func (c *Client) RunOnce(ctx context.Context, agent *Agent, input string, opts ...RunOption) (*RunResult, error)
```

---

### 5.7 Custom provider registration

Provider authors implement `gateway.ProviderPort` (see pith `examples/11-custom-provider`).

SDK users register once on Client:

```go
type ProviderRegistration struct {
    Provider gateway.ProviderPort

    // Credentials — one of:
    APIKey     string
    APIKeyEnv  string                              // e.g. "ANTHROPIC_API_KEY"
    Credential func(providerID string) (string, error)

    Models []ModelPreset
}

type ModelPreset struct {
    ID            string // "claude-sonnet-4-20250514"
    Name          string // optional display name
    ContextWindow int    // 0 = SDK default
    MaxTokens     int    // 0 = SDK default
}
```

Usage:

```go
client.RegisterProvider(pithsdk.ProviderRegistration{
    Provider:  anthropic.New(anthropic.Config{}),
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

SDK internally: register provider → wire credentials → build ModelDescriptors → done.

---

## 6. v1 scope — include vs defer

### Include in v1

- [ ] `Client`, `Agent`, `Session`, `Run`, `RunOnce`
- [ ] Built-in OpenAI-compat provider (auto on `NewClient`)
- [ ] `RegisterProvider` for custom providers
- [ ] `NewTool[T]` with struct-based schema
- [ ] `ToolContext.Local` for run-scoped deps
- [ ] Model string shorthand + `ModelPreset`
- [ ] Env-based credentials (`APIKey`, `APIKeyEnv`)
- [ ] `RunOptions`: stream callback, maxTurns, local context, instructions override
- [ ] `RunResult.Text` + simplified message/usage summaries
- [ ] `Session.Reset()`, `Session.Messages()`
- [ ] `NewClientFromGateway` escape hatch
- [ ] Examples: hello, tools, multi-turn, custom anthropic provider
- [ ] CI: vet + test (no live API keys required for unit tests)

### Defer to v2+

| Feature | Notes |
|---------|-------|
| Handoffs / multi-agent | New type (`Orchestrator`) — don't add to Agent |
| Structured output (`outputType`) | Pick schema approach first |
| Dynamic instructions callback | `RunOptions.InstructionsFunc` |
| Guardrails / HITL / approvals | Hooks exist in primitives |
| MCP servers | Integration surface still evolving |
| Built-in tools (bash, read_file) | Optional; start with custom tools only |
| Compaction helper | `Session.CompactIfNeeded()` — primitive exists |
| Tracing / observability | Separate concern |
| Hosted platform tools | OpenAI-specific |
| Provider preset packages | `pith-sdk/providers/anthropic` when maintained |
| `client.Use(bundle)` sugar | After first-party provider packages exist |

---

## 7. Target hello worlds

### OpenAI (default)

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/chinudotdev/pith-sdk"
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

### Custom Anthropic provider

```go
client, _ := pithsdk.NewClient(pithsdk.ClientConfig{})

client.RegisterProvider(pithsdk.ProviderRegistration{
    Provider:  myanthropic.New(myanthropic.Config{}),
    APIKeyEnv: "ANTHROPIC_API_KEY",
    Models: []pithsdk.ModelPreset{
        {ID: "claude-sonnet-4-20250514", ContextWindow: 200_000, MaxTokens: 8192},
    },
})

agent, _ := pithsdk.NewAgent(pithsdk.AgentConfig{
    Name:         "Assistant",
    Instructions: "You are helpful.",
    Model:        "anthropic/claude-sonnet-4-20250514",
})
```

---

## 8. New repo structure

```
pith-sdk/
├── plan.md                 # this file
├── README.md               # app developer focused; link to pith for primitives
├── LICENSE                 # same as pith (MIT)
├── go.mod                  # module github.com/chinudotdev/pith-sdk
├── go.sum
├── .github/
│   └── workflows/
│       └── ci.yml          # go vet + go test
├── client.go               # Client, ClientConfig, NewClient
├── agent.go                # Agent, AgentConfig, NewAgent
├── session.go              # Session, Run, RunOnce
├── tool.go                 # Tool, NewTool, ToolContext, RawTool
├── provider.go             # ProviderRegistration, ModelPreset, RegisterProvider
├── model.go                # ModelSettings, model string resolution
├── result.go               # RunResult, MessageSummary, UsageSummary, RunOptions
├── options.go              # RunOption functional options
├── defaults.go             # built-in openai provider + default model presets
├── gateway.go              # internal gateway wiring (or internal/ package)
├── internal/
│   ├── resolve/            # model string → ModelDescriptor
│   ├── schema/             # struct tags → JSON Schema
│   └── wire/               # StreamFn, agent.NewAgent wiring
├── client_test.go
├── session_test.go
├── tool_test.go
├── integration_test.go     # uses gateway.FauxProvider — no API keys
└── examples/
    ├── 01-hello/
    ├── 02-tools/
    ├── 03-multi-turn/
    └── 04-anthropic-provider/
```

### Package name

- Module: `github.com/chinudotdev/pith-sdk`
- Package: `pithsdk` (avoids generic `sdk` in consumer code)

```go
import "github.com/chinudotdev/pith-sdk"
// usage: pithsdk.NewClient(...)
```

Or root package `sdk` with import alias — pick one and document in README.

---

## 9. Implementation phases

### Phase 0 — Repo bootstrap (day 1)

- [ ] Create `github.com/chinudotdev/pith-sdk` repo
- [ ] Copy this plan.md
- [ ] `go mod init github.com/chinudotdev/pith-sdk`
- [ ] Add pith primitive dependencies
- [ ] CI workflow (vet + test)
- [ ] README with positioning + link to pith

### Phase 1 — Core run path (MVP)

- [ ] `Client` with built-in OpenAI-compat + env API key
- [ ] `Agent` definition
- [ ] Internal gateway + StreamFn wiring
- [ ] `Session.Run` → blocking text result
- [ ] Default model preset for `gpt-4o-mini`
- [ ] Example `01-hello`
- [ ] Integration test with `gateway.FauxProvider`

**Exit criteria:** `Session.Run` returns text without manual EventBus subscription.

### Phase 2 — Tools

- [ ] `NewTool[T]` with struct → JSON Schema
- [ ] `ToolContext` with Local context from RunOptions
- [ ] Tool execution loop works end-to-end
- [ ] Example `02-tools`
- [ ] Tests: tool called, result in transcript

### Phase 3 — Multi-turn + streaming

- [ ] `Session` holds transcript across runs
- [ ] `Session.Reset()`, `Session.Messages()`
- [ ] `RunOptions.Stream` callback
- [ ] Example `03-multi-turn`
- [ ] `RunOnce` sugar

### Phase 4 — Custom providers

- [ ] `RegisterProvider` + `ModelPreset`
- [ ] Model string resolution (`provider/model`)
- [ ] Credential wiring per provider
- [ ] Example `04-anthropic-provider` (port from pith example 11)
- [ ] `NewClientFromGateway` escape hatch

### Phase 5 — Polish + release

- [ ] Error messages (missing API key, unknown model)
- [ ] godoc on all public types
- [ ] CHANGELOG
- [ ] Tag `v0.1.0`
- [ ] Update pith README: "Getting started → pith-sdk"

---

## 10. Testing strategy

### Unit tests (no network)

- Model string parsing (`gpt-4o-mini`, `anthropic/claude-...`)
- Schema generation from struct tags
- Client/provider registration

### Integration tests (FauxProvider)

Use `gateway.NewFauxProvider` + `gateway.FauxModel()` like pith's `agent/integration_test.go`:

- Run returns expected text
- Tool execution produces tool result in transcript
- Multi-turn session accumulates messages
- Custom provider registration routes to correct provider

### Manual smoke tests (optional, local)

- `OPENAI_API_KEY=... go run examples/01-hello/main.go`
- Anthropic example with real key

Do **not** require API keys in CI.

---

## 11. Cross-repo maintenance

### pith repo (`chinudotdev/pith`)

- Add to README top: "Build agents quickly → [pith-sdk](https://github.com/chinudotdev/pith-sdk)"
- Keep examples 01–11 as "under the hood" primitive docs
- CONTRIBUTING: feature requests for app DX → pith-sdk; primitive gaps → discuss in pith issues first

### pith-sdk repo

- Issues: SDK bugs, DX, examples, presets
- Never modify pith primitives from pith-sdk PRs — open upstream issue instead
- Changelog tracks SDK-only changes

### Local dev (both repos cloned)

```bash
# ~/dev/pith
# ~/dev/pith-sdk

# pith-sdk/go.mod (dev only — remove before release)
replace (
    github.com/chinudotdev/pith/agent    => ../pith/agent
    github.com/chinudotdev/pith/gateway  => ../pith/gateway
    github.com/chinudotdev/pith/loop     => ../pith/loop
    github.com/chinudotdev/pith/protocol => ../pith/protocol
)
```

---

## 12. Open decisions (resolve before or during Phase 1)

| Decision | Recommendation | Decide by |
|----------|----------------|-----------|
| Package name | `pithsdk` | Phase 0 |
| Session vs stateless Run as default | `Session` default; `RunOnce` for scripts | Phase 1 |
| Built-in provider presets | OpenAI only in v1; groq/ollama via RegisterProvider docs | Phase 4 |
| Struct tag convention | `json` + optional `desc:"..."` | Phase 2 |
| MessageSummary shape | `{Role, Text}` simplified — not protocol.Message | Phase 1 |
| Generic import path alias | Document `pithsdk "github.com/chinudotdev/pith-sdk"` | Phase 0 |

---

## 13. Non-goals (explicit)

- Competing with Eino, Genkit, or Google ADK on features
- Shipping Temporal/durable execution
- Shipping MCP, handoffs, guardrails in v1
- Moving primitive code into pith-sdk
- Monorepo with pith (separate repos permanently)

---

## 14. References

- Pith primitives: https://github.com/chinudotdev/pith
- Pith custom provider example: `examples/11-custom-provider/main.go`
- OpenAI Agents SDK (API north star): https://developers.openai.com/api/docs/guides/agents
- OpenAI Agents JS quickstart: https://openai.github.io/openai-agents-js/guides/quickstart/
- OpenAI agent definitions: https://developers.openai.com/api/docs/guides/agents (Agent properties table)

---

## 15. Checklist — copy to new repo and start

```
[ ] Create github.com/chinudotdev/pith-sdk
[ ] go mod init github.com/chinudotdev/pith-sdk
[ ] Copy plan.md + write README
[ ] Add CI
[ ] Phase 1: Client + Agent + Session.Run
[ ] Phase 2: NewTool[T]
[ ] Phase 3: multi-turn + streaming
[ ] Phase 4: RegisterProvider + anthropic example
[ ] Tag v0.1.0
[ ] Link from pith README
```
