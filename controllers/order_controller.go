package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderController struct {
	DB *gorm.DB
}

func NewOrderController(DB *gorm.DB) OrderController {
	return OrderController{DB}
}

// CreateOrder creates a new order from the user's cart
func (oc *OrderController) CreateOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var request models.CreateOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cart models.Cart
	if err := oc.DB.Preload("CartItems.Product").Where("user_id = ?", userID).First(&cart).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	if len(cart.CartItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	var totalAmount float64
	order := models.Order{
		UserID:          userID,
		Status:          models.OrderStatusPending,
		ShippingAddress: request.ShippingAddress,
		PaymentMethod:   request.PaymentMethod,
	}

	tx := oc.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	for _, item := range cart.CartItems {
		if item.Product.Stock < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Not enough stock for %s", item.Product.Name)})
			return
		}

		totalAmount += item.Product.Price * float64(item.Quantity)

		order.Items = append(order.Items, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.Product.Price,
		})

		if err := tx.Model(&item.Product).Update("stock", item.Product.Stock-item.Quantity).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	order.TotalAmount = totalAmount

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

// GetOrders retrieves all orders for the user
func (oc *OrderController) GetOrders(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var orders []models.Order
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	if err := oc.DB.Preload("Items.Product").Where("user_id = ?", userID).
		Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetOrder retrieves a specific order
func (oc *OrderController) GetOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("orderId")

	var order models.Order
	if err := oc.DB.Preload("Items.Product").Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// UpdateOrderStatus updates an order's status (admin only)
func (oc *OrderController) UpdateOrderStatus(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" && role != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	orderID := c.Param("orderId")
	var request models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := oc.DB.First(&order, orderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	order.Status = request.Status
	if err := oc.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// CancelOrder allows a user to cancel their order
func (oc *OrderController) CancelOrder(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderID := c.Param("orderId")

	var order models.Order
	if err := oc.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order"})
		}
		return
	}

	if order.Status != models.OrderStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be cancelled"})
		return
	}

	order.Status = models.OrderStatusCancelled
	if err := oc.DB.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order cancelled successfully"})
}
