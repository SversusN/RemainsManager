package models

// InactiveStockProduct представляет товар без движения более N дней
// @Description Информация о товаре без движения
// @Description - **Name**: Название товара
// @Description - **Qty**: Остаток на складе
// @Description - **PriceSal**: Цена продажи
// @Description - **PriceProd**: Себестоимость
// @Description - **DaysNoMovement**: Количество дней без движения
// @Description - **BestBefore**: Срок годности (опционально)
type InactiveStockProduct struct {
	Name           string  `json:"name"`
	Qty            float64 `json:"qty"`
	PriceSal       float64 `json:"price_sal"`
	PriceProd      float64 `json:"price_prod"`
	DaysNoMovement int     `json:"days_no_movement"`
	BestBefore     string  `json:"best_before,omitempty"`
}

type ProductStockWithSalesSpeed struct {
	Name        string  `json:"name"`
	Qty         float64 `json:"qty"`
	PriceSal    float64 `json:"price_sal"`
	PriceProd   float64 `json:"price_prod"`
	BestBefore  string  `json:"best_before,omitempty"`
	TotalSold   float64 `json:"total_sold_last_30_days"`
	SalesPerDay float64 `json:"sales_per_day"`
	ActiveDays  int     `json:"active_days"`
}
