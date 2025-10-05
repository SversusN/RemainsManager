package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"RemainsManager/internal/models"
)

type ProductRepository struct {
	db      *sql.DB
	timeout int
}

func NewProductRepository(timeout int, db *sql.DB) *ProductRepository {
	return &ProductRepository{timeout: timeout, db: db}
}

func (r *ProductRepository) GetInactiveStockProducts(contractGlobalID string, days, page, limit int, nameFilter *string) ([]models.InactiveStockProduct, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()
	var rows *sql.Rows
	var err error
	if nameFilter == nil || *nameFilter == "" {
		rows, err = r.db.QueryContext(ctx, `
			EXEC GetInactiveStockProducts 
				@DAYS = @days, 
				@CONTRACTOR = @contractor, 
				@PAGE = @page, 
				@LIMIT = @limit`,
			sql.Named("days", days),
			sql.Named("contractor", contractGlobalID),
			sql.Named("page", page),
			sql.Named("limit", limit),
		)
	} else {
		rows, err = r.db.QueryContext(ctx, `
			EXEC GetInactiveStockProducts 
				@DAYS = @days, 
				@CONTRACTOR = @contractor, 
				@PAGE = @page, 
				@LIMIT = @limit, 
				@NAME = @name`,
			sql.Named("days", days),
			sql.Named("contractor", contractGlobalID),
			sql.Named("page", page),
			sql.Named("limit", limit),
			sql.Named("name", *nameFilter),
		)
	}
	if err != nil {
		return nil, 0, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var products []models.InactiveStockProduct
	for rows.Next() {
		var p models.InactiveStockProduct
		err := rows.Scan(&p.Name, &p.Qty, &p.PriceSal, &p.PriceProd, &p.DaysNoMovement, &p.BestBefore, &p.IdGoodsGlobal)
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

func (r *ProductRepository) GetProductStockWithSalesSpeed(contractGlobalID string, days int, goodsID string) ([]models.ProductStockWithSalesSpeed, error) {
	var rows *sql.Rows
	var err error

	query := "EXEC GetProductStockWithSalesSpeed @DAYS = @days, @CONTRACTOR = @contractor"
	args := []interface{}{
		sql.Named("days", days),
		sql.Named("contractor", contractGlobalID),
	}

	if goodsID != "" {
		query += ", @GOODS_ID = @goods_id"
		args = append(args, sql.Named("goods_id", goodsID))
	}

	rows, err = r.db.Query(query, args...)

	if err != nil {
		return nil, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var products []models.ProductStockWithSalesSpeed
	for rows.Next() {
		var p models.ProductStockWithSalesSpeed
		err := rows.Scan(&p.Name, &p.IdGoodsGlobal, &p.ContractorName, &p.IdContractorGlobal, &p.Qty, &p.PriceSal, &p.PriceProd, &p.BestBefore, &p.TotalSold, &p.SalesPerDay, &p.ActiveDays)
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
