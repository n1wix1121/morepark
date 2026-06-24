package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketTypeRepository struct {
	db *pgxpool.Pool
}

func NewTicketTypeRepository(db *pgxpool.Pool) *TicketTypeRepository {
	return &TicketTypeRepository{db: db}
}

func (r *TicketTypeRepository) GetAll(ctx context.Context) ([]domain.TicketType, error) {
	query := `
		SELECT id, type, name, price, duration_hours, COALESCE(description, ''), created_at
		FROM ticket_types
		ORDER BY price ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticket types: %w", err)
	}
	defer rows.Close()

	var types []domain.TicketType
	for rows.Next() {
		var t domain.TicketType
		err := rows.Scan(
			&t.ID,
			&t.Type,
			&t.Name,
			&t.Price,
			&t.DurationHours,
			&t.Description,
			&t.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket type: %w", err)
		}
		types = append(types, t)
	}
	return types, nil
}

func (r *TicketTypeRepository) GetByID(ctx context.Context, id string) (*domain.TicketType, error) {
	query := `
		SELECT id, type, name, price, duration_hours, COALESCE(description, ''), created_at
		FROM ticket_types WHERE id = $1
	`
	var t domain.TicketType
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID,
		&t.Type,
		&t.Name,
		&t.Price,
		&t.DurationHours,
		&t.Description,
		&t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
