# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

This is the **gorest-spellcheck** plugin, a comprehensive template/boilerplate for creating new GoREST plugins. It demonstrates all common patterns, structures, and best practices used across the GoREST plugin ecosystem.

**Purpose**: Clone this repository as a starting point when creating new plugins. It includes complete CRUD operations, database integration, configuration management, testing, and CI/CD.

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
├── models.go              # Data models (database structs)
├── dtos.go                # Request/Response structs
├── handlers.go            # HTTP request handlers
├── repository.go          # Database operations layer
├── *_test.go              # Unit tests
├── migrations/            # Database migrations
└── examples/basic/        # Usage example
```

## Key Components

### Configuration (config.go)

Defines plugin settings with validation:

```go
type Config struct {
    Enabled  bool  // Enable/disable plugin
    MaxItems int   // Max items per request (1-1000)
}
```

**Important**: Always call `config.Validate()` in `Initialize()` to catch invalid configuration early.

### Repository Pattern (repository.go)

Encapsulates all database operations:

```go
type Repository interface {
    Create(ctx context.Context, item *models.SpellcheckItem) error
    GetByID(ctx context.Context, id string) (*models.SpellcheckItem, error)
    List(ctx context.Context, limit, offset int) ([]models.SpellcheckItem, int, error)
    Update(ctx context.Context, id string, item *models.SpellcheckItem) error
    Delete(ctx context.Context, id string) error
}
```

**Why Repository Pattern?**
- Separates business logic from database operations
- Makes testing easier (can mock repository)
- Consistent interface across different storage backends
- Simplifies handler code

### Handler Layer (handlers.go)

Handles HTTP requests and responses:

```go
type Handler struct {
    repo   Repository
    config *Config
}
```

**Handler Responsibilities**:
1. Parse and validate request bodies (DTOs)
2. Convert DTOs to models
3. Call repository methods
4. Return standardized responses

### Models vs DTOs

**Models** (`models.go`): Database representations with `db` tags
```go
type SpellcheckItem struct {
    Id          string     `json:"id" db:"id"`
    Name        string     `json:"name" db:"name"`
    Description string     `json:"description" db:"description"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}
```

**DTOs** (`dtos.go`): Request/Response formats
```go
type CreateSpellcheckItemRequest struct {
    Name        string  `json:"name" validate:"required,min=1,max=255"`
    Description string  `json:"description"`
}
```

**Key Difference**: DTOs exclude system-managed fields (id, created_at, updated_at) and include validation tags.

## Database Migrations

Located in `migrations/` directory with format: `YYYYMMDDHHMMSSMMM_description.sql`

```sql
-- migrations/001_create_spellcheck_items.sql
CREATE TABLE spellcheck_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);
```

**Migration Best Practices**:
- Use timestamps in filename for ordering
- Include both UP and DOWN migrations
- Test migrations on all supported databases (PostgreSQL, MySQL, SQLite)
- Never modify existing migrations (create new ones instead)

## API Endpoints

The spellcheck registers these routes under `/api/spellcheck`:

```
POST   /api/spellcheck       - Create new item
GET    /api/spellcheck/:id   - Get item by ID
GET    /api/spellcheck       - List all items (paginated)
PUT    /api/spellcheck/:id   - Update item
DELETE /api/spellcheck/:id   - Delete item
```

### Request/Response Examples

**Create Item**:
```json
POST /api/spellcheck
{
  "name": "Example Item",
  "description": "This is an example"
}

Response (201):
{
  "id": "uuid-here",
  "name": "Example Item",
  "description": "This is an example",
  "created_at": "2025-01-01T00:00:00Z"
}
```

**List Items** (with pagination):
```json
GET /api/spellcheck?limit=10&offset=0

Response (200):
{
  "items": [...],
  "total": 42,
  "limit": 10,
  "offset": 0
}
```

## Important Patterns

### Error Handling

Always return consistent error responses:

```go
if err != nil {
    return c.Status(500).JSON(fiber.Map{
        "error": "Failed to create item",
        "details": err.Error(),
    })
}
```

### Validation

Use validator tags in DTOs and validate before processing:

```go
if err := req.Validate(); err != nil {
    return c.Status(400).JSON(fiber.Map{
        "error": "Validation failed",
        "details": err.Error(),
    })
}
```

### Logging

Use structured logging throughout:

```go
logger.Log.Info("Creating spellcheck item", "name", req.Name)
logger.Log.Error("Failed to create item", "error", err, "name", req.Name)
```

### Context Propagation

Always pass context down the call chain for proper timeout/cancellation handling:

```go
func (h *Handler) Create(c *fiber.Ctx) error {
    ctx := c.Context()
    // Pass ctx to repository methods
    err := h.repo.Create(ctx, item)
}
```

## Customizing This Template

When creating a new plugin from this spellcheck:

1. **Clone and rename**: `cp -r gorest-spellcheck gorest-yourplugin`
2. **Global find/replace**: `spellcheck` → `yourplugin`, `Spellcheck` → `YourPlugin`
3. **Update go.mod**: Module name, dependencies
4. **Customize models**: Update `models.go` and `dtos.go` for your domain
5. **Update migrations**: Create appropriate database schema
6. **Modify handlers**: Add plugin-specific business logic
7. **Update docs**: README.md, CLAUDE.md, configuration examples
8. **Test thoroughly**: Add tests for all custom logic

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

### "Database not initialized" warnings

The plugin can run without a database for testing middleware-only features. If you see this warning but need database features, ensure `gorest.yaml` includes database configuration.

### Migration conflicts

If migrations fail, check:
1. Database connection is valid
2. Migration version doesn't conflict with existing versions
3. SQL syntax is compatible with target database (PostgreSQL/MySQL/SQLite)
