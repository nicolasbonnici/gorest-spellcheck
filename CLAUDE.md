# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the **gorest-spellcheck** plugin, providing automatic spelling validation middleware and on-demand spell checking for GoREST applications.

**Purpose**: Validates text fields in HTTP requests for spelling errors using a built-in English dictionary with 500+ words. Provides both automatic middleware validation (using struct tags or field name heuristics) and an explicit HTTP endpoint for on-demand checking.

## Build & Test Commands

```bash
# Install development tools and git hooks (run this first)
make install

# Run all checks (lint, test, build)
make all

# Run tests with race detector
make test

# Run tests with HTML coverage report
make test-coverage

# Run linter
make lint

# Auto-fix linting issues
make lint-fix

# Run all Go Report Card quality checks
make audit

# Build verification
make build

# Clean build artifacts and caches
make clean
```

### Running Specific Tests

```bash
# Run tests for a specific file
go test -v -race ./handler_test.go

# Run a single test function
go test -v -race -run TestHandler_Create

# Run integration tests (when available)
go test -v -race -tags=integration ./...
```

## Architecture & Plugin System

### Plugin Interface Implementation

The spellcheck implements the core GoREST plugin interface in `plugin.go`:

```go
type Plugin interface {
    Name() string                                    // Returns "spellcheck"
    Initialize(config map[string]interface{}) error  // Setup and validation
    Handler() fiber.Handler                          // Middleware (if needed)
    SetupEndpoints(app *fiber.App) error            // Register HTTP routes
}
```

### Automatic Config Injection

The plugin automatically receives these values from GoREST core:
- `database` - Database connection (database.Database)
- `enabled` - Enable/disable plugin (bool)
- Plugin-specific config from `gorest.yaml` under `plugins[].config`

### Plugin Structure

```
spellcheck/
├── plugin.go              # Plugin interface implementation
├── config.go              # Configuration with validation
├── models.go              # Request/Response structs (CheckRequest, CheckResponse)
├── errors.go              # Custom error types
├── spellchecker.go        # Core spell checking logic (fuzzy matching)
├── tag_parser.go          # Struct tag parsing with caching
├── middleware.go          # Automatic validation middleware
├── handlers.go            # HTTP endpoint handler (/api/spellcheck)
├── *_test.go              # Unit tests (82.6% coverage)
├── plugin_integration_test.go  # End-to-end integration tests
└── examples/basic/        # Usage example (if needed)
```

## Key Components

### Configuration (config.go)

Defines plugin settings with comprehensive validation:

```go
type Config struct {
    Enabled            bool     // Enable/disable plugin
    DefaultLanguage    string   // Default: "en"
    SupportedLanguages []string // Initially: ["en"]
    MaxTextLength      int      // Default: 10000 (1-1,000,000)
    MaxSuggestions     int      // Default: 5 (1-20)
    IgnoredWords       []string // Custom ignored words
    CustomDictionary   string   // Path to custom dictionary
    CaseSensitive      bool     // Default: false
    MinWordLength      int      // Default: 2 (1-10)
}
```

**Important**: `config.Validate()` enforces strict ranges on all values to prevent resource exhaustion.

### Spellchecker (spellchecker.go)

Core spell checking logic using `github.com/sajari/fuzzy`:

```go
type Spellchecker struct {
    model          *fuzzy.Model
    config         *Config
    ignoredWords   map[string]bool
    caseSensitive  bool
    minWordLength  int
    maxSuggestions int
}
```

**Key Methods**:
- `Check(text string) (*SpellingErrors, error)` - Check entire text
- `CheckField(fieldName, text string) (*SpellingErrors, error)` - Check with field context
- `extractWords(text string) []wordInfo` - Extract words with positions
- `isCorrect(word string) bool` - Check if word is in dictionary
- `getSuggestions(word string) []string` - Get correction suggestions

**Dictionary**: 500+ built-in English words including common tech terms (API, JSON, HTTP, etc.)

### TagParser (tag_parser.go)

