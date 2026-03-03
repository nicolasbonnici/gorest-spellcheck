package spellcheck

// CheckRequest represents a request to check spelling
type CheckRequest struct {
	Text     string            `json:"text" validate:"required"`
	Language string            `json:"language,omitempty"`
	Context  []string          `json:"context,omitempty"` // Additional words to ignore for this request
	Options  *CheckOptions     `json:"options,omitempty"`
}

// CheckOptions contains optional parameters for spell checking
type CheckOptions struct {
	CaseSensitive  *bool `json:"case_sensitive,omitempty"`
	MaxSuggestions *int  `json:"max_suggestions,omitempty"`
}

// CheckResponse represents the response from spell checking
type CheckResponse struct {
	Valid       bool                       `json:"valid"`
	Errors      []*SpellingError           `json:"errors,omitempty"`
	Suggestions map[string][]string        `json:"suggestions,omitempty"`
	Text        string                     `json:"text,omitempty"`
}

// Validate validates the CheckRequest
func (r *CheckRequest) Validate() error {
	if r.Text == "" {
		return NewValidationError("text is required", map[string]string{
			"text": "cannot be empty",
		})
	}

	return nil
}
