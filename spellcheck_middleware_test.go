package spellcheck

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestNewMiddleware(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker failed: %v", err)
	}

	middleware, err := NewMiddleware(&config, checker)
	if err != nil {
		t.Fatalf("NewMiddleware failed: %v", err)
	}

	if middleware == nil {
		t.Fatal("NewMiddleware returned nil")
	}

	if !middleware.IsEnabled() {
		t.Error("Expected middleware to be enabled")
	}
}

func TestMiddleware_Validate(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker failed: %v", err)
	}

	middleware, err := NewMiddleware(&config, checker)
	if err != nil {
		t.Fatalf("NewMiddleware failed: %v", err)
	}

	app := fiber.New()
	app.Use(middleware.Validate())

	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	tests := []struct {
		name           string
		method         string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name:   "valid text",
			method: http.MethodPost,
			body: map[string]interface{}{
				"title":   "Hello World",
				"content": "This is a test",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "misspelled text",
			method: http.MethodPost,
			body: map[string]interface{}{
				"title": "Teh quik test",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			method:         http.MethodPost,
			body:           map[string]interface{}{},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "non-text fields only",
			method: http.MethodPost,
			body: map[string]interface{}{
				"id":    "123",
				"count": 42,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request - skip validation",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed, // No GET handler defined
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req = httptest.NewRequest(tt.method, "/test", bytes.NewReader(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, "/test", nil)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestMiddleware_ValidateStruct(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker failed: %v", err)
	}

	middleware, err := NewMiddleware(&config, checker)
	if err != nil {
		t.Fatalf("NewMiddleware failed: %v", err)
	}

	type Article struct {
		Title   string `json:"title" spellcheck:"true"`
		Content string `json:"content" spellcheck:"true"`
		Slug    string `json:"slug"` // No spellcheck
	}

	tests := []struct {
		name        string
		article     Article
		expectError bool
	}{
		{
			name: "valid article",
			article: Article{
				Title:   "Hello World",
				Content: "This is a test",
				Slug:    "hello-world",
			},
			expectError: false,
		},
		{
			name: "misspelled title",
			article: Article{
				Title:   "Teh quik test",
				Content: "Valid content",
			},
			expectError: true,
		},
		{
			name: "misspelled content",
			article: Article{
				Title:   "Valid title",
				Content: "Teh wrld",
			},
			expectError: true,
		},
		{
			name: "empty fields",
			article: Article{
				Title:   "",
				Content: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := middleware.ValidateStruct(tt.article)
			if err != nil {
				t.Fatalf("ValidateStruct failed: %v", err)
			}

			if tt.expectError && !errors.HasErrors() {
				t.Error("Expected spelling errors but got none")
			}

			if !tt.expectError && errors.HasErrors() {
				t.Errorf("Expected no errors but got: %v", errors)
			}
		})
	}
}

func TestMiddleware_ShouldCheckField(t *testing.T) {
	config := DefaultConfig()
	checker, _ := NewSpellchecker(&config)
	middleware, _ := NewMiddleware(&config, checker)

	tests := []struct {
		fieldName string
		expected  bool
	}{
		{"title", true},
		{"content", true},
		{"description", true},
		{"body", true},
		{"text", true},
		{"message", true},
		{"id", false},
		{"slug", false},
		{"email", false},
		{"password", false},
		{"username", false},
	}

	for _, tt := range tests {
		t.Run(tt.fieldName, func(t *testing.T) {
			result := middleware.shouldCheckField(tt.fieldName)
			if result != tt.expected {
				t.Errorf("shouldCheckField(%q) = %v, want %v", tt.fieldName, result, tt.expected)
			}
		})
	}
}

func TestMiddleware_ValidateStructPointer(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker failed: %v", err)
	}

	middleware, err := NewMiddleware(&config, checker)
	if err != nil {
		t.Fatalf("NewMiddleware failed: %v", err)
	}

	type Article struct {
		Title string `json:"title" spellcheck:"true"`
	}

	article := &Article{
		Title: "Teh test",
	}

	errors, err := middleware.ValidateStruct(article)
	if err != nil {
		t.Fatalf("ValidateStruct failed: %v", err)
	}

	if !errors.HasErrors() {
		t.Error("Expected spelling errors for pointer to struct")
	}
}

func TestMiddleware_NonJSONContentType(t *testing.T) {
	config := DefaultConfig()
	checker, _ := NewSpellchecker(&config)
	middleware, _ := NewMiddleware(&config, checker)

	app := fiber.New()
	app.Use(middleware.Validate())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("plain text")))
	req.Header.Set("Content-Type", "text/plain")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should pass through without validation
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}
