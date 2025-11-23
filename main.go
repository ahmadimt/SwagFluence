package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Swagger/OpenAPI structures
type Swagger struct {
	Paths       map[string]Path       `json:"paths"`
	Info        Info                  `json:"info"`
	Definitions map[string]Definition `json:"definitions"` // Swagger 2.0
	Components  *Components           `json:"components"`  // OpenAPI 3.0
}

type Components struct {
	Schemas map[string]Definition `json:"schemas"`
}

type Info struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

type Path map[string]Operation

type Operation struct {
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	OperationID string       `json:"operationId"`
	Tags        []string     `json:"tags"`
	Parameters  []Parameter  `json:"parameters"`
	RequestBody *RequestBody `json:"requestBody"` // OpenAPI 3.0
	Consumes    []string     `json:"consumes"`
	Produces    []string     `json:"produces"`
}

type Parameter struct {
	Name        string  `json:"name"`
	In          string  `json:"in"`
	Description string  `json:"description"`
	Required    bool    `json:"required"`
	Type        string  `json:"type"`
	Format      string  `json:"format"`
	Schema      *Schema `json:"schema"`
}

type Schema struct {
	Type       string              `json:"type"`
	Ref        string              `json:"$ref"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
	Items      *Schema             `json:"items"`
}

type Property struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Format      string  `json:"format"`
	Ref         string  `json:"$ref"`
	Items       *Schema `json:"items"`
	Example     any     `json:"example"` // Changed to any to handle any type
	MinLength   int     `json:"minLength"`
	MaxLength   int     `json:"maxLength"`
	Minimum     float64 `json:"minimum"`
	Maximum     float64 `json:"maximum"`
	Pattern     string  `json:"pattern"`
	ReadOnly    bool    `json:"readOnly"`
}

type RequestBody struct {
	Description string               `json:"description"`
	Required    bool                 `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Definition struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
	Ref        string              `json:"$ref"`
}

// Confluence structures
type ConfluencePage struct {
	Type      string         `json:"type"`
	Title     string         `json:"title"`
	Space     SpaceKey       `json:"space"`
	Body      Body           `json:"body"`
	Version   *Version       `json:"version,omitempty"`
	Ancestors []PageAncestor `json:"ancestors,omitempty"`
}

type PageAncestor struct {
	ID string `json:"id"`
}

type SpaceKey struct {
	Key string `json:"key"`
}

type Body struct {
	Storage Storage `json:"storage"`
}

type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

type Version struct {
	Number int `json:"number"`
}

type ConfluenceConfig struct {
	BaseURL      string
	Username     string
	APIToken     string
	SpaceKey     string
	ParentPageID string
}

type EndpointInfo struct {
	Path      string
	Method    string
	Operation Operation
	Title     string
}

var swaggerSpec Swagger // Global variable to hold the full swagger spec

