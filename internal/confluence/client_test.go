package confluence

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ahmadimt/SwagFluence/internal/config"
)

type MockClient struct {
	cfg config.ConfluenceConfig
}

func NewMockClient(cfg config.ConfluenceConfig) Client {
	return &MockClient{
		cfg: cfg,
	}
}

func (m *MockClient) CreateOrUpdatePage(ctx context.Context, title, content, parentPageID string) (string, error) {

	switch title {
	case "Test Page":
		if m.cfg.Enabled {
			return "12345", nil
		}
	}
	return "", nil

}

func (m *MockClient) CreateParentPage(ctx context.Context, apiTitle string) (string, error) {
	return "", nil
}

func TestClient_CreateOrUpdatePage_Disabled(t *testing.T) {

	cfg := config.ConfluenceConfig{
		Enabled: false,
	}
	client := NewMockClient(cfg)
	pageID, err := client.CreateOrUpdatePage(context.Background(), "Test Page", "Content", "")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if pageID != "" {
		t.Errorf("expected empty pageID when disabled, got %s", pageID)
	}
}

func TestClient_CreatePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "12345", "title": "Test Page"}`))
	}))
	defer server.Close()

	cfg := config.ConfluenceConfig{
		BaseURL:  server.URL,
		Username: "user",
		APIToken: "token",
		SpaceKey: "TEST",
		Enabled:  true,
	}

	client := NewMockClient(cfg)
	pageID, err := client.CreateOrUpdatePage(context.Background(), "Test Page", "Content", "")

	if err != nil {
		t.Fatalf("CreateOrUpdatePage() error = %v", err)
	}

	if pageID != "12345" {
		t.Errorf("expected pageID '12345', got '%s'", pageID)
	}
}
