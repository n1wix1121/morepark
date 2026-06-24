package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MigrateV3(db *pgxpool.Pool) error {
	log.Println("🗄️  Running V3 migrations...")

	// 1. Добавляем is_refund в таблицу sales
	_, err := db.Exec(context.Background(), `
		ALTER TABLE sales 
		ADD COLUMN IF NOT EXISTS is_refund BOOLEAN DEFAULT false
	`)
	if err != nil {
		return err
	}
	log.Println("✅ Column 'is_refund' added to 'sales'")

	// 2. Добавляем created_at в ticket_types (на всякий случай)
	_, err = db.Exec(context.Background(), `
		ALTER TABLE ticket_types 
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW()
	`)
	if err != nil {
		return err
	}
	log.Println("✅ Column 'created_at' added to 'ticket_types'")

	log.Println("🎉 V3 migrations completed!")
	return nil
}
