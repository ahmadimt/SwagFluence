package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ahmadimt/SwagFluence/internal/config"
	"github.com/ahmadimt/SwagFluence/internal/confluence"
	"github.com/ahmadimt/SwagFluence/internal/swagger"
	"github.com/ahmadimt/SwagFluence/pkg/converter"
)

const (
	exitCodeSuccess = 0
	exitCodeError   = 1
)

func main() {
	os.Exit(run())
}

func run() int {
	// Setup context with cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Parse command line arguments
	if len(os.Args) < 2 {
		printUsage()
		return exitCodeError
	}

	swaggerURL := os.Args[1]

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return exitCodeError
	}

	// Initialize components
	swaggerParser := swagger.NewParser()
	confluenceClient := confluence.NewClient(cfg.Confluence)
	conv := converter.New(swaggerParser, confluenceClient)

	// Execute conversion
	if err := conv.Convert(ctx, swaggerURL); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitCodeError
	}

	return exitCodeSuccess
}

func printUsage() {
	fmt.Println("Usage: swagfluence <swagger-url>")
	fmt.Println("\nExample:")
	fmt.Println("  swagfluence https://petstore.swagger.io/v2/swagger.json")
	fmt.Println("\nEnvironment variables (optional for Confluence integration):")
	fmt.Println("  CONFLUENCE_BASE_URL       - Base URL of your Confluence instance")
	fmt.Println("  CONFLUENCE_USERNAME       - Your Confluence username/email")
	fmt.Println("  CONFLUENCE_API_TOKEN      - Your Confluence API token")
	fmt.Println("  CONFLUENCE_SPACE_KEY      - Space key where pages will be created")
	fmt.Println("  CONFLUENCE_PARENT_PAGE_ID - (Optional) Parent page ID for documentation")
	fmt.Println("  CONFLUENCE_ENABLED        - Whether write to Confluence")
}