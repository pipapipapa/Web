package main

import (
	"log"
	"orbit-calc/internal/app/config"
	"orbit-calc/internal/app/dsn"
	"orbit-calc/internal/app/handler"
	"orbit-calc/internal/app/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	postgresDSN := dsn.FromEnv()
	if postgresDSN == "" {
		log.Fatal("DB connection string is not set")
	}

	repo, err := repository.New(postgresDSN)
	if err != nil {
		log.Fatalf("failed to initialize repository: %v", err)
	}

	h := handler.NewHandler(repo)

	router := gin.Default()

	h.RegisterAPI(router)

	serverAddress := cfg.ServiceHost + ":" + strconv.Itoa(cfg.ServicePort)

	log.Printf("Starting server at %s", serverAddress)
	if err := router.Run(serverAddress); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
