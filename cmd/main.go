package main

import (
	"RemainsManager/config"
	"RemainsManager/internal/handlers"
	"RemainsManager/internal/middleware"
	"RemainsManager/internal/repositories"
	"RemainsManager/internal/services"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/microsoft/go-mssqldb"
	swagger "github.com/swaggo/http-swagger"
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
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Проверка подключения с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Database.Timeout)*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}

	// Миграции
	m, err := migrate.New("file://migrations", connString)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Migration error: %v", err)
	}

	// Инициализация репозиториев
	authRepo := repositories.NewAuthRepository(cfg.Database.Timeout, db)
	userRepo := repositories.NewUserRepository(cfg.Database.Timeout, db)
	pharmacyRepo := repositories.NewPharmacyRepository(cfg.Database.Timeout, db)
	productsRepo := repositories.NewProductRepository(cfg.Database.Timeout, db)
	routsRepo := repositories.NewRouteRepository(cfg.Database.Timeout, db)
	offerRepo := repositories.NewOfferRepository(cfg.Database.Timeout, db)

	// Инициализация сервисов
	authService := services.NewAuthService(authRepo, cfg.Security.JWTSecret)
	userService := services.NewUserService(userRepo)
	pharmacyService := services.NewPharmacyService(pharmacyRepo)
	productService := services.NewProductService(productsRepo)
	routeService := services.NewRouteService(routsRepo)
	offerService := services.NewOfferService(offerRepo)

	// Инициализация хендлеров
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	pharmacyHandler := handlers.NewPharmacyHandler(pharmacyService)
	productHandler := handlers.NewProductHandler(productService)
	routeHandler := handlers.NewRouteHandler(routeService)
	offerHandler := handlers.NewOfferHandler(offerService)

	// Роутер
	r := chi.NewRouter()
	r.Use(middleware.EnableCORS)
	r.Post("/login", authHandler.Login)
	r.Get("/users", userHandler.GetAllUsers)

	// Защищённые маршруты
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
		//r.Get("/route-items/{id}", routeHandler.GetRouteItems)
		r.Delete("/route-items/{id}", routeHandler.DeleteRouteItem)
		r.Put("/routes/{id}/items", routeHandler.UpdateRouteItems)

		//Заявки
		r.Get("/offer", offerHandler.GetOrCreateOffer)
		r.Post("/offer-items", offerHandler.AddOfferItems)

		// Журнал и детали
		r.Get("/offers/journal", offerHandler.GetOfferJournal)
		r.Get("/offers/{id}/details", offerHandler.GetOfferDetails)
		r.Put("/offer-items/{id}", offerHandler.UpdateOfferItem)
		r.Delete("/offer-items/{id}", offerHandler.DeleteOfferItem)
	})
	// Swagger
	r.Group(func(r chi.Router) {
		r.Get("/swagger/*", swagger.Handler(
			swagger.URL("http://localhost:8080/swagger/doc.json"),
		))
	})

	// HTTP-сервер
	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("Server is running on %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Канал для перехвата сигналов (Ctrl+C, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Ожидаем сигнал остановки
	<-stop
	log.Println("Shutting down server gracefully...")

	// Контекст для graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем HTTP-сервер
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced shutdown: %v", err)
	} else {
		log.Println("Server stopped gracefully")
	}

	// Закрываем соединение с БД
	if err := db.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
	} else {
		log.Println("Database connection closed")
	}
}
