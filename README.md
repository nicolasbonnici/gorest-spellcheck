# GoREST Spellcheck Plugin

[![CI](https://github.com/nicolasbonnici/gorest-spellcheck/actions/workflows/ci.yml/badge.svg)](https://github.com/nicolasbonnici/gorest-spellcheck/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nicolasbonnici/gorest-spellcheck)](https://goreportcard.com/report/github.com/nicolasbonnici/gorest-spellcheck)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**A comprehensive boilerplate/spellcheck plugin for the GoREST framework - your starting point for building new plugins.**

This plugin serves as a reference implementation and template for creating GoREST plugins. It includes all the common patterns, structures, and best practices used across the GoREST plugin ecosystem.

## Purpose

This spellcheck plugin is designed to:

1. **Provide a starting template** - Clone and customize this to create your own plugin
2. **Demonstrate best practices** - Shows proper structure, patterns, and conventions
3. **Serve as documentation** - Complete example of all plugin features
4. **Ensure consistency** - Maintains the same patterns across all GoREST plugins

## Features

This spellcheck includes examples of:

- **Complete CRUD operations** - Create, Read, Update, Delete handlers
- **Database integration** - Repository pattern with database operations
- **Configuration management** - Config struct with validation
- **Request/Response models** - Proper data structures and validation
- **Error handling** - Consistent error responses
- **Logging** - Structured logging throughout
- **User authentication** - Integration with auth plugin patterns
- **Database migrations** - SQL migration files
- **Testing** - Unit test examples
- **CI/CD** - GitHub Actions workflow
- **Documentation** - Complete README and development guide

## Quick Start

### For Using This Plugin

```bash
go get github.com/nicolasbonnici/gorest-spellcheck
```

```go
package main

import (
	"github.com/nicolasbonnici/gorest"
	"github.com/nicolasbonnici/gorest/pluginloader"

	spellcheckplugin "github.com/nicolasbonnici/gorest-spellcheck"
)

func init() {
	pluginloader.RegisterPluginFactory("spellcheck", spellcheckplugin.NewPlugin)
}

func main() {
	cfg := gorest.Config{
		ConfigPath: ".",
	}

	gorest.Start(cfg)
}
```

### For Creating Your Own Plugin

See [DEVELOPMENT.md](DEVELOPMENT.md) for detailed instructions on forking and customizing this spellcheck.

**Quick steps:**

1. Fork/clone this repository
2. Replace "spellcheck" with your plugin name throughout
3. Customize the data model, handlers, and business logic
4. Update documentation
5. Run tests and push!

## Development Environment

To set up your development environment:

```bash
make install
```

This will:
- Install Go dependencies
- Install development tools (golangci-lint)
- Set up git hooks (pre-commit linting and tests)

## Configuration

Add to your `gorest.yaml`:

```yaml
database:
  url: "${DATABASE_URL}"

plugins:
  - name: spellcheck
    enabled: true
    config:
      max_items: 100  # Maximum items per query (default: 100)
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable/disable the plugin |
| `max_items` | int | `100` | Maximum number of items returned per list query (1-1000) |

## Database Setup

Run the migration to create the required table:

```bash
# Automatic migration (if migrations.auto_migrate: true in config)
go run main.go

# Or manually
gorest migrate up
```

The plugin creates a `spellcheck_items` table with:

```sql
CREATE TABLE spellcheck_items (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    user_id UUID NOT NULL,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ
);
```

## API Endpoints

All endpoints require authentication (JWT token in `Authorization` header).

### Create Item

```bash
POST /api/spellcheck
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "My Item",
  "description": "Item description",
  "active": true
}
```

**Response (201 Created):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Item",
  "description": "Item description",
  "user_id": "750e8400-e29b-41d4-a716-446655440000",
  "active": true,
  "created_at": "2025-12-31T10:00:00Z"
}
```

### Get Item by ID

```bash
GET /api/spellcheck/:id
Authorization: Bearer <token>
```

**Response (200 OK):**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "My Item",
  "description": "Item description",
  "user_id": "750e8400-e29b-41d4-a716-446655440000",
  "active": true,
  "created_at": "2025-12-31T10:00:00Z",
  "updated_at": "2025-12-31T11:00:00Z"
}
```

### List Items

```bash
GET /api/spellcheck?limit=20&offset=0
Authorization: Bearer <token>
```

**Query Parameters:**

- `limit` - Number of items to return (default: 20, max: configured max_items)
- `offset` - Number of items to skip (default: 0)

**Response (200 OK):**

```json
{
  "items": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "My Item",
      "description": "Item description",
      "user_id": "750e8400-e29b-41d4-a716-446655440000",
      "active": true,
      "created_at": "2025-12-31T10:00:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

### Update Item

```bash
PUT /api/spellcheck/:id
Content-Type: application/json
Authorization: Bearer <token>

{
  "name": "Updated Name",
  "description": "Updated description",
  "active": false
}
```

All fields are optional. Only provided fields will be updated.

**Response (200 OK):** Returns the updated item.

### Delete Item

```bash
DELETE /api/spellcheck/:id
Authorization: Bearer <token>
```

**Response (204 No Content)**

## Project Structure

```
gorest-spellcheck/
├── plugin.go              # Main plugin implementation
├── config.go              # Configuration structure and validation
├── handlers.go            # HTTP request handlers (CRUD operations)
├── models.go              # Data models and request/response structures
├── repository.go          # Database operations (repository pattern)
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── README.md              # This file
├── DEVELOPMENT.md         # Guide for creating plugins from this spellcheck
├── LICENSE                # MIT License
├── Makefile               # Development commands
├── .gitignore             # Git ignore patterns
├── .github/
│   └── workflows/
│       └── ci.yml         # GitHub Actions CI/CD pipeline
├── migrations/
│   └── 001_create_spellcheck_items.sql  # Database migration
├── examples/
│   └── basic/
│       ├── main.go        # Example application
│       ├── go.mod         # Example dependencies
│       ├── gorest.yaml    # Example configuration
│       └── README.md      # Example documentation
└── tests/
    ├── plugin_test.go     # Plugin tests
    ├── handlers_test.go   # Handler tests
    ├── config_test.go     # Config tests
    └── repository_test.go # Repository tests
```

## Architecture Patterns

This spellcheck demonstrates the following architectural patterns used across GoREST plugins:

### 1. Plugin Interface Implementation

```go
type SpellcheckPlugin struct {
    config  Config
    db      database.Database
    handler *Handler
    repo    Repository
}

func NewPlugin() plugin.Plugin {
    return &SpellcheckPlugin{}
}

func (p *SpellcheckPlugin) Name() string
func (p *SpellcheckPlugin) Initialize(config map[string]interface{}) error
func (p *SpellcheckPlugin) Handler() fiber.Handler
func (p *SpellcheckPlugin) SetupEndpoints(app *fiber.App) error
```

### 2. Repository Pattern

Separates database logic from business logic:

```go
type Repository interface {
    Create(ctx context.Context, item *Item) error
    GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
    List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Item, int, error)
    Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates map[string]interface{}) error
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}
```

### 3. Handler Pattern

HTTP handlers with proper error handling and validation:

```go
type Handler struct {
    repo   Repository
    config *Config
}

func (h *Handler) Create(c *fiber.Ctx) error
func (h *Handler) GetByID(c *fiber.Ctx) error
func (h *Handler) List(c *fiber.Ctx) error
func (h *Handler) Update(c *fiber.Ctx) error
func (h *Handler) Delete(c *fiber.Ctx) error
```

### 4. Configuration Management

Type-safe configuration with validation:

```go
type Config struct {
    Database database.Database
    Enabled  bool
    MaxItems int
}

func (c *Config) Validate() error
func DefaultConfig() Config
```

## Security Features

This spellcheck includes examples of:

- **Authentication checks** - Validates user from JWT token
- **User ownership** - Users can only access their own items
- **Input validation** - Validates all request inputs
- **SQL injection prevention** - Uses parameterized queries
- **Error message safety** - Doesn't leak sensitive information

## Error Handling

Standard HTTP status codes:

| Status | Meaning |
|--------|---------|
| `200` | Success (GET, PUT) |
| `201` | Created (POST) |
| `204` | No Content (DELETE) |
| `400` | Bad Request (validation error) |
| `401` | Unauthorized (missing/invalid auth) |
| `404` | Not Found |
| `500` | Internal Server Error |

**Error Response Format:**

```json
{
  "error": "Error message here"
}
```

## Development

### Prerequisites

- Go 1.25.1 or later
- PostgreSQL (or compatible database)
- Make (optional, for using Makefile)

### Setup

```bash
# Clone the repository
git clone https://github.com/nicolasbonnici/gorest-spellcheck.git
cd gorest-spellcheck

# Install dependencies
go mod download

# Install development tools
make install

# Run tests
make test

# Run linter
make lint
```

### Available Make Targets

```bash
make help       # Show all available targets
make install    # Install development tools (golangci-lint)
make test       # Run tests with coverage
make lint       # Run linter
make lint-fix   # Run linter with auto-fix
make build      # Build verification
make clean      # Clean build artifacts and caches
```

## Testing

### Unit Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# View coverage report
go tool cover -html=coverage.out

# Or use Make
make test
make test-coverage  # Generates HTML coverage report
```

### Integration Tests

Integration tests run against real databases and are tagged with `//go:build integration`.

**Prerequisites:**
- PostgreSQL or MySQL running locally
- Test database created
- Environment variables set:
  - `POSTGRES_URL` (e.g., `postgres://test:test@localhost:5432/test?sslmode=disable`)
  - `MYSQL_URL` (e.g., `test:test@tcp(localhost:3306)/test`)

**Running integration tests:**

```bash
# Run integration tests only
go test -v -race -tags=integration ./...

# Run with specific database
POSTGRES_URL="postgres://test:test@localhost:5432/test?sslmode=disable" \
  go test -v -race -tags=integration ./...
```

**CI/CD:**

Integration tests run automatically in GitHub Actions using PostgreSQL and MySQL service containers.

## Integration with Other Plugins

This plugin works seamlessly with other GoREST plugins, particularly:

### With Auth Plugin

```go
import (
    authplugin "github.com/nicolasbonnici/gorest-auth"
    spellcheckplugin "github.com/nicolasbonnici/gorest-spellcheck"
)

func init() {
    pluginloader.RegisterPluginFactory("auth", authplugin.NewPlugin)
    pluginloader.RegisterPluginFactory("spellcheck", spellcheckplugin.NewPlugin)
}
```

The spellcheck plugin automatically uses user authentication from the auth plugin.

## Examples

See the [examples/basic](examples/basic) directory for a complete working example.

```bash
cd examples/basic
cp .env.example .env
# Edit .env with your database URL and JWT secret
go run main.go
```

## Creating Your Own Plugin

This spellcheck is designed to be forked and customized. See [DEVELOPMENT.md](DEVELOPMENT.md) for a complete guide on:

1. Forking and renaming the plugin
2. Customizing the data model
3. Implementing your business logic
4. Adding custom endpoints
5. Writing tests
6. Publishing your plugin

## Contributing

Contributions to improve this spellcheck are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Ensure CI passes
5. Submit a pull request

## Common Use Cases

This spellcheck can be adapted for various plugin types:

- **Resource plugins** - Manage custom data entities (posts, products, etc.)
- **Service plugins** - Provide specific services (email, notifications, etc.)
- **Integration plugins** - Connect to external APIs (payment, social media, etc.)
- **Utility plugins** - Add utility features (search, analytics, etc.)

## FAQ

**Q: Can I use this spellcheck for commercial projects?**
A: Yes! It's MIT licensed.

**Q: Do I need to keep the spellcheck branding?**
A: No, customize everything for your plugin.

**Q: Is database support required?**
A: No, you can remove database functionality if not needed.

**Q: Can I add more endpoints?**
A: Absolutely! Modify `SetupEndpoints` and add handlers as needed.

**Q: How do I handle migrations?**
A: Place SQL files in `migrations/` - GoREST handles the rest.

## Troubleshooting

### "user not authenticated" errors

Ensure the auth plugin is registered and configured properly. The spellcheck plugin requires authenticated users for most operations.

### Migration errors

Check that your database URL is correct and migrations are enabled in your config.

### Import path errors

After renaming the plugin, make sure to update all import paths and run `go mod tidy`.

---

## Git Hooks

This directory contains git hooks for the GoREST plugin to maintain code quality.

### Available Hooks

#### pre-commit

Runs before each commit to ensure code quality:
- **Linting**: Runs `make lint` to check code style and potential issues
- **Tests**: Runs `make test` to verify all tests pass

### Installation

#### Automatic Installation

Run the install script from the project root:

```bash
./.githooks/install.sh
```

#### Manual Installation

Copy the hooks to your `.git/hooks` directory:

```bash
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---


## License

MIT License - See [LICENSE](LICENSE) file for details.

## Related Projects

- [GoREST](https://github.com/nicolasbonnici/gorest) - The main GoREST framework
- [GoREST Auth Plugin](https://github.com/nicolasbonnici/gorest-auth) - JWT authentication
- [GoREST Status Plugin](https://github.com/nicolasbonnici/gorest-status) - Health checks
- [GoREST Blog Plugin](https://github.com/nicolasbonnici/gorest-blog) - Blog functionality

## Support

For questions, issues, or contributions:

- GitHub Issues: https://github.com/nicolasbonnici/gorest-spellcheck/issues
- GoREST Documentation: https://github.com/nicolasbonnici/gorest

## Changelog

### v1.0.0 (2025-12-31)

- Initial spellcheck plugin release
- Complete CRUD operations example
- Repository pattern implementation
- Full documentation and examples
- CI/CD pipeline
- Test coverage examples
