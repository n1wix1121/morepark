package handler

import (
	"net/http"
	"strconv"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type EquipmentHandler struct {
	equipmentService *service.EquipmentService
}

func NewEquipmentHandler(equipmentService *service.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{equipmentService: equipmentService}
}

// GetAll возвращает всё оборудование
func (h *EquipmentHandler) GetAll(c *gin.Context) {
	equipment, err := h.equipmentService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"equipment": equipment})
}

// GetByID возвращает оборудование по ID
func (h *EquipmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	e, err := h.equipmentService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Оборудование не найдено"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"equipment": e})
}

// CreateRequest — запрос на создание
type CreateEquipmentRequest struct {
	Name            string  `json:"name" binding:"required"`
	SerialNumber    string  `json:"serial_number" binding:"required"`
	ZoneID          *string `json:"zone_id"`
	Status          string  `json:"status" binding:"required,oneof=working maintenance broken"`
	NextMaintenance *string `json:"next_maintenance"`
}

// Create создаёт оборудование (только директор)
func (h *EquipmentHandler) Create(c *gin.Context) {
	var req CreateEquipmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e, err := h.equipmentService.Create(c.Request.Context(), service.CreateEquipmentRequest{
		Name:            req.Name,
		SerialNumber:    req.SerialNumber,
		ZoneID:          req.ZoneID,
		Status:          req.Status,
		NextMaintenance: req.NextMaintenance,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"equipment": e})
}

// UpdateRequest — запрос на обновление
type UpdateEquipmentRequest struct {
	Name            string  `json:"name" binding:"required"`
	SerialNumber    string  `json:"serial_number" binding:"required"`
	ZoneID          *string `json:"zone_id"`
	Status          string  `json:"status" binding:"required,oneof=working maintenance broken"`
	NextMaintenance *string `json:"next_maintenance"`
}

// Update обновляет оборудование (только директор)
func (h *EquipmentHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req UpdateEquipmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	e, err := h.equipmentService.Update(c.Request.Context(), id, service.UpdateEquipmentRequest{
		Name:            req.Name,
		SerialNumber:    req.SerialNumber,
		ZoneID:          req.ZoneID,
		Status:          req.Status,
		NextMaintenance: req.NextMaintenance,
	})

	if err != nil {
		if err == service.ErrEquipmentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"equipment": e})
}

// CompleteMaintenanceRequest — запрос на завершение ТО
type CompleteMaintenanceRequest struct {
	Description     string `json:"description" binding:"required"`
	NextMaintenance string `json:"next_maintenance" binding:"required"`
}

// CompleteMaintenance фиксирует ТО
func (h *EquipmentHandler) CompleteMaintenance(c *gin.Context) {
	id := c.Param("id")

	var req CompleteMaintenanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	technicianID, _ := c.Get("user_id")

	log, err := h.equipmentService.CompleteMaintenance(c.Request.Context(), id, service.CompleteMaintenanceRequest{
		Description:     req.Description,
		NextMaintenance: req.NextMaintenance,
		TechnicianID:    technicianID.(string),
	})

	if err != nil {
		if err == service.ErrEquipmentNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "ТО успешно зафиксировано",
		"maintenance": log,
	})
}

// GetMaintenanceLogs возвращает историю ТО
func (h *EquipmentHandler) GetMaintenanceLogs(c *gin.Context) {
	id := c.Param("id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	logs, err := h.equipmentService.GetMaintenanceLogs(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"maintenance_logs": logs})
}

// GetUpcomingMaintenance возвращает оборудование с предстоящим ТО
func (h *EquipmentHandler) GetUpcomingMaintenance(c *gin.Context) {
	withinDays := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			withinDays = parsed
		}
	}

	upcoming, err := h.equipmentService.GetUpcomingMaintenance(c.Request.Context(), withinDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upcoming_maintenance": upcoming,
		"within_days":          withinDays,
		"count":                len(upcoming),
	})
}
