package domain

import "time"

// InventoryItem — товар на складе
type InventoryItem struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Category    string     `json:"category"` // chemical, drinks, food, supplies
	Quantity    float64    `json:"quantity"`
	Unit        string     `json:"unit"`         // литр, кг, шт
	MinQuantity float64    `json:"min_quantity"` // минимальный остаток для авто-заявки
	ExpiryDate  *time.Time `json:"expiry_date"`  // может быть null
	Price       float64    `json:"price"`
	CreatedAt   time.Time  `json:"created_at"`

	// Вычисляемые поля
	IsExpiringSoon bool `json:"is_expiring_soon"` // истекает в ближайшие 30 дней
	IsExpired      bool `json:"is_expired"`       // просрочено
	IsLowStock     bool `json:"is_low_stock"`     // ниже минимума
}

// InventoryMovement — движение товара (приход/расход)
type InventoryMovement struct {
	ID          string    `json:"id"`
	InventoryID string    `json:"inventory_id"`
	Type        string    `json:"type"` // in (приход), out (расход)
	Quantity    float64   `json:"quantity"`
	UserID      string    `json:"user_id"`
	Reason      string    `json:"reason"` // причина (закупка, списание, продажа)
	CreatedAt   time.Time `json:"created_at"`

	// Вложенные данные
	Inventory *InventoryItem `json:"inventory,omitempty"`
	User      *User          `json:"user,omitempty"`
}

// LowStockAlert — уведомление о низком остатке
type LowStockAlert struct {
	InventoryID string  `json:"inventory_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Quantity    float64 `json:"quantity"`
	MinQuantity float64 `json:"min_quantity"`
	Unit        string  `json:"unit"`
	Deficit     float64 `json:"deficit"` // сколько нужно до минимума
}

// ExpiringAlert — уведомление о скором истечении срока
type ExpiringAlert struct {
	InventoryID string    `json:"inventory_id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Quantity    float64   `json:"quantity"`
	Unit        string    `json:"unit"`
	ExpiryDate  time.Time `json:"expiry_date"`
	DaysLeft    int       `json:"days_left"` // дней до истечения
}
