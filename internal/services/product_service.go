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

func (s *ProductService) GetInactiveStockProducts(
	contractGlobalID string,
	days, page, limit int,
	nameFilter *string,
) ([]models.InactiveStockProduct, int, error) {
	return s.repo.GetInactiveStockProducts(contractGlobalID, days, page, limit, nameFilter)
}

func (s *ProductService) GetProductStockWithSalesSpeed(contractGlobalID string, days int, goodsID string, speedOrRout int) ([]models.ProductStockWithSalesSpeed, error) {
	return s.repo.GetProductStockWithSalesSpeed(contractGlobalID, days, goodsID, speedOrRout)
}
