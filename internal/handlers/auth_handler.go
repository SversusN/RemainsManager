package handlers

import (
	"RemainsManager/internal/models"
	"encoding/json"
	"net/http"

	"RemainsManager/internal/services"
)

type AuthHandler struct {
	service *services.AuthService
}

func NewAuthHandler(service *services.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login godoc
// @Summary		Вход в систему и получение токена
// @Description	Аутентифицирует пользователя и возвращает Bearer-токен
// @Tags			auth
// @Accept			json
// @Produce		json
//
// @Param			body	body		models.AuthRequest	true	"Логин и пароль"
//
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]string
// @Failure		401	{object}	map[string]string
//
// @Router			/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.service.Authenticate(req.Username, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
