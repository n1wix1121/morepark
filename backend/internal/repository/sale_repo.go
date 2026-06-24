package repository

import (
	"context"
	"fmt"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SaleRepository struct {
	db *pgxpool.Pool
}

func NewSaleRepository(db *pgxpool.Pool) *SaleRepository {
	return &SaleRepository{db: db}
}

// Create создаёт запись о продаже
func (r *SaleRepository) Create(ctx context.Context, tx pgx.Tx, sale *domain.Sale) error {
	query := `
		INSERT INTO sales (ticket_id, cashier_id, amount, payment_method, is_refund)
		VALUES ($1, $2, $3, $4, $5)
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
		sale.TicketID,
		sale.CashierID,
		sale.Amount,
		sale.PaymentMethod,
		sale.IsRefund,
	).Scan(&sale.ID, &sale.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create sale: %w", err)
	}
	return nil
}

// GetAll возвращает историю продаж
func (r *SaleRepository) GetAll(ctx context.Context, limit int) ([]domain.Sale, error) {
	// 👇 ВАЖНО: LEFT JOIN + COALESCE для защиты от NULL
	query := `
		SELECT 
			s.id, 
			s.ticket_id, 
			COALESCE(s.cashier_id::text, ''),
			s.amount, 
			COALESCE(s.payment_method, 'cash'),
			COALESCE(s.is_refund, false),
			s.created_at,
			COALESCE(u.full_name, 'Онлайн'),
			COALESCE(z.name, '—')
		FROM sales s
		LEFT JOIN tickets t ON s.ticket_id = t.id
		LEFT JOIN users u ON s.cashier_id = u.id
		LEFT JOIN zones z ON t.zone_id = z.id
		ORDER BY s.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query sales: %w", err)
	}
	defer rows.Close()

	var sales []domain.Sale
	for rows.Next() {
		var s domain.Sale
		var cashierID string
		var cashier domain.User
		var zone domain.Zone

		err := rows.Scan(
			&s.ID, &s.TicketID, &cashierID, &s.Amount, &s.PaymentMethod, &s.IsRefund, &s.CreatedAt,
			&cashier.FullName,
			&zone.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale: %w", err)
		}

		s.CashierID = cashierID
		if cashierID != "" {
			cashier.ID = cashierID
			s.Cashier = &cashier
		}

		zone.ID = ""
		s.Ticket = &domain.Ticket{Zone: &zone}

		sales = append(sales, s)
	}
	return sales, rows.Err()
}
