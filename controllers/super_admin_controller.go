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

// Get single user
func (ac *AuthController) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := ac.DB.Preload("Roles").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Update user roles
func (ac *AuthController) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var payload struct {
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := ac.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update roles
	if err := ac.DB.Model(&user).Association("Roles").Replace(&payload.RoleIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated user
	if err := ac.DB.Preload("Roles").First(&user, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Hard delete user
func (ac *AuthController) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := ac.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete related records manually if needed
	ac.DB.Where("user_id = ?", id).Delete(&models.Token{})
	ac.DB.Where("user_id = ?", id).Delete(&models.Vendor{})
	ac.DB.Where("user_id = ?", id).Delete(&models.Cart{})
	ac.DB.Where("user_id = ?", id).Delete(&models.WishlistItem{})
	ac.DB.Where("user_id = ?", id).Delete(&models.Order{})

	// Hard delete user
	if err := ac.DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// List all categories
func (cc *CategoryController) ListCategories(c *gin.Context) {
	var categories []models.Category
	if err := cc.DB.Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categories)
}

// Create category
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var payload struct {
		Name     string `json:"name" binding:"required"`
		ParentID *uint  `json:"parent_id"` // optional
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := models.Category{
		Name:     payload.Name,
		ParentID: payload.ParentID,
	}
	if err := cc.DB.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

// Update category
func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := cc.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	var payload struct {
		Name     string `json:"name" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = payload.Name
	category.ParentID = payload.ParentID

	if err := cc.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, category)
}

// Hard delete category and related products
func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category
	if err := cc.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Delete related products
	cc.DB.Where("category_id = ?", id).Delete(&models.Product{})

	// Hard delete category
	if err := cc.DB.Unscoped().Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}
