# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
