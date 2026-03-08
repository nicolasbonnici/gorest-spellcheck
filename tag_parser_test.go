package spellcheck

import (
	"reflect"
	"testing"
)

// Test structs
type TestArticle struct {
	ID        string `json:"id"`
	Title     string `json:"title" spellcheck:"true"`
	Content   string `json:"content" spellcheck:"true"`
	Slug      string `json:"slug"` // No spellcheck
	ViewCount int    `json:"view_count"`
	//lint:ignore U1000 intentionally unexported to test field filtering
	unexported string `spellcheck:"true"`
}

type TestProduct struct {
	Name        string  `json:"name" spellcheck:"true"`
	Description string  `json:"description,omitempty" spellcheck:"true"`
	SKU         string  `json:"sku"` // No spellcheck
	Price       float64 `json:"price"`
}

type TestEmptyStruct struct{}

type TestNoSpellcheckFields struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type TestNonStringField struct {
	ID    string `json:"id"`
	Count int    `json:"count" spellcheck:"true"` // Should be ignored (not string)
	Title string `json:"title" spellcheck:"true"`
}

func TestNewTagParser(t *testing.T) {
	parser := NewTagParser()
	if parser == nil {
		t.Fatal("NewTagParser() returned nil")
	}
}

func TestTagParser_Parse(t *testing.T) {
	parser := NewTagParser()

	t.Run("parse article struct", func(t *testing.T) {
		article := TestArticle{}
		fields := parser.Parse(article)

		if len(fields) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(fields))
		}

		// Check Title field
		found := false
		for _, f := range fields {
			if f.Name == "Title" && f.JSONName == "title" && f.SpellCheck {
				found = true
				break
			}
		}
		if !found {
			t.Error("Title field not found or incorrect")
		}

		// Check Content field
		found = false
		for _, f := range fields {
			if f.Name == "Content" && f.JSONName == "content" && f.SpellCheck {
				found = true
				break
			}
		}
		if !found {
			t.Error("Content field not found or incorrect")
		}
	})

	t.Run("parse pointer to struct", func(t *testing.T) {
		article := &TestArticle{}
		fields := parser.Parse(article)

		if len(fields) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(fields))
		}
	})

	t.Run("parse product with omitempty", func(t *testing.T) {
		product := TestProduct{}
		fields := parser.Parse(product)

		if len(fields) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(fields))
		}

		// Check Description field (has omitempty)
		found := false
		for _, f := range fields {
			if f.Name == "Description" && f.JSONName == "description" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Description field not parsed correctly with omitempty")
		}
	})

	t.Run("parse empty struct", func(t *testing.T) {
		empty := TestEmptyStruct{}
		fields := parser.Parse(empty)

		if len(fields) != 0 {
			t.Errorf("Expected 0 fields, got %d", len(fields))
		}
	})

	t.Run("parse struct with no spellcheck fields", func(t *testing.T) {
		item := TestNoSpellcheckFields{}
		fields := parser.Parse(item)

		if len(fields) != 0 {
			t.Errorf("Expected 0 fields, got %d", len(fields))
		}
	})

	t.Run("ignore non-string fields", func(t *testing.T) {
		item := TestNonStringField{}
		fields := parser.Parse(item)

		// Should only have Title (string), not Count (int)
		if len(fields) != 1 {
			t.Errorf("Expected 1 field, got %d", len(fields))
		}

		if fields[0].Name != "Title" {
			t.Errorf("Expected Title field, got %s", fields[0].Name)
		}
	})

	t.Run("ignore unexported fields", func(t *testing.T) {
		article := TestArticle{}
		fields := parser.Parse(article)

		// Should not include unexported field
		for _, f := range fields {
			if f.Name == "unexported" {
				t.Error("Unexported field should not be included")
			}
		}
	})

	t.Run("parse non-struct returns nil", func(t *testing.T) {
		str := "not a struct"
		fields := parser.Parse(str)

		if fields != nil {
			t.Error("Expected nil for non-struct, got fields")
		}
	})
}

func TestTagParser_Caching(t *testing.T) {
	parser := NewTagParser()

	article := TestArticle{}

	// Parse once
	fields1 := parser.Parse(article)

	// Parse again - should use cache
	fields2 := parser.Parse(article)

	// Should return same data
	if len(fields1) != len(fields2) {
		t.Error("Cached result differs from original")
	}

	// Check cache size
	size := parser.CacheSize()
	if size != 1 {
		t.Errorf("Expected cache size 1, got %d", size)
	}

	// Parse different type
	product := TestProduct{}
	parser.Parse(product)

	size = parser.CacheSize()
	if size != 2 {
		t.Errorf("Expected cache size 2, got %d", size)
	}
}

