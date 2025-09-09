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

	// origin := os.Getenv("ALLOW_ORIGINS")
	//  CORS CONFIG
	corsConfig := cors.Config{
		AllowOrigins:     []string{"https://e-com-nextjs-six.vercel.app"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	//  CONTROLLERS
	authController := controllers.NewAuthController(db)
	superAdminController := controllers.NewSuperAdminController(db)
	productController := controllers.NewProductController(db)
	cartController := controllers.NewCartController(db)
	orderController := controllers.NewOrderController(db)
	categoryController := controllers.NewCategoryController(db)
	wishlistController := controllers.NewWishlistController(db)
	vendorController := controllers.NewVendorController(db, &authController)

	//  PUBLIC ROUTES
	r.POST("/register", authController.Register)
	r.POST("/login", authController.Login)
	r.POST("/refresh", authController.RefreshToken)
	r.POST("/logout", authController.Logout)
	r.GET("/categorys", categoryController.GetCategories)

	//  AUTH GROUP
	authGroup := r.Group("/auth")
	authGroup.Use(middlewares.AuthMiddleware(db))
	{
		authGroup.GET("/me", authController.GetProfile)
	}

	//  PRODUCT ROUTES
	productGroup := r.Group("/api/v1/products")
	{
		// public's - customer's
		customerPruduct := productGroup.Group("/customer")
		{
			customerPruduct.GET("", productController.GetProductsCustomer)
			customerPruduct.GET("/:id", productController.GetProductByIDCustomer)
		}

		// protected (admin or seller or vendor)
		vendorProduct := productGroup.Group("/vendor")
		vendorProduct.Use(middlewares.AuthMiddleware(db), middlewares.AdminSellerMiddleware())
		{

			vendorProduct.GET("", productController.GetProductsVendor)
			vendorProduct.GET("/:id", productController.GetProductByIDVendor)
			vendorProduct.POST("", productController.CreateProduct)
			vendorProduct.PUT("/:id", productController.UpdateProduct)
			vendorProduct.DELETE("/:id", productController.DeleteProduct)

			// status update
			vendorProduct.PUT("/:id/status", productController.UpdateStatus)
		}

		// superadmin can view all product (all without draft)
		superadminProduct := productGroup.Group("/superadmin")
		{
			superadminProduct.GET("", productController.GetProductsSuperadmin)
			superadminProduct.GET("/:id", productController.GetProductByIDSuperadmin)
		}
	}

	//  ORDER ROUTES
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

	//  WISHLIST ROUTES
	wishlistGroup := authGroup.Group("/wishlist")
	{
		wishlistGroup.POST("/add", wishlistController.AddToWishlist)
		wishlistGroup.GET("", wishlistController.GetWishlist)
		wishlistGroup.DELETE("/:id", wishlistController.RemoveFromWishlist)
		wishlistGroup.DELETE("/clear", wishlistController.ClearWishlist)
		wishlistGroup.PUT("/update/:id", wishlistController.UpdateWishlistItem)
		wishlistGroup.POST("/import", wishlistController.ImportWishlist)
	}

	//  CART ROUTES
	cartGroup := authGroup.Group("/cart")
	{
		cartGroup.GET("/", cartController.GetCart)
		cartGroup.POST("/items", cartController.AddToCart)
		cartGroup.PUT("/items/:itemId", cartController.UpdateCartItem)
		cartGroup.DELETE("/items/:itemId", cartController.RemoveFromCart)
		cartGroup.DELETE("/clear", cartController.ClearCart)
	}

	//  PRODUCT CATEGORY ROUTES
	categoryRoutes := r.Group("/categories")
	{
		categoryRoutes.GET("/", categoryController.GetCategories)
		categoryRoutes.GET("/:id", categoryController.GetCategories)
	}

	//  SUPERADMIN  MANAGEMENT
	superAdminGroup := r.Group("/super-admin")
	superAdminGroup.Use(middlewares.AuthMiddleware(db), middlewares.SuperAdminMiddleware(db))
	{
		superAdminGroup.GET("/users", superAdminController.ListUsers)
		superAdminGroup.DELETE("/users/:id", superAdminController.DeleteUserByID)
		// user related
		superAdminGroup.GET("/users/:id", authController.GetUser)
		superAdminGroup.PUT("/users/:id/role", authController.UpdateUserRole)
		// superAdminGroup.DELETE("/users/:id", authController.DeleteUser) // hard delete

		superAdminGroup.GET("/categories", categoryController.ListCategories)
		superAdminGroup.POST("/categories", categoryController.CreateCategory)
		superAdminGroup.PUT("/categories/:id", categoryController.UpdateCategory)
		superAdminGroup.DELETE("/categories/:id", categoryController.DeleteCategory)

	}

	//  VENDOR ROUTES
	vendorRoutes := r.Group("/vendors")
	vendorRoutes.GET("/:id", vendorController.GetVendor)
	vendorRoutes.Use(middlewares.AuthMiddleware(db))
	{
		vendorRoutes.POST("/apply", vendorController.VendorApply) // user apply
		// apply for delete vendor account
	}

	//  SUPERADMIN VENDOR MANAGEMENT
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

// products - who can see

// user/customer -> /api/v1/products -> getProducts_user
// vendor/admin -> /auth/v1/products -> getProducts_admin
// superadmin -> /superadmin/v1/products -> getPrducts_superadmin
