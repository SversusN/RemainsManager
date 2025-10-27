package services

import (
	"context"

	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

type ReportService struct {
	repo *repositories.ReportRepository
}

func NewReportService(repo *repositories.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

func (s *ReportService) BuildReport(ctx context.Context, offerID int64) (*models.GroupedReport, error) {
	items, err := s.repo.GetOfferItems(ctx, offerID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, nil // или ошибку "not found"
	}

	// Заголовок
	header := models.OfferHeader{
		MnemoCode:      items[0].MnemoCode,
		ContractorFrom: items[0].ContractorFrom,
	}

	// Группировка
	groups := make(map[string][]models.OfferItemReport)
	for _, item := range items {
		groups[item.ContractorTo] = append(groups[item.ContractorTo], item)
	}

	report := &models.GroupedReport{
		Header: header,
		Groups: groups,
	}

	return report, nil
}
