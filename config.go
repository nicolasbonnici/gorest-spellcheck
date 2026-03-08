package spellcheck

import (
	"errors"
	"fmt"
)

type Config struct {
	Enabled            bool
	DefaultLanguage    string
	SupportedLanguages []string
	MaxTextLength      int
	MaxSuggestions     int
	IgnoredWords       []string
	CustomDictionary   string
	CaseSensitive      bool
	MinWordLength      int
}

func DefaultConfig() Config {
	return Config{
		Enabled:            true,
		DefaultLanguage:    "en",
		SupportedLanguages: []string{"en"},
		MaxTextLength:      10000,
		MaxSuggestions:     5,
		IgnoredWords:       []string{},
		CustomDictionary:   "",
		CaseSensitive:      false,
		MinWordLength:      2,
	}
}

func (c *Config) Validate() error {
	if c.MaxTextLength < 1 {
		return errors.New("max_text_length must be greater than 0")
	}

	if c.MaxTextLength > 1000000 {
		return fmt.Errorf("max_text_length cannot exceed 1000000, got %d", c.MaxTextLength)
	}

	if c.MaxSuggestions < 0 {
		return errors.New("max_suggestions cannot be negative")
	}

	if c.MaxSuggestions > 100 {
		return fmt.Errorf("max_suggestions cannot exceed 100, got %d", c.MaxSuggestions)
	}

	if c.MinWordLength < 1 {
		return errors.New("min_word_length must be at least 1")
	}

	if c.MinWordLength > 50 {
		return fmt.Errorf("min_word_length cannot exceed 50, got %d", c.MinWordLength)
	}

	if c.DefaultLanguage == "" {
		return errors.New("default_language cannot be empty")
	}

	if len(c.SupportedLanguages) == 0 {
		return errors.New("supported_languages cannot be empty")
	}

	found := false
	for _, lang := range c.SupportedLanguages {
		if lang == c.DefaultLanguage {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("default_language %q must be in supported_languages %v", c.DefaultLanguage, c.SupportedLanguages)
	}

	return nil
}