func TestTagParser_ClearCache(t *testing.T) {
	parser := NewTagParser()

	article := TestArticle{}
	parser.Parse(article)

	if parser.CacheSize() != 1 {
		t.Error("Cache should have 1 entry")
	}

	parser.ClearCache()

	if parser.CacheSize() != 0 {
		t.Error("Cache should be empty after clear")
	}

	// Parsing after clear should work
	fields := parser.Parse(article)
	if len(fields) != 2 {
		t.Error("Parsing after cache clear should still work")
	}
}

func TestTagParser_ParseValue(t *testing.T) {
	parser := NewTagParser()

	article := TestArticle{
		Title:   "Test",
		Content: "Content",
	}

	val := reflect.ValueOf(article)
	fields := parser.ParseValue(val)

	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Test with pointer
	valPtr := reflect.ValueOf(&article)
	fields = parser.ParseValue(valPtr)

	if len(fields) != 2 {
		t.Errorf("Expected 2 fields from pointer, got %d", len(fields))
	}
}

func TestTagParser_GetFieldValue(t *testing.T) {
	parser := NewTagParser()

	article := TestArticle{
		ID:      "123",
		Title:   "My Title",
		Content: "My Content",
		Slug:    "my-title",
	}

	t.Run("get existing string field", func(t *testing.T) {
		val, ok := parser.GetFieldValue(article, "Title")
		if !ok {
			t.Error("GetFieldValue should return true for existing field")
		}
		if val != "My Title" {
			t.Errorf("Expected 'My Title', got %q", val)
		}
	})

	t.Run("get field from pointer", func(t *testing.T) {
		val, ok := parser.GetFieldValue(&article, "Content")
		if !ok {
			t.Error("GetFieldValue should work with pointer")
		}
		if val != "My Content" {
			t.Errorf("Expected 'My Content', got %q", val)
		}
	})

	t.Run("get non-existent field", func(t *testing.T) {
		_, ok := parser.GetFieldValue(article, "NonExistent")
		if ok {
			t.Error("GetFieldValue should return false for non-existent field")
		}
	})

	t.Run("get non-string field", func(t *testing.T) {
		item := TestNonStringField{Count: 42}
		_, ok := parser.GetFieldValue(item, "Count")
		if ok {
			t.Error("GetFieldValue should return false for non-string field")
		}
	})

	t.Run("get from non-struct", func(t *testing.T) {
		str := "not a struct"
		_, ok := parser.GetFieldValue(str, "Field")
		if ok {
			t.Error("GetFieldValue should return false for non-struct")
		}
	})
}

func TestTagParser_ExtractFieldsFromMap(t *testing.T) {
	parser := NewTagParser()

	article := TestArticle{}
	fieldInfos := parser.Parse(article)

	data := map[string]interface{}{
		"id":      "123",
		"title":   "My Title",
		"content": "My Content",
		"slug":    "my-title",
	}

	t.Run("extract spellcheck fields", func(t *testing.T) {
		result := parser.ExtractFieldsFromMap(data, fieldInfos)

		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}

		if result["title"] != "My Title" {
			t.Errorf("Expected 'My Title', got %q", result["title"])
		}

		if result["content"] != "My Content" {
			t.Errorf("Expected 'My Content', got %q", result["content"])
		}

		// Slug should not be included
		if _, ok := result["slug"]; ok {
			t.Error("Slug should not be in result (no spellcheck tag)")
		}
	})

	t.Run("handle missing fields", func(t *testing.T) {
		partialData := map[string]interface{}{
			"title": "Only Title",
		}

		result := parser.ExtractFieldsFromMap(partialData, fieldInfos)

		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}

		if result["title"] != "Only Title" {
			t.Errorf("Expected 'Only Title', got %q", result["title"])
		}
	})

	t.Run("handle non-string values", func(t *testing.T) {
		badData := map[string]interface{}{
			"title":   123, // Not a string
			"content": "Valid",
		}

		result := parser.ExtractFieldsFromMap(badData, fieldInfos)

		// Should only include content (valid string)
		if len(result) != 1 {
			t.Errorf("Expected 1 field, got %d", len(result))
		}

		if _, ok := result["title"]; ok {
			t.Error("Non-string title should not be included")
		}
	})

	t.Run("handle empty map", func(t *testing.T) {
		emptyData := map[string]interface{}{}
		result := parser.ExtractFieldsFromMap(emptyData, fieldInfos)

		if len(result) != 0 {
			t.Errorf("Expected 0 fields, got %d", len(result))
		}
	})
}

func TestTagParser_ConcurrentAccess(t *testing.T) {
	parser := NewTagParser()

	// Test concurrent parsing (should not panic due to sync.Map)
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			article := TestArticle{}
			_ = parser.Parse(article)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should have cached the type once
	if parser.CacheSize() != 1 {
		t.Errorf("Expected cache size 1, got %d", parser.CacheSize())
	}
}
