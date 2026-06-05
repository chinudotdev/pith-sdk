package resolve_test

import (
	"strings"
	"testing"

	"github.com/chinudotdev/pith/gateway"
	"github.com/chinudotdev/pith/protocol"
	"github.com/chinudotdev/pith-sdk/internal/resolve"
)

func TestModelBareID(t *testing.T) {
	gw := gateway.NewLLMGateway()
	gw.Catalog.Register("openai", protocol.ModelDescriptor{
		ID:       "gpt-4o-mini",
		Provider: "openai",
	})

	model, err := resolve.Model(gw, "openai", "gpt-4o-mini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != "gpt-4o-mini" {
		t.Fatalf("expected gpt-4o-mini, got %q", model.ID)
	}
}

func TestModelProviderSyntax(t *testing.T) {
	gw := gateway.NewLLMGateway()
	gw.Catalog.Register("anthropic", protocol.ModelDescriptor{
		ID:       "claude-test",
		Provider: "anthropic",
	})

	model, err := resolve.Model(gw, "openai", "anthropic/claude-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.ID != "claude-test" {
		t.Fatalf("expected claude-test, got %q", model.ID)
	}
}

func TestModelUnknownProvider(t *testing.T) {
	gw := gateway.NewLLMGateway()
	_, err := resolve.Model(gw, "openai", "anthropic/claude-sonnet")
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
	if !strings.Contains(err.Error(), `unknown provider "anthropic": not registered`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModelUnknownModelInRegisteredProvider(t *testing.T) {
	gw := gateway.NewLLMGateway()
	gw.Catalog.Register("anthropic", protocol.ModelDescriptor{
		ID:       "claude-test",
		Provider: "anthropic",
	})

	_, err := resolve.Model(gw, "openai", "anthropic/claude-foo")
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if !strings.Contains(err.Error(), `unknown model "anthropic/claude-foo": provider "anthropic" registered but model not found`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModelUnknownBareID(t *testing.T) {
	gw := gateway.NewLLMGateway()
	gw.Catalog.Register("openai", protocol.ModelDescriptor{
		ID:       "gpt-4o-mini",
		Provider: "openai",
	})

	_, err := resolve.Model(gw, "openai", "missing-model")
	if err == nil {
		t.Fatal("expected error for unknown model")
	}
	if !strings.Contains(err.Error(), `unknown model "missing-model"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestModelEmptyID(t *testing.T) {
	gw := gateway.NewLLMGateway()
	_, err := resolve.Model(gw, "openai", "")
	if err == nil {
		t.Fatal("expected error for empty model ID")
	}
}
