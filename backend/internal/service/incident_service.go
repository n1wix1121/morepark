package service

import (
	"context"
	"errors"
	"fmt"

	"morepark/internal/domain"
	"morepark/internal/repository"
)

var (
	ErrIncidentNotFound      = errors.New("инцидент не найден")
	ErrInvalidSeverity       = errors.New("недопустимый уровень серьёзности")
	ErrInvalidIncidentStatus = errors.New("недопустимый статус инцидента") // ← ПЕРЕИМЕНОВАНО
	ErrAlreadyClosed         = errors.New("инцидент уже закрыт")
)

// Допустимые уровни серьёзности
var validSeverities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

// Допустимые статусы инцидентов  ← ПЕРЕИМЕНОВАНО
var validIncidentStatuses = map[string]bool{
	"open":        true,
	"in_progress": true,
	"closed":      true,
}

type IncidentService struct {
	incidentRepo *repository.IncidentRepository
	zoneRepo     *repository.ZoneRepository
}

func NewIncidentService(incidentRepo *repository.IncidentRepository, zoneRepo *repository.ZoneRepository) *IncidentService {
	return &IncidentService{
		incidentRepo: incidentRepo,
		zoneRepo:     zoneRepo,
	}
}

// CreateRequest — запрос на создание инцидента
type CreateIncidentRequest struct {
	ZoneID      string `json:"zone_id"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	LifeguardID string `json:"lifeguard_id"`
}

// Create создаёт новый инцидент
func (s *IncidentService) Create(ctx context.Context, req CreateIncidentRequest) (*domain.Incident, error) {
	// 1. Валидация серьёзности
	if !validSeverities[req.Severity] {
		return nil, ErrInvalidSeverity
	}

	// 2. Проверяем, что зона существует
	zone, err := s.zoneRepo.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("зона не найдена: %w", err)
	}

	// 3. Создаём инцидент
	incident := &domain.Incident{
		ZoneID:      req.ZoneID,
		LifeguardID: req.LifeguardID,
		Description: req.Description,
		Severity:    req.Severity,
		Status:      "open",
	}

	if err := s.incidentRepo.Create(ctx, incident); err != nil {
		return nil, err
	}

	// 4. Заполняем вложенные данные
	incident.Zone = zone

	// 5. Логируем
	fmt.Printf("🚨 ИНЦИДЕНТ (%s) в зоне '%s': %s\n", req.Severity, zone.Name, req.Description)

	return incident, nil
}

// GetAll возвращает все инциденты
func (s *IncidentService) GetAll(ctx context.Context, limit int) ([]domain.Incident, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.incidentRepo.GetAll(ctx, limit)
}

// GetByID возвращает инцидент по ID
func (s *IncidentService) GetByID(ctx context.Context, id string) (*domain.Incident, error) {
	return s.incidentRepo.GetByID(ctx, id)
}

// GetActive возвращает только активные инциденты
func (s *IncidentService) GetActive(ctx context.Context) ([]domain.Incident, error) {
	return s.incidentRepo.GetActive(ctx)
}

// GetByZone возвращает инциденты по зоне
func (s *IncidentService) GetByZone(ctx context.Context, zoneID string, limit int) ([]domain.Incident, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.incidentRepo.GetByZone(ctx, zoneID, limit)
}

// UpdateStatusRequest — запрос на изменение статуса
type UpdateStatusRequest struct {
	Status     string  `json:"status"`
	Resolution *string `json:"resolution"`
	ResolvedBy string  `json:"resolved_by"`
}

// UpdateStatus изменяет статус инцидента
func (s *IncidentService) UpdateStatus(ctx context.Context, id string, req UpdateStatusRequest) (*domain.Incident, error) {
	// 1. Валидация статуса  ← ИСПОЛЬЗУЕМ НОВОЕ ИМЯ
	if !validIncidentStatuses[req.Status] {
		return nil, ErrInvalidIncidentStatus
	}

	// 2. Получаем текущий инцидент
	incident, err := s.incidentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrIncidentNotFound
	}

	// 3. Проверяем, не закрыт ли уже
	if incident.Status == "closed" {
		return nil, ErrAlreadyClosed
	}

	// 4. Если закрываем — требуем описание решения
	if req.Status == "closed" {
		if req.Resolution == nil || *req.Resolution == "" {
			return nil, errors.New("при закрытии инцидента необходимо указать описание решения")
		}
	}

	// 5. Обновляем статус
	err = s.incidentRepo.UpdateStatus(ctx, id, req.Status, &req.ResolvedBy, req.Resolution)
	if err != nil {
		return nil, err
	}

	// 6. Возвращаем обновлённый инцидент
	return s.incidentRepo.GetByID(ctx, id)
}
