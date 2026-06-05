# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
