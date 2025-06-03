package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gophermart/internal/handlers"
	"gophermart/internal/initialize"
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

	db := initialize.InitDB()
	defer db.Close()

	r := gin.Default()

	userHandler := handlers.New(db, jwtSecret)
	routes.SetupUserRoutes(r, userHandler)

	log.Printf("Server is running at %s", runAddress)
	if err := r.Run(runAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
