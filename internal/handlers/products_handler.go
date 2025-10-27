package handlers

import (
	"RemainsManager/internal/models"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"net/http"
	"strconv"
	"time"

	"RemainsManager/internal/services"
)

type ProductHandler struct {
	service *services.ProductService
}

func NewProductHandler(service *services.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

// GetInactiveStockProducts godoc
// @Summary		Получить товары без движения
// @Description	Возвращает список товаров, которые не продавались N дней
// @Tags			products
// @Accept			json
// @Produce		json
//
// @Param			contractor_id	query	string	true	"ID контрагента"
// @Param			days			query	int		false	"Количество дней без движения"	default(30)
// @Param			page			query	int		false	"Номер страницы"					default(1)
// @Param			limit			query	int		false	"Размер страницы"					default(50)
// @Param			name			query	string	false	"Фильтр по наименованию (частичное совпадение)"
//
// @Success		200	{array}	models.InactiveStockProduct
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/inactive-products [get]
func (h *ProductHandler) GetInactiveStockProducts(w http.ResponseWriter, r *http.Request) {
	contractorID := r.URL.Query().Get("contractor_id")
	daysStr := r.URL.Query().Get("days")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	nameFilter := r.URL.Query().Get("name")

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

	var namePtr *string
	if nameFilter != "" {
		namePtr = &nameFilter
	}

	products, totalPages, err := h.service.GetInactiveStockProducts(contractorID, days, page, limit, namePtr)
	if err != nil {
		http.Error(w, "Failed to fetch inactive products", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Total-Pages", strconv.Itoa(totalPages))
	w.Header().Set("Access-Control-Expose-Headers", "X-Total-Pages")
	json.NewEncoder(w).Encode(products)
}

func (h *ProductHandler) GetProductStockWithSalesSpeed(w http.ResponseWriter, r *http.Request) {
	contractorID := r.URL.Query().Get("contractor_id")
	daysStr := r.URL.Query().Get("days")
	goodsIDStr := r.URL.Query().Get("goods_id")
	speedOrRoutStr := r.URL.Query().Get("speed_or_rout")

	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}
	speedOrRout := 0
	if s, err := strconv.Atoi(speedOrRoutStr); err == nil && s > 0 {
		speedOrRout = s
	}

	if contractorID == "" {
		http.Error(w, "missing contractor_id", http.StatusBadRequest)
		return
	}

	products, err := h.service.GetProductStockWithSalesSpeed(contractorID, days, goodsIDStr, speedOrRout)
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

// ExportInactiveStockProductsExcel godoc
// @Summary		Экспорт неактивных товаров в Excel
// @Description	Возвращает файл .xlsx с товарами без движения
// @Tags			products
// @Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Param			contractor_id	query		string	true	"ID контрагента"
// @Param			days			query		int		false	"Количество дней без движения"	default(30)
// @Success		200	{file}	file
// @Failure		400	{object}	map[string]string
// @Failure		500	{object}	map[string]string
// @Security		ApiKeyAuth
// @Router			/inactive-products/export [get]
func (h *ProductHandler) ExportInactiveStockProductsExcel(w http.ResponseWriter, r *http.Request) {
	contractorID := r.URL.Query().Get("contractor_id")
	daysStr := r.URL.Query().Get("days")
	limitStr := r.URL.Query().Get("limit")

	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	if contractorID == "" {
		http.Error(w, "missing contractor_id", http.StatusBadRequest)
		return
	}
	limit := 1000
	if s, err := strconv.Atoi(limitStr); err == nil && s > 0 {
		limit = s
	}

	// Получаем все данные (ограничим лимит, например, 1000)
	products, _, err := h.service.GetInactiveStockProducts(contractorID, days, 1, limit, nil)
	if err != nil {
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}

	// Создаём Excel-файл
	file := excelize.NewFile()
	sheet := "Данные"
	file.SetSheetName("Sheet1", sheet)

	// Заголовки
	headers := []string{"Наименование", "Остаток", "Цена продажи", "Себестоимость", "Дней без движения", "Срок годности", "ID товара"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		file.SetCellValue(sheet, cell, h)
	}

	// Данные
	for rIdx, p := range products {
		row := rIdx + 2
		file.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.Name)
		file.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.Qty)
		file.SetCellValue(sheet, fmt.Sprintf("C%d", row), p.PriceSal)
		file.SetCellValue(sheet, fmt.Sprintf("D%d", row), p.PriceProd)
		file.SetCellValue(sheet, fmt.Sprintf("E%d", row), p.DaysNoMovement)
		file.SetCellValue(sheet, fmt.Sprintf("F%d", row), p.BestBefore)
	}

	// Автоширина колонок (опционально)
	_ = file.SetColWidth(sheet, "A", "G", 20)

	// Подготовка к отправке
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="inactive_products_%s.xlsx"`, time.Now().Format("2006-01-02")))

	// Сохраняем и отправляем
	if err := file.Write(w); err != nil {
		log.Printf("Failed to write Excel file: %v", err)
		http.Error(w, "Failed to generate file", http.StatusInternalServerError)
	}
}
