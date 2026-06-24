package domain

import "time"

type TicketType struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	DurationHours int       `json:"duration_hours"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}
