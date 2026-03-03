package spellcheck

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSpellchecker(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)

	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	if checker == nil {
		t.Fatal("NewSpellchecker() returned nil")
	}

	if checker.model == nil {
		t.Error("Spellchecker model is nil")
	}

	if checker.config == nil {
		t.Error("Spellchecker config is nil")
	}
}

func TestSpellchecker_Check(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	tests := []struct {
		name          string
		text          string
		expectErrors  bool
		expectedWords []string
	}{
		{
			name:         "correct text",
			text:         "The quick brown fox jumps over the lazy dog",
			expectErrors: false,
		},
		{
			name:          "single misspelled word",
			text:          "The quik brown fox",
			expectErrors:  true,
			expectedWords: []string{"quik"},
		},
		{
			name:          "multiple misspelled words",
			text:          "Teh quik brwn fox",
			expectErrors:  true,
			expectedWords: []string{"Teh", "quik", "brwn"},
		},
		{
			name:         "text with numbers",
			text:         "Version 1.2.3 is ready",
			expectErrors: false,
		},
		{
			name:         "text with punctuation",
			text:         "Hello, world! How are you?",
			expectErrors: false,
		},
		{
			name:         "empty text",
			text:         "",
			expectErrors: false,
		},
		{
			name:         "text with apostrophes",
			text:         "It's a wonderful day, isn't it?",
			expectErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := checker.Check(tt.text)
			if err != nil {
				t.Fatalf("Check() failed: %v", err)
			}

			if tt.expectErrors && !errors.HasErrors() {
				t.Errorf("Expected errors but got none")
			}

			if !tt.expectErrors && errors.HasErrors() {
				t.Errorf("Expected no errors but got: %v", errors)
			}

			if tt.expectErrors && len(tt.expectedWords) > 0 {
				foundWords := make(map[string]bool)
				for _, e := range errors.Errors {
					foundWords[e.Word] = true
				}

				for _, expected := range tt.expectedWords {
					if !foundWords[expected] {
						t.Errorf("Expected to find misspelled word %q but didn't", expected)
					}
				}
			}
		})
	}
}

func TestSpellchecker_CheckTextTooLong(t *testing.T) {
	config := DefaultConfig()
	config.MaxTextLength = 50
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	longText := string(make([]byte, 100))
	_, err = checker.Check(longText)

	if err == nil {
		t.Error("Expected TextTooLongError but got nil")
	}

	if _, ok := err.(*TextTooLongError); !ok {
		t.Errorf("Expected TextTooLongError, got %T", err)
	}
}

func TestSpellchecker_CheckField(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	errors, err := checker.CheckField("title", "Teh quik fox")
	if err != nil {
		t.Fatalf("CheckField() failed: %v", err)
	}

	if !errors.HasErrors() {
		t.Fatal("Expected errors but got none")
	}

	// Check that all errors have the field name set
	for _, e := range errors.Errors {
		if e.Field != "title" {
			t.Errorf("Expected field name 'title', got %q", e.Field)
		}
	}
}

func TestSpellchecker_IgnoredWords(t *testing.T) {
	config := DefaultConfig()
	config.IgnoredWords = []string{"quik", "brwn"}
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	// These words should be ignored
	errors, err := checker.Check("The quik brwn fox")
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if errors.HasErrors() {
		t.Errorf("Expected no errors (words should be ignored), got: %v", errors)
	}
}

func TestSpellchecker_MinWordLength(t *testing.T) {
	config := DefaultConfig()
	config.MinWordLength = 5 // Only check words with 5+ characters
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	// "teh" is too short to check (3 chars), so no errors expected
	errors, err := checker.Check("teh quick")
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if errors.HasErrors() {
		t.Errorf("Expected no errors (word too short), got: %v", errors)
	}
}

func TestSpellchecker_CaseSensitive(t *testing.T) {
	configInsensitive := DefaultConfig()
	configInsensitive.CaseSensitive = false

	checkerInsensitive, err := NewSpellchecker(&configInsensitive)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	// "THE" in all caps should be recognized as correct (case-insensitive)
	errors, err := checkerInsensitive.Check("THE quick fox")
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	if errors.HasErrors() {
		t.Errorf("Expected no errors (case-insensitive), got: %v", errors)
	}
}

func TestSpellchecker_GetSuggestions(t *testing.T) {
	config := DefaultConfig()
	config.MaxSuggestions = 3
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	suggestions := checker.getSuggestions("teh")

	if len(suggestions) == 0 {
		t.Error("Expected suggestions but got none")
	}

	if len(suggestions) > 3 {
		t.Errorf("Expected max 3 suggestions, got %d", len(suggestions))
	}

	// "the" should be in suggestions for "teh"
	found := false
	for _, s := range suggestions {
		if s == "the" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'the' in suggestions for 'teh', got: %v", suggestions)
	}
}

func TestSpellchecker_ExtractWords(t *testing.T) {
	config := DefaultConfig()
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "simple sentence",
			text:     "Hello world",
			expected: []string{"Hello", "world"},
		},
		{
			name:     "with punctuation",
			text:     "Hello, world!",
			expected: []string{"Hello", "world"},
		},
		{
			name:     "with apostrophe",
			text:     "It's a test",
			expected: []string{"It's", "a", "test"},
		},
		{
			name:     "with hyphen",
			text:     "well-known phrase",
			expected: []string{"well-known", "phrase"},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []string{},
		},
		{
			name:     "only punctuation",
			text:     "!!!",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := checker.extractWords(tt.text)

			if len(words) != len(tt.expected) {
				t.Errorf("Expected %d words, got %d", len(tt.expected), len(words))
				return
			}

			for i, expected := range tt.expected {
				if words[i].word != expected {
					t.Errorf("Expected word %q at position %d, got %q", expected, i, words[i].word)
				}
			}
		})
	}
}

func TestSpellchecker_ContainsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"hello", false},
		{"hello123", true},
		{"123", true},
		{"v1.2.3", true},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsDigit(tt.input)
			if result != tt.expected {
				t.Errorf("containsDigit(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSpellchecker_CustomDictionary(t *testing.T) {
	// Create a temporary custom dictionary file
	tmpDir := t.TempDir()
	dictFile := filepath.Join(tmpDir, "custom.txt")

	customWords := "customword\nanotherword\nthirdword"
	if err := os.WriteFile(dictFile, []byte(customWords), 0644); err != nil {
		t.Fatalf("Failed to create custom dictionary: %v", err)
	}

	config := DefaultConfig()
	config.CustomDictionary = dictFile

	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() with custom dictionary failed: %v", err)
	}

	// Check that custom words are recognized
	errors, err := checker.Check("This is a customword test")
	if err != nil {
		t.Fatalf("Check() failed: %v", err)
	}

	// "customword" should be recognized (no errors)
	for _, e := range errors.Errors {
		if e.Word == "customword" {
			t.Errorf("Custom word 'customword' should be recognized but was flagged as error")
		}
	}
}

func TestSpellchecker_SuggestionCapitalization(t *testing.T) {
	config := DefaultConfig()
	config.CaseSensitive = false
	checker, err := NewSpellchecker(&config)
	if err != nil {
		t.Fatalf("NewSpellchecker() failed: %v", err)
	}

	// Check capitalized misspelled word
	suggestions := checker.getSuggestions("Teh")

	if len(suggestions) == 0 {
		t.Fatal("Expected suggestions but got none")
	}

	// First suggestion should be capitalized to match input
	if len(suggestions[0]) > 0 && suggestions[0][0] != 'T' {
		t.Errorf("Expected first suggestion to be capitalized, got: %v", suggestions)
	}
}
