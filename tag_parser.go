package spellcheck

import (
	"reflect"
	"sync"
)

// FieldInfo contains metadata about a struct field for spellchecking
type FieldInfo struct {
	Name          string // Go field name
	JSONName      string // JSON tag name
	SpellCheck    bool   // Whether to spellcheck this field
	IsStringField bool   // Whether the field is a string type
}

// TagParser parses struct tags and caches the results
type TagParser struct {
	cache sync.Map // map[reflect.Type][]FieldInfo
}

// NewTagParser creates a new TagParser
func NewTagParser() *TagParser {
	return &TagParser{}
}

// Parse extracts spellcheck metadata from a struct
// Returns a slice of FieldInfo for fields that should be spellchecked
func (p *TagParser) Parse(v interface{}) []FieldInfo {
	// Get the type
	t := reflect.TypeOf(v)

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Only works with structs
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Check cache first
	if cached, ok := p.cache.Load(t); ok {
		return cached.([]FieldInfo)
	}

	// Parse the struct
	fields := p.parseStruct(t)

	// Cache the result
	p.cache.Store(t, fields)

	return fields
}

// ParseValue extracts spellcheck metadata from a reflect.Value
// This is useful when you already have a reflect.Value
func (p *TagParser) ParseValue(v reflect.Value) []FieldInfo {
	t := v.Type()

	// Handle pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}

	// Only works with structs
	if t.Kind() != reflect.Struct {
		return nil
	}

	// Check cache first
	if cached, ok := p.cache.Load(t); ok {
		return cached.([]FieldInfo)
	}

	// Parse the struct
	fields := p.parseStruct(t)

	// Cache the result
	p.cache.Store(t, fields)

	return fields
}

// parseStruct does the actual parsing of struct tags
func (p *TagParser) parseStruct(t reflect.Type) []FieldInfo {
	var fields []FieldInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the spellcheck tag
		spellcheckTag := field.Tag.Get("spellcheck")

		// Only include fields with spellcheck:"true"
		if spellcheckTag != "true" {
			continue
		}

		// Only string fields can be spellchecked
		if field.Type.Kind() != reflect.String {
			continue
		}

		// Get JSON tag name (use field name if not present)
		jsonName := field.Tag.Get("json")
		if jsonName == "" {
			jsonName = field.Name
		} else {
			// Handle json:",omitempty" format
			for idx := 0; idx < len(jsonName); idx++ {
				if jsonName[idx] == ',' {
					jsonName = jsonName[:idx]
					break
				}
			}
		}

		fields = append(fields, FieldInfo{
			Name:          field.Name,
			JSONName:      jsonName,
			SpellCheck:    true,
			IsStringField: true,
		})
	}

	return fields
}

// GetFieldValue extracts the string value of a field from a struct
func (p *TagParser) GetFieldValue(v interface{}, fieldName string) (string, bool) {
	val := reflect.ValueOf(v)

	// Handle pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Only works with structs
	if val.Kind() != reflect.Struct {
		return "", false
	}

	// Get the field by name
	fieldVal := val.FieldByName(fieldName)
	if !fieldVal.IsValid() {
		return "", false
	}

	// Ensure it's a string
	if fieldVal.Kind() != reflect.String {
		return "", false
	}

	return fieldVal.String(), true
}

// ExtractFieldsFromMap extracts fields from a map[string]interface{}
// This is useful for checking JSON request bodies before they're unmarshaled
func (p *TagParser) ExtractFieldsFromMap(data map[string]interface{}, fieldInfos []FieldInfo) map[string]string {
	result := make(map[string]string)

	for _, info := range fieldInfos {
		// Try to get value using JSON name
		if val, ok := data[info.JSONName]; ok {
			// Convert to string if it's a string
			if strVal, ok := val.(string); ok {
				result[info.JSONName] = strVal
			}
		}
	}

	return result
}

// ClearCache clears the tag parser cache
// Useful for testing or if types are redefined at runtime
func (p *TagParser) ClearCache() {
	p.cache = sync.Map{}
}

// CacheSize returns the approximate number of types cached
// Note: This is O(n) as sync.Map doesn't expose size directly
func (p *TagParser) CacheSize() int {
	count := 0
	p.cache.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
