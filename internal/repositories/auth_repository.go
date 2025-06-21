package repositories

import (
	"database/sql"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) GetUserByUsername(username string) (string, error) {
	var password string
	err := r.db.QueryRow("EXEC GetUserByUsername @Username=@username", sql.Named("username", username)).Scan(&password)
	if err != nil {
		return "", err
	}
	return password, nil
}
