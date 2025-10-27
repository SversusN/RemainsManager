package repositories

import (
	"context"
	"database/sql"
	"fmt"

	"RemainsManager/internal/models"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetOfferItems(ctx context.Context, offerID int64) ([]models.OfferItemReport, error) {
	query := `
SELECT 
  o.NAME as mnemocode,
  c.NAME as contractor_from,
  g.NAME as goods_name,
  oi.QUANTITY as qty,
  c2.NAME as contractor_to,
  lot.PRICE_SAL as price_sal,
  lot.LOT_NAME as lot_name
FROM OFFER o
INNER JOIN OFFER_ITEM oi ON o.ID_OFFER = oi.ID_OFFER
INNER JOIN CONTRACTOR c ON c.ID_CONTRACTOR_GLOBAL = o.ID_CONTRACTOR_GLOBAL_FROM
INNER JOIN GOODS g ON g.ID_GOODS_GLOBAL = oi.GOODS_ID
INNER JOIN CONTRACTOR c2 ON c2.ID_CONTRACTOR_GLOBAL = oi.ID_CONTRACTOR_GLOBAL_TO
INNER JOIN LOT ON lot.ID_LOT_GLOBAL = oi.ID_LOT_GLOBAL
WHERE o.ID_OFFER = @offer_id`

	rows, err := r.db.QueryContext(ctx, query, sql.Named("offer_id", offerID))
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var items []models.OfferItemReport
	for rows.Next() {
		var item models.OfferItemReport
		err := rows.Scan(
			&item.MnemoCode,
			&item.ContractorFrom,
			&item.GoodsName,
			&item.Qty,
			&item.ContractorTo,
			&item.PriceSal,
			&item.LotName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return items, nil
}
