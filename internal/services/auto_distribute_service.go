package services

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"

	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

// AutoDistributeService отвечает за автоматическое формирование заявок
// на основе скорости продаж контрагентов.
type AutoDistributeService struct {
	productService *ProductService
	offerRepo      *repositories.OfferRepository
}

func NewAutoDistributeService(
	productService *ProductService,
	offerRepo *repositories.OfferRepository,
) *AutoDistributeService {
	return &AutoDistributeService{
		productService: productService,
		offerRepo:      offerRepo,
	}
}

// Distribute автоматически распределяет неактивные товары из одной аптеки
// по контрагентам с наибольшей скоростью продаж.
// Добавляет позиции в сегодняшнюю активную заявку.
func (s *AutoDistributeService) Distribute(ctx context.Context, fromID string, days int) error {
	// 1. Получаем неактивные товары из указанной аптеки
	inactive, _, err := s.productService.GetInactiveStockProducts(fromID, days, 1, 1000, nil)
	if err != nil {
		return fmt.Errorf("failed to get inactive products: %w", err)
	}

	if len(inactive) == 0 {
		log.Printf("No inactive products found for contractor %s", fromID)
		return nil
	}

	// 2. Получаем или создаём сегодняшнюю заявку
	fromName := s.offerRepo.GetContractorName(ctx, fromID)
	if fromName == "" {
		fromName = "Автоматическая заявка"
	}

	offer, err := s.offerRepo.GetOrCreateTodayOffer(ctx, fromID, fromName)
	if err != nil {
		return fmt.Errorf("failed to get or create today's offer: %w", err)
	}

	var allItems []models.OfferItem

	// 3. Для каждого неактивного товара:
	for _, prod := range inactive {
		// Получаем всех контрагентов, у которых этот товар в продаже
		speedData, err := s.productService.GetProductStockWithSalesSpeed(fromID, days, prod.IdGoodsGlobal, 0)
		if err != nil {
			log.Printf("Error fetching sales data for product %s: %v", prod.IdGoodsGlobal, err)
			continue
		}

		// Фильтруем: исключаем отправителя
		var filtered []models.ProductStockWithSalesSpeed
		for _, d := range speedData {
			if d.IdContractorGlobal != fromID {
				filtered = append(filtered, d)
			}
		}

		if len(filtered) == 0 {
			continue
		}

		// Сортируем по SalesPerDay (по убыванию)
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].SalesPerDay > filtered[j].SalesPerDay
		})

		// Берём топ-3 получателя
		topReceivers := filtered[:int(math.Min(3, float64(len(filtered))))]

		// Распределяем количество: поровну + остаток первому
		qtyPerReceiver := math.Floor(prod.Qty / float64(len(topReceivers)))
		remainder := int(prod.Qty) - int(qtyPerReceiver*float64(len(topReceivers)))

		for i, receiver := range topReceivers {
			qty := qtyPerReceiver
			if i == 0 {
				qty += float64(remainder)
			}

			if qty <= 0 {
				continue
			}

			item := models.OfferItem{
				OfferID:                offer.ID,
				IdContractorGlobalFrom: fromID,
				IdContractorGlobalTo:   receiver.IdContractorGlobal,
				GoodsId:                prod.IdGoodsGlobal,
				Quantity:               int(qty),
				IdLotGlobal:            prod.IdLotGlobal,
			}

			allItems = append(allItems, item)
		}
	}

	// 4. Массово добавляем все позиции в заявку
	if len(allItems) > 0 {
		if err := s.offerRepo.AddItems(ctx, allItems); err != nil {
			return fmt.Errorf("failed to add items to offer: %w", err)
		}
		log.Printf("Auto-distributed %d items to offer ID=%d", len(allItems), offer.ID)
	} else {
		log.Println("No items were distributed during auto-distribution")
	}

	return nil
}
