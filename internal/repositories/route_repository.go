package repositories

import (
	"RemainsManager/internal/models"
	"context"
	"database/sql"
	"time"
)

type RouteRepository struct {
	db      *sql.DB
	timeout int // таймаут в секундах
}

func NewRouteRepository(timeout int, db *sql.DB) *RouteRepository {
	return &RouteRepository{db: db, timeout: timeout}
}

// CreateRoute создает маршрут
func (r *RouteRepository) CreateRoute(name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	var id int64
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO ROUTE (NAME) 
        OUTPUT INSERTED.ID_ROUTE 
        VALUES (@p1)`,
		name).Scan(&id)
	return id, err
}

// GetRoutes возвращает все маршруты
func (r *RouteRepository) GetRoutes() ([]models.Route, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
        SELECT ID_ROUTE, NAME 
        FROM ROUTE 
        ORDER BY NAME`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var routes []models.Route
	for rows.Next() {
		var route models.Route
		if err := rows.Scan(&route.ID, &route.Name); err != nil {
			return nil, err
		}
		routes = append(routes, route)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return routes, nil
}

// DeleteRoute удаляет маршрут и все его пункты
func (r *RouteRepository) DeleteRoute(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "DELETE FROM ROUTE WHERE ID_ROUTE = ?", id)
	return err
}

// AddRouteItem добавляет пункт маршрута
func (r *RouteRepository) AddRouteItem(routeID int64, contractorGlobal string, order int, name string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	var id int64
	err := r.db.QueryRowContext(ctx, `
        INSERT INTO ROUTE_ITEM (ID_ROUTE, ID_CONTRACTOR_GLOBAL, DISPLAY_ORDER, NAME)
        OUTPUT INSERTED.ID_ROUTE_ITEM
        VALUES (@p1, @p2, @p3, @p4)`,
		routeID, contractorGlobal, order, name).Scan(&id)
	return id, err
}

// GetRouteItems возвращает пункты маршрута
func (r *RouteRepository) GetRouteItems(routeID int64) ([]models.RouteItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	rows, err := r.db.QueryContext(ctx, `
        SELECT ID_ROUTE_ITEM, ID_ROUTE, ID_CONTRACTOR_GLOBAL, DISPLAY_ORDER, NAME
        FROM ROUTE_ITEM
        WHERE ID_ROUTE = ?
        ORDER BY DISPLAY_ORDER`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.RouteItem
	for rows.Next() {
		var item models.RouteItem
		if err := rows.Scan(&item.ID, &item.RouteID, &item.ContractorGlobal, &item.DisplayOrder, &item.Name); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// DeleteRouteItem удаляет пункт маршрута
func (r *RouteRepository) DeleteRouteItem(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeout)*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, "DELETE FROM ROUTE_ITEM WHERE ID_ROUTE_ITEM = ?", id)
	return err
}
