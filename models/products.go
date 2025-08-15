package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	UserID      uint    `json:"user_id" gorm:"not null"` // Seller ID
	User        User    `json:"-" gorm:"foreignKey:UserID"`
	Name        string  `json:"name" gorm:"not null"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	ImageURL    string  `json:"image_url"`
	CategoryID  uint    `gorm:"not null"`
	Status      string  `json:"status" gorm:"type:enum('published','private','pending');default:'pending'"`
}
