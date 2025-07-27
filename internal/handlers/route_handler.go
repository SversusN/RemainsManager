package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"

	"RemainsManager/internal/models"
	"RemainsManager/internal/services"
)

type RouteHandler struct {
	service *services.RouteService
}

func NewRouteHandler(service *services.RouteService) *RouteHandler {
	return &RouteHandler{service: service}
}

// CreateRoute godoc
// @Summary		Создать маршрут
// @Description	Создаёт новый маршрут
// @Tags			routes
// @Accept			json
// @Produce		json
// @Param			body	body		models.Route	true	"Название маршрута"
// @Success		201	{object}	map[string]int64
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/routes [post]
func (h *RouteHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	var req models.Route
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateRoute(req.Name)
	if err != nil {
		http.Error(w, "Failed to create route", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetRoutes godoc
// @Summary		Получить все маршруты
// @Description	Возвращает список всех маршрутов
// @Tags			routes
// @Produce		json
// @Success		200	{array}	models.Route
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/routes [get]
func (h *RouteHandler) GetRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.service.GetRoutes()
	if err != nil {
		http.Error(w, "Failed to fetch routes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routes)
}

// DeleteRoute godoc
// @Summary		Удалить маршрут
// @Description	Удаляет маршрут по ID
// @Tags			routes
// @Param			id	path		int	true	"ID маршрута"
// @Success		204
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/routes/{id} [delete]
func (h *RouteHandler) DeleteRoute(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteRoute(id); err != nil {
		http.Error(w, "Failed to delete route", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddRouteItem godoc
// @Summary		Добавить пункт маршрута
// @Description	Добавляет контрагента в маршрут
// @Tags			route-items
// @Accept			json
// @Produce		json
// @Param			body	body		models.RouteItem	true	"Пункт маршрута"
// @Success		201	{object}	map[string]int64
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/route-items [post]
func (h *RouteHandler) AddRouteItem(w http.ResponseWriter, r *http.Request) {
	var req models.RouteItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.service.AddRouteItem(req.RouteID, req.ContractorGlobal, req.DisplayOrder, req.Name)
	if err != nil {
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

// GetRouteItems godoc
// @Summary		Получить пункты маршрута
// @Description	Возвращает все пункты указанного маршрута
// @Tags			route-items
// @Produce		json
// @Param			route_id	query		int	true	"ID маршрута"
// @Success		200	{array}	models.RouteItem
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/route-items [get]
func (h *RouteHandler) GetRouteItems(w http.ResponseWriter, r *http.Request) {
	routeIDStr := r.URL.Query().Get("route_id")
	routeID, err := strconv.ParseInt(routeIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid route_id", http.StatusBadRequest)
		return
	}

	items, err := h.service.GetRouteItems(routeID)
	if err != nil {
		http.Error(w, "Failed to fetch items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// DeleteRouteItem godoc
// @Summary		Удалить пункт маршрута
// @Description	Удаляет пункт маршрута по ID
// @Tags			route-items
// @Param			id	path		int	true	"ID пункта маршрута"
// @Success		204
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/route-items/{id} [delete]
func (h *RouteHandler) DeleteRouteItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteRouteItem(id); err != nil {
		http.Error(w, "Failed to delete item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
