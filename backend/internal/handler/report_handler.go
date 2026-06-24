package handler

import (
	"net/http"

	"morepark/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) sendExcel(c *gin.Context, data []byte, filename string) {
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}

// ExportSales — отчёт по продажам и возвратам
func (h *ReportHandler) ExportSales(c *gin.Context) {
	data, filename, err := h.reportService.ExportSales(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка формирования отчёта"})
		return
	}
	h.sendExcel(c, data, filename)
}

// ExportInventory — отчёт по складу (остатки + движения)
func (h *ReportHandler) ExportInventory(c *gin.Context) {
	data, filename, err := h.reportService.ExportInventory(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка формирования отчёта"})
		return
	}
	h.sendExcel(c, data, filename)
}

// ExportSummary — сводный отчёт для бухгалтерии
func (h *ReportHandler) ExportSummary(c *gin.Context) {
	data, filename, err := h.reportService.ExportSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка формирования отчёта"})
		return
	}
	h.sendExcel(c, data, filename)
}
