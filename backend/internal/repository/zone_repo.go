package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ZoneRepository struct {
	db *pgxpool.Pool
}

func NewZoneRepository(db *pgxpool.Pool) *ZoneRepository {
	return &ZoneRepository{db: db}
}

// GetAll возвращает все активные зоны
func (r *ZoneRepository) GetAll(ctx context.Context) ([]domain.Zone, error) {
	query := `
		SELECT id, name, capacity, current_count, description, is_active, created_at
		FROM zones
		WHERE is_active = true
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query zones: %w", err)
	}
	defer rows.Close()

	var zones []domain.Zone
	for rows.Next() {
		var zone domain.Zone
		err := rows.Scan(
			&zone.ID,
			&zone.Name,
			&zone.Capacity,
			&zone.CurrentCount,
			&zone.Description,
			&zone.IsActive,
			&zone.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan zone: %w", err)
		}
		zones = append(zones, zone)
	}

	return zones, nil
}

// GetByID возвращает зону по ID
func (r *ZoneRepository) GetByID(ctx context.Context, id string) (*domain.Zone, error) {
	query := `
		SELECT id, name, capacity, current_count, description, is_active, created_at
		FROM zones
		WHERE id = $1
	`

	var zone domain.Zone
	err := r.db.QueryRow(ctx, query, id).Scan(
		&zone.ID,
		&zone.Name,
		&zone.Capacity,
		&zone.CurrentCount,
		&zone.Description,
		&zone.IsActive,
		&zone.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("zone not found: %w", err)
	}

	return &zone, nil
}

// Create создаёт новую зону
func (r *ZoneRepository) Create(ctx context.Context, zone *domain.Zone) error {
	query := `
		INSERT INTO zones (name, capacity, current_count, description, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		zone.Name,
		zone.Capacity,
		zone.CurrentCount,
		zone.Description,
		zone.IsActive,
	).Scan(&zone.ID, &zone.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create zone: %w", err)
	}

	return nil
}

// Update обновляет зону
func (r *ZoneRepository) Update(ctx context.Context, zone *domain.Zone) error {
	query := `
		UPDATE zones
		SET name = $1, capacity = $2, description = $3, is_active = $4
		WHERE id = $5
	`

	_, err := r.db.Exec(ctx, query,
		zone.Name,
		zone.Capacity,
		zone.Description,
		zone.IsActive,
		zone.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update zone: %w", err)
	}

	return nil
}

// Delete удаляет зону (мягкое удаление)
func (r *ZoneRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE zones SET is_active = false WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete zone: %w", err)
	}

	return nil
}

// IncrementCount увеличивает счётчик посетителей
func (r *ZoneRepository) IncrementCount(ctx context.Context, id string) error {
	query := `UPDATE zones SET current_count = current_count + 1 WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment count: %w", err)
	}

	return nil
}

// DecrementCount уменьшает счётчик посетителей
func (r *ZoneRepository) DecrementCount(ctx context.Context, id string) error {
	query := `UPDATE zones SET current_count = GREATEST(0, current_count - 1) WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to decrement count: %w", err)
	}

	return nil
}

// IncrementCountTx увеличивает счётчик в рамках транзакции
func (r *ZoneRepository) IncrementCountTx(ctx context.Context, tx pgx.Tx, id string) error {
	query := `UPDATE zones SET current_count = current_count + 1 WHERE id = $1`
	_, err := tx.Exec(ctx, query, id)
	return err
}

// DecrementCountTx уменьшает счётчик в рамках транзакции
func (r *ZoneRepository) DecrementCountTx(ctx context.Context, tx pgx.Tx, id string) error {
	query := `UPDATE zones SET current_count = GREATEST(0, current_count - 1) WHERE id = $1`
	_, err := tx.Exec(ctx, query, id)
	return err
}

// IncrementCountByTx увеличивает счётчик на N в рамках транзакции
func (r *ZoneRepository) IncrementCountByTx(ctx context.Context, tx pgx.Tx, id string, count int) error {
	query := `UPDATE zones SET current_count = current_count + $2 WHERE id = $1`
	_, err := tx.Exec(ctx, query, id, count)
	return err
}

// DecrementCountByTx уменьшает счётчик на N в рамках транзакции
func (r *ZoneRepository) DecrementCountByTx(ctx context.Context, tx pgx.Tx, id string, count int) error {
	query := `UPDATE zones SET current_count = GREATEST(0, current_count - $2) WHERE id = $1`
	_, err := tx.Exec(ctx, query, id, count)
	return err
}
