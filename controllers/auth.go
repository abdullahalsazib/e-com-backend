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
		7*24*60*60, // 7 দিন (সেকেন্ডে)
		"/",
		"localhost", // ডেভেলপমেন্টে localhost, প্রোডাকশনে তোমার ডোমেইন দিবে
		false,       // ডেভেলপমেন্টে false, প্রোডাকশনে true (https এর জন্য)
		true,        // HttpOnly true
	)

	// Return access token in JSON response
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successfully",
		"access_token": accessToken,
		"expires_in":   15 * 60, // ১৫ মিনিট
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

	// এখানে Preload("Roles") ব্যবহার করতে হবে
	if err := ac.DB.Preload("Roles").First(&user, userID).Error; err != nil {
		switch err {
		case gorm.ErrRecordNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching user profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile fetched successfully",
		"data": gin.H{
			"id":        user.ID,
			"name":      user.Name,
			"email":     user.Email,
			"roles":     user.Roles, // এখন এটা null হবে না
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		},
	})
}

// refresh
func (ac *AuthController) RefreshToken(c *gin.Context) {
	// Refresh Token cookie থেকে নেওয়া
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not found"})
		return
	}

	// Token Parse ও Validate করা
	claims, err := utils.ParseToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// ডাটাবেজ থেকে Refresh Token ভ্যালিডেট করা
	var token models.Token
	if err := ac.DB.Where("token = ? AND user_id = ?", refreshToken, claims.UserID).First(&token).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token not valid"})
		return
	}

	// Token মেয়াদ শেষ হয়েছে কি না চেক করা
	if time.Now().Unix() > token.ExpiresAt {
		// মেয়াদ শেষ হলে Token ডিলিট করাও ভালো (Optional)
		ac.DB.Delete(&token)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}

	// ইউজারের তথ্য ও Roles লোড করা
	var user models.User
	if err := ac.DB.Preload("Roles").First(&user, claims.UserID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Roles থেকে Slug নিয়ে আসা
	var roleSlugs []string
	for _, role := range user.Roles {
		roleSlugs = append(roleSlugs, role.Slug)
	}
	rolesStr := strings.Join(roleSlugs, ",")

	// নতুন Access Token জেনারেট (Refresh Token আগেরই থাকবে)
	newAccessToken, _, err := utils.GenerateTokenPair(user.ID, rolesStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	// নতুন Access Token রেসপন্স পাঠানো
	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccessToken,
		"expires_in":   15 * 60, // ১৫ মিনিট
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
