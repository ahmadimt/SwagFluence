package swagger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Parser handles Swagger/OpenAPI specification parsing
type Parser struct {
	httpClient *http.Client
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Parse fetches and parses a Swagger/OpenAPI specification from a URL
func (p *Parser) Parse(ctx context.Context, url string) (*Spec, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swagger: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var spec Spec
	if err := json.Unmarshal(body, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse swagger: %w", err)
	}

	return &spec, nil
}

// ExtractEndpoints extracts all endpoints from a specification
func (p *Parser) ExtractEndpoints(spec *Spec) []EndpointInfo {
	var endpoints []EndpointInfo

	for path, pathItem := range spec.Paths {
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

	return endpoints
}

// isHTTPMethod checks if a string is a valid HTTP method
func isHTTPMethod(method string) bool {
	validMethods := map[string]bool{
		"get":     true,
		"post":    true,
		"put":     true,
		"delete":  true,
		"patch":   true,
		"options": true,
		"head":    true,
	}
	return validMethods[strings.ToLower(method)]
}

// generatePageTitle generates a page title for an endpoint
func generatePageTitle(path, method string, operation Operation) string {
	if operation.Summary != "" {
		return operation.Summary
	}

	if operation.OperationID != "" {
		return cleanOperationID(operation.OperationID)
	}

	return generateTitleFromPath(path, method)
}

// cleanOperationID converts operation ID to a readable title
func cleanOperationID(operationID string) string {
	// Replace underscores with spaces
	result := strings.ReplaceAll(operationID, "_", " ")

	// Add space before capital letters
	var builder strings.Builder
	for i, r := range result {
		if i > 0 && r >= 'A' && r <= 'Z' {
			builder.WriteRune(' ')
		}
		builder.WriteRune(r)
	}

	// Title case the result
	titleCaser := cases.Title(language.Und)
	return titleCaser.String(strings.ToLower(builder.String()))
}

// generateTitleFromPath generates a title from the path and method
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