package domain

import "time"

type Incident struct {
	ID          string     `json:"id"`
	ZoneID      string     `json:"zone_id"`
	LifeguardID string     `json:"lifeguard_id"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"`              // low, medium, high
	Status      string     `json:"status"`                // open, in_progress, closed
	ResolvedBy  *string    `json:"resolved_by,omitempty"` // кто закрыл
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"` // когда закрыли
	Resolution  *string    `json:"resolution,omitempty"`  // описание решения
	CreatedAt   time.Time  `json:"created_at"`

	// Вложенные данные
	Zone      *Zone `json:"zone,omitempty"`
	Lifeguard *User `json:"lifeguard,omitempty"`
}
