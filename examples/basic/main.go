package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	spellcheck "github.com/nicolasbonnici/gorest-spellcheck"
)

// Article demonstrates a model with spellcheck tags
type Article struct {
	ID        string    `json:"id"`
	Title     string    `json:"title" spellcheck:"true"`
	Content   string    `json:"content" spellcheck:"true"`
	Author    string    `json:"author"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateArticleRequest is the input DTO
type CreateArticleRequest struct {
	Title   string   `json:"title" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Author  string   `json:"author" validate:"required"`
	Tags    []string `json:"tags"`
}

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Spellcheck Example",
	})

	// Initialize spellcheck plugin
	plugin := spellcheck.NewPlugin()
	config := map[string]interface{}{
		"enabled":          true,
		"default_language": "en",
		"max_text_length":  10000,
		"max_suggestions":  5,
		"ignored_words":    []string{"API", "JSON", "HTTP", "UUID"},
	}

	if err := plugin.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize spellcheck plugin: %v", err)
	}

	// Apply middleware globally (will validate all POST/PUT/PATCH requests)
	app.Use(plugin.Handler())

	// Setup spellcheck API endpoint
	spellcheckPlugin := plugin.(*spellcheck.SpellcheckPlugin)
	if err := spellcheckPlugin.SetupEndpoints(app); err != nil {
		log.Fatalf("Failed to setup endpoints: %v", err)
	}

	// Create articles endpoint
	app.Post("/api/articles", createArticle)
	app.Get("/api/articles/:id", getArticle)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"time":   time.Now(),
		})
	})

	log.Println("Starting server on http://localhost:3000")
	log.Println("Try these requests:")
	log.Println("")
	log.Println("1. POST /api/articles with correct spelling:")
	log.Println("   curl -X POST http://localhost:3000/api/articles \\")
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{\"title\":\"Hello World\",\"content\":\"This is a test article\",\"author\":\"John\"}'")
	log.Println("")
	log.Println("2. POST /api/articles with spelling errors:")
	log.Println("   curl -X POST http://localhost:3000/api/articles \\")
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{\"title\":\"Teh quik test\",\"content\":\"This has speling erors\",\"author\":\"John\"}'")
	log.Println("")
	log.Println("3. POST /api/spellcheck for on-demand checking:")
	log.Println("   curl -X POST http://localhost:3000/api/spellcheck \\")
	log.Println("     -H 'Content-Type: application/json' \\")
	log.Println("     -d '{\"text\":\"Teh quik brown fox\"}'")
	log.Println("")

	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func createArticle(c *fiber.Ctx) error {
	var req CreateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Note: The spellcheck middleware has already validated spelling
	// in the request body before reaching this handler

	// Create article (simplified - no database)
	article := Article{
		ID:        "123",
		Title:     req.Title,
		Content:   req.Content,
		Author:    req.Author,
		Tags:      req.Tags,
		CreatedAt: time.Now(),
	}

	return c.Status(201).JSON(article)
}

func getArticle(c *fiber.Ctx) error {
	id := c.Params("id")

	// Simplified - return mock data
	article := Article{
		ID:        id,
		Title:     "Sample Article",
		Content:   "This is sample content",
		Author:    "John Doe",
		Tags:      []string{"sample", "test"},
		CreatedAt: time.Now(),
	}

	return c.JSON(article)
}
