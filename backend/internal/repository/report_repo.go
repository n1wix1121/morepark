package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SaleExportRow struct {
	ID            string
	Amount        float64
	PaymentMethod string
	IsRefund      bool
	CreatedAt     time.Time
	CashierName   string
	ZoneName      string
	TicketNumber  string
	TicketType    string
	Quantity      int
	CustomerName  string
	Source        string
}

type InventoryExportRow struct {
	Name        string
	Category    string
	Quantity    float64
	Unit        string
	MinQuantity float64
	Price       float64
	TotalValue  float64
	ExpiryDate  *time.Time
}

type MovementExportRow struct {
	CreatedAt  time.Time
	ItemName   string
	Category   string
	Type       string
	Quantity   float64
	Unit       string
	Price      float64
	TotalValue float64
	Reason     string
	UserName   string
}

type ReportRepository struct {
	db *pgxpool.Pool
}

func NewReportRepository(db *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{db: db}
}

// GetSalesForExport — продажи для Excel
func (r *ReportRepository) GetSalesForExport(ctx context.Context) ([]SaleExportRow, error) {
	// 👇 ВАЖНО: LEFT JOIN вместо JOIN для защиты от NULL
	query := `
		SELECT 
			s.id, 
			s.amount, 
			s.payment_method, 
			COALESCE(s.is_refund, false), 
			s.created_at,
			COALESCE(u.full_name, 'Онлайн'),
			COALESCE(z.name, '—'),
			COALESCE(t.ticket_number, '—'),
			COALESCE(tt.name, '—'),
			COALESCE(t.quantity, 1),
			COALESCE(t.customer_name, '—'),
			COALESCE(t.source, 'cashier')
		FROM sales s
		JOIN tickets t ON s.ticket_id = t.id
		LEFT JOIN zones z ON t.zone_id = z.id
		LEFT JOIN users u ON s.cashier_id = u.id
		LEFT JOIN ticket_types tt ON t.ticket_type_id = tt.id
		ORDER BY s.created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to export sales: %w", err)
	}
	defer rows.Close()

	var result []SaleExportRow
	for rows.Next() {
		var row SaleExportRow
		if err := rows.Scan(
			&row.ID, &row.Amount, &row.PaymentMethod, &row.IsRefund, &row.CreatedAt,
			&row.CashierName, &row.ZoneName, &row.TicketNumber, &row.TicketType,
			&row.Quantity, &row.CustomerName, &row.Source,
		); err != nil {
			return nil, fmt.Errorf("failed to scan sale row: %w", err)
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

// GetInventoryForExport — остатки для Excel
func (r *ReportRepository) GetInventoryForExport(ctx context.Context) ([]InventoryExportRow, error) {
	query := `
		SELECT 
			name, 
			category, 
			quantity, 
			unit, 
			min_quantity, 
			COALESCE(price, 0),
			COALESCE(quantity * price, 0),
			expiry_date
		FROM inventory
		ORDER BY category, name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to export inventory: %w", err)
	}
	defer rows.Close()

	var result []InventoryExportRow
	for rows.Next() {
		var row InventoryExportRow
		if err := rows.Scan(
			&row.Name, &row.Category, &row.Quantity, &row.Unit,
			&row.MinQuantity, &row.Price, &row.TotalValue, &row.ExpiryDate,
		); err != nil {
			return nil, fmt.Errorf("failed to scan inventory row: %w", err)
		}
		result = append(result, row)
	}

	return result, rows.Err()
}

// GetMovementsForExport — движения склада для Excel
func (r *ReportRepository) GetMovementsForExport(ctx context.Context) ([]MovementExportRow, error) {
	// 👇 Проверяем существование таблицы
	var tableExists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_name = 'inventory_movements'
		)
	`).Scan(&tableExists)

	if err != nil {
		return []MovementExportRow{}, nil
	}

	if !tableExists {
		return []MovementExportRow{}, nil
	}

	// 👇 ВАЖНО: LEFT JOIN для users
	query := `
		SELECT 
			m.created_at, 
			i.name, 
			i.category, 
			m.type, 
			m.quantity, 
			COALESCE(i.unit, 'шт'),
			COALESCE(i.price, 0),
			COALESCE(m.quantity * i.price, 0),
			COALESCE(m.reason, '—'),
			COALESCE(u.full_name, '—')
		FROM inventory_movements m
		JOIN inventory i ON m.inventory_id = i.id
		LEFT JOIN users u ON m.user_id = u.id
		ORDER BY m.created_at DESC
		LIMIT 5000
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to export movements: %w", err)
	}
	defer rows.Close()

	var result []MovementExportRow
	for rows.Next() {
		var row MovementExportRow
		if err := rows.Scan(
			&row.CreatedAt, &row.ItemName, &row.Category, &row.Type,
			&row.Quantity, &row.Unit, &row.Price, &row.TotalValue,
			&row.Reason, &row.UserName,
		); err != nil {
			return nil, fmt.Errorf("failed to scan movement row: %w", err)
		}
		result = append(result, row)
	}

	return result, rows.Err()
}
