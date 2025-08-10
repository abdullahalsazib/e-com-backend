package controllers

import "gorm.io/gorm"

type SuperAdminController struct {
	DB *gorm.DB
}
