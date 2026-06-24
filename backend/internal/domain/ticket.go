package domain

import "time"

type Ticket struct {
	ID            string    `json:"id"`
	TicketTypeID  string    `json:"ticket_type_id"`
	ZoneID        string    `json:"zone_id"`
	Status        string    `json:"status"`
	CustomerName  string    `json:"customer_name"`
	CustomerPhone string    `json:"customer_phone"`
	CustomerEmail string    `json:"customer_email,omitempty"` // НОВОЕ
	TicketNumber  string    `json:"ticket_number"`            // НОВОЕ
	Source        string    `json:"source"`                   // НОВОЕ: cashier / online
	Quantity      int       `json:"quantity"`                 // НОВОЕ
	QRCode        string    `json:"qr_code,omitempty"`        // НОВОЕ: base64
	ValidFrom     time.Time `json:"valid_from"`
	ValidUntil    time.Time `json:"valid_until"`
	CreatedAt     time.Time `json:"created_at"`

	// Вложенные данные
	TicketType *TicketType `json:"ticket_type,omitempty"`
	Zone       *Zone       `json:"zone,omitempty"`
}
