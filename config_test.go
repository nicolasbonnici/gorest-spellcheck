package spellcheck

import "testing"

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Expected default Enabled to be true")
	}

	if cfg.MaxItems != 100 {
		t.Errorf("Expected default MaxItems to be 100, got %d", cfg.MaxItems)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		wantError bool
	}{
		{
			name: "valid config",
			config: Config{
				Enabled:  true,
				MaxItems: 100,
			},
			wantError: false,
		},
		{
			name: "max items at minimum",
			config: Config{
				Enabled:  true,
				MaxItems: 1,
			},
			wantError: false,
		},
		{
			name: "max items at maximum",
			config: Config{
				Enabled:  true,
				MaxItems: 1000,
			},
			wantError: false,
		},
		{
			name: "max items too low",
			config: Config{
				Enabled:  true,
				MaxItems: 0,
			},
			wantError: true,
		},
		{
			name: "max items negative",
			config: Config{
				Enabled:  true,
				MaxItems: -1,
			},
			wantError: true,
		},
		{
			name: "max items too high",
			config: Config{
				Enabled:  true,
				MaxItems: 1001,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}
