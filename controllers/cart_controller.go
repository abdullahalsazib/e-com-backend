package controllers

import (
	"net/http"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CartController struct {
	DB *gorm.DB
}

func NewCartController(DB *gorm.DB) CartController {
	return CartController{DB}
}

// GetCart retrieves the user's cart with items
func (cc *CartController) GetCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var cart models.Cart
	result := cc.DB.Preload("CartItems.Product").Where("user_id = ?", userID).First(&cart)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create a new cart if not exists
			newCart := models.Cart{UserID: userID}
			if err := cc.DB.Create(&newCart).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": newCart})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": cart})
}

// AddToCart adds an item to the cart
func (cc *CartController) AddToCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var request models.AddToCartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify product exists
	var product models.Product
	if err := cc.DB.First(&product, request.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get or create user's cart
	var cart models.Cart
	if err := cc.DB.Where("user_id = ?", userID).FirstOrCreate(&cart, models.Cart{UserID: userID}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get/create cart"})
		return
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	result := cc.DB.Where("cart_id = ? AND id = ?", cart.ID, request.ProductID).First(&existingItem)

	if result.Error == nil {
		// Update quantity if item exists
		existingItem.Quantity += request.Quantity
		if err := cc.DB.Save(&existingItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	} else {
		// Create new cart item
		newItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: request.ProductID,
			Quantity:  request.Quantity,
		}
		if err := cc.DB.Create(&newItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}
	}

	// Return updated cart
	cc.DB.Preload("CartItems.Product").First(&cart, cart.ID)
	c.JSON(http.StatusOK, gin.H{"data": cart})
}

// UpdateCartItem updates a cart item's quantity
func (cc *CartController) UpdateCartItem(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("itemId")

	var request models.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify cart item exists and belongs to user
	var cartItem models.CartItem
	if err := cc.DB.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
		First(&cartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	// Update quantity
	cartItem.Quantity = request.Quantity
	if err := cc.DB.Save(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
		return
	}

	// Return updated cart
	var cart models.Cart
	cc.DB.Preload("CartItems.Product").First(&cart, cartItem.CartID)
	c.JSON(http.StatusOK, gin.H{"data": cart})
}

// RemoveFromCart removes an item from the cart
func (cc *CartController) RemoveFromCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	itemID := c.Param("itemId")

	// Verify cart item exists and belongs to user
	var cartItem models.CartItem
	if err := cc.DB.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ?", itemID, userID).
		First(&cartItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
		return
	}

	if err := cc.DB.Delete(&cartItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
		return
	}

	// Return updated cart
	var cart models.Cart
	cc.DB.Preload("CartItems.Product").First(&cart, cartItem.CartID)
	c.JSON(http.StatusOK, gin.H{"data": cart})
}

// ClearCart removes all items from the cart
func (cc *CartController) ClearCart(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Get user's cart
	var cart models.Cart
	if err := cc.DB.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Delete all cart items
	if err := cc.DB.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}
