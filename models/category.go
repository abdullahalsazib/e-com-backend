package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name        string    `gorm:"unique;not null"`
	Slug        string    `gorm:"unique;not null"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	ParentID    *uint     `json:"parent_id"`
	Parent      *Category `gorm:"foreignKey:ParentID"`
}
