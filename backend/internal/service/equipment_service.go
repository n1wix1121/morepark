package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"morepark/internal/domain"
	"morepark/internal/repository"
)

var (
	ErrEquipmentNotFound = errors.New("оборудование не найдено")
	ErrInvalidStatus     = errors.New("недопустимый статус")
)

// Допустимые статусы
var validStatuses = map[string]bool{
	"working":     true,
	"maintenance": true,
	"broken":      true,
}

type EquipmentService struct {
	equipmentRepo *repository.EquipmentRepository
}

func NewEquipmentService(equipmentRepo *repository.EquipmentRepository) *EquipmentService {
	return &EquipmentService{equipmentRepo: equipmentRepo}
}

// GetAll возвращает всё оборудование
func (s *EquipmentService) GetAll(ctx context.Context) ([]domain.Equipment, error) {
	return s.equipmentRepo.GetAll(ctx)
}

// GetByID возвращает оборудование по ID
func (s *EquipmentService) GetByID(ctx context.Context, id string) (*domain.Equipment, error) {
	return s.equipmentRepo.GetByID(ctx, id)
}

// CreateRequest — запрос на создание оборудования
type CreateEquipmentRequest struct {
	Name            string  `json:"name"`
	SerialNumber    string  `json:"serial_number"`
	ZoneID          *string `json:"zone_id"`
	Status          string  `json:"status"`
	NextMaintenance *string `json:"next_maintenance"` // формат: YYYY-MM-DD
}

// Create создаёт новое оборудование
func (s *EquipmentService) Create(ctx context.Context, req CreateEquipmentRequest) (*domain.Equipment, error) {
	// Валидация статуса
	if !validStatuses[req.Status] {
		return nil, ErrInvalidStatus
	}

	e := &domain.Equipment{
		Name:         req.Name,
		SerialNumber: req.SerialNumber,
		ZoneID:       req.ZoneID,
		Status:       req.Status,
	}

	// Парсим дату следующего ТО
	if req.NextMaintenance != nil && *req.NextMaintenance != "" {
		t, err := time.Parse("2006-01-02", *req.NextMaintenance)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты: %w", err)
		}
		e.NextMaintenance = &t
	}

	if err := s.equipmentRepo.Create(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// UpdateRequest — запрос на обновление
type UpdateEquipmentRequest struct {
	Name            string  `json:"name"`
	SerialNumber    string  `json:"serial_number"`
	ZoneID          *string `json:"zone_id"`
	Status          string  `json:"status"`
	NextMaintenance *string `json:"next_maintenance"`
}

// Update обновляет оборудование
func (s *EquipmentService) Update(ctx context.Context, id string, req UpdateEquipmentRequest) (*domain.Equipment, error) {
	// Получаем текущее оборудование
	e, err := s.equipmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrEquipmentNotFound
	}

	if !validStatuses[req.Status] {
		return nil, ErrInvalidStatus
	}

	e.Name = req.Name
	e.SerialNumber = req.SerialNumber
	e.ZoneID = req.ZoneID
	e.Status = req.Status

	if req.NextMaintenance != nil && *req.NextMaintenance != "" {
		t, err := time.Parse("2006-01-02", *req.NextMaintenance)
		if err != nil {
			return nil, fmt.Errorf("неверный формат даты: %w", err)
		}
		e.NextMaintenance = &t
	}

	if err := s.equipmentRepo.Update(ctx, e); err != nil {
		return nil, err
	}

	return e, nil
}

// CompleteMaintenanceRequest — запрос на завершение ТО
type CompleteMaintenanceRequest struct {
	Description     string `json:"description" binding:"required"`
	NextMaintenance string `json:"next_maintenance" binding:"required"` // YYYY-MM-DD
	TechnicianID    string `json:"technician_id"`
}

// CompleteMaintenance фиксирует ТО и планирует следующее
func (s *EquipmentService) CompleteMaintenance(ctx context.Context, equipmentID string, req CompleteMaintenanceRequest) (*domain.MaintenanceLog, error) {
	// 1. Проверяем, что оборудование существует
	e, err := s.equipmentRepo.GetByID(ctx, equipmentID)
	if err != nil {
		return nil, ErrEquipmentNotFound
	}

	// 2. Парсим дату следующего ТО
	nextDate, err := time.Parse("2006-01-02", req.NextMaintenance)
	if err != nil {
		return nil, fmt.Errorf("неверный формат даты: %w", err)
	}

	// 3. Создаём запись в журнале ТО
	log := &domain.MaintenanceLog{
		EquipmentID:  equipmentID,
		TechnicianID: req.TechnicianID,
		Description:  req.Description,
	}

	if err := s.equipmentRepo.CreateMaintenanceLog(ctx, log); err != nil {
		return nil, err
	}

	// 4. Обновляем оборудование: last_maintenance = сегодня, next_maintenance = новая дата
	now := time.Now()
	e.LastMaintenance = &now
	e.NextMaintenance = &nextDate
	e.Status = "working" // после ТО оборудование снова в работе

	if err := s.equipmentRepo.Update(ctx, e); err != nil {
		return nil, err
	}

	log.Equipment = e
	return log, nil
}

// GetMaintenanceLogs возвращает историю ТО
func (s *EquipmentService) GetMaintenanceLogs(ctx context.Context, equipmentID string, limit int) ([]domain.MaintenanceLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.equipmentRepo.GetMaintenanceLogs(ctx, equipmentID, limit)
}

// GetUpcomingMaintenance возвращает оборудование с предстоящим ТО
func (s *EquipmentService) GetUpcomingMaintenance(ctx context.Context, withinDays int) ([]domain.UpcomingMaintenance, error) {
	if withinDays <= 0 {
		withinDays = 7 // по умолчанию — неделя
	}
	return s.equipmentRepo.GetUpcomingMaintenance(ctx, withinDays)
}