func main() {
	// Check for CLI argument
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <swagger-url>")
		fmt.Println("\nExample:")
		fmt.Println("  go run main.go https://petstore.swagger.io/v2/swagger.json")
		fmt.Println("\nEnvironment variables (optional for Confluence integration):")
		fmt.Println("  CONFLUENCE_BASE_URL")
		fmt.Println("  CONFLUENCE_USERNAME")
		fmt.Println("  CONFLUENCE_API_TOKEN")
		fmt.Println("  CONFLUENCE_SPACE_KEY")
		fmt.Println("  CONFLUENCE_PARENT_PAGE_ID (optional)")
		os.Exit(1)
	}

	swaggerURL := os.Args[1]

	config := ConfluenceConfig{
		BaseURL:      os.Getenv("CONFLUENCE_BASE_URL"),
		Username:     os.Getenv("CONFLUENCE_USERNAME"),
		APIToken:     os.Getenv("CONFLUENCE_API_TOKEN"),
		SpaceKey:     os.Getenv("CONFLUENCE_SPACE_KEY"),
		ParentPageID: os.Getenv("CONFLUENCE_PARENT_PAGE_ID"),
	}

	fmt.Printf("Fetching Swagger specification from: %s\n", swaggerURL)
	endpoints, apiTitle, err := parseSwaggerSpec(swaggerURL)
	if err != nil {
		fmt.Printf("Error parsing Swagger: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d endpoints in %s\n\n", len(endpoints), apiTitle)

	if config.BaseURL != "" && config.Username != "" && config.APIToken != "" {
		parentPageID := config.ParentPageID
		if parentPageID == "" {
			parentTitle := fmt.Sprintf("%s - API Documentation", apiTitle)
			fmt.Printf("Creating/updating parent page: %s\n", parentTitle)
			parentPageID, err = createOrUpdateParentPage(config, parentTitle, apiTitle)
			if err != nil {
				fmt.Printf("Error creating parent page: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Parent page ID: %s\n\n", parentPageID)
		}

		successCount := 0
		for i, endpoint := range endpoints {
			fmt.Printf("[%d/%d] Processing: %s %s\n", i+1, len(endpoints),
				strings.ToUpper(endpoint.Method), endpoint.Path)

			confluenceMarkup := generateOperationTable(endpoint.Path, endpoint.Method, endpoint.Operation)

			pageID, err := createOrUpdateEndpointPage(config, endpoint.Title, confluenceMarkup, parentPageID)
			if err != nil {
				fmt.Printf("  ✗ Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("  ✓ Success - Page ID: %s\n", pageID)
			fmt.Printf("     URL: %s/pages/viewpage.action?pageId=%s\n\n", config.BaseURL, pageID)
			successCount++
		}

		fmt.Printf("\n=================================\n")
		fmt.Printf("Summary: %d/%d pages created/updated successfully\n", successCount, len(endpoints))
		fmt.Printf("Parent page: %s/pages/viewpage.action?pageId=%s\n", config.BaseURL, parentPageID)

	} else {
		fmt.Println("Confluence credentials not provided. Showing generated content only.")
		fmt.Println("\nTo create pages directly, set these environment variables:")
		fmt.Println("  CONFLUENCE_BASE_URL")
		fmt.Println("  CONFLUENCE_USERNAME")
		fmt.Println("  CONFLUENCE_API_TOKEN")
		fmt.Println("  CONFLUENCE_SPACE_KEY")
		fmt.Println("  CONFLUENCE_PARENT_PAGE_ID (optional)")
		fmt.Println("\n=================================\n")

		for i, endpoint := range endpoints {
			fmt.Printf("\n[%d] Page Title: %s\n", i+1, endpoint.Title)
			fmt.Printf("Endpoint: %s %s\n", strings.ToUpper(endpoint.Method), endpoint.Path)
			fmt.Println("---")
			confluenceMarkup := generateOperationTable(endpoint.Path, endpoint.Method, endpoint.Operation)
			fmt.Println(confluenceMarkup)
			fmt.Println("\n=================================\n")
		}
	}
}

func parseSwaggerSpec(swaggerURL string) ([]EndpointInfo, string, error) {
	resp, err := http.Get(swaggerURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch swagger: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &swaggerSpec); err != nil {
		return nil, "", fmt.Errorf("failed to parse swagger: %w", err)
	}

	var endpoints []EndpointInfo

	for path, pathItem := range swaggerSpec.Paths {
		for method, operation := range pathItem {
			if isHTTPMethod(method) {
				title := generatePageTitle(path, method, operation)
				endpoints = append(endpoints, EndpointInfo{
					Path:      path,
					Method:    method,
					Operation: operation,
					Title:     title,
				})
			}
		}
	}

	return endpoints, swaggerSpec.Info.Title, nil
}

func generatePageTitle(path, method string, operation Operation) string {
	if operation.Summary != "" {
		return operation.Summary
	}

	if operation.OperationID != "" {
		return cleanOperationID(operation.OperationID)
	}

	return generateTitleFromPath(path, method)
}

func cleanOperationID(operationID string) string {
	result := strings.ReplaceAll(operationID, "_", " ")

	var builder strings.Builder
	for i, r := range result {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteRune(' ')
		}
		builder.WriteRune(r)
	}
	titleCaser := cases.Title(language.Und)

	return titleCaser.String(strings.ToLower(builder.String()))
}

func generateTitleFromPath(path, method string) string {
	cleanPath := strings.TrimPrefix(path, "/")
	cleanPath = strings.ReplaceAll(cleanPath, "{", "")
	cleanPath = strings.ReplaceAll(cleanPath, "}", "")

	parts := strings.Split(cleanPath, "/")
	titleCaser := cases.Title(language.Und)
	var titleParts []string

	for _, part := range parts {
		if part != "" {
			titleParts = append(titleParts, titleCaser.String(part))
		}
	}

	methodVerb := strings.ToUpper(method)

	if len(titleParts) == 0 {
		return fmt.Sprintf("%s Root", methodVerb)
	}

	return fmt.Sprintf("%s %s", methodVerb, strings.Join(titleParts, " "))
}

func isHTTPMethod(method string) bool {
	httpMethods := []string{"get", "post", "put", "delete", "patch", "options", "head"}
	for _, m := range httpMethods {
		if method == m {
			return true
		}
	}
	return false
}

func generateOperationTable(path, method string, operation Operation) string {
	var table strings.Builder

	// Add layout section for full width in Confluence Cloud
	table.WriteString("<ac:layout>\n")
	table.WriteString("<ac:layout-section ac:type=\"single\">\n")
	table.WriteString("<ac:layout-cell>\n")

	// Header with method badge
	table.WriteString("<h2>")
	table.WriteString("<ac:structured-macro ac:name=\"status\">")
	table.WriteString("<ac:parameter ac:name=\"colour\">Blue</ac:parameter>")
	table.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"title\">%s</ac:parameter>", strings.ToUpper(method)))
	table.WriteString("</ac:structured-macro>")
	table.WriteString(fmt.Sprintf(" %s</h2>\n", path))

	// Description
	if operation.Description != "" {
		table.WriteString(fmt.Sprintf("<p>%s</p>\n", operation.Description))
	}

	// Operation ID
	if operation.OperationID != "" {
		table.WriteString(fmt.Sprintf("<p><strong>Operation ID:</strong> <code>%s</code></p>\n", operation.OperationID))
	}

	// Tags
	if len(operation.Tags) > 0 {
		table.WriteString("<p><strong>Tags:</strong> ")
		for i, tag := range operation.Tags {
			if i > 0 {
				table.WriteString(", ")
			}
			table.WriteString("<ac:structured-macro ac:name=\"status\">")
			table.WriteString("<ac:parameter ac:name=\"colour\">Grey</ac:parameter>")
			table.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"title\">%s</ac:parameter>", tag))
			table.WriteString("</ac:structured-macro>")
		}
		table.WriteString("</p>\n")
	}

	// Content types
	if len(operation.Consumes) > 0 {
		table.WriteString("<p><strong>Consumes:</strong> ")
		table.WriteString(fmt.Sprintf("<code>%s</code>", strings.Join(operation.Consumes, ", ")))
		table.WriteString("</p>\n")
	}

	if len(operation.Produces) > 0 {
		table.WriteString("<p><strong>Produces:</strong> ")
		table.WriteString(fmt.Sprintf("<code>%s</code>", strings.Join(operation.Produces, ", ")))
		table.WriteString("</p>\n")
	}

	// REQUEST BODY SECTION (NEW)
	table.WriteString(generateRequestBodySection(operation))

	// PARAMETERS SECTION
	table.WriteString("<h3>Parameters</h3>\n")
	table.WriteString("<table>\n")
	table.WriteString("<tr><th>Parameter</th><th>Description</th></tr>\n")

	if len(operation.Parameters) > 0 {
		for _, param := range operation.Parameters {
			// Skip body parameters as they're handled in the request body section
			if param.In == "body" {
				continue
			}

			table.WriteString("<tr>\n")
			table.WriteString(fmt.Sprintf("<td><code>%s</code></td>\n", param.Name))
			table.WriteString("<td>")

			if param.Required {
				table.WriteString("<ac:structured-macro ac:name=\"status\">")
				table.WriteString("<ac:parameter ac:name=\"colour\">Red</ac:parameter>")
				table.WriteString("<ac:parameter ac:name=\"title\">MANDATORY</ac:parameter>")
				table.WriteString("</ac:structured-macro>")
			} else {
				table.WriteString("<ac:structured-macro ac:name=\"status\">")
				table.WriteString("<ac:parameter ac:name=\"colour\">Green</ac:parameter>")
				table.WriteString("<ac:parameter ac:name=\"title\">OPTIONAL</ac:parameter>")
				table.WriteString("</ac:structured-macro>")
			}

			table.WriteString("<br/><br/>")

			if param.Description != "" {
				table.WriteString(param.Description)
			} else {
				table.WriteString("No description provided")
			}

			paramType := getParameterType(param)
			if paramType != "" {
				table.WriteString(fmt.Sprintf("<br/><br/><strong>Type:</strong> <code>%s</code>", paramType))
			}

			if param.In != "" {
				table.WriteString(fmt.Sprintf("<br/><br/><strong>Location:</strong> %s", param.In))
			}

			table.WriteString("</td>\n")
			table.WriteString("</tr>\n")
		}
	} else {
		table.WriteString("<tr>\n")
		table.WriteString("<td colspan=\"2\"><em>This endpoint requires no parameters</em></td>\n")
		table.WriteString("</tr>\n")
	}

	table.WriteString("</table>\n")

	// Close layout section
	table.WriteString("</ac:layout-cell>\n")
	table.WriteString("</ac:layout-section>\n")
	table.WriteString("</ac:layout>\n")

	return table.String()
}

