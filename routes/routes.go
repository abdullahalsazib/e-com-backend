package routes

import (
	"time"

	"github.com/abdullahalsazib/e-com-backend/controllers"
	"github.com/abdullahalsazib/e-com-backend/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB) *gin.Engine {
	r := gin.Default()

	config := cors.Config{
		AllowOrigins: []string{"http://localhost:3000"},
		// AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(config))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// initilize controllers
	// auth
	authController := controllers.NewAuthController(db)
	// product
	productController := controllers.NewProductController(db)
	// cart
	cartController := controllers.NewCartController(db)
	// order
	orderController := controllers.NewOrderController(db)
	// seller

	// categories
	categoryController := controllers.NewCategoryController(db)

	// wishlist
	wishlistController := controllers.NewWishlistController(db)

	// public route
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)
	r.POST("/refresh", authController.Refresh)
	r.GET("/categorys", categoryController.GetCategories)
	//
	authGroup := r.Group("/auth")
	{
		authGroup.Use(middlewares.AuthMiddleware(db))
		{
			authGroup.GET("/me", authController.GetProfile)
			authGroup.POST("/logout", authController.Logout)
		}

	}
	// product routes
	productGrup := r.Group("/api/v1/products")
	{
		// public route (no auth use)
		productGrup.GET("", productController.GetProducts)
		productGrup.GET("/:id", productController.GetProduct)

		// protected routes (require auth - admin or !)
		protectedGroup := productGrup.Group("")
		protectedGroup.Use(middlewares.AuthMiddleware(db))
		protectedGroup.Use(middlewares.AdminSellerMiddleware())
		{
			// auth is required
			protectedGroup.POST("/", productController.CreateProduct)
			protectedGroup.PUT("/:id", productController.UpdateProduct)
			protectedGroup.DELETE("/:id", productController.DeleteProduct)
		}

	}

	order := authGroup.Group("/orders")
	{
		order.POST("/", orderController.CreateOrder)
		order.GET("/", orderController.GetOrders)
		order.GET("/:orderId", orderController.GetOrder)
		order.PUT("/:orderId/cancel", orderController.CancelOrder)

		// admin only route
		admin := order.Group("/")
		{
			admin.Use(middlewares.AdminMiddleware(db))
			admin.PUT("/:orderId/status", orderController.UpdateOrderStatus)
		}
	}

	wishlistGroup := authGroup.Group("/wishlist")
	{
		wishlistGroup.Use(middlewares.AuthMiddleware(db))
		wishlistGroup.POST("/add", wishlistController.AddToWishlist)
		wishlistGroup.GET("", wishlistController.GetWishlist)
		wishlistGroup.DELETE("/:id", wishlistController.RemoveFromWishlist)
		wishlistGroup.DELETE("/clear", wishlistController.ClearWishlist)
		wishlistGroup.PUT("/update/:id", wishlistController.UpdateWishlistItem)
		wishlistGroup.POST("/import", wishlistController.ImportWishlist)

	}

	cartGroup := authGroup.Group("/cart")
	{
		cartGroup.Use(middlewares.AuthMiddleware(db))
		cartGroup.GET("/", cartController.GetCart)
		cartGroup.POST("/items", cartController.AddToCart)
		cartGroup.PUT("/items/:itemId", cartController.UpdateCartItem)
		cartGroup.DELETE("/items/:itemId", cartController.RemoveFromCart)
		cartGroup.DELETE("/clear", cartController.ClearCart)
	}

	return r
}
