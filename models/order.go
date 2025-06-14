package models

import (
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	gorm.Model
	UserID          uint        `json:"user_id" gorm:"not null"`
	Status          OrderStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TotalAmount     float64     `json:"total_amount" gorm:"not null"`
	Items           []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	ShippingAddress string      `json:"shipping_address" gorm:"not null"`
	PaymentMethod   string      `json:"payment_method" gorm:"not null"`
}

type OrderItem struct {
	gorm.Model
	OrderID   uint    `json:"order_id" gorm:"not null"`
	ProductID uint    `json:"product_id" gorm:"not null"`
	Quantity  int     `json:"quantity" gorm:"not null"`
	UnitPrice float64 `json:"unit_price" gorm:"not null"`
	Product   Product `json:"product" gorm:"foreignKey:ProductID"`
}

type CreateOrderRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required"`
	PaymentMethod   string `json:"payment_method" binding:"required"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required,oneof=pending processing shipped delivered cancelled"`
}
