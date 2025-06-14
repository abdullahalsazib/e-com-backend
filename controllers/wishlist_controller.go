package controllers

import (
	"net/http"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WishlistController struct {
	DB *gorm.DB
}

func NewWishlistController(DB *gorm.DB) WishlistController {
	return WishlistController{DB}
}

// AddToWishlist adds a new item to the user's wishlist
func (ws *WishlistController) AddToWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var item struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if product exists (implementation depends on your product service)
	if !ws.productExists(item.ProductID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product does not exist"})
		return
	}

	// Check if item already exists in wishlist
	var existingItem models.WishlistItem
	if err := ws.DB.Where("user_id = ? AND product_id = ?", userID, item.ProductID).First(&existingItem).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Product already in wishlist"})
		return
	}

	wishlistItem := models.WishlistItem{
		UserID:    userID,
		ProductID: item.ProductID,
	}

	if err := ws.DB.Create(&wishlistItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to wishlist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "product add successfully.",
	})
}

// GetWishlist retrieves all wishlist items for the authenticated user
func (ws *WishlistController) GetWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var items []models.WishlistItem

	if err := ws.DB.Where("user_id = ?", userID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wishlist"})
		return
	}

	// You might want to join with product data here
	var response []map[string]interface{}
	for _, item := range items {
		product, err := ws.getProductDetails(item.ProductID) // Implement this function
		if err != nil {
			continue // or handle error
		}

		response = append(response, map[string]interface{}{
			"id":         item.ID,
			"product":    product,
			"created_at": item.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// RemoveFromWishlist deletes a specific item from the user's wishlist
func (ws *WishlistController) RemoveFromWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("id")

	// First verify the item belongs to the user
	var item models.WishlistItem
	if err := ws.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wishlist item not found"})
		return
	}

	if err := ws.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from wishlist"})
}

// ClearWishlist removes all items from the user's wishlist
func (ws *WishlistController) ClearWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := ws.DB.Where("user_id = ?", userID).Delete(&models.WishlistItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wishlist cleared successfully"})
}

// UpdateWishlistItem modifies an existing wishlist item
func (ws *WishlistController) UpdateWishlistItem(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("id")

	var updateData struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the item exists and belongs to the user
	var item models.WishlistItem
	if err := ws.DB.Where("id = ? AND user_id = ?", itemID, userID).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wishlist item not found"})
		return
	}

	// Check if new product exists
	if !ws.productExists(updateData.ProductID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product does not exist"})
		return
	}

	// Update the item
	item.ProductID = updateData.ProductID
	if err := ws.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wishlist item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ImportWishlist imports items from localStorage to the database
func (ws *WishlistController) ImportWishlist(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	var items []struct {
		ProductID uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&items); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate all products exist and prepare for bulk insert
	var validItems []models.WishlistItem
	for _, item := range items {
		if ws.productExists(item.ProductID) {
			validItems = append(validItems, models.WishlistItem{
				UserID:    userID,
				ProductID: item.ProductID,
			})
		}
	}

	// Use transaction for bulk insert
	tx := ws.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(&validItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import wishlist"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit wishlist import"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wishlist imported successfully",
		"count":   len(validItems),
	})
}

// productExists checks if a product with the given ID exists
func (ws *WishlistController) productExists(productID uint) bool {
	// Implement based on your product service
	// This could be a database check or API call
	var count int64
	ws.DB.Model(&models.Product{}).Where("id = ?", productID).Count(&count)
	return count > 0
}

// getProductDetails fetches product information
func (ws *WishlistController) getProductDetails(productID uint) (map[string]interface{}, error) {
	// Implement based on your product service
	var product models.Product
	if err := ws.DB.Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":    product.ID,
		"name":  product.Name,
		"price": product.Price,
		// Add other product fields as needed
	}, nil
}
