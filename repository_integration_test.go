//go:build integration
// +build integration

package spellcheck

import (
	"context"
	"os"
	"testing"

	"github.com/nicolasbonnici/gorest/database/postgres"
)

// TestRepository_Integration demonstrates integration testing with a real database.
// Run with: go test -v -race -tags=integration ./...
//
// Prerequisites:
//   - PostgreSQL running on localhost:5432
//   - Test database created
//   - Environment variable POSTGRES_URL set (e.g., postgres://test:test@localhost:5432/test?sslmode=disable)
func TestRepository_Integration(t *testing.T) {
	// Skip if no database URL provided
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		t.Skip("POSTGRES_URL not set, skipping integration test")
	}

	// Connect to test database
	db, err := postgres.New(dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create test schema
	ctx := context.Background()
	_, err = db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS spellcheck_items (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("failed to create test schema: %v", err)
	}

	// Cleanup after test
	defer func() {
		_, _ = db.Exec(ctx, "DROP TABLE IF EXISTS spellcheck_items")
	}()

	// Initialize repository
	repo := NewRepository(db)

	t.Run("Create", func(t *testing.T) {
		item := &SpellcheckItem{
			Name:        "Integration Test Item",
			Description: "Created during integration test",
		}

		err := repo.Create(ctx, item)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		if item.Id == "" {
			t.Error("Expected ID to be set after creation")
		}

		// Cleanup
		defer func() {
			_ = repo.Delete(ctx, item.Id)
		}()

		// Verify item was created
		retrieved, err := repo.GetByID(ctx, item.Id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.Name != item.Name {
			t.Errorf("Expected name %q, got %q", item.Name, retrieved.Name)
		}
	})

	t.Run("List", func(t *testing.T) {
		// Create test items
		items := []*SpellcheckItem{
			{Name: "Test Item 1", Description: "First test item"},
			{Name: "Test Item 2", Description: "Second test item"},
		}

		for _, item := range items {
			if err := repo.Create(ctx, item); err != nil {
				t.Fatalf("Failed to create test item: %v", err)
			}
			defer func(id string) {
				_ = repo.Delete(ctx, id)
			}(item.Id)
		}

		// Test pagination
		retrieved, total, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}

		if total < 2 {
			t.Errorf("Expected at least 2 items, got %d", total)
		}

		if len(retrieved) == 0 {
			t.Error("Expected to retrieve items, got none")
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Create item
		item := &SpellcheckItem{
			Name:        "Original Name",
			Description: "Original description",
		}

		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}
		defer func() {
			_ = repo.Delete(ctx, item.Id)
		}()

		// Update item
		updatedItem := &SpellcheckItem{
			Name:        "Updated Name",
			Description: "Updated description",
		}

		if err := repo.Update(ctx, item.Id, updatedItem); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		// Verify update
		retrieved, err := repo.GetByID(ctx, item.Id)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("Expected name %q, got %q", "Updated Name", retrieved.Name)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		// Create item
		item := &SpellcheckItem{
			Name:        "To Be Deleted",
			Description: "This will be deleted",
		}

		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("Failed to create item: %v", err)
		}

		// Delete item
		if err := repo.Delete(ctx, item.Id); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify deletion
		_, err := repo.GetByID(ctx, item.Id)
		if err == nil {
			t.Error("Expected error when getting deleted item, got nil")
		}
	})
}
