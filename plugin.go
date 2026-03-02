package spellcheck

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nicolasbonnici/gorest/database"
	"github.com/nicolasbonnici/gorest/logger"
	"github.com/nicolasbonnici/gorest/plugin"
)

type SpellcheckPlugin struct {
	config  Config
	db      database.Database
	handler *Handler
	repo    Repository
}

func NewPlugin() plugin.Plugin {
	return &SpellcheckPlugin{}
}

func (p *SpellcheckPlugin) Name() string {
	return "spellcheck"
}

func (p *SpellcheckPlugin) Initialize(config map[string]interface{}) error {
	p.config = DefaultConfig()

	if db, ok := config["database"].(database.Database); ok {
		p.db = db
	}

	if enabled, ok := config["enabled"].(bool); ok {
		p.config.Enabled = enabled
	}

	if maxItems, ok := config["max_items"].(int); ok {
		p.config.MaxItems = maxItems
	}

	if err := p.config.Validate(); err != nil {
		logger.Log.Error("Invalid spellcheck plugin configuration", "error", err)
		return err
	}

	if p.db != nil {
		p.repo = NewRepository(p.db)
		p.handler = NewHandler(p.repo, &p.config)
		logger.Log.Info("Spellcheck plugin database initialized")
	} else {
		logger.Log.Warn("Spellcheck plugin initialized without database - endpoints will not be available")
	}

	logger.Log.Info("Spellcheck plugin initialized successfully", "enabled", p.config.Enabled)
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

	api.Post("/", p.handler.Create)
	api.Get("/:id", p.handler.GetByID)
	api.Get("/", p.handler.List)
	api.Put("/:id", p.handler.Update)
	api.Delete("/:id", p.handler.Delete)

	logger.Log.Info("Spellcheck plugin endpoints registered", "prefix", "/api/spellcheck")
	return nil
}

func (p *SpellcheckPlugin) Config() *Config {
	return &p.config
}
