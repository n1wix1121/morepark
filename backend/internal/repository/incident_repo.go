package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IncidentRepository struct {
	db *pgxpool.Pool
}

func NewIncidentRepository(db *pgxpool.Pool) *IncidentRepository {
	return &IncidentRepository{db: db}
}

// Create создаёт новый инцидент
func (r *IncidentRepository) Create(ctx context.Context, inc *domain.Incident) error {
	query := `
		INSERT INTO incidents (zone_id, lifeguard_id, description, severity, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		inc.ZoneID,
		inc.LifeguardID,
		inc.Description,
		inc.Severity,
		inc.Status,
	).Scan(&inc.ID, &inc.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// GetAll возвращает все инциденты с деталями
// GetAll возвращает все инциденты с деталями
func (r *IncidentRepository) GetAll(ctx context.Context, limit int) ([]domain.Incident, error) {
	query := `
		SELECT 
			i.id, i.zone_id, i.lifeguard_id, i.description, i.severity, i.status,
			i.resolved_by, i.resolved_at, i.resolution, i.created_at,
			COALESCE(z.name, '—') as zone_name,
			COALESCE(u.full_name, '—') as lifeguard_name
		FROM incidents i
		LEFT JOIN zones z ON i.zone_id = z.id
		LEFT JOIN users u ON i.lifeguard_id = u.id
		ORDER BY i.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		var zoneName, lifeguardName string

		err := rows.Scan(
			&inc.ID, &inc.ZoneID, &inc.LifeguardID, &inc.Description,
			&inc.Severity, &inc.Status, &inc.ResolvedBy, &inc.ResolvedAt,
			&inc.Resolution, &inc.CreatedAt,
			&zoneName,
			&lifeguardName,
		)
		if err != nil {
			return nil, err
		}

		inc.Zone = &domain.Zone{ID: inc.ZoneID, Name: zoneName}
		inc.Lifeguard = &domain.User{ID: inc.LifeguardID, FullName: lifeguardName}

		incidents = append(incidents, inc)
	}

	return incidents, nil
}

// GetByID возвращает инцидент по ID
func (r *IncidentRepository) GetByID(ctx context.Context, id string) (*domain.Incident, error) {
	query := `
		SELECT 
			i.id, i.zone_id, i.lifeguard_id, i.description, i.severity, i.status,
			i.resolved_by, i.resolved_at, i.resolution, i.created_at,
			z.name as zone_name,
			u.full_name as lifeguard_name
		FROM incidents i
		JOIN zones z ON i.zone_id = z.id
		JOIN users u ON i.lifeguard_id = u.id
		WHERE i.id = $1
	`

	var inc domain.Incident
	var zone domain.Zone
	var lifeguard domain.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&inc.ID, &inc.ZoneID, &inc.LifeguardID, &inc.Description,
		&inc.Severity, &inc.Status, &inc.ResolvedBy, &inc.ResolvedAt,
		&inc.Resolution, &inc.CreatedAt,
		&zone.Name,
		&lifeguard.FullName,
	)
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}

	zone.ID = inc.ZoneID
	lifeguard.ID = inc.LifeguardID
	inc.Zone = &zone
	inc.Lifeguard = &lifeguard

	return &inc, nil
}

// UpdateStatus обновляет статус инцидента
func (r *IncidentRepository) UpdateStatus(ctx context.Context, id string, status string, resolvedBy *string, resolution *string) error {
	query := `
		UPDATE incidents
		SET status = $1, resolved_by = $2, resolved_at = $3, resolution = $4
		WHERE id = $5
	`

	var resolvedAt interface{}
	if status == "closed" {
		// Используем CURRENT_TIMESTAMP для resolved_at
		query = `
			UPDATE incidents
			SET status = $1, resolved_by = $2, resolved_at = CURRENT_TIMESTAMP, resolution = $3
			WHERE id = $4
		`
		_, err := r.db.Exec(ctx, query, status, resolvedBy, resolution, id)
		return err
	}

	_, err := r.db.Exec(ctx, query, status, resolvedBy, resolvedAt, resolution, id)
	return err
}

// GetActive возвращает только открытые и в работе инциденты
func (r *IncidentRepository) GetActive(ctx context.Context) ([]domain.Incident, error) {
	query := `
		SELECT 
			i.id, i.zone_id, i.lifeguard_id, i.description, i.severity, i.status,
			i.resolved_by, i.resolved_at, i.resolution, i.created_at,
			z.name as zone_name,
			u.full_name as lifeguard_name
		FROM incidents i
		JOIN zones z ON i.zone_id = z.id
		JOIN users u ON i.lifeguard_id = u.id
		WHERE i.status IN ('open', 'in_progress')
		ORDER BY 
			CASE i.severity 
				WHEN 'high' THEN 1 
				WHEN 'medium' THEN 2 
				WHEN 'low' THEN 3 
			END,
			i.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		var zone domain.Zone
		var lifeguard domain.User

		err := rows.Scan(
			&inc.ID, &inc.ZoneID, &inc.LifeguardID, &inc.Description,
			&inc.Severity, &inc.Status, &inc.ResolvedBy, &inc.ResolvedAt,
			&inc.Resolution, &inc.CreatedAt,
			&zone.Name,
			&lifeguard.FullName,
		)
		if err != nil {
			return nil, err
		}

		zone.ID = inc.ZoneID
		lifeguard.ID = inc.LifeguardID
		inc.Zone = &zone
		inc.Lifeguard = &lifeguard

		incidents = append(incidents, inc)
	}

	return incidents, nil
}

// GetByZone возвращает инциденты по конкретной зоне
func (r *IncidentRepository) GetByZone(ctx context.Context, zoneID string, limit int) ([]domain.Incident, error) {
	query := `
		SELECT 
			i.id, i.zone_id, i.lifeguard_id, i.description, i.severity, i.status,
			i.resolved_by, i.resolved_at, i.resolution, i.created_at,
			z.name as zone_name,
			u.full_name as lifeguard_name
		FROM incidents i
		JOIN zones z ON i.zone_id = z.id
		JOIN users u ON i.lifeguard_id = u.id
		WHERE i.zone_id = $1
		ORDER BY i.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, zoneID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var inc domain.Incident
		var zone domain.Zone
		var lifeguard domain.User

		err := rows.Scan(
			&inc.ID, &inc.ZoneID, &inc.LifeguardID, &inc.Description,
			&inc.Severity, &inc.Status, &inc.ResolvedBy, &inc.ResolvedAt,
			&inc.Resolution, &inc.CreatedAt,
			&zone.Name,
			&lifeguard.FullName,
		)
		if err != nil {
			return nil, err
		}

		zone.ID = inc.ZoneID
		lifeguard.ID = inc.LifeguardID
		inc.Zone = &zone
		inc.Lifeguard = &lifeguard

		incidents = append(incidents, inc)
	}

	return incidents, nil
}
