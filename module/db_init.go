package module

import (
	"golang_training/models"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("queue.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Migrate the schema for Queue and User models
	if err := DB.AutoMigrate(&models.Queue{}, &models.User{}); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	log.Println("Database connection and migration successful.")
}
