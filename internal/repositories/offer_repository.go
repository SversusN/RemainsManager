package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"RemainsManager/internal/models"
)

type OfferRepository struct {
	db      *sql.DB
	timeout int
}

func NewOfferRepository(timeout int, db *sql.DB) *OfferRepository {
	return &OfferRepository{db: db, timeout: timeout}
}

// GetOrCreateTodayOffer возвращает существующую или создаёт новую заявку на сегодня
func (r *OfferRepository) GetOrCreateTodayOffer(ctx context.Context, fromID, fromName string) (*models.Offer, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	var offer models.Offer

	err := r.db.QueryRowContext(ctx, `
		SELECT ID_OFFER, NAME, ID_CONTRACTOR_GLOBAL_FROM, CREATED_AT
		FROM OFFER
		WHERE ID_CONTRACTOR_GLOBAL_FROM = @from_id
		  AND CAST(CREATED_AT AS DATE) = CAST(GETDATE() AS DATE)
	`, sql.Named("from_id", fromID)).Scan(&offer.ID, &offer.Name, &offer.IdContractorGlobalFrom, &offer.CreatedAt)

	if err == nil {
		// Заявка найдена — загружаем позиции
		items, err := r.loadOfferItems(ctx, offer.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load offer items: %w", err)
		}
		offer.OfferItems = items
		return &offer, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Заявка не найдена — создаём новую
	name := time.Now().Format("02.01.2006") + " - " + fromName

	var newID int64
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO OFFER (NAME, ID_CONTRACTOR_GLOBAL_FROM, CREATED_AT, STATUS)
		OUTPUT INSERTED.ID_OFFER
		VALUES (@name, @from_id, GETDATE(), 0)
	`, sql.Named("name", name), sql.Named("from_id", fromID)).Scan(&newID)

	if err != nil {
		return nil, fmt.Errorf("failed to create offer: %w", err)
	}

	offer.ID = newID
	offer.Name = name
	offer.IdContractorGlobalFrom = fromID
	offer.CreatedAt = time.Now()
	offer.OfferItems = []models.OfferItem{}

	return &offer, nil
}

// AddItems обновляет или добавляет позиции в заявку (объединяет по GOODS_ID)
func (r *OfferRepository) AddItems(ctx context.Context, items []models.OfferItem) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, item := range items {
		_, err := tx.ExecContext(ctx, `
			IF EXISTS (
				SELECT 1 FROM OFFER_ITEM 
				WHERE ID_OFFER = @offer_id AND GOODS_ID = @goods_id AND ID_CONTRACTOR_GLOBAL_TO = @to_id AND ID_LOT_GLOBAL = @ID_LOT_GLOBAL
			)
				UPDATE OFFER_ITEM 
				SET QUANTITY = @quantity
				WHERE ID_OFFER = @offer_id AND GOODS_ID = @goods_id AND ID_CONTRACTOR_GLOBAL_TO = @to_id AND ID_LOT_GLOBAL = @ID_LOT_GLOBAL
			ELSE
				INSERT INTO OFFER_ITEM 
				(ID_OFFER, ID_CONTRACTOR_GLOBAL_FROM, ID_CONTRACTOR_GLOBAL_TO, GOODS_ID, QUANTITY, ID_LOT_GLOBAL)
				VALUES (@offer_id, @from_id, @to_id, @goods_id, @quantity, @ID_LOT_GLOBAL)
		`,
			sql.Named("offer_id", item.OfferID),
			sql.Named("goods_id", item.GoodsId),
			sql.Named("quantity", item.Quantity),
			sql.Named("from_id", item.IdContractorGlobalFrom),
			sql.Named("to_id", item.IdContractorGlobalTo),
			sql.Named("id_lot_global", item.IdLotGlobal),
		)
		if err != nil {
			return fmt.Errorf("failed to upsert item for goods_id=%s: %w", item.GoodsId, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// loadOfferItems загружает все позиции заявки
func (r *OfferRepository) loadOfferItems(ctx context.Context, offerID int64) ([]models.OfferItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ID_OFFER_ITEM, ID_CONTRACTOR_GLOBAL_FROM, ID_CONTRACTOR_GLOBAL_TO, GOODS_ID, QUANTITY
		FROM OFFER_ITEM
		WHERE ID_OFFER = @offer_id
	`, sql.Named("offer_id", offerID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OfferItem
	for rows.Next() {
		var i models.OfferItem
		if err := rows.Scan(&i.ID, &i.IdContractorGlobalFrom, &i.IdContractorGlobalTo, &i.GoodsId, &i.Quantity); err != nil {
			return nil, err
		}
		i.OfferID = offerID
		items = append(items, i)
	}
	return items, nil
}

// GetOfferJournal возвращает журнал заявок с фильтрацией по диапазону дат
// GetOfferJournal возвращает журнал заявок с фильтрацией по датам и контрагенту
func (r *OfferRepository) GetOfferJournal(ctx context.Context, from, to time.Time, contractorGlobal *string) ([]models.OfferJournalItem, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	// Базовый запрос
	query := `
		SELECT 
			o.ID_OFFER,
			o.NAME AS mnemocode,
			c.NAME AS contractor,
			CAST(o.CREATED_AT AS DATE) AS created,
			o.STATUS
		FROM OFFER o
		INNER JOIN CONTRACTOR c ON c.ID_CONTRACTOR_GLOBAL = o.ID_CONTRACTOR_GLOBAL_FROM
		WHERE CAST(o.CREATED_AT AS DATE) BETWEEN @from AND @to
	`

	// Добавляем фильтр по контрагенту, если указан
	if contractorGlobal != nil && *contractorGlobal != "" {
		query += " AND o.ID_CONTRACTOR_GLOBAL_FROM = @contractor_global"
	}

	query += " ORDER BY o.CREATED_AT DESC"

	// Подготавливаем параметры
	var rows *sql.Rows
	var err error

	if contractorGlobal != nil && *contractorGlobal != "" {
		rows, err = r.db.QueryContext(ctx, query,
			sql.Named("from", from.Format("2006-01-02")),
			sql.Named("to", to.Format("2006-01-02")),
			sql.Named("contractor_global", *contractorGlobal),
		)
	} else {
		rows, err = r.db.QueryContext(ctx, query,
			sql.Named("from", from.Format("2006-01-02")),
			sql.Named("to", to.Format("2006-01-02")),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to execute journal query: %w", err)
	}
	defer rows.Close()

	var items []models.OfferJournalItem
	for rows.Next() {
		var item models.OfferJournalItem
		err := rows.Scan(&item.ID, &item.Mnemocode, &item.Contractor, &item.CreatedAt, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// GetOfferDetails возвращает детали заявки по ID
func (r *OfferRepository) GetOfferDetails(ctx context.Context, offerID int64) ([]models.OfferDetailItem, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	query := `
		SELECT 
			CONCAT(g.NAME, ' | ', p.NAME) AS goods_name,
			c.NAME AS contractor_to,
			oi.QUANTITY AS quantity,
			oi.id_offer_item as itemId
		FROM OFFER o
		INNER JOIN OFFER_ITEM oi ON o.ID_OFFER = oi.ID_OFFER
		INNER JOIN CONTRACTOR c ON c.ID_CONTRACTOR_GLOBAL = oi.ID_CONTRACTOR_GLOBAL_TO
		INNER JOIN GOODS g ON g.ID_GOODS_GLOBAL = oi.GOODS_ID
		INNER JOIN PRODUCER p ON p.ID_PRODUCER = g.ID_PRODUCER
		WHERE o.ID_OFFER = @offer_id
		ORDER BY g.NAME
	`

	rows, err := r.db.QueryContext(ctx, query, sql.Named("offer_id", offerID))
	if err != nil {
		return nil, fmt.Errorf("failed to execute details query: %w", err)
	}
	defer rows.Close()

	var items []models.OfferDetailItem
	for rows.Next() {
		var item models.OfferDetailItem
		err := rows.Scan(&item.GoodsName, &item.ContractorTo, &item.Quantity, &item.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detail row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return items, nil
}

// UpdateOfferItem обновляет количество позиции в заявке
func (r *OfferRepository) UpdateOfferItem(ctx context.Context, id int64, quantity int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `
		UPDATE OFFER_ITEM 
		SET QUANTITY = @quantity 
		WHERE ID_OFFER_ITEM = @id`,
		sql.Named("quantity", quantity),
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf("failed to update offer item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("offer item with id %d not found", id)
	}

	return nil
}

// DeleteOfferItem удаляет позицию из заявки по ID
func (r *OfferRepository) DeleteOfferItem(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `
		DELETE FROM OFFER_ITEM 
		WHERE ID_OFFER_ITEM = @id`,
		sql.Named("id", id),
	)
	if err != nil {
		return fmt.Errorf("failed to delete offer item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("offer item with id %d not found", id)
	}

	return nil

}

// UpdateOfferStatus обновляет статус заявки по ID
func (r *OfferRepository) UpdateOfferStatus(ctx context.Context, offerID int64, status int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `
		UPDATE OFFER 
		SET STATUS = @status 
		WHERE ID_OFFER = @id`,
		sql.Named("status", status),
		sql.Named("id", offerID),
	)
	if err != nil {
		return fmt.Errorf("failed to update offer status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("offer with id %d not found", offerID)
	}

	return nil
}

// ProcessOffer вызывает usp_GenerateInterfirmMovingFromOffer и возвращает сырые строки
func (r *OfferRepository) ProcessOffer(ctx context.Context, offerID int64) ([]models.InterfirmRow, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
		EXEC usp_GenerateInterfirmMovingFromOffer @id_offer = @offer_id
	`, sql.Named("offer_id", offerID))
	if err != nil {
		return nil, fmt.Errorf("failed to execute usp_GenerateInterfirmMovingFromOffer: %w", err)
	}
	defer rows.Close()

	var results []models.InterfirmRow
	for rows.Next() {
		var row models.InterfirmRow
		err := rows.Scan(
			&row.IDContractorGlobalTo,
			&row.IDInterfirmMoving,
			&row.IDInterfirmMovingGlobal,
			&row.Mnemocode,
			&row.IDStoreFromMain,
			&row.IDStoreFromTransit,
			&row.IDContractorTo,
			&row.IDStoreToMain,
			&row.IDStoreToTransit,
			&row.Date,
			&row.DocumentState,
			&row.Comment,
			&row.IDUser,
			&row.IDUser2,
			&row.SumSupplierHeader,
			&row.SVatSupplierHeader,
			&row.SumRetailHeader,
			&row.SVatRetailHeader,
			&row.GoodsSent,
			&row.AuthNum,
			&row.AuthValidPeriod,
			&row.IDInterfirmMovingItem,
			&row.IDInterfirmMovingItemGlobal,
			&row.Quantity,
			&row.IDLotFrom,
			&row.IDLotTo,
			&row.SumSupplier,
			&row.SVatSupplier,
			&row.PVatRetail,
			&row.VatRetail,
			&row.IsWeight,
			&row.Kiz,   // ← NULL
			&row.IsKiz, // ← 1 или 0
		)
		if err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, nil
}

// SaveInterfirmMoving вызывает USP_INTERFIRM_MOVING_SAVE с XML
func (r *OfferRepository) SaveInterfirmMoving(ctx context.Context, xmlData string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `
		EXEC USP_INTERFIRM_MOVING_SAVE @XML_DATA = @xml
	`, sql.Named("xml", xmlData))
	if err != nil {
		return fmt.Errorf("failed to save interfirm moving: %w", err)
	}

	return nil
}

// getContractorName получает имя контрагента по его ID_CONTRACTOR_GLOBAL
func (r *OfferRepository) GetContractorName(ctx context.Context, id string) string {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(r.timeout)*time.Second)
	defer cancel()
	var name string
	err := r.db.QueryRowContext(ctx, `
		SELECT TOP 1 NAME FROM CONTRACTOR 
		WHERE ID_CONTRACTOR_GLOBAL = @id
	`, map[string]interface{}{"id": id}).Scan(&name)
	if err != nil {
		log.Printf("Failed to get contractor name for %s: %v", id, err)
		return ""
	}
	return name
}
