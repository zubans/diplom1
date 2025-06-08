package main

import (
	"gophermart/config"
	"gophermart/internal/repos"
	"gophermart/internal/services"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gophermart/internal/handlers"
	"gophermart/internal/initialize"
	"gophermart/internal/middlewares"
	"gophermart/internal/routes"
)

var cfg = config.NewServerConfig()

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	runAddress := os.Getenv("RUN_ADDRESS")
	if runAddress == "" {
		runAddress = ":8080"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	authMW := middlewares.AuthMiddleware([]byte(jwtSecret))
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	accrualURL := os.Getenv("ACCRUAL_URL")
	if accrualURL == "" {
		log.Fatal("ACCRUAL_URL must be set")
	}

	db := initialize.InitDB(cfg)
	defer db.Close()

	r := gin.Default()

	postgresRepo := repos.NewPostgresUserRepository(db)
	authService := services.NewAuthService(postgresRepo, jwtSecret)
	authHandler := handlers.New(authService)
	routes.SetupUserRoutes(r, authHandler)

	orderRepo := repos.NewPostgresOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, cfg.Accrual_address)
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

	log.Printf("Server is running at %s", runAddress)
	if err := r.Run(runAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
