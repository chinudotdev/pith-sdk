# pith-sdk — Non-Goals

This document lists features **deliberately excluded** from the pith-sdk public API. It complements [plan.md](plan.md), which describes what we ship and how.

When the SDK is locked at **v1.0**, these remain out of scope unless a future major version explicitly revisits them.

---

## 1. Framework features (use `pith` or app code)

| Feature | Why excluded | Alternative |
|---------|--------------|-------------|
| **Real handoffs** | Multi-agent orchestration is a separate concern from “define agent → run → get text.” Requires agent graphs, control transfer, and cross-agent run scope — not composable from `Tool` alone. | Manager-as-tool pattern (sub-agent inside a tool handler), app-level routing, or drop to `pith/agent` directly. See [plan.md §8](plan.md#8-feature-evaluation-handoffs). |
| **Agents-as-tools (manager pattern)** | Same as handoffs — orchestration, not a single-agent runner. | `client.RunOnce(subAgent, ...)` inside a tool handler. |
| **Guardrails (parallel LLM validation)** | Heavy, opinionated safety layer. OpenAI-style tripwires need concurrent model calls and run-level abort semantics the minimal SDK does not own. | Validate input before `Session.Run()`; use `BeforeToolCall` / `AfterToolCall` hooks for HITL and output checks. |
| **Agent graphs / Temporal / durable execution** | Out of scope for a thin SDK. | Use workflow engines or `pith` primitives directly. |
| **Built-in tracing dashboard** | Observability platform, not library responsibility. | `SessionID`, `RunID`, `CallID` + hooks → ship to OpenTelemetry, Datadog, etc. |
| **Sandbox / hosted workspace agents** | Platform-specific (OpenAI Sandbox, isolated filesystem, etc.). | App provides workspace; tools access it via `WithContext`. |
| **OpenAI prompt templates (`prompt.id`)** | Vendor lock-in to OpenAI platform. | `WithInstructions` or dynamic prompt building in app code. |
| **Hosted tool search** | OpenAI Responses API–specific. | Register tools explicitly via `NewTool` or `mcp.Tools()`. |
| **`Agent.clone()`** | Trivial in Go without SDK support. | Copy `AgentConfig`, change fields. |

---

## 2. Platform-specific features

| Feature | Why excluded |
|---------|--------------|
| OpenAI Responses API as default transport | SDK is provider-neutral; OpenAI-compat Chat Completions is the default. |
| First-party provider packages (`pith-sdk/providers/anthropic`) | Examples and `RegisterProvider` are sufficient for v1. Preset packages are maintenance burden. |
| Built-in tools (bash, read_file, web search) | Opinionated and platform-dependent. Users define tools or use MCP. |
| Compaction helper on `Session` | Primitive exists (`agent.Compact()`); deferred to keep v1 small. Escape hatch: `NewClientFromGateway`. |

---

## 3. Blocked on primitives (not SDK decisions alone)

| Feature | Status | Policy |
|---------|--------|--------|
| **Structured output** (`output_type`) | Not in `pith` protocol/gateway today | **Add to SDK when primitive supports it.** Track upstream in `chinudotdev/pith`. Do not shim in the SDK. |
| **`tool_choice`** (force / require / disable tools) | Not in `pith` today | **Deferred.** See [plan.md §8](plan.md#8-feature-evaluation-tool_choice). Revisit only if pith adds provider-level support and user demand is clear. |

---

## 4. Evaluated and deferred (SDK could expose, but we choose not to)

These exist in `pith` primitives but are intentionally not part of the locked SDK surface:

| Feature | Why deferred for v1 |
|---------|---------------------|
| `Session.Steer()` / `FollowUp()` | Mid-run injection is power-user / HITL; adds API without helping the 15-line hello world. |
| `Session.Continue()` | Pause-resume flows; niche for minimal SDK. |
| Full `EventBus` / LLM lifecycle hooks | `WithHooks` + `WithStream` cover most production needs. Full event surface duplicates `pith/agent`. |
| Parallel tool execution policy | Performance knob; default behavior is sufficient for v1. |
| Thinking / reasoning controls | Primitive exists; SDK hardcodes off until there is clear cross-provider demand. |
| Rich `ModelSettings` (`top_p`, etc.) | Expose only what pith gateway consistently supports across providers. |

Power users: `NewClientFromGateway` + `pith/agent` directly.

---

## 5. What “locked” means

After **v1.0**:

- **No new core types** on `Client`, `Agent`, or `Session` without a major version bump.
- **Patch/minor releases**: bugfixes, docs, examples, provider presets, `pithsdk/mcp` improvements.
- **New orchestration features** (handoffs, graphs): separate module or repo — not added to root `pithsdk`.
- **Primitive-gated features** (structured output, `tool_choice`): SDK follows `pith`; no SDK-only shims.

---

## 6. Where to request changes

| Request type | Open issue in |
|--------------|---------------|
| SDK bug, DX, docs, examples | `chinudotdev/pith-sdk` |
| Primitive gap (structured output, `tool_choice`, etc.) | `chinudotdev/pith` first, then SDK exposure |
| Multi-agent orchestration | Discuss as separate package; not a pith-sdk v1 issue |
