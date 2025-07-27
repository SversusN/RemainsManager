package models

// Route представляет маршрут
type Route struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// RouteItem представляет пункт маршрута
type RouteItem struct {
	ID               int64  `json:"id"`
	RouteID          int64  `json:"route_id"`
	ContractorGlobal string `json:"contractor_global"`
	DisplayOrder     int    `json:"display_order"`
	Name             string `json:"name"`
}
