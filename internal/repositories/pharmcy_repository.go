package repositories

import (
	"database/sql"
	"fmt"

	"RemainsManager/internal/models"
)

type PharmacyRepository struct {
	db *sql.DB
}

func NewPharmacyRepository(db *sql.DB) *PharmacyRepository {
	return &PharmacyRepository{db: db}
}

func (r *PharmacyRepository) GetPharmacies() ([]models.Pharmacy, error) {
	rows, err := r.db.Query("EXEC GetPharmacies")
	if err != nil {
		return nil, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var pharmacies []models.Pharmacy
	for rows.Next() {
		var p models.Pharmacy
		err := rows.Scan(&p.Name, &p.Address, &p.Phone, &p.INN)
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
