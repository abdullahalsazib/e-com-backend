// package routes

// import (
// 	"time"

// 	"github.com/abdullahalsazib/e-com-backend/controllers"
// 	"github.com/abdullahalsazib/e-com-backend/middlewares"
// 	"github.com/gin-contrib/cors"
// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func SetupRoutes(db *gorm.DB) *gin.Engine {
// 	r := gin.Default()

// 	config := cors.Config{
// 		AllowOrigins: []string{"http://localhost:3000"},
// 		// AllowOrigins:     []string{"*"},
// 		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
// 		ExposeHeaders:    []string{"Content-Length"},
// 		AllowCredentials: true,
// 		MaxAge:           12 * time.Hour,
// 	}

// 	r.Use(cors.New(config))

// 	r.Use(cors.New(cors.Config{
// 		AllowOrigins:     []string{"http://localhost:3000"},
// 		AllowMethods:     []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
// 		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
// 		ExposeHeaders:    []string{"Content-Length"},
// 		AllowCredentials: true,
// 	}))
// 	// initilize controllers
// 	// auth
// 	authController := controllers.NewAuthController(db)

// 	// super admin
// 	superAdminController := controllers.NewSuperAdminController(db)

// 	// product
// 	productController := controllers.NewProductController(db)
// 	// cart
// 	cartController := controllers.NewCartController(db)
// 	// order
// 	orderController := controllers.NewOrderController(db)
// 	// seller

// 	// categories
// 	categoryController := controllers.NewCategoryController(db)

// 	// wishlist
// 	wishlistController := controllers.NewWishlistController(db)

// 	// public route
// 	r.POST("/register", authController.Register)
// 	r.POST("/login", authController.Login)
// 	r.POST("/refresh", authController.RefreshToken)
// 	r.POST("/logout", authController.Logout)
// 	r.GET("/categorys", categoryController.GetCategories)
// 	//
// 	authGroup := r.Group("/auth")
// 	{
// 		authGroup.Use(middlewares.AuthMiddleware(db))
// 		{
// 			authGroup.GET("/me", authController.GetProfile)
// 			authGroup.POST("/logout", authController.Logout)
// 		}

// 	}
// 	// product routes
// 	productGrup := r.Group("/api/v1/products")
// 	{
// 		// public route (no auth use)
// 		productGrup.GET("", productController.GetProducts)
// 		productGrup.GET("/:id", productController.GetProduct)

// 		// protected routes (require auth - admin or !)
// 		protectedGroup := productGrup.Group("")
// 		protectedGroup.Use(middlewares.AuthMiddleware(db))
// 		protectedGroup.Use(middlewares.AdminSellerMiddleware())
// 		{
// 			// auth is required
// 			protectedGroup.POST("/", productController.CreateProduct)
// 			protectedGroup.PUT("/:id", productController.UpdateProduct)
// 			protectedGroup.DELETE("/:id", productController.DeleteProduct)
// 		}

// 	}

// 	order := authGroup.Group("/orders")
// 	{
// 		order.POST("/", orderController.CreateOrder)
// 		order.GET("/", orderController.GetOrders)
// 		order.GET("/:orderId", orderController.GetOrder)
// 		order.PUT("/:orderId/cancel", orderController.CancelOrder)

// 		// admin only route
// 		admin := order.Group("/")
// 		{
// 			admin.Use(middlewares.AdminMiddleware(db))
// 			admin.PUT("/:orderId/status", orderController.UpdateOrderStatus)
// 		}
// 	}

// 	wishlistGroup := authGroup.Group("/wishlist")
// 	{
// 		wishlistGroup.Use(middlewares.AuthMiddleware(db))
// 		wishlistGroup.POST("/add", wishlistController.AddToWishlist)
// 		wishlistGroup.GET("", wishlistController.GetWishlist)
// 		wishlistGroup.DELETE("/:id", wishlistController.RemoveFromWishlist)
// 		wishlistGroup.DELETE("/clear", wishlistController.ClearWishlist)
// 		wishlistGroup.PUT("/update/:id", wishlistController.UpdateWishlistItem)
// 		wishlistGroup.POST("/import", wishlistController.ImportWishlist)

// 	}

// 	cartGroup := authGroup.Group("/cart")
// 	{
// 		cartGroup.Use(middlewares.AuthMiddleware(db))
// 		cartGroup.GET("/", cartController.GetCart)
// 		cartGroup.POST("/items", cartController.AddToCart)
// 		cartGroup.PUT("/items/:itemId", cartController.UpdateCartItem)
// 		cartGroup.DELETE("/items/:itemId", cartController.RemoveFromCart)
// 		cartGroup.DELETE("/clear", cartController.ClearCart)
// 	}

// 	superAdminRoutes := r.Group("/super-admin")
// 	{

// 		superAdminRoutes.GET("/vendors", superAdminController.GetVendors)
// 		superAdminRoutes.GET("/vendors/:id", superAdminController.GetVendorByID)
// 		superAdminRoutes.PUT("/vendors/:id/approve", superAdminController.ApproveVendor)
// 		superAdminRoutes.PUT("/vendors/:id/reject", superAdminController.RejectVendor)
// 	}

// 	// vendor routes
// 	vc := controllers.NewVendorController(db)

// 	adminGroup := r.Group("/admin")
// 	adminGroup.GET("/vendors", vc.ListVendors)
// 	adminGroup.GET("/vendors/:id", vc.GetVendor)
// 	adminGroup.PUT("/vendors/:id/approve", vc.ApproveVendor)
// 	adminGroup.PUT("/vendors/:id/reject", vc.RejectVendor)
// 	adminGroup.PUT("/vendors/:id/suspend", vc.SuspendVendor)
// 	return r
// }

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
	// superAdminController := controllers.NewSuperAdminController(db)
	productController := controllers.NewProductController(db)
	cartController := controllers.NewCartController(db)
	orderController := controllers.NewOrderController(db)
	categoryController := controllers.NewCategoryController(db)
	wishlistController := controllers.NewWishlistController(db)
	vendorController := controllers.NewVendorController(db)

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
