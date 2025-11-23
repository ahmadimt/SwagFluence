package swagger

import (
	"testing"
)

func TestResolver_ResolveSchema(t *testing.T) {
	spec := &Spec{
		Definitions: map[string]Definition{
			"User": {
				Type: "object",
				Properties: map[string]Property{
					"id": {
						Type: "integer",
					},
					"name": {
						Type: "string",
					},
				},
			},
		},
	}

	resolver := NewResolver(spec)

	schema := &Schema{
		Ref: "#/definitions/User",
	}

	resolved, err := resolver.ResolveSchema(schema)
	if err != nil {
		t.Fatalf("ResolveSchema() error = %v", err)
	}

	if resolved.Type != "object" {
		t.Errorf("expected type 'object', got '%s'", resolved.Type)
	}

	if len(resolved.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(resolved.Properties))
	}
}

func TestResolver_ResolveNestedSchema(t *testing.T) {
	spec := &Spec{
		Components: &Components{
			Schemas: map[string]Definition{
				"Address": {
					Type: "object",
					Properties: map[string]Property{
						"street": {Type: "string"},
						"city":   {Type: "string"},
					},
				},
				"User": {
					Type: "object",
					Properties: map[string]Property{
						"name": {Type: "string"},
						"address": {
							Ref: "#/components/schemas/Address",
						},
					},
				},
			},
		},
	}

	resolver := NewResolver(spec)

	schema := &Schema{
		Ref: "#/components/schemas/User",
	}

	resolved, err := resolver.ResolveSchema(schema)
	if err != nil {
		t.Fatalf("ResolveSchema() error = %v", err)
	}

	if len(resolved.Properties) != 2 {
		t.Errorf("expected 2 properties, got %d", len(resolved.Properties))
	}
}

func TestExtractRefName(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{
			ref:  "#/definitions/User",
			want: "User",
		},
		{
			ref:  "#/components/schemas/Pet",
			want: "Pet",
		},
		{
			ref:  "User",
			want: "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := ExtractRefName(tt.ref)
			if got != tt.want {
				t.Errorf("ExtractRefName() = %v, want %v", got, tt.want)
			}
		})
	}
}