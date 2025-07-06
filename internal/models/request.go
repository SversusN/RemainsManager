package models

// AuthRequest представляет данные для авторизации
type AuthRequest struct {
	// Логин пользователя
	// Example: admin
	Username string `json:"username"`

	// Пароль пользователя
	// Example: password123
	Password string `json:"password"`
}
