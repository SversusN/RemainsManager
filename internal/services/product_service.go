package services

import (
	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

type ProductService struct {
	repo *repositories.ProductRepository
}

func NewProductService(repo *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) GetInactiveStockProducts(contractGlobalID string, days, page, limit int) ([]models.InactiveStockProduct, int, error) {
	return s.repo.GetInactiveStockProducts(contractGlobalID, days, page, limit)
}

func (s *ProductService) GetProductStockWithSalesSpeed(contractGlobalID string, days int, goodsID *int64) ([]models.ProductStockWithSalesSpeed, error) {
	return s.repo.GetProductStockWithSalesSpeed(contractGlobalID, days, goodsID)
}
