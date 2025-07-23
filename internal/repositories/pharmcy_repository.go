package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"RemainsManager/internal/models"
)

type PharmacyRepository struct {
	db      *sql.DB
	timeout int
}

func NewPharmacyRepository(timeout int, db *sql.DB) *PharmacyRepository {
	return &PharmacyRepository{timeout: timeout, db: db}
}

func (r *PharmacyRepository) GetPharmacies() ([]models.Pharmacy, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()
	rows, err := r.db.QueryContext(ctx, "EXEC GetPharmacies")
	if err != nil {
		return nil, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var pharmacies []models.Pharmacy
	for rows.Next() {
		var p models.Pharmacy
		err := rows.Scan(&p.ID_CONTRACTOR_GLOBAL, &p.Name, &p.Address, &p.Phone, &p.INN)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		pharmacies = append(pharmacies, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return pharmacies, nil
}
