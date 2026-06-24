package service

import (
	"context"
	"fmt"

	"morepark/internal/domain"
	"morepark/internal/repository"
)

// Нормы СанПиН для бассейнов
const (
	PHMin        = 7.2
	PHMax        = 7.6
	ChlorineMax  = 0.5 // мг/л
	TurbidityMax = 1.5 // ЕМФ
)

type WaterService struct {
	waterRepo *repository.WaterRepository
	zoneRepo  *repository.ZoneRepository
}

func NewWaterService(waterRepo *repository.WaterRepository, zoneRepo *repository.ZoneRepository) *WaterService {
	return &WaterService{
		waterRepo: waterRepo,
		zoneRepo:  zoneRepo,
	}
}

// CreateMeasurementRequest — запрос на ввод замера
type CreateMeasurementRequest struct {
	ZoneID       string  `json:"zone_id"`
	PH           float64 `json:"ph"`
	Chlorine     float64 `json:"chlorine"`
	Turbidity    float64 `json:"turbidity"`
	TechnicianID string  `json:"technician_id"`
}

// ValidateBySanPiN проверяет показатели по нормам СанПиН
// Возвращает список нарушений (если пустой — всё в норме)
func (s *WaterService) ValidateBySanPiN(m domain.WaterQuality) []string {
	var violations []string

	if m.PH < PHMin {
		violations = append(violations, fmt.Sprintf("pH ниже нормы (%.2f < %.2f)", m.PH, PHMin))
	}
	if m.PH > PHMax {
		violations = append(violations, fmt.Sprintf("pH выше нормы (%.2f > %.2f)", m.PH, PHMax))
	}
	if m.Chlorine > ChlorineMax {
		violations = append(violations, fmt.Sprintf("хлор превышен (%.2f > %.2f мг/л)", m.Chlorine, ChlorineMax))
	}
	if m.Turbidity > TurbidityMax {
		violations = append(violations, fmt.Sprintf("мутность превышена (%.2f > %.2f ЕМФ)", m.Turbidity, TurbidityMax))
	}

	return violations
}

// CreateMeasurement сохраняет замер и проверяет нормы
func (s *WaterService) CreateMeasurement(ctx context.Context, req CreateMeasurementRequest) (*domain.WaterQuality, error) {
	// 1. Проверяем, что зона существует
	zone, err := s.zoneRepo.GetByID(ctx, req.ZoneID)
	if err != nil {
		return nil, fmt.Errorf("зона не найдена: %w", err)
	}

	// 2. Создаём модель
	measurement := domain.WaterQuality{
		ZoneID:       req.ZoneID,
		PH:           req.PH,
		Chlorine:     req.Chlorine,
		Turbidity:    req.Turbidity,
		TechnicianID: req.TechnicianID,
	}

	// 3. Валидация по СанПиН
	violations := s.ValidateBySanPiN(measurement)
	measurement.Violations = violations
	measurement.IsNormal = len(violations) == 0

	// 4. Сохраняем в БД
	if err := s.waterRepo.Create(ctx, &measurement); err != nil {
		return nil, err
	}

	// 5. Заполняем вложенные данные для ответа
	measurement.Zone = zone

	// 6. Если есть нарушения — логируем (позже заменим на WebSocket-уведомления)
	if !measurement.IsNormal {
		fmt.Printf("⚠️  НАРУШЕНИЕ САнПиН в зоне '%s': %v\n", zone.Name, violations)
		// TODO: отправить уведомление директору и спасателям через WebSocket
	}

	return &measurement, nil
}

// GetMeasurements возвращает все замеры
func (s *WaterService) GetMeasurements(ctx context.Context, limit int) ([]domain.WaterQuality, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.waterRepo.GetAll(ctx, limit)
}

// GetMeasurementsByZone возвращает замеры по конкретной зоне
func (s *WaterService) GetMeasurementsByZone(ctx context.Context, zoneID string, limit int) ([]domain.WaterQuality, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.waterRepo.GetByZoneID(ctx, zoneID, limit)
}

// GetAlerts возвращает все замеры с нарушениями
func (s *WaterService) GetAlerts(ctx context.Context, limit int) ([]domain.WaterQuality, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.waterRepo.GetAlerts(ctx, limit)
}
