package domain

import "time"

type Sale struct {
	ID            string    `json:"id"`
	TicketID      string    `json:"ticket_id"`
	CashierID     string    `json:"cashier_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"` // cash, card, online
	IsRefund      bool      `json:"is_refund"`
	CreatedAt     time.Time `json:"created_at"`

	// Вложенные данные
	Ticket  *Ticket `json:"ticket,omitempty"`
	Cashier *User   `json:"cashier,omitempty"`
}
