package service

import (
	"context"
	"errors"
	"time"

	"morepark/internal/domain"
	"morepark/internal/repository"
)

var (
	ErrItemNotFound    = errors.New("товар не найден")
	ErrInsufficientQty = errors.New("недостаточно товара на складе")
	ErrInvalidCategory = errors.New("недопустимая категория")
	ErrInvalidMoveType = errors.New("недопустимый тип движения")
)

// Допустимые категории
var validCategories = map[string]bool{
	"chemical": true, // химия
	"drinks":   true, // напитки
	"food":     true, // еда
	"supplies": true, // расходники
}

type InventoryService struct {
	inventoryRepo *repository.InventoryRepository
}

func NewInventoryService(inventoryRepo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

// GetAll возвращает все товары с вычисляемыми полями
func (s *InventoryService) GetAll(ctx context.Context) ([]domain.InventoryItem, error) {
	items, err := s.inventoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Вычисляем флаги для каждого товара
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for i := range items {
		items[i].IsLowStock = items[i].Quantity <= items[i].MinQuantity

		if items[i].ExpiryDate != nil {
			daysUntil := int(items[i].ExpiryDate.Sub(today).Hours() / 24)
			items[i].IsExpired = daysUntil < 0
			items[i].IsExpiringSoon = daysUntil >= 0 && daysUntil <= 30
		}
	}

	return items, nil
}

// GetByID возвращает товар по ID
func (s *InventoryService) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	return s.inventoryRepo.GetByID(ctx, id)
}

// CreateRequest — запрос на создание товара
type CreateInventoryRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinQuantity float64 `json:"min_quantity"`
	ExpiryDate  *string `json:"expiry_date"` // YYYY-MM-DD
	Price       float64 `json:"price"`
}

// Create создаёт новый товар
func (s *InventoryService) Create(ctx context.Context, req CreateInventoryRequest) (*domain.InventoryItem, error) {
	if !validCategories[req.Category] {
		return nil, ErrInvalidCategory
	}

	item := &domain.InventoryItem{
		Name:        req.Name,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinQuantity: req.MinQuantity,
		Price:       req.Price,
	}

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return nil, errors.New("неверный формат даты")
		}
		item.ExpiryDate = &t
	}

	if err := s.inventoryRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateRequest — запрос на обновление
type UpdateInventoryRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	MinQuantity float64 `json:"min_quantity"`
	ExpiryDate  *string `json:"expiry_date"`
	Price       float64 `json:"price"`
}

// Update обновляет товар
func (s *InventoryService) Update(ctx context.Context, id string, req UpdateInventoryRequest) (*domain.InventoryItem, error) {
	item, err := s.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrItemNotFound
	}

	if !validCategories[req.Category] {
		return nil, ErrInvalidCategory
	}

	item.Name = req.Name
	item.Category = req.Category
	item.Quantity = req.Quantity
	item.Unit = req.Unit
	item.MinQuantity = req.MinQuantity
	item.Price = req.Price

	if req.ExpiryDate != nil && *req.ExpiryDate != "" {
		t, err := time.Parse("2006-01-02", *req.ExpiryDate)
		if err != nil {
			return nil, errors.New("неверный формат даты")
		}
		item.ExpiryDate = &t
	}

	if err := s.inventoryRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// MoveRequest — запрос на движение товара
type MoveRequest struct {
	Type     string  `json:"type"` // in, out
	Quantity float64 `json:"quantity"`
	Reason   string  `json:"reason"`
	UserID   string  `json:"user_id"`
}

// Move выполняет приход или расход товара
func (s *InventoryService) Move(ctx context.Context, inventoryID string, req MoveRequest) (*domain.InventoryMovement, error) {
	// 1. Валидация типа движения
	if req.Type != "in" && req.Type != "out" {
		return nil, ErrInvalidMoveType
	}

	// 2. Получаем товар
	item, err := s.inventoryRepo.GetByID(ctx, inventoryID)
	if err != nil {
		return nil, ErrItemNotFound
	}

	// 3. Проверяем количество для расхода
	if req.Type == "out" && item.Quantity < req.Quantity {
		return nil, ErrInsufficientQty
	}

	// 4. Создаём движение
	movement := &domain.InventoryMovement{
		InventoryID: inventoryID,
		Type:        req.Type,
		Quantity:    req.Quantity,
		UserID:      req.UserID,
		Reason:      req.Reason,
	}

	if err := s.inventoryRepo.CreateMovement(ctx, movement); err != nil {
		return nil, err
	}

	// 5. Обновляем количество товара
	var newQuantity float64
	if req.Type == "in" {
		newQuantity = item.Quantity + req.Quantity
	} else {
		newQuantity = item.Quantity - req.Quantity
	}

	if err := s.inventoryRepo.UpdateQuantity(ctx, inventoryID, newQuantity); err != nil {
		return nil, err
	}

	// 6. Заполняем вложенные данные
	movement.Inventory = item
	movement.Inventory.Quantity = newQuantity

	return movement, nil
}

// GetMovements возвращает историю движений
func (s *InventoryService) GetMovements(ctx context.Context, inventoryID string, limit int) ([]domain.InventoryMovement, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.inventoryRepo.GetMovements(ctx, inventoryID, limit)
}

// GetLowStock возвращает товары с низким остатком
func (s *InventoryService) GetLowStock(ctx context.Context) ([]domain.LowStockAlert, error) {
	return s.inventoryRepo.GetLowStock(ctx)
}

// GetExpiring возвращает товары с истекающим сроком
func (s *InventoryService) GetExpiring(ctx context.Context, withinDays int) ([]domain.ExpiringAlert, error) {
	if withinDays <= 0 {
		withinDays = 30
	}
	return s.inventoryRepo.GetExpiring(ctx, withinDays)
}

// GetExpired возвращает просроченные товары
func (s *InventoryService) GetExpired(ctx context.Context) ([]domain.ExpiringAlert, error) {
	return s.inventoryRepo.GetExpired(ctx)
}
