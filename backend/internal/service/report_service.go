package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"morepark/internal/repository"

	"github.com/xuri/excelize/v2"
)

type ReportService struct {
	reportRepo *repository.ReportRepository
}

func NewReportService(reportRepo *repository.ReportRepository) *ReportService {
	return &ReportService{reportRepo: reportRepo}
}

var categoryLabels = map[string]string{
	"chemical": "Химия",
	"drinks":   "Напитки",
	"food":     "Еда",
	"supplies": "Расходники",
}

var paymentLabels = map[string]string{
	"cash":   "Наличные",
	"card":   "Карта",
	"online": "Онлайн",
	"refund": "Возврат",
}

func (s *ReportService) ExportSales(ctx context.Context) ([]byte, string, error) {
	rows, err := s.reportRepo.GetSalesForExport(ctx)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Продажи"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"№", "Дата", "Время", "Тип операции", "Сумма (₽)", "Способ оплаты",
		"Канал", "Зона", "Тип билета", "Кол-во чел.", "Клиент", "Кассир", "№ билета",
	}
	writeHeader(f, sheet, headers)

	var totalSales, totalRefunds float64
	var salesCount, refundCount int

	for i, row := range rows {
		r := i + 2
		opType := "Продажа"
		amount := row.Amount
		if row.IsRefund {
			opType = "Возврат"
			totalRefunds += abs(row.Amount)
			refundCount++
			amount = -abs(row.Amount)
		} else {
			totalSales += row.Amount
			salesCount++
		}

		channel := "Касса"
		if row.Source == "online" {
			channel = "Онлайн"
		}

		values := []interface{}{
			i + 1,
			row.CreatedAt.Format("02.01.2006"),
			row.CreatedAt.Format("15:04"),
			opType,
			amount,
			paymentLabels[row.PaymentMethod],
			channel,
			row.ZoneName,
			row.TicketType,
			row.Quantity,
			row.CustomerName,
			row.CashierName,
			row.TicketNumber,
		}
		writeRow(f, sheet, r, values)
	}

	summaryRow := len(rows) + 3
	f.SetCellValue(sheet, fmt.Sprintf("A%d", summaryRow), "ИТОГО:")
	f.SetCellValue(sheet, fmt.Sprintf("B%d", summaryRow), fmt.Sprintf("Продаж: %d на %.2f ₽", salesCount, totalSales))
	f.SetCellValue(sheet, fmt.Sprintf("C%d", summaryRow), fmt.Sprintf("Возвратов: %d на %.2f ₽", refundCount, totalRefunds))
	f.SetCellValue(sheet, fmt.Sprintf("D%d", summaryRow), fmt.Sprintf("Чистая выручка: %.2f ₽", totalSales-totalRefunds))

	styleHeader(f, sheet, len(headers))
	setColWidths(f, sheet, []float64{5, 12, 8, 12, 12, 14, 10, 18, 20, 10, 22, 22, 16})

	return saveWorkbook(f, "prodazhi")
}

func (s *ReportService) ExportInventory(ctx context.Context) ([]byte, string, error) {
	items, err := s.reportRepo.GetInventoryForExport(ctx)
	if err != nil {
		return nil, "", err
	}
	movements, err := s.reportRepo.GetMovementsForExport(ctx)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()

	// Лист 1: Остатки
	stockSheet := "Остатки"
	f.SetSheetName("Sheet1", stockSheet)

	stockHeaders := []string{
		"№", "Наименование", "Категория", "Остаток", "Ед. изм.",
		"Мин. остаток", "Цена за ед. (₽)", "Сумма (₽)", "Срок годности",
	}
	writeHeader(f, stockSheet, stockHeaders)

	var totalStockValue float64
	for i, item := range items {
		r := i + 2
		totalStockValue += item.TotalValue
		expiry := "—"
		if item.ExpiryDate != nil {
			expiry = item.ExpiryDate.Format("02.01.2006")
		}
		writeRow(f, stockSheet, r, []interface{}{
			i + 1, item.Name, labelCategory(item.Category),
			item.Quantity, item.Unit, item.MinQuantity,
			item.Price, item.TotalValue, expiry,
		})
	}

	summaryRow := len(items) + 3
	f.SetCellValue(stockSheet, fmt.Sprintf("A%d", summaryRow), "ИТОГО стоимость склада:")
	f.SetCellValue(stockSheet, fmt.Sprintf("H%d", summaryRow), totalStockValue)

	styleHeader(f, stockSheet, len(stockHeaders))
	setColWidths(f, stockSheet, []float64{5, 28, 14, 10, 10, 12, 14, 14, 14})

	// Лист 2: Движения
	movSheet := "Движения"
	f.NewSheet(movSheet)
	movHeaders := []string{
		"№", "Дата", "Время", "Товар", "Категория", "Тип",
		"Кол-во", "Ед.", "Цена (₽)", "Сумма (₽)", "Причина", "Сотрудник",
	}
	writeHeader(f, movSheet, movHeaders)

	for i, m := range movements {
		r := i + 2
		movType := "Приход"
		qty := m.Quantity
		sum := m.TotalValue
		if m.Type == "out" {
			movType = "Расход"
			sum = -m.TotalValue
		}
		writeRow(f, movSheet, r, []interface{}{
			i + 1,
			m.CreatedAt.Format("02.01.2006"),
			m.CreatedAt.Format("15:04"),
			m.ItemName, labelCategory(m.Category), movType,
			qty, m.Unit, m.Price, sum, m.Reason, m.UserName,
		})
	}

	styleHeader(f, movSheet, len(movHeaders))
	setColWidths(f, movSheet, []float64{5, 12, 8, 24, 14, 10, 10, 8, 12, 12, 24, 22})

	return saveWorkbook(f, "sklad")
}

