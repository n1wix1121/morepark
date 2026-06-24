package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func Seed(db *pgxpool.Pool) error {
	log.Println("🌱 Seeding database...")
	ctx := context.Background()

	if err := seedUsers(ctx, db); err != nil {
		return err
	}
	if err := seedZones(ctx, db); err != nil {
		return err
	}
	if err := seedTicketTypes(ctx, db); err != nil {
		return err
	}
	if err := seedEquipment(ctx, db); err != nil {
		return err
	}
	if err := seedInventory(ctx, db); err != nil {
		return err
	}

	log.Println("🎉 Database seed check completed!")
	return nil
}

func tableCount(ctx context.Context, db *pgxpool.Pool, table string) (int, error) {
	var count int
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count)
	return count, err
}

func seedUsers(ctx context.Context, db *pgxpool.Pool) error {
	count, err := tableCount(ctx, db, "users")
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("⚠️  Users already exist, skipping")
		return nil
	}

	password := "test123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	users := []struct {
		email string
		name  string
		role  string
	}{
		{"director@morepark.ru", "Крылова Алла Владимировна", "director"},
		{"cashier@morepark.ru", "Иванова Мария Петровна", "cashier"},
		{"lifeguard@morepark.ru", "Петров Алексей Сергеевич", "lifeguard"},
		{"technician@morepark.ru", "Сидоров Дмитрий Иванович", "technician"},
		{"barman@morepark.ru", "Козлова Анна Михайловна", "barman"},
	}

	for _, u := range users {
		_, err := db.Exec(ctx, `
			INSERT INTO users (email, password_hash, full_name, role, is_active)
			VALUES ($1, $2, $3, $4, true)
		`, u.email, string(hashedPassword), u.name, u.role)
		if err != nil {
			return err
		}
	}
	log.Println("✅ Users created (5)")
	log.Println("📝 Test password for all users: test123")
	return nil
}

func seedZones(ctx context.Context, db *pgxpool.Pool) error {
	count, err := tableCount(ctx, db, "zones")
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("⚠️  Zones already exist, skipping")
		return nil
	}

	zones := []struct {
		name        string
		capacity    int
		description string
	}{
		{"Волновой бассейн", 60, "Главный бассейн с волнами"},
		{"Детская зона", 30, "Для детей до 12 лет"},
		{"Джакузи", 15, "Горячие ванны с гидромассажем"},
		{"Горка Цунами", 20, "Экстремальная горка высотой 15 метров"},
		{"Ленивая река", 40, "Спокойный поток для отдыха"},
	}

	for _, z := range zones {
		_, err := db.Exec(ctx, `
			INSERT INTO zones (name, capacity, current_count, description, is_active)
			VALUES ($1, $2, 0, $3, true)
		`, z.name, z.capacity, z.description)
		if err != nil {
			return err
		}
	}
	log.Println("✅ Zones created (5)")
	return nil
}

func seedTicketTypes(ctx context.Context, db *pgxpool.Pool) error {
	count, err := tableCount(ctx, db, "ticket_types")
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("⚠️  Ticket types already exist, skipping")
		return nil
	}

	ticketTypes := []struct {
		typeName      string
		name          string
		price         float64
		durationHours int
		description   string
	}{
		{"single", "Разовый билет", 1500.00, 4, "Посещение одной зоны на 4 часа (1 человек)"},
		{"subscription", "Абонемент на день", 2500.00, 12, "Безлимитное посещение всех зон (1 человек)"},
		{"group", "Групповой (5 чел)", 6000.00, 4, "Для компании до 5 человек — засчитывается 5 мест"},
	}

	for _, t := range ticketTypes {
		_, err := db.Exec(ctx, `
			INSERT INTO ticket_types (type, name, price, duration_hours, description)
			VALUES ($1, $2, $3, $4, $5)
		`, t.typeName, t.name, t.price, t.durationHours, t.description)
		if err != nil {
			return err
		}
	}
	log.Println("✅ Ticket types created (3)")
	return nil
}

func seedEquipment(ctx context.Context, db *pgxpool.Pool) error {
	count, err := tableCount(ctx, db, "equipment")
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("⚠️  Equipment already exists, skipping")
		return nil
	}

	equipment := []struct {
		name         string
		serialNumber string
		zoneName     string
		status       string
	}{
		{"Горка Цунами", "TSN-001", "Горка Цунами", "working"},
		{"Волновой генератор", "WVG-002", "Волновой бассейн", "working"},
		{"Фильтр основной", "FLT-003", "", "working"},
		{"Система хлорирования", "CHL-004", "", "working"},
		{"Детская горка", "KDS-005", "Детская зона", "working"},
	}

	for _, e := range equipment {
		var zoneID *string
		if e.zoneName != "" {
			err := db.QueryRow(ctx, "SELECT id FROM zones WHERE name = $1", e.zoneName).Scan(&zoneID)
			if err != nil {
				log.Printf("⚠️  Zone %s not found for equipment %s", e.zoneName, e.name)
				continue
			}
		}

		_, err := db.Exec(ctx, `
			INSERT INTO equipment (name, serial_number, zone_id, status)
			VALUES ($1, $2, $3, $4)
		`, e.name, e.serialNumber, zoneID, e.status)
		if err != nil {
			return err
		}
	}
	log.Println("✅ Equipment created (5)")
	return nil
}

func seedInventory(ctx context.Context, db *pgxpool.Pool) error {
	count, err := tableCount(ctx, db, "inventory")
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("⚠️  Inventory already exists, skipping")
		return nil
	}

	inventory := []struct {
		name        string
		category    string
		quantity    float64
		unit        string
		minQuantity float64
		price       float64
	}{
		{"Хлор жидкий", "chemical", 50.0, "литр", 10.0, 150.0},
		{"pH-минус", "chemical", 30.0, "кг", 5.0, 300.0},
		{"Тест-полоски", "chemical", 100.0, "шт", 20.0, 50.0},
		{"Кола", "drinks", 48.0, "шт", 12.0, 80.0},
		{"Вода минеральная", "drinks", 60.0, "шт", 15.0, 60.0},
		{"Сок апельсиновый", "drinks", 36.0, "шт", 10.0, 90.0},
		{"Стаканчики", "supplies", 500.0, "шт", 100.0, 5.0},
		{"Салфетки", "supplies", 200.0, "шт", 50.0, 10.0},
	}

	for _, i := range inventory {
		_, err := db.Exec(ctx, `
			INSERT INTO inventory (name, category, quantity, unit, min_quantity, price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, i.name, i.category, i.quantity, i.unit, i.minQuantity, i.price)
		if err != nil {
			return err
		}
	}
	log.Println("✅ Inventory created (8)")
	return nil
}
