package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"RemainsManager/internal/models"
)

type UserRepository struct {
	db      *sql.DB
	timeout int
}

func NewUserRepository(timeout int, db *sql.DB) *UserRepository {
	return &UserRepository{timeout: timeout, db: db}
}

func (r *UserRepository) GetAllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()
	rows, err := r.db.QueryContext(ctx, "EXEC GetUsers")
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
