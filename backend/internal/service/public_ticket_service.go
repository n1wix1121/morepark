package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"time"

	"morepark/internal/domain"
	"morepark/internal/repository"
	"morepark/internal/util"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/skip2/go-qrcode"
)

var (
	ErrNotEnoughSlots = errors.New("недостаточно мест на выбранное время")
)

type PublicTicketService struct {
	ticketTypeRepo *repository.TicketTypeRepository
	ticketRepo     *repository.TicketRepository
	saleRepo       *repository.SaleRepository
	zoneRepo       *repository.ZoneRepository
	db             *pgxpool.Pool
}

func NewPublicTicketService(
	ticketTypeRepo *repository.TicketTypeRepository,
	ticketRepo *repository.TicketRepository,
	saleRepo *repository.SaleRepository,
	zoneRepo *repository.ZoneRepository,
	db *pgxpool.Pool,
) *PublicTicketService {
	return &PublicTicketService{
		ticketTypeRepo: ticketTypeRepo,
		ticketRepo:     ticketRepo,
		saleRepo:       saleRepo,
		zoneRepo:       zoneRepo,
		db:             db,
	}
}

// PublicTicketTypes возвращает типы билетов для публичного просмотра
func (s *PublicTicketService) PublicTicketTypes(ctx context.Context) ([]domain.TicketType, error) {
	return s.ticketTypeRepo.GetAll(ctx)
}

// PublicZones возвращает зоны с загрузкой для публичного просмотра
func (s *PublicTicketService) PublicZones(ctx context.Context) ([]domain.Zone, error) {
	return s.zoneRepo.GetAll(ctx)
}

// CheckAvailabilityRequest — запрос на проверку доступности
type CheckAvailabilityRequest struct {
	ZoneID   string    `json:"zone_id"`
	DateTime time.Time `json:"datetime"`
	Quantity int       `json:"quantity"`
}

// CheckAvailabilityResponse — ответ о доступности
type CheckAvailabilityResponse struct {
	Available    bool   `json:"available"`
	TotalSlots   int    `json:"total_slots"`
	CurrentCount int    `json:"current_count"`
	Reserved     int    `json:"reserved"`
	AvailableNow int    `json:"available_now"`
	Message      string `json:"message"`
}

// CheckAvailability проверяет, есть ли места на выбранное время
func (s *PublicTicketService) CheckAvailability(ctx context.Context, req CheckAvailabilityRequest) (*CheckAvailabilityResponse, error) {
	zone, err := s.zoneRepo.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, errors.New("зона не найдена")
	}

	// Считаем забронированные онлайн-билеты на это время
	reserved, err := s.ticketRepo.CountReservedForSlot(ctx, req.ZoneID, req.DateTime)
	if err != nil {
		return nil, err
	}

	availableNow := zone.Capacity - zone.CurrentCount - reserved

	resp := &CheckAvailabilityResponse{
		TotalSlots:   zone.Capacity,
		CurrentCount: zone.CurrentCount,
		Reserved:     reserved,
		AvailableNow: availableNow,
		Available:    availableNow >= req.Quantity,
	}

	if resp.Available {
		resp.Message = fmt.Sprintf("Доступно %d мест", availableNow)
	} else {
		resp.Message = fmt.Sprintf("Недостаточно мест. Доступно: %d, нужно: %d", availableNow, req.Quantity)
	}

	return resp, nil
}