Thread-safe struct tag parser with caching:

```go
type TagParser struct {
    cache sync.Map // map[reflect.Type][]FieldInfo
}

type FieldInfo struct {
    Name          string // Go field name
    JSONName      string // JSON tag name
    SpellCheck    bool   // Has spellcheck:"true"
    IsStringField bool
}
```

**Caching Strategy**: Parses each struct type once using reflection, stores in sync.Map for concurrent access.

### Middleware (middleware.go)

Automatic validation for POST/PUT/PATCH requests:

```go
type Middleware struct {
    config       *Config
    spellchecker *Spellchecker
    tagParser    *TagParser
}
```

**Two Validation Methods**:
1. `Validate() fiber.Handler` - Middleware for automatic request validation
2. `ValidateStruct(v interface{}) (*SpellingErrors, error)` - Explicit struct validation

**Field Selection**:
- Uses `shouldCheckField()` heuristic for common field names (title, content, description, etc.)
- Only validates POST/PUT/PATCH with JSON content-type
- Skips validation for GET/DELETE/HEAD requests

### Handler (handlers.go)

HTTP endpoint for on-demand spell checking:

```go
type Handler struct {
    config       *Config
    spellchecker *Spellchecker
}
```

**Endpoint**: `POST /api/spellcheck`

**Request Model**:
```go
type CheckRequest struct {
    Text     string            `json:"text" validate:"required"`
    Language string            `json:"language,omitempty"`
    Context  []string          `json:"context,omitempty"` // Temporary ignored words
    Options  *CheckOptions     `json:"options,omitempty"`
}
```

**Response Model**:
```go
type CheckResponse struct {
    Valid       bool                       `json:"valid"`
    Errors      []*SpellingError           `json:"errors,omitempty"`
    Suggestions map[string][]string        `json:"suggestions,omitempty"`
    Text        string                     `json:"text,omitempty"`
}
```

## API Endpoint

The spellcheck plugin registers a single endpoint: `POST /api/spellcheck`

### On-Demand Spell Checking

**Request**:
```json
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

**Response (400 Bad Request)** - Text has spelling errors:
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

**Response (200 OK)** - Text is correct:
```json
{
  "valid": true,
  "errors": [],
  "text": "The quick brown fox jumps over the lazy dog"
}
```

## Middleware Behavior

### Automatic Validation

The middleware validates POST/PUT/PATCH requests with JSON bodies:

1. **Request Filtering**:
   - Only validates POST/PUT/PATCH methods
   - Only validates application/json content-type
   - Skips empty bodies
   - Passes through GET/DELETE/HEAD requests

2. **Field Selection**:
   - Checks common text field names: title, content, description, body, text, message, comment, note, summary, excerpt, caption, bio, about
   - Ignores structural fields: id, slug, email, password, username

3. **Error Response** (400 Bad Request):
```json
{
  "error": "Spelling errors found",
  "errors": [
    {
      "field": "title",
      "word": "Teh",
      "position": 0,
      "suggestions": ["The", "Tea", "Ten"]
    }
  ]
}
```

### Struct Tag Validation

Use `ValidateStruct()` for explicit validation with tags:

```go
type Article struct {
    Title   string `json:"title" spellcheck:"true"`   // ✓ Checked
    Content string `json:"content" spellcheck:"true"` // ✓ Checked
    Slug    string `json:"slug"`                      // ✗ Not checked
}

middleware.ValidateStruct(&article)
```

## Important Patterns

### Error Types

The plugin provides specialized error types:

```go
// SpellingError - Single word error with suggestions
type SpellingError struct {
    Field       string
    Word        string
    Position    int
    Suggestions []string
}

// TextTooLongError - Prevents resource exhaustion
type TextTooLongError struct {
    Length    int
    MaxLength int
}

