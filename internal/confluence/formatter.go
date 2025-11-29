package confluence

import (
	"fmt"
	"strings"

	"github.com/ahmadimt/SwagFluence/internal/example"
	"github.com/ahmadimt/SwagFluence/internal/swagger"
)

// Formatter generates Confluence storage format markup
type Formatter struct {
	exampleGen *example.Generator
}

// NewFormatter creates a new Formatter
func NewFormatter() *Formatter {
	return &Formatter{
		exampleGen: example.NewGenerator(),
	}
}

// FormatEndpointPage generates markup for an endpoint page
func (f *Formatter) FormatEndpointPage(path, method string, op swagger.Operation, resolver *swagger.Resolver) string {
	var sb strings.Builder

	// Add layout section for full width
	sb.WriteString("<ac:layout>\n")
	sb.WriteString("<ac:layout-section ac:type=\"single\">\n")
	sb.WriteString("<ac:layout-cell>\n")

	// Header with method badge
	sb.WriteString("<h2>")
	sb.WriteString(f.methodBadge(method))
	sb.WriteString(fmt.Sprintf(" %s</h2>\n", path))

	// Description
	if op.Description != "" {
		sb.WriteString(fmt.Sprintf("<p>%s</p>\n", op.Description))
	}

	// Operation ID
	if op.OperationID != "" {
		sb.WriteString(fmt.Sprintf("<p><strong>Operation ID:</strong> <code>%s</code></p>\n", op.OperationID))
	}

	// Tags
	if len(op.Tags) > 0 {
		sb.WriteString(f.formatTags(op.Tags))
	}

	// Content types
	if len(op.Consumes) > 0 {
		sb.WriteString(fmt.Sprintf("<p><strong>Consumes:</strong> <code>%s</code></p>\n", strings.Join(op.Consumes, ", ")))
	}
	if len(op.Produces) > 0 {
		sb.WriteString(fmt.Sprintf("<p><strong>Produces:</strong> <code>%s</code></p>\n", strings.Join(op.Produces, ", ")))
	}

	// Request body section
	sb.WriteString(f.formatRequestBodySection(op, resolver))

	// Parameters section
	sb.WriteString(f.formatParametersSection(op.Parameters))

	// Close layout
	sb.WriteString("</ac:layout-cell>\n")
	sb.WriteString("</ac:layout-section>\n")
	sb.WriteString("</ac:layout>\n")

	return sb.String()
}

// methodBadge creates a colored status badge for HTTP method
func (f *Formatter) methodBadge(method string) string {
	colors := map[string]string{
		"GET":    "Blue",
		"POST":   "Green",
		"PUT":    "Yellow",
		"DELETE": "Red",
		"PATCH":  "Purple",
	}

	color, ok := colors[strings.ToUpper(method)]
	if !ok {
		color = "Grey"
	}

	return fmt.Sprintf("<ac:structured-macro ac:name=\"status\">"+
		"<ac:parameter ac:name=\"colour\">%s</ac:parameter>"+
		"<ac:parameter ac:name=\"title\">%s</ac:parameter>"+
		"</ac:structured-macro>", color, strings.ToUpper(method))
}

// formatTags formats API tags
func (f *Formatter) formatTags(tags []string) string {
	var sb strings.Builder
	sb.WriteString("<p><strong>Tags:</strong> ")
	for i, tag := range tags {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("<ac:structured-macro ac:name=\"status\">")
		sb.WriteString("<ac:parameter ac:name=\"colour\">Grey</ac:parameter>")
		sb.WriteString(fmt.Sprintf("<ac:parameter ac:name=\"title\">%s</ac:parameter>", tag))
		sb.WriteString("</ac:structured-macro>")
	}
	sb.WriteString("</p>\n")
	return sb.String()
}

