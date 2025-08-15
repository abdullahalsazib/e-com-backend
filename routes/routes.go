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

	// ==== CORS CONFIG ====
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// ==== CONTROLLERS ====
	authController := controllers.NewAuthController(db)
	superAdminController := controllers.NewSuperAdminController(db)
	productController := controllers.NewProductController(db)
	cartController := controllers.NewCartController(db)
	orderController := controllers.NewOrderController(db)
	categoryController := controllers.NewCategoryController(db)
	wishlistController := controllers.NewWishlistController(db)
	vendorController := controllers.NewVendorController(db, &authController)

	// ==== PUBLIC ROUTES ====
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)
	r.POST("/refresh", authController.RefreshToken)
	r.POST("/logout", authController.Logout)
	r.GET("/categorys", categoryController.GetCategories)

	// ==== AUTH GROUP ====
	authGroup := r.Group("/auth")
	authGroup.Use(middlewares.AuthMiddleware(db))
	{
		authGroup.GET("/me", authController.GetProfile)
		// authGroup.POST("/logout", authController.Logout)
	}

	// ==== PRODUCT ROUTES ====
	productGroup := r.Group("/api/v1/products")
	{
		// public
		productGroup.GET("", productController.GetProducts)
		productGroup.GET("/:id", productController.GetProduct)

		// protected (admin or seller)
		protectedProduct := productGroup.Group("")
		protectedProduct.Use(middlewares.AuthMiddleware(db), middlewares.AdminSellerMiddleware())
		{
			protectedProduct.POST("/", productController.CreateProduct)
			protectedProduct.PUT("/:id", productController.UpdateProduct)
			protectedProduct.DELETE("/:id", productController.DeleteProduct)
		}
	}

	// ==== ORDER ROUTES ====
	order := authGroup.Group("/orders")
	{
		order.POST("/", orderController.CreateOrder)
		order.GET("/", orderController.GetOrders)
		order.GET("/:orderId", orderController.GetOrder)
		order.PUT("/:orderId/cancel", orderController.CancelOrder)

		// admin only
		adminOrder := order.Group("/")
		adminOrder.Use(middlewares.AdminMiddleware(db))
		{
			adminOrder.PUT("/:orderId/status", orderController.UpdateOrderStatus)
		}
	}

	// ==== WISHLIST ROUTES ====
	wishlistGroup := authGroup.Group("/wishlist")
	{
		wishlistGroup.POST("/add", wishlistController.AddToWishlist)
		wishlistGroup.GET("", wishlistController.GetWishlist)
		wishlistGroup.DELETE("/:id", wishlistController.RemoveFromWishlist)
		wishlistGroup.DELETE("/clear", wishlistController.ClearWishlist)
		wishlistGroup.PUT("/update/:id", wishlistController.UpdateWishlistItem)
		wishlistGroup.POST("/import", wishlistController.ImportWishlist)
	}

	// ==== CART ROUTES ====
	cartGroup := authGroup.Group("/cart")
	{
		cartGroup.GET("/", cartController.GetCart)
		cartGroup.POST("/items", cartController.AddToCart)
		cartGroup.PUT("/items/:itemId", cartController.UpdateCartItem)
		cartGroup.DELETE("/items/:itemId", cartController.RemoveFromCart)
		cartGroup.DELETE("/clear", cartController.ClearCart)
	}
	// ==== SUPER ADMIN ROUTES ====

	superAdmin := r.Group("/super-admin")
	superAdmin.Use(middlewares.AuthMiddleware(db), middlewares.SuperAdminMiddleware(db))
	{
		superAdmin.GET("/users", superAdminController.ListUsers)
		superAdmin.DELETE("/users/:id", superAdminController.DeleteUserByID)
	}

	// ==== PRODUCT CATEGORY ROUTES ====
	categoryRoutes := r.Group("/categories")
	{
		categoryRoutes.GET("/", categoryController.GetCategories)
		categoryRoutes.GET("/:id", categoryController.GetCategories)
	}

	// ==== SUPERADMIN  MANAGEMENT ====
	superAdminGroup := r.Group("/super-admin")
	superAdminGroup.Use(middlewares.AuthMiddleware(db), middlewares.SuperAdminMiddleware(db))
	{
		superAdminGroup.GET("/users", superAdminController.ListUsers)
		superAdminGroup.DELETE("/users/:id", superAdminController.DeleteUserByID)
		// superAdminGroup.GET("/users/:id", authController.GetUser)
		// superAdminGroup.PUT("/users/:id/role", authController.UpdateUserRole)
		// superAdminGroup.DELETE("/users/:id", authController.DeleteUser)

		// superAdminGroup.GET("/categories", categoryController.ListCategories)
		// superAdminGroup.POST("/categories", categoryController.CreateCategory)
		// superAdminGroup.PUT("/categories/:id", categoryController.UpdateCategory)
		// superAdminGroup.DELETE("/categories/:id", categoryController.DeleteCategory)
	}

	// ==== VENDOR ROUTES ====
	vendorRoutes := r.Group("/vendors")
	vendorRoutes.Use(middlewares.AuthMiddleware(db))
	{
		vendorRoutes.POST("/apply", vendorController.VendorApply) // user apply
	}

	// ==== SUPERADMIN VENDOR MANAGEMENT ====
	superAdminVendor := r.Group("/super-admin/vendors")
	superAdminVendor.Use(middlewares.AuthMiddleware(db), middlewares.SuperAdminMiddleware(db))
	{
		superAdminVendor.GET("/", vendorController.ListVendors)
		superAdminVendor.GET("/:id", vendorController.GetVendor)
		superAdminVendor.PUT("/:id/approve", vendorController.ApproveVendor)
		superAdminVendor.PUT("/:id/reject", vendorController.RejectVendor)
		superAdminVendor.PUT("/:id/suspend", vendorController.SuspendVendor)
	}

	return r
}
