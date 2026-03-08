package spellcheck

import (
	"testing"
)

func TestCheckRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CheckRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: CheckRequest{
				Text: "Hello world",
			},
			wantErr: false,
		},
		{
			name: "valid request with language",
			req: CheckRequest{
				Text:     "Hello world",
				Language: "en",
			},
			wantErr: false,
		},
		{
			name: "valid request with context",
			req: CheckRequest{
				Text:    "API endpoint",
				Context: []string{"API", "endpoint"},
			},
			wantErr: false,
		},
		{
			name: "empty text",
			req: CheckRequest{
				Text: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckResponse_Structure(t *testing.T) {
	// Test that CheckResponse can be created with all fields
	response := CheckResponse{
		Valid: false,
		Errors: []*SpellingError{
			{
				Word:        "teh",
				Position:    0,
				Suggestions: []string{"the", "tea"},
			},
		},
		Suggestions: map[string][]string{
			"teh": {"the", "tea"},
		},
		Text: "teh test",
	}

	if response.Valid {
		t.Error("Expected Valid to be false when errors present")
	}

	if len(response.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(response.Errors))
	}

	if len(response.Suggestions) != 1 {
		t.Errorf("Expected 1 suggestion, got %d", len(response.Suggestions))
	}

	if response.Text == "" {
		t.Error("Expected Text to be non-empty")
	}
}

func TestCheckOptions_Structure(t *testing.T) {
	caseSensitive := true
	maxSuggestions := 10

	options := CheckOptions{
		CaseSensitive:  &caseSensitive,
		MaxSuggestions: &maxSuggestions,
	}

	if options.CaseSensitive == nil || *options.CaseSensitive != true {
		t.Error("Expected CaseSensitive to be settable")
	}

	if options.MaxSuggestions == nil || *options.MaxSuggestions != 10 {
		t.Error("Expected MaxSuggestions to be settable")
	}
}
