package spellcheck

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nicolasbonnici/gorest/database"
)

type Repository interface {
	Create(ctx context.Context, item *Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*Item, error)
	List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Item, int, error)
	Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates map[string]interface{}) error
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type repository struct {
	db database.Database
}

func NewRepository(db database.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, item *Item) error {
	query := `
		INSERT INTO spellcheck_items (id, name, description, user_id, active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		item.ID,
		item.Name,
		item.Description,
		item.UserID,
		item.Active,
		item.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Item, error) {
	query := `
		SELECT id, name, description, user_id, active, created_at, updated_at
		FROM spellcheck_items
		WHERE id = $1
	`

	var item Item
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.UserID,
		&item.Active,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	return &item, nil
}

func (r *repository) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]Item, int, error) {
	countQuery := `SELECT COUNT(*) FROM spellcheck_items WHERE user_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count items: %w", err)
	}

	query := `
		SELECT id, name, description, user_id, active, created_at, updated_at
		FROM spellcheck_items
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list items: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.UserID,
			&item.Active,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating items: %w", err)
	}

	return items, total, nil
}

func (r *repository) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, updates map[string]interface{}) error {
	query := `
		UPDATE spellcheck_items
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    active = COALESCE($3, active),
		    updated_at = NOW()
		WHERE id = $4 AND user_id = $5
	`

	result, err := r.db.Exec(
		ctx,
		query,
		updates["name"],
		updates["description"],
		updates["active"],
		id,
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("item not found or permission denied")
	}

	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM spellcheck_items WHERE id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("item not found or permission denied")
	}

	return nil
}
