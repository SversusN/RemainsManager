package repositories

import (
	"database/sql"
	"fmt"

	"RemainsManager/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	rows, err := r.db.Query("EXEC GetUsers")
	if err != nil {
		return nil, fmt.Errorf("error executing stored procedure: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.Name, &u.FullName, &u.UserNum)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
