package repository

import (
	"context"
	"fmt"
	"time"

	"morepark/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type InventoryRepository struct {
	db *pgxpool.Pool
}

func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// GetAll возвращает все товары
func (r *InventoryRepository) GetAll(ctx context.Context) ([]domain.InventoryItem, error) {
	query := `
		SELECT id, name, category, quantity, unit, min_quantity, expiry_date, price, created_at
		FROM inventory
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query inventory: %w", err)
	}
	defer rows.Close()

	var items []domain.InventoryItem
	for rows.Next() {
		var item domain.InventoryItem
		err := rows.Scan(
			&item.ID, &item.Name, &item.Category, &item.Quantity, &item.Unit,
			&item.MinQuantity, &item.ExpiryDate, &item.Price, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetByID возвращает товар по ID
func (r *InventoryRepository) GetByID(ctx context.Context, id string) (*domain.InventoryItem, error) {
	query := `
		SELECT id, name, category, quantity, unit, min_quantity, expiry_date, price, created_at
		FROM inventory
		WHERE id = $1
	`

	var item domain.InventoryItem
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.Name, &item.Category, &item.Quantity, &item.Unit,
		&item.MinQuantity, &item.ExpiryDate, &item.Price, &item.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("item not found: %w", err)
	}

	return &item, nil
}

// Create создаёт новый товар
func (r *InventoryRepository) Create(ctx context.Context, item *domain.InventoryItem) error {
	query := `
		INSERT INTO inventory (name, category, quantity, unit, min_quantity, expiry_date, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		item.Name, item.Category, item.Quantity, item.Unit,
		item.MinQuantity, item.ExpiryDate, item.Price,
	).Scan(&item.ID, &item.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory item: %w", err)
	}

	return nil
}

// Update обновляет товар
func (r *InventoryRepository) Update(ctx context.Context, item *domain.InventoryItem) error {
	query := `
		UPDATE inventory
		SET name = $1, category = $2, quantity = $3, unit = $4,
		    min_quantity = $5, expiry_date = $6, price = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(ctx, query,
		item.Name, item.Category, item.Quantity, item.Unit,
		item.MinQuantity, item.ExpiryDate, item.Price, item.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update inventory item: %w", err)
	}

	return nil
}

// UpdateQuantity обновляет количество товара
func (r *InventoryRepository) UpdateQuantity(ctx context.Context, id string, quantity float64) error {
	query := `UPDATE inventory SET quantity = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, quantity, id)
	return err
}

// CreateMovement создаёт запись о движении товара
func (r *InventoryRepository) CreateMovement(ctx context.Context, m *domain.InventoryMovement) error {
	query := `
		INSERT INTO inventory_movements (inventory_id, type, quantity, user_id, reason)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(ctx, query,
		m.InventoryID, m.Type, m.Quantity, m.UserID, m.Reason,
	).Scan(&m.ID, &m.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create movement: %w", err)
	}

	return nil
}

// GetMovements возвращает историю движений товара
func (r *InventoryRepository) GetMovements(ctx context.Context, inventoryID string, limit int) ([]domain.InventoryMovement, error) {
	query := `
		SELECT 
			m.id, m.inventory_id, m.type, m.quantity, m.user_id, m.reason, m.created_at,
			u.full_name as user_name,
			i.name as item_name
		FROM inventory_movements m
		JOIN users u ON m.user_id = u.id
		JOIN inventory i ON m.inventory_id = i.id
		WHERE m.inventory_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, inventoryID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movements []domain.InventoryMovement
	for rows.Next() {
		var m domain.InventoryMovement
		var user domain.User
		var item domain.InventoryItem

		err := rows.Scan(
			&m.ID, &m.InventoryID, &m.Type, &m.Quantity, &m.UserID, &m.Reason, &m.CreatedAt,
			&user.FullName,
			&item.Name,
		)
		if err != nil {
			return nil, err
		}

		user.ID = m.UserID
		item.ID = m.InventoryID
		m.User = &user
		m.Inventory = &item

		movements = append(movements, m)
	}

	return movements, nil
}

// GetLowStock возвращает товары с остатком ниже минимума
func (r *InventoryRepository) GetLowStock(ctx context.Context) ([]domain.LowStockAlert, error) {
	query := `
		SELECT id, name, category, quantity, min_quantity, unit
		FROM inventory
		WHERE quantity <= min_quantity
		ORDER BY (quantity / NULLIF(min_quantity, 0)) ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []domain.LowStockAlert
	for rows.Next() {
		var a domain.LowStockAlert
		err := rows.Scan(&a.InventoryID, &a.Name, &a.Category, &a.Quantity, &a.MinQuantity, &a.Unit)
		if err != nil {
			return nil, err
		}
		a.Deficit = a.MinQuantity - a.Quantity
		alerts = append(alerts, a)
	}

	return alerts, nil
}

// GetExpiring возвращает товары с истекающим сроком годности
func (r *InventoryRepository) GetExpiring(ctx context.Context, withinDays int) ([]domain.ExpiringAlert, error) {
	query := `
		SELECT id, name, category, quantity, unit, expiry_date
		FROM inventory
		WHERE expiry_date IS NOT NULL
		  AND expiry_date <= CURRENT_DATE + $1 * INTERVAL '1 day'
		  AND expiry_date >= CURRENT_DATE
		ORDER BY expiry_date ASC
	`

	rows, err := r.db.Query(ctx, query, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var alerts []domain.ExpiringAlert
	for rows.Next() {
		var a domain.ExpiringAlert
		err := rows.Scan(&a.InventoryID, &a.Name, &a.Category, &a.Quantity, &a.Unit, &a.ExpiryDate)
		if err != nil {
			return nil, err
		}
		a.DaysLeft = int(a.ExpiryDate.Sub(today).Hours() / 24)
		alerts = append(alerts, a)
	}

	return alerts, nil
}

// GetExpired возвращает просроченные товары
func (r *InventoryRepository) GetExpired(ctx context.Context) ([]domain.ExpiringAlert, error) {
	query := `
		SELECT id, name, category, quantity, unit, expiry_date
		FROM inventory
		WHERE expiry_date IS NOT NULL
		  AND expiry_date < CURRENT_DATE
		ORDER BY expiry_date ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var alerts []domain.ExpiringAlert
	for rows.Next() {
		var a domain.ExpiringAlert
		err := rows.Scan(&a.InventoryID, &a.Name, &a.Category, &a.Quantity, &a.Unit, &a.ExpiryDate)
		if err != nil {
			return nil, err
		}
		a.DaysLeft = int(a.ExpiryDate.Sub(today).Hours() / 24) // будет отрицательным
		alerts = append(alerts, a)
	}

	return alerts, nil
}
