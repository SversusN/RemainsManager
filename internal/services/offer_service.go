package services

import (
	"context"
	"fmt"
	"time"

	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

type OfferService struct {
	repo *repositories.OfferRepository
}

func NewOfferService(repo *repositories.OfferRepository) *OfferService {
	return &OfferService{repo: repo}
}

func (s *OfferService) GetOrCreateTodayOffer(ctx context.Context, fromID, fromName string) (*models.Offer, error) {
	return s.repo.GetOrCreateTodayOffer(ctx, fromID, fromName)
}

func (s *OfferService) AddItems(ctx context.Context, items []models.OfferItem) error {
	return s.repo.AddItems(ctx, items)
}

func (s *OfferService) GetOfferJournal(ctx context.Context, from, to time.Time, contractorGlobal *string) ([]models.OfferJournalItem, error) {
	return s.repo.GetOfferJournal(ctx, from, to, contractorGlobal)
}

func (s *OfferService) GetOfferDetails(ctx context.Context, offerID int64) ([]models.OfferDetailItem, error) {
	return s.repo.GetOfferDetails(ctx, offerID)
}

func (s *OfferService) UpdateOfferItem(ctx context.Context, id int64, quantity int) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be greater than zero")
	}
	return s.repo.UpdateOfferItem(ctx, id, quantity)
}

func (s *OfferService) DeleteOfferItem(ctx context.Context, id int64) error {
	return s.repo.DeleteOfferItem(ctx, id)
}