func generateRequestBodySection(operation Operation) string {
	var body strings.Builder

	// Check for body parameter (Swagger 2.0)
	var bodyParam *Parameter
	for _, param := range operation.Parameters {
		if param.In == "body" {
			bodyParam = &param
			break
		}
	}

	// Check for requestBody (OpenAPI 3.0)
	hasRequestBody := operation.RequestBody != nil || bodyParam != nil

	if !hasRequestBody {
		return ""
	}

	body.WriteString("<h3>Request Body</h3>\n")

	var schemaToUse *Schema

	// Handle OpenAPI 3.0 requestBody
	if operation.RequestBody != nil {
		if operation.RequestBody.Description != "" {
			body.WriteString(fmt.Sprintf("<p>%s</p>\n", operation.RequestBody.Description))
		}

		if operation.RequestBody.Required {
			body.WriteString("<p>")
			body.WriteString("<ac:structured-macro ac:name=\"status\">")
			body.WriteString("<ac:parameter ac:name=\"colour\">Red</ac:parameter>")
			body.WriteString("<ac:parameter ac:name=\"title\">REQUIRED</ac:parameter>")
			body.WriteString("</ac:structured-macro>")
			body.WriteString("</p>\n")
		}

		for contentType, mediaType := range operation.RequestBody.Content {
			body.WriteString(fmt.Sprintf("<p><strong>Content-Type:</strong> <code>%s</code></p>\n", contentType))
			schemaToUse = &mediaType.Schema
			body.WriteString(generateSchemaTable(&mediaType.Schema))
		}
	}

	// Handle Swagger 2.0 body parameter
	if bodyParam != nil {
		if bodyParam.Description != "" {
			body.WriteString(fmt.Sprintf("<p>%s</p>\n", bodyParam.Description))
		}

		if bodyParam.Required {
			body.WriteString("<p>")
			body.WriteString("<ac:structured-macro ac:name=\"status\">")
			body.WriteString("<ac:parameter ac:name=\"colour\">Red</ac:parameter>")
			body.WriteString("<ac:parameter ac:name=\"title\">REQUIRED</ac:parameter>")
			body.WriteString("</ac:structured-macro>")
			body.WriteString("</p>\n")
		}

		if bodyParam.Schema != nil {
			schemaToUse = bodyParam.Schema
			body.WriteString(generateSchemaTable(bodyParam.Schema))
		}
	}

	// Add Example JSON section
	if schemaToUse != nil {
		body.WriteString(generateExampleJSON(schemaToUse))
	}

	return body.String()
}

