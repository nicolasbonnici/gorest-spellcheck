package spellcheck

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestPluginIntegration(t *testing.T) {
	// Create and initialize plugin
	plugin := NewPlugin()
	config := map[string]interface{}{
		"enabled":          true,
		"default_language": "en",
		"max_text_length":  10000,
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize plugin: %v", err)
	}

	// Create Fiber app with plugin middleware
	app := fiber.New()
	app.Use(plugin.Handler())

	// Setup plugin endpoints (cast to *SpellcheckPlugin to access SetupEndpoints)
	spellcheckPlugin := plugin.(*SpellcheckPlugin)
	err = spellcheckPlugin.SetupEndpoints(app)
	if err != nil {
		t.Fatalf("Failed to setup endpoints: %v", err)
	}

	// Add a test endpoint that accepts articles
	app.Post("/articles", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"message": "Article created successfully",
		})
	})

	t.Run("middleware blocks misspelled text in POST request", func(t *testing.T) {
		body := map[string]interface{}{
			"title":   "Teh quik brown fox",
			"content": "This has speling erors",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)

		if response["error"] != "Spelling errors found" {
			t.Errorf("Expected spelling error message, got: %v", response["error"])
		}

		if response["errors"] == nil {
			t.Error("Expected errors field in response")
		}
	})

	t.Run("middleware allows correct spelling", func(t *testing.T) {
		body := map[string]interface{}{
			"title":   "The quick brown fox",
			"content": "This is correct text",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusCreated {
			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			t.Errorf("Expected status 201, got %d. Response: %+v", resp.StatusCode, response)
		}
	})

	t.Run("on-demand spellcheck endpoint works", func(t *testing.T) {
		body := CheckRequest{
			Text: "Teh quik test",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/spellcheck", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}

		var response CheckResponse
		json.NewDecoder(resp.Body).Decode(&response)

		if response.Valid {
			t.Error("Expected valid to be false")
		}

		if len(response.Errors) == 0 {
			t.Error("Expected errors to be returned")
		}
	})

	t.Run("on-demand endpoint with correct text", func(t *testing.T) {
		body := CheckRequest{
			Text: "The quick brown fox jumps over the lazy dog",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/spellcheck", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var response CheckResponse
		json.NewDecoder(resp.Body).Decode(&response)

		if !response.Valid {
			t.Errorf("Expected valid to be true, got errors: %v", response.Errors)
		}
	})
}

func TestPluginIntegrationDisabled(t *testing.T) {
	// Create and initialize plugin with middleware disabled
	plugin := NewPlugin()
	config := map[string]interface{}{
		"enabled":          false,
		"default_language": "en",
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Failed to initialize plugin: %v", err)
	}

	// Create Fiber app with plugin middleware
	app := fiber.New()
	app.Use(plugin.Handler())

	app.Post("/articles", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"message": "Article created successfully",
		})
	})

	t.Run("middleware disabled - allows misspelled text", func(t *testing.T) {
		body := map[string]interface{}{
			"title":   "Teh quik brown fox",
			"content": "This has speling erors",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/articles", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}

		// Should pass through without validation when disabled
		if resp.StatusCode != http.StatusCreated {
			t.Errorf("Expected status 201 (pass-through), got %d", resp.StatusCode)
		}
	})
}
