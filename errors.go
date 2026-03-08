package spellcheck

import "fmt"

// SpellingError represents a single spelling error with suggestions
type SpellingError struct {
	Field       string   `json:"field,omitempty"`       // Field name (for middleware validation)
	Word        string   `json:"word"`                  // Misspelled word
	Position    int      `json:"position"`              // Position in text
	Suggestions []string `json:"suggestions,omitempty"` // Suggested corrections
}

func (e *SpellingError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("spelling error in field %q: word %q at position %d", e.Field, e.Word, e.Position)
	}
	return fmt.Sprintf("spelling error: word %q at position %d", e.Word, e.Position)
}

// SpellingErrors is a collection of spelling errors
type SpellingErrors struct {
	Errors []*SpellingError `json:"errors"`
}

func (e *SpellingErrors) Error() string {
	if len(e.Errors) == 0 {
		return "no spelling errors"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	return fmt.Sprintf("%d spelling errors found", len(e.Errors))
}

func (e *SpellingErrors) Add(err *SpellingError) {
	e.Errors = append(e.Errors, err)
}

func (e *SpellingErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// ValidationError represents errors during request validation
type ValidationError struct {
	Message string
	Details map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationError(message string, details map[string]string) *ValidationError {
	return &ValidationError{
		Message: message,
		Details: details,
	}
}

// TextTooLongError indicates text exceeds maximum length
type TextTooLongError struct {
	Length    int
	MaxLength int
}

func (e *TextTooLongError) Error() string {
	return fmt.Sprintf("text length %d exceeds maximum allowed length %d", e.Length, e.MaxLength)
}

// UnsupportedLanguageError indicates requested language is not supported
type UnsupportedLanguageError struct {
	Language           string
	SupportedLanguages []string
}

func (e *UnsupportedLanguageError) Error() string {
	return fmt.Sprintf("language %q is not supported (supported: %v)", e.Language, e.SupportedLanguages)
}
