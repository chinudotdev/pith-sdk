package pithsdk_test

import (
	"context"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestSessionRunFauxGateway(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Hello! I am a helpful assistant."},
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

	result, err := session.Run(context.Background(), "Hi there!")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Text != "Hello! I am a helpful assistant." {
		t.Fatalf("expected assistant text, got %q", result.Text)
	}
	if len(result.Messages) == 0 {
		t.Fatal("expected messages in result")
	}
}

func TestSessionRunInstructionsOverride(t *testing.T) {
	gw := gateway.NewFauxGateway(
		gateway.FauxResponse{Text: "Override works."},
	)
	client := pithsdk.NewClientFromGateway(gw)

	agent, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Instructions: "Original instructions.",
		Model:        "faux-model",
	})
	if err != nil {
		t.Fatalf("NewAgent: %v", err)
	}

	session, err := client.NewSession(agent)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	_, err = session.Run(context.Background(), "Test", pithsdk.WithInstructions("Temporary instructions."))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Ensure original prompt is restored — second run should still succeed.
	if _, err = session.Run(context.Background(), "Again"); err != nil {
		t.Fatalf("second Run: %v", err)
	}
}