func (f *Formatter) formatRequestBodySection(op swagger.Operation, resolver *swagger.Resolver) string {
	var sb strings.Builder

	// Check for body parameter (Swagger 2.0)
	var bodyParam *swagger.Parameter
	for i := range op.Parameters {
		if op.Parameters[i].In == "body" {
			bodyParam = &op.Parameters[i]
			break
		}
	}

	// Check for requestBody (OpenAPI 3.0)
	hasRequestBody := op.RequestBody != nil || bodyParam != nil
	if !hasRequestBody {
		return ""
	}

	sb.WriteString("<h3>Request Body</h3>\n")

	var schemaToUse *swagger.Schema

	// Handle OpenAPI 3.0 requestBody
	if op.RequestBody != nil {
		if op.RequestBody.Description != "" {
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", op.RequestBody.Description))
		}

		if op.RequestBody.Required {
			sb.WriteString(f.requiredBadge())
		}

		for contentType, mediaType := range op.RequestBody.Content {
			sb.WriteString(fmt.Sprintf("<p><strong>Content-Type:</strong> <code>%s</code></p>\n", contentType))
			schemaToUse = &mediaType.Schema
			resolvedSchema, _ := resolver.ResolveSchema(&mediaType.Schema)
			if resolvedSchema != nil {
				sb.WriteString(f.formatSchemaTable(resolvedSchema))
			}
		}
	}

	// Handle Swagger 2.0 body parameter
	if bodyParam != nil {
		if bodyParam.Description != "" {
			sb.WriteString(fmt.Sprintf("<p>%s</p>\n", bodyParam.Description))
		}

		if bodyParam.Required {
			sb.WriteString(f.requiredBadge())
		}

		if bodyParam.Schema != nil {
			schemaToUse = bodyParam.Schema
			resolvedSchema, _ := resolver.ResolveSchema(bodyParam.Schema)
			if resolvedSchema != nil {
				sb.WriteString(f.formatSchemaTable(resolvedSchema))
			}
		}
	}

	// Add Example JSON section
	if schemaToUse != nil {
		resolvedSchema, _ := resolver.ResolveSchema(schemaToUse)
		if resolvedSchema != nil {
			exampleJSON := f.exampleGen.GenerateExampleJSON(resolvedSchema)
			sb.WriteString(f.formatExampleJSON(exampleJSON))
		}
	}

	return sb.String()
}

// formatParametersSection formats the parameters table
func (f *Formatter) formatParametersSection(params []swagger.Parameter) string {
	var sb strings.Builder

	sb.WriteString("<h3>Parameters</h3>\n")
	sb.WriteString("<table>\n")
	sb.WriteString("<tr><th>Parameter</th><th>Description</th></tr>\n")

	hasNonBodyParams := false
	for _, param := range params {
		if param.In != "body" {
			hasNonBodyParams = true
			sb.WriteString(f.formatParameter(param))
		}
	}

	if !hasNonBodyParams {
		sb.WriteString("<tr>\n")
		sb.WriteString("<td colspan=\"2\"><em>This endpoint requires no parameters</em></td>\n")
		sb.WriteString("</tr>\n")
	}

	sb.WriteString("</table>\n")
	return sb.String()
}

// formatParameter formats a single parameter row
func (f *Formatter) formatParameter(param swagger.Parameter) string {
	var sb strings.Builder

	sb.WriteString("<tr>\n")
	sb.WriteString(fmt.Sprintf("<td><code>%s</code></td>\n", param.Name))
	sb.WriteString("<td>")

	// Required badge
	if param.Required {
		sb.WriteString(f.requiredBadge())
	} else {
		sb.WriteString(f.optionalBadge())
	}

	sb.WriteString("<br/><br/>")

	// Description
	if param.Description != "" {
		sb.WriteString(param.Description)
	} else {
		sb.WriteString("No description provided")
	}

	// Type
	paramType := getParameterType(param)
	if paramType != "" {
		sb.WriteString(fmt.Sprintf("<br/><br/><strong>Type:</strong> <code>%s</code>", paramType))
	}

	// Location
	if param.In != "" {
		sb.WriteString(fmt.Sprintf("<br/><br/><strong>Location:</strong> %s", param.In))
	}

	sb.WriteString("</td>\n")
	sb.WriteString("</tr>\n")

	return sb.String()
}

// formatSchemaTable formats a schema as an HTML table
func (f *Formatter) formatSchemaTable(schema *swagger.Schema) string {
	if schema == nil || len(schema.Properties) == 0 {
		return "<p><em>No properties defined for this schema</em></p>\n"
	}

	var sb strings.Builder

	// Handle array type
	if schema.Type == "array" && schema.Items != nil {
		sb.WriteString("<p><strong>Type:</strong> Array</p>\n")
		if schema.Items.Ref != "" {
			sb.WriteString(fmt.Sprintf("<p><strong>Items:</strong> %s</p>\n", swagger.ExtractRefName(schema.Items.Ref)))
		}
	}

	sb.WriteString("<table>\n")
	sb.WriteString("<tr><th>Field</th><th>Type</th><th>Description</th><th>Constraints</th><th>Example</th></tr>\n")

	for fieldName, prop := range schema.Properties {
		sb.WriteString(f.formatPropertyRow(fieldName, prop, schema.Required))
	}

	sb.WriteString("</table>\n")

	if len(schema.Required) > 0 {
		sb.WriteString("<p><em>* indicates required field</em></p>\n")
	}

	return sb.String()
}

