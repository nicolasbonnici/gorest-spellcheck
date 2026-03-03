package spellcheck

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/logger"
)

// Middleware returns a Fiber middleware that validates request bodies for spelling errors
type Middleware struct {
	config       *Config
	spellchecker *Spellchecker
	tagParser    *TagParser
}

// NewMiddleware creates a new Middleware instance
func NewMiddleware(config *Config, spellchecker *Spellchecker) (*Middleware, error) {
	return &Middleware{
		config:       config,
		spellchecker: spellchecker,
		tagParser:    NewTagParser(),
	}, nil
}

// Validate is the middleware handler that checks spelling in request bodies
func (m *Middleware) Validate() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only validate POST, PUT, PATCH requests
		method := c.Method()
		if method != fiber.MethodPost && method != fiber.MethodPut && method != fiber.MethodPatch {
			return c.Next()
		}

		// Only validate if Content-Type is JSON
		contentType := string(c.Request().Header.ContentType())
		if contentType != "application/json" && contentType != "" {
			return c.Next()
		}

		// Read the body (Fiber's c.Body() returns []byte directly)
		bodyBytes := c.Body()

		// If body is empty, skip validation
		if len(bodyBytes) == 0 {
			return c.Next()
		}

		// Restore the body for the next handler (Fiber already keeps it)
		c.Request().SetBodyRaw(bodyBytes)

		// Parse JSON body
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
			// Not valid JSON, let the next handler deal with it
			return c.Next()
		}

		// Try to determine the model type from the route
		// For now, we'll check all string fields in the map
		// A more sophisticated approach would use route context to know the expected model type

		// Find string fields in the body
		spellingErrors := &SpellingErrors{}

		for fieldName, value := range bodyMap {
			// Only check string values
			strValue, ok := value.(string)
			if !ok {
				continue
			}

			// Skip empty strings
			if strValue == "" {
				continue
			}

			// Check if this looks like a field that should be spellchecked
			// We could make this more sophisticated by checking against known model types
			// For now, check common field names that typically contain text
			if m.shouldCheckField(fieldName) {
				errors, err := m.spellchecker.CheckField(fieldName, strValue)
				if err != nil {
					logger.Log.Error("Spell check failed", "field", fieldName, "error", err)
					continue
				}

				if errors.HasErrors() {
					spellingErrors.Errors = append(spellingErrors.Errors, errors.Errors...)
				}
			}
		}

		// If we found spelling errors, return 400
		if spellingErrors.HasErrors() {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":  "Spelling errors found",
				"errors": spellingErrors.Errors,
			})
		}

		// No errors, continue to next handler
		return c.Next()
	}
}

// ValidateStruct validates a specific struct instance for spelling errors
// This can be called manually from handlers if you want explicit validation
func (m *Middleware) ValidateStruct(v interface{}) (*SpellingErrors, error) {
	// Parse the struct to get fields with spellcheck tags
	fieldInfos := m.tagParser.Parse(v)

	if len(fieldInfos) == 0 {
		// No fields to check
		return &SpellingErrors{}, nil
	}

	spellingErrors := &SpellingErrors{}

	// Check each field
	for _, fieldInfo := range fieldInfos {
		// Get the field value
		value, ok := m.tagParser.GetFieldValue(v, fieldInfo.Name)
		if !ok {
			continue
		}

		// Skip empty strings
		if value == "" {
			continue
		}

		// Check spelling
		errors, err := m.spellchecker.CheckField(fieldInfo.JSONName, value)
		if err != nil {
			return nil, err
		}

		if errors.HasErrors() {
			spellingErrors.Errors = append(spellingErrors.Errors, errors.Errors...)
		}
	}

	return spellingErrors, nil
}

// shouldCheckField determines if a field should be spellchecked based on its name
// This is a heuristic approach - ideally we'd use struct tags
func (m *Middleware) shouldCheckField(fieldName string) bool {
	// Common field names that typically contain text to spellcheck
	checkFields := map[string]bool{
		"title":       true,
		"content":     true,
		"description": true,
		"body":        true,
		"text":        true,
		"message":     true,
		"comment":     true,
		"note":        true,
		"summary":     true,
		"excerpt":     true,
		"caption":     true,
		"bio":         true,
		"about":       true,
	}

	return checkFields[fieldName]
}

// Enable/disable middleware at runtime
func (m *Middleware) IsEnabled() bool {
	return m.config.Enabled
}
