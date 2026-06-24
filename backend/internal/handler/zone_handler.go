package handler

import (
	"net/http"

	"morepark/internal/domain"
	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type ZoneHandler struct {
	zoneService *service.ZoneService
}

func NewZoneHandler(zoneService *service.ZoneService) *ZoneHandler {
	return &ZoneHandler{zoneService: zoneService}
}

// GetAll возвращает список всех зон
func (h *ZoneHandler) GetAll(c *gin.Context) {
	zones, err := h.zoneService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zones": zones})
}

// GetByID возвращает зону по ID
func (h *ZoneHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	zone, err := h.zoneService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Zone not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zone": zone})
}

// CreateRequest — запрос на создание зоны
type CreateZoneRequest struct {
	Name        string `json:"name" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	Description string `json:"description"`
}

// Create создаёт новую зону (только директор)
func (h *ZoneHandler) Create(c *gin.Context) {
	var req CreateZoneRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone := &domain.Zone{
		Name:        req.Name,
		Capacity:    req.Capacity,
		Description: req.Description,
	}

	if err := h.zoneService.Create(c.Request.Context(), zone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"zone": zone})
}

// UpdateRequest — запрос на обновление зоны
type UpdateZoneRequest struct {
	Name        string `json:"name" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

// Update обновляет зону (только директор)
// Update обновляет зону
func (h *ZoneHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone, err := h.zoneService.Update(c.Request.Context(), id, service.UpdateZoneRequest{
		Name:        req.Name,
		Description: req.Description,
		Capacity:    req.Capacity,
	})

	if err != nil {
		if err == service.ErrZoneNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"zone": zone})
}

// Delete удаляет зону (только директор)
func (h *ZoneHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.zoneService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Zone deleted successfully"})
}