// PurchaseRequest — запрос на покупку
type PurchaseRequest struct {
	TicketTypeID  string    `json:"ticket_type_id"`
	ZoneID        string    `json:"zone_id"`
	DateTime      time.Time `json:"datetime"`
	Quantity      int       `json:"quantity"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	CustomerEmail string    `json:"customer_email"`
}

// PurchaseResponse — ответ после покупки
type PurchaseResponse struct {
	Ticket       *domain.Ticket `json:"ticket"`
	TicketNumber string         `json:"ticket_number"`
	QRCodeBase64 string         `json:"qr_code_base64"`
	Amount       float64        `json:"amount"`
	Message      string         `json:"message"`
}

// PurchaseTicket покупает билет онлайн
func (s *PublicTicketService) PurchaseTicket(ctx context.Context, req PurchaseRequest) (*PurchaseResponse, error) {
	// 1. Валидация количества
	if req.Quantity <= 0 || req.Quantity > 10 {
		return nil, errors.New("количество должно быть от 1 до 10")
	}

	// 2. Получаем тип билета
	ticketType, err := s.ticketTypeRepo.GetByID(ctx, req.TicketTypeID)
	if err != nil {
		return nil, errors.New("тип билета не найден")
	}

	// 3. Получаем зону
	zone, err := s.zoneRepo.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, errors.New("зона не найдена")
	}

	// 4. Проверяем доступность
	reserved, err := s.ticketRepo.CountReservedForSlot(ctx, req.ZoneID, req.DateTime)
	if err != nil {
		return nil, err
	}

	available := zone.Capacity - zone.CurrentCount - reserved
	if available < req.Quantity {
		return nil, ErrNotEnoughSlots
	}

	// 5. Начинаем транзакцию
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 6. Генерируем номер билета
	ticketNumber := util.GenerateTicketNumber()

	// 7. Создаём билет
	validFrom := req.DateTime
	validUntil := req.DateTime.Add(time.Duration(ticketType.DurationHours) * time.Hour)

	ticket := &domain.Ticket{
		TicketTypeID:  req.TicketTypeID,
		ZoneID:        req.ZoneID,
		Status:        "active",
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
		TicketNumber:  ticketNumber,
		Source:        "online",
		Quantity:      req.Quantity,
		ValidFrom:     validFrom,
		ValidUntil:    validUntil,
	}

	if err := s.ticketRepo.Create(ctx, tx, ticket); err != nil {
		return nil, err
	}

	// 8. Увеличиваем счётчик зоны (бронируем места)
	if err := s.zoneRepo.IncrementCountByTx(ctx, tx, zone.ID, req.Quantity); err != nil {
		return nil, err
	}

	// 9. Создаём запись о продаже (без кассира — онлайн)
	sale := &domain.Sale{
		TicketID:      ticket.ID,
		CashierID:     "", // онлайн-продажа
		Amount:        ticketType.Price * float64(req.Quantity),
		PaymentMethod: "online",
		IsRefund:      false,
	}

	// Используем Exec напрямую для онлайн-продаж без кассира
	_, err = tx.Exec(ctx, `
		INSERT INTO sales (ticket_id, amount, payment_method, is_refund)
		VALUES ($1, $2, $3, $4)
	`, sale.TicketID, sale.Amount, sale.PaymentMethod, sale.IsRefund)
	if err != nil {
		return nil, err
	}

	// 10. Коммитим транзакцию
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// 11. Генерируем QR-код
	qrCodeBase64, err := generateQRCode(ticketNumber)
	if err != nil {
		// Не критично — билет создан, просто без QR
		qrCodeBase64 = ""
	} else {
		// Сохраняем QR в БД
		s.ticketRepo.UpdateQRCode(ctx, ticket.ID, qrCodeBase64)
		ticket.QRCode = qrCodeBase64
	}

	// 12. Запускаем авто-отмену через 30 минут после времени входа
	go s.scheduleAutoCancel(ticket.ID, ticket.ValidFrom.Add(30*time.Minute))

	return &PurchaseResponse{
		Ticket:       ticket,
		TicketNumber: ticketNumber,
		QRCodeBase64: qrCodeBase64,
		Amount:       ticketType.Price * float64(req.Quantity),
		Message:      "Билет успешно забронирован!",
	}, nil
}

// scheduleAutoCancel отменяет билет, если посетитель не пришёл
func (s *PublicTicketService) scheduleAutoCancel(ticketID string, cancelAt time.Time) {
	delay := time.Until(cancelAt)
	if delay <= 0 {
		return
	}

	time.Sleep(delay)

	ctx := context.Background()
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return
	}

	// Отменяем только если билет ещё активен
	if ticket.Status == "active" && ticket.Source == "online" {
		tx, err := s.db.Begin(ctx)
		if err != nil {
			return
		}
		defer tx.Rollback(ctx)

		s.ticketRepo.UpdateStatus(ctx, tx, ticketID, "cancelled")
		s.zoneRepo.DecrementCountByTx(ctx, tx, ticket.ZoneID, ticket.Quantity)
		tx.Commit(ctx)

		fmt.Printf("⏰ Авто-отмена билета %s (посетитель не пришёл)\n", ticket.TicketNumber)
	}
}

// GetTicketByNumber возвращает билет по номеру (для проверки статуса)
func (s *PublicTicketService) GetTicketByNumber(ctx context.Context, number string) (*domain.Ticket, error) {
	return s.ticketRepo.GetByNumber(ctx, number)
}

// ==================== ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ ====================

// generateQRCode генерирует QR-код в base64
func generateQRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	png.Encode(&buf, qr.Image(256))
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
