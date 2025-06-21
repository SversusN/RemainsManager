package main

import (
	"RemainsManager/internal/middleware"
	"log"
	"net/http"
	"time"

	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlserver"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/microsoft/go-mssqldb"

	"RemainsManager/config"
	"RemainsManager/internal/handlers"
	_ "RemainsManager/internal/middleware"
	"RemainsManager/internal/repositories"
	"RemainsManager/internal/services"
)

func main() {

	cfg := config.LoadConfig("config.yaml")

	// Подключение к БД
	connString := "sqlserver://" + cfg.Database.User + ":" + cfg.Database.Password +
		"@" + cfg.Database.Host +
		"?database=" + cfg.Database.Name + "&encrypt=disable"
	db, err := sql.Open("sqlserver", connString)
	defer db.Close()
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
	authRepo := repositories.NewAuthRepository(db)
	//userRepo := repositories.NewUserRepository(db)

	// Инициализация сервисов
	authService := services.NewAuthService(authRepo, cfg.Security.JWTSecret)
	//userService := services.NewUserService(userRepo)

	// Инициализация хендлеров
	authHandler := handlers.NewAuthHandler(authService)
	//userHandler := handlers.NewUserHandler(userService)

	// Роутинг
	r := chi.NewRouter()

	r.Post("/login", authHandler.Login)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

	})

	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server running on %s", cfg.Server.Port)
	log.Fatal(server.ListenAndServe())
}
