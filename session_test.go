package pithsdk_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chinudotdev/pith/gateway"
	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestNewAgentAcceptsTools(t *testing.T) {
	tool := pithsdk.NewTool("noop", "No-op tool.",
		func(ctx pithsdk.ToolContext, args struct{}) (string, error) {
			return "ok", nil
		},
	)
	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Tools: []pithsdk.Tool{tool},
	})
	if err != nil {
		t.Fatalf("NewAgent with tools: %v", err)
	}
	if agent == nil {
		t.Fatal("expected agent")
	}
}

func TestConcurrentRunRejected(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	firstActive := make(chan struct{})
	releaseFirst := make(chan struct{})
	var once sync.Once

	go func() {
		_, _ = session.Run(context.Background(), "First", pithsdk.WithStream(func(c pithsdk.TextChunk) {
			once.Do(func() {
				close(firstActive)
				<-releaseFirst
			})
		}))
	}()

	select {
	case <-firstActive:
	case <-time.After(2 * time.Second):
		t.Fatal("first Run did not start in time")
	}

	_, err = session.Run(context.Background(), "Second")
	if err == nil {
		t.Fatal("expected error for concurrent Run")
	}
	if !strings.Contains(err.Error(), "concurrent Session.Run is not supported") {
		t.Fatalf("unexpected error: %v", err)
	}

	close(releaseFirst)
}

func TestNewClientMissingAPIKeyDeferred(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	client, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient should succeed without API key: %v", err)
	}

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "You are helpful.",
		Model:        "gpt-4o-mini",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	_, err = session.Run(context.Background(), "Hi")
	if err == nil {
		t.Fatal("expected auth error when OpenAI key is missing at run time")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Fatalf("unexpected error: %v", err)
	}
}
