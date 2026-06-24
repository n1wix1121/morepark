package repository

import (
	"context"
	"fmt"
	"time"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

// Create создаёт билет в рамках транзакции
func (r *TicketRepository) Create(ctx context.Context, tx pgx.Tx, ticket *domain.Ticket) error {
	source := ticket.Source
	if source == "" {
		source = "cashier"
	}
	quantity := ticket.Quantity
	if quantity <= 0 {
		quantity = 1
	}

	query := `
		INSERT INTO tickets (
			ticket_type_id, zone_id, status, customer_name, customer_phone,
			valid_from, valid_until, ticket_number, customer_email, source, quantity
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), $10, $11)
		RETURNING id, created_at
	`

	var executor interface {
		QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	}

	if tx != nil {
		executor = tx
	} else {
		executor = r.db
	}

	err := executor.QueryRow(ctx, query,
		ticket.TicketTypeID,
		ticket.ZoneID,
		ticket.Status,
		ticket.CustomerName,
		ticket.CustomerPhone,
		ticket.ValidFrom,
		ticket.ValidUntil,
		ticket.TicketNumber,
		ticket.CustomerEmail,
		source,
		quantity,
	).Scan(&ticket.ID, &ticket.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}
	return nil
}

// GetByID возвращает билет с вложенными данными
func (r *TicketRepository) GetByID(ctx context.Context, id string) (*domain.Ticket, error) {
	query := `
		SELECT 
			t.id, t.ticket_type_id, t.zone_id, t.status, 
			t.customer_name, t.customer_phone, t.valid_from, t.valid_until, t.created_at,
			COALESCE(t.customer_email, ''),
			COALESCE(t.ticket_number, ''),
			COALESCE(t.source, 'cashier'),
			COALESCE(t.quantity, 1),
			COALESCE(t.qr_code, ''),
			tt.name as type_name, tt.price, tt.duration_hours,
			z.name as zone_name
		FROM tickets t
		JOIN ticket_types tt ON t.ticket_type_id = tt.id
		JOIN zones z ON t.zone_id = z.id
		WHERE t.id = $1
	`

	var ticket domain.Ticket
	var ticketType domain.TicketType
	var zone domain.Zone

	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.TicketTypeID, &ticket.ZoneID, &ticket.Status,
		&ticket.CustomerName, &ticket.CustomerPhone, &ticket.ValidFrom, &ticket.ValidUntil, &ticket.CreatedAt,
		&ticket.CustomerEmail,
		&ticket.TicketNumber,
		&ticket.Source,
		&ticket.Quantity,
		&ticket.QRCode,
		&ticketType.Name, &ticketType.Price, &ticketType.DurationHours,
		&zone.Name,
	)
	if err != nil {
		return nil, err
	}

	ticketType.ID = ticket.TicketTypeID
	zone.ID = ticket.ZoneID
	ticket.TicketType = &ticketType
	ticket.Zone = &zone

	return &ticket, nil
}

// UpdateStatus обновляет статус билета
func (r *TicketRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, id string, status string) error {
	query := `UPDATE tickets SET status = $1 WHERE id = $2`

	if tx != nil {
		_, err := tx.Exec(ctx, query, status, id)
		return err
	}
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

// GetByNumber возвращает билет по номеру
func (r *TicketRepository) GetByNumber(ctx context.Context, number string) (*domain.Ticket, error) {
	query := `
		SELECT 
			t.id, t.ticket_type_id, t.zone_id, t.status, 
			t.customer_name, t.customer_phone, t.valid_from, t.valid_until, t.created_at,
			COALESCE(t.customer_email, ''),
			COALESCE(t.ticket_number, ''),
			COALESCE(t.source, 'cashier'),
			COALESCE(t.quantity, 1),
			COALESCE(t.qr_code, ''),
			tt.name as type_name, tt.price, tt.duration_hours,
			z.name as zone_name
		FROM tickets t
		JOIN ticket_types tt ON t.ticket_type_id = tt.id
		JOIN zones z ON t.zone_id = z.id
		WHERE t.ticket_number = $1
	`

	var ticket domain.Ticket
	var ticketType domain.TicketType
	var zone domain.Zone

	err := r.db.QueryRow(ctx, query, number).Scan(
		&ticket.ID, &ticket.TicketTypeID, &ticket.ZoneID, &ticket.Status,
		&ticket.CustomerName, &ticket.CustomerPhone, &ticket.ValidFrom, &ticket.ValidUntil, &ticket.CreatedAt,
		&ticket.CustomerEmail,
		&ticket.TicketNumber,
		&ticket.Source,
		&ticket.Quantity,
		&ticket.QRCode,
		&ticketType.Name, &ticketType.Price, &ticketType.DurationHours,
		&zone.Name,
	)
	if err != nil {
		return nil, err
	}

	ticketType.ID = ticket.TicketTypeID
	zone.ID = ticket.ZoneID
	ticket.TicketType = &ticketType
	ticket.Zone = &zone

	return &ticket, nil
}

// CountReservedForSlot считает забронированные билеты на конкретное время
func (r *TicketRepository) CountReservedForSlot(ctx context.Context, zoneID string, validFrom time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(quantity), 0)
		FROM tickets
		WHERE zone_id = $1
		  AND status = 'active'
		  AND source = 'online'
		  AND valid_from <= $2
		  AND valid_until >= $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, zoneID, validFrom).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateQRCode сохраняет QR-код для билета
func (r *TicketRepository) UpdateQRCode(ctx context.Context, id string, qrCode string) error {
	query := `UPDATE tickets SET qr_code = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, qrCode, id)
	return err
}
