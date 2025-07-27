package main

import (
	"RemainsManager/internal/middleware"
	"context"
	swagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"
	"time"

	"RemainsManager/config"
	"RemainsManager/internal/handlers"
	_ "RemainsManager/internal/middleware"
	"RemainsManager/internal/repositories"
	"RemainsManager/internal/services"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/microsoft/go-mssqldb"
)

// @title           REST API для управления остатками
// @version         1.0
// @description     API для работы с остатками, пользователями и контрагентами
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
func main() {

	cfg := config.LoadConfig("config.yaml")

	// Подключение к БД
	connString := "sqlserver://" + cfg.Database.User + ":" + cfg.Database.Password +
		"@" + cfg.Database.Host +
		"?database=" + cfg.Database.Name + "&encrypt=disable"
	db, err := sql.Open("sqlserver", connString)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.Timeout)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Миграции
	m, err := migrate.New("file://migrations", connString)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Printf("Migration failed:", err)
	}
	// Инициализация репозиториев
	authRepo := repositories.NewAuthRepository(cfg.Database.Timeout, db)
	userRepo := repositories.NewUserRepository(cfg.Database.Timeout, db)
	pharmacyRepo := repositories.NewPharmacyRepository(cfg.Database.Timeout, db)
	productsRepo := repositories.NewProductRepository(cfg.Database.Timeout, db)
	routsRepo := repositories.NewRouteRepository(cfg.Database.Timeout, db)
	defer db.Close()
	// Инициализация сервисов
	authService := services.NewAuthService(authRepo, cfg.Security.JWTSecret)
	userService := services.NewUserService(userRepo)
	pharmacyService := services.NewPharmacyService(pharmacyRepo)
	productService := services.NewProductService(productsRepo)
	routeService := services.NewRouteService(routsRepo)

	// Инициализация хендлеров
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	pharmacyHandler := handlers.NewPharmacyHandler(pharmacyService)
	productHandler := handlers.NewProductHandler(productService)
	routeHandler := handlers.NewRouteHandler(routeService)

	r := chi.NewRouter()
	r.Use(middleware.EnableCORS)
	r.Post("/login", authHandler.Login)
	r.Get("/users", userHandler.GetAllUsers)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Get("/pharmacies", pharmacyHandler.GetPharmacies)
		r.Get("/inactive-products", productHandler.GetInactiveStockProducts)
		r.Get("/products-with-sales-speed", productHandler.GetProductStockWithSalesSpeed)

		// Маршруты
		r.Post("/routes", routeHandler.CreateRoute)
		r.Get("/routes", routeHandler.GetRoutes)
		r.Delete("/routes/{id}", routeHandler.DeleteRoute)

		// Пункты маршрута
		r.Post("/route-items", routeHandler.AddRouteItem)
		r.Get("/route-items", routeHandler.GetRouteItems)
		r.Delete("/route-items/{id}", routeHandler.DeleteRouteItem)
	})
	// Инициализация Swagger
	r.Group(func(r chi.Router) {
		r.Get("/swagger/*", swagger.Handler(
			swagger.URL("http://localhost:8080/swagger/doc.json"),
		))
	})

	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("Server running on %s", cfg.Server.Port)
	log.Fatal(server.ListenAndServe())
}
