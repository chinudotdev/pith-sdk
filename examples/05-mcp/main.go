// 05-mcp demonstrates composing local and MCP tools and running an agent session.
//
// From the repo root:
//
//	OPENAI_API_KEY="sk-..." go run ./examples/05-mcp/main.go
//
// Or copy into a new module (include echo-server/):
//
//	mkdir my-agent && cd my-agent && go mod init my-agent
//	cp -r examples/05-mcp/* . && go get github.com/chinudotdev/pith-sdk@latest
//	OPENAI_API_KEY="sk-..." go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	pithsdk "github.com/chinudotdev/pith-sdk"
	"github.com/chinudotdev/pith-sdk/mcp"
)

func main() {
	ctx := context.Background()

	serverBin, cleanup := buildEchoServer()
	defer cleanup()

	mcpTools, closeMCP, err := mcp.Tools(ctx, mcp.Config{
		Command: serverBin,
	})
	if err != nil {
		log.Fatalf("mcp.Tools: %v", err)
	}
	defer closeMCP()

	localTool := pithsdk.NewTool("greet", "Greets a person by name.",
		func(ctx pithsdk.ToolContext, args struct {
			Name string `json:"name"`
		}) (string, error) {
			return fmt.Sprintf("Hello, %s!", args.Name), nil
		},
	)

	allTools := append([]pithsdk.Tool{localTool}, mcpTools...)
	fmt.Printf("Discovered %d MCP tools, %d total tools\n", len(mcpTools), len(allTools))

	client, err := pithsdk.NewClient(pithsdk.ClientConfig{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "client: %v\n", err)
		os.Exit(1)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are a helpful assistant. Use the echo tool when asked to echo text.",
		Model:        "gpt-4o-mini",
		Tools:        allTools,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "agent: %v\n", err)
		os.Exit(1)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "session: %v\n", err)
		os.Exit(1)
	}

	result, err := session.Run(ctx, "Use the echo tool to echo: hello from MCP")
	if err != nil {
		fmt.Fprintf(os.Stderr, "run: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Text)
}

// buildEchoServer compiles examples/05-mcp/echo-server into a temp binary for stdio MCP.
func buildEchoServer() (string, func()) {
	dir, err := os.MkdirTemp("", "pith-sdk-mcp-echo-*")
	if err != nil {
		log.Fatal(err)
	}
	bin := filepath.Join(dir, "echo-server")

	paths := []string{"./examples/05-mcp/echo-server", "./echo-server"}
	var out []byte
	var buildErr error
	for _, src := range paths {
		cmd := exec.Command("go", "build", "-o", bin, src)
		out, buildErr = cmd.CombinedOutput()
		if buildErr == nil {
			return bin, func() { os.RemoveAll(dir) }
		}
	}
	os.RemoveAll(dir)
	log.Fatalf("build echo server: %v\n%s", buildErr, out)
	return "", nil
}
