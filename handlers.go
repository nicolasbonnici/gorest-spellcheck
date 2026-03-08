package spellcheck

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/logger"
)

// Handler handles HTTP requests for the spellcheck endpoint
type Handler struct {
	config       *Config
	spellchecker *Spellchecker
}

// NewHandler creates a new Handler instance
func NewHandler(config *Config, spellchecker *Spellchecker) *Handler {
	return &Handler{
		config:       config,
		spellchecker: spellchecker,
	}
}

// Check handles POST /api/spellcheck requests for on-demand spell checking
func (h *Handler) Check(c *fiber.Ctx) error {
	// Parse request body
	var req CheckRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Log.Error("Failed to parse spell check request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		logger.Log.Error("Spell check request validation failed", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check language support
	if req.Language != "" && req.Language != h.config.DefaultLanguage {
		// Check if language is supported
		supported := false
		for _, lang := range h.config.SupportedLanguages {
			if lang == req.Language {
				supported = true
				break
			}
		}
		if !supported {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": &UnsupportedLanguageError{
					Language:           req.Language,
					SupportedLanguages: h.config.SupportedLanguages,
				},
			})
		}
	}

	// Create a temporary spellchecker with context words if provided
	checker := h.spellchecker
	if len(req.Context) > 0 || req.Options != nil {
		// Create modified config
		tempConfig := *h.config

		// Add context words to ignored words
		if len(req.Context) > 0 {
			tempConfig.IgnoredWords = append(tempConfig.IgnoredWords, req.Context...)
		}

		// Apply options
		if req.Options != nil {
			if req.Options.CaseSensitive != nil {
				tempConfig.CaseSensitive = *req.Options.CaseSensitive
			}
			if req.Options.MaxSuggestions != nil {
				tempConfig.MaxSuggestions = *req.Options.MaxSuggestions
			}
		}

		// Create temporary checker
		var err error
		checker, err = NewSpellchecker(&tempConfig)
		if err != nil {
			logger.Log.Error("Failed to create temporary spellchecker", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to initialize spell checker",
			})
		}
	}

	// Check spelling
	errors, err := checker.Check(req.Text)
	if err != nil {
		// Handle specific error types
		if textErr, ok := err.(*TextTooLongError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": textErr.Error(),
			})
		}

		logger.Log.Error("Spell check failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Spell check failed",
		})
	}

	// Build response
	response := CheckResponse{
		Valid: !errors.HasErrors(),
		Text:  req.Text,
	}

	if errors.HasErrors() {
		response.Errors = errors.Errors

		// Build suggestions map
		response.Suggestions = make(map[string][]string)
		for _, e := range errors.Errors {
			if len(e.Suggestions) > 0 {
				response.Suggestions[e.Word] = e.Suggestions
			}
		}

		// Return 400 for validation failures
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	// Return 200 for valid text
	return c.JSON(response)
}