func (s *ReportService) ExportSummary(ctx context.Context) ([]byte, string, error) {
	sales, err := s.reportRepo.GetSalesForExport(ctx)
	if err != nil {
		return nil, "", err
	}
	inventory, err := s.reportRepo.GetInventoryForExport(ctx)
	if err != nil {
		return nil, "", err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheet := "Сводка"
	f.SetSheetName("Sheet1", sheet)

	f.SetCellValue(sheet, "A1", "Сводный отчёт для бухгалтерии")
	f.SetCellValue(sheet, "A2", "Аквапарк «Море Парк»")
	f.SetCellValue(sheet, "A3", fmt.Sprintf("Дата формирования: %s", time.Now().Format("02.01.2006 15:04")))

	var totalSales, totalRefunds float64
	var salesCount, refundCount, onlineCount, cashCount, cardCount int

	for _, row := range sales {
		if row.IsRefund {
			totalRefunds += abs(row.Amount)
			refundCount++
		} else {
			totalSales += row.Amount
			salesCount++
			switch row.PaymentMethod {
			case "online":
				onlineCount++
			case "cash":
				cashCount++
			case "card":
				cardCount++
			}
		}
	}

	var stockValue float64
	for _, item := range inventory {
		stockValue += item.TotalValue
	}

	rows := [][]interface{}{
		{"", ""},
		{"Показатель", "Значение"},
		{"Выручка (продажи)", fmt.Sprintf("%.2f ₽", totalSales)},
		{"Возвраты", fmt.Sprintf("%.2f ₽", totalRefunds)},
		{"Чистая выручка", fmt.Sprintf("%.2f ₽", totalSales-totalRefunds)},
		{"Кол-во продаж", salesCount},
		{"Кол-во возвратов", refundCount},
		{"", ""},
		{"По способу оплаты", ""},
		{"  Наличные", cashCount},
		{"  Карта", cardCount},
		{"  Онлайн", onlineCount},
		{"", ""},
		{"Стоимость склада ТМЦ", fmt.Sprintf("%.2f ₽", stockValue)},
		{"Позиций на складе", len(inventory)},
	}

	for i, row := range rows {
		r := i + 5
		f.SetCellValue(sheet, fmt.Sprintf("A%d", r), row[0])
		if row[1] != "" {
			f.SetCellValue(sheet, fmt.Sprintf("B%d", r), row[1])
		}
	}

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Color: "1E40AF"},
	})
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"E0F2FE"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "left"},
	})
	f.SetCellStyle(sheet, "A7", "B7", headerStyle)

	setColWidths(f, sheet, []float64{30, 20})

	return saveWorkbook(f, "svodka_buhgalteriya")
}

func writeHeader(f *excelize.File, sheet string, headers []string) {
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
}

func writeRow(f *excelize.File, sheet string, row int, values []interface{}) {
	for i, v := range values {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		f.SetCellValue(sheet, cell, v)
	}
}

func styleHeader(f *excelize.File, sheet string, colCount int) {
	endCol, _ := excelize.CoordinatesToCellName(colCount, 1)
	style, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"0284C7"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheet, "A1", endCol, style)
}

func setColWidths(f *excelize.File, sheet string, widths []float64) {
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}
}

func saveWorkbook(f *excelize.File, prefix string) ([]byte, string, error) {
	filename := fmt.Sprintf("%s_%s.xlsx", prefix, time.Now().Format("2006-01-02"))
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), filename, nil
}

func labelCategory(cat string) string {
	if label, ok := categoryLabels[cat]; ok {
		return label
	}
	return cat
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
