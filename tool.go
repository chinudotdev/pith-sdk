package pithsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/chinudotdev/pith/loop"
	"github.com/chinudotdev/pith-sdk/internal/schema"
	"github.com/chinudotdev/pith-sdk/internal/wire"
)

// ToolContext is passed to tool handlers during execution.
type ToolContext struct {
	// Run is the context for the current Session.Run call.
	Run context.Context
	// Local holds run-scoped dependencies from WithContext; not sent to the model.
	Local any
	// CallID is the provider-assigned tool call identifier.
	CallID string
}

// Tool is an opaque tool definition. Create with NewTool or RawTool.
type Tool struct {
	typed *typedTool
	raw   *loop.AgentTool
}

type typedTool struct {
	name        string
	description string
	parameters  map[string]any
	decode      func(map[string]any) (any, error)
	invoke      func(ToolContext, any) (string, error)
}

// NewTool creates a typed tool from a struct argument type T and handler function.
// T must be a struct; JSON Schema is generated from json and optional desc struct tags.
func NewTool[T any](name, description string, fn func(ToolContext, T) (string, error)) Tool {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("pithsdk.NewTool: type parameter must be a struct, got %s", t.Kind()))
	}

	parameters, err := schema.FromType(t)
	if err != nil {
		panic(fmt.Sprintf("pithsdk.NewTool: %v", err))
	}

	decode := func(params map[string]any) (any, error) {
		b, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("decode tool args: %w", err)
		}
		var args T
		if err := json.Unmarshal(b, &args); err != nil {
			return nil, fmt.Errorf("decode tool args: %w", err)
		}
		return args, nil
	}

	invoke := func(ctx ToolContext, decoded any) (string, error) {
		args, ok := decoded.(T)
		if !ok {
			return "", fmt.Errorf("internal tool arg type mismatch")
		}
		return fn(ctx, args)
	}

	return Tool{
		typed: &typedTool{
			name:        name,
			description: description,
			parameters:  parameters,
			decode:      decode,
			invoke:      invoke,
		},
	}
}

// RawTool wraps a pre-built loop.AgentTool without ToolContext injection.
func RawTool(t loop.AgentTool) Tool {
	cp := t
	return Tool{raw: &cp}
}

func toWireTools(tools []Tool, holder *wire.RunScopeHolder) []loop.AgentTool {
	var typed []wire.TypedTool
	var raw []loop.AgentTool

	for _, tool := range tools {
		if tool.raw != nil {
			raw = append(raw, *tool.raw)
			continue
		}
		if tool.typed == nil {
			continue
		}
		td := tool.typed
		typed = append(typed, wire.TypedTool{
			Name:        td.name,
			Description: td.description,
			Parameters:  td.parameters,
			Handler: func(holder *wire.RunScopeHolder, callID string, params map[string]any) (string, error) {
				decoded, err := td.decode(params)
				if err != nil {
					return "", err
				}
				var runCtx context.Context
				var local any
				if scope := holder.Current(); scope != nil {
					runCtx = scope.Ctx
					local = scope.Local
				}
				if runCtx == nil {
					runCtx = context.Background()
				}
				return td.invoke(ToolContext{
					Run:    runCtx,
					Local:  local,
					CallID: callID,
				}, decoded)
			},
		})
	}

	return wire.ToAgentTools(typed, raw, holder)
}
