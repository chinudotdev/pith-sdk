package schema

import (
	"reflect"
	"testing"
)

func TestFromType_BasicStruct(t *testing.T) {
	type args struct {
		City    string `json:"city" desc:"City name"`
		Country string `json:"country,omitempty"`
	}

	schema, err := FromType(reflect.TypeOf(args{}))
	if err != nil {
		t.Fatalf("FromType: %v", err)
	}

	if schema["type"] != "object" {
		t.Fatalf("expected type object, got %v", schema["type"])
	}
	if schema["additionalProperties"] != false {
		t.Fatal("expected additionalProperties false")
	}

	props, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("expected properties map")
	}

	city, ok := props["city"].(map[string]any)
	if !ok {
		t.Fatal("expected city property schema")
	}
	if city["type"] != "string" {
		t.Fatalf("expected city type string, got %v", city["type"])
	}
	if city["description"] != "City name" {
		t.Fatalf("expected city description, got %v", city["description"])
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("expected required slice")
	}
	if len(required) != 1 || required[0] != "city" {
		t.Fatalf("expected required [city], got %v", required)
	}
}

func TestFromType_ScalarTypes(t *testing.T) {
	type args struct {
		Flag   bool    `json:"flag"`
		Count  int     `json:"count"`
		Ratio  float64 `json:"ratio"`
		Hidden string  `json:"-"`
	}

	schema, err := FromType(reflect.TypeOf(args{}))
	if err != nil {
		t.Fatalf("FromType: %v", err)
	}

	props := schema["properties"].(map[string]any)
	if props["flag"].(map[string]any)["type"] != "boolean" {
		t.Fatal("expected boolean type for flag")
	}
	if props["count"].(map[string]any)["type"] != "integer" {
		t.Fatal("expected integer type for count")
	}
	if props["ratio"].(map[string]any)["type"] != "number" {
		t.Fatal("expected number type for ratio")
	}
	if _, ok := props["hidden"]; ok {
		t.Fatal("hidden field should be omitted")
	}
}

func TestFromType_NonStruct(t *testing.T) {
	_, err := FromType(reflect.TypeOf(""))
	if err == nil {
		t.Fatal("expected error for non-struct type")
	}
}

func TestFromType_UnsupportedField(t *testing.T) {
	type args struct {
		Nested struct{ X int } `json:"nested"`
	}
	_, err := FromType(reflect.TypeOf(args{}))
	if err == nil {
		t.Fatal("expected error for nested struct field")
	}
}
