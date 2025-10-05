package handlers

import (
	"RemainsManager/internal/models"
	"RemainsManager/internal/services"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type OfferHandler struct {
	service *services.OfferService
}

func NewOfferHandler(service *services.OfferService) *OfferHandler {
	return &OfferHandler{service: service}
}

// GetOrCreateOffer godoc
// @Summary		Получить или создать заявку на сегодня
// @Description	Возвращает активную заявку или создаёт новую
// @Tags			offers
// @Accept			json
// @Produce		json
// @Param			from_id	query		string	true	"ID контрагента-отправителя"
// @Param			from_name	query		string	true	"Наименование контрагента"
// @Success		200	{object}	models.Offer
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offers/today [get]
func (h *OfferHandler) GetOrCreateOffer(w http.ResponseWriter, r *http.Request) {
	fromID := r.URL.Query().Get("from_id")
	fromName := r.URL.Query().Get("from_name")

	if fromID == "" || fromName == "" {
		http.Error(w, "missing from_id or from_name", http.StatusBadRequest)
		return
	}

	offer, err := h.service.GetOrCreateTodayOffer(r.Context(), fromID, fromName)
	if err != nil {
		http.Error(w, "Failed to get or create offer", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offer)
}

// AddOfferItems godoc
// @Summary		Добавить несколько позиций в заявку
// @Description	Добавляет массив товаров в текущую заявку
// @Tags			offers
// @Accept			json
// @Produce		json
// @Param			body	body		[]models.OfferItem	true	"Массив позиций заявки"
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offer-items/bulk [post]
func (h *OfferHandler) AddOfferItems(w http.ResponseWriter, r *http.Request) {
	var items []models.OfferItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(items) == 0 {
		http.Error(w, "no items provided", http.StatusBadRequest)
		return
	}

	for _, item := range items {
		if item.OfferID == 0 {
			http.Error(w, "offer_id is required for all items", http.StatusBadRequest)
			return
		}
	}

	if err := h.service.AddItems(r.Context(), items); err != nil {
		http.Error(w, "Failed to add items: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "added": strconv.Itoa(len(items))})
}

// GetOfferJournal godoc
// @Summary		Получить журнал заявок
// @Description	Возвращает список заявок с фильтрацией по дате и контрагенту
// @Tags			offers
// @Produce		json
// @Param			from				query		string	false	"Дата начала (YYYY-MM-DD)"	default(2025-01-01)
// @Param			to					query		string	false	"Дата окончания (YYYY-MM-DD)"	default(2025-12-31)
// @Param			contractor_id		query		string	false	"ID контрагента-отправителя (GUID)"
// @Success		200	{array}	models.OfferJournalItem
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offers/journal [get]
func (h *OfferHandler) GetOfferJournal(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	contractorID := r.URL.Query().Get("contractor_id")

	from, err := parseDate(fromStr, 7) // по умолчанию — 7 дней назад
	if err != nil {
		http.Error(w, "invalid 'from' date", http.StatusBadRequest)
		return
	}

	to, err := parseDate(toStr, 0) // по умолчанию — сегодня
	if err != nil {
		http.Error(w, "invalid 'to' date", http.StatusBadRequest)
		return
	}

	if from.After(to) {
		http.Error(w, "'from' date must be before or equal to 'to'", http.StatusBadRequest)
		return
	}

	var contractorPtr *string
	if contractorID != "" {
		contractorPtr = &contractorID
	}

	items, err := h.service.GetOfferJournal(r.Context(), from, to, contractorPtr)
	if err != nil {
		http.Error(w, "Failed to fetch journal", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// Вспомогательная функция парсинга даты
func parseDate(dateStr string, daysAgo int) (time.Time, error) {
	if dateStr == "" {
		t := time.Now().AddDate(0, 0, -daysAgo)
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()), nil
	}
	return time.Parse("2006-01-02", dateStr)
}

// GetOfferDetails godoc
// @Summary		Получить детали заявки
// @Description	Возвращает список позиций указанной заявки
// @Tags			offers
// @Produce		json
// @Param			id	path		int	true	"ID заявки"
// @Success		200	{array}	models.OfferDetailItem
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offers/{id}/details [get]
func (h *OfferHandler) GetOfferDetails(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid offer ID", http.StatusBadRequest)
		return
	}

	items, err := h.service.GetOfferDetails(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to fetch details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// UpdateOfferItem godoc
// @Summary		Обновить количество в позиции заявки
// @Description	Изменяет количество товара в существующей позиции
// @Tags			offers
// @Accept			json
// @Produce		json
// @Param			id		path		int				true	"ID позиции заявки"
// @Param			body	body		UpdateQuantityRequest	true	"Новое количество"
// @Success		200	{object}	map[string]string
// @Failure		400	{object}	map[string]string
// @Failure		404	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offer-items/{id} [put]
func (h *OfferHandler) UpdateOfferItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	var req UpdateQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Quantity <= 0 {
		http.Error(w, "Quantity must be greater than 0", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateOfferItem(r.Context(), id, req.Quantity); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Offer item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to update item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

// Вспомогательная структура для запроса
type UpdateQuantityRequest struct {
	Quantity int `json:"quantity"`
}

// DeleteOfferItem godoc
// @Summary		Удалить позицию из заявки
// @Description	Удаляет позицию по ID
// @Tags			offers
// @Param			id	path		int	true	"ID позиции заявки"
// @Success		204
// @Failure		404	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/offer-items/{id} [delete]
func (h *OfferHandler) DeleteOfferItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteOfferItem(r.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Offer item not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete item: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
