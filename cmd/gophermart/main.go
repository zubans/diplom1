package main

import (
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

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	runAddress := os.Getenv("RUN_ADDRESS")
	if runAddress == "" {
		runAddress = ":8080"
	}
	//dbURI := os.Getenv("DATABASE_URI")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	accrualURL := os.Getenv("ACCRUAL_URL")
	if accrualURL == "" {
		log.Fatal("ACCRUAL_URL must be set")
	}

	db := initialize.InitDB()
	defer db.Close()

	r := gin.Default()

	userRepo := repos.NewPostgresUserRepository(db)
	authService := services.NewAuthService(userRepo, jwtSecret)
	authHandler := handlers.New(authService)
	routes.SetupUserRoutes(r, authHandler)

	orderRepo := repos.NewPostgresOrderRepository(db)
	orderService := services.NewOrderService(orderRepo, accrualURL)
	orderHandler := handlers.NewOrderHandler(orderService)
	authMW := middlewares.AuthMiddleware([]byte(jwtSecret))

	routes.SetupOrderRoutes(r, orderHandler, authMW)

	log.Printf("Server is running at %s", runAddress)
	if err := r.Run(runAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