// UnsupportedLanguageError - Language validation
type UnsupportedLanguageError struct {
    Language           string
    SupportedLanguages []string
}
```

### Thread-Safe Caching

TagParser uses sync.Map for concurrent access:

```go
// Parse struct once, cache forever
func (p *TagParser) Parse(v interface{}) []FieldInfo {
    typ := reflect.TypeOf(v)
    if typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }

    // Check cache first
    if cached, ok := p.cache.Load(typ); ok {
        return cached.([]FieldInfo)
    }

    // Parse and cache
    fieldInfos := p.parseStruct(typ)
    p.cache.Store(typ, fieldInfos)
    return fieldInfos
}
```

### Fuzzy Matching Configuration

Spellchecker uses configurable Levenshtein distance:

```go
model := fuzzy.NewModel()
model.SetDepth(4)      // Max 4 character differences
model.SetThreshold(1)  // Suggestion threshold
```

### Dictionary Loading

Batch training for optimal performance:

```go
// Train with all words at once (more efficient)
model.Train(commonWords)

// Also train with capitalized versions
for _, word := range commonWords {
    if len(word) > 0 {
        capitalized := strings.ToUpper(word[:1]) + word[1:]
        model.Train([]string{capitalized})
    }
}
```

### Logging

Use structured logging for spell check operations:

```go
logger.Log.Info("Spellcheck plugin initialized successfully",
    "enabled", config.Enabled,
    "default_language", config.DefaultLanguage,
    "max_text_length", config.MaxTextLength)

logger.Log.Error("Spell check failed", "field", fieldName, "error", err)
```

### Performance Optimizations

1. **Struct parsing**: Cached by type using sync.Map
2. **Dictionary**: Loaded once at initialization
3. **Request filtering**: Early exit for non-JSON, GET requests
4. **Word extraction**: Single pass with position tracking
5. **Ignored words**: Map lookup O(1) instead of slice O(n)

## Testing Strategy

### Unit Tests

Comprehensive test coverage (82.6%):

- `config_test.go` - Configuration validation edge cases
- `errors_test.go` - Custom error types
- `spellchecker_test.go` - Core spelling logic, dictionary loading
- `tag_parser_test.go` - Struct parsing, caching, concurrent access
- `models_test.go` - Request/response validation
- `handlers_test.go` - HTTP endpoint behavior
- `middleware_test.go` - Automatic validation, field selection
- `plugin_test.go` - Plugin initialization and configuration
- `plugin_integration_test.go` - End-to-end integration scenarios

### Running Tests

```bash
# All tests
make test

# With coverage report
make test-coverage

# Specific test
go test -v -run TestSpellchecker_Check

# Integration tests
go test -v -run TestPluginIntegration
```

## Pre-commit Hooks

Git pre-commit hooks are installed via:
```bash
make install
# or manually:
./.githooks/install.sh
```

The hook runs `make lint && make test` before each commit to ensure code quality.

**Skip hooks** (use sparingly): `git commit --no-verify`

## Related Files

- **Main GoREST framework**: `/home/nicolas/Projects/go/gorest/`
- **Other plugins**: `/home/nicolas/Projects/go/gorest-*/`
- **Plugin docs**: See `/home/nicolas/Projects/go/gorest/PLUGINS.md`

## Common Issues

### Common words flagged as errors

The built-in dictionary contains 500+ common words. If you encounter false positives:
- Add words to `ignored_words` configuration
- Use a custom dictionary file with domain-specific terms
- Check if the word is a technical term or brand name

### Middleware not validating requests

Ensure:
1. Plugin is enabled in configuration (`enabled: true`)
2. Request Content-Type is `application/json`
3. Request method is POST, PUT, or PATCH (GET/DELETE are skipped)
4. Field names match the heuristic or struct tags are used

### Performance with large texts

Set `max_text_length` appropriately:
- Default: 10,000 characters
- For long documents: Increase to 50,000-100,000
- For short messages: Decrease to 1,000-5,000

### Suggestions not accurate

The fuzzy library may not always return perfect suggestions. This is acceptable as long as:
- Misspelled words are detected correctly
- Some suggestions are provided (even if not perfect)
- The primary use case (blocking misspelled content) works
