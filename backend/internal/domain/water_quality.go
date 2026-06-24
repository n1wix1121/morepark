package domain

import "time"

type WaterQuality struct {
	ID           string    `json:"id"`
	ZoneID       string    `json:"zone_id"`
	PH           float64   `json:"ph"`
	Chlorine     float64   `json:"chlorine"`  // хлор, мг/л
	Turbidity    float64   `json:"turbidity"` // мутность, ЕМФ
	TechnicianID string    `json:"technician_id"`
	MeasuredAt   time.Time `json:"measured_at"`
	IsNormal     bool      `json:"is_normal"`
	Violations   []string  `json:"violations,omitempty"` // список нарушений

	// Вложенные данные для удобства
	Zone       *Zone `json:"zone,omitempty"`
	Technician *User `json:"technician,omitempty"`
}
