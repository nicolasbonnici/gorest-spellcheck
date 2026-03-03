package spellcheck

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/plugin"
)

type SpellcheckPlugin struct {
	config  Config
	handler *Handler
}

func NewPlugin() plugin.Plugin {
	return &SpellcheckPlugin{}
}

func (p *SpellcheckPlugin) Name() string {
	return "spellcheck"
}

func (p *SpellcheckPlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	// Extract config values from map
	if enabled, ok := config["enabled"].(bool); ok {
		p.config.Enabled = enabled
	}

	if defaultLang, ok := config["default_language"].(string); ok {
		p.config.DefaultLanguage = defaultLang
	}

	if maxTextLen, ok := config["max_text_length"].(int); ok {
		p.config.MaxTextLength = maxTextLen
	}

	if maxSuggestions, ok := config["max_suggestions"].(int); ok {
		p.config.MaxSuggestions = maxSuggestions
	}

	if minWordLen, ok := config["min_word_length"].(int); ok {
		p.config.MinWordLength = minWordLen
	}

	if caseSensitive, ok := config["case_sensitive"].(bool); ok {
		p.config.CaseSensitive = caseSensitive
	}

	// Validate configuration
	if err := p.config.Validate(); err != nil {
		logger.Log.Error("Invalid spellcheck plugin configuration", "error", err)
		return err
	}

	// Initialize spellchecker
	spellchecker, err := NewSpellchecker(&p.config)
	if err != nil {
		logger.Log.Error("Failed to initialize spellchecker", "error", err)
		return err
	}

	// Initialize handler
	p.handler = NewHandler(&p.config, spellchecker)

	logger.Log.Info("Spellcheck plugin initialized successfully",
		"enabled", p.config.Enabled,
		"default_language", p.config.DefaultLanguage,
		"max_text_length", p.config.MaxTextLength)
	return nil
}

func (p *SpellcheckPlugin) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}

func (p *SpellcheckPlugin) SetupEndpoints(app *fiber.App) error {
	if p.handler == nil {
		logger.Log.Warn("Spellcheck plugin handler not initialized, skipping endpoint registration")
		return nil
	}

	api := app.Group("/api/spellcheck")

	// POST /api/spellcheck - On-demand spell check
	api.Post("/", p.handler.Check)

	logger.Log.Info("Spellcheck plugin endpoints registered", "prefix", "/api/spellcheck")
	return nil
}

func (p *SpellcheckPlugin) Config() *Config {
	return &p.config
}