func generateSchemaTable(schema *Schema) string {
	if schema == nil {
		return ""
	}

	var table strings.Builder

	// Resolve reference if present
	if schema.Ref != "" {
		resolvedSchema := resolveSchemaRef(schema.Ref)
		if resolvedSchema != nil {
			schema = resolvedSchema
		} else {
			// If we can't resolve, show the reference name
			table.WriteString(fmt.Sprintf("<p><strong>Schema:</strong> %s</p>\n", extractRefName(schema.Ref)))
			return table.String()
		}
	}

	// Handle array type
	if schema.Type == "array" && schema.Items != nil {
		table.WriteString("<p><strong>Type:</strong> Array</p>\n")
		if schema.Items.Ref != "" {
			table.WriteString(fmt.Sprintf("<p><strong>Items:</strong> %s</p>\n", extractRefName(schema.Items.Ref)))
			resolvedSchema := resolveSchemaRef(schema.Items.Ref)
			if resolvedSchema != nil {
				schema = resolvedSchema
			}
		} else if schema.Items.Type != "" {
			table.WriteString(fmt.Sprintf("<p><strong>Items Type:</strong> %s</p>\n", schema.Items.Type))
		}
	}

	// Generate properties table
	if len(schema.Properties) > 0 {
		table.WriteString("<table>\n")
		table.WriteString("<tr><th>Field</th><th>Type</th><th>Description</th><th>Constraints</th><th>Example</th></tr>\n")

		// Sort properties by name for consistent output
		var propNames []string
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}

		for _, fieldName := range propNames {
			prop := schema.Properties[fieldName]
			table.WriteString("<tr>\n")

			// Field name with required indicator
			table.WriteString("<td><code>")
			table.WriteString(fieldName)
			if isFieldRequired(fieldName, schema.Required) {
				table.WriteString(" *")
			}
			table.WriteString("</code></td>\n")

			// Type
			table.WriteString("<td><code>")
			if prop.Ref != "" {
				table.WriteString(extractRefName(prop.Ref))
			} else if prop.Type != "" {
				table.WriteString(prop.Type)
				if prop.Format != "" {
					table.WriteString(fmt.Sprintf(" (%s)", prop.Format))
				}
				if prop.Type == "array" && prop.Items != nil {
					if prop.Items.Ref != "" {
						table.WriteString(fmt.Sprintf("[%s]", extractRefName(prop.Items.Ref)))
					} else if prop.Items.Type != "" {
						table.WriteString(fmt.Sprintf("[%s]", prop.Items.Type))
					}
				}
			}
			table.WriteString("</code></td>\n")

			// Description
			table.WriteString("<td>")
			if prop.Description != "" {
				table.WriteString(prop.Description)
			} else {
				table.WriteString("-")
			}
			table.WriteString("</td>\n")

			// Constraints (NEW COLUMN)
			table.WriteString("<td>")
			var constraints []string

			if isFieldRequired(fieldName, schema.Required) {
				constraints = append(constraints, "<strong>Required</strong>")
			}

			// Add min/max length constraints
			minLen := getIntValue(prop, "minLength")
			maxLen := getIntValue(prop, "maxLength")

			if minLen > 0 && maxLen > 0 {
				constraints = append(constraints, fmt.Sprintf("Length: %d-%d", minLen, maxLen))
			} else if minLen > 0 {
				constraints = append(constraints, fmt.Sprintf("Min length: %d", minLen))
			} else if maxLen > 0 {
				constraints = append(constraints, fmt.Sprintf("Max length: %d", maxLen))
			}

			// Add pattern if exists
			if prop.Pattern != "" {
				constraints = append(constraints, fmt.Sprintf("Pattern: <code>%s</code>", prop.Pattern))
			}

			if len(constraints) > 0 {
				table.WriteString(strings.Join(constraints, "<br/>"))
			} else {
				table.WriteString("-")
			}
			table.WriteString("</td>\n")

			// Example (NEW COLUMN)
			table.WriteString("<td>")
			if prop.Example != nil {
				exampleStr := fmt.Sprintf("%v", prop.Example)
				table.WriteString(fmt.Sprintf("<code>%s</code>", exampleStr))
			} else {
				table.WriteString("-")
			}
			table.WriteString("</td>\n")

			table.WriteString("</tr>\n")
		}

		table.WriteString("</table>\n")

		if len(schema.Required) > 0 {
			table.WriteString("<p><em>* indicates required field</em></p>\n")
		}
	} else {
		table.WriteString("<p><em>No properties defined for this schema</em></p>\n")
	}

	return table.String()
}

