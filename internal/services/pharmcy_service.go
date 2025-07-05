package services

import (
	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

type PharmacyService struct {
	repo *repositories.PharmacyRepository
}

func NewPharmacyService(repo *repositories.PharmacyRepository) *PharmacyService {
	return &PharmacyService{repo: repo}
}

func (s *PharmacyService) GetPharmacies() ([]models.Pharmacy, error) {
	return s.repo.GetPharmacies()
}
