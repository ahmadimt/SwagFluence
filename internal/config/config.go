package config

import (
	"os"
)

// Config holds all application configuration
type Config struct {
	Confluence ConfluenceConfig
}

// ConfluenceConfig holds Confluence-specific settings
type ConfluenceConfig struct {
	BaseURL      string
	Username     string
	APIToken     string
	SpaceKey     string
	ParentPageID string
	Enabled      bool
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		Confluence: ConfluenceConfig{
			BaseURL:      os.Getenv("CONFLUENCE_BASE_URL"),
			Username:     os.Getenv("CONFLUENCE_USERNAME"),
			APIToken:     os.Getenv("CONFLUENCE_API_TOKEN"),
			SpaceKey:     os.Getenv("CONFLUENCE_SPACE_KEY"),
			ParentPageID: os.Getenv("CONFLUENCE_PARENT_PAGE_ID"),
		},
	}

	// Enable Confluence only if all required fields are present
	cfg.Confluence.Enabled = cfg.Confluence.BaseURL != "" &&
		cfg.Confluence.Username != "" &&
		cfg.Confluence.APIToken != "" &&
		cfg.Confluence.SpaceKey != ""

	return cfg, nil
}

// IsConfluenceEnabled returns true if Confluence integration is enabled
func (c *Config) IsConfluenceEnabled() bool {
	return c.Confluence.Enabled
}
