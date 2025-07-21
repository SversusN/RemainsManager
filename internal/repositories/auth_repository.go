package repositories

import (
	"context"
	"database/sql"
	"time"
)

type AuthRepository struct {
	db      *sql.DB
	timeout int
}

func NewAuthRepository(timeout int, db *sql.DB) *AuthRepository {
	return &AuthRepository{timeout: timeout, db: db}
}

func (r *AuthRepository) GetUserByUsername(username string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()
	var password string
	err := r.db.QueryRowContext(ctx, "EXEC GetUserByUsername @Username=@username", sql.Named("username", username)).Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}
