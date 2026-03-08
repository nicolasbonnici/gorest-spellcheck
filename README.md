# GoREST Spellcheck Plugin

[![CI](https://github.com/nicolasbonnici/gorest-spellcheck/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-spellcheck/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-spellcheck)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-spellcheck)
[![Coverage](https://img.shields.io/badge/coverage-82.6%25-brightgreen)](https://github.com/nicolasbonnici/gorest-spellcheck)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Automatic spelling validation middleware and on-demand spell checking for GoREST applications.

## Features

- **Automatic Middleware Validation** - Validates request bodies for spelling errors before they reach your handlers
- **On-Demand Spell Checking** - HTTP endpoint for explicit spell checking with detailed error reports
- **Struct Tag Annotations** - Use `spellcheck:"true"` tags to mark fields for validation
- **Pure Go Implementation** - Uses `github.com/sajari/fuzzy` library (no external dependencies)
- **Configurable** - Customize ignored words, max text length, suggestions, and more
- **Built-in Dictionary** - 500+ common English words including tech terms (API, JSON, HTTP, etc.)
- **Custom Dictionaries** - Load your own dictionary files
- **Thread-Safe** - Caches struct metadata for optimal performance
- **Comprehensive Testing** - 82.6% test coverage with integration tests

## Installation

```bash
go get github.com/nicolasbonnici/gorest-spellcheck
```

## Quick Start

### With GoREST Framework

```go
package main

import (
	"github.com/nicolasbonnici/gorest"
	"github.com/nicolasbonnici/gorest/pluginloader"
	spellcheck "github.com/nicolasbonnici/gorest-spellcheck"
)

func init() {
	pluginloader.RegisterPluginFactory("spellcheck", spellcheck.NewPlugin)
}

func main() {
	gorest.Start(gorest.Config{
		ConfigPath: ".",
	})
}
```

### Standalone Usage

```go
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	spellcheck "github.com/nicolasbonnici/gorest-spellcheck"
)

func main() {
	app := fiber.New()

	// Initialize plugin
	plugin := spellcheck.NewPlugin()
	config := map[string]interface{}{
		"enabled":          true,
		"default_language": "en",
		"max_text_length":  10000,
	}

	if err := plugin.Initialize(config); err != nil {
		log.Fatal(err)
	}

	// Apply middleware
	app.Use(plugin.Handler())

	// Setup spellcheck endpoint
	spellcheckPlugin := plugin.(*spellcheck.SpellcheckPlugin)
	spellcheckPlugin.SetupEndpoints(app)

	// Your routes here
	app.Post("/articles", createArticle)

	app.Listen(":3000")
}
```

## Configuration

Add to your `gorest.yaml`:

```yaml
plugins:
  - name: spellcheck
    enabled: true
    config:
      default_language: "en"
      max_text_length: 10000      # Maximum text length to check
      max_suggestions: 5           # Max suggestions per error
      min_word_length: 2           # Minimum word length to check
      case_sensitive: false        # Case-sensitive checking
      ignored_words:               # Custom words to ignore
        - "API"
        - "JSON"
        - "HTTP"
        - "UUID"
      custom_dictionary: "/path/to/dictionary.txt"  # Optional custom dictionary
```

### Configuration Options

| Option | Type | Default | Range | Description |
|--------|------|---------|-------|-------------|
| `enabled` | bool | `true` | - | Enable/disable the plugin |
| `default_language` | string | `"en"` | - | Default language for spell checking |
| `supported_languages` | []string | `["en"]` | - | List of supported languages |
| `max_text_length` | int | `10000` | 1-1,000,000 | Maximum text length to validate |
| `max_suggestions` | int | `5` | 1-20 | Maximum suggestions per error |
| `min_word_length` | int | `2` | 1-10 | Minimum word length to check |
| `case_sensitive` | bool | `false` | - | Enable case-sensitive checking |
| `ignored_words` | []string | `[]` | - | Words to always treat as correct |
| `custom_dictionary` | string | `""` | - | Path to custom dictionary file |

## Usage

### 1. Automatic Middleware Validation

The middleware automatically validates POST/PUT/PATCH requests with JSON bodies:

```go
type Article struct {
	Title   string `json:"title" spellcheck:"true"`
	Content string `json:"content" spellcheck:"true"`
	Author  string `json:"author"` // Not checked
}
```

**Request with spelling errors:**

```bash
POST /api/articles
Content-Type: application/json

{
  "title": "Teh quik brown fox",
  "content": "This has speling erors"
}
```

**Response (400 Bad Request):**

```json
{
  "error": "Spelling errors found",
  "errors": [
    {
      "field": "title",
      "word": "Teh",
      "position": 0,
      "suggestions": ["The", "Tea", "Ten"]
    },
    {
      "field": "title",
      "word": "quik",
      "position": 4,
      "suggestions": ["quick", "quit", "quiz"]
    },
    {
      "field": "content",
      "word": "speling",
      "position": 9,
      "suggestions": ["spelling", "spieling"]
    },
    {
      "field": "content",
      "word": "erors",
      "position": 17,
      "suggestions": ["errors", "eros"]
    }
  ]
}
```

### 2. On-Demand Spell Checking

Use the `/api/spellcheck` endpoint for explicit validation:

```bash
POST /api/spellcheck
Content-Type: application/json

{
  "text": "Teh quik brown fox jumps over the lazy dog",
  "language": "en",
  "context": ["API", "JSON"],
  "options": {
    "case_sensitive": false,
    "max_suggestions": 3
  }
}
```

**Response (400 Bad Request):**

```json
{
  "valid": false,
  "errors": [
    {
      "word": "Teh",
      "position": 0,
      "suggestions": ["The", "Tea", "Ten"]
    },
    {
      "word": "quik",
      "position": 4,
      "suggestions": ["quick", "quit", "quiz"]
    }
  ],
  "suggestions": {
    "Teh": ["The", "Tea", "Ten"],
    "quik": ["quick", "quit", "quiz"]
  },
  "text": "Teh quik brown fox jumps over the lazy dog"
}
```

**Response with correct spelling (200 OK):**

```json
{
  "valid": true,
  "errors": [],
  "text": "The quick brown fox jumps over the lazy dog"
}
```

### 3. Manual Struct Validation

```go
middleware, _ := spellcheck.NewMiddleware(&config, spellchecker)

article := &Article{
	Title:   "Teh test",
	Content: "This has erors",
}

errors, err := middleware.ValidateStruct(article)
if err != nil {
	log.Fatal(err)
}

if errors.HasErrors() {
	for _, e := range errors.Errors {
		fmt.Printf("Field %s: '%s' at position %d, suggestions: %v\n",
			e.Field, e.Word, e.Position, e.Suggestions)
	}
}
```

## Field Checking Behavior

### With Struct Tags

When using `ValidateStruct()`, only fields marked with `spellcheck:"true"` are validated:

```go
type Article struct {
	Title   string `json:"title" spellcheck:"true"`   // ✓ Checked
	Content string `json:"content" spellcheck:"true"` // ✓ Checked
	Slug    string `json:"slug"`                      // ✗ Not checked
}
```

### With Middleware (JSON Requests)

The middleware uses a heuristic based on common field names:

**Checked fields:** `title`, `content`, `description`, `body`, `text`, `message`, `comment`, `note`, `summary`, `excerpt`, `caption`, `bio`, `about`

**Not checked:** `id`, `slug`, `email`, `password`, `username`, etc.

```json
{
  "title": "Test article",    // ✓ Checked
  "content": "Body text",      // ✓ Checked
  "slug": "test-article",      // ✗ Not checked
  "email": "user@example.com"  // ✗ Not checked
}
```

## Custom Dictionary

Create a text file with one word per line:

```text
# Custom dictionary
customword
techterm
brandname
```

Configure in `gorest.yaml`:

```yaml
plugins:
  - name: spellcheck
    config:
      custom_dictionary: "./custom-words.txt"
```

## Error Types

The plugin provides detailed error types:

```go
// SpellingError - Single word error
type SpellingError struct {
	Field       string   // Field name (empty for on-demand checks)
	Word        string   // Misspelled word
	Position    int      // Character position in text
	Suggestions []string // Correction suggestions
}

// SpellingErrors - Collection of errors
type SpellingErrors struct {
	Errors []*SpellingError
}

// TextTooLongError - Text exceeds max length
type TextTooLongError struct {
	Length    int
	MaxLength int
}

// UnsupportedLanguageError - Language not supported
type UnsupportedLanguageError struct {
	Language           string
	SupportedLanguages []string
}

// ValidationError - Request validation failed
type ValidationError struct {
	Message string
	Fields  map[string]string
}
```

## Development

### Prerequisites

- Go 1.23+ or later
- Make (optional)

### Setup

```bash
# Clone and install dependencies
git clone https://github.com/nicolasbonnici/gorest-spellcheck.git
cd gorest-spellcheck
make install  # Installs golangci-lint and git hooks

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint
```

### Available Make Targets

```bash
make install        # Install development tools and git hooks
make test           # Run tests with race detector
make test-coverage  # Generate HTML coverage report
make lint           # Run linter
make lint-fix       # Run linter with auto-fix
make build          # Verify build
make audit          # Run all quality checks
make clean          # Clean artifacts
make all            # Run lint, test, and build
```

### Running Tests

```bash
# All tests
go test -v ./...

# Specific test
go test -v -run TestSpellchecker_Check

# With coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out

# Integration tests
go test -v -run TestPluginIntegration
```

## Architecture

```
gorest-spellcheck/
├── plugin.go              # Plugin interface implementation
├── config.go              # Configuration with validation
├── models.go              # Request/Response structs
├── errors.go              # Custom error types
├── spellchecker.go        # Core spell checking logic
├── tag_parser.go          # Struct tag parsing with caching
├── spellcheck_middleware.go # Automatic validation middleware
├── handlers.go            # HTTP endpoint handler
├── *_test.go              # Unit tests (82.6% coverage)
└── examples/              # Usage examples
```

### Key Components

- **Spellchecker** - Core logic using `github.com/sajari/fuzzy`
- **TagParser** - Thread-safe reflection-based tag parsing with caching
- **Middleware** - Fiber middleware for automatic validation
- **Handler** - HTTP endpoint for on-demand checking
- **Plugin** - GoREST plugin interface implementation

## Performance Considerations

- **Tag Parsing Caching** - Struct metadata is parsed once and cached per type
- **Concurrent Access** - Uses `sync.Map` for thread-safe caching
- **Dictionary Loading** - Dictionary loaded once at initialization
- **Request Filtering** - Only validates POST/PUT/PATCH with JSON content
- **Fast Word Matching** - Uses Levenshtein distance with configurable depth

## Security

- **Text Length Limits** - Prevents resource exhaustion with max_text_length
- **Input Sanitization** - Skips numbers and special characters
- **No Sensitive Data** - Never logs or exposes checked text content
- **Configurable Validation** - Can be disabled per environment

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass and coverage remains high
5. Run linter (`make lint`)
6. Submit a pull request

## Troubleshooting

### Common words flagged as errors

The built-in dictionary contains 500+ common words. If you encounter false positives, add words to `ignored_words` or use a custom dictionary.

### Performance issues with large texts

Set `max_text_length` appropriately for your use case. The default 10,000 characters is suitable for most applications.

### Middleware not validating

Ensure:
- Plugin is enabled in configuration
- Content-Type is `application/json`
- Request method is POST, PUT, or PATCH
- Field names match the heuristic or struct tags are used

## License

MIT License - See [LICENSE](LICENSE) file for details.

## Related Projects

- [GoREST](https://github.com/nicolasbonnici/gorest) - The GoREST framework
- [GoREST Auth Plugin](https://github.com/nicolasbonnici/gorest-auth) - JWT authentication
- [GoREST RBAC Plugin](https://github.com/nicolasbonnici/gorest-rbac) - Role-based access control
- [GoREST Blog Plugin](https://github.com/nicolasbonnici/gorest-blog) - Blog functionality

## Support

- GitHub Issues: https://github.com/nicolasbonnici/gorest-spellcheck/issues
- GoREST Documentation: https://github.com/nicolasbonnici/gorest

## Changelog

### v0.1.0 (2026-03-03)

- Initial release
- Automatic middleware validation
- On-demand spell checking endpoint
- Struct tag annotations
- Built-in English dictionary (500+ words)
- Custom dictionary support
- Thread-safe tag parsing with caching
- 82.6% test coverage
- Comprehensive integration tests
