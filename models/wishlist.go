package models

import "time"

type WishlistItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"` // Foreign key to user
	ProductID uint      `gorm:"not null" json:"product_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Add any additional fields needed
}