// formatPropertyRow formats a single property row in the schema table
func (f *Formatter) formatPropertyRow(fieldName string, prop swagger.Property, required []string) string {
	var sb strings.Builder

	sb.WriteString("<tr>\n")

	// Field name with required indicator
	sb.WriteString("<td><code>")
	sb.WriteString(fieldName)
	if isFieldRequired(fieldName, required) {
		sb.WriteString(" *")
	}
	sb.WriteString("</code></td>\n")

	// Type
	sb.WriteString("<td><code>")
	sb.WriteString(getPropertyType(prop))
	sb.WriteString("</code></td>\n")

	// Description
	sb.WriteString("<td>")
	if prop.Description != "" {
		sb.WriteString(prop.Description)
	} else {
		sb.WriteString("-")
	}
	sb.WriteString("</td>\n")

	// Constraints
	sb.WriteString("<td>")
	sb.WriteString(formatConstraints(fieldName, prop, required))
	sb.WriteString("</td>\n")

	// Example
	sb.WriteString("<td>")
	if prop.Example != nil {
		sb.WriteString(fmt.Sprintf("<code>%v</code>", prop.Example))
	} else {
		sb.WriteString("-")
	}
	sb.WriteString("</td>\n")

	sb.WriteString("</tr>\n")

	return sb.String()
}

// formatExampleJSON formats example JSON in a code block
func (f *Formatter) formatExampleJSON(exampleJSON string) string {
	var sb strings.Builder

	sb.WriteString("<h4>Example JSON</h4>\n")
	sb.WriteString("<ac:structured-macro ac:name=\"code\">\n")
	sb.WriteString("<ac:parameter ac:name=\"language\">json</ac:parameter>\n")
	sb.WriteString("<ac:plain-text-body><![CDATA[")
	sb.WriteString(exampleJSON)
	sb.WriteString("]]></ac:plain-text-body>\n")
	sb.WriteString("</ac:structured-macro>\n")

	return sb.String()
}

// Helper functions

func (f *Formatter) requiredBadge() string {
	return "<ac:structured-macro ac:name=\"status\">" +
		"<ac:parameter ac:name=\"colour\">Red</ac:parameter>" +
		"<ac:parameter ac:name=\"title\">REQUIRED</ac:parameter>" +
		"</ac:structured-macro>\n"
}

func (f *Formatter) optionalBadge() string {
	return "<ac:structured-macro ac:name=\"status\">" +
		"<ac:parameter ac:name=\"colour\">Green</ac:parameter>" +
		"<ac:parameter ac:name=\"title\">OPTIONAL</ac:parameter>" +
		"</ac:structured-macro>"
}

func getParameterType(param swagger.Parameter) string {
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
			return swagger.ExtractRefName(param.Schema.Ref)
		}
	}

	return ""
}

func getPropertyType(prop swagger.Property) string {
	if prop.Ref != "" {
		return swagger.ExtractRefName(prop.Ref)
	}

	typeStr := prop.Type
	if prop.Format != "" {
		typeStr += fmt.Sprintf(" (%s)", prop.Format)
	}

	if prop.Type == "array" && prop.Items != nil {
		if prop.Items.Ref != "" {
			typeStr += fmt.Sprintf("[%s]", swagger.ExtractRefName(prop.Items.Ref))
		} else if prop.Items.Type != "" {
			typeStr += fmt.Sprintf("[%s]", prop.Items.Type)
		}
	}

	return typeStr
}

func formatConstraints(fieldName string, prop swagger.Property, required []string) string {
	var constraints []string

	if isFieldRequired(fieldName, required) {
		constraints = append(constraints, "<strong>Required</strong>")
	}

	if prop.MinLength > 0 && prop.MaxLength > 0 {
		constraints = append(constraints, fmt.Sprintf("Length: %d-%d", prop.MinLength, prop.MaxLength))
	} else if prop.MinLength > 0 {
		constraints = append(constraints, fmt.Sprintf("Min length: %d", prop.MinLength))
	} else if prop.MaxLength > 0 {
		constraints = append(constraints, fmt.Sprintf("Max length: %d", prop.MaxLength))
	}

	if prop.Pattern != "" {
		constraints = append(constraints, fmt.Sprintf("Pattern: <code>%s</code>", prop.Pattern))
	}

	if len(constraints) > 0 {
		return strings.Join(constraints, "<br/>")
	}
	return "-"
}

func isFieldRequired(fieldName string, required []string) bool {
	for _, req := range required {
		if req == fieldName {
			return true
		}
	}
	return false
}