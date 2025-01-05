package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"example.com/monolithic/configs"
	"example.com/monolithic/internal/core/services"
	"example.com/monolithic/internal/handlers"
	custommw "example.com/monolithic/internal/middleware"
	"example.com/monolithic/internal/platform/database"
	"example.com/monolithic/internal/platform/database/migrations"
	"example.com/monolithic/internal/repositories"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "APP: ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database configuration
	dbConfig := database.Config{
		Host:        cfg.Database.Host,
		Port:        cfg.Database.Port,
		User:        cfg.Database.User,
		Password:    cfg.Database.Password,
		Database:    cfg.Database.DBName,
		MaxPoolSize: 10,
		MinPoolSize: 2,
		MaxIdleTime: 15 * time.Minute,
		MaxLifetime: 1 * time.Hour,
		HealthCheck: 30 * time.Second,
		SSLMode:     cfg.Database.SSLMode,
	}

	// Initialize database connection
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database health check
	if err := db.Ping(context.Background()); err != nil {
		logger.Fatalf("Database health check failed: %v", err)
	}
	logger.Println("Successfully connected to database")

	// Run db migrations
	if err := migrations.RunMigrations(dbConfig.GetConnectionURL()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	//productRepo := repositories.NewProductRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo)
	//productService := services.NewProductService(productRepo)

	// Initialize HTTP handlers
	userHandler := handlers.NewUserHandler(userService)
	//productHandler := handlers.NewProductHandler(productService)

	// Create Chi router
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second)) // maximum duration of 60 seconds for all HTTP requests handled by your server
	r.Use(custommw.CORS)
	r.Use(custommw.Authentication)

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Users endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.Timeout(200 * time.Second)) // route specific middleware
			r.Mount("/users", userHandler.Routes())
		})
	})

	// Create server
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		ErrorLog:     logger,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, _ := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				logger.Fatal("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			logger.Printf("Shutdown error: %v\n", err)
		}
		serverStopCtx()
	}()

	// Start server
	logger.Printf("Server is starting on %s\n", srv.Addr)
	err = srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatalf("Server error: %v\n", err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	logger.Println("Server stopped gracefully")
}
