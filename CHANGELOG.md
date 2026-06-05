# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.3.0] - 2026-06-06

### Removed

- `Client.RunOnce` — use `NewSession` + `Run`
- `RawTool` — use `NewClientFromGateway` + `pith/loop` for advanced cases
- `NewDynamicTool` — use `mcp.Tools()` for schema-driven tools
- `Hooks.ShouldStopAfterTurn` and `TurnContext` — use `WithMaxTurns` or app-side abort
- `AgentConfig.Name` and `AgentName` on hook contexts — pass labels via `WithContext`

### Changed

- `MessageSummary` is now a type alias to `internal/summary.MessageSummary` (no copy loop)
- Streaming text delta subscription inlined into `session.go`; `internal/stream` removed
- `MarshalSchema` moved to `mcp/schema.go`; `internal/dynamic_tool.go` removed
- CI builds `examples/05-mcp` only when `mcp/**` or `examples/05-mcp/**` changes

### Migration

```go
// Before (v0.2.x)
result, _ := client.RunOnce(ctx, agent, input)

// After (v0.3.0)
session, _ := client.NewSession(agent)
result, _ := session.Run(ctx, input)
```

## [0.2.1] - 2026-06-06

### Changed

- Unified hook execution: `WrapRawTool` now routes through `RunWithHooks` (single hook path)
- `RunWithHooks` shallow-copies tool params before hooks and invoke
- `Session.Run` rejects concurrent calls with an explicit error
- MCP tests build mock server once via `TestMain` (faster CI/local runs)
- `examples/05-mcp` tries monorepo and standalone echo-server paths

### Added

- `internal/wire/hooks_test.go` — block, override, and hook-error wire tests
- `TestConcurrentRunRejected` in `session_test.go`
- CI step: `go test -race ./...`
- README "Defended Requirements" section; trimmed `plan.md`

## [0.2.0] - 2026-06-05

### Added

- **Tracing IDs**: `Session.ID()`, `WithSessionID`, `WithRunID`, `RunResult.RunID`
- **ToolContext expansion**: `RunID`, `SessionID`, `ToolName` fields on `ToolContext`
- **Hooks**: `WithHooks` run option with `BeforeToolCall`, `AfterToolCall`, `ShouldStopAfterTurn`
- **MCP tools**: `pithsdk/mcp` package with `mcp.Tools()` for discovering MCP server tools
- **Dynamic tools**: `internal.DynamicTool` for schema-driven tool creation (backs MCP adapter)
- **Agent name**: `Agent.Name` propagated to hook and tracing contexts
- Example: 05-mcp demonstrating MCP tool composition
- Auto-generated UUIDs via `crypto/rand` (no external dependencies)
- Full godoc for all new public symbols

### Changed

- `Client.NewSession` now accepts variadic `SessionOption` (backward compatible)
- `internal/wire.RunScope` carries `SessionID`, `RunID`, and `Hooks`

## [0.1.0] - 2026-06-05

### Added

- `Client`, `Agent`, `Session`, and `RunOnce` for the core agent run path
- Built-in OpenAI-compatible provider via `NewClient`
- `NewTool[T]` with struct-based JSON Schema generation
- `ToolContext` with run-scoped `Local` dependencies via `WithContext`
- Multi-turn sessions with `Session.Messages()` and `Session.Reset()`
- Streaming assistant text via `WithStream`
- `WithMaxTurns` to limit tool-calling loop iterations (default 10)
- `RegisterProvider`, `ModelPreset`, and `provider/model` model resolution
- `NewClientFromGateway` escape hatch for tests and custom setups
- Examples: hello, tools, multi-turn, and custom Anthropic provider
