package handler

import (
	"net/http"
	"strconv"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

// GetTicketTypes возвращает список типов билетов
func (h *TicketHandler) GetTicketTypes(c *gin.Context) {
	types, err := h.ticketService.GetTicketTypes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket_types": types})
}

// SellRequest — запрос на продажу
type SellTicketRequest struct {
	TicketTypeID  string `json:"ticket_type_id" binding:"required"`
	ZoneID        string `json:"zone_id" binding:"required"`
	CustomerName  string `json:"customer_name"`
	CustomerPhone string `json:"customer_phone"`
	PaymentMethod string `json:"payment_method" binding:"required,oneof=cash card online"`
}

// SellTicket продаёт билет
func (h *TicketHandler) SellTicket(c *gin.Context) {
	var req SellTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Берём ID кассира из JWT
	cashierID, _ := c.Get("user_id")

	sale, err := h.ticketService.SellTicket(c.Request.Context(), service.SellRequest{
		TicketTypeID:  req.TicketTypeID,
		ZoneID:        req.ZoneID,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		PaymentMethod: req.PaymentMethod,
		CashierID:     cashierID.(string),
	})

	if err != nil {
		if err == service.ErrZoneFull {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Зона заполнена, продажа невозможна",
				"code":  "ZONE_FULL",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Билет успешно продан",
		"sale":    sale,
	})
}

// GetSales возвращает историю продаж
func (h *TicketHandler) GetSales(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	sales, err := h.ticketService.GetSales(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sales": sales})
}

// GetTicket возвращает билет по ID
func (h *TicketHandler) GetTicket(c *gin.Context) {
	id := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Билет не найден"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

// RefundTicket оформляет возврат
func (h *TicketHandler) RefundTicket(c *gin.Context) {
	id := c.Param("id")
	cashierID, _ := c.Get("user_id")

	refund, err := h.ticketService.RefundTicket(c.Request.Context(), id, service.RefundRequest{
		CashierID: cashierID.(string),
	})

	if err != nil {
		switch err {
		case service.ErrTicketNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case service.ErrTicketUsed:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Билет уже использован, возврат невозможен"})
		case service.ErrTicketCancelled:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Билет уже отменён"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Возврат оформлен",
		"refund":  refund,
	})
}
