package main

import (
	"context"
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"gophermart/config"
	"gophermart/internal/services"
	"gophermart/internal/storage"
	"gophermart/internal/storage/repos"
	"gophermart/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gophermart/internal/handlers"
	"gophermart/internal/middlewares"
	"gophermart/internal/routes"
)

var cfg = config.NewServerConfig()

func main() {
	if err := logger.Init("INFO"); err != nil {
		panic(err)
	}
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	authMW := middlewares.AuthMiddleware([]byte(jwtSecret))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	db, err := storage.NewDB(storage.Config{
		DBCfg:      cfg.DBCfg,
		Migrations: cfg.Migrations,
	})
	if err != nil {
		logger.Log.Error("Failed to initialize database", zap.Error(err))
		os.Exit(1)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Log.Info("Error DB close", zap.String("address", cfg.RunAddr))
		}
	}(db)

	r := gin.Default()

	postgresRepo := repos.NewPostgresUserRepository(db)
	authService := services.NewAuthService(postgresRepo, jwtSecret)
	authHandler := handlers.New(authService)
	routes.SetupUserRoutes(r, authHandler)

	orderRepo := repos.NewPostgresOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, cfg.AccrualAddress)

	if err := orderService.RecoverPendingOrders(context.Background()); err != nil {
		log.Fatalf("Failed to recover pending orders: %v", err)
	}

	orderHandler := handlers.NewOrderHandler(orderService)

	routes.SetupOrderRoutes(r, orderHandler, authMW)

	balanceRepo := repos.NewPostgresBalanceRepository(db)
	balanceService := services.NewBalanceService(balanceRepo, 50)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	routes.SetupBalanceRoutes(r, balanceHandler, authMW)

	withdrawalRepo := repos.NewPostgresWithdrawalRepository(db)
	withdrawalService := services.NewWithdrawalService(withdrawalRepo)
	withdrawalHandler := handlers.NewWithdrawalHandler(withdrawalService)

	routes.SetupWithdrawalRoutes(r, withdrawalHandler, authMW)

	logger.Log.Info("Server is running at ", zap.String("address", cfg.RunAddr))

	srv := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	serverErr := make(chan error, 1)

	go func() {
		logger.Log.Info("Server is running", zap.String("address", cfg.RunAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Log.Info("Shutting down server...")
	case err := <-serverErr:
		logger.Log.Error("Server error", zap.Error(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server is shutdown", zap.Error(err))
	}

	logger.Log.Info("Server stopped gracefully")
}
