package controllers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abdullahalsazib/e-com-backend/models"
	"github.com/abdullahalsazib/e-com-backend/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(DB *gorm.DB) AuthController {
	return AuthController{DB}
}

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4"` // Minimum 6-8 characters
}

func (ac *AuthController) Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := ac.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User with this email already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error checking for existing user", "details": err.Error()})
		return
	}

	// Hash password
	hashedPass, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	// Find default "user" role
	var userRole models.Role
	if err := ac.DB.Where("slug = ?", "user").First(&userRole).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Default role not found"})
		return
	}

	// Create user
	user := models.User{
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPass),
		IsActive:  true,
		Roles:     []models.Role{userRole}, // Initially no roles
		CreatedBy: 0,                       // system-created
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	// Ensure "user" role exists
	var role models.Role
	if err := ac.DB.Where("name = ?", "User").First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			role = models.Role{Name: "user"}
			ac.DB.Create(&role)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find/create role"})
			return
		}
	}

	// Assign role to user
	ac.DB.Create(&models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// login

func (ac *AuthController) Login(c *gin.Context) {
	var payload struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// JSON bind and validation
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	// Preload roles to get user's roles
	if err := ac.DB.Preload("Roles").Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Collect role slugs
	var roleSlugs []string
	for _, role := range user.Roles {
		roleSlugs = append(roleSlugs, role.Slug)
	}
	rolesStr := strings.Join(roleSlugs, ",")

	// Generate access and refresh tokens
	accessToken, refreshToken, err := utils.GenerateTokenPair(user.ID, rolesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating token"})
		return
	}

	// Save refresh token in DB
	refreshTokenExp := time.Now().Add(7 * 24 * time.Hour).Unix()
	token := models.Token{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: refreshTokenExp,
		Role:      rolesStr,
	}
	if err := ac.DB.Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving refresh token"})
		return
	}

	// Set refresh token as HttpOnly cookie
	c.SetCookie(
		"refresh_token",
		refreshToken,
		7*24*60*60, // 7 day
		"/",
		"localhost",
		false,
		true, // HttpOnly true
	)

	// Return access token in JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successfully",
		"access_token": accessToken,
		"expires_in":   15 * 60, // ১৫ min
		"role":         roleSlugs,
	})
}

// GET /profile
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	var user models.User

	if err := ac.DB.Preload("Roles").Preload("Vendor").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user profile"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "User profile fetched successfully",
		"data": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"roles": user.Roles,
			"vendor": func() interface{} {
				if user.Vendor != nil {
					return gin.H{
						"vendor_id":     user.Vendor.ID,
						"user_id":       user.Vendor.UserID,
						"vendor_name":   user.Vendor.ShopName,
						"vendor_status": user.Vendor.Status,
						"approved_by":   user.Vendor.ApprovedBy,
					}
				}
				return nil
			}(),
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		},
	})

}

// refresh
func (ac *AuthController) RefreshToken(c *gin.Context) {
	// Refresh Token cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	var token models.Token
	if err := ac.DB.Where("token = ? AND user_id = ?", refreshToken, claims.UserID).First(&token).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not valid"})
		return
	}

	if time.Now().Unix() > token.ExpiresAt {
		ac.DB.Delete(&token)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	var user models.User
	if err := ac.DB.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Roles  Slug
	var roleSlugs []string
	for _, role := range user.Roles {
		roleSlugs = append(roleSlugs, role.Slug)
	}
	rolesStr := strings.Join(roleSlugs, ",")

	// new Access Token generate (Refresh Token if before have)
	newAccessToken, _, err := utils.GenerateTokenPair(user.ID, rolesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// new Access Token
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_in":   15 * 60,
		"role":         roleSlugs,
		"message":      "Token refreshed successfully",
	})
}

// logout
func (ac *AuthController) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token not found"})
		return
	}

	// Hard delete the token from DB
	if err := ac.DB.Unscoped().Where("token = ?", refreshToken).Delete(&models.Token{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete refresh token"})
		return
	}

	// Clear cookie
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"localhost",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func (ac *AuthController) AddAdminRoleByUser(user *models.User) error {
	var adminRole models.Role
	if err := ac.DB.Where("slug = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}

	for _, r := range user.Roles {
		if r.ID == adminRole.ID {
			return nil
		}
	}

	return ac.DB.Model(user).Association("Roles").Append(&adminRole)
}

func (ac *AuthController) RemoveAdminRoleByUser(user *models.User) error {
	var adminRole models.Role
	if err := ac.DB.Where("slug = ?", "admin").First(&adminRole).Error; err != nil {
		return err
	}
	return ac.DB.Model(user).Association("Roles").Delete(&adminRole)
}

func (ac *AuthController) AddRoleByUserSlug(user *models.User, slug string) error {
	var role models.Role
	if err := ac.DB.Where("slug = ?", slug).First(&role).Error; err != nil {
		return err
	}

	for _, r := range user.Roles {
		if r.ID == role.ID {
			return nil
		}
	}

	return ac.DB.Model(user).Association("Roles").Append(&role)
}
