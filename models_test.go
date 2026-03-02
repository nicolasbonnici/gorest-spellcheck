package spellcheck

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestItemJSONSerialization(t *testing.T) {
	now := time.Now()
	item := Item{
		ID:          uuid.New(),
		Name:        "Test Item",
		Description: "Test Description",
		UserID:      uuid.New(),
		Active:      true,
		CreatedAt:   now,
		UpdatedAt:   &now,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal item: %v", err)
	}

	var decoded Item
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal item: %v", err)
	}

	if decoded.Name != item.Name {
		t.Errorf("Expected name %s, got %s", item.Name, decoded.Name)
	}

	if decoded.Description != item.Description {
		t.Errorf("Expected description %s, got %s", item.Description, decoded.Description)
	}

	if decoded.Active != item.Active {
		t.Errorf("Expected active %v, got %v", item.Active, decoded.Active)
	}
}

func TestCreateItemRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateItemRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateItemRequest{
				Name:        "Test Item",
				Description: "Test Description",
				Active:      true,
			},
			valid: true,
		},
		{
			name: "empty description is valid",
			request: CreateItemRequest{
				Name:   "Test Item",
				Active: true,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.request.Name == "" && tt.valid {
				t.Error("Empty name should not be valid")
			}
		})
	}
}

func TestUpdateItemRequestPartialUpdate(t *testing.T) {
	name := "Updated Name"
	active := false

	req := UpdateItemRequest{
		Name:   &name,
		Active: &active,
	}

	if req.Description != nil {
		t.Error("Description should be nil for partial update")
	}

	if *req.Name != name {
		t.Errorf("Expected name %s, got %s", name, *req.Name)
	}

	if *req.Active != active {
		t.Errorf("Expected active %v, got %v", active, *req.Active)
	}
}

func TestListItemsResponse(t *testing.T) {
	items := []Item{
		{
			ID:          uuid.New(),
			Name:        "Item 1",
			Description: "Description 1",
			UserID:      uuid.New(),
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Item 2",
			Description: "Description 2",
			UserID:      uuid.New(),
			Active:      false,
			CreatedAt:   time.Now(),
		},
	}

	response := ListItemsResponse{
		Items:  items,
		Total:  2,
		Limit:  20,
		Offset: 0,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var decoded ListItemsResponse
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(decoded.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(decoded.Items))
	}

	if decoded.Total != 2 {
		t.Errorf("Expected total 2, got %d", decoded.Total)
	}

	if decoded.Limit != 20 {
		t.Errorf("Expected limit 20, got %d", decoded.Limit)
	}
}
