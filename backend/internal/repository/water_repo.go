package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type WaterRepository struct {
	db *pgxpool.Pool
}

func NewWaterRepository(db *pgxpool.Pool) *WaterRepository {
	return &WaterRepository{db: db}
}

// Create сохраняет замер воды
func (r *WaterRepository) Create(ctx context.Context, m *domain.WaterQuality) error {
	query := `
		INSERT INTO water_quality (zone_id, ph, chlorine, turbidity, technician_id, is_normal)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, measured_at
	`

	err := r.db.QueryRow(ctx, query,
		m.ZoneID,
		m.PH,
		m.Chlorine,
		m.Turbidity,
		m.TechnicianID,
		m.IsNormal,
	).Scan(&m.ID, &m.MeasuredAt)

	if err != nil {
		return fmt.Errorf("failed to create water measurement: %w", err)
	}

	return nil
}

// GetAll возвращает все замеры (с деталями)
func (r *WaterRepository) GetAll(ctx context.Context, limit int) ([]domain.WaterQuality, error) {
	query := `
		SELECT 
			w.id, w.zone_id, w.ph, w.chlorine, w.turbidity, 
			w.technician_id, w.measured_at, w.is_normal,
			z.name as zone_name,
			u.full_name as technician_name
		FROM water_quality w
		JOIN zones z ON w.zone_id = z.id
		JOIN users u ON w.technician_id = u.id
		ORDER BY w.measured_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query water measurements: %w", err)
	}
	defer rows.Close()

	var measurements []domain.WaterQuality
	for rows.Next() {
		var m domain.WaterQuality
		var zone domain.Zone
		var tech domain.User

		err := rows.Scan(
			&m.ID, &m.ZoneID, &m.PH, &m.Chlorine, &m.Turbidity,
			&m.TechnicianID, &m.MeasuredAt, &m.IsNormal,
			&zone.Name,
			&tech.FullName,
		)
		if err != nil {
			return nil, err
		}

		zone.ID = m.ZoneID
		tech.ID = m.TechnicianID
		m.Zone = &zone
		m.Technician = &tech

		measurements = append(measurements, m)
	}

	return measurements, nil
}

// GetByZoneID возвращает замеры по конкретной зоне
func (r *WaterRepository) GetByZoneID(ctx context.Context, zoneID string, limit int) ([]domain.WaterQuality, error) {
	query := `
		SELECT 
			w.id, w.zone_id, w.ph, w.chlorine, w.turbidity, 
			w.technician_id, w.measured_at, w.is_normal,
			z.name as zone_name,
			u.full_name as technician_name
		FROM water_quality w
		JOIN zones z ON w.zone_id = z.id
		JOIN users u ON w.technician_id = u.id
		WHERE w.zone_id = $1
		ORDER BY w.measured_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, zoneID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []domain.WaterQuality
	for rows.Next() {
		var m domain.WaterQuality
		var zone domain.Zone
		var tech domain.User

		err := rows.Scan(
			&m.ID, &m.ZoneID, &m.PH, &m.Chlorine, &m.Turbidity,
			&m.TechnicianID, &m.MeasuredAt, &m.IsNormal,
			&zone.Name,
			&tech.FullName,
		)
		if err != nil {
			return nil, err
		}

		zone.ID = m.ZoneID
		tech.ID = m.TechnicianID
		m.Zone = &zone
		m.Technician = &tech

		measurements = append(measurements, m)
	}

	return measurements, nil
}

// GetAlerts возвращает все замеры с нарушениями
func (r *WaterRepository) GetAlerts(ctx context.Context, limit int) ([]domain.WaterQuality, error) {
	query := `
		SELECT 
			w.id, w.zone_id, w.ph, w.chlorine, w.turbidity, 
			w.technician_id, w.measured_at, w.is_normal,
			z.name as zone_name,
			u.full_name as technician_name
		FROM water_quality w
		JOIN zones z ON w.zone_id = z.id
		JOIN users u ON w.technician_id = u.id
		WHERE w.is_normal = false
		ORDER BY w.measured_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []domain.WaterQuality
	for rows.Next() {
		var m domain.WaterQuality
		var zone domain.Zone
		var tech domain.User

		err := rows.Scan(
			&m.ID, &m.ZoneID, &m.PH, &m.Chlorine, &m.Turbidity,
			&m.TechnicianID, &m.MeasuredAt, &m.IsNormal,
			&zone.Name,
			&tech.FullName,
		)
		if err != nil {
			return nil, err
		}

		zone.ID = m.ZoneID
		tech.ID = m.TechnicianID
		m.Zone = &zone
		m.Technician = &tech

		alerts = append(alerts, m)
	}

	return alerts, nil
}
