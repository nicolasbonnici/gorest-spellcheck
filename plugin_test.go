package spellcheck

import (
	"context"
	"testing"

	"github.com/nicolasbonnici/gorest/database"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()

	if plugin == nil {
		t.Fatal("NewPlugin returned nil")
	}

	spellcheckPlugin, ok := plugin.(*SpellcheckPlugin)
	if !ok {
		t.Fatal("NewPlugin did not return *SpellcheckPlugin")
	}

	if spellcheckPlugin.Name() != "spellcheck" {
		t.Errorf("Expected plugin name 'spellcheck', got '%s'", spellcheckPlugin.Name())
	}
}

func TestPluginInitialize(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{
		"enabled":   true,
		"max_items": 50,
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !plugin.config.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if plugin.config.MaxItems != 50 {
		t.Errorf("Expected MaxItems to be 50, got %d", plugin.config.MaxItems)
	}
}

func TestPluginInitializeWithDefaults(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !plugin.config.Enabled {
		t.Error("Expected default Enabled to be true")
	}

	if plugin.config.MaxItems != 100 {
		t.Errorf("Expected default MaxItems to be 100, got %d", plugin.config.MaxItems)
	}
}

func TestPluginInitializeWithInvalidConfig(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{
		"max_items": 2000,
	}

	err := plugin.Initialize(config)
	if err == nil {
		t.Fatal("Expected error for invalid max_items, got nil")
	}
}

func TestPluginHandler(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	handler := plugin.Handler()
	if handler == nil {
		t.Fatal("Handler returned nil")
	}
}

type mockDatabase struct{}

func (m *mockDatabase) Connect(ctx context.Context, dsn string) error { return nil }
func (m *mockDatabase) Close() error                                  { return nil }
func (m *mockDatabase) Ping(ctx context.Context) error                { return nil }
func (m *mockDatabase) Query(ctx context.Context, query string, args ...interface{}) (database.Rows, error) {
	return nil, nil
}
func (m *mockDatabase) QueryRow(ctx context.Context, query string, args ...interface{}) database.Row {
	return nil
}
func (m *mockDatabase) Exec(ctx context.Context, query string, args ...interface{}) (database.Result, error) {
	return nil, nil
}
func (m *mockDatabase) Begin(ctx context.Context) (database.Tx, error) { return nil, nil }
func (m *mockDatabase) Dialect() database.Dialect                      { return nil }
func (m *mockDatabase) DriverName() string                             { return "mock" }
func (m *mockDatabase) Introspector() database.SchemaIntrospector      { return nil }

func TestPluginInitializeWithDatabase(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	mockDB := &mockDatabase{}

	config := map[string]interface{}{
		"database": database.Database(mockDB),
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize with database failed: %v", err)
	}

	if plugin.db == nil {
		t.Error("Expected database to be set")
	}

	if plugin.repo == nil {
		t.Error("Expected repository to be initialized")
	}

	if plugin.handler == nil {
		t.Error("Expected handler to be initialized")
	}
}
