package handler

import (
	"net/http"
	"strconv"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	inventoryService *service.InventoryService
}

func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// GetAll возвращает все товары
func (h *InventoryHandler) GetAll(c *gin.Context) {
	items, err := h.inventoryService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"inventory": items})
}

// GetByID возвращает товар по ID
func (h *InventoryHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	item, err := h.inventoryService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Товар не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

// CreateRequest — запрос на создание
type CreateInventoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required,oneof=chemical drinks food supplies"`
	Quantity    float64 `json:"quantity" binding:"required,min=0"`
	Unit        string  `json:"unit" binding:"required"`
	MinQuantity float64 `json:"min_quantity" binding:"required,min=0"`
	ExpiryDate  *string `json:"expiry_date"`
	Price       float64 `json:"price" binding:"required,min=0"`
}

// Create создаёт товар (только директор)
func (h *InventoryHandler) Create(c *gin.Context) {
	var req CreateInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.inventoryService.Create(c.Request.Context(), service.CreateInventoryRequest{
		Name:        req.Name,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinQuantity: req.MinQuantity,
		ExpiryDate:  req.ExpiryDate,
		Price:       req.Price,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": item})
}

// UpdateRequest — запрос на обновление
type UpdateInventoryRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required,oneof=chemical drinks food supplies"`
	Quantity    float64 `json:"quantity" binding:"required,min=0"`
	Unit        string  `json:"unit" binding:"required"`
	MinQuantity float64 `json:"min_quantity" binding:"required,min=0"`
	ExpiryDate  *string `json:"expiry_date"`
	Price       float64 `json:"price" binding:"required,min=0"`
}

// Update обновляет товар (только директор)
func (h *InventoryHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.inventoryService.Update(c.Request.Context(), id, service.UpdateInventoryRequest{
		Name:        req.Name,
		Category:    req.Category,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		MinQuantity: req.MinQuantity,
		ExpiryDate:  req.ExpiryDate,
		Price:       req.Price,
	})

	if err != nil {
		if err == service.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": item})
}

// MoveRequest — запрос на движение
type MoveRequest struct {
	Type     string  `json:"type" binding:"required,oneof=in out"`
	Quantity float64 `json:"quantity" binding:"required,gt=0"`
	Reason   string  `json:"reason" binding:"required"`
}

// Move выполняет приход/расход
func (h *InventoryHandler) Move(c *gin.Context) {
	id := c.Param("id")

	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")

	movement, err := h.inventoryService.Move(c.Request.Context(), id, service.MoveRequest{
		Type:     req.Type,
		Quantity: req.Quantity,
		Reason:   req.Reason,
		UserID:   userID.(string),
	})

	if err != nil {
		switch err {
		case service.ErrItemNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case service.ErrInsufficientQty:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Недостаточно товара на складе"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Движение успешно зарегистрировано",
		"movement": movement,
	})
}

// GetMovements возвращает историю движений
func (h *InventoryHandler) GetMovements(c *gin.Context) {
	id := c.Param("id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	movements, err := h.inventoryService.GetMovements(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movements": movements})
}

// GetLowStock возвращает товары с низким остатком
func (h *InventoryHandler) GetLowStock(c *gin.Context) {
	alerts, err := h.inventoryService.GetLowStock(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"low_stock_alerts": alerts,
		"count":            len(alerts),
	})
}

// GetExpiring возвращает товары с истекающим сроком
func (h *InventoryHandler) GetExpiring(c *gin.Context) {
	withinDays := 30
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			withinDays = parsed
		}
	}

	alerts, err := h.inventoryService.GetExpiring(c.Request.Context(), withinDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expiring_alerts": alerts,
		"within_days":     withinDays,
		"count":           len(alerts),
	})
}

// GetExpired возвращает просроченные товары
func (h *InventoryHandler) GetExpired(c *gin.Context) {
	alerts, err := h.inventoryService.GetExpired(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"expired_alerts": alerts,
		"count":          len(alerts),
	})
}
