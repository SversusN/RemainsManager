package models

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
