package handlers

import (
	"RemainsManager/internal/models"
	"encoding/json"
	"net/http"
	"strconv"

	"RemainsManager/internal/services"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// GetInactiveStockProducts godoc
//
// @Summary      Получить неактивные товары
// @Description  Возвращает список товаров, которые не продавались дольше N дней
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        contractor_id query string true "ID контрагента (GUID)"
// @Param        days query integer false "Количество дней без движения" default(30)
// @Param        page query integer false "Номер страницы" default(1)
// @Param        limit query integer false "Размер страницы" default(50)
// @Success      200 {object} models.InactiveStockProduct
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Failure		401	{object}	map[string]string
// @Security     ApiKeyAuth
// @Router       /api/inactive-products [get]
func (h *ProductHandler) GetInactiveStockProducts(w http.ResponseWriter, r *http.Request) {
	contractorID := r.URL.Query().Get("contractor_id")
	daysStr := r.URL.Query().Get("days")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	days := 30
	page := 1
	limit := 50

	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	if contractorID == "" {
		http.Error(w, "missing contractor_id", http.StatusBadRequest)
		return
	}

	products, totalCount, err := h.service.GetInactiveStockProducts(contractorID, days, page, limit)
	if err != nil {
		http.Error(w, "failed to fetch inactive products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Count", strconv.Itoa(totalCount))
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductStockWithSalesSpeed(w http.ResponseWriter, r *http.Request) {
	contractorID := r.URL.Query().Get("contractor_id")
	daysStr := r.URL.Query().Get("days")
	goodsIDStr := r.URL.Query().Get("goods_id")

	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	var goodsID *int64
	if goodsIDStr != "" {
		id, _ := strconv.ParseInt(goodsIDStr, 10, 64)
		goodsID = &id
	}

	if contractorID == "" {
		http.Error(w, "missing contractor_id", http.StatusBadRequest)
		return
	}

	products, err := h.service.GetProductStockWithSalesSpeed(contractorID, days, goodsID)
	if err != nil {
		http.Error(w, "failed to fetch product stock", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if products == nil {
		products = make([]models.ProductStockWithSalesSpeed, 0)
	}
	json.NewEncoder(w).Encode(products)
}
