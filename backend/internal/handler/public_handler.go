package handler

import (
	"net/http"
	"time"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	publicService *service.PublicTicketService
}

func NewPublicHandler(publicService *service.PublicTicketService) *PublicHandler {
	return &PublicHandler{publicService: publicService}
}

// GetZones возвращает зоны с загрузкой (публичный)
func (h *PublicHandler) GetZones(c *gin.Context) {
	zones, err := h.publicService.PublicZones(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Добавляем флаг доступности
	type ZoneResponse struct {
		ID           string  `json:"id"`
		Name         string  `json:"name"`
		Capacity     int     `json:"capacity"`
		CurrentCount int     `json:"current_count"`
		Description  string  `json:"description"`
		IsAvailable  bool    `json:"is_available"`
		LoadPercent  float64 `json:"load_percent"`
	}

	response := make([]ZoneResponse, len(zones))
	for i, z := range zones {
		loadPercent := 0.0
		if z.Capacity > 0 {
			loadPercent = float64(z.CurrentCount) / float64(z.Capacity) * 100
		}
		response[i] = ZoneResponse{
			ID:           z.ID,
			Name:         z.Name,
			Capacity:     z.Capacity,
			CurrentCount: z.CurrentCount,
			Description:  z.Description,
			IsAvailable:  z.CurrentCount < z.Capacity,
			LoadPercent:  loadPercent,
		}
	}

	c.JSON(http.StatusOK, gin.H{"zones": response})
}

// GetTicketTypes возвращает типы билетов (публичный)
func (h *PublicHandler) GetTicketTypes(c *gin.Context) {
	types, err := h.publicService.PublicTicketTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket_types": types})
}

// CheckAvailabilityRequest — запрос на проверку
type CheckAvailabilityRequest struct {
	ZoneID   string `json:"zone_id" binding:"required"`
	DateTime string `json:"datetime" binding:"required"` // ISO формат
	Quantity int    `json:"quantity" binding:"required,min=1,max=10"`
}

// CheckAvailability проверяет доступность
func (h *PublicHandler) CheckAvailability(c *gin.Context) {
	var req CheckAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dt, err := time.Parse(time.RFC3339, req.DateTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты. Используйте ISO 8601"})
		return
	}

	resp, err := h.publicService.CheckAvailability(c.Request.Context(), service.CheckAvailabilityRequest{
		ZoneID:   req.ZoneID,
		DateTime: dt,
		Quantity: req.Quantity,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// PurchaseRequest — запрос на покупку
type PurchaseRequest struct {
	TicketTypeID  string `json:"ticket_type_id" binding:"required"`
	ZoneID        string `json:"zone_id" binding:"required"`
	DateTime      string `json:"datetime" binding:"required"`
	Quantity      int    `json:"quantity" binding:"required,min=1,max=10"`
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	CustomerEmail string `json:"customer_email" binding:"required,email"`
}

// PurchaseTicket покупает билет онлайн
func (h *PublicHandler) PurchaseTicket(c *gin.Context) {
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dt, err := time.Parse(time.RFC3339, req.DateTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты"})
		return
	}

	resp, err := h.publicService.PurchaseTicket(c.Request.Context(), service.PurchaseRequest{
		TicketTypeID:  req.TicketTypeID,
		ZoneID:        req.ZoneID,
		DateTime:      dt,
		Quantity:      req.Quantity,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		CustomerEmail: req.CustomerEmail,
	})

	if err != nil {
		if err == service.ErrNotEnoughSlots {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Недостаточно мест на выбранное время",
				"code":  "NOT_ENOUGH_SLOTS",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetTicketStatus возвращает статус билета по номеру
func (h *PublicHandler) GetTicketStatus(c *gin.Context) {
	number := c.Param("number")

	ticket, err := h.publicService.GetTicketByNumber(c.Request.Context(), number)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Билет не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticket_number": ticket.TicketNumber,
		"status":        ticket.Status,
		"zone_name":     ticket.Zone.Name,
		"valid_from":    ticket.ValidFrom,
		"valid_until":   ticket.ValidUntil,
		"quantity":      ticket.Quantity,
	})
}
