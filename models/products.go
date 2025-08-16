package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user" gorm:"foreignKey:UserID"`

	VendorID uint   `json:"vendor_id" gorm:"not null"`
	Vendor   Vendor `json:"vendor" gorm:"foreignKey:VendorID"`

	Name        string   `json:"name" gorm:"not null"`
	Description string   `json:"description"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	ImageURL    string   `json:"image_url"`
	CategoryID  uint     `json:"category_id" gorm:"not null"`
	Category    Category `json:"category" gorm:"foreignKey:CategoryID"`
	Status      string   `gorm:"size:50;default:'draft'" json:"status"`
}
