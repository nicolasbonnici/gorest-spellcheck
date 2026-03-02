package spellcheck

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/logger"
)

type Handler struct {
	repo   Repository
	config *Config
}

func NewHandler(repo Repository, config *Config) *Handler {
	return &Handler{
		repo:   repo,
		config: config,
	}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateItemRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Log.Error("Failed to parse create request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		logger.Log.Error("Invalid user ID", "user_id", userID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	item := &Item{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		UserID:      userUUID,
		Active:      req.Active,
		CreatedAt:   time.Now(),
	}

	if err := h.repo.Create(c.Context(), item); err != nil {
		logger.Log.Error("Failed to create item", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create item",
		})
	}

	logger.Log.Info("Item created successfully", "id", item.ID, "user_id", userUUID)
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	item, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		logger.Log.Error("Failed to get item", "id", id, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Item not found",
		})
	}

	return c.JSON(item)
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			if l > h.config.MaxItems {
				limit = h.config.MaxItems
			} else {
				limit = l
			}
		}
	}

	offset := 0
	if offsetParam := c.Query("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	items, total, err := h.repo.List(c.Context(), userUUID, limit, offset)
	if err != nil {
		logger.Log.Error("Failed to list items", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list items",
		})
	}

	if items == nil {
		items = []Item{}
	}

	return c.JSON(ListItemsResponse{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req UpdateItemRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Log.Error("Failed to parse update request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Active != nil {
		updates["active"] = *req.Active
	}

	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No fields to update",
		})
	}

	if err := h.repo.Update(c.Context(), id, userUUID, updates); err != nil {
		logger.Log.Error("Failed to update item", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update item",
		})
	}

	item, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated item",
		})
	}

	logger.Log.Info("Item updated successfully", "id", id, "user_id", userUUID)
	return c.JSON(item)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ID format",
		})
	}

	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := h.repo.Delete(c.Context(), id, userUUID); err != nil {
		logger.Log.Error("Failed to delete item", "id", id, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Item not found or permission denied",
		})
	}

	logger.Log.Info("Item deleted successfully", "id", id, "user_id", userUUID)
	return c.Status(fiber.StatusNoContent).Send(nil)
}
