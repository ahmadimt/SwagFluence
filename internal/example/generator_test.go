package example

import (
	"encoding/json"
	"testing"

	"github.com/ahmadimt/SwagFluence/internal/swagger"
)

func TestGenerator_GenerateExampleJSON(t *testing.T) {
	schema := &swagger.Schema{
		Type: "object",
		Properties: map[string]swagger.Property{
			"name": {
				Type: "string",
			},
			"age": {
				Type: "integer",
			},
			"active": {
				Type: "boolean",
			},
		},
	}

	gen := NewGenerator()
	result := gen.GenerateExampleJSON(schema)

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatalf("failed to parse generated JSON: %v", err)
	}

	if _, ok := obj["name"]; !ok {
		t.Error("expected 'name' field in generated JSON")
	}

	if _, ok := obj["age"]; !ok {
		t.Error("expected 'age' field in generated JSON")
	}

	if _, ok := obj["active"]; !ok {
		t.Error("expected 'active' field in generated JSON")
	}
}

func TestGenerator_BuildArrayExample(t *testing.T) {
	schema := &swagger.Schema{
		Type: "array",
		Items: &swagger.Schema{
			Type: "object",
			Properties: map[string]swagger.Property{
				"id": {Type: "integer"},
			},
		},
	}

	gen := NewGenerator()
	result := gen.buildExample(schema, 0)

	arr, ok := result.([]interface{})
	if !ok {
		t.Fatal("expected array result")
	}

	if len(arr) != 1 {
		t.Errorf("expected 1 item in array, got %d", len(arr))
	}
}
