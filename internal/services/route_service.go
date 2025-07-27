package services

import (
	"RemainsManager/internal/models"
	"RemainsManager/internal/repositories"
)

type RouteService struct {
	repo *repositories.RouteRepository
}

func NewRouteService(repo *repositories.RouteRepository) *RouteService {
	return &RouteService{repo: repo}
}

func (s *RouteService) CreateRoute(name string) (int64, error) {
	return s.repo.CreateRoute(name)
}

func (s *RouteService) GetRoutes() ([]models.Route, error) {
	return s.repo.GetRoutes()
}

func (s *RouteService) DeleteRoute(id int64) error {
	return s.repo.DeleteRoute(id)
}

func (s *RouteService) AddRouteItem(routeID int64, contractorGlobal string, order int, name string) (int64, error) {
	return s.repo.AddRouteItem(routeID, contractorGlobal, order, name)
}

func (s *RouteService) GetRouteItems(routeID int64) ([]models.RouteItem, error) {
	return s.repo.GetRouteItems(routeID)
}

func (s *RouteService) DeleteRouteItem(id int64) error {
	return s.repo.DeleteRouteItem(id)
}
