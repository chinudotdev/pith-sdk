package pithsdk_test

import (
	"testing"

	pithsdk "github.com/chinudotdev/pith-sdk"
)

func TestNewAgentRejectsTools(t *testing.T) {
	_, err := pithsdk.NewAgent(pithsdk.AgentConfig{
		Tools: []pithsdk.Tool{{}},
	})
	if err == nil {
		t.Fatal("expected error when tools are provided")
	}
}

func TestNewClientMissingAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := pithsdk.NewClient(pithsdk.ClientConfig{})
	if err == nil {
		t.Fatal("expected error when API key is missing")
	}
}
