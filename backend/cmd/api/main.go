package main

import (
	"log"

	"morepark/internal/config"
	"morepark/internal/database"
	"morepark/internal/server"
)

func main() {
	cfg := config.Load()
	log.Printf("📋 Config loaded: port=%s, db=%s", cfg.Port, cfg.DBName)

	db, err := database.NewPool(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	if err := database.MigrateV2(db); err != nil {
		log.Fatalf("❌ V2 Migration failed: %v", err)
	}
	if err := database.MigrateV3(db); err != nil {
		log.Fatalf("❌ V3 Migration failed: %v", err)
	}
	if err := database.MigrateV4(db); err != nil {
		log.Fatalf("❌ V4 Migration failed: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("❌ Seed failed: %v", err)
	}

	srv := server.New(cfg, db)

	log.Printf("🚀 Server starting on http://localhost:%s", cfg.Port)
	log.Printf("📊 Health check: http://localhost:%s/health", cfg.Port)
	log.Printf("🗄️  DB check: http://localhost:%s/health/db", cfg.Port)

	if err := srv.Run(); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
