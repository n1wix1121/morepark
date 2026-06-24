package repository

import (
	"context"
	"fmt"
	"time"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EquipmentRepository struct {
	db *pgxpool.Pool
}

func NewEquipmentRepository(db *pgxpool.Pool) *EquipmentRepository {
	return &EquipmentRepository{db: db}
}

// GetAll возвращает всё оборудование с деталями
func (r *EquipmentRepository) GetAll(ctx context.Context) ([]domain.Equipment, error) {
	query := `
		SELECT 
			e.id, e.name, e.serial_number, e.zone_id, e.status,
			e.last_maintenance, e.next_maintenance, e.created_at,
			z.name as zone_name
		FROM equipment e
		LEFT JOIN zones z ON e.zone_id = z.id
		ORDER BY e.name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query equipment: %w", err)
	}
	defer rows.Close()

	var equipment []domain.Equipment
	for rows.Next() {
		var e domain.Equipment
		var zoneName *string

		err := rows.Scan(
			&e.ID, &e.Name, &e.SerialNumber, &e.ZoneID, &e.Status,
			&e.LastMaintenance, &e.NextMaintenance, &e.CreatedAt,
			&zoneName,
		)
		if err != nil {
			return nil, err
		}

		if zoneName != nil && e.ZoneID != nil {
			e.Zone = &domain.Zone{
				ID:   *e.ZoneID,
				Name: *zoneName,
			}
		}

		equipment = append(equipment, e)
	}

	return equipment, nil
}

// GetByID возвращает оборудование по ID
func (r *EquipmentRepository) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	query := `
		SELECT 
			e.id, e.name, e.serial_number, e.zone_id, e.status,
			e.last_maintenance, e.next_maintenance, e.created_at,
			z.name as zone_name
		FROM equipment e
		LEFT JOIN zones z ON e.zone_id = z.id
		WHERE e.id = $1
	`

	var e domain.Equipment
	var zoneName *string

	err := r.db.QueryRow(ctx, query, id).Scan(
		&e.ID, &e.Name, &e.SerialNumber, &e.ZoneID, &e.Status,
		&e.LastMaintenance, &e.NextMaintenance, &e.CreatedAt,
		&zoneName,
	)
	if err != nil {
		return nil, fmt.Errorf("equipment not found: %w", err)
	}

	if zoneName != nil && e.ZoneID != nil {
		e.Zone = &domain.Zone{
			ID:   *e.ZoneID,
			Name: *zoneName,
		}
	}

	return &e, nil
}

// Create создаёт новое оборудование
func (r *EquipmentRepository) Create(ctx context.Context, e *domain.Equipment) error {
	query := `
		INSERT INTO equipment (name, serial_number, zone_id, status, last_maintenance, next_maintenance)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		e.Name, e.SerialNumber, e.ZoneID, e.Status,
		e.LastMaintenance, e.NextMaintenance,
	).Scan(&e.ID, &e.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create equipment: %w", err)
	}

	return nil
}

// Update обновляет оборудование
func (r *EquipmentRepository) Update(ctx context.Context, e *domain.Equipment) error {
	query := `
		UPDATE equipment
		SET name = $1, serial_number = $2, zone_id = $3, status = $4,
		    last_maintenance = $5, next_maintenance = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(ctx, query,
		e.Name, e.SerialNumber, e.ZoneID, e.Status,
		e.LastMaintenance, e.NextMaintenance, e.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	return nil
}

// GetUpcomingMaintenance возвращает оборудование с ТО в ближайшие N дней
func (r *EquipmentRepository) GetUpcomingMaintenance(ctx context.Context, withinDays int) ([]domain.UpcomingMaintenance, error) {
	query := `
		SELECT 
			e.id, e.name, e.serial_number, e.next_maintenance
		FROM equipment e
		WHERE e.next_maintenance IS NOT NULL
		  AND e.next_maintenance >= CURRENT_DATE
		  AND e.next_maintenance <= CURRENT_DATE + $1 * INTERVAL '1 day'
		  AND e.status != 'broken'
		ORDER BY e.next_maintenance ASC
	`

	rows, err := r.db.Query(ctx, query, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var upcoming []domain.UpcomingMaintenance
	for rows.Next() {
		var u domain.UpcomingMaintenance
		err := rows.Scan(&u.EquipmentID, &u.EquipmentName, &u.SerialNumber, &u.NextMaintenance)
		if err != nil {
			return nil, err
		}

		// Считаем дни до ТО
		daysUntil := int(u.NextMaintenance.Sub(today).Hours() / 24)
		u.DaysUntil = daysUntil

		// Определяем уровень срочности
		switch {
		case daysUntil <= 0:
			u.UrgencyLevel = "critical"
		case daysUntil == 1:
			u.UrgencyLevel = "high"
		case daysUntil <= 3:
			u.UrgencyLevel = "medium"
		default:
			u.UrgencyLevel = "low"
		}

		upcoming = append(upcoming, u)
	}

	return upcoming, nil
}

// GetMaintenanceLogs возвращает историю ТО для оборудования
func (r *EquipmentRepository) GetMaintenanceLogs(ctx context.Context, equipmentID string, limit int) ([]domain.MaintenanceLog, error) {
	query := `
		SELECT 
			m.id, m.equipment_id, m.technician_id, m.description, m.completed_at,
			u.full_name as technician_name,
			e.name as equipment_name
		FROM maintenance_logs m
		JOIN users u ON m.technician_id = u.id
		JOIN equipment e ON m.equipment_id = e.id
		WHERE m.equipment_id = $1
		ORDER BY m.completed_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, equipmentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.MaintenanceLog
	for rows.Next() {
		var l domain.MaintenanceLog
		var tech domain.User
		var eq domain.Equipment

		err := rows.Scan(
			&l.ID, &l.EquipmentID, &l.TechnicianID, &l.Description, &l.CompletedAt,
			&tech.FullName,
			&eq.Name,
		)
		if err != nil {
			return nil, err
		}

		tech.ID = l.TechnicianID
		eq.ID = l.EquipmentID
		l.Technician = &tech
		l.Equipment = &eq

		logs = append(logs, l)
	}

	return logs, nil
}

// CreateMaintenanceLog создаёт запись о ТО
func (r *EquipmentRepository) CreateMaintenanceLog(ctx context.Context, log *domain.MaintenanceLog) error {
	query := `
		INSERT INTO maintenance_logs (equipment_id, technician_id, description)
		VALUES ($1, $2, $3)
		RETURNING id, completed_at
	`

	err := r.db.QueryRow(ctx, query,
		log.EquipmentID, log.TechnicianID, log.Description,
	).Scan(&log.ID, &log.CompletedAt)

	if err != nil {
		return fmt.Errorf("failed to create maintenance log: %w", err)
	}

	return nil
}
