package swagger

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantError  bool
	}{
		{
			name: "successful parse",
			response: `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				},
				"paths": {}
			}`,
			statusCode: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "invalid JSON",
			response:   `invalid json`,
			statusCode: http.StatusOK,
			wantError:  true,
		},
		{
			name:       "HTTP error",
			response:   `{}`,
			statusCode: http.StatusInternalServerError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			parser := NewParser()
			spec, err := parser.Parse(context.Background(), server.URL)

			if (err != nil) != tt.wantError {
				t.Errorf("Parse() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && spec.Info.Title != "Test API" {
				t.Errorf("expected title 'Test API', got '%s'", spec.Info.Title)
			}
		})
	}
}

func TestParser_ExtractEndpoints(t *testing.T) {
	spec := &Spec{
		Paths: map[string]PathItem{
			"/users": {
				"get": Operation{
					Summary: "Get Users",
				},
				"post": Operation{
					Summary: "Create User",
				},
			},
			"/users/{id}": {
				"get": Operation{
					Summary: "Get User",
				},
			},
		},
	}

	parser := NewParser()
	endpoints := parser.ExtractEndpoints(spec)

	if len(endpoints) != 3 {
		t.Errorf("expected 3 endpoints, got %d", len(endpoints))
	}
}

func TestGeneratePageTitle(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		method    string
		operation Operation
		want      string
	}{
		{
			name:   "with summary",
			path:   "/users",
			method: "get",
			operation: Operation{
				Summary: "Get All Users",
			},
			want: "Get All Users",
		},
		{
			name:   "with operation ID",
			path:   "/users",
			method: "get",
			operation: Operation{
				OperationID: "getAllUsers",
			},
			want: "Get All Users",
		},
		{
			name:      "from path",
			path:      "/users/{id}/posts",
			method:    "get",
			operation: Operation{},
			want:      "GET Users Id Posts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePageTitle(tt.path, tt.method, tt.operation)
			if got != tt.want {
				t.Errorf("generatePageTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}
