package domain

import "time"

type Zone struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Capacity     int       `json:"capacity"`
	CurrentCount int       `json:"current_count"`
	Description  string    `json:"description"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}
