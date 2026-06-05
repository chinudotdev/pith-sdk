package pithsdk

import (
	"reflect"
	"testing"

	"github.com/chinudotdev/pith-sdk/internal/schema"
)

func TestNewTool_Schema(t *testing.T) {
	tool := NewTool("get_weather", "Return weather for a city.",
		func(ctx ToolContext, args struct {
			City string `json:"city" desc:"City name"`
		}) (string, error) {
			return "ok", nil
		},
	)

	if tool.typed == nil {
		t.Fatal("expected typed tool")
	}
	if tool.typed.name != "get_weather" {
		t.Fatalf("unexpected name %q", tool.typed.name)
	}

	schema, err := schema.FromType(reflect.TypeOf(struct {
		City string `json:"city" desc:"City name"`
	}{}))
	if err != nil {
		t.Fatalf("FromType: %v", err)
	}

	props := schema["properties"].(map[string]any)
	if tool.typed.parameters["type"] != "object" {
		t.Fatal("expected object schema")
	}
	toolProps := tool.typed.parameters["properties"].(map[string]any)
	if toolProps["city"] == nil {
		t.Fatal("expected city property in tool schema")
	}
	_ = props
}

func TestNewTool_PanicsOnNonStruct(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-struct type parameter")
		}
	}()
	_ = NewTool("bad", "bad", func(ctx ToolContext, args string) (string, error) {
		return "", nil
	})
}
