package main

import (
	"log"

	"github.com/abdullahalsazib/e-com-backend/config"
	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/abdullahalsazib/e-com-backend/routes"
	"github.com/abdullahalsazib/e-com-backend/seed"
	"github.com/joho/godotenv"
)

func main() {
	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// connect to db
	config.ConnectDB()
	db := config.DB
	db.AutoMigrate(
		&models.Product{},
		&models.User{},
		&models.Token{},
		&models.Cart{},
		&models.CartItem{},
		&models.Category{},
		&models.Order{},
		&models.OrderItem{},
		&models.WishlistItem{},
	)

	// seed category
	seed.SeedCategories(db)
	// setup models
	r := routes.SetupRoutes(db)

	r.Run(":8080")
}
