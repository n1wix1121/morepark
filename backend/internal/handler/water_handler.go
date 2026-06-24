package handler

import (
	"net/http"
	"strconv"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type WaterHandler struct {
	waterService *service.WaterService
}

func NewWaterHandler(waterService *service.WaterService) *WaterHandler {
	return &WaterHandler{waterService: waterService}
}

// CreateMeasurementRequest — запрос на ввод замера
type CreateMeasurementRequest struct {
	ZoneID    string  `json:"zone_id" binding:"required"`
	PH        float64 `json:"ph" binding:"required,min=0,max=14"`
	Chlorine  float64 `json:"chlorine" binding:"required,min=0"`
	Turbidity float64 `json:"turbidity" binding:"required,min=0"`
}

// CreateMeasurement сохраняет замер (только тех. служба и директор)
func (h *WaterHandler) CreateMeasurement(c *gin.Context) {
	var req CreateMeasurementRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Берём ID техника из JWT
	technicianID, _ := c.Get("user_id")

	measurement, err := h.waterService.CreateMeasurement(c.Request.Context(), service.CreateMeasurementRequest{
		ZoneID:       req.ZoneID,
		PH:           req.PH,
		Chlorine:     req.Chlorine,
		Turbidity:    req.Turbidity,
		TechnicianID: technicianID.(string),
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Формируем ответ с информацией о нарушениях
	response := gin.H{
		"measurement": measurement,
		"is_normal":   measurement.IsNormal,
	}

	if !measurement.IsNormal {
		response["message"] = "⚠️ Обнаружены нарушения норм СанПиН!"
		response["violations"] = measurement.Violations
	} else {
		response["message"] = "✅ Показатели в норме"
	}

	c.JSON(http.StatusCreated, response)
}

// GetMeasurements возвращает все замеры
func (h *WaterHandler) GetMeasurements(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	measurements, err := h.waterService.GetMeasurements(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"measurements": measurements})
}

// GetMeasurementsByZone возвращает замеры по конкретной зоне
func (h *WaterHandler) GetMeasurementsByZone(c *gin.Context) {
	zoneID := c.Param("id")

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	measurements, err := h.waterService.GetMeasurementsByZone(c.Request.Context(), zoneID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"measurements": measurements})
}

// GetAlerts возвращает все замеры с нарушениями
func (h *WaterHandler) GetAlerts(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	alerts, err := h.waterService.GetAlerts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
