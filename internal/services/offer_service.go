package services

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"
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

func (s *OfferService) MarkAsSent(ctx context.Context, offerID int64) error {
	return s.repo.UpdateOfferStatus(ctx, offerID, models.OfferStatusSent)
}

func (s *OfferService) DeleteOffer(ctx context.Context, offerID int64) error {
	return s.repo.UpdateOfferStatus(ctx, offerID, models.OfferStatusDeleted)
}

func (s *OfferService) ProcessOffer(ctx context.Context, offerID int64) error {
	// 1. Получаем данные из процедуры
	rows, err := s.repo.ProcessOffer(ctx, offerID)
	if err != nil {
		return fmt.Errorf("failed to generate interfirm data: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Errorf("no items found in offer %d", offerID)
	}

	// 2. Группируем по ID_CONTRACTOR_GLOBAL_TO
	grouped := make(map[string][]models.InterfirmRow)
	for _, row := range rows {
		key := row.IDContractorGlobalTo
		grouped[key] = append(grouped[key], row)
	}

	// 3. Для каждой группы генерируем XML и сохраняем
	for contractorID, items := range grouped {
		log.Printf("[Offer %d] Building interfirm moving for contractor: %s, item count: %d",
			offerID, contractorID, len(items))

		// Генерируем XML
		xmlDoc, err := s.buildInterfirmXML(items)
		if err != nil {
			return fmt.Errorf("failed to build XML for contractor %s: %w", contractorID, err)
		}

		// Сериализуем в строку
		xmlData, err := xml.Marshal(xmlDoc)
		if err != nil {
			return fmt.Errorf("failed to marshal XML for %s: %w", contractorID, err)
		}

		fullXML := string(xmlData)

		// 🔹 ЛОГ: выводим XML (в одну строку, без переносов)
		log.Printf("[Offer %d][Contractor %s] Generated XML:\n%s",
			offerID, contractorID, fullXML)

		// 🔹 ЛОГ: начало сохранения
		log.Printf("[Offer %d][Contractor %s] Sending XML to USP_INTERFIRM_MOVING_SAVE...", offerID, contractorID)

		// Сохраняем через хранимку
		if err := s.repo.SaveInterfirmMoving(ctx, fullXML); err != nil {
			log.Printf("[ERROR][Offer %d][Contractor %s] Failed to save interfirm moving: %v",
				offerID, contractorID, err)
			return fmt.Errorf("failed to save interfirm moving for %s: %w", contractorID, err)
		}

		// 🔹 ЛОГ: успех
		log.Printf("[SUCCESS][Offer %d][Contractor %s] Interfirm moving saved successfully",
			offerID, contractorID)
	}

	// 4. Меняем статус заявки на "отработана"
	if err := s.repo.UpdateOfferStatus(ctx, offerID, models.OfferStatusProcessed); err != nil {
		log.Printf("[ERROR][Offer %d] Failed to update offer status: %v", offerID, err)
		return fmt.Errorf("failed to update offer status to processed: %w", err)
	}

	log.Printf("[SUCCESS][Offer %d] Offer processed and marked as 'processed'", offerID)
	return nil
}

func (s *OfferService) buildInterfirmXML(rows []models.InterfirmRow) (*models.InterfirmXML, error) {
	if len(rows) == 0 {
		return nil, fmt.Errorf("no rows provided")
	}

	first := rows[0]

	items := make([]models.InterfirmItemXML, 0, len(rows))
	for _, r := range rows {
		items = append(items, models.InterfirmItemXML{
			IDInterfirmMovingItem:       r.IDInterfirmMovingItem,
			IDInterfirmMovingItemGlobal: r.IDInterfirmMovingItemGlobal,
			Quantity:                    r.Quantity,
			IDLotFrom:                   r.IDLotFrom,
			IDLotTo:                     r.IDLotTo,
			SumSupplier:                 r.SumSupplier,
			SVatSupplier:                r.SVatSupplier,
			PVatRetail:                  r.PVatRetail,
			VatRetail:                   r.VatRetail,
			IsWeight:                    r.IsWeight,
			Kiz:                         r.Kiz,
			IsKiz:                       r.IsKiz,
		})
	}

	// Формат даты как в примере: 2025-10-12T09:42:59.237
	dateStr := first.Date.Format("2006-01-02T15:04:05.000")

	var authValidPeriodStr *string
	if first.AuthValidPeriod != nil {
		s := first.AuthValidPeriod.Format("2006-01-02T15:04:05.000")
		authValidPeriodStr = &s
	}

	return &models.InterfirmXML{
		Moving: models.InterfirmMovingXML{
			IDInterfirmMoving:       first.IDInterfirmMoving,
			IDInterfirmMovingGlobal: first.IDInterfirmMovingGlobal,
			Mnemocode:               first.Mnemocode,
			IDStoreFromMain:         first.IDStoreFromMain,
			IDStoreFromTransit:      first.IDStoreFromTransit,
			IDContractorTo:          first.IDContractorTo,
			IDStoreToMain:           first.IDStoreToMain,
			IDStoreToTransit:        first.IDStoreToTransit,
			Date:                    dateStr,
			DocumentState:           first.DocumentState,
			Comment:                 first.Comment,
			IDUser:                  first.IDUser,
			IDUser2:                 first.IDUser2,
			SumSupplier:             first.SumSupplierHeader,
			SVatSupplier:            first.SVatSupplierHeader,
			SumRetail:               first.SumRetailHeader,
			SVatRetail:              first.SVatRetailHeader,
			GoodsSent:               first.GoodsSent,
			AuthNum:                 first.AuthNum,
			AuthValidPeriod:         authValidPeriodStr,
			Items:                   items,
		},
	}, nil
}
