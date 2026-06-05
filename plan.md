# pith-sdk — Implementation Plan

Blueprint for **`github.com/chinudotdev/pith-sdk`**, a minimal OpenAI Agents–style Go SDK on top of stable [Pith](https://github.com/chinudotdev/pith) primitives.

**Related:** explicit exclusions live in [NON_GOALS.md](NON_GOALS.md).

---

## 1. Goals

### What pith-sdk is

A **minimal agent SDK** for Go app developers:

- Define an agent (name, instructions, model, tools)
- Run it (`Session.Run` → get text back)
- Register custom providers (`RegisterProvider`)
- Compose MCP tools like any other tool
- Observe runs via hooks and tracing IDs

No gateway wiring, no `ModelDescriptor`, no EventBus parsing.

### What pith-sdk is not

See [NON_GOALS.md](NON_GOALS.md). In short: not a multi-agent framework, not a guardrails platform, not a tracing SaaS.

### Success criteria

| Milestone | Criteria |
|-----------|----------|
| **v0.1.0** ✅ | Zero → tool-calling agent in ~15 lines; OpenAI-compat or custom provider; no direct `protocol`/`gateway`/`loop` imports. |
| **v0.2.0** ✅ | v0.1 + hooks + tracing IDs + `pithsdk/mcp`; public API frozen per [NON_GOALS.md §5](NON_GOALS.md#5-what-locked-means). |

---

## 2. Current status

### Shipped — v0.1.0 (2026-06-05)

- [x] `Client`, `Agent`, `Session`, `RunOnce`
- [x] Built-in OpenAI-compatible provider (`NewClient`)
- [x] `RegisterProvider`, `ModelPreset`, `provider/model` resolution
- [x] `NewTool[T]` with struct-based JSON Schema
- [x] `ToolContext.Local` via `WithContext`
- [x] Multi-turn: `Session.Messages()`, `Session.Reset()`
- [x] `WithStream`, `WithMaxTurns`, `WithInstructions`
- [x] `NewClientFromGateway`, `RawTool` escape hatches
- [x] Examples: hello, tools, multi-turn, anthropic provider
- [x] CI: vet + test (FauxProvider, no live API keys)

### Shipped — v0.2.0 (2026-06-05)

- [x] Hooks (`WithHooks`)
- [x] Tracing IDs (`SessionID`, `RunID`, `CallID`, `ToolName`)
- [x] `pithsdk/mcp` — `mcp.Tools()` → `[]pithsdk.Tool`
- [x] `NewDynamicTool` (public; backs MCP adapter)
- [x] README + godoc sync for all public symbols
- [x] [NON_GOALS.md](NON_GOALS.md) published and linked

### Conditional — when `pith` supports it

- [ ] **Structured output** — expose `output_type` / typed `RunResult` only after `protocol` + gateway + providers support response schema constraints. No SDK shim. See [NON_GOALS.md §3](NON_GOALS.md#3-blocked-on-primitives-not-sdk-decisions-alone).

---

## 3. Repository split

| Repo | Module | Audience | Change policy |
|------|--------|----------|---------------|
| **`chinudotdev/pith`** | `github.com/chinudotdev/pith/{protocol,loop,gateway,agent}` | Library authors, provider implementers | Slow |
| **`chinudotdev/pith-sdk`** | `github.com/chinudotdev/pith-sdk` | App developers | Fast (pre-v1); patch/minor only (post-v1 lock) |

```
App code → pith-sdk → pith/{agent,gateway,loop,protocol}
```

Primitives never import pith-sdk.

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
│  Hooks · Tracing IDs                    │
│  pithsdk/mcp (tool adapter)             │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  pith/agent + pith/gateway              │
│  (wired internally — not exposed)       │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  pith/loop + pith/protocol              │
└─────────────────────────────────────────┘
```

### Tool call flow (local + MCP)

```
User → Session.Run → tool loop → tool call
                                    ├─ Local Tool  (NewTool handler)
                                    └─ MCP Tool    (handler → MCP server)
```

MCP needs no new primitives — it is a tool source factory.

### Design rules (API stability)

1. **`Agent` stays small** — identity only: name, instructions, model, tools, settings
2. **Run behavior on `Session` / `RunOption`** — stream, max turns, hooks, IDs
3. **Infrastructure on `Client`** — keys, providers, defaults
4. **String model IDs only** — never expose `ModelDescriptor` in public docs
5. **One multi-turn path** — `Session`
6. **Escape hatch** — `NewClientFromGateway`, `RawTool`, drop to `pith/*`
7. **MCP is composition** — `append(localTools, mcpTools...)`; no MCP on `AgentConfig`

---

## 5. Public API

### 5.1 Shipped (v0.1.0)

```go
// Client
NewClient(cfg ClientConfig) (*Client, error)
NewClientFromGateway(gw *gateway.LLMGateway) *Client
(c *Client) RegisterProvider(reg ProviderRegistration) error
(c *Client) NewSession(agent *Agent) (*Session, error)
(c *Client) RunOnce(ctx, agent, input, opts ...RunOption) (*RunResult, error)

// Agent
NewAgent(cfg AgentConfig) (*Agent, error)

// Tool
NewTool[T](name, desc string, fn func(ToolContext, T) (string, error)) Tool
RawTool(t loop.AgentTool) Tool

// Session
(s *Session) Run(ctx, input, opts ...RunOption) (*RunResult, error)
(s *Session) Messages() []MessageSummary
(s *Session) Reset()

// Run options
WithContext(local any) RunOption
WithInstructions(instructions string) RunOption
WithStream(fn func(TextChunk)) RunOption
WithMaxTurns(n int) RunOption
```

### 5.2 Shipped (v0.2.0)

#### Tracing IDs

```go
// Session
WithSessionID(id string) SessionOption   // auto UUID if omitted
(s *Session) ID() string

// Run
WithRunID(id string) RunOption           // auto UUID per Run if omitted

// RunResult
type RunResult struct {
    RunID    string
    Text     string
    Messages []MessageSummary
    Usage    *UsageSummary
}

// ToolContext (expanded)
type ToolContext struct {
    Run       context.Context
    RunID     string
    SessionID string
    Local     any
    ToolName  string
    CallID    string
}
```

ID semantics:

| ID | Scope | Purpose |
|----|-------|---------|
| `SessionID` | Conversation | Thread / conversation correlation |
| `RunID` | Single `Run()` | One user turn + agent response cycle |
| `CallID` | Tool invocation | Provider-assigned tool call ID |
| `ToolName` | Tool invocation | Human-readable tool name (local or MCP) |

#### Hooks

```go
type Hooks struct {
    BeforeToolCall      func(BeforeToolContext) (*BeforeToolResult, error)
    AfterToolCall       func(AfterToolContext) (*AfterToolResult, error)
    ShouldStopAfterTurn func(TurnContext) bool
}

WithHooks(h Hooks) RunOption
```

Hook context types carry `RunID`, `SessionID`, `AgentName`, `ToolName`, `CallID`, and tool args.

Use cases: HITL approval, logging/metrics, output redaction, `stop_on_first_tool` via `ShouldStopAfterTurn`.

#### MCP subpackage

```go
// github.com/chinudotdev/pith-sdk/mcp
type Config struct {
    Command string
    Args    []string
    Env     []string
    // URL (remote transport) — deferred; stdio via Command is supported today.
}

func Tools(ctx context.Context, cfg Config) (tools []pithsdk.Tool, close func() error, err error)
```

`NewDynamicTool` generates schema-driven tools from JSON Schema maps (used by the MCP adapter).

### 5.3 Conditional (primitive-gated)

```go
// Future — only when pith protocol/gateway supports response schema
type AgentConfig struct {
    // ...
    OutputType OutputSchema // TBD — shape follows pith primitive API
}
```

Do not design or ship until `chinudotdev/pith` adds structured output support.

---

## 6. Implementation phases

### Phase 0–4 — Complete ✅

Bootstrap, core run path, tools, multi-turn/streaming, custom providers. See [CHANGELOG.md](CHANGELOG.md).

### Phase 5 — v0.2.0 (complete)

| Step | Deliverable |
|------|-------------|
| 5a | `WithSessionID`, `Session.ID()`, `WithRunID`, `RunResult.RunID`, expand `ToolContext` |
| 5b | `WithHooks` + hook context types; wire to `loop.LoopHooks` |
| 5c | `NewDynamicTool`; `pithsdk/mcp` package + example |
| 5d | Wire `Agent.Name` into hook/tracing context |
| 5e | README, godoc, [NON_GOALS.md](NON_GOALS.md); tag **v0.2.0** |

**Exit criteria:** Integration tests for hooks, IDs, and MCP (stdio mock); no new core types beyond plan; API matches §5.2.

### Phase 6 — Post-lock (patch/minor only)

- Provider preset examples (Groq, Ollama)
- Optional thin exposures if demand is clear: `WithImages`, `Session.Steer` / `FollowUp` (primitives exist — see [NON_GOALS.md §4](NON_GOALS.md#4-evaluated-and-deferred-sdk-could-expose-but-we-choose-not-to))
- Structured output SDK exposure — **only after pith ships it**

---

## 7. OpenAI Agents SDK comparison

| OpenAI | pith-sdk v1 | Policy |
|--------|-------------|--------|
| Agent, Runner, Tools | Agent, Session, NewTool | ✅ Keep |
| Sessions, streaming | Session, WithStream | ✅ Keep |
| Context DI | WithContext | ✅ Keep |
| MCP | mcp.Tools() composed | ✅ Keep |
| Tool hooks | WithHooks | ✅ Keep |
| Tracing | IDs + hooks (DIY) | ✅ Keep |
| Handoffs | — | ⏭ [§8 evaluation](#8-feature-evaluation-handoffs) |
| Guardrails | — | ⏭ [NON_GOALS.md](NON_GOALS.md) |
| Structured output | — | ⏳ When primitive supports |
| tool_choice | — | ⏭ [§8 evaluation](#8-feature-evaluation-tool_choice) |

---

## 8. Feature evaluation

### tool_choice

**What it is:** Force the model to call a specific tool, require any tool, or forbid tools (`auto` / `required` / `none` / named tool).

**Current state:** Not in `pith` protocol, gateway, or providers.

#### Is it needed for pith-sdk v1?

**No — defer.**

| For | Against |
|-----|---------|
| Niche flows: guaranteed tool invocation, pipeline steps that must call an API | Not in primitives — SDK cannot ship without upstream work |
| OpenAI SDK parity | Most apps work with good tool descriptions + instructions |
| | `BeforeToolCall` hook can block wrong tools (HITL-style) |
| | Forcing tools via prompt is unreliable but often sufficient |
| | Adds `ModelSettings` complexity across providers with inconsistent support |

**Recommendation:** Do not add to SDK roadmap until `pith` adds provider-level `tool_choice`. Document workaround: strong instructions + tool naming. Revisit if extraction/ETL use cases demand it.

**Workarounds today:**

```go
// Instruction-based (good enough for most cases)
Instructions: "You must call get_weather before answering."

// Hook-based gate
BeforeToolCall: func(h BeforeToolContext) (*BeforeToolResult, error) {
    if h.ToolName != "get_weather" {
        return &BeforeToolResult{Block: true, Reason: "only get_weather allowed"}, nil
    }
    return nil, nil
}
```

---

### Real handoffs

**What it is:** Agent A delegates the conversation to Agent B; B inherits history, system prompt, tools, and run scope; framework tracks active agent and emits `on_handoff`.

**Current state:** No handoff primitive in `pith`. Composable only via hacks (tool that calls another session).

#### Is it needed for pith-sdk v1?

**No — exclude from SDK; optional future orchestrator package.**

| For | Against |
|-----|---------|
| OpenAI SDK parity | Conflicts with “smallest API surface” positioning |
| Triage → specialist routing | Manager-as-tool pattern covers 80% without framework support |
| | Requires new types (`Orchestrator`, handoff graph, run scope across agents) |
| | Hooks scope becomes ambiguous (which agent’s hooks fire?) |
| | Most pith-sdk users want single-agent apps in ~15 lines |

**Recommendation:** Keep handoffs out of root `pithsdk`. If demand grows, ship `pith-sdk/orchestrator` or document patterns in examples — never inflate `AgentConfig`.

**Workarounds today:**

```go
// Manager-as-tool: specialist agent inside a tool handler
billingTool := pithsdk.NewTool("billing_expert", "Handles billing questions.",
    func(tc pithsdk.ToolContext, args struct{ Query string `json:"query"` }) (string, error) {
        result, err := client.RunOnce(tc.Run, billingAgent, args.Query)
        if err != nil {
            return "", err
        }
        return result.Text, nil
    },
)
```

This is not a true handoff (separate transcript, no `on_handoff`), but sufficient for many routing tasks.

---

### Structured output

**What it is:** Agent returns a typed struct (e.g. `CalendarEvent`) instead of plain text.

**Current state:** Not in `pith`.

**Policy:** **Add to SDK when primitive supports it.** Do not shim JSON parsing in the SDK. Track as upstream work in `chinudotdev/pith`; then expose via `AgentConfig` or `RunOption` matching the primitive API shape.

---

## 9. Repository structure

```
pith-sdk/
├── plan.md
├── NON_GOALS.md
├── README.md
├── CHANGELOG.md
├── client.go
├── agent.go
├── session.go
├── tool.go
├── hooks.go              # Phase 5b
├── tracing.go            # Phase 5a (or merged into session/result)
├── provider.go
├── model.go
├── result.go
├── options.go
├── defaults.go
├── credentials.go
├── mcp/
│   └── mcp.go            # Phase 5c
├── internal/
│   ├── resolve/
│   ├── schema/
│   ├── dynamic_tool.go   # Phase 5c (internal)
│   ├── stream/
│   ├── summary/
│   └── wire/
├── examples/
│   ├── 01-hello/
│   ├── 02-tools/
│   ├── 03-multi-turn/
│   ├── 04-anthropic-provider/
│   └── 05-mcp/           # Phase 5c
└── .github/workflows/ci.yml
```

---

## 10. Testing strategy

### Unit tests (no network)

- Model string resolution
- JSON Schema from struct tags
- Hook wiring, ID propagation
- MCP schema → dynamic tool mapping (mock server)

### Integration tests (FauxProvider / mock MCP)

- Run returns text; tools execute; multi-turn accumulates
- Hooks: block tool, override result, stop after turn
- `RunID` / `SessionID` present on result and `ToolContext`
- MCP tools discovered and invoked end-to-end (stdio mock)

Do **not** require live API keys or external MCP servers in CI.

---

## 11. Version roadmap

| Version | Contents |
|---------|----------|
| **v0.1.0** ✅ | Core run path, tools, sessions, providers |
| **v0.2.0** ✅ | Hooks, tracing IDs, `pithsdk/mcp`, API lock |
| **v1.x** | Bugfixes, docs, examples, optional thin primitive exposures |
| **v2.0** | Only if intentional breaking API change |
| **TBD** | Structured output — when `pith` supports it |
| **Unlikely** | `tool_choice`, handoffs in root package — see §8 |

---

## 12. References

- Pith primitives: https://github.com/chinudotdev/pith
- Non-goals: [NON_GOALS.md](NON_GOALS.md)
- OpenAI Agents SDK: https://openai.github.io/openai-agents-python/agents/
- Pith custom provider: `pith/examples/11-custom-provider/`
