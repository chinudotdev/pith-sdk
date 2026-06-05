# 04 — Custom Anthropic provider

Registers a custom `ProviderPort` for the Anthropic Messages API using `RegisterProvider`.

Requires `ANTHROPIC_API_KEY` in the environment.

```bash
ANTHROPIC_API_KEY="sk-ant-..." go run ./examples/04-anthropic-provider/
```

The provider implementation lives in `anthropic.go` (ported from [pith example 11](https://github.com/chinudotdev/pith/tree/main/examples/11-custom-provider)). The SDK handles gateway wiring, credentials, and model catalog registration.
