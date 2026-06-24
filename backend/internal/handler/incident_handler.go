package handler

import (
	"net/http"
	"strconv"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type IncidentHandler struct {
	incidentService *service.IncidentService
}

func NewIncidentHandler(incidentService *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{incidentService: incidentService}
}

// CreateRequest — запрос на создание
type CreateIncidentRequest struct {
	ZoneID      string `json:"zone_id" binding:"required"`
	Description string `json:"description" binding:"required,min=10"`
	Severity    string `json:"severity" binding:"required,oneof=low medium high"`
}

// Create создаёт инцидент (спасатель или директор)
func (h *IncidentHandler) Create(c *gin.Context) {
	var req CreateIncidentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Берём ID спасателя из JWT
	lifeguardID, _ := c.Get("user_id")

	incident, err := h.incidentService.Create(c.Request.Context(), service.CreateIncidentRequest{
		ZoneID:      req.ZoneID,
		Description: req.Description,
		Severity:    req.Severity,
		LifeguardID: lifeguardID.(string),
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Инцидент зарегистрирован",
		"incident": incident,
	})
}

// GetAll возвращает все инциденты
func (h *IncidentHandler) GetAll(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	incidents, err := h.incidentService.GetAll(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

// GetActive возвращает только активные инциденты
func (h *IncidentHandler) GetActive(c *gin.Context) {
	incidents, err := h.incidentService.GetActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents": incidents,
		"count":     len(incidents),
	})
}

// GetByID возвращает инцидент по ID
func (h *IncidentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	incident, err := h.incidentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Инцидент не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incident": incident})
}

// GetByZone возвращает инциденты по зоне
func (h *IncidentHandler) GetByZone(c *gin.Context) {
	zoneID := c.Param("id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	incidents, err := h.incidentService.GetByZone(c.Request.Context(), zoneID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"incidents": incidents})
}

// UpdateStatusRequest — запрос на изменение статуса
type UpdateStatusRequest struct {
	Status     string  `json:"status" binding:"required,oneof=open in_progress closed"`
	Resolution *string `json:"resolution"`
}

// UpdateStatus изменяет статус инцидента
func (h *IncidentHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resolvedBy, _ := c.Get("user_id")

	incident, err := h.incidentService.UpdateStatus(c.Request.Context(), id, service.UpdateStatusRequest{
		Status:     req.Status,
		Resolution: req.Resolution,
		ResolvedBy: resolvedBy.(string),
	})

	if err != nil {
		switch err {
		case service.ErrIncidentNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case service.ErrAlreadyClosed:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Статус инцидента обновлён",
		"incident": incident,
	})
}
