package main

import (
	"log"
	"orbit-calc/internal/app/ds"
	"orbit-calc/internal/app/dsn"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	err := godotenv.Load(".env") 
	if err != nil {
		log.Fatalf("Error loading .env file from relative path: %v", err)
	}

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	err = db.AutoMigrate(
		&ds.User{},
		&ds.OrbitType{},
		&ds.Mission{},
		&ds.MissionOrbitItem{},
	)
	if err != nil {
		log.Fatal("cant migrate db: ", err)
	}
	log.Println("Migration successful!")
}