package models

import "gorm.io/gorm"

type Token struct {
	gorm.Model
	UserID    uint   `json:"user_id" gorm:"not null"`
	Token     string `json:"token" gorm:"unique;not null"`
	Role      string `json:"role" gorm:"not null"`
	ExpiresAt int64  `json:"expires_at" gorm:"not null"`
}
