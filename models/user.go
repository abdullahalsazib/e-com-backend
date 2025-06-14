package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string `json:"name"`
	Email     string `json:"email" gorm:"uniqe" binding:"required,email"`
	Password  string `json:"-" binding:"required"`
	Role      string `json:"role" gorm:"default:user" binding:"required"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	LastLogin string `json:"last_login"`
	CreatedBy uint   `json:"created_by"`
}
