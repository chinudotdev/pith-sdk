# pith-sdk — Implementation Plan

Blueprint for **`github.com/chinudotdev/pith-sdk`**, a minimal OpenAI Agents–style Go SDK on top of stable [Pith](https://github.com/chinudotdev/pith) primitives.

**Related:** explicit exclusions live in [NON_GOALS.md](NON_GOALS.md).

---

## 1. Goals

### What pith-sdk is

A **minimal agent SDK** for Go app developers:

- Define an agent (instructions, model, tools)
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

- [x] `Client`, `Agent`, `Session`
- [x] Built-in OpenAI-compatible provider (`NewClient`)
- [x] `RegisterProvider`, `ModelPreset`, `provider/model` resolution
- [x] `NewTool[T]` with struct-based JSON Schema
- [x] `ToolContext.Local` via `WithContext`
- [x] Multi-turn: `Session.Messages()`, `Session.Reset()`
- [x] `WithStream`, `WithMaxTurns`, `WithInstructions`
- [x] `NewClientFromGateway` escape hatch
- [x] Examples: hello, tools, multi-turn, anthropic provider
- [x] CI: vet + test (FauxProvider, no live API keys)

### Shipped — v0.2.0 (2026-06-05)

- [x] Hooks (`WithHooks`)
- [x] Tracing IDs (`SessionID`, `RunID`, `CallID`, `ToolName`)
- [x] `pithsdk/mcp` — `mcp.Tools()` → `[]pithsdk.Tool`
- [x] README + godoc sync for all public symbols
- [x] [NON_GOALS.md](NON_GOALS.md) published and linked

### Shipped — v0.2.1 (2026-06-06)

- [x] Unified hook execution path (`RunWithHooks` for all tools)
- [x] Params shallow-copy in hooks; wire-level hook tests
- [x] Concurrent `Session.Run` guard
- [x] MCP test binary cache (`TestMain`)
- [x] CI race detector (`go test -race`)

### Shipped — v0.3.0 (2026-06-06)

- [x] Removed escape-hatch APIs (`RunOnce`, `RawTool`, `NewDynamicTool`, `ShouldStopAfterTurn`, `Agent.Name`)
- [x] Internal package collapse (`internal/stream` inlined, `MarshalSchema` in `mcp/`)
- [x] CI path-filter for example builds

### Conditional — when `pith` supports it

- [ ] **Structured output** — expose typed `RunResult` only after `protocol` + gateway + providers support response schema constraints. See [NON_GOALS.md §3](NON_GOALS.md#3-blocked-on-primitives-not-sdk-decisions-alone).

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

### Design rules (API stability)

1. **`Agent` stays small** — instructions, model, tools, settings
2. **Run behavior on `Session` / `RunOption`** — stream, max turns, hooks, IDs
3. **Infrastructure on `Client`** — keys, providers, defaults
4. **String model IDs only** — never expose `ModelDescriptor` in public docs
5. **One multi-turn path** — `Session`
6. **Escape hatch** — `NewClientFromGateway`, drop to `pith/*`
7. **MCP is composition** — `append(localTools, mcpTools...)`; no MCP on `AgentConfig`

---

## 5. Version roadmap

| Version | Contents |
|---------|----------|
| **v0.1.0** ✅ | Core run path, tools, sessions, providers |
| **v0.2.0** ✅ | Hooks, tracing IDs, `pithsdk/mcp` |
| **v0.2.1** ✅ | Hook dedup, run guard, faster MCP tests, race CI |
| **v0.3.0** ✅ | API trim before v1 lock |
| **v1.x** | Bugfixes, docs, examples, optional thin primitive exposures |
| **TBD** | Structured output — when `pith` supports it |

Deferred feature evaluations (`tool_choice`, handoffs, structured output) live in [NON_GOALS.md](NON_GOALS.md) only.

---

## 6. References

- Pith primitives: https://github.com/chinudotdev/pith
- Non-goals: [NON_GOALS.md](NON_GOALS.md)
- OpenAI Agents SDK: https://openai.github.io/openai-agents-python/agents/
