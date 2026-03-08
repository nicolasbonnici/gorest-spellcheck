package spellcheck

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if cfg.DefaultLanguage != "en" {
		t.Errorf("Expected DefaultLanguage to be 'en', got %q", cfg.DefaultLanguage)
	}

	if len(cfg.SupportedLanguages) != 1 || cfg.SupportedLanguages[0] != "en" {
		t.Errorf("Expected SupportedLanguages to be ['en'], got %v", cfg.SupportedLanguages)
	}

	if cfg.MaxTextLength != 10000 {
		t.Errorf("Expected MaxTextLength to be 10000, got %d", cfg.MaxTextLength)
	}

	if cfg.MaxSuggestions != 5 {
		t.Errorf("Expected MaxSuggestions to be 5, got %d", cfg.MaxSuggestions)
	}

	if cfg.MinWordLength != 2 {
		t.Errorf("Expected MinWordLength to be 2, got %d", cfg.MinWordLength)
	}

	if cfg.CaseSensitive {
		t.Error("Expected CaseSensitive to be false")
	}

	if len(cfg.IgnoredWords) != 0 {
		t.Errorf("Expected IgnoredWords to be empty, got %v", cfg.IgnoredWords)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "max_text_length at minimum",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: false,
		},
		{
			name: "max_text_length at maximum",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: false,
		},
		{
			name: "max_text_length too low",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      0,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "max_text_length must be greater than 0",
		},
		{
			name: "max_text_length negative",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      -100,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "max_text_length must be greater than 0",
		},
		{
			name: "max_text_length too high",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      2000000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "max_text_length cannot exceed 1000000",
		},
		{
			name: "max_suggestions negative",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000,
				MaxSuggestions:     -5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "max_suggestions cannot be negative",
		},
		{
			name: "max_suggestions too high",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000,
				MaxSuggestions:     200,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "max_suggestions cannot exceed 100",
		},
		{
			name: "min_word_length zero",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      0,
			},
			wantErr: true,
			errMsg:  "min_word_length must be at least 1",
		},
		{
			name: "min_word_length too high",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      100,
			},
			wantErr: true,
			errMsg:  "min_word_length cannot exceed 50",
		},
		{
			name: "empty default_language",
			config: Config{
				DefaultLanguage:    "",
				SupportedLanguages: []string{"en"},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "default_language cannot be empty",
		},
		{
			name: "empty supported_languages",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "supported_languages cannot be empty",
		},
		{
			name: "default_language not in supported_languages",
			config: Config{
				DefaultLanguage:    "fr",
				SupportedLanguages: []string{"en", "es"},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: true,
			errMsg:  "default_language",
		},
		{
			name: "multiple supported languages",
			config: Config{
				DefaultLanguage:    "en",
				SupportedLanguages: []string{"en", "fr", "es"},
				MaxTextLength:      1000,
				MaxSuggestions:     5,
				MinWordLength:      1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() == "" {
					t.Errorf("Expected error message containing %q, got empty string", tt.errMsg)
				}
			}
		})
	}
}
