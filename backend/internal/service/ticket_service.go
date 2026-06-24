package service

import (
	"context"
	"errors"
	"time"

	"morepark/internal/domain"
	"morepark/internal/repository"
	"morepark/internal/util"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTicketNotFound  = errors.New("билет не найден")
	ErrTicketUsed      = errors.New("билет уже использован")
	ErrTicketCancelled = errors.New("билет отменён")
)

type TicketService struct {
	ticketTypeRepo *repository.TicketTypeRepository
	ticketRepo     *repository.TicketRepository
	saleRepo       *repository.SaleRepository
	zoneRepo       *repository.ZoneRepository
	db             *pgxpool.Pool
}

func NewTicketService(
	ticketTypeRepo *repository.TicketTypeRepository,
	ticketRepo *repository.TicketRepository,
	saleRepo *repository.SaleRepository,
	zoneRepo *repository.ZoneRepository,
	db *pgxpool.Pool,
) *TicketService {
	return &TicketService{
		ticketTypeRepo: ticketTypeRepo,
		ticketRepo:     ticketRepo,
		saleRepo:       saleRepo,
		zoneRepo:       zoneRepo,
		db:             db,
	}
}

// GetTicketTypes возвращает все типы билетов
func (s *TicketService) GetTicketTypes(ctx context.Context) ([]domain.TicketType, error) {
	return s.ticketTypeRepo.GetAll(ctx)
}

// SellRequest — запрос на продажу билета
type SellRequest struct {
	TicketTypeID  string `json:"ticket_type_id"`
	ZoneID        string `json:"zone_id"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	PaymentMethod string `json:"payment_method"`
	CashierID     string `json:"cashier_id"`
}

// SellTicket продаёт билет с проверкой лимита зоны
// ВСЁ В ТРАНЗАКЦИИ — чтобы не было рассинхрона!
func (s *TicketService) SellTicket(ctx context.Context, req SellRequest) (*domain.Sale, error) {
	// 1. Получаем тип билета
	ticketType, err := s.ticketTypeRepo.GetByID(ctx, req.TicketTypeID)
	if err != nil {
		return nil, errors.New("тип билета не найден")
	}

	// 2. Получаем зону и проверяем лимит
	zone, err := s.zoneRepo.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, errors.New("зона не найдена")
	}

	if zone.CurrentCount+guestCount(ticketType) > zone.Capacity {
		return nil, ErrZoneFull
	}

	// 3. Начинаем ТРАНЗАКЦИЮ
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) // откатится, если не сделаем Commit

	// 4. Создаём билет
	now := time.Now()
	quantity := guestCount(ticketType)
	ticket := &domain.Ticket{
		TicketTypeID:  req.TicketTypeID,
		ZoneID:        req.ZoneID,
		Status:        "active",
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		TicketNumber:  util.GenerateTicketNumber(),
		Source:        "cashier",
		Quantity:      quantity,
		ValidFrom:     now,
		ValidUntil:    now.Add(time.Duration(ticketType.DurationHours) * time.Hour),
	}

	if err := s.ticketRepo.Create(ctx, tx, ticket); err != nil {
		return nil, err
	}

	// 5. Увеличиваем счётчик зоны
	if err := s.zoneRepo.IncrementCountByTx(ctx, tx, zone.ID, quantity); err != nil {
		return nil, err
	}

	// 6. Создаём запись о продаже
	sale := &domain.Sale{
		TicketID:      ticket.ID,
		CashierID:     req.CashierID,
		Amount:        ticketType.Price,
		PaymentMethod: req.PaymentMethod,
		IsRefund:      false,
	}

	if err := s.saleRepo.Create(ctx, tx, sale); err != nil {
		return nil, err
	}

	// 7. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// 8. Возвращаем sale с вложенными данными
	sale.Ticket = ticket
	return sale, nil
}

// RefundRequest — запрос на возврат
type RefundRequest struct {
	CashierID string `json:"cashier_id"`
}

// RefundTicket оформляет возврат билета
func (s *TicketService) RefundTicket(ctx context.Context, ticketID string, req RefundRequest) (*domain.Sale, error) {
	// 1. Получаем билет
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, ErrTicketNotFound
	}

	// 2. Проверяем статус
	if ticket.Status == "used" {
		return nil, ErrTicketUsed
	}
	if ticket.Status == "cancelled" {
		return nil, ErrTicketCancelled
	}

	// 3. Транзакция
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 4. Отменяем билет
	if err := s.ticketRepo.UpdateStatus(ctx, tx, ticketID, "cancelled"); err != nil {
		return nil, err
	}

	// 5. Уменьшаем счётчик зоны
	quantity := ticket.Quantity
	if quantity <= 0 {
		quantity = 1
	}
	if err := s.zoneRepo.DecrementCountByTx(ctx, tx, ticket.ZoneID, quantity); err != nil {
		return nil, err
	}

	// 6. Создаём запись о возврате (с отрицательной суммой)
	refundAmount := ticket.TicketType.Price * float64(quantity)
	refund := &domain.Sale{
		TicketID:      ticketID,
		CashierID:     req.CashierID,
		Amount:        -refundAmount,
		PaymentMethod: "refund",
		IsRefund:      true,
	}

	if err := s.saleRepo.Create(ctx, tx, refund); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	refund.Ticket = ticket
	return refund, nil
}

// GetSales возвращает историю продаж
func (s *TicketService) GetSales(ctx context.Context, limit int) ([]domain.Sale, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.saleRepo.GetAll(ctx, limit)
}

// GetTicket возвращает билет по ID
func (s *TicketService) GetTicket(ctx context.Context, id string) (*domain.Ticket, error) {
	return s.ticketRepo.GetByID(ctx, id)
}

func guestCount(ticketType *domain.TicketType) int {
	if ticketType.Type == "group" {
		return 5
	}
	return 1
}
