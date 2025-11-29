package confluence

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/ahmadimt/SwagFluence/internal/config"
)

type Client interface {
	CreateOrUpdatePage(ctx context.Context, title, content, parentPageID string) (string, error)
	CreateParentPage(ctx context.Context, apiTitle string) (string, error)
}

// Client handles Confluence API interactions
type ConfluenceClient struct {
	cfg        config.ConfluenceConfig
	httpClient *http.Client
}

// NewClient creates a new Confluence client
func NewClient(cfg config.ConfluenceConfig) Client {
	return &ConfluenceClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateOrUpdatePage creates or updates a Confluence page
func (c *ConfluenceClient) CreateOrUpdatePage(ctx context.Context, title, content, parentPageID string) (string, error) {
	if !c.cfg.Enabled {
		// Print to console if Confluence is disabled
		fmt.Printf("\n=== Page: %s ===\n%s\n\n", title, content)
		return "", nil
	}

	// Check if page exists
	existingPageID, version, err := c.findPageByTitle(ctx, title)
	if err != nil {
		return "", fmt.Errorf("failed to check existing page: %w", err)
	}

	page := Page{
		Type:  "page",
		Title: title,
		Space: Space{Key: c.cfg.SpaceKey},
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

	if existingPageID != "" {
		// Update existing page
		page.ID = existingPageID
		page.Version = &Version{Number: version + 1}
		return c.updatePage(ctx, &page)
	}

	// Create new page
	return c.createPage(ctx, &page)
}

// createPage creates a new page
func (c *ConfluenceClient) createPage(ctx context.Context, page *Page) (string, error) {
	apiURL := fmt.Sprintf("%s/rest/api/content", c.cfg.BaseURL)

	body, err := json.Marshal(page)
	if err != nil {
		return "", fmt.Errorf("failed to marshal page: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.cfg.Username, c.cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result Page
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", c.cfg.BaseURL, result.ID)
	fmt.Printf("✓ Created page: %s - %s\n", page.Title, pageURL)

	return result.ID, nil
}

// updatePage updates an existing page
func (c *ConfluenceClient) updatePage(ctx context.Context, page *Page) (string, error) {
	apiURL := fmt.Sprintf("%s/rest/api/content/%s", c.cfg.BaseURL, page.ID)

	body, err := json.Marshal(page)
	if err != nil {
		return "", fmt.Errorf("failed to marshal page: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.cfg.Username, c.cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to update page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	pageURL := fmt.Sprintf("%s/pages/viewpage.action?pageId=%s", c.cfg.BaseURL, page.ID)
	fmt.Printf("✓ Updated page: %s - %s\n", page.Title, pageURL)

	return page.ID, nil
}

// findPageByTitle finds a page by title
func (c *ConfluenceClient) findPageByTitle(ctx context.Context, title string) (string, int, error) {
	apiURL := fmt.Sprintf("%s/rest/api/content?spaceKey=%s&title=%s&expand=version",
		c.cfg.BaseURL, c.cfg.SpaceKey, url.QueryEscape(title))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.cfg.Username, c.cfg.APIToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("failed to search page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Results) == 0 {
		return "", 0, nil
	}

	page := result.Results[0]
	version := 0
	if page.Version != nil {
		version = page.Version.Number
	}

	return page.ID, version, nil
}

// CreateParentPage creates or updates the parent documentation page
func (c *ConfluenceClient) CreateParentPage(ctx context.Context, apiTitle string) (string, error) {
	title := fmt.Sprintf("%s - API Documentation", apiTitle)
	content := fmt.Sprintf(`<h1>%s</h1>
<p>This page contains the API documentation for %s. Each endpoint has its own page below.</p>
<p><strong>Generated automatically from Swagger/OpenAPI specification</strong></p>
<p><ac:structured-macro ac:name="children">
<ac:parameter ac:name="all">true</ac:parameter>
</ac:structured-macro></p>`, apiTitle, apiTitle)

	return c.CreateOrUpdatePage(ctx, title, content, "")
}
