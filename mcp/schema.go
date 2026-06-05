package mcp

import (
	"encoding/json"
	"fmt"
)

// marshalSchema converts a Go value to a map[string]any JSON Schema.
func marshalSchema(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal schema: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("unmarshal schema: %w", err)
	}
	return m, nil
}
