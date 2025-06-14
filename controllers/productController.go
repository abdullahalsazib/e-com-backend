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

func NewProductController(DB *gorm.DB) ProductController {
	return ProductController{DB}
}

// Create Product - POST
func (pc *ProductController) CreateProduct(c *gin.Context) {
	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert userId to uint
	var userIDUint uint
	if id, ok := userId.(uint); ok {
		userIDUint = id
	} else if idf, ok := userId.(float64); ok {
		userIDUint = uint(idf)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	// Parse JSON
	var payload struct {
		Name        string  `json:"product_name"  binding:"required"`
		Description string  `json:"description" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		Stock       int     `json:"stock" binding:"required"`
		ImageURL    string  `json:"image_url" binding:"required"`
		CategoryID  uint    `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if category exists
	var category models.Category
	if err := pc.DB.First(&category, payload.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	// Create product
	newProduct := models.Product{
		Name:        payload.Name,
		Description: payload.Description,
		Price:       payload.Price,
		Stock:       payload.Stock,
		ImageURL:    payload.ImageURL,
		CategoryID:  payload.CategoryID,
		UserID:      userIDUint,
		CreatedBy:   userIDUint,
	}

	if err := pc.DB.Create(&newProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Product created successfully", "data": newProduct})
}

func (pc *ProductController) GetProducts(c *gin.Context) {
	var products []models.Product
	result := pc.DB.Find(&products)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, products)
}

// get product by id
func (pc *ProductController) GetProduct(c *gin.Context) {
	produtId := c.Param("id")

	var product models.Product
	result := pc.DB.First(&product, "ID = ?", produtId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": product})
}

func (pc *ProductController) UpdateProduct(c *gin.Context) {
	productIdStr := c.Param("id")
	productId, err := strconv.ParseUint(productIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	userId, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var userIDUint uint
	switch v := userId.(type) {
	case uint:
		userIDUint = v
	case float64:
		userIDUint = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var payload struct {
		Name        string  `json:"product_name" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Price       float64 `json:"price" binding:"required"`
		Stock       int     `json:"stock" binding:"required"`
		ImageURL    string  `json:"image_url" binding:"required"`
		CategoryID  uint    `json:"category_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updatedProduct models.Product

	// fmt.Println("-------------------: ")
	// fmt.Println("paramis: ", productIdStr)
	// fmt.Println("-------------------: ")
	result := pc.DB.First(&updatedProduct, productId)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if updatedProduct.CreatedBy != userIDUint {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own products"})
		return
	}

	updatedProduct.Name = payload.Name
	updatedProduct.Description = payload.Description
	updatedProduct.Price = payload.Price
	updatedProduct.Stock = payload.Stock
	updatedProduct.ImageURL = payload.ImageURL
	updatedProduct.CategoryID = payload.CategoryID

	if err := pc.DB.Save(&updatedProduct).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully", "data": updatedProduct})
}

func (pc *ProductController) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	productId, err := strconv.Atoi(id)
	userID := c.GetUint("user_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product id"})
		return
	}

	var product models.Product
	if err := pc.DB.First(&product, productId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding product"})
		}
		return
	}

	if product.CreatedBy != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own products"})
		return
	}

	if err := pc.DB.Unscoped().Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error Delete Product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product delete succesfully"})
}

type CategoryController struct {
	DB *gorm.DB
}

func NewCategoryController(DB *gorm.DB) CategoryController {
	return CategoryController{DB}
}

// GET /categories
func (cc *CategoryController) GetCategories(c *gin.Context) {
	var categories []models.Category
	// var categories models.Category
	if err := cc.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get categories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}
