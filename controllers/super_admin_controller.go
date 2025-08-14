package controllers

import (
	"fmt"
	"net/http"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SuperAdminController struct {
	DB *gorm.DB
}

func NewSuperAdminController(DB *gorm.DB) SuperAdminController {
	return SuperAdminController{DB}
}

// List all users
func (sac *SuperAdminController) ListUsers(c *gin.Context) {
	var users []models.User

	// Preload Roles for all users
	if err := sac.DB.Preload("Roles").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Delete a user by ID
func (sac *SuperAdminController) DeleteUserByID(c *gin.Context) {
	// Parse user ID from URL parameter
	userIDParam := c.Param("id")
	var userID uint
	if _, err := fmt.Sscanf(userIDParam, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var currentUser models.User
	if err := sac.DB.Preload("Roles").First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	//  Delete related Vendor records (hard delete)
	if err := sac.DB.Unscoped().Where("user_id = ?", userID).Delete(&models.Vendor{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vendor records"})
		return
	}

	//  Delete related Product records (hard delete)
	if err := sac.DB.Unscoped().Where("user_id = ?", userID).Delete(&models.Product{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product records"})
		return
	}

	// Finally delete the user (hard delete)
	if err := sac.DB.Unscoped().Delete(&currentUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User and all related records deleted successfully"})
}
