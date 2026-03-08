package spellcheck

import (
	"testing"
)

func TestSpellingError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *SpellingError
		want string
	}{
		{
			name: "error with field",
			err: &SpellingError{
				Field:    "title",
				Word:     "teh",
				Position: 0,
			},
			want: `spelling error in field "title": word "teh" at position 0`,
		},
		{
			name: "error without field",
			err: &SpellingError{
				Word:     "wrld",
				Position: 10,
			},
			want: `spelling error: word "wrld" at position 10`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("SpellingError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSpellingErrors(t *testing.T) {
	errors := &SpellingErrors{}

	if errors.HasErrors() {
		t.Error("Expected HasErrors() to be false for empty errors")
	}

	if got := errors.Error(); got != "no spelling errors" {
		t.Errorf("Expected error message 'no spelling errors', got %q", got)
	}

	// Add one error
	errors.Add(&SpellingError{Word: "teh", Position: 0})

	if !errors.HasErrors() {
		t.Error("Expected HasErrors() to be true after adding error")
	}

	if len(errors.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors.Errors))
	}

	if got := errors.Error(); got != `spelling error: word "teh" at position 0` {
		t.Errorf("Expected single error message, got %q", got)
	}

	// Add another error
	errors.Add(&SpellingError{Word: "wrld", Position: 10})

	if len(errors.Errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors.Errors))
	}

	if got := errors.Error(); got != "2 spelling errors found" {
		t.Errorf("Expected '2 spelling errors found', got %q", got)
	}
}

func TestValidationError(t *testing.T) {
	details := map[string]string{
		"field1": "error1",
		"field2": "error2",
	}

	err := NewValidationError("validation failed", details)

	if err.Message != "validation failed" {
		t.Errorf("Expected message 'validation failed', got %q", err.Message)
	}

	if len(err.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(err.Details))
	}

	if err.Error() != "validation failed" {
		t.Errorf("Expected Error() to return 'validation failed', got %q", err.Error())
	}
}

func TestTextTooLongError(t *testing.T) {
	err := &TextTooLongError{
		Length:    20000,
		MaxLength: 10000,
	}

	expected := "text length 20000 exceeds maximum allowed length 10000"
	if got := err.Error(); got != expected {
		t.Errorf("TextTooLongError.Error() = %v, want %v", got, expected)
	}
}

func TestUnsupportedLanguageError(t *testing.T) {
	err := &UnsupportedLanguageError{
		Language:           "fr",
		SupportedLanguages: []string{"en", "es"},
	}

	expected := `language "fr" is not supported (supported: [en es])`
	if got := err.Error(); got != expected {
		t.Errorf("UnsupportedLanguageError.Error() = %v, want %v", got, expected)
	}
}
