package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Migrate(db *pgxpool.Pool) error {
	log.Println("🗄️  Running database migrations...")

	// 1. Таблица пользователей
	_, err := db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            email VARCHAR(255) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            full_name VARCHAR(255) NOT NULL,
            role VARCHAR(50) NOT NULL CHECK (role IN ('director', 'cashier', 'lifeguard', 'technician', 'barman')),
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}
	log.Println("✅ Table 'users' created")

	// 2. Таблица зон
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS zones (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            capacity INT NOT NULL,
            current_count INT DEFAULT 0,
            description TEXT,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create zones table: %w", err)
	}
	log.Println("✅ Table 'zones' created")

	// 3. Таблица типов билетов
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS ticket_types (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            type VARCHAR(50) NOT NULL,
            name VARCHAR(255) NOT NULL,
            price DECIMAL(10,2) NOT NULL,
            duration_hours INT NOT NULL,
            description TEXT
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create ticket_types table: %w", err)
	}
	log.Println("✅ Table 'ticket_types' created")

	// 4. Таблица билетов
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS tickets (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            ticket_type_id UUID REFERENCES ticket_types(id),
            zone_id UUID REFERENCES zones(id),
            status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'used', 'cancelled')),
            customer_name VARCHAR(255),
            customer_phone VARCHAR(20),
            valid_from TIMESTAMP NOT NULL,
            valid_until TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create tickets table: %w", err)
	}
	log.Println("✅ Table 'tickets' created")

	// 5. Таблица продаж
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS sales (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            ticket_id UUID REFERENCES tickets(id),
            cashier_id UUID REFERENCES users(id),
            amount DECIMAL(10,2) NOT NULL,
            payment_method VARCHAR(50),
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create sales table: %w", err)
	}
	log.Println("✅ Table 'sales' created")

	// 6. Таблица качества воды
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS water_quality (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            zone_id UUID REFERENCES zones(id),
            ph DECIMAL(3,2) NOT NULL,
            chlorine DECIMAL(5,2) NOT NULL,
            turbidity DECIMAL(5,2) NOT NULL,
            technician_id UUID REFERENCES users(id),
            measured_at TIMESTAMP DEFAULT NOW(),
            is_normal BOOLEAN DEFAULT true
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create water_quality table: %w", err)
	}
	log.Println("✅ Table 'water_quality' created")

	// 7. Таблица оборудования
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS equipment (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            serial_number VARCHAR(100) UNIQUE,
            zone_id UUID REFERENCES zones(id),
            status VARCHAR(50) DEFAULT 'working' CHECK (status IN ('working', 'maintenance', 'broken')),
            last_maintenance DATE,
            next_maintenance DATE,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create equipment table: %w", err)
	}
	log.Println("✅ Table 'equipment' created")

	// 8. Таблица журналов ТО
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS maintenance_logs (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            equipment_id UUID REFERENCES equipment(id),
            technician_id UUID REFERENCES users(id),
            description TEXT,
            completed_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create maintenance_logs table: %w", err)
	}
	log.Println("✅ Table 'maintenance_logs' created")

	// 9. Таблица склада
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS inventory (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name VARCHAR(255) NOT NULL,
            category VARCHAR(50),
            quantity DECIMAL(10,2) NOT NULL,
            unit VARCHAR(50) NOT NULL,
            min_quantity DECIMAL(10,2) NOT NULL,
            expiry_date DATE,
            price DECIMAL(10,2),
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create inventory table: %w", err)
	}
	log.Println("✅ Table 'inventory' created")

	// 10. Таблица движений товаров
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS inventory_movements (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            inventory_id UUID REFERENCES inventory(id),
            type VARCHAR(20) CHECK (type IN ('in', 'out')),
            quantity DECIMAL(10,2) NOT NULL,
            user_id UUID REFERENCES users(id),
            reason TEXT DEFAULT '',
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create inventory_movements table: %w", err)
	}
	log.Println("✅ Table 'inventory_movements' created")

	// 11. Таблица инцидентов
	_, err = db.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS incidents (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            zone_id UUID REFERENCES zones(id),
            lifeguard_id UUID REFERENCES users(id),
            description TEXT NOT NULL,
            severity VARCHAR(20) CHECK (severity IN ('low', 'medium', 'high')),
            status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'closed')),
            resolved_by UUID REFERENCES users(id),
            resolved_at TIMESTAMP,
            resolution TEXT,
            created_at TIMESTAMP DEFAULT NOW()
        )
    `)
	if err != nil {
		return fmt.Errorf("failed to create incidents table: %w", err)
	}
	log.Println("✅ Table 'incidents' created")

	log.Println("🎉 All migrations completed successfully!")
	return nil
}
