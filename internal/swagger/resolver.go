package swagger

import (
	"fmt"
	"strings"
)

// Resolver handles $ref resolution in schemas
type Resolver struct {
	spec *Spec
}

// NewResolver creates a new Resolver
func NewResolver(spec *Spec) *Resolver {
	return &Resolver{spec: spec}
}

// ResolveSchema resolves $ref references in a schema
func (r *Resolver) ResolveSchema(schema *Schema) (*Schema, error) {
	if schema == nil {
		return nil, nil
	}

	if schema.Ref != "" {
		return r.resolveRef(schema.Ref)
	}

	// Resolve nested schemas in properties
	if len(schema.Properties) > 0 {
		resolvedProperties := make(map[string]Property)
		for key, prop := range schema.Properties {
			resolvedProp, err := r.resolveProperty(prop)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve property %s: %w", key, err)
			}
			resolvedProperties[key] = resolvedProp
		}
		schema.Properties = resolvedProperties
	}

	// Resolve array items
	if schema.Items != nil {
		resolved, err := r.ResolveSchema(schema.Items)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve items: %w", err)
		}
		schema.Items = resolved
	}

	return schema, nil
}

// resolveProperty resolves a property, including its references
func (r *Resolver) resolveProperty(prop Property) (Property, error) {
	if prop.Ref != "" {
		schema, err := r.resolveRef(prop.Ref)
		if err != nil {
			return prop, err
		}
		// Convert schema back to property
		prop.Type = schema.Type
	
	}

	if prop.Items != nil && prop.Items.Ref != "" {
		resolved, err := r.resolveRef(prop.Items.Ref)
		if err != nil {
			return prop, err
		}
		prop.Items = resolved
	}

	return prop, nil
}

// resolveRef resolves a $ref string to a schema
func (r *Resolver) resolveRef(ref string) (*Schema, error) {
	// Handle #/components/schemas/... (OpenAPI 3.x)
	if strings.HasPrefix(ref, "#/components/schemas/") {
		name := strings.TrimPrefix(ref, "#/components/schemas/")
		if r.spec.Components != nil {
			if def, ok := r.spec.Components.Schemas[name]; ok {
				return &Schema{
					Type:       def.Type,
					Properties: def.Properties,
					Required:   def.Required,
				}, nil
			}
		}
		return nil, fmt.Errorf("schema not found: %s", name)
	}

	// Handle #/definitions/... (Swagger 2.0)
	if strings.HasPrefix(ref, "#/definitions/") {
		name := strings.TrimPrefix(ref, "#/definitions/")
		if def, ok := r.spec.Definitions[name]; ok {
			return &Schema{
				Type:       def.Type,
				Properties: def.Properties,
				Required:   def.Required,
			}, nil
		}
		return nil, fmt.Errorf("definition not found: %s", name)
	}

	return nil, fmt.Errorf("unsupported $ref format: %s", ref)
}

// ExtractRefName extracts the name from a $ref string
func ExtractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}
