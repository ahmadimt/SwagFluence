package swagger

// Spec represents a parsed Swagger/OpenAPI specification
type Spec struct {
	OpenAPI     string                `json:"openapi"`
	Swagger     string                `json:"swagger"`
	Info        Info                  `json:"info"`
	Paths       map[string]PathItem   `json:"paths"`
	Components  *Components           `json:"components,omitempty"`
	Definitions map[string]Definition `json:"definitions,omitempty"`
	Tags        []Tag                 `json:"tags,omitempty"`
}

// Info contains API metadata
type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// PathItem describes operations available on a single path
type PathItem map[string]Operation

// Operation describes a single API operation
type Operation struct {
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	OperationID string       `json:"operationId"`
	Tags        []string     `json:"tags"`
	Parameters  []Parameter  `json:"parameters"`
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	Consumes    []string     `json:"consumes,omitempty"`
	Produces    []string     `json:"produces,omitempty"`
	Responses   Responses    `json:"responses"`
}

// Parameter describes a single operation parameter
type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Type        string  `json:"type,omitempty"`
	Format      string  `json:"format,omitempty"`
	Schema      *Schema `json:"schema,omitempty"`
}

// RequestBody describes a single request body
type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

// MediaType describes media type with schema
type MediaType struct {
	Schema *Schema `json:"schema"`
}

// Responses is a map of response codes to response objects
type Responses map[string]Response

// Response describes a single response
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Schema      *Schema              `json:"schema,omitempty"` // Swagger 2.0
}

// Schema describes a data schema
type Schema struct {
	Type       string              `json:"type,omitempty"`
	Format     string              `json:"format,omitempty"`
	Ref        string              `json:"$ref,omitempty"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
	Items      *Schema             `json:"items,omitempty"`
}

// Property describes a schema property
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Format      string      `json:"format,omitempty"`
	Ref         string      `json:"$ref,omitempty"`
	Items       *Schema     `json:"items,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	MinLength   int         `json:"minLength,omitempty"`
	MaxLength   int         `json:"maxLength,omitempty"`
	Minimum     float64     `json:"minimum,omitempty"`
	Maximum     float64     `json:"maximum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	ReadOnly    bool        `json:"readOnly,omitempty"`
}

// Components holds reusable objects (OpenAPI 3.x)
type Components struct {
	Schemas map[string]Definition `json:"schemas"`
}

// Definition represents a schema definition
type Definition struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
	Ref        string              `json:"$ref,omitempty"`
}

// Tag describes an API tag
type Tag struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EndpointInfo contains information about a single endpoint
type EndpointInfo struct {
	Path      string
	Method    string
	Operation Operation
	Title     string
}
