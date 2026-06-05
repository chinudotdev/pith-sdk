# 01-hello

Minimal pith-sdk example: create a client, define an agent, run one prompt.

## Run from repo root

```bash
OPENAI_API_KEY="sk-..." go run ./examples/01-hello/main.go
```

## Run as standalone module

```bash
mkdir my-agent && cd my-agent
go mod init my-agent
cp /path/to/pith-sdk/examples/01-hello/main.go .
go get github.com/chinudotdev/pith-sdk@latest
OPENAI_API_KEY="sk-..." go run main.go
```
