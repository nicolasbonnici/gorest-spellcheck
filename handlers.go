package spellcheck

import (
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for the spellcheck endpoint
type Handler struct {
	config *Config
}

// NewHandler creates a new Handler instance
func NewHandler(config *Config) *Handler {
	return &Handler{
		config: config,
	}
}

// Check handles POST /api/spellcheck requests for on-demand spell checking
func (h *Handler) Check(c *fiber.Ctx) error {
	// TODO: Implement spell check endpoint
	return c.Status(501).JSON(fiber.Map{
		"error": "not yet implemented",
	})
}
