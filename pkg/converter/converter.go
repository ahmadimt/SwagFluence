package converter

import (
	"context"
	"fmt"

	"github.com/ahmadimt/SwagFluence/internal/confluence"
	"github.com/ahmadimt/SwagFluence/internal/swagger"
)

// Converter orchestrates the conversion process
type Converter struct {
	parser    *swagger.Parser
	client    confluence.Client
	formatter *confluence.Formatter
}

// New creates a new Converter
func New(parser *swagger.Parser, client confluence.Client) *Converter {
	return &Converter{
		parser:    parser,
		client:    client,
		formatter: confluence.NewFormatter(),
	}
}

// Convert performs the full conversion from Swagger to Confluence
func (c *Converter) Convert(ctx context.Context, swaggerURL string) error {
	fmt.Printf("Fetching Swagger specification from: %s\n", swaggerURL)

	// Parse Swagger specification
	spec, err := c.parser.Parse(ctx, swaggerURL)
	if err != nil {
		return fmt.Errorf("failed to parse swagger: %w", err)
	}

	fmt.Printf("Successfully parsed: %s v%s\n", spec.Info.Title, spec.Info.Version)

	// Extract endpoints
	endpoints := c.parser.ExtractEndpoints(spec)
	fmt.Printf("Found %d endpoints\n\n", len(endpoints))

	// Create resolver for $ref resolution
	resolver := swagger.NewResolver(spec)

	// Create parent page if Confluence is enabled
	parentPageID := ""
	if c.client != nil {
		var err error
		parentPageID, err = c.client.CreateParentPage(ctx, spec.Info.Title)
		if err != nil {
			return fmt.Errorf("failed to create parent page: %w", err)
		}
		if parentPageID != "" {
			fmt.Printf("Parent page ID: %s\n\n", parentPageID)
		}
	}

	// Process each endpoint
	successCount := 0
	for i, endpoint := range endpoints {
		fmt.Printf("[%d/%d] Processing: %s %s\n", i+1, len(endpoints),
			endpoint.Method, endpoint.Path)

		if err := c.processEndpoint(ctx, resolver, endpoint, parentPageID); err != nil {
			return fmt.Errorf("failed to process %s %s: %w", endpoint.Method, endpoint.Path, err)
		}

		successCount++
	}

	fmt.Printf("\n=================================\n")
	fmt.Printf("Summary: %d/%d pages processed successfully\n", successCount, len(endpoints))

	return nil
}

func (c *Converter) processEndpoint(ctx context.Context, resolver *swagger.Resolver, endpoint swagger.EndpointInfo, parentPageID string) error {
	// Generate Confluence markup
	content := c.formatter.FormatEndpointPage(endpoint.Path, endpoint.Method, endpoint.Operation, resolver)

	// Create/update page
	_, err := c.client.CreateOrUpdatePage(ctx, endpoint.Title, content, parentPageID)
	if err != nil {
		return fmt.Errorf("failed to create/update page: %w", err)
	}

	return nil
}