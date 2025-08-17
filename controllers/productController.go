package controllers

import (
	"net/http"
	"strconv"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductController struct {
	DB *gorm.DB
}

func NewProductController(db *gorm.DB) *ProductController {
	return &ProductController{DB: db}
}

// CreateProduct creates a new product
func (pc *ProductController) CreateProduct(c *gin.Context) {
	var payload struct {
		CategoryID  uint    `json:"category_id" binding:"required"` // Category FK
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required"`
		Stock       int     `json:"stock"`
		ImageURL    string  `json:"image_url"`
		Status      string  `json:"status"` // draft, published, private, archived
	}

	// Bind JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Get Vendor record for this user
	var vendor models.Vendor
	if err := pc.DB.Where("user_id = ?", userID).First(&vendor).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vendor not found for this user"})
		return
	}

	// Create Product instance
	product := models.Product{
		UserID:      userID,
		VendorID:    vendor.ID,
		CategoryID:  payload.CategoryID,
		Name:        payload.Name,
		Description: payload.Description,
		Price:       payload.Price,
		Stock:       payload.Stock,
		ImageURL:    payload.ImageURL,
		Status:      payload.Status,
	}

	// Save to DB
	if err := pc.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Preload related fields for response
	if err := pc.DB.Preload("Category").Preload("User").Preload("Vendor").First(&product, product.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load created product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates a product's details
func (pc *ProductController) UpdateProduct(c *gin.Context) {
	// 1. Parse product ID
	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// 2. Fetch product from DB
	var product models.Product
	if err := pc.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// 3. Bind JSON payload
	var payload struct {
		Name        string  `json:"name" binding:"required"`
		Description string  `json:"description"`
		Price       float64 `json:"price" binding:"required"`
		Stock       int     `json:"stock" binding:"required"`
		ImageURL    string  `json:"image_url"`
		CategoryID  uint    `json:"category_id" binding:"required"`
		Status      string  `json:"status"` // optional: draft, published, private, archived
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. Optional: validate status if provided
	if payload.Status != "" {
		validStatuses := map[string]bool{"draft": true, "published": true, "private": true, "archived": true}
		if !validStatuses[payload.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
			return
		}
		product.Status = payload.Status
	}

	// 5. Update allowed fields
	product.Name = payload.Name
	product.Description = payload.Description
	product.Price = payload.Price
	product.Stock = payload.Stock
	product.ImageURL = payload.ImageURL
	product.CategoryID = payload.CategoryID

	if err := pc.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Preload relations for response
	if err := pc.DB.
		Preload("Category").
		Preload("User").Preload("User.Roles").
		Preload("Vendor").Preload("Vendor.User").
		First(&product, productID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Delete Product
func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if err := pc.DB.Delete(&models.Product{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// Update Product Status
func (pc *ProductController) UpdateStatus(c *gin.Context) {

	//  1. Get and validate product ID
	id := c.Param("id")
	productID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	//  2. Fetch product from DB
	var product models.Product
	if err := pc.DB.First(&product, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	//  3. Bind JSON payload
	var req struct {
		Status string `json:"status" binding:"required"` // draft, published, private, archived
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//  4. Validate allowed status
	validStatuses := map[string]bool{"draft": true, "published": true, "private": true, "archived": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value"})
		return
	}

	//  6. Update product status
	product.Status = req.Status
	if err := pc.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := pc.DB.
		Preload("Category").
		Preload("User").Preload("User.Roles").
		Preload("Vendor").Preload("Vendor.User").
		// Preload("Vendor.ApprovedByUser"). // optional
		First(&product, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Customer
func (pc *ProductController) GetProductsCustomer(c *gin.Context) {
	var products []models.Product
	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("status = ?", "published").
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Vendor
func (pc *ProductController) GetProductsVendor(c *gin.Context) {
	rolesVal, ok := c.Get("role")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "roles not found in context"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	userID := userIDVal.(uint)
	roleList := rolesVal.([]string)

	isAdmin := false
	for _, r := range roleList {
		if r == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can access vendor products"})
		return
	}

	var products []models.Product
	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("user_id = ?", userID).
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Superadmin
func (pc *ProductController) GetProductsSuperadmin(c *gin.Context) {
	var products []models.Product
	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("status != ?", "draft").
		Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Customer
func (pc *ProductController) GetProductByIDCustomer(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("id = ? AND status = ?", id, "published").
		First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found or not published"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Vendor/Admin
func (pc *ProductController) GetProductByIDVendor(c *gin.Context) {
	id := c.Param("id")

	rolesVal, ok := c.Get("role")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "roles not found in context"})
		return
	}

	userIDVal, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}
	userID := userIDVal.(uint)
	roleList := rolesVal.([]string)

	// check admin role
	isAdmin := false
	for _, r := range roleList {
		if r == "admin" {
			isAdmin = true
			break
		}
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "only admins can access vendor products"})
		return
	}

	var product models.Product
	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("id = ? AND user_id = ?", id, userID).
		First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// Superadmin
func (pc *ProductController) GetProductByIDSuperadmin(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	if err := pc.DB.
		Preload("Category").
		Preload("User").
		Preload("Vendor").
		Where("id = ? AND status != ?", id, "draft").
		First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found or draft"})
		return
	}

	c.JSON(http.StatusOK, product)
}
