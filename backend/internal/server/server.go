package server

import (
	"morepark/internal/config"
	"morepark/internal/handler"
	"morepark/internal/middleware"
	"morepark/internal/repository"
	"morepark/internal/service"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	cfg    *config.Config
	db     *pgxpool.Pool
}

func New(cfg *config.Config, db *pgxpool.Pool) *Server {
	router := gin.Default()

	// CORS
	allowedOrigins := strings.Split(cfg.AllowedOrigins, ",")
	router.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// ==================== РЕПОЗИТОРИИ ====================
	userRepo := repository.NewUserRepository(db)
	zoneRepo := repository.NewZoneRepository(db)
	ticketTypeRepo := repository.NewTicketTypeRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	saleRepo := repository.NewSaleRepository(db)
	waterRepo := repository.NewWaterRepository(db)
	equipmentRepo := repository.NewEquipmentRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	incidentRepo := repository.NewIncidentRepository(db)

	// ==================== СЕРВИСЫ ====================
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	zoneService := service.NewZoneService(zoneRepo)
	ticketService := service.NewTicketService(ticketTypeRepo, ticketRepo, saleRepo, zoneRepo, db)
	waterService := service.NewWaterService(waterRepo, zoneRepo)
	equipmentService := service.NewEquipmentService(equipmentRepo)
	inventoryService := service.NewInventoryService(inventoryRepo)
	incidentService := service.NewIncidentService(incidentRepo, zoneRepo)
	userService := service.NewUserService(userRepo)
	reportService := service.NewReportService(repository.NewReportRepository(db))

	// 🆕 НОВЫЙ СЕРВИС для публичной части
	publicTicketService := service.NewPublicTicketService(
		ticketTypeRepo, ticketRepo, saleRepo, zoneRepo, db,
	)

	// ==================== ХЕНДЛЕРЫ ====================
	authHandler := handler.NewAuthHandler(authService)
	zoneHandler := handler.NewZoneHandler(zoneService)
	ticketHandler := handler.NewTicketHandler(ticketService)
	waterHandler := handler.NewWaterHandler(waterService)
	equipmentHandler := handler.NewEquipmentHandler(equipmentService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	incidentHandler := handler.NewIncidentHandler(incidentService)
	userHandler := handler.NewUserHandler(userService)
	reportHandler := handler.NewReportHandler(reportService)

	// 🆕 НОВЫЙ ХЕНДЛЕР для публичной части
	publicHandler := handler.NewPublicHandler(publicTicketService)

	// ==================== HEALTH CHECK ====================
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "morepark"})
	})

	router.GET("/health/db", func(c *gin.Context) {
		err := db.Ping(c.Request.Context())
		if err != nil {
			c.JSON(500, gin.H{"status": "error", "message": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "database": "connected"})
	})

	// ==================== API v1 ====================
	api := router.Group("/api/v1")
	{
		// 🌐 ПУБЛИЧНЫЕ МАРШРУТЫ (без авторизации!)
		public := api.Group("/public")
		{
			public.GET("/zones", publicHandler.GetZones)
			public.GET("/ticket-types", publicHandler.GetTicketTypes)
			public.POST("/check-availability", publicHandler.CheckAvailability)
			public.POST("/tickets", publicHandler.PurchaseTicket)
			public.GET("/tickets/:number", publicHandler.GetTicketStatus)
		}

		// 🔐 Публичный вход
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// 🔒 ЗАЩИЩЁННЫЕ МАРШРУТЫ (требуют JWT)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.GET("/auth/me", authHandler.Me)

			// 🏊 Зоны
			zones := protected.Group("/zones")
			{
				zones.GET("", zoneHandler.GetAll)
				zones.GET("/:id", zoneHandler.GetByID)
				zones.POST("", middleware.RequireRoles("director"), zoneHandler.Create)
				zones.PUT("/:id", middleware.RequireRoles("director"), zoneHandler.Update)
				zones.DELETE("/:id", middleware.RequireRoles("director"), zoneHandler.Delete)
			}

			// 🎫 Типы билетов
			protected.GET("/ticket-types", ticketHandler.GetTicketTypes)

			// 🎫 Билеты и продажи
			tickets := protected.Group("/tickets")
			{
				tickets.GET("", ticketHandler.GetSales)
				tickets.GET("/:id", ticketHandler.GetTicket)
				tickets.POST("/sell",
					middleware.RequireRoles("cashier", "director"),
					ticketHandler.SellTicket,
				)
				tickets.POST("/:id/refund",
					middleware.RequireRoles("cashier", "director"),
					ticketHandler.RefundTicket,
				)
			}

			// 💧 Водоподготовка
			water := protected.Group("/water")
			{
				water.GET("/measurements", waterHandler.GetMeasurements)
				water.GET("/alerts", waterHandler.GetAlerts)
				water.GET("/zone/:id", waterHandler.GetMeasurementsByZone)
				water.POST("/measurements",
					middleware.RequireRoles("technician", "director"),
					waterHandler.CreateMeasurement,
				)
			}

			// 🔧 Оборудование и ТО
			equipment := protected.Group("/equipment")
			{
				equipment.GET("", equipmentHandler.GetAll)
				equipment.GET("/:id", equipmentHandler.GetByID)
				equipment.GET("/:id/maintenance", equipmentHandler.GetMaintenanceLogs)
				equipment.POST("",
					middleware.RequireRoles("director"),
					equipmentHandler.Create,
				)
				equipment.PUT("/:id",
					middleware.RequireRoles("director"),
					equipmentHandler.Update,
				)
				equipment.POST("/:id/maintenance",
					middleware.RequireRoles("technician", "director"),
					equipmentHandler.CompleteMaintenance,
				)
			}

			// 🔔 Напоминания о ТО
			protected.GET("/maintenance/upcoming", equipmentHandler.GetUpcomingMaintenance)

			// 📦 Склад ТМЦ
			inventory := protected.Group("/inventory")
			{
				inventory.GET("", inventoryHandler.GetAll)
				inventory.GET("/low-stock", inventoryHandler.GetLowStock)
				inventory.GET("/expiring", inventoryHandler.GetExpiring)
				inventory.GET("/expired", inventoryHandler.GetExpired)
				inventory.GET("/:id", inventoryHandler.GetByID)
				inventory.GET("/:id/movements", inventoryHandler.GetMovements)
				inventory.POST("",
					middleware.RequireRoles("director"),
					inventoryHandler.Create,
				)
				inventory.PUT("/:id",
					middleware.RequireRoles("director"),
					inventoryHandler.Update,
				)
				inventory.POST("/:id/move",
					middleware.RequireRoles("director", "barman"),
					inventoryHandler.Move,
				)
			}

			// 🚨 Инциденты
			incidents := protected.Group("/incidents")
			{
				incidents.GET("", incidentHandler.GetAll)
				incidents.GET("/active", incidentHandler.GetActive)
				incidents.GET("/:id", incidentHandler.GetByID)
				incidents.GET("/zone/:id", incidentHandler.GetByZone)
				incidents.POST("",
					middleware.RequireRoles("lifeguard", "director"),
					incidentHandler.Create,
				)
				incidents.PATCH("/:id/status",
					middleware.RequireRoles("lifeguard", "director"),
					incidentHandler.UpdateStatus,
				)
			}

			// 👥 Пользователи (только директор)
			users := protected.Group("/users")
			users.Use(middleware.RequireRoles("director"))
			{
				users.GET("", userHandler.GetAll)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			// 📊 Отчёты для бухгалтерии (Excel)
			reports := protected.Group("/reports")
			{
				reports.GET("/sales/excel",
					middleware.RequireRoles("director", "cashier"),
					reportHandler.ExportSales,
				)
				reports.GET("/inventory/excel",
					middleware.RequireRoles("director", "barman"),
					reportHandler.ExportInventory,
				)
				reports.GET("/summary/excel",
					middleware.RequireRoles("director"),
					reportHandler.ExportSummary,
				)
			}
		}
	}

	return &Server{router: router, cfg: cfg, db: db}
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.cfg.Port)
}
