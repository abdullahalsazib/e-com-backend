package main

import (
	"log"
	"os"

	"github.com/abdullahalsazib/e-com-backend/config"
	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/abdullahalsazib/e-com-backend/routes"
	"github.com/abdullahalsazib/e-com-backend/seed"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// if production then use this releaseMode
	gin.SetMode(gin.ReleaseMode)


	// load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// connect to db
	config.ConnectDB()
	db := config.DB
	db.AutoMigrate(
		// Add Role model migration
		&models.User{},
		&models.Vendor{},
		&models.AuditLog{},
		&models.Token{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
		&models.UserRole{},

		// Add other models
		&models.Product{},
		&models.Cart{},
		&models.CartItem{},
		&models.Category{},
		&models.Order{},
		&models.OrderItem{},
		&models.WishlistItem{},
	)

	// seed category
	seed.SeedCategories(db)
	// seed roles and super admin
	seed.SeedRoles(db)
	seed.SeedSuperAdmin(db)

	// setup models
	r := routes.SetupRoutes(db)

	// port := os.Getenv("PORT")
	port := os.Getenv("PORT")
	if port == "" {
		port = "10000"
	}

	// Start server
	err = r.Run(":" + port)
	if err != nil {
		log.Fatal("Server failed to start:", err)
	}

	// r.Run(port)
}
