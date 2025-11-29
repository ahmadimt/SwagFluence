package example

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ahmadimt/SwagFluence/internal/swagger"
)

// Generator generates example JSON from schemas
type Generator struct{}

// NewGenerator creates a new Generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateExampleJSON generates example JSON from a schema
func (g *Generator) GenerateExampleJSON(schema *swagger.Schema) string {
	example := g.buildExample(schema, 0)
	bytes, _ := json.MarshalIndent(example, "", "  ")
	return string(bytes)
}

// buildExample recursively builds an example object from a schema
func (g *Generator) buildExample(schema *swagger.Schema, depth int) interface{} {
	if schema == nil || depth > 10 { // Prevent infinite recursion
		return nil
	}

	switch schema.Type {
	case "object":
		return g.buildObjectExample(schema, depth)
	case "array":
		return g.buildArrayExample(schema, depth)
	case "string":
		return g.buildStringExample(schema)
	case "integer":
		return 0
	case "number":
		return 0.0
	case "boolean":
		return false
	default:
		return nil
	}
}

func (g *Generator) buildObjectExample(schema *swagger.Schema, depth int) map[string]interface{} {
	obj := make(map[string]interface{})

	if schema.Properties != nil {
		for name, prop := range schema.Properties {
			obj[name] = g.buildPropertyExample(name, prop, depth+1)
		}
	}

	return obj
}

func (g *Generator) buildArrayExample(schema *swagger.Schema, depth int) []interface{} {
	if schema.Items == nil {
		return []interface{}{}
	}

	itemExample := g.buildExample(schema.Items, depth+1)
	return []interface{}{itemExample}
}

func (g *Generator) buildStringExample(schema *swagger.Schema) string {
	// Use example if available
	if schema.Format == "date" {
		return "2024-01-15"
	}
	if schema.Format == "date-time" {
		return "2024-01-15T10:30:00Z"
	}
	if schema.Format == "email" {
		return "user@example.com"
	}
	return "string"
}

func (g *Generator) buildPropertyExample(fieldName string, prop swagger.Property, depth int) interface{} {
	// Use explicit example if available
	if prop.Example != nil {
		return prop.Example
	}

	// Handle references
	if prop.Ref != "" {
		return fmt.Sprintf("<%s>", swagger.ExtractRefName(prop.Ref))
	}

	// Handle arrays
	if prop.Type == "array" && prop.Items != nil {
		itemExample := g.buildExample(prop.Items, depth+1)
		return []interface{}{itemExample}
	}

	// Generate default values based on type and field name
	switch prop.Type {
	case "string":
		return g.generateStringValue(fieldName, prop)
	case "integer", "number":
		return 0
	case "boolean":
		return false
	case "object":
		return map[string]interface{}{}
	default:
		return nil
	}
}

func (g *Generator) generateStringValue(fieldName string, prop swagger.Property) string {
	fieldLower := strings.ToLower(fieldName)

	if prop.Format == "date" {
		return "2024-01-15"
	}
	if prop.Format == "date-time" {
		return "2024-01-15T10:30:00Z"
	}
	if prop.Format == "email" || strings.Contains(fieldLower, "email") {
		return "user@example.com"
	}
	if strings.Contains(fieldLower, "name") {
		return fmt.Sprintf("Sample %s", fieldName)
	}
	if strings.Contains(fieldLower, "id") {
		return "123e4567-e89b-12d3-a456-426614174000"
	}

	return "string"
}