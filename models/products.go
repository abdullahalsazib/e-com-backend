package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	UserID      uint    `json:"user_id" gorm:"not null"` // Seller ID	ProductName string `json:"product_name"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
	CreatedBy   uint    `json:"created_by"`
	User        User    `json:"-" gorm:"foreignKey:UserID"`
	CategoryID  uint    `gorm:"not null"`
}
