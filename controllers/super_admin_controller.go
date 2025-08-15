package controllers

import (
	"net/http"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SuperAdminController struct {
	DB *gorm.DB
}

func NewSuperAdminController(db *gorm.DB) *SuperAdminController {
	return &SuperAdminController{DB: db}
}

// List all users with roles
func (sac *SuperAdminController) ListUsers(c *gin.Context) {
	var users []models.User
	if err := sac.DB.Preload("Roles").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// Delete user by ID with cascade delete
func (sac *SuperAdminController) DeleteUserByID(c *gin.Context) {
	id := c.Param("id")

	// Delete related vendors first
	if err := sac.DB.Unscoped().Where("user_id = ?", id).Delete(&models.Vendor{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related vendors"})
		return
	}

	// Delete related products
	if err := sac.DB.Unscoped().Where("user_id = ?", id).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related products"})
		return
	}

	// Finally delete the user
	if err := sac.DB.Unscoped().Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User and related data deleted successfully"})
}
