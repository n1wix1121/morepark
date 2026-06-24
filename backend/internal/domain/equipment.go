package domain

import "time"

type Equipment struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	SerialNumber    string     `json:"serial_number"`
	ZoneID          *string    `json:"zone_id"`          // может быть без зоны (фильтры и т.д.)
	Status          string     `json:"status"`           // working, maintenance, broken
	LastMaintenance *time.Time `json:"last_maintenance"` // может быть null
	NextMaintenance *time.Time `json:"next_maintenance"` // может быть null
	CreatedAt       time.Time  `json:"created_at"`

	// Вложенные данные
	Zone *Zone `json:"zone,omitempty"`
}

type MaintenanceLog struct {
	ID           string    `json:"id"`
	EquipmentID  string    `json:"equipment_id"`
	TechnicianID string    `json:"technician_id"`
	Description  string    `json:"description"`
	CompletedAt  time.Time `json:"completed_at"`

	// Вложенные данные
	Equipment  *Equipment `json:"equipment,omitempty"`
	Technician *User      `json:"technician,omitempty"`
}

// UpcomingMaintenance — информация о предстоящем ТО
type UpcomingMaintenance struct {
	EquipmentID     string    `json:"equipment_id"`
	EquipmentName   string    `json:"equipment_name"`
	SerialNumber    string    `json:"serial_number"`
	NextMaintenance time.Time `json:"next_maintenance"`
	DaysUntil       int       `json:"days_until"`    // дней до ТО
	UrgencyLevel    string    `json:"urgency_level"` // low (7 дней), medium (3 дня), high (1 день), critical (сегодня)
}
