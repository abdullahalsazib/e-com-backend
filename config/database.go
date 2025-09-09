package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	// DATABASE_URL

	// get DATABASE_URL from .env
	// webDsn = os.Getenv("DATABASE_URL")
	dsn := os.Getenv("DATABASE_URL")
	// dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	DB = db

	fmt.Println("Database connected successfully")
}
