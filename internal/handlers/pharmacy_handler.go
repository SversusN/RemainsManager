package handlers

import (
	"encoding/json"
	"net/http"

	"RemainsManager/internal/services"
)

type PharmacyHandler struct {
	service *services.PharmacyService
}

func NewPharmacyHandler(service *services.PharmacyService) *PharmacyHandler {
	return &PharmacyHandler{service: service}
}

func (h *PharmacyHandler) GetPharmacies(w http.ResponseWriter, r *http.Request) {
	pharmacies, err := h.service.GetPharmacies()
	if err != nil {
		http.Error(w, "Failed to fetch pharmacies", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pharmacies)
}