func resolveSchemaRef(ref string) *Schema {
	// Extract definition name from $ref
	// Supports both Swagger 2.0 (#/definitions/Name) and OpenAPI 3.0 (#/components/schemas/Name)
	parts := strings.Split(ref, "/")
	if len(parts) < 2 {
		return nil
	}

	defName := parts[len(parts)-1]

	// Try OpenAPI 3.0 components/schemas first
	if swaggerSpec.Components != nil && swaggerSpec.Components.Schemas != nil {
		if def, exists := swaggerSpec.Components.Schemas[defName]; exists {
			return &Schema{
				Type:       def.Type,
				Properties: def.Properties,
				Required:   def.Required,
			}
		}
	}

	// Fall back to Swagger 2.0 definitions
	if def, exists := swaggerSpec.Definitions[defName]; exists {
		return &Schema{
			Type:       def.Type,
			Properties: def.Properties,
			Required:   def.Required,
		}
	}

	return nil
}

func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func isFieldRequired(fieldName string, required []string) bool {
	for _, req := range required {
		if req == fieldName {
			return true
		}
	}
	return false
}

func getIntValue(prop Property, fieldName string) int {
	// This is a helper to extract minLength/maxLength from Property
	// Since Property struct needs to be extended
	switch fieldName {
	case "minLength":
		return prop.MinLength
	case "maxLength":
		return prop.MaxLength
	}
	return 0
}

