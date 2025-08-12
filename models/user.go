package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name      string `json:"name"`
	Email     string `json:"email" gorm:"unique" binding:"required,email"`
	Password  string `json:"-" binding:"required"`
	Roles     []Role `gorm:"many2many:user_roles;" json:"roles"`
	IsActive  bool   `json:"is_active" gorm:"default:true"`
	LastLogin string `json:"last_login"`
	CreatedBy uint   `json:"created_by"`
}
