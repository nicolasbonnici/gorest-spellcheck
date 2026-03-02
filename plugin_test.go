package spellcheck

import (
	"testing"
)

func TestNewPlugin(t *testing.T) {
	plugin := NewPlugin()
	if plugin == nil {
		t.Fatal("NewPlugin() returned nil")
	}

	if plugin.Name() != "spellcheck" {
		t.Errorf("Expected plugin name 'spellcheck', got %q", plugin.Name())
	}
}

func TestPluginInitialize(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{
		"enabled": true,
	}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if !plugin.config.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if plugin.handler == nil {
		t.Error("Expected handler to be initialized")
	}
}

func TestPluginInitializeWithDefaults(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{}

	err := plugin.Initialize(config)
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Should use defaults
	if !plugin.config.Enabled {
		t.Error("Expected default Enabled to be true")
	}

	if plugin.config.DefaultLanguage != "en" {
		t.Errorf("Expected default DefaultLanguage to be 'en', got %q", plugin.config.DefaultLanguage)
	}
}

func TestPluginInitializeWithInvalidConfig(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{
		"max_text_length": -1000, // Invalid
	}

	err := plugin.Initialize(config)
	if err == nil {
		t.Error("Expected Initialize() to fail with invalid config")
	}
}

func TestPluginHandler(t *testing.T) {
	plugin := NewPlugin().(*SpellcheckPlugin)

	config := map[string]interface{}{
		"enabled": true,
	}

	if err := plugin.Initialize(config); err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	handler := plugin.Handler()
	if handler == nil {
		t.Error("Expected Handler() to return a non-nil handler")
	}
}