func getParameterType(param Parameter) string {
	if param.Type != "" {
		typeStr := param.Type
		if param.Format != "" {
			typeStr += fmt.Sprintf(" (%s)", param.Format)
		}
		return typeStr
	}

	if param.Schema != nil {
		if param.Schema.Type != "" {
			return param.Schema.Type
		}
		if param.Schema.Ref != "" {
			parts := strings.Split(param.Schema.Ref, "/")
			return parts[len(parts)-1]
		}
	}

	return ""
}

func generateExampleJSON(schema *Schema) string {
	var example strings.Builder

	example.WriteString("<h4>Example JSON</h4>\n")
	example.WriteString("<ac:structured-macro ac:name=\"code\">\n")
	example.WriteString("<ac:parameter ac:name=\"language\">json</ac:parameter>\n")
	example.WriteString("<ac:plain-text-body><![CDATA[")

	// Resolve reference if present
	if schema.Ref != "" {
		resolvedSchema := resolveSchemaRef(schema.Ref)
		if resolvedSchema != nil {
			schema = resolvedSchema
		}
	}

	// Generate JSON from schema
	jsonStr := buildJSONFromSchema(schema, 0)
	example.WriteString(jsonStr)

	example.WriteString("]]></ac:plain-text-body>\n")
	example.WriteString("</ac:structured-macro>\n")

	return example.String()
}

func buildJSONFromSchema(schema *Schema, indentLevel int) string {
	if schema == nil {
		return ""
	}

	indent := strings.Repeat("  ", indentLevel)
	nextIndent := strings.Repeat("  ", indentLevel+1)

	var json strings.Builder

	// Handle array type
	if schema.Type == "array" && schema.Items != nil {
		json.WriteString("[\n")

		// Resolve array items if it's a reference
		itemSchema := schema.Items
		if itemSchema.Ref != "" {
			resolvedSchema := resolveSchemaRef(itemSchema.Ref)
			if resolvedSchema != nil {
				itemSchema = resolvedSchema
			}
		}

		json.WriteString(nextIndent)
		json.WriteString(buildJSONFromSchema(itemSchema, indentLevel+1))
		json.WriteString("\n")
		json.WriteString(indent)
		json.WriteString("]")
		return json.String()
	}

	// Handle object type
	if len(schema.Properties) > 0 {
		json.WriteString("{\n")

		// Sort properties for consistent output
		var propNames []string
		for name := range schema.Properties {
			propNames = append(propNames, name)
		}

		for i, propName := range propNames {
			prop := schema.Properties[propName]

			json.WriteString(nextIndent)
			json.WriteString(fmt.Sprintf("\"%s\": ", propName))

			// Get example value or generate default
			exampleValue := getExampleValue(prop, propName)
			json.WriteString(exampleValue)

			// Add comma if not last property
			if i < len(propNames)-1 {
				json.WriteString(",")
			}
			json.WriteString("\n")
		}

		json.WriteString(indent)
		json.WriteString("}")
	}

	return json.String()
}

func getExampleValue(prop Property, fieldName string) string {
	// Use example if available
	if prop.Example != nil {
		switch v := prop.Example.(type) {
		case string:
			return fmt.Sprintf("\"%s\"", escapeJSON(v))
		case float64, int:
			return fmt.Sprintf("%v", v)
		case bool:
			return fmt.Sprintf("%v", v)
		default:
			return fmt.Sprintf("\"%v\"", v)
		}
	}

	// Handle references
	if prop.Ref != "" {
		resolvedSchema := resolveSchemaRef(prop.Ref)
		if resolvedSchema != nil {
			return buildJSONFromSchema(resolvedSchema, 1)
		}
		return "\"...\""
	}

	// Handle arrays
	if prop.Type == "array" && prop.Items != nil {
		if prop.Items.Ref != "" {
			resolvedSchema := resolveSchemaRef(prop.Items.Ref)
			if resolvedSchema != nil {
				return fmt.Sprintf("[\n    %s\n  ]", buildJSONFromSchema(resolvedSchema, 2))
			}
		}
		return "[]"
	}

	// Generate default values based on type
	switch prop.Type {
	case "string":
		if prop.Format == "date" {
			return "\"2024-01-15\""
		} else if prop.Format == "date-time" {
			return "\"2024-01-15T10:30:00Z\""
		} else if prop.Format == "email" {
			return "\"user@example.com\""
		} else if strings.Contains(strings.ToLower(fieldName), "email") {
			return "\"user@example.com\""
		} else if strings.Contains(strings.ToLower(fieldName), "name") {
			return fmt.Sprintf("\"Sample %s\"", fieldName)
		}
		return "\"string\""
	case "integer", "number":
		return "0"
	case "boolean":
		return "false"
	case "object":
		return "{}"
	default:
		return "null"
	}
}

