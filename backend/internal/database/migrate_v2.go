package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MigrateV2(db *pgxpool.Pool) error {
	log.Println("🗄️  Running V2 migrations...")

	// Добавляем поля в tickets для онлайн-бронирования
	_, err := db.Exec(context.Background(), `
		ALTER TABLE tickets 
		ADD COLUMN IF NOT EXISTS ticket_number VARCHAR(50) UNIQUE,
		ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255),
		ADD COLUMN IF NOT EXISTS source VARCHAR(20) DEFAULT 'cashier' CHECK (source IN ('cashier', 'online')),
		ADD COLUMN IF NOT EXISTS quantity INT DEFAULT 1,
		ADD COLUMN IF NOT EXISTS qr_code TEXT
	`)
	if err != nil {
		return err
	}
	log.Println("✅ Table 'tickets' extended with online booking fields")

	// Заполняем ticket_number для существующих записей
	_, err = db.Exec(context.Background(), `
		UPDATE tickets 
		SET ticket_number = 'MP-' || LPAD(id::text, 6, '0')
		WHERE ticket_number IS NULL
	`)
	if err != nil {
		return err
	}
	log.Println("✅ Existing tickets assigned numbers")

	log.Println("🎉 V2 migrations completed!")
	return nil
}
