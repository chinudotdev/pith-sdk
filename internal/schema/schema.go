package schema

import (
	"fmt"
	"reflect"
	"strings"
)

// FromType generates a JSON Schema object for a struct type T.
func FromType(t reflect.Type) (map[string]any, error) {
	if t == nil {
		return nil, fmt.Errorf("schema: nil type")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema: expected struct, got %s", t.Kind())
	}

	properties := make(map[string]any)
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		name, omitEmpty := parseJSONTag(jsonTag)
		if name == "" {
			name = field.Name
		}

		propSchema, err := fieldSchema(field.Type)
		if err != nil {
			return nil, fmt.Errorf("schema: field %q: %w", name, err)
		}

		if desc := field.Tag.Get("desc"); desc != "" {
			propSchema["description"] = desc
		}

		properties[name] = propSchema
		if !omitEmpty {
			required = append(required, name)
		}
	}

	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema, nil
}

func parseJSONTag(tag string) (name string, omitEmpty bool) {
	if tag == "" {
		return "", false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	for _, part := range parts[1:] {
		if part == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty
}

func fieldSchema(t reflect.Type) (map[string]any, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}, nil
	case reflect.Bool:
		return map[string]any{"type": "boolean"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}, nil
	default:
		return nil, fmt.Errorf("unsupported type %s", t.Kind())
	}
}