func escapeJSON(s string) string {
	// Basic JSON string escaping
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func createOrUpdateParentPage(config ConfluenceConfig, title, apiTitle string) (string, error) {
	content := fmt.Sprintf(`<h1>%s</h1>
<p>This page contains the API documentation for %s. Each endpoint has its own page below.</p>
<p><strong>Generated automatically from Swagger/OpenAPI specification</strong></p>
<p><ac:structured-macro ac:name="children">
<ac:parameter ac:name="all">true</ac:parameter>
</ac:structured-macro></p>`, apiTitle, apiTitle)

	pageID, err := createOrUpdateConfluencePage(config, title, content, "")
	return pageID, err
}

func createOrUpdateEndpointPage(config ConfluenceConfig, title, content, parentPageID string) (string, error) {
	return createOrUpdateConfluencePage(config, title, content, parentPageID)
}

func createOrUpdateConfluencePage(config ConfluenceConfig, title, content, parentPageID string) (string, error) {
	pageID, version, err := getPageByTitle(config, title)
	if err != nil {
		return "", err
	}

	page := ConfluencePage{
		Type:  "page",
		Title: title,
		Space: SpaceKey{Key: config.SpaceKey},
		Body: Body{
			Storage: Storage{
				Value:          content,
				Representation: "storage",
			},
		},
	}

	if parentPageID != "" {
		page.Ancestors = []PageAncestor{{ID: parentPageID}}
	}

	var method string
	var url string

	if pageID != "" {
		method = "PUT"
		url = fmt.Sprintf("%s/rest/api/content/%s", config.BaseURL, pageID)
		page.Version = &Version{Number: version + 1}
	} else {
		method = "POST"
		url = fmt.Sprintf("%s/rest/api/content", config.BaseURL)
	}

	jsonData, err := json.Marshal(page)
	if err != nil {
		return "", fmt.Errorf("failed to marshal page: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(config.Username, config.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("confluence API error: %s - Response: %s", resp.Status, string(respBody))
	}

	if len(respBody) == 0 {
		return "", fmt.Errorf("empty response from Confluence API")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w - Response body: %s", err, string(respBody))
	}

	resultPageID, ok := result["id"].(string)
	if !ok {
		return "", fmt.Errorf("page ID not found in response: %s", string(respBody))
	}

	return resultPageID, nil
}

func getPageByTitle(config ConfluenceConfig, title string) (string, int, error) {
	apiURL := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&title=%s&expand=version",
		config.BaseURL, config.SpaceKey, url.QueryEscape(title))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(config.Username, config.APIToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("confluence API error: %s - %s", resp.Status, string(body))
	}

	if len(body) == 0 {
		return "", 0, fmt.Errorf("empty response from Confluence API")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to parse JSON response: %w - Body: %s", err, string(body))
	}

	results, ok := result["results"].([]interface{})
	if !ok {
		return "", 0, fmt.Errorf("unexpected response format: missing 'results' field")
	}

	if len(results) > 0 {
		page, ok := results[0].(map[string]interface{})
		if !ok {
			return "", 0, fmt.Errorf("unexpected page format in results")
		}

		pageID, ok := page["id"].(string)
		if !ok {
			return "", 0, fmt.Errorf("page ID not found or invalid type")
		}

		versionMap, ok := page["version"].(map[string]interface{})
		if !ok {
			return "", 0, fmt.Errorf("version info not found or invalid type")
		}

		versionNum, ok := versionMap["number"].(float64)
		if !ok {
			return "", 0, fmt.Errorf("version number not found or invalid type")
		}

		return pageID, int(versionNum), nil
	}

	return "", 0, nil
}