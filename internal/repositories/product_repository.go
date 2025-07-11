package repositories

import (
	"database/sql"
	"fmt"

	"RemainsManager/internal/models"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetInactiveStockProducts(contractGlobalID string, days, page, limit int) ([]models.InactiveStockProduct, int, error) {
	rows, err := r.db.Query("EXEC GetInactiveStockProducts @DAYS = @days, @CONTRACTOR = @contractor, @PAGE = @page, @LIMIT = @limit",
		sql.Named("days", days),
		sql.Named("contractor", contractGlobalID),
		sql.Named("page", page),
		sql.Named("limit", limit),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var products []models.InactiveStockProduct
	for rows.Next() {
		var p models.InactiveStockProduct
		err := rows.Scan(&p.Name, &p.Qty, &p.PriceSal, &p.PriceProd, &p.DaysNoMovement, &p.BestBefore)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}
		products = append(products, p)
	}

	// Переходим ко второму набору результатов
	if hasMore := rows.NextResultSet(); !hasMore {
		return nil, 0, fmt.Errorf("failed to read next result set: %w", err)
	}

	var totalCount int
	if rows.Next() {
		if err := rows.Scan(&totalCount); err != nil {
			return nil, 0, fmt.Errorf("failed to scan total count: %w", err)
		}
	} else {
		totalCount = len(products) // fallback
	}

	return products, totalCount, nil
}

func (r *ProductRepository) GetProductStockWithSalesSpeed(contractGlobalID string, days int, goodsID *int64) ([]models.ProductStockWithSalesSpeed, error) {
	var rows *sql.Rows
	var err error

	if goodsID == nil || *goodsID == 0 {
		rows, err = r.db.Query("EXEC GetProductStockWithSalesSpeed @DAYS = ?, @CONTRACTOR = ?", days, contractGlobalID)
	} else {
		rows, err = r.db.Query("EXEC GetProductStockWithSalesSpeed @DAYS = ?, @CONTRACTOR = ?, @GOODS_ID = ?", days, contractGlobalID, *goodsID)
	}

	if err != nil {
		return nil, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var products []models.ProductStockWithSalesSpeed
	for rows.Next() {
		var p models.ProductStockWithSalesSpeed
		err := rows.Scan(&p.Name, &p.Qty, &p.PriceSal, &p.PriceProd, &p.BestBefore, &p.TotalSold, &p.SalesPerDay, &p.ActiveDays)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}
