package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MigrateV4(db *pgxpool.Pool) error {
	log.Println("🗄️  Running V4 migrations...")

	queries := []struct {
		name string
		sql  string
	}{
		{
			name: "incidents.resolved_by",
			sql:  `ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolved_by UUID REFERENCES users(id)`,
		},
		{
			name: "incidents.resolved_at",
			sql:  `ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolved_at TIMESTAMP`,
		},
		{
			name: "incidents.resolution",
			sql:  `ALTER TABLE incidents ADD COLUMN IF NOT EXISTS resolution TEXT`,
		},
		{
			name: "equipment.created_at",
			sql:  `ALTER TABLE equipment ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW()`,
		},
		{
			name: "inventory.created_at",
			sql:  `ALTER TABLE inventory ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW()`,
		},
		{
			name: "inventory_movements.reason",
			sql:  `ALTER TABLE inventory_movements ADD COLUMN IF NOT EXISTS reason TEXT DEFAULT ''`,
		},
		{
			name: "zones.reactivate_hidden",
			sql:  `UPDATE zones SET is_active = true WHERE is_active = false`,
		},
	}

	for _, q := range queries {
		if _, err := db.Exec(context.Background(), q.sql); err != nil {
			return err
		}
		log.Printf("✅ %s", q.name)
	}

	log.Println("🎉 V4 migrations completed!")
	return nil
}
