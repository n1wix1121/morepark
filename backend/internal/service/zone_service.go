package service

import (
	"context"
	"errors"

	"morepark/internal/domain"
	"morepark/internal/repository"
)

var (
	ErrZoneFull     = errors.New("zone is full")
	ErrZoneNotFound = errors.New("зона не найдена")
)

type ZoneService struct {
	zoneRepo *repository.ZoneRepository
}

func NewZoneService(zoneRepo *repository.ZoneRepository) *ZoneService {
	return &ZoneService{zoneRepo: zoneRepo}
}

func (s *ZoneService) GetAll(ctx context.Context) ([]domain.Zone, error) {
	return s.zoneRepo.GetAll(ctx)
}

func (s *ZoneService) GetByID(ctx context.Context, id string) (*domain.Zone, error) {
	return s.zoneRepo.GetByID(ctx, id)
}

func (s *ZoneService) Create(ctx context.Context, zone *domain.Zone) error {
	zone.CurrentCount = 0
	zone.IsActive = true
	return s.zoneRepo.Create(ctx, zone)
}

type UpdateZoneRequest struct {
	Name        string
	Description string
	Capacity    int
}

func (s *ZoneService) Update(ctx context.Context, id string, req UpdateZoneRequest) (*domain.Zone, error) {
	zone, err := s.zoneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrZoneNotFound
	}

	if req.Capacity < zone.CurrentCount {
		return nil, errors.New("вместимость не может быть меньше текущего количества посетителей")
	}

	zone.Name = req.Name
	zone.Description = req.Description
	zone.Capacity = req.Capacity

	if err := s.zoneRepo.Update(ctx, zone); err != nil {
		return nil, err
	}

	return zone, nil
}

func (s *ZoneService) Delete(ctx context.Context, id string) error {
	return s.zoneRepo.Delete(ctx, id)
}

func (s *ZoneService) CheckAvailability(ctx context.Context, zoneID string) (bool, error) {
	zone, err := s.zoneRepo.GetByID(ctx, zoneID)
	if err != nil {
		return false, err
	}

	return zone.CurrentCount < zone.Capacity, nil
}
